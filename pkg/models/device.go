package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// --------------------------------------------------------------------
// Data model for device
//
// This is the main struct for Device

type Device struct {
	Id primitive.ObjectID `json:"id" bson:"_id,omitempty,omitempty"`

	// Device information
	// Note: DeviceId is a unique identifier for the device, it can be used to identify the device in the system.
	// DeviceType is used to identify the type of device, such as camera, sensor, access control, etc.
	// Version and ReleaseHash are used to identify the version of the device software.
	// Deployment is used to identify the deployment type of the device, such as factory, docker, docker compose, kubernetes, etc.
	Name        string `json:"name" bson:"name,omitempty"` // e.g. "Front Door Camera"
	DeviceId    string `json:"deviceId" bson:"deviceId,omitempty"`
	DeviceType  string `json:"deviceType" bson:"deviceType,omitempty"` // e.g. "camera", "sensor", "access_control"
	Version     string `json:"version" bson:"version,omitempty"`
	ReleaseHash string `json:"releaseHash" bson:"releaseHash,omitempty"` // e.g. "v1.0.0-abcdef123456"
	Deployment  string `json:"deployment" bson:"deployment,omitempty"`   // e.g. "factory", "docker", "docker compose", "kubernetes"

	// RBAC information
	// Note: SiteId is used to identify the site where the device is located.
	// GroupId is used to identify the group of devices.
	// OrganisationId is used to identify the organisation that owns the device.
	// FeaturePermissions is used to identify the permissions of the device, such as read, write, delete, etc.
	SiteId         string `json:"siteId" bson:"siteId,omitempty"`
	GroupId        string `json:"groupId" bson:"groupId,omitempty"`
	OrganisationId string `json:"organisationId" bson:"organisationId,omitempty"`

	// FeaturePermissions is used to identify the permissions of the device, such as read, write, delete, etc.
	// It is a map of feature names to permissions.
	// For example, "camera" can have permissions like "read", "write", "delete", etc.
	// This allows for fine-grained control over what features are accessible by users or groups.
	// FeaturePermissions can be used to implement Role-Based Access Control (RBAC) for devices.
	FeaturePermissions FeaturePermissions `json:"featurePermissions" bson:"featurePermissions"`

	// Device status
	// Note: Status is used to identify the status of the device, such as online, offline, maintenance, etc.
	// LastSeenTimestamp is used to identify the last time the device was seen online.
	Status            string `json:"status" bson:"status,omitempty"`                       // e.g. "connected", "idle"
	LastSeenTimestamp int64  `json:"lastSeenTimestamp" bson:"lastSeenTimestamp,omitempty"` //

	// Metadata
	Metadata []DeviceMetadata `json:"metadata,omitempty" bson:"metadata,omitempty"`

	// Camera Metadata
	CameraMetadata CameraMetadata `json:"cameraMetadata,omitempty" bson:"cameraMetadata,omitempty"`

	// Location metadata
	LocationMetadata LocationMetadata `json:"locationMetadata,omitempty" bson:"locationMetadata,omitempty"`
}

// We can store additional metadata for media files, such as tags and classifications.
type DeviceMetadata struct {
	Mute             int64  `json:"mute" bson:"mute,omitempty"`
	Color            string `json:"color" bson:"color,omitempty"`
	Brand            string `json:"brand" bson:"brand,omitempty"`
	Model            string `json:"model" bson:"model,omitempty"`
	Description      string `json:"description" bson:"description,omitempty"`
	LastMaintenance  int64  `json:"lastMaintenance" bson:"lastMaintenance,omitempty"`
	InstallationDate int64  `json:"installationDate" bson:"installationDate,omitempty"`
	Location         string `json:"location" bson:"location,omitempty"` // e.g. "Front Door", "Back Door", "Living Room"
}

// CameraMetadata contains metadata specific to camera devices.
type CameraMetadata struct {
	Resolution string   `json:"resolution" bson:"resolution,omitempty"`
	FrameRate  int64    `json:"frameRate" bson:"frameRate,omitempty"`
	Bitrate    int64    `json:"bitrate" bson:"bitrate,omitempty"`
	Codec      string   `json:"codec" bson:"codec,omitempty"`
	Presets    []Preset `json:"presets" bson:"presets,omitempty"`
	Tours      []Tour   `json:"tours" bson:"tours,omitempty"`
}

// LocationMetadata contains metadata about the physical location of the device.
type LocationMetadata struct {
	Latitude     float64  `json:"latitude" bson:"latitude,omitempty"`
	Longitude    float64  `json:"longitude" bson:"longitude,omitempty"`
	Altitude     float64  `json:"altitude" bson:"altitude,omitempty"`
	Address      string   `json:"address" bson:"address,omitempty"` // e.g. "123 Main St, Anytown, USA"
	OnFloorPlans []string `json:"onFloorPlans" bson:"onFloorPlans,omitempty"`
	FieldOfView  float64  `json:"fieldOfView" bson:"fieldOfView,omitempty"`
}

// FeaturePermissions is a map of feature names to permissions.
type FeaturePermissions struct {
	PTZ          int `json:"ptz" bson:"ptz"`
	Liveview     int `json:"liveview" bson:"liveview"`
	RemoteConfig int `json:"remoteConfig" bson:"remoteConfig"`
	IO           int `json:"io" bson:"io"`
	FloorPlans   int `json:"floorPlans" bson:"floorPlans"`
	// Talk..
	// ...
}
