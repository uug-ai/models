package models

import "go.mongodb.org/mongo-driver/bson/primitive"

const (
	MARKER_FOUND           string = "One or more markers where found."
	MARKER_NOT_FOUND       string = "One or more markers not found, returning empty list."
	MARKER_ADDED_SUCCESS   string = "Marker added successfully."
	MARKER_ADDED_FAILED    string = "Failed to add marker, please try again."
	MARKER_UPDATED_SUCCESS string = "Marker updated successfully."
	MARKER_UPDATED_FAILED  string = "Failed to update marker, please try again."
	MARKER_DELETED_SUCCESS string = "Marker deleted successfully."
	MARKER_DELETED_FAILED  string = "Failed to delete marker, please try again."
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
	Timestamp  int64    `json:"timestamp,omitempty" bson:"timestamp,omitempty"` // Timestamp of the marker
	MarkerId   string   `json:"markerId" bson:"markerId,omitempty"`             // unique identifier for the marker
	MarkerType string   `json:"markerType" bson:"markerType,omitempty"`         // e.g. "event", "alert", "notification"
	Comments   *Comment `json:"comments,omitempty" bson:"comments,omitempty"`   // Additional comments or description of the marker
	Tags       []string `json:"tags,omitempty" bson:"tags,omitempty"`           // Tags associated with the marker for categorization

	// Audit information
	Audit *Audit `json:"audit,omitempty" bson:"audit,omitempty"`
}
