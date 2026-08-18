package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// Project is an optional access boundary within an organisation. Resources
// remain organisation-wide unless their ProjectId is set.
type Project struct {
	Id             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	OrganisationId primitive.ObjectID `json:"organisationId" bson:"organisationId"`
	Name           string             `json:"name" bson:"name"`
	Slug           string             `json:"slug" bson:"slug"`
	Description    string             `json:"description,omitempty" bson:"description,omitempty"`
	IsActive       bool               `json:"isActive" bson:"isActive"`
	Audit          Audit              `json:"audit" bson:"audit"`
}

// DefaultProjectSlug is the reserved slug of the single default project minted
// per organisation. The default project is the sentinel-free "organisation-wide"
// selection: it is a real document (never a stored NilObjectID) so a user's
// active project can always point at something concrete.
const DefaultProjectSlug = "default"

// GetProjectsInput lists the projects in the caller's active organisation.
// Soft-deleted (inactive) projects are omitted unless IncludeInactive is set.
type GetProjectsInput struct {
	User            User `json:"user"`
	IncludeInactive bool `json:"includeInactive"`
}

type GetProjectsOutput struct {
	Projects []Project `json:"projects"`
}

// GetProjectInput resolves a single project by id within the caller's active
// organisation.
type GetProjectInput struct {
	User      User   `json:"user"`
	ProjectId string `json:"projectId"`
}

type GetProjectOutput struct {
	Project *Project `json:"project"`
}

// GetCurrentProjectInput resolves the caller's active project (the one whose id
// equals the caller's projectId, or the organisation default when unset).
type GetCurrentProjectInput struct {
	User User `json:"user"`
}

type GetCurrentProjectOutput struct {
	Project *Project `json:"project"`
}

// SetCurrentProjectInput selects the project used to scope subsequent requests
// for the caller. The project must belong to the caller's active organisation.
type SetCurrentProjectInput struct {
	User      User   `json:"user"`
	ProjectId string `json:"projectId"`
}

type SetCurrentProjectOutput struct {
	Project *Project `json:"project"`
}

// CreateProjectInput creates a project inside the caller's active organisation.
// Ownership, identity and audit stamps are server-managed and ignored on the
// supplied Project.
type CreateProjectInput struct {
	User    User    `json:"user"`
	Project Project `json:"project"`
}

type CreateProjectOutput struct {
	Project *Project `json:"project"`
}

// ProjectUpdate is a partial-update (PATCH) payload for a project. Every field
// is a pointer so a nil value means "not provided" (leave the stored value
// untouched), which is distinct from an explicit zero value (e.g. clearing the
// description or setting IsActive to false). Identity, ownership and audit
// fields are intentionally absent because they are managed server-side.
type ProjectUpdate struct {
	Name        *string `json:"name,omitempty"`
	Slug        *string `json:"slug,omitempty"`
	Description *string `json:"description,omitempty"`
	IsActive    *bool   `json:"isActive,omitempty"`
}

// UpdateProjectInput applies a partial update to a project in the caller's
// active organisation. Only the fields set on the ProjectUpdate payload are
// changed.
type UpdateProjectInput struct {
	User      User          `json:"user"`
	ProjectId string        `json:"projectId"`
	Project   ProjectUpdate `json:"project"`
}

type UpdateProjectOutput struct {
	Project *Project `json:"project"`
}

// DeleteProjectInput soft-deletes a project: the document is retained and its
// IsActive flag is cleared, so resources that still carry its id keep a
// resolvable owner. The organisation's default project cannot be deleted.
type DeleteProjectInput struct {
	User      User   `json:"user"`
	ProjectId string `json:"projectId"`
}

type DeleteProjectOutput struct {
	Project *Project `json:"project"`
}
