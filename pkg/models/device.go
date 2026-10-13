package models

import (
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DeviceSchemaVersion is the current shape of the device document. Writers
// stamp it; the backfill upgrades documents below it.
const DeviceSchemaVersion = 2

// --------------------------------------------------------------------
// Device type, capabilities and status

// DeviceType is the device's primary kind. It drives the icon, default view
// and grouping in the UI. What a device can do is given by which capability
// sub-documents are set, not by its type.
type DeviceType string

const (
	DeviceTypeCamera        DeviceType = "camera"
	DeviceTypeSensor        DeviceType = "sensor"
	DeviceTypeAccessControl DeviceType = "access_control"
	DeviceTypeGateway       DeviceType = "gateway"
)

func (t DeviceType) Valid() bool {
	switch t {
	case DeviceTypeCamera, DeviceTypeSensor, DeviceTypeAccessControl, DeviceTypeGateway:
		return true
	}
	return false
}

// DeviceCapability names a capability sub-document on Device. Each value
// equals the BSON field name of that sub-document.
type DeviceCapability string

const (
	DeviceCapabilityCamera         DeviceCapability = "camera"
	DeviceCapabilityLightSensor    DeviceCapability = "lightSensor"
	DeviceCapabilityHumiditySensor DeviceCapability = "humiditySensor"
	DeviceCapabilityGPS            DeviceCapability = "gps"
)

// deviceCapabilities is the single registry of capability sub-documents.
// Adding a capability: add the *XxxMetadata field on Device, a constant above
// and an entry here. TestDeviceCapabilitiesRegistered enforces it.
var deviceCapabilities = []struct {
	Capability DeviceCapability
	Has        func(Device) bool
}{
	{DeviceCapabilityCamera, func(d Device) bool { return d.Camera != nil }},
	{DeviceCapabilityLightSensor, func(d Device) bool { return d.LightSensor != nil }},
	{DeviceCapabilityHumiditySensor, func(d Device) bool { return d.HumiditySensor != nil }},
	{DeviceCapabilityGPS, func(d Device) bool { return d.GPS != nil }},
}

// DeviceConnectionStatus is derived at read time from the last-seen
// timestamps and thresholds; it is never stored.
type DeviceConnectionStatus string

const (
	DeviceConnectionHealthy      DeviceConnectionStatus = "healthy"
	DeviceConnectionIdle         DeviceConnectionStatus = "idle"
	DeviceConnectionDisconnected DeviceConnectionStatus = "disconnected"
)

// --------------------------------------------------------------------
// Data model for device
//
// Device is the persisted document.
//
// Ownership of sub-documents:
//   - top level:    identity, tenancy and lifecycle (hub-api)
//   - Agent:        written only by the heartbeat
//   - Metadata:     written only by user edits (DevicePatch)
//   - capabilities: reported fields by the heartbeat, saved configuration by users
//
// Writers $set explicit paths, never the whole struct, so a nil capability
// never removes one registered elsewhere. Removing a capability is an explicit
// $unset. Booleans that can flip back to false carry no omitempty.

type Device struct {
	Id            primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	SchemaVersion int                `json:"schemaVersion" bson:"schemaVersion,omitempty"`
	Key           string             `json:"key" bson:"key,omitempty"`   // Unique device key reported by the agent
	Name          string             `json:"name" bson:"name,omitempty"` // e.g. "Front Door Camera"
	Type          DeviceType         `json:"type" bson:"type,omitempty"`

	// Tenancy. Heartbeat credentials live on the project: the cloud key resolves
	// to an organisation and project, which must match these fields.
	OrganisationId string              `json:"organisationId" bson:"organisationId,omitempty"`
	ProjectId      *primitive.ObjectID `json:"projectId,omitempty" bson:"projectId,omitempty"` // nil or zero resolves to the organisation's default project
	SiteIds        []string            `json:"siteIds" bson:"siteIds,omitempty"`
	GroupIds       []string            `json:"groupIds" bson:"groupIds,omitempty"`

	// Lifecycle
	Active     bool  `json:"active" bson:"isActive"`
	MutedUntil int64 `json:"mutedUntil,omitempty" bson:"mute,omitempty"` // Notifications muted until (milliseconds since epoch)

	// Liveness (milliseconds since epoch)
	ConnectionStart int64             `json:"connectionStart" bson:"connectionStart,omitempty"` // Last (re)connect of the agent, e.g. after a reboot
	AgentLastSeen   int64             `json:"agentLastSeen" bson:"agentLastSeen,omitempty"`     // Last heartbeat from the agent
	DeviceLastSeen  int64             `json:"deviceLastSeen" bson:"deviceLastSeen,omitempty"`   // Last time the agent could reach the device itself
	Thresholds      *DeviceThresholds `json:"thresholds,omitempty" bson:"thresholds,omitempty"` // Optional override; nil uses project defaults

	// Latest media uploaded by the device, from any capability. Kept at the top
	// level: device lists sort on LatestMediaTimestamp.
	LatestMediaTimestamp int64           `json:"latestMediaTimestamp" bson:"latestMediaTimestamp,omitempty"` // milliseconds since epoch
	LatestMedia          *DeviceMediaRef `json:"latestMedia,omitempty" bson:"latestMedia,omitempty"`

	Location *Location `json:"location,omitempty" bson:"location,omitempty"` // Fixed placement, not a real-time position

	Agent    *DeviceAgent    `json:"agent,omitempty" bson:"agent,omitempty"`
	Metadata *DeviceMetadata `json:"metadata,omitempty" bson:"metadata,omitempty"`

	// Capabilities. Each is optional and any combination is valid, e.g. a camera
	// with a built-in humidity sensor sets both Camera and HumiditySensor.
	// Capability sub-documents hold capabilities and configuration only; live
	// readings go to a time-series collection.
	Camera         *CameraMetadata         `json:"camera,omitempty" bson:"camera,omitempty"`
	LightSensor    *LightSensorMetadata    `json:"lightSensor,omitempty" bson:"lightSensor,omitempty"`
	HumiditySensor *HumiditySensorMetadata `json:"humiditySensor,omitempty" bson:"humiditySensor,omitempty"`
	GPS            *GPSMetadata            `json:"gps,omitempty" bson:"gps,omitempty"`

	FeaturePermissions *DeviceFeaturePermissions `json:"featurePermissions,omitempty" bson:"featurePermissions,omitempty"`
	Audit              *Audit                    `json:"audit,omitempty" bson:"audit,omitempty"`

	// AtRuntimeMetadata is computed when the device is read and never stored.
	AtRuntimeMetadata *DeviceAtRuntimeMetadata `json:"atRuntimeMetadata,omitempty" bson:"-"`

	// @LEGACY - removed once all readers use the fields above.
	UserId              string                 `json:"user_id,omitempty" bson:"user_id,omitempty"`
	Enterprise          bool                   `json:"enterprise,omitempty" bson:"enterprise,omitempty"` // Licence concept; moves to the organisation
	DeprecatedAnalytics *[]DeprecatedHeartbeat `json:"analytics,omitempty" bson:"analytics,omitempty"`
}

// HasCapability reports whether the capability sub-document is set.
func (d Device) HasCapability(c DeviceCapability) bool {
	for _, entry := range deviceCapabilities {
		if entry.Capability == c {
			return entry.Has(d)
		}
	}
	return false
}

// Capabilities lists the capabilities the device has, in registry order.
func (d Device) Capabilities() []DeviceCapability {
	var capabilities []DeviceCapability
	for _, entry := range deviceCapabilities {
		if entry.Has(d) {
			capabilities = append(capabilities, entry.Capability)
		}
	}
	return capabilities
}

var (
	ErrDeviceKeyMissing          = errors.New("device key is required")
	ErrDeviceOrganisationMissing = errors.New("device organisationId is required")
	ErrDeviceTypeInvalid         = errors.New("device type is invalid")
)

// Validate checks the invariants every stored device must satisfy.
func (d Device) Validate() error {
	var errs []error
	if d.Key == "" {
		errs = append(errs, ErrDeviceKeyMissing)
	}
	if d.OrganisationId == "" {
		errs = append(errs, ErrDeviceOrganisationMissing)
	}
	if d.Type != "" && !d.Type.Valid() {
		errs = append(errs, fmt.Errorf("%w: %q", ErrDeviceTypeInvalid, d.Type))
	}
	if d.Thresholds != nil {
		errs = append(errs, d.Thresholds.Validate())
	}
	if d.LightSensor != nil {
		errs = append(errs, validRange("lightSensor lux", d.LightSensor.MinLux, d.LightSensor.MaxLux))
	}
	if d.HumiditySensor != nil {
		errs = append(errs, validRange("humiditySensor humidity", d.HumiditySensor.MinHumidity, d.HumiditySensor.MaxHumidity))
	}
	return errors.Join(errs...)
}

func validRange(name string, min, max *float64) error {
	if min != nil && max != nil && *min > *max {
		return fmt.Errorf("%s range: min %v exceeds max %v", name, *min, *max)
	}
	return nil
}

// DeviceThresholds overrides the project's liveness thresholds (milliseconds).
//   - healthy:      now - AgentLastSeen  <= Healthy
//   - idle:         now - DeviceLastSeen >  Idle (agent still reporting)
//   - disconnected: now - AgentLastSeen  >  Disconnected
type DeviceThresholds struct {
	Healthy      int64 `json:"healthy,omitempty" bson:"healthy,omitempty"`
	Idle         int64 `json:"idle,omitempty" bson:"idle,omitempty"`
	Disconnected int64 `json:"disconnected,omitempty" bson:"disconnected,omitempty"`
}

func (t DeviceThresholds) Validate() error {
	if t.Healthy < 0 || t.Idle < 0 || t.Disconnected < 0 {
		return errors.New("device thresholds must not be negative")
	}
	if t.Healthy > 0 && t.Disconnected > 0 && t.Healthy > t.Disconnected {
		return errors.New("device healthy threshold exceeds disconnected threshold")
	}
	return nil
}

// DeviceMediaRef is the subset of Media needed to render and sign the latest
// media. Field names match Media, so documents that still embed a full Media
// decode into it unchanged.
type DeviceMediaRef struct {
	Id                primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Duration          int                `json:"duration,omitempty" bson:"duration,omitempty"`
	StorageSolution   string             `json:"storageSolution,omitempty" bson:"storageSolution,omitempty"`
	VideoFile         string             `json:"videoFile,omitempty" bson:"videoFile,omitempty"`
	VideoProvider     string             `json:"videoProvider,omitempty" bson:"videoProvider,omitempty"`
	ThumbnailFile     string             `json:"thumbnailFile,omitempty" bson:"thumbnailFile,omitempty"`
	ThumbnailProvider string             `json:"thumbnailProvider,omitempty" bson:"thumbnailProvider,omitempty"`
	SpriteFile        string             `json:"spriteFile,omitempty" bson:"spriteFile,omitempty"`
	SpriteProvider    string             `json:"spriteProvider,omitempty" bson:"spriteProvider,omitempty"`

	// AtRuntimeMetadata carries the signed URLs and display time filled when
	// the device is read; never stored.
	AtRuntimeMetadata *MediaAtRuntimeMetadata `json:"atRuntimeMetadata,omitempty" bson:"-"`
}

// DeviceAgent is reported by the agent and written only by the heartbeat.
type DeviceAgent struct {
	Version     string `json:"version" bson:"version,omitempty"`         // Release tag of the agent
	ReleaseHash string `json:"releaseHash" bson:"releaseHash,omitempty"` // Build hash of the agent
	Repository  string `json:"repository" bson:"repository,omitempty"`
	Deployment  string `json:"deployment" bson:"deployment,omitempty"` // e.g. "factory", "docker", "kubernetes"

	Hostname     string   `json:"hostname" bson:"hostname,omitempty"`
	Architecture string   `json:"architecture" bson:"architecture,omitempty"`
	IPAddresses  []string `json:"ipAddresses" bson:"ipAddresses,omitempty"`
	MacAddresses []string `json:"macAddresses" bson:"macAddresses,omitempty"`

	BootTime      int64 `json:"bootTime" bson:"bootTime,omitempty"` // milliseconds since epoch
	UptimeSeconds int64 `json:"uptimeSeconds" bson:"uptimeSeconds,omitempty"`
	MemoryTotal   int64 `json:"memoryTotal" bson:"memoryTotal,omitempty"`     // bytes
	MemoryFree    int64 `json:"memoryFree" bson:"memoryFree,omitempty"`       // bytes
	MemoryUsed    int64 `json:"memoryUsed" bson:"memoryUsed,omitempty"`       // bytes
	MemoryProcess int64 `json:"memoryProcess" bson:"memoryProcess,omitempty"` // bytes

	HubEncryption bool `json:"hubEncryption" bson:"hubEncryption"`
	E2EEncryption bool `json:"e2eEncryption" bson:"e2eEncryption"`
}

// DeviceMetadata is edited by users and written only through DevicePatch.
type DeviceMetadata struct {
	Hidden           bool     `json:"hidden" bson:"hidden"`
	Color            string   `json:"color" bson:"color,omitempty"` // hex, e.g. "#FF5733"
	Brand            string   `json:"brand" bson:"brand,omitempty"`
	Model            string   `json:"model" bson:"model,omitempty"`
	Description      string   `json:"description" bson:"description,omitempty"`
	LastMaintenance  int64    `json:"lastMaintenance" bson:"lastMaintenance,omitempty"`   // milliseconds since epoch
	InstallationDate int64    `json:"installationDate" bson:"installationDate,omitempty"` // milliseconds since epoch
	OnFloorPlans     []string `json:"onFloorPlans" bson:"onFloorPlans,omitempty"`

	// VLM holds the stable ("normal") scenes used as context for VLM analysis.
	VLM *DeviceVLMMetadata `json:"vlm,omitempty" bson:"vlm,omitempty"`
}

// --------------------------------------------------------------------
// Capabilities
//
// One typed struct per capability kind. Each holds capabilities and
// configuration only, never live readings. Use a pointer field on Device by
// default; use a slice only for a kind that has several instances per device.

// CameraMetadata holds camera capabilities (heartbeat) and saved
// configuration (users).
type CameraMetadata struct {
	// Stream, reported by the agent
	Resolution  string  `json:"resolution" bson:"resolution,omitempty"` // e.g. "1920x1080"
	FrameRate   int64   `json:"frameRate" bson:"frameRate,omitempty"`   // fps
	Bitrate     int64   `json:"bitrate" bson:"bitrate,omitempty"`       // kbps
	Codec       string  `json:"codec" bson:"codec,omitempty"`           // e.g. "H.264"
	FieldOfView float64 `json:"fieldOfView" bson:"fieldOfView,omitempty"`

	// State, reported by the agent
	Connected           bool `json:"connected" bson:"connected"`
	ContinuousRecording bool `json:"continuousRecording" bson:"continuousRecording"`

	// Capabilities, reported by the agent
	HasOnvif       bool `json:"hasOnvif" bson:"hasOnvif"`
	HasAudio       bool `json:"hasAudio" bson:"hasAudio"`
	HasZoom        bool `json:"hasZoom" bson:"hasZoom"`
	HasPanTilt     bool `json:"hasPanTilt" bson:"hasPanTilt"`
	HasBackChannel bool `json:"hasBackChannel" bson:"hasBackChannel"`

	OnvifPresets []CameraPreset `json:"onvifPresets" bson:"onvifPresets,omitempty"` // Presets stored on the camera (agent)
	IOs          []IO           `json:"ios" bson:"ios,omitempty"`                   // Digital inputs and relay outputs (agent)

	// Saved configuration, edited by users
	Presets []CameraPreset `json:"presets" bson:"presets,omitempty"`
	Tours   []CameraTour   `json:"tours" bson:"tours,omitempty"`
}

type LightSensorMetadata struct {
	MinLux *float64 `json:"minLux,omitempty" bson:"minLux,omitempty"` // Measurable range
	MaxLux *float64 `json:"maxLux,omitempty" bson:"maxLux,omitempty"`
}

type HumiditySensorMetadata struct {
	MinHumidity *float64 `json:"minHumidity,omitempty" bson:"minHumidity,omitempty"` // Measurable range, percent
	MaxHumidity *float64 `json:"maxHumidity,omitempty" bson:"maxHumidity,omitempty"`
}

type GPSMetadata struct {
	Accuracy float64 `json:"accuracy" bson:"accuracy,omitempty"` // Rated accuracy in meters
	UpdateHz float64 `json:"updateHz" bson:"updateHz,omitempty"` // Fix rate

	// LastFix is a single overwritten snapshot so lists can show a position
	// without querying the time series.
	LastFix *GPSFix `json:"lastFix,omitempty" bson:"lastFix,omitempty"`
}

type GPSFix struct {
	Latitude  float64 `json:"latitude" bson:"latitude"`
	Longitude float64 `json:"longitude" bson:"longitude"`
	Timestamp int64   `json:"timestamp" bson:"timestamp"` // milliseconds since epoch
}

// --------------------------------------------------------------------
// Runtime and patch

type DeviceAtRuntimeMetadata struct {
	Sites        []SiteOption           `json:"sites,omitempty"`
	Groups       []GroupOption          `json:"groups,omitempty"`
	Capabilities []DeviceCapability     `json:"capabilities,omitempty"`
	Status       DeviceConnectionStatus `json:"status"`
}

// DevicePatch is the request body for a device update. It targets the
// user-owned paths only; use ToSet to build the $set document. Nil fields are
// left untouched. Dates stay time.Time in the API and are stored as
// milliseconds since epoch.
type DevicePatch struct {
	Name               *string                   `json:"name,omitempty"`
	Color              *string                   `json:"color,omitempty"`
	Brand              *string                   `json:"brand,omitempty"`
	Model              *string                   `json:"model,omitempty"`
	Description        *string                   `json:"description,omitempty"`
	Hidden             *bool                     `json:"hidden,omitempty"`
	LastMaintenance    *time.Time                `json:"lastMaintenance,omitempty"`
	InstallationDate   *time.Time                `json:"installationDate,omitempty"`
	OnFloorPlans       *[]string                 `json:"onFloorPlans,omitempty"`
	FieldOfView        *float64                  `json:"fieldOfView,omitempty"`
	FeaturePermissions *DeviceFeaturePermissions `json:"featurePermissions,omitempty"`
}

func (p DevicePatch) ToSet() map[string]interface{} {
	set := map[string]interface{}{}
	setIf(set, "name", p.Name)
	setIf(set, "metadata.color", p.Color)
	setIf(set, "metadata.brand", p.Brand)
	setIf(set, "metadata.model", p.Model)
	setIf(set, "metadata.description", p.Description)
	setIf(set, "metadata.hidden", p.Hidden)
	setIf(set, "metadata.lastMaintenance", unixMilli(p.LastMaintenance))
	setIf(set, "metadata.installationDate", unixMilli(p.InstallationDate))
	setIf(set, "metadata.onFloorPlans", p.OnFloorPlans)
	setIf(set, "camera.fieldOfView", p.FieldOfView)
	if p.FeaturePermissions != nil {
		set["featurePermissions"] = p.FeaturePermissions
	}
	return set
}

func unixMilli(t *time.Time) *int64 {
	if t == nil {
		return nil
	}
	ms := t.UnixMilli()
	return &ms
}

func setIf[T any](set map[string]interface{}, path string, v *T) {
	if v != nil {
		set[path] = *v
	}
}

// --------------------------------------------------------------------
// Shared and legacy types

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

// @Deprecated - to be removed in future versions
type DeprecatedHeartbeat struct {
	CloudPublicKey string `json:"cloudpublicKey,omitempty" bson:"cloudpublickey,omitempty"`
	Encrypted      bool   `json:"encrypted,omitempty"`
	EncryptedData  []byte `json:"encryptedData,omitempty"`
	// -----------
	Key              string                  `json:"key,omitempty"`
	HubEncryption    string                  `json:"hub_encryption,omitempty" bson:"hub_encryption,omitempty"`
	E2EEncryption    string                  `json:"e2e_encryption,omitempty" bson:"e2e_encryption,omitempty"`
	Enterprise       bool                    `json:"enterprise,omitempty"`
	Hash             string                  `json:"hash,omitempty"`
	Version          string                  `json:"version,omitempty"`
	Release          string                  `json:"release,omitempty"`
	MACs             []string                `json:"mac_list,omitempty" bson:"mac_list,omitempty"`
	IPs              []string                `json:"ip_list,omitempty" bson:"ip_list,omitempty"`
	CpuID            string                  `json:"cpuid,omitempty" bson:"cpuid,omitempty"`
	CloudUser        string                  `json:"clouduser,omitempty" bson:"clouduser,omitempty"`
	CameraName       string                  `json:"cameraname,omitempty" bson:"cameraname,omitempty"`
	CameraType       string                  `json:"cameratype,omitempty" bson:"cameratype,omitempty"`
	Architecture     string                  `json:"architecture,omitempty"`
	Hostname         string                  `json:"hostname,omitempty"`
	FreeMemory       string                  `json:"freeMemory,omitempty" bson:"freeMemory,omitempty"`
	TotalMemory      string                  `json:"totalMemory,omitempty" bson:"totalMemory,omitempty"`
	UsedMemory       string                  `json:"usedMemory,omitempty" bson:"usedMemory,omitempty"`
	ProcessMemory    string                  `json:"processMemory,omitempty" bson:"processMemory,omitempty"`
	Kubernetes       bool                    `json:"kubernetes,omitempty"`
	Docker           bool                    `json:"docker,omitempty"`
	Kios             bool                    `json:"kios,omitempty"`
	Raspberrypi      bool                    `json:"raspberrypi,omitempty"`
	Uptime           string                  `json:"uptime,omitempty" bson:"uptime,omitempty"`
	BootTime         string                  `json:"boot_time,omitempty" bson:"boot_time,omitempty"`
	Timestamp        int64                   `json:"timestamp" bson:"timestamp"`
	SiteID           string                  `json:"siteID" bson:"siteID,omitempty"`
	ONVIF            string                  `json:"onvif" bson:"onvif,omitempty"`
	ONVIFZoom        string                  `json:"onvif_zoom" bson:"onvif_zoom,omitempty"`
	ONVIFPanTilt     string                  `json:"onvif_pantilt" bson:"onvif_pantilt,omitempty"`
	ONVIFPresets     string                  `json:"onvif_presets" bson:"onvif_presets,omitempty"`
	ONVIFPresetsList []DeprecatedPreset      `json:"onvif_presets_list" bson:"onvif_presets_list,omitempty"`
	ONVIFEventsList  []DeprecatedONVIFEvents `json:"onvif_events_list" bson:"onvif_events_list,omitempty"`
	CameraConnected  string                  `json:"cameraConnected" bson:"cameraConnected,omitempty"`
	HasBackChannel   string                  `json:"hasBackChannel" bson:"hasBackChannel,omitempty"`
	//Board           string   `json:"board,omitempty"`
	//Disk1Size       string   `json:"disk1size,omitempty" bson:"disk1size,omitempty"`
	//Disk3Size       string   `json:"disk3size,omitempty" bson:"disk3size,omitempty"`
	//DiskVDASize     string   `json:"diskvdasize,omitempty" bson:"diskvdasize,omitempty"`
	//NumberOfFiles   string   `json:"numberofiles,omitempty" bson:"numberoffiles,omitempty"`
	//Temperature string `json:"temperature,omitempty"`
	//WifiSSID        string   `json:"wifissid,omitempty" bson:"wifissid,omitempty"`
	//WifiStrength    string   `json:"wifistrength,omitempty" bson:"wifisstrength,omitempty"`
}

// @Deprecated - to be removed in future versions
type DeprecatedPreset struct {
	Name  string  `json:"name" bson:"name"`
	Token string  `json:"token" bson:"token"`
	X     float64 `json:"x" bson:"x"`
	Y     float64 `json:"y" bson:"y"`
	Z     float64 `json:"z" bson:"z"`
}

// @Deprecated - to be removed in future versions
type DeprecatedONVIFEvents struct {
	Key       string `json:"key" bson:"key,omitempty"`
	Type      string `json:"type" bson:"type,omitempty"`
	Value     string `json:"value" bson:"value,omitempty"`
	Timestamp int64  `json:"timestamp" bson:"timestamp,omitempty"`
}
