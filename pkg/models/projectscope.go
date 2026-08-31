package models

import (
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TenancyMode selects the persisted ownership contract used by readers during
// the organisation/project migration.
type TenancyMode string

const (
	TenancyModeLegacy        TenancyMode = "legacy"
	TenancyModeCompatibility TenancyMode = "compatibility"
	TenancyModeCanonical     TenancyMode = "canonical"
)

// ParseTenancyMode parses deployment configuration. An empty value preserves
// migration-safe compatibility reads; unknown values are rejected so a typo
// cannot silently select a different ownership contract.
func ParseTenancyMode(value string) (TenancyMode, error) {
	mode := TenancyMode(strings.ToLower(strings.TrimSpace(value)))
	if mode == "" {
		return TenancyModeCompatibility, nil
	}
	if !mode.IsValid() {
		return "", fmt.Errorf("invalid tenancy mode %q", value)
	}
	return mode, nil
}

func (m TenancyMode) IsValid() bool {
	return m == TenancyModeLegacy || m == TenancyModeCompatibility || m == TenancyModeCanonical
}

// The helpers below define which project owns an organisation's resources
// during the hidden single-project rollout.
//
// The rollout has no project UI, no switcher, and no API surface, so there is
// no authoritative record to consult: every organisation behaves as if it has
// exactly one project. That makes the default a *definition* rather than a
// lookup, and the distinction matters.
//
// A lookup — resolving the project by slug, by "first document", or by an
// isActive flag — can return a different value in two services for the same
// organisation. It diverges when one service is deployed ahead of another, when
// a query orders results differently, or when the collection is empty on an
// instance where the migration has not run. Derived resources would then be
// stamped with projects their organisation's readers never select, and the
// media would simply disappear from the UI.
//
// A pure function cannot diverge. It performs no I/O, so it returns the same
// value on every instance regardless of deployment order or collection state,
// and it produces exactly the value the project materialisation will persist.
// Callers must therefore never "improve" these helpers by adding a query, and
// they must stay free of any dependency that would make one possible — that is
// why they live in this module rather than alongside the device ownership
// resolution in the database module, which does query records.

// DefaultProjectId returns the project that owns an organisation's resources
// while no project has been explicitly assigned.
//
// The default reuses the organisation id. That is not an arbitrary choice: it
// makes the value derivable from data every service already holds, so no
// service needs a project read to agree with the others, and it stays stable
// once real projects are introduced because the default project is minted with
// this id rather than a fresh one.
func DefaultProjectId(organisationId primitive.ObjectID) primitive.ObjectID {
	return organisationId
}

// ResolveProjectId returns the project owning a resource, preferring an
// explicitly stored assignment over the organisation default.
//
// A stored project is an authoritative fact written by a service that knew the
// resource's placement, so it wins. A nil or zero value means "not assigned",
// not "assigned to nothing", and falls back to DefaultProjectId. Treating zero
// as a real project would stamp derived resources with NilObjectID and hide
// them from every project-scoped read.
func ResolveProjectId(organisationId primitive.ObjectID, stored *primitive.ObjectID) primitive.ObjectID {
	if stored != nil && !stored.IsZero() {
		return *stored
	}
	return DefaultProjectId(organisationId)
}

// ProjectScopeFilter returns the clause a reader adds to narrow an
// organisation's documents to one project.
//
// It is the read half of ResolveProjectId, and it lives here for the same
// reason that function does: a writer stamps whatever ResolveProjectId returns,
// and a reader must match exactly that. Split the two across modules and they
// drift — silently, because a predicate that no longer matches what was stamped
// returns zero documents rather than an error. Keeping the pair in one file is
// what makes the drift impossible.
//
// This is NOT a tenant boundary. It narrows within an organisation, so callers
// must apply their own organisation clause alongside it; on its own it would
// match another tenant's documents that happen to share a project id.
//
// The predicate has two shapes, and the asymmetry is deliberate:
//
//   - For a real project it is strict equality. Only documents explicitly
//     stamped for that project belong to it.
//   - For the organisation's default project it also matches documents with no
//     project at all. Every document written before the project axis existed is
//     unstamped, and the default project is where they belong; a strict clause
//     would make the entire pre-rollout history vanish from the UI.
//
// Relaxing only for the default project is the important half. An unconditional
// tolerance reads identically today — during the rollout every project IS the
// organisation default — but the moment a second project exists it would make
// that project's reads match every unstamped document in the organisation. That
// is a cross-project leak that no test written today would catch, because today
// there is no second project to catch it with.
//
// A zero project, or an organisation id that is not an ObjectID, yields nil:
// "do not narrow". That degrades a read to organisation-wide rather than to
// nothing, mirroring the writer's own degradation for an unresolvable
// organisation. It is not a leak — the caller's organisation clause still
// bounds the result — whereas failing to an unmatchable predicate would blank a
// tenant's screen over a resolution glitch.
//
// The tolerant form is a two-element $in rather than an $or over an equality,
// an $exists and a null arm. All three shapes select the same documents —
// {projectId: nil} already matches a missing field in MongoDB, so the $exists
// arm never added a document — but only the $in is answerable from an index.
// A branch testing {$exists: false} cannot be: a missing field and an explicit
// null are stored as the same index entry, so the planner cannot decide the
// difference without fetching the document, and an AND-ed clause it cannot
// push into an index turns an otherwise-indexed read into a full collection
// scan followed by a blocking sort. The $in yields two point bounds instead,
// which is what lets a reader keep using an existing {organisationId, sort key}
// index.
//
// The redundant $or shape this replaces was chosen so the Hub API call sites
// that hand-rolled it could adopt this helper as a pure substitution. They have
// all adopted it, so that reason is spent and the shape is now free to be the
// one the planner can use.
func ProjectScopeFilter(organisationId string, projectId primitive.ObjectID) bson.M {
	return ProjectScopeFilterForMode(TenancyModeCompatibility, organisationId, projectId)
}

// ProjectScopeFilterForMode returns the project predicate for the selected
// migration contract. Legacy mode has no project boundary, compatibility mode
// admits unstamped rows only into the deterministic default project, and
// canonical mode requires an exact projectId.
func ProjectScopeFilterForMode(mode TenancyMode, organisationId string, projectId primitive.ObjectID) bson.M {
	if !mode.IsValid() {
		mode = TenancyModeCompatibility
	}
	if mode == TenancyModeLegacy || projectId.IsZero() {
		return nil
	}
	if mode == TenancyModeCanonical {
		return bson.M{"projectId": projectId}
	}

	organisation, err := primitive.ObjectIDFromHex(organisationId)
	if err != nil {
		return nil
	}

	// Asking DefaultProjectId rather than comparing the hex strings directly
	// keeps this in step with the definition: if the default ever stops being
	// the organisation id, the tolerance follows it automatically.
	if DefaultProjectId(organisation) != projectId {
		return bson.M{"projectId": projectId}
	}

	return bson.M{"projectId": bson.M{"$in": bson.A{projectId, nil}}}
}
