package models

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	PermissionDevicesRead   Permission = "devices.read"
	PermissionDevicesUpdate Permission = "devices.update"
	PermissionDevicesDelete Permission = "devices.delete"
	PermissionDevicesConfig Permission = "devices.config"
	PermissionLivePreview   Permission = "live.preview"
	PermissionLiveStream    Permission = "live.stream"
	PermissionPTZZoom       Permission = "ptz.zoom"
	PermissionPTZPanTilt    Permission = "ptz.panTilt"
	PermissionPTZControl    Permission = "ptz.control"
	PermissionPTZPresets    Permission = "ptz.presets"
)

var devicePermissions = []Permission{
	PermissionDevicesRead,
	PermissionDevicesUpdate,
	PermissionDevicesDelete,
	PermissionDevicesConfig,
	PermissionLivePreview,
	PermissionLiveStream,
	PermissionPTZZoom,
	PermissionPTZPanTilt,
	PermissionPTZControl,
	PermissionPTZPresets,
}

// devicePermissionImplications encodes the graded legacy feature scales
// (PermissionLevel PTZ 2-5, live view 2-3) as permissions that imply their
// lower grades, so a migrated level maps to exactly one permission.
var devicePermissionImplications = map[Permission][]Permission{
	PermissionLiveStream: {PermissionLivePreview},
	PermissionPTZPanTilt: {PermissionPTZZoom},
	PermissionPTZControl: {PermissionPTZPanTilt},
	PermissionPTZPresets: {PermissionPTZControl},
}

// DevicePermissions returns the canonical device permission catalog.
// PermissionDevicesConfig covers remote agent configuration; live view and PTZ
// are graded capabilities, see devicePermissionImplications.
func DevicePermissions() []Permission {
	return append([]Permission(nil), devicePermissions...)
}

// PermissionDevicesManageSystemTags governs platform-defined device system tags
// and is only effective organisation-wide.
const (
	PermissionDevicesManageSystemTags Permission = "devices.manageSystemTags"
)

var deviceAdministrationPermissions = []Permission{
	PermissionDevicesManageSystemTags,
}

// DeviceAdministrationPermissions returns the canonical organisation-level device permission catalog.
func DeviceAdministrationPermissions() []Permission {
	return append([]Permission(nil), deviceAdministrationPermissions...)
}

// DeviceSchemaVersion is the current shape of the device document. Only a
// writer that leaves the whole document in this shape stamps it: provisioning
// (DeviceHeartbeat.ProvisionSet) and the backfill, which upgrades documents
// below it. Partial writers such as the heartbeat's FullSet must not, or the
// backfill would skip documents whose user-edited fields still sit on the
// legacy paths.
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

// DeviceConnectionStatus is derived at read time from AgentLastSeen,
// DeviceLastSeen and the thresholds (see DeviceThresholds); it is never stored.
type DeviceConnectionStatus string

const (
	DeviceConnectionConnected    DeviceConnectionStatus = "connected"    // agent and device are both reporting
	DeviceConnectionIdle         DeviceConnectionStatus = "idle"         // agent reports, but cannot reach the device
	DeviceConnectionDisconnected DeviceConnectionStatus = "disconnected" // agent stopped reporting
)

// --------------------------------------------------------------------
// Data model for device
//
// Device is the persisted document.
//
// Ownership of sub-documents:
//   - top level:    identity, tenancy and lifecycle (hub-api)
//   - Agent:        written only by the heartbeat (DeviceHeartbeat)
//   - Metadata:     written only by user edits (DevicePatch)
//   - capabilities: reported fields by the heartbeat, saved configuration by users
//
// DeviceHeartbeat and DevicePatch build the $set documents for the two
// writers; their paths never overlap.
//
// Writers $set explicit paths, never the whole struct, so a nil capability
// never removes one registered elsewhere. Removing a capability is an explicit
// $unset. Booleans that can flip back to false carry no omitempty.

type Device struct {
	Id            primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	SchemaVersion int                `json:"schemaVersion" bson:"schemaVersion,omitempty"`
	Key           string             `json:"key" bson:"key,omitempty"`   // Device key reported by the agent, unique within its project
	Name          string             `json:"name" bson:"name,omitempty"` // Optional stored name (DevicePatch.Name); display falls back to Agent.Name, the name the agent reports
	Type          DeviceType         `json:"type" bson:"type,omitempty"`

	// Tenancy. Heartbeat credentials live on the project: the cloud key resolves
	// to an organisation and project, which must match these fields.
	OrganisationId string              `json:"organisationId" bson:"organisationId,omitempty"`
	ProjectId      *primitive.ObjectID `json:"projectId,omitempty" bson:"projectId,omitempty"` // nil or zero resolves to the organisation's default project
	SiteIds        []string            `json:"siteIds" bson:"siteIds,omitempty"`
	GroupIds       []string            `json:"groupIds" bson:"groupIds,omitempty"`

	// Tags are short identifiers on a device, grouped by type (DeviceTags).
	Tags *DeviceTags `json:"tags,omitempty" bson:"tags,omitempty"`

	// Lifecycle
	MutedUntil int64 `json:"mute,omitempty" bson:"mute,omitempty"` // Notifications muted until (seconds since epoch)

	// Liveness: the inputs the connection status is derived from, kept together.
	// Timestamps are points in time on the hub's clock (milliseconds since
	// epoch), written by every heartbeat; readers derive durations and status.
	AgentLastSeen  int64          `json:"agentLastSeen" bson:"agentLastSeen,omitempty"`   // Last heartbeat from the agent
	DeviceLastSeen int64          `json:"deviceLastSeen" bson:"deviceLastSeen,omitempty"` // Last heartbeat in which the agent could reach the device itself
	Monitor        *DeviceMonitor `json:"monitor,omitempty" bson:"monitor,omitempty"`     // Last notified state, owned by hub-monitor-device

	// Latest media uploaded by the device, from any capability. Kept at the top
	// level: device lists sort on LatestMediaTimestamp. LatestMedia is a value,
	// not a pointer, so readers never need a nil check; an empty VideoFile means
	// the device has no media yet (the media pipeline does not store the Id).
	LatestMediaTimestamp int64          `json:"latestMediaTimestamp" bson:"latestMediaTimestamp,omitempty"` // seconds since epoch (Media.StartTimestamp)
	LatestMedia          DeviceMediaRef `json:"latestMedia" bson:"latestMedia,omitempty"`

	// Location is the fixed placement owners set (DevicePatch), e.g. an
	// address picked on a map. It is not a real-time position.
	Location *Location `json:"location,omitempty" bson:"location,omitempty"`

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
	Audit              *Audit                    `json:"audit,omitempty" bson:"audit,omitempty"` // Provisioning sets CreatedAt (first heartbeat)

	// AtRuntimeMetadata is computed when the device is read and never stored.
	AtRuntimeMetadata *DeviceAtRuntimeMetadata `json:"atRuntimeMetadata,omitempty" bson:"-"`

	// @LEGACY - removed once all readers use the fields above.
	UserId              string                 `json:"user_id,omitempty" bson:"user_id,omitempty"`
	IsActive            bool                   `json:"isActive" bson:"isActive"` // No defined meaning; never set true for devices
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
	ErrDeviceTagInvalid          = errors.New("device tag is invalid")
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
	if d.Tags != nil {
		if normalized, err := NormalizeDeviceTags(d.Tags.System); err != nil {
			errs = append(errs, err)
		} else if !slices.Equal(normalized, d.Tags.System) && len(d.Tags.System) > 0 {
			errs = append(errs, fmt.Errorf("%w: system tags are not normalized", ErrDeviceTagInvalid))
		}
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

// DeviceThresholds are the liveness thresholds (milliseconds), one per failure
// mode. They are not configurable or stored yet; per-device or per-project
// overrides can be added later without a migration. The status is derived in
// this order:
//   - disconnected: now - AgentLastSeen  > Disconnected (the agent is gone)
//   - idle:         now - DeviceLastSeen > Idle (the agent reports, but cannot reach the device)
//   - connected:    otherwise
type DeviceThresholds struct {
	Idle         int64 `json:"idle,omitempty" bson:"idle,omitempty"`
	Disconnected int64 `json:"disconnected,omitempty" bson:"disconnected,omitempty"`
}

// Default liveness thresholds (milliseconds), used for any threshold that is
// zero. Disconnected matches the 180 s the hub-api videowall uses today.
const (
	DefaultDeviceIdleThreshold         int64 = 180_000
	DefaultDeviceDisconnectedThreshold int64 = 180_000
)

// ConnectionStatus derives the device's status at now (milliseconds since
// epoch), in the order documented on DeviceThresholds. Zero thresholds fall
// back to the defaults. A device without a capability that reports its own
// reachability (today: Camera) is never idle. Every reader derives the status
// here, so services cannot drift apart.
func (d Device) ConnectionStatus(now int64, thresholds DeviceThresholds) DeviceConnectionStatus {
	idle := thresholds.Idle
	if idle <= 0 {
		idle = DefaultDeviceIdleThreshold
	}
	disconnected := thresholds.Disconnected
	if disconnected <= 0 {
		disconnected = DefaultDeviceDisconnectedThreshold
	}
	if d.AgentLastSeen <= 0 || now-d.AgentLastSeen > disconnected {
		return DeviceConnectionDisconnected
	}
	if d.Camera != nil && (d.DeviceLastSeen <= 0 || now-d.DeviceLastSeen > idle) {
		return DeviceConnectionIdle
	}
	return DeviceConnectionConnected
}

func (t DeviceThresholds) Validate() error {
	if t.Idle < 0 || t.Disconnected < 0 {
		return errors.New("device thresholds must not be negative")
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
	// CloudPublicKey is the public key the agent presented in its last
	// heartbeat. Agents within a project can be on different keys (e.g. a
	// legacy user key after the project credentials were regenerated), so it
	// is recorded per device for routing and key status. Authentication still
	// resolves key -> project; private and symmetric keys never live on the
	// device.
	CloudPublicKey string `json:"cloudPublicKey" bson:"cloudPublicKey,omitempty"`
	Name           string `json:"name" bson:"name,omitempty"` // Name configured on the agent (AGENT_NAME), as last reported

	Version     string `json:"version" bson:"version,omitempty"`         // Release tag of the agent
	ReleaseHash string `json:"releaseHash" bson:"releaseHash,omitempty"` // Build hash of the agent
	Repository  string `json:"repository" bson:"repository,omitempty"`
	// Deployment is how the agent runs: "factory", "kubernetes", "docker",
	// "kios" or "raspberrypi". It replaces the agent's legacy enterprise flag,
	// which only meant running under Kerberos Factory or Kubernetes.
	Deployment string `json:"deployment" bson:"deployment,omitempty"`

	Hostname     string   `json:"hostname" bson:"hostname,omitempty"`
	Architecture string   `json:"architecture" bson:"architecture,omitempty"`
	IPAddresses  []string `json:"ipAddresses" bson:"ipAddresses,omitempty"`
	MacAddresses []string `json:"macAddresses" bson:"macAddresses,omitempty"`

	// StartedAt (agent process start) and BootTime (host boot) are points in
	// time on the hub's clock (milliseconds since epoch), displayed as agent and
	// host uptime (now - StartedAt, now - BootTime). Durations are never
	// stored: they are stale as soon as they are written. Agents report both as
	// human-readable durations ("10 hours"), so the hub derives them on a
	// detected restart or reboot; they are approximate until the next one.
	StartedAt     int64 `json:"startedAt" bson:"startedAt,omitempty"`
	BootTime      int64 `json:"bootTime" bson:"bootTime,omitempty"`
	MemoryTotal   int64 `json:"memoryTotal" bson:"memoryTotal,omitempty"`     // bytes
	MemoryFree    int64 `json:"memoryFree" bson:"memoryFree,omitempty"`       // bytes
	MemoryUsed    int64 `json:"memoryUsed" bson:"memoryUsed,omitempty"`       // bytes
	MemoryProcess int64 `json:"memoryProcess" bson:"memoryProcess,omitempty"` // bytes

	HubEncryption bool `json:"hubEncryption" bson:"hubEncryption"`
	E2EEncryption bool `json:"e2eEncryption" bson:"e2eEncryption"`
}

// DeviceMonitor is owned by hub-monitor-device: the last state it notified
// about, used to detect transitions. Readers use the derived
// AtRuntimeMetadata.Status instead.
type DeviceMonitor struct {
	Status    string `json:"status" bson:"status,omitempty"`       // "active" or "inactive"
	ChangedAt int64  `json:"changedAt" bson:"changedAt,omitempty"` // milliseconds since epoch
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
	Resolution string `json:"resolution" bson:"resolution,omitempty"` // e.g. "1920x1080"
	FrameRate  int64  `json:"frameRate" bson:"frameRate,omitempty"`   // fps
	Bitrate    int64  `json:"bitrate" bson:"bitrate,omitempty"`       // kbps
	Codec      string `json:"codec" bson:"codec,omitempty"`           // e.g. "H.264"

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
	FieldOfView float64        `json:"fieldOfView" bson:"fieldOfView,omitempty"` // degrees, used on floor plans
	Presets     []CameraPreset `json:"presets" bson:"presets,omitempty"`
	Tours       []CameraTour   `json:"tours" bson:"tours,omitempty"`
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
// Runtime, patch and heartbeat

type DeviceAtRuntimeMetadata struct {
	Sites        []SiteOption           `json:"sites,omitempty"`
	Groups       []GroupOption          `json:"groups,omitempty"`
	Capabilities []DeviceCapability     `json:"capabilities,omitempty"`
	Status       DeviceConnectionStatus `json:"status"`
}

// DevicePatch is the request body for a device update. It targets the
// user-owned paths only. Nil fields are left untouched. Dates stay time.Time
// in the API and are stored as milliseconds since epoch.
//
// ToSet builds the $set for paths every device has. CameraToSet builds the
// camera.* paths separately: apply it with {"camera": {"$exists": true}} in the
// match, otherwise the $set creates a camera sub-document and the device gains
// the camera capability.
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
	Location           *Location                 `json:"location,omitempty"`   // Replaces the whole location
	SystemTags         *[]string                 `json:"systemTags,omitempty"` // Replaces all system tags; normalized by ToSet, check with Validate
}

// Validate checks the fields of a patch that can be invalid on their own.
func (p DevicePatch) Validate() error {
	if p.SystemTags != nil {
		if _, err := NormalizeDeviceTags(*p.SystemTags); err != nil {
			return err
		}
	}
	return nil
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
	if p.FeaturePermissions != nil {
		set["featurePermissions"] = p.FeaturePermissions
	}
	if p.Location != nil {
		set["location"] = p.Location
	}
	if p.SystemTags != nil {
		// Validate rejects tags that do not normalize; ToSet stores the
		// normalized form so stored tags always match trigger tags.
		tags, _ := NormalizeDeviceTags(*p.SystemTags)
		set["tags.system"] = tags
	}
	return set
}

// CameraToSet builds the $set for the camera.* paths; see DevicePatch.
func (p DevicePatch) CameraToSet() map[string]interface{} {
	set := map[string]interface{}{}
	setIf(set, "camera.fieldOfView", p.FieldOfView)
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

// DeviceHeartbeat is the agent-reported state from one heartbeat, normalised
// from the agent protocol by the receiving service. It is the heartbeat's
// counterpart of DevicePatch: DeviceHeartbeat writes only the paths the
// agent owns, DevicePatch only the paths users own, and the two never
// overlap (TestDeviceHeartbeatAndPatchPathsAreDisjoint).
//
// All writes use explicit paths, never whole sub-documents, so user-owned
// camera configuration and restart-derived times are never overwritten.
// It is not stored itself.
type DeviceHeartbeat struct {
	// ReceivedAt is when the hub received the heartbeat, on the hub's clock
	// (milliseconds since epoch). Agent clocks are not trusted.
	ReceivedAt int64 `json:"receivedAt" bson:"-"`

	// Type is the device's primary kind as known to the agent; empty keeps the
	// stored type.
	Type DeviceType `json:"type,omitempty" bson:"-"`

	// Agent holds the agent-reported fields. StartedAt and BootTime are only
	// set when the receiver detected a restart or reboot; 0 keeps the stored
	// value.
	Agent DeviceAgent `json:"agent" bson:"-"`

	// Camera holds the agent-reported camera fields; the user-owned ones
	// (FieldOfView, Presets, Tours) are ignored. Nil for devices without the
	// camera capability.
	Camera *CameraMetadata `json:"camera,omitempty" bson:"-"`
}

// RefreshSet builds the $set for the fields that change on every beat. It is
// enough for a heartbeat whose stable content did not change.
func (h DeviceHeartbeat) RefreshSet() map[string]interface{} {
	set := map[string]interface{}{
		"agentLastSeen":       h.ReceivedAt,
		"agent.memoryFree":    h.Agent.MemoryFree,
		"agent.memoryUsed":    h.Agent.MemoryUsed,
		"agent.memoryProcess": h.Agent.MemoryProcess,
	}
	// The device itself is reachable when its camera stream is connected.
	if h.Camera != nil && h.Camera.Connected {
		set["deviceLastSeen"] = h.ReceivedAt
	}
	setIfPositive(set, "agent.startedAt", h.Agent.StartedAt)
	setIfPositive(set, "agent.bootTime", h.Agent.BootTime)
	return set
}

// FullSet builds the $set for every heartbeat-owned field: RefreshSet plus
// the stable agent-reported fields.
func (h DeviceHeartbeat) FullSet() map[string]interface{} {
	set := h.RefreshSet()
	if h.Type != "" {
		set["type"] = h.Type
	}

	agent := h.Agent
	set["agent.cloudPublicKey"] = agent.CloudPublicKey
	set["agent.name"] = agent.Name
	set["agent.version"] = agent.Version
	set["agent.releaseHash"] = agent.ReleaseHash
	set["agent.repository"] = agent.Repository
	set["agent.deployment"] = agent.Deployment
	set["agent.hostname"] = agent.Hostname
	set["agent.architecture"] = agent.Architecture
	set["agent.ipAddresses"] = agent.IPAddresses
	set["agent.macAddresses"] = agent.MacAddresses
	set["agent.memoryTotal"] = agent.MemoryTotal
	set["agent.hubEncryption"] = agent.HubEncryption
	set["agent.e2eEncryption"] = agent.E2EEncryption

	if camera := h.Camera; camera != nil {
		set["camera.resolution"] = camera.Resolution
		set["camera.frameRate"] = camera.FrameRate
		set["camera.bitrate"] = camera.Bitrate
		set["camera.codec"] = camera.Codec
		set["camera.connected"] = camera.Connected
		set["camera.continuousRecording"] = camera.ContinuousRecording
		set["camera.hasOnvif"] = camera.HasOnvif
		set["camera.hasAudio"] = camera.HasAudio
		set["camera.hasZoom"] = camera.HasZoom
		set["camera.hasPanTilt"] = camera.HasPanTilt
		set["camera.hasBackChannel"] = camera.HasBackChannel
		set["camera.onvifPresets"] = camera.OnvifPresets
		set["camera.ios"] = camera.IOs
	}
	return set
}

// ProvisionSet builds the $setOnInsert for a device created by its first
// heartbeat: FullSet plus the creation time and the schema version, since a
// new device is complete in the current shape. Name is left empty so display
// follows Agent.Name, the name the agent reports (its friendly name from the
// remote config, else its configured name), and renames on the agent show.
func (h DeviceHeartbeat) ProvisionSet() map[string]interface{} {
	set := h.FullSet()
	set["schemaVersion"] = DeviceSchemaVersion
	set["audit.createdAt"] = time.UnixMilli(h.ReceivedAt).UTC()
	return set
}

func setIfPositive(set map[string]interface{}, path string, value int64) {
	if value > 0 {
		set[path] = value
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
	// ContinuousRecording is "true" or "false"; empty when the agent does not
	// report it (older agents), unlike CameraMetadata.ContinuousRecording.
	ContinuousRecording string `json:"continuousRecording,omitempty" bson:"continuousRecording,omitempty"`
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

// DeviceTags holds a device's tags per type, each in NormalizeDeviceTags form.
// Each type has its own writer and permissions and its own path (tags.<type>),
// so writing one type never touches another.
type DeviceTags struct {
	// System tags come from the deployment's catalog (Helm) and are set by
	// owners (DevicePatch.SystemTags). Workflow triggers select devices by
	// system tag (WorkflowTrigger.SystemTags), which lets a deployment-global
	// workflow run on specific devices across organisations; a deployment that
	// defines such workflows must trust its owners (self-hosted, not
	// multi-tenant SaaS).
	System []string `json:"system,omitempty" bson:"system,omitempty"`
}

// Device tag limits, per tag type.
const (
	MaxDeviceTags      = 32
	MaxDeviceTagLength = 64
)

var deviceTagPattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9._:-]*[a-z0-9])?$`)

// NormalizeDeviceTags returns tags in their stored form: trimmed, lowercased,
// de-duplicated and sorted. A tag is 1-64 characters of a-z, 0-9, ".", "_",
// ":" and "-", starting and ending with a letter or digit, and a device has at
// most 32 of each type. The empty list is valid (no tags). Triggers use the
// same form, so a tag always matches regardless of how it was typed.
func NormalizeDeviceTags(tags []string) ([]string, error) {
	normalized := make([]string, 0, len(tags))
	seen := make(map[string]bool, len(tags))
	for _, tag := range tags {
		tag = strings.ToLower(strings.TrimSpace(tag))
		if len(tag) == 0 || len(tag) > MaxDeviceTagLength || !deviceTagPattern.MatchString(tag) {
			return nil, fmt.Errorf("%w: %q must be 1-%d characters of a-z, 0-9, \".\", \"_\", \":\" or \"-\", starting and ending with a letter or digit", ErrDeviceTagInvalid, tag, MaxDeviceTagLength)
		}
		if seen[tag] {
			continue
		}
		seen[tag] = true
		normalized = append(normalized, tag)
	}
	if len(normalized) > MaxDeviceTags {
		return nil, fmt.Errorf("%w: a device has at most %d tags of one type", ErrDeviceTagInvalid, MaxDeviceTags)
	}
	sort.Strings(normalized)
	return normalized, nil
}
