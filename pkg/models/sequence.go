package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Sequence struct {
	Id             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	OrganisationId string             `json:"organisationId" bson:"organisationId,omitempty"`

	// ProjectId optionally places the sequence in a project within its organisation.
	// A nil value keeps the sequence organisation-wide.
	ProjectId *primitive.ObjectID `json:"projectId,omitempty" bson:"projectId,omitempty"`

	Start    int64                  `json:"start,omitempty"`
	End      int64                  `json:"end,omitempty"`
	UserId   string                 `json:"user_id" bson:"user_id,omitempty"`
	Images   []Media                `json:"images,omitempty"`
	Analysis map[string]interface{} `bson:"analysis,omitempty"`
	Devices  []string               `json:"devices,omitempty"`
	Notified bool                   `json:"notified,omitempty"`
}
