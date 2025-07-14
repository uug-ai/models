package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Marker struct {
	Id primitive.ObjectID `json:"id" bson:"_id,omitempty,omitempty"`

	// Device information
	// Note: DeviceId is a unique identifier for the device, it can be used to identify the device in the system.
	// DeviceType is used to identify the type of device, such as camera, sensor, access control, etc.
	DeviceId   string `json:"deviceId" bson:"deviceId,omitempty"`
	DeviceType string `json:"deviceType" bson:"deviceType,omitempty"` // e.g. "camera", "sensor", "access_control"

	// RBAC information
	// Note: SiteId is used to identify the site where the device is located.
	// GroupId is used to identify the group of devices.
	// OrganisationId is used to identify the organisation that owns the device.
	// FeaturePermissions is used to identify the permissions of the device, such as read, write, delete, etc.
	SiteId         string `json:"siteId" bson:"siteId,omitempty"`
	GroupId        string `json:"groupId" bson:"groupId,omitempty"`
	OrganisationId string `json:"organisationId" bson:"organisationId,omitempty"`

	// Marker information
	Timestamp  int64  `json:"timestamp,omitempty" bson:"timestamp,omitempty"` // Timestamp of the marker
	MarkerId   string `json:"markerId" bson:"markerId,omitempty"`             // unique identifier for the marker
	MarkerType string `json:"markerType" bson:"markerType,omitempty"`         // e.g. "event", "alert", "notification"
}
