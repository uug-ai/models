package models

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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
// The tolerant form carries both a null arm and an $exists arm even though
// {projectId: nil} already matches a missing field in MongoDB. The redundancy
// is intentional: it is the exact shape the Hub API already hand-rolls, so
// those call sites can adopt this helper as a pure substitution with no
// behaviour change to argue about in review.
func ProjectScopeFilter(organisationId string, projectId primitive.ObjectID) bson.M {
	if projectId.IsZero() {
		return nil
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

	return bson.M{"$or": []bson.M{
		{"projectId": projectId},
		{"projectId": bson.M{"$exists": false}},
		{"projectId": nil},
	}}
}
