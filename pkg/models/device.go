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
	Key  string `json:"key" bson:"key,omitempty"`
	Name string `json:"name" bson:"name,omitempty"` // e.g. "Front Door Camera"
	Type string `json:"type" bson:"type,omitempty"` // Type of device e.g. "camera", "sensor", "access_control"

	// Versioning information
	// Note: Version is used to identify the version of the device software.
	// ReleaseHash is used to identify the release hash of the device software, it can be used to identify the specific build of the device software.
	// Deployment is used to identify the deployment type of the device, such as factory, docker, docker compose, kubernetes, etc.
	Repository string `json:"repository" bson:"repository,omitempty"` // Repository URL of the agent codebase
	Version    string `json:"version" bson:"version,omitempty"`       // Version of the agent, injected on build. Reflects the release tag of the agent.
	Deployment string `json:"deployment" bson:"deployment,omitempty"` // Type of deployment used e.g. "factory", "docker", "docker compose", "kubernetes"

	// RBAC information
	// Note: SiteId is used to identify the site where the device is located.
	// GroupId is used to identify the group of devices.
	// OrganisationId is used to identify the organisation that owns the device.
	// FeaturePermissions is used to identify the permissions of the device, such as read, write, delete, etc.
	OrganisationId string   `json:"organisationId" bson:"organisationId,omitempty"`
	SiteIds        []string `json:"siteIds" bson:"siteIds,omitempty"`
	GroupIds       []string `json:"groupIds" bson:"groupIds,omitempty"`

	// @LEGACY FIELDS - to be removed in future versions
	UserId string `json:"user_id" bson:"user_id,omitempty"`

	// Latest Media information
	LatestMedia          string `json:"latestMedia" bson:"latestMedia,omitempty"`                   // ID of the latest media captured by the device
	LatestMediaTimestamp int64  `json:"latestMediaTimestamp" bson:"latestMediaTimestamp,omitempty"` // Timestamp of the latest media captured by the device (milliseconds since epoch)

	// Device status
	// Note: Status is used to identify the status of the device, such as online, offline, maintenance, etc."
	// LastSeenTimestamp is used to identify the last time the device was seen online.
	ConnectionStart       int64 `json:"connectionStart" bson:"connectionStart,omitempty"`             // timestamp in milliseconds when the device last connected e.g. after a reboot
	DeviceLastSeen        int64 `json:"deviceLastSeen" bson:"deviceLastSeen,omitempty"`               // last time the device itself reported being online (timestamp in milliseconds)
	AgentLastSeen         int64 `json:"agentLastSeen" bson:"agentLastSeen,omitempty"`                 // last time the agent reported being online (timestamp in milliseconds)
	IdleThreshold         int64 `json:"idleThreshold" bson:"idleThreshold,omitempty"`                 // threshold in milliseconds to consider a device idle, based on deviceLastSeen
	DisconnectedThreshold int64 `json:"disconnectedThreshold" bson:"disconnectedThreshold,omitempty"` // threshold in milliseconds to consider a device offline, based on agentLastSeen
	HealthyThreshold      int64 `json:"healthyThreshold" bson:"healthyThreshold,omitempty"`           // threshold in milliseconds to consider a device healthy, based on agentLastSeen

	// Location of the device, not real time postion. e.g. "Office 1st Floor", "Lobby", "Kilian's Car", etc.
	Location *Location `json:"location" bson:"location,omitempty"`

	// Metadata
	Metadata *DeviceMetadata `json:"metadata,omitempty" bson:"metadata,omitempty"`

	// Device Specific Metadata
	CameraMetadata         *CameraMetadata         `json:"cameraMetadata,omitempty" bson:"cameraMetadata,omitempty"`
	LightSensorMetadata    *LightSensorMetadata    `json:"lightSensorMetadata,omitempty" bson:"lightSensorMetadata,omitempty"`
	HumiditySensorMetadata *HumiditySensorMetadata `json:"humiditySensorMetadata,omitempty" bson:"humiditySensorMetadata,omitempty"`
	GPSMetadata            *GPSMetadata            `json:"gpsMetadata,omitempty" bson:"gpsMetadata,omitempty"`
	// ... other device specific metadata structs

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

	// Audit information
	Audit *Audit `json:"audit,omitempty" bson:"audit,omitempty"`
}

// We can store additional metadata for media files, such as tags and classifications.
type DeviceMetadata struct {
	Hidden           bool     `json:"hidden" bson:"hidden,omitempty"`                     // Indicates if the device is hidden from UI
	Color            string   `json:"color" bson:"color,omitempty"`                       // e.g. "#FF5733" (hex color code)
	Brand            string   `json:"brand" bson:"brand,omitempty"`                       // e.g. "Nest", "Ring"
	Model            string   `json:"model" bson:"model,omitempty"`                       // e.g. "Nest Cam", "Ring Doorbell"
	Description      string   `json:"description" bson:"description,omitempty"`           // e.g. "Outdoor camera with night vision"
	LastMaintenance  int64    `json:"lastMaintenance" bson:"lastMaintenance,omitempty"`   // Last maintenance date in milliseconds since epoch
	InstallationDate int64    `json:"installationDate" bson:"installationDate,omitempty"` // Installation date in milliseconds since epoch
	OnFloorPlans     []string `json:"onFloorPlans" bson:"onFloorPlans,omitempty"`         // Floor plans this device is associated with
	IPAddress        string   `json:"ipAddress" bson:"ipAddress,omitempty"`               // e.g. "192.168.1.1"
	MacAddress       string   `json:"macAddress" bson:"macAddress,omitempty"`             // e.g. "00:1A:2B:3C:4D:5E"
	BootTime         string   `json:"bootTime" bson:"bootTime,omitempty"`
	Architecture     string   `json:"architecture" bson:"architecture,omitempty"`
	Hostname         string   `json:"hostname" bson:"hostname,omitempty"`
	FreeMemory       string   `json:"freeMemory" bson:"freeMemory,omitempty"`
	TotalMemory      string   `json:"totalMemory" bson:"totalMemory,omitempty"`
	UsedMemory       string   `json:"usedMemory" bson:"usedMemory,omitempty"`
	ProcessMemory    string   `json:"processMemory" bson:"processMemory,omitempty"`
	Encrypted        bool     `json:"encrypted" bson:"encrypted,omitempty"`
	EncryptedData    []byte   `json:"encryptedData" bson:"encryptedData,omitempty"`
	HubEncryption    string   `json:"hubEncryption" bson:"hubEncryption,omitempty"`
	E2EEncryption    string   `json:"e2eEncryption" bson:"e2eEncryption,omitempty"`
}

// CameraMetadata contains metadata specific to camera devices.
type CameraMetadata struct {
	Resolution  string  `json:"resolution" bson:"resolution,omitempty"` // e.g. "1920x1080", "1280x720"
	FieldOfView float64 `json:"fieldOfView" bson:"fieldOfView,omitempty"`
	FrameRate   int64   `json:"frameRate" bson:"frameRate,omitempty"` // Frame rate in fps
	Bitrate     int64   `json:"bitrate" bson:"bitrate,omitempty"`     // Bitrate in kbps
	Codec       string  `json:"codec" bson:"codec,omitempty"`         // e.g. "H.264", "H.265"

	// Camera ONVIF and PTZ capabilities
	HasOnvif       bool           `json:"hasOnvif" bson:"hasOnvif,omitempty"`             // Indicates if the camera supports ONVIF protocol
	HasAudio       bool           `json:"hasAudio" bson:"hasAudio,omitempty"`             // Indicates if the camera supports audio
	HasZoom        bool           `json:"hasZoom" bson:"hasZoom,omitempty"`               // Indicates if the camera supports zoom functionality
	HasBackChannel bool           `json:"hasBackChannel" bson:"hasBackChannel,omitempty"` // Indicates if the camera supports backchannel audio
	HasPanTilt     bool           `json:"hasPanTilt" bson:"hasPanTilt,omitempty"`         // Indicates if the camera supports pan and tilt functionality
	HasPresets     bool           `json:"hasPresets" bson:"hasPresets,omitempty"`         // Indicates if the camera supports presets
	HasIO          bool           `json:"hasIO" bson:"hasIO,omitempty"`                   // Indicates if the camera has input/output capabilities
	Presets        []CameraPreset `json:"presets" bson:"presets,omitempty"`               // Presets for the camera, used for quick positioning
	Tours          []CameraTour   `json:"tours" bson:"tours,omitempty"`                   // Tours for the camera, used for automated movements through presets
	IOs            []IO           `json:"ios" bson:"ios,omitempty"`
}

type LightSensorMetadata struct {
	// Sensitivity indicates the sensitivity level of the light sensor (e.g. 1-100)
	Sensitivity int `json:"sensitivity" bson:"sensitivity,omitempty"`
	// LuxRange specifies the measurable range in lux (e.g. "0-10000")
	LuxRange string `json:"luxRange" bson:"luxRange,omitempty"`
	// CurrentLux is the latest measured light level in lux
	CurrentLux float64 `json:"currentLux" bson:"currentLux,omitempty"`
	// Unit is the measurement unit, typically "lux"
	Unit string `json:"unit" bson:"unit,omitempty"`
}

// HumiditySensorMetadata contains metadata specific to humidity sensor devices.
type HumiditySensorMetadata struct {
	// Sensitivity indicates the sensitivity level of the humidity sensor (e.g. 1-100)
	Sensitivity int `json:"sensitivity" bson:"sensitivity,omitempty"`
	// HumidityRange specifies the measurable range in percentage (e.g. "0-100")
	HumidityRange string `json:"humidityRange" bson:"humidityRange,omitempty"`
	// CurrentHumidity is the latest measured humidity level in percentage
	CurrentHumidity float64 `json:"currentHumidity" bson:"currentHumidity,omitempty"`
	// Unit is the measurement unit, typically "%"
	Unit string `json:"unit" bson:"unit,omitempty"`
}

// GPSMetadata contains metadata specific to GPS devices.
type GPSMetadata struct {
	// Latitude is the current latitude of the device.
	Latitude float64 `json:"latitude" bson:"latitude,omitempty"`
	// Longitude is the current longitude of the device.
	Longitude float64 `json:"longitude" bson:"longitude,omitempty"`
	// Altitude is the current altitude of the device in meters.
	Altitude float64 `json:"altitude" bson:"altitude,omitempty"`
	// Accuracy is the accuracy of the GPS reading in meters.
	Accuracy float64 `json:"accuracy" bson:"accuracy,omitempty"`
	// Timestamp is the time when the GPS data was recorded (milliseconds since epoch).
	Timestamp int64 `json:"timestamp" bson:"timestamp,omitempty"`
	// Speed is the current speed of the device in meters per second.
	Speed float64 `json:"speed" bson:"speed,omitempty"`
	// Heading is the direction of movement in degrees.
	Heading float64 `json:"heading" bson:"heading,omitempty"`
}

// DeviceAtRuntimeMetadata contains metadata that is generated at runtime, which can include
// more verbose information about the device's current state, capabilities, or configuration.
type DeviceAtRuntimeMetadata struct {
	// LinkedSites contains details about the sites that the device is linked to.
	//Sites []Sites `json:"sites,omitempty" bson:"sites,omitempty"`
	// LinkedGroups contains details about the groups that the device is linked to.
	//Groups []Groups `json:"groups,omitempty" bson:"groups,omitempty"`

	// Status is derived from the last seen timestamps and thresholds.
	Status string `json:"status" bson:"status,omitempty"` // e.g. "connected", "idle", "disconnected", "healthy"
}

type DeviceKey struct {
	Key  string `json:"key" bson:"key,omitempty"`
	Name string `json:"name" bson:"name,omitempty"`
}

type DeviceOption struct {
	Id    string `bson:"id" json:"id"`
	Value string `bson:"value" json:"value"`
	Text  string `bson:"text" json:"text"`
}

type CameraPreset struct {
	Name  string  `json:"name" bson:"name"`
	Token string  `json:"token" bson:"token"`
	X     float64 `json:"x" bson:"x"`
	Y     float64 `json:"y" bson:"y"`
	Z     float64 `json:"z" bson:"z"`
}

type CameraTour struct {
	Name    string         `json:"name" bson:"name,omitempty"`
	Presets []CameraPreset `json:"presets" bson:"presets,omitempty"`
	Current int            `json:"current" bson:"current,omitempty"`
	Running bool           `json:"running" bson:"running,omitempty"`
	Loop    bool           `json:"loop" bson:"loop,omitempty"`
	Speed   float64        `json:"speed" bson:"speed,omitempty"`
}
