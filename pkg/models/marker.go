package models

import "go.mongodb.org/mongo-driver/bson/primitive"

const (
	MARKER_BINDING_FAILED string = "Failed to bind marker from request body"
	MARKER_NAME_EXISTS    string = "Marker with the same name already exists"
	MARKER_MISSING_INFO   string = "Marker is missing required information"
	MARKER_FOUND          string = "One or more markers where found"
	MARKER_NOT_FOUND      string = "One or more markers not found, returning empty list"
	MARKER_ADD_SUCCESS    string = "Marker added successfully"
	MARKER_ADD_FAILED     string = "Failed to add marker"
	MARKER_UPDATE_SUCCESS string = "Marker updated successfully"
	MARKER_UPDATE_FAILED  string = "Failed to update marker"
	MARKER_DELETE_SUCCESS string = "Marker deleted successfully"
	MARKER_DELETE_FAILED  string = "Failed to delete marker"
)

type Marker struct {
	Id primitive.ObjectID `json:"id" bson:"_id,omitempty,omitempty"`

	// Device information
	DeviceId string `json:"deviceId" bson:"deviceId,omitempty"` // DeviceId is used to identify the device associated with the marker

	// RBAC information
	SiteId         string `json:"siteId" bson:"siteId,omitempty"`                 // SiteId is used to identify the site where the marker is located
	GroupId        string `json:"groupId" bson:"groupId,omitempty"`               // GroupId is used to identify the group of markers
	OrganisationId string `json:"organisationId" bson:"organisationId,omitempty"` // OrganisationId is used to identify the organisation that owns the marker

	// Timing information
	StartTimestamp int64 `json:"startTimestamp,omitempty" bson:"startTimestamp,omitempty"` // Start timestamp of the marker in milliseconds since epoch
	EndTimestamp   int64 `json:"endTimestamp,omitempty" bson:"endTimestamp,omitempty"`     // End timestamp of the marker in milliseconds since epoch
	Duration       int64 `json:"duration,omitempty" bson:"duration,omitempty"`             // Duration of the marker in milliseconds

	Name        string `json:"name,omitempty" bson:"name,omitempty"`               // Name of the marker
	Description string `json:"description,omitempty" bson:"description,omitempty"` // Description of the marker
	Type        string `json:"type,omitempty" bson:"type,omitempty"`               // Type of the marker, e.g., "alert", "event", etc.

	// Additional metadata
	MetaData *MarkerMetadata `json:"metaData,omitempty" bson:"metaData,omitempty"` // Metadata associated with the marker, such as comments and tags

	// Synchronize
	Synchronize Synchronize `json:"synchronize,omitempty" bson:"synchronize,omitempty"` // Synchronization status with external systems

	// Audit information
	Audit *Audit `json:"audit,omitempty" bson:"audit,omitempty"` // Audit information for tracking changes to the marker
}

type MarkerMetadata struct {
	Comments *Comment `json:"comments,omitempty" bson:"comments,omitempty"` // Additional comments or description of the marker
	Tags     []string `json:"tags,omitempty" bson:"tags,omitempty"`         // Tags associated with the marker for categorization
}
