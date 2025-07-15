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
	// Note: DeviceId is a unique identifier for the device, it can be used to identify the device in the system.
	DeviceId string `json:"deviceId" bson:"deviceId,omitempty"`

	// RBAC information
	// Note: SiteId is used to identify the site where the device is located.
	// GroupId is used to identify the group of devices.
	// OrganisationId is used to identify the organisation that owns the device.
	// FeaturePermissions is used to identify the permissions of the device, such as read, write, delete, etc.
	SiteId         string `json:"siteId" bson:"siteId,omitempty"`
	GroupId        string `json:"groupId" bson:"groupId,omitempty"`
	OrganisationId string `json:"organisationId" bson:"organisationId,omitempty"`

	// Marker information
	Timestamp   int64    `json:"timestamp,omitempty" bson:"timestamp,omitempty"`     // Timestamp of the marker
	Name        string   `json:"name,omitempty" bson:"name,omitempty"`               // Name of the marker
	Description string   `json:"description,omitempty" bson:"description,omitempty"` // Description of the marker
	Type        string   `json:"type,omitempty" bson:"type,omitempty"`               // Type of the marker, e.g., "alert", "event", etc.
	Comments    *Comment `json:"comments,omitempty" bson:"comments,omitempty"`       // Additional comments or description of the marker
	Tags        []string `json:"tags,omitempty" bson:"tags,omitempty"`               // Tags associated with the marker for categorization

	// Additional metadata
	MetaData *MarkerMetadata `json:"metaData,omitempty" bson:"metaData,omitempty"`

	// Synchronize
	Synchronize Synchronize `json:"synchronize,omitempty" bson:"synchronize,omitempty"` // Synchronization status with external systems

	// Audit information
	Audit *Audit `json:"audit,omitempty" bson:"audit,omitempty"`
}

type MarkerMetadata struct {
}
