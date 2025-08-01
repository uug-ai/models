package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Marker struct {
	Id primitive.ObjectID `json:"id" bson:"_id,omitempty,omitempty"`

	// Media information
	MediaIds []string `json:"mediaIds,omitempty" bson:"mediaIds,omitempty"` // MediaIds is used to link the marker to media items such as images or videos

	// Device information
	DeviceId string `json:"deviceId,omitempty" bson:"deviceId,omitempty"` // DeviceId is used to identify the device associated with the marker

	// RBAC information
	SiteId         string `json:"siteId,omitempty" bson:"siteId,omitempty"`                 // SiteId is used to identify the site where the marker is located
	GroupId        string `json:"groupId,omitempty" bson:"groupId,omitempty"`               // GroupId is used to identify the group of markers
	OrganisationId string `json:"organisationId,omitempty" bson:"organisationId,omitempty"` // OrganisationId is used to identify the organisation that owns the marker

	// Timing information
	StartTimestamp int64 `json:"startTimestamp,omitempty" bson:"startTimestamp,omitempty"` // Start timestamp of the marker in milliseconds since epoch
	EndTimestamp   int64 `json:"endTimestamp,omitempty" bson:"endTimestamp,omitempty"`     // End timestamp of the marker in milliseconds since epoch
	Duration       int64 `json:"duration,omitempty" bson:"duration,omitempty"`             // Duration of the marker in milliseconds

	Name        string `json:"name,omitempty" bson:"name,omitempty"`               // Name or identifier for the marker e.g., "a license plate (2-HCP-007), an unique identifier (transaction_id, point of sale), etc."
	Type        string `json:"type,omitempty" bson:"type,omitempty"`               // Type of the marker e.g., "alert", "event", "door_opened", "person", "car" etc.
	Description string `json:"description,omitempty" bson:"description,omitempty"` // Description of the marker

	// Additional metadata
	MetaData *MarkerMetadata `json:"metaData,omitempty" bson:"metaData,omitempty"` // Metadata associated with the marker, such as comments and tags

	// Synchronize
	Synchronize *Synchronize `json:"synchronize,omitempty" bson:"synchronize,omitempty"` // Synchronization status with external systems

	// Audit information
	Audit *Audit `json:"audit,omitempty" bson:"audit,omitempty"` // Audit information for tracking changes to the marker
}

type MarkerMetadata struct {
	Comments *Comment `json:"comments,omitempty" bson:"comments,omitempty"` // Additional comments or description of the marker
	Tags     []string `json:"tags,omitempty" bson:"tags,omitempty"`         // Tags associated with the marker for categorization
}
