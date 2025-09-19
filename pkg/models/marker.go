package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Marker struct {
	Id primitive.ObjectID `json:"id" bson:"_id" example:"507f1f77bcf86cd799439011" required:"true"` // Unique identifier for the marker, generated automatically§§§

	// Media information
	MediaIds []string `json:"mediaIds" bson:"mediaIds" example:"[\"img_20230918_001.jpg\", \"vid_20230918_002.mp4\"]" required:"true"` // MediaIds is used to link the marker to media items such as images or videos

	// Device information
	DeviceId string `json:"deviceId" bson:"deviceId" example:"686a906345c1df594939f9j25f4" required:"true"` // DeviceId is used to identify the device associated with the marker

	// RBAC information
	SiteId         string `json:"siteId,omitempty" bson:"siteId,omitempty" example:"686a906345c1df594pcsr3r45"`  // SiteId is used to identify the site where the marker is located
	GroupId        string `json:"groupId,omitempty" bson:"groupId,omitempty" example:"686a906345c1df594pmt41w4"` // GroupId is used to identify the group of markers
	OrganisationId string `json:"organisationId" bson:"organisationId" example:"686a906345c1df594pad69f0"`       // OrganisationId is used to identify the organisation that owns the marker, retrieved from the user's access token

	// Timing information (all timestamps are in seconds)
	StartTimestamp int64 `json:"startTimestamp" bson:"startTimestamp" example:"1752482068" required:"true"` // Start timestamp of the marker in seconds since epoch
	EndTimestamp   int64 `json:"endTimestamp" bson:"endTimestamp" example:"1752482079" required:"true"`     // End timestamp of the marker in seconds since epoch
	Duration       int64 `json:"duration" bson:"duration" example:"11" required:"true"`                     // Duration of the marker in seconds

	Name        string `json:"name" bson:"name" example:"2-HCP-007" required:"true"`                                       // Name or identifier for the marker e.g., "a license plate (2-HCP-007), an unique identifier (transaction_id, point of sale), etc."
	Type        string `json:"type,omitempty" bson:"type,omitempty" example:"door-forced"`                                 // Type of the marker e.g., "alert", "event", "door_opened", "person", etc.
	Description string `json:"description,omitempty" bson:"description,omitempty" example:"Person forcably opened a door"` // Description of the marker

	// Additional metadata
	MetaData *MarkerMetadata `json:"metaData,omitempty" bson:"metaData,omitempty"` // Metadata associated with the marker, such as comments and tags

	// Synchronize
	Synchronize *Synchronize `json:"synchronize,omitempty" bson:"synchronize,omitempty"` // Synchronization status with external systems

	// Audit information
	Audit *Audit `json:"audit,omitempty" bson:"audit,omitempty"` // Audit information for tracking changes to the marker
}

type MarkerMetadata struct {
	Comments *Comment `json:"comments,omitempty" bson:"comments,omitempty"`                                                // Additional comments or description of the marker
	Tags     []string `json:"tags,omitempty" bson:"tags,omitempty" example:"[\"vehicle\",\"license plate\",\"security\"]"` // Tags associated with the marker for categorization
}
