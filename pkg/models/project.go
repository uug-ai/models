package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// Project is an optional access boundary within an organisation. Resources
// remain organisation-wide unless they are assigned a ProjectResourceScope.
type Project struct {
	Id             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	OrganisationId primitive.ObjectID `json:"organisationId" bson:"organisationId"`
	Name           string             `json:"name" bson:"name"`
	Slug           string             `json:"slug" bson:"slug"`
	Description    string             `json:"description,omitempty" bson:"description,omitempty"`
	IsActive       bool               `json:"isActive" bson:"isActive"`
	Audit          Audit              `json:"audit" bson:"audit"`
}

// ProjectResourceScope optionally places a resource in one owning project and
// shares it with additional projects. A nil ProjectId keeps the resource at
// organisation scope; SharedWithProjectIds must then be empty.
type ProjectResourceScope struct {
	ProjectId            *primitive.ObjectID  `json:"projectId,omitempty" bson:"projectId,omitempty"`
	SharedWithProjectIds []primitive.ObjectID `json:"sharedWithProjectIds,omitempty" bson:"sharedWithProjectIds,omitempty"`
}

// IsOrganisationWide reports whether the resource remains at organisation scope.
func (s ProjectResourceScope) IsOrganisationWide() bool {
	return s.ProjectId == nil
}
