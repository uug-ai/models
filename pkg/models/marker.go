package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Marker struct {
	Id primitive.ObjectID `json:"id" bson:"_id" example:"507f1f77bcf86cd799439011" required:"true"` // Unique identifier for the marker, generated automatically§§§

	// RBAC information
	DeviceId       string `json:"deviceId" bson:"deviceId" example:"686a906345c1df594939f9j25f4" required:"true"` // DeviceId is used to identify the device associated with the marker
	SiteId         string `json:"siteId,omitempty" bson:"siteId,omitempty" example:"686a906345c1df594pcsr3r45"`   // SiteId is used to identify the site where the marker is located
	GroupId        string `json:"groupId,omitempty" bson:"groupId,omitempty" example:"686a906345c1df594pmt41w4"`  // GroupId is used to identify the group of markers
	OrganisationId string `json:"organisationId" bson:"organisationId" example:"686a906345c1df594pad69f0"`        // OrganisationId is used to identify the organisation that owns the marker, retrieved from the user's access token

	// Timing information (all timestamps are in seconds)
	StartTimestamp int64 `json:"startTimestamp" bson:"startTimestamp" example:"1752482068" required:"true"` // Start timestamp of the marker in seconds since epoch
	EndTimestamp   int64 `json:"endTimestamp" bson:"endTimestamp" example:"1752482079" required:"true"`     // End timestamp of the marker in seconds since epoch
	Duration       int64 `json:"duration" bson:"duration" example:"11" required:"true"`                     // Duration of the marker in seconds

	Name        string        `json:"name" bson:"name" example:"2-HCP-007" required:"true"`                                        // Name or identifier for the marker e.g., "a license plate (2-HCP-007), an unique identifier (transaction_id, point of sale), etc."
	Events      []MarkerEvent `json:"events,omitempty" bson:"events,omitempty"`                                                    // Events associated with the marker, such as motion detected, sound detected, etc.
	Tags        []MarkerTag   `json:"tags,omitempty" bson:"tags,omitempty" example:"[\"vehicle\",\"license plate\",\"security\"]"` // Tags associated with the marker for categorization
	Description string        `json:"description,omitempty" bson:"description,omitempty" example:"Person forcably opened a door"`  // Description of the marker

	// Additional metadata
	Metadata *MarkerMetadata `json:"metadata,omitempty" bson:"metadata,omitempty"` // Metadata associated with the marker, such as comments and tags

	// AtRuntimeMetadata contains metadata that is generated at runtime, which can include
	// more verbose information about the device's current state, capabilities, or configuration.
	// for example the linked sites details, etc.
	AtRuntimeMetadata *MarkerAtRuntimeMetadata `json:"atRuntimeMetadata,omitempty" bson:"atRuntimeMetadata,omitempty"`

	// Synchronize
	Synchronize *Synchronize `json:"synchronize,omitempty" bson:"synchronize,omitempty"` // Synchronization status with external systems

	// Audit information
	Audit *Audit `json:"audit,omitempty" bson:"audit,omitempty"` // Audit information for tracking changes to the marker
}

type MarkerMetadata struct {
	Comments *Comment `json:"comments,omitempty" bson:"comments,omitempty"` // Additional comments or description of the marker
}

type MarkerAtRuntimeMetadata struct {
}

/* Marker options */
type MarkerOption struct {
	Value     string `bson:"value" json:"value"`
	Text      string `bson:"text" json:"text"`
	CreatedAt int64  `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	UpdatedAt int64  `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
}

type MarkerOptionTimeRange struct {
	Value     string `bson:"value" json:"value"`
	Text      string `bson:"text" json:"text"`
	Start     int64  `json:"start,omitempty" bson:"start,omitempty"`
	End       int64  `json:"end,omitempty" bson:"end,omitempty"`
	DeviceId  string `bson:"deviceId" json:"deviceId"` // Tags associated with the event for categorization
	CreatedAt int64  `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	UpdatedAt int64  `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
}

/* Marker Event */
type MarkerEvent struct { // Unique identifier for the event, generated automatically
	Timestamp   int64    `json:"timestamp" bson:"timestamp" example:"1752482070" required:"true"`                                // Timestamp of the event in seconds since epoch
	Name        string   `json:"name,omitempty" bson:"name,omitempty" example:"Motion Detected"`                                 // Name or identifier for the event e.g., "Motion Detected", "Sound Detected", etc.
	Description string   `json:"description,omitempty" bson:"description,omitempty" example:"Motion detected in the lobby area"` // Description of the event
	Tags        []string `json:"tags,omitempty" bson:"tags,omitempty" example:"[\"urgent\",\"review-needed\"]"`                  // Tags associated with the event for categorization
}

type MarkerEventOption struct {
	Value     string `bson:"value" json:"value"`
	Text      string `bson:"text" json:"text"`
	CreatedAt int64  `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	UpdatedAt int64  `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
}

type MarkerEventTimeRange struct {
	Value     string `bson:"value" json:"value"`
	Text      string `bson:"text" json:"text"`
	Start     int64  `json:"start,omitempty" bson:"start,omitempty"`
	End       int64  `json:"end,omitempty" bson:"end,omitempty"`
	DeviceId  string `bson:"deviceId" json:"deviceId"`
	CreatedAt int64  `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	UpdatedAt int64  `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
}

/* Marker Tag */
type MarkerTag struct {
	Name string `json:"name,omitempty" bson:"name,omitempty" example:"Motion Detected"`
}

type MarkerTagOption struct {
	Value     string `bson:"value" json:"value"`
	Text      string `bson:"text" json:"text"`
	CreatedAt int64  `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	UpdatedAt int64  `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
}

type MarkerTagTimeRange struct {
	Value     string `bson:"value" json:"value"`
	Text      string `bson:"text" json:"text"`
	Start     int64  `json:"start,omitempty" bson:"start,omitempty"`
	End       int64  `json:"end,omitempty" bson:"end,omitempty"`
	DeviceId  string `bson:"deviceId" json:"deviceId"`
	CreatedAt int64  `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	UpdatedAt int64  `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
}
