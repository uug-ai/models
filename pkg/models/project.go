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
type GetProjectsInput struct {
	User User `json:"user"`
}

type GetProjectsOutput struct {
	Projects []Project `json:"projects"`
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
