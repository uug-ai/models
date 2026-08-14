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
