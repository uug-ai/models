package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Comment struct {
	Id primitive.ObjectID `json:"id" bson:"_id,omitempty,omitempty"`

	// RBAC information
	// DeviceId is a unique identifier for the device, it can be used to identify the device in the system.
	// OrganisationId is used to identify the organisation that owns the device.
	DeviceId       string `json:"deviceId" bson:"deviceId,omitempty"` // device identifier
	OrganisationId string `json:"organisationId" bson:"organisationId,omitempty"`

	// Comment information
	Type       string  `json:"type" bson:"type,omitempty"`
	Author     string  `json:"author" bson:"author,omitempty"`
	Content    string  `json:"content" bson:"comment,omitempty"`
	Timestamps []int64 `json:"timestamps" bson:"timestamps,omitempty"`
	ParentId   string  `json:"parentId,omitempty" bson:"parentId,omitempty"` // ID of the parent comment, if this is a reply

	// Audit information
	Audit *Audit `json:"audit,omitempty" bson:"audit,omitempty"`
}
