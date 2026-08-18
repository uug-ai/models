package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// The two helpers below define which project owns an organisation's resources
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
