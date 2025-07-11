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
	Name       string `json:"name" bson:"name,omitempty"` // e.g. "Front Door Camera"
	DeviceId   string `json:"deviceId" bson:"deviceId,omitempty"`
	DeviceType string `json:"deviceType" bson:"deviceType,omitempty"` // e.g. "camera", "sensor", "access_control"

	// Versioning information
	// Note: Version is used to identify the version of the device software.
	// ReleaseHash is used to identify the release hash of the device software, it can be used to identify the specific build of the device software.
	// Deployment is used to identify the deployment type of the device, such as factory, docker, docker compose, kubernetes, etc.
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

	// Device status
	// Note: Status is used to identify the status of the device, such as online, offline, maintenance, etc.
	// LastSeenTimestamp is used to identify the last time the device was seen online.
	Status            string `json:"status" bson:"status,omitempty"`                       // e.g. "connected", "idle"
	LastSeenTimestamp int64  `json:"lastSeenTimestamp" bson:"lastSeenTimestamp,omitempty"` // last time the device was seen online (timestamp in milliseconds)

	// Metadata
	Metadata *DeviceMetadata `json:"metadata,omitempty" bson:"metadata,omitempty"`

	// Camera Metadata
	CameraMetadata *DeviceCameraMetadata `json:"cameraMetadata,omitempty" bson:"cameraMetadata,omitempty"`

	// Location metadata
	LocationMetadata *DeviceLocationMetadata `json:"locationMetadata,omitempty" bson:"locationMetadata,omitempty"`

	// FeaturePermissions is used to identify the permissions of the device, such as read, write, delete, etc.
	// It is a map of feature names to permissions.
	// For example, "camera" can have permissions like "read", "write", "delete", etc.
	// This allows for fine-grained control over what features are accessible by users or groups.
	// FeaturePermissions can be used to implement Role-Based Access Control (RBAC) for devices.
	FeaturePermissions *DeviceFeaturePermissions `json:"featurePermissions" bson:"featurePermissions"`

	// AtRuntimeMetadata contains metadata that is generated at runtime, which can include
	// more verbose information about the device's current state, capabilities, or configuration.
	// for example the linked sites details, etc.
	AtRuntimeMetadata *DeviceAtRuntimeMetadata `json:"atRuntimeMetadata,omitempty" bson:"atRuntimeMetadata,omitempty"`
}

// We can store additional metadata for media files, such as tags and classifications.
type DeviceMetadata struct {
	Mute             bool   `json:"mute" bson:"mute,omitempty"`                         // Mute status, e.g. false for unmuted, true for muted
	Color            string `json:"color" bson:"color,omitempty"`                       // e.g. "#FF5733" (hex color code)
	Brand            string `json:"brand" bson:"brand,omitempty"`                       // e.g. "Nest", "Ring"
	Model            string `json:"model" bson:"model,omitempty"`                       // e.g. "Nest Cam", "Ring Doorbell"
	Description      string `json:"description" bson:"description,omitempty"`           // e.g. "Outdoor camera with night vision"
	LastMaintenance  int64  `json:"lastMaintenance" bson:"lastMaintenance,omitempty"`   // Last maintenance date in milliseconds since epoch
	InstallationDate int64  `json:"installationDate" bson:"installationDate,omitempty"` // Installation date in milliseconds since epoch
}

// CameraMetadata contains metadata specific to camera devices.
type DeviceCameraMetadata struct {
	Resolution string   `json:"resolution" bson:"resolution,omitempty"` // e.g. "1920x1080", "1280x720"
	FrameRate  int64    `json:"frameRate" bson:"frameRate,omitempty"`   // Frame rate in fps
	Bitrate    int64    `json:"bitrate" bson:"bitrate,omitempty"`       // Bitrate in kbps
	Codec      string   `json:"codec" bson:"codec,omitempty"`           // e.g. "H.264", "H.265"
	HasOnvif   bool     `json:"hasOnvif" bson:"hasOnvif,omitempty"`     // Indicates if the camera supports ONVIF protocol
	HasAudio   bool     `json:"hasAudio" bson:"hasAudio,omitempty"`     // Indicates if the camera supports audio
	HasZoom    bool     `json:"hasZoom" bson:"hasZoom,omitempty"`       // Indicates if the camera supports zoom functionality
	HasPanTilt bool     `json:"hasPanTilt" bson:"hasPanTilt,omitempty"` // Indicates if the camera supports pan and tilt functionality
	HasPresets bool     `json:"hasPresets" bson:"hasPresets,omitempty"` // Indicates if the camera supports presets
	Presets    []Preset `json:"presets" bson:"presets,omitempty"`       // Presets for the camera, used for quick positioning
	Tours      []Tour   `json:"tours" bson:"tours,omitempty"`           // Tours for the camera, used for automated movements through presets
	HasIO      bool     `json:"hasIO" bson:"hasIO,omitempty"`           // Indicates if the camera has input/output capabilities
	IOs        []IO     `json:"ios" bson:"ios,omitempty"`               // Input/Output capabilities of the camera (such as alarms, relays, etc.)
}

// LocationMetadata contains metadata about the physical location of the device.
type DeviceLocationMetadata struct {
	Latitude     float64  `json:"latitude" bson:"latitude,omitempty"`
	Longitude    float64  `json:"longitude" bson:"longitude,omitempty"`
	Altitude     float64  `json:"altitude" bson:"altitude,omitempty"`
	Address      string   `json:"address" bson:"address,omitempty"` // e.g. "123 Main St, Anytown, USA"
	OnFloorPlans []string `json:"onFloorPlans" bson:"onFloorPlans,omitempty"`
	FieldOfView  float64  `json:"fieldOfView" bson:"fieldOfView,omitempty"`
}

// FeaturePermissions is a map of feature names to permissions.
type DeviceFeaturePermissions struct {
	PTZ          int `json:"ptz" bson:"ptz"`
	Liveview     int `json:"liveview" bson:"liveview"`
	RemoteConfig int `json:"remoteConfig" bson:"remoteConfig"`
	IO           int `json:"io" bson:"io"`
	FloorPlans   int `json:"floorPlans" bson:"floorPlans"`
	// Talk..
	// ...
}

// DeviceAtRuntimeMetadata contains metadata that is generated at runtime, which can include
// more verbose information about the device's current state, capabilities, or configuration.
type DeviceAtRuntimeMetadata struct {
	// LinkedSites contains details about the sites that the device is linked to.
	Sites []Sites `json:"sites,omitempty" bson:"sites,omitempty"`
	// LinkedGroups contains details about the groups that the device is linked to.
	Groups []Groups `json:"groups,omitempty" bson:"groups,omitempty"`
}
