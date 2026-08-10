package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type LabelWrapper struct {
	Label Label `json:"label" bson:"label"`
}

type Label struct {
	Id             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	OrganisationId string             `json:"organisationId" bson:"organisationId,omitempty"`
	Name           string             `json:"name" bson:"name,omitempty"`
	Description    string             `json:"description" bson:"description,omitempty"`
	Color          string             `json:"color" bson:"color,omitempty"`
	IsPrivate      bool               `json:"is_private" bson:"is_private"`
	Types          []string           `json:"type" bson:"type,omitempty"`

	UserId    string `json:"user_id" bson:"user_id,omitempty"`
	OwnerId   string `json:"owner_id" bson:"owner_id,omitempty"`
	CreatedAt int64  `json:"created_at" bson:"created_at,omitempty"`

	Audit *Audit `json:"audit,omitempty" bson:"audit,omitempty"`
}
