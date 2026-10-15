package models

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Every *XxxMetadata capability field on Device must be in deviceCapabilities,
// and its constant must equal the field's BSON name.
func TestDeviceCapabilitiesRegistered(t *testing.T) {
	registered := map[DeviceCapability]bool{}
	for _, entry := range deviceCapabilities {
		registered[entry.Capability] = true
	}
	deviceType := reflect.TypeOf(Device{})
	found := 0
	for i := 0; i < deviceType.NumField(); i++ {
		field := deviceType.Field(i)
		elem := field.Type
		if elem.Kind() == reflect.Pointer || elem.Kind() == reflect.Slice {
			elem = elem.Elem()
		}
		name := elem.Name()
		if !strings.HasSuffix(name, "Metadata") || name == "DeviceMetadata" || strings.HasPrefix(name, "DeviceAtRuntime") {
			continue
		}
		found++
		bsonName := strings.Split(field.Tag.Get("bson"), ",")[0]
		if !registered[DeviceCapability(bsonName)] {
			t.Errorf("capability field %s (%s) is not in deviceCapabilities", field.Name, bsonName)
		}
	}
	if found != len(deviceCapabilities) {
		t.Errorf("found %d capability fields, registry has %d entries", found, len(deviceCapabilities))
	}
}

func TestDeviceCapabilities(t *testing.T) {
	device := Device{Camera: &CameraMetadata{}, HumiditySensor: &HumiditySensorMetadata{}}
	want := []DeviceCapability{DeviceCapabilityCamera, DeviceCapabilityHumiditySensor}
	if got := device.Capabilities(); !reflect.DeepEqual(got, want) {
		t.Fatalf("Capabilities() = %v, want %v", got, want)
	}
	if !device.HasCapability(DeviceCapabilityHumiditySensor) || device.HasCapability(DeviceCapabilityGPS) {
		t.Fatal("HasCapability does not match the set sub-documents")
	}
	if (Device{}).Capabilities() != nil {
		t.Fatal("a device without capability sub-documents has no capabilities")
	}
}

// Booleans without omitempty must be written as false so a $set clears them.
func TestDeviceFalseBooleansAreWritten(t *testing.T) {
	raw, err := bson.Marshal(Device{
		Camera:   &CameraMetadata{HasPanTilt: false},
		Metadata: &DeviceMetadata{Hidden: false},
		Agent:    &DeviceAgent{HubEncryption: false},
	})
	if err != nil {
		t.Fatal(err)
	}
	doc := bson.Raw(raw)
	for _, path := range []string{"camera.hasPanTilt", "metadata.hidden", "agent.hubEncryption", "isActive"} {
		value, err := doc.LookupErr(strings.Split(path, ".")...)
		if err != nil {
			t.Errorf("%s is not written: %v", path, err)
			continue
		}
		if value.Boolean() {
			t.Errorf("%s = true, want false", path)
		}
	}
}

func TestDeviceOptionalFieldsAreNotWrittenWhenNil(t *testing.T) {
	raw, err := bson.Marshal(Device{
		Key:               "cam-1",
		AtRuntimeMetadata: &DeviceAtRuntimeMetadata{Status: DeviceConnectionConnected},
	})
	if err != nil {
		t.Fatal(err)
	}
	doc := bson.Raw(raw)
	for _, key := range []string{"featurePermissions", "camera", "atRuntimeMetadata", "agent", "metadata"} {
		if _, err := doc.LookupErr(key); err == nil {
			t.Errorf("%s is written but should be omitted", key)
		}
	}
}

// Devices written before DeviceMediaRef embed a full Media under latestMedia.
func TestDeviceMediaRefDecodesLegacyEmbeddedMedia(t *testing.T) {
	mediaId := primitive.NewObjectID()
	legacy := bson.M{
		"key":                  "cam-1",
		"latestMediaTimestamp": int64(1706000000),
		"latestMedia": Media{
			Id:             mediaId,
			StartTimestamp: 1706000000,
			Duration:       30,
			DeviceKey:      "cam-1",
			VideoFile:      "org/cam-1/video.mp4",
			VideoProvider:  "vault",
			ThumbnailFile:  "org/cam-1/thumb.jpg",
			SpriteFile:     "org/cam-1/sprite.jpg",
			MarkerNames:    []string{"person"},
			AtRuntimeMetadata: &MediaAtRuntimeMetadata{
				VideoUrl: "https://stale.example/video.mp4",
			},
		},
	}
	raw, err := bson.Marshal(legacy)
	if err != nil {
		t.Fatal(err)
	}
	var device Device
	if err := bson.Unmarshal(raw, &device); err != nil {
		t.Fatal(err)
	}
	want := DeviceMediaRef{
		Id:            mediaId,
		Duration:      30,
		VideoFile:     "org/cam-1/video.mp4",
		VideoProvider: "vault",
		ThumbnailFile: "org/cam-1/thumb.jpg",
		SpriteFile:    "org/cam-1/sprite.jpg",
	}
	if !reflect.DeepEqual(device.LatestMedia, want) {
		t.Fatalf("LatestMedia = %+v, want %+v", device.LatestMedia, want)
	}
	if device.LatestMediaTimestamp != 1706000000 {
		t.Fatalf("LatestMediaTimestamp = %d", device.LatestMediaTimestamp)
	}
}

func TestDevicePatchToSet(t *testing.T) {
	name := "Front door"
	color := "#FF5733"
	hidden := false
	installed := time.UnixMilli(1706000000000).UTC()
	fieldOfView := 90.0
	patch := DevicePatch{
		Name:             &name,
		Color:            &color,
		Hidden:           &hidden,
		InstallationDate: &installed,
		FieldOfView:      &fieldOfView,
	}
	want := map[string]interface{}{
		"name":                      name,
		"metadata.color":            color,
		"metadata.hidden":           false,
		"metadata.installationDate": int64(1706000000000),
	}
	if got := patch.ToSet(); !reflect.DeepEqual(got, want) {
		t.Fatalf("ToSet() = %v, want %v", got, want)
	}
	wantCamera := map[string]interface{}{"camera.fieldOfView": fieldOfView}
	if got := patch.CameraToSet(); !reflect.DeepEqual(got, wantCamera) {
		t.Fatalf("CameraToSet() = %v, want %v", got, wantCamera)
	}
	if got := (DevicePatch{}).ToSet(); len(got) != 0 {
		t.Fatalf("empty patch ToSet() = %v, want empty", got)
	}
	if got := (DevicePatch{}).CameraToSet(); len(got) != 0 {
		t.Fatalf("empty patch CameraToSet() = %v, want empty", got)
	}
}

func TestDeviceValidate(t *testing.T) {
	float := func(v float64) *float64 { return &v }
	valid := Device{Key: "cam-1", OrganisationId: "org-1", Type: DeviceTypeCamera}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid device: %v", err)
	}

	tests := []struct {
		name   string
		device Device
		want   error
	}{
		{"missing key", Device{OrganisationId: "org-1"}, ErrDeviceKeyMissing},
		{"missing organisation", Device{Key: "cam-1"}, ErrDeviceOrganisationMissing},
		{"unknown type", Device{Key: "cam-1", OrganisationId: "org-1", Type: "toaster"}, ErrDeviceTypeInvalid},
	}
	for _, tt := range tests {
		if err := tt.device.Validate(); !errors.Is(err, tt.want) {
			t.Errorf("%s: Validate() = %v, want %v", tt.name, err, tt.want)
		}
	}

	invalid := []struct {
		name   string
		device Device
	}{
		{"inverted lux range", Device{Key: "k", OrganisationId: "o", LightSensor: &LightSensorMetadata{MinLux: float(10), MaxLux: float(1)}}},
		{"inverted humidity range", Device{Key: "k", OrganisationId: "o", HumiditySensor: &HumiditySensorMetadata{MinHumidity: float(90), MaxHumidity: float(10)}}},
	}
	for _, tt := range invalid {
		if err := tt.device.Validate(); err == nil {
			t.Errorf("%s: Validate() = nil, want an error", tt.name)
		}
	}
}

// Signed URLs are filled at read time and returned to clients, never stored.
func TestDeviceMediaRefRuntimeMetadataIsNotStored(t *testing.T) {
	device := Device{LatestMedia: DeviceMediaRef{
		VideoFile:         "org/cam-1/video.mp4",
		AtRuntimeMetadata: &MediaAtRuntimeMetadata{VideoUrl: "https://signed.example/video.mp4"},
	}}
	raw, err := bson.Marshal(device)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := bson.Raw(raw).LookupErr("latestMedia", "atRuntimeMetadata"); err == nil {
		t.Fatal("latestMedia.atRuntimeMetadata is written to BSON")
	}
	out, err := json.Marshal(device)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), `"atRuntimeMetadata":{"videoUrl":"https://signed.example/video.mp4"}`) {
		t.Fatalf("JSON does not carry the signed URL: %s", out)
	}
}

// JSON names that existing clients read must not change with the model.
func TestDeviceLegacyJSONNames(t *testing.T) {
	out, err := json.Marshal(Device{MutedUntil: 1706000000, IsActive: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{`"mute":1706000000`, `"isActive":true`} {
		if !strings.Contains(string(out), name) {
			t.Errorf("JSON does not contain %s: %s", name, out)
		}
	}
}

// A device without media decodes to a zero LatestMedia that is safe to read.
func TestDeviceWithoutLatestMediaIsSafeToRead(t *testing.T) {
	raw, err := bson.Marshal(bson.M{"key": "cam-1"})
	if err != nil {
		t.Fatal(err)
	}
	var device Device
	if err := bson.Unmarshal(raw, &device); err != nil {
		t.Fatal(err)
	}
	if device.LatestMedia.VideoFile != "" || !device.LatestMedia.Id.IsZero() {
		t.Fatalf("LatestMedia = %+v, want zero value", device.LatestMedia)
	}
	device.LatestMedia.AtRuntimeMetadata = &MediaAtRuntimeMetadata{VideoUrl: "https://signed.example/video.mp4"}
}

// A device document as stored today (abridged from production): legacy
// permission names and a latestMedia written by the media pipeline without _id.
func TestDeviceDecodesStoredDocument(t *testing.T) {
	stored := bson.M{
		"key":            "kRhhZAtFGeIHGQSxhQCzZfTkpEEROV",
		"organisationId": "6754047347321302010cafee",
		"featurePermissions": bson.M{
			"ptz": int32(1), "liveview": int32(2), "remote_config": int32(3),
			"io": int32(4), "floorPlans": int32(2), "download_recordings": int32(1),
		},
		"latestMedia": bson.M{
			"startTimestamp":  int64(1791961138),
			"duration":        int32(29715),
			"deviceKey":       "kRhhZAtFGeIHGQSxhQCzZfTkpEEROV",
			"storageSolution": "kstorage",
			"videoFile":       "uugai/1791961138_3-639_rtsp_28-179-165-461_2096_29715.mp4",
			"videoProvider":   "archive",
			"sourceVaultId":   "vault-edge-bijlokehof",
			"metadata":        bson.M{"fps": 25.0, "fileSize": int64(1570263)},
		},
		"latestMediaTimestamp": int64(1791961138),
		"status":               "active",
		"ownershipGeneration":  int64(1),
	}
	raw, err := bson.Marshal(stored)
	if err != nil {
		t.Fatal(err)
	}
	var device Device
	if err := bson.Unmarshal(raw, &device); err != nil {
		t.Fatal(err)
	}

	want := &DeviceFeaturePermissions{PTZ: 1, Liveview: 2, RemoteConfig: 3, IO: 4, FloorPlans: 2, DownloadRecordings: 1}
	if !reflect.DeepEqual(device.FeaturePermissions, want) {
		t.Fatalf("FeaturePermissions = %+v, want %+v", device.FeaturePermissions, want)
	}
	if device.LatestMedia.VideoFile == "" || device.LatestMedia.VideoProvider != "archive" || device.LatestMedia.StorageSolution != "kstorage" {
		t.Fatalf("LatestMedia = %+v", device.LatestMedia)
	}
	if !device.LatestMedia.Id.IsZero() {
		t.Fatalf("LatestMedia.Id = %s, want zero: the pipeline does not store it", device.LatestMedia.Id.Hex())
	}

	// Permissions are written back under the stored names.
	out, err := bson.Marshal(Device{FeaturePermissions: want})
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"remote_config", "download_recordings"} {
		if _, err := bson.Raw(out).LookupErr("featurePermissions", key); err != nil {
			t.Errorf("featurePermissions.%s is not written: %v", key, err)
		}
	}
}

// The agent's cloud public key routes MQTT traffic per device, so clients
// receive it; it is never a private key.
func TestDeviceAgentFields(t *testing.T) {
	device := Device{
		Agent:   &DeviceAgent{CloudPublicKey: "AKIA5V6EGLUWSI42UQFH", CameraName: "rtsp"},
		Monitor: &DeviceMonitor{Status: "active", ChangedAt: 1791961313000},
	}
	raw, err := bson.Marshal(device)
	if err != nil {
		t.Fatal(err)
	}
	doc := bson.Raw(raw)
	if got := doc.Lookup("agent", "cloudPublicKey").StringValue(); got != "AKIA5V6EGLUWSI42UQFH" {
		t.Fatalf("agent.cloudPublicKey = %q", got)
	}
	if got := doc.Lookup("agent", "cameraName").StringValue(); got != "rtsp" {
		t.Fatalf("agent.cameraName = %q", got)
	}
	if got := doc.Lookup("monitor", "status").StringValue(); got != "active" {
		t.Fatalf("monitor.status = %q", got)
	}
	out, err := json.Marshal(device)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), `"cloudPublicKey":"AKIA5V6EGLUWSI42UQFH"`) {
		t.Fatalf("JSON does not carry the agent cloud public key: %s", out)
	}
}

func fullHeartbeat() DeviceHeartbeat {
	return DeviceHeartbeat{
		ReceivedAt: 1791961313000,
		Type:       DeviceTypeCamera,
		Agent: DeviceAgent{
			CloudPublicKey: "AKIA5V6EGLUWSI42UQFH",
			CameraName:     "rtsp",
			Version:        "v3.12.9",
			ReleaseHash:    "abc",
			Repository:     "kerberos-io/agent",
			Deployment:     "docker",
			Hostname:       "agent-rtsp",
			Architecture:   "x86_64",
			IPAddresses:    []string{"10.42.0.200/24"},
			MacAddresses:   []string{"6e:a1:b2:26:71:b6"},
			StartedAt:      1791925313000,
			BootTime:       1791356513000,
			MemoryTotal:    32171696128,
			MemoryFree:     1772171264,
			MemoryUsed:     30399524864,
			MemoryProcess:  117075968,
			HubEncryption:  true,
		},
		Camera: &CameraMetadata{
			Resolution:   "1920x1080",
			Connected:    true,
			HasPanTilt:   true,
			OnvifPresets: []CameraPreset{{Name: "door", Token: "t1"}},
			IOs:          []IO{{Key: "input-1"}},
			// User-owned fields: must never be written by a heartbeat.
			FieldOfView: 90,
			Presets:     []CameraPreset{{Name: "saved"}},
			Tours:       []CameraTour{{Name: "patrol"}},
		},
	}
}

// bsonFieldPaths lists prefix.<bson name> for every stored field of a struct.
func bsonFieldPaths(prefix string, value interface{}) []string {
	var paths []string
	structType := reflect.TypeOf(value)
	for i := 0; i < structType.NumField(); i++ {
		name := strings.Split(structType.Field(i).Tag.Get("bson"), ",")[0]
		if name != "" && name != "-" {
			paths = append(paths, prefix+name)
		}
	}
	return paths
}

// Every heartbeat-owned path is either written by a heartbeat or
// deliberately user-owned, so a new field cannot silently go unwritten.
func TestDeviceHeartbeatCoversAgentAndCameraFields(t *testing.T) {
	set := fullHeartbeat().FullSet()
	for _, path := range bsonFieldPaths("agent.", DeviceAgent{}) {
		if _, ok := set[path]; !ok {
			t.Errorf("FullSet does not write %s", path)
		}
	}
	userOwned := map[string]bool{"camera.fieldOfView": true, "camera.presets": true, "camera.tours": true}
	for _, path := range bsonFieldPaths("camera.", CameraMetadata{}) {
		_, written := set[path]
		if userOwned[path] && written {
			t.Errorf("FullSet writes user-owned %s", path)
		}
		if !userOwned[path] && !written {
			t.Errorf("FullSet does not write agent-reported %s", path)
		}
	}
}

// The ownership rule: heartbeats and user edits never write the same path.
func TestDeviceHeartbeatAndPatchPathsAreDisjoint(t *testing.T) {
	text := "x"
	flag := true
	when := time.Now()
	number := 1.0
	floorPlans := []string{"plan"}
	patch := DevicePatch{
		Name: &text, Color: &text, Brand: &text, Model: &text, Description: &text,
		Hidden: &flag, LastMaintenance: &when, InstallationDate: &when,
		OnFloorPlans: &floorPlans, FieldOfView: &number, FeaturePermissions: &DeviceFeaturePermissions{},
		Location: &Location{Address: text},
		Labels:   &floorPlans,
	}
	userPaths := map[string]bool{
		// Written by the preset endpoints rather than DevicePatch.
		"camera.presets": true,
		"camera.tours":   true,
	}
	for path := range patch.ToSet() {
		userPaths[path] = true
	}
	for path := range patch.CameraToSet() {
		userPaths[path] = true
	}

	heartbeat := fullHeartbeat()
	for path := range heartbeat.ProvisionSet() {
		for userPath := range userPaths {
			if path == userPath || strings.HasPrefix(path, userPath+".") || strings.HasPrefix(userPath, path+".") {
				t.Errorf("heartbeat path %s overlaps user path %s", path, userPath)
			}
		}
	}
}

func TestDeviceHeartbeatRefreshSet(t *testing.T) {
	heartbeat := fullHeartbeat()
	heartbeat.Agent.StartedAt = 0
	heartbeat.Agent.BootTime = 0
	want := map[string]interface{}{
		"agentLastSeen":       int64(1791961313000),
		"deviceLastSeen":      int64(1791961313000),
		"agent.memoryFree":    int64(1772171264),
		"agent.memoryUsed":    int64(30399524864),
		"agent.memoryProcess": int64(117075968),
	}
	if got := heartbeat.RefreshSet(); !reflect.DeepEqual(got, want) {
		t.Fatalf("RefreshSet() = %v, want %v", got, want)
	}

	// A disconnected camera keeps its DeviceLastSeen.
	heartbeat.Camera.Connected = false
	if _, ok := heartbeat.RefreshSet()["deviceLastSeen"]; ok {
		t.Fatal("RefreshSet advances deviceLastSeen for a disconnected camera")
	}
	heartbeat.Camera.Connected = true

	// A detected restart writes the derived times.
	heartbeat.Agent.StartedAt = 1791925313000
	set := heartbeat.RefreshSet()
	if set["agent.startedAt"] != int64(1791925313000) {
		t.Fatalf("agent.startedAt = %v", set["agent.startedAt"])
	}
	if _, ok := set["agent.bootTime"]; ok {
		t.Fatal("RefreshSet writes agent.bootTime without a detected reboot")
	}
}

func TestDeviceHeartbeatRefreshSetIsSubsetOfFullSet(t *testing.T) {
	heartbeat := fullHeartbeat()
	full := heartbeat.FullSet()
	for path, value := range heartbeat.RefreshSet() {
		if !reflect.DeepEqual(full[path], value) {
			t.Errorf("FullSet[%s] = %v, RefreshSet has %v", path, full[path], value)
		}
	}
	if full["type"] != DeviceTypeCamera {
		t.Fatalf("type = %v", full["type"])
	}
	// The heartbeat leaves user-edited fields on legacy paths until the
	// backfill, so it must not claim the document is in the current shape.
	if _, stamped := full["schemaVersion"]; stamped {
		t.Fatal("FullSet stamps schemaVersion; the backfill would skip this device")
	}
}

func TestDeviceHeartbeatWithoutCameraOrType(t *testing.T) {
	heartbeat := fullHeartbeat()
	heartbeat.Camera = nil
	heartbeat.Type = ""
	for path := range heartbeat.FullSet() {
		if path == "type" || strings.HasPrefix(path, "camera") {
			t.Errorf("FullSet writes %s without a camera or type", path)
		}
	}
}

func TestDeviceHeartbeatProvisionSet(t *testing.T) {
	set := fullHeartbeat().ProvisionSet()
	if got, want := set["audit.createdAt"], time.UnixMilli(1791961313000).UTC(); got != want {
		t.Fatalf("audit.createdAt = %v, want %v", got, want)
	}
	if _, ok := set["name"]; ok {
		t.Fatal("ProvisionSet sets name; display falls back to agent.cameraName")
	}
	if set["schemaVersion"] != DeviceSchemaVersion {
		t.Fatalf("schemaVersion = %v, want %d for a new device", set["schemaVersion"], DeviceSchemaVersion)
	}
}

func TestDeviceThresholdsValidate(t *testing.T) {
	if err := (DeviceThresholds{Idle: 180_000, Disconnected: 180_000}).Validate(); err != nil {
		t.Fatalf("valid thresholds: %v", err)
	}
	for _, thresholds := range []DeviceThresholds{{Idle: -1}, {Disconnected: -1}} {
		if err := thresholds.Validate(); err == nil {
			t.Errorf("Validate(%+v) = nil, want an error", thresholds)
		}
	}
}

func TestDeviceConnectionStatus(t *testing.T) {
	const now = int64(1791961313000)
	camera := &CameraMetadata{}
	tests := []struct {
		name       string
		device     Device
		thresholds DeviceThresholds
		want       DeviceConnectionStatus
	}{
		{"never seen", Device{Camera: camera}, DeviceThresholds{}, DeviceConnectionDisconnected},
		{"agent and camera fresh", Device{Camera: camera, AgentLastSeen: now - 10_000, DeviceLastSeen: now - 10_000}, DeviceThresholds{}, DeviceConnectionConnected},
		{"agent gone", Device{Camera: camera, AgentLastSeen: now - 181_000, DeviceLastSeen: now - 181_000}, DeviceThresholds{}, DeviceConnectionDisconnected},
		{"camera unreachable", Device{Camera: camera, AgentLastSeen: now - 10_000, DeviceLastSeen: now - 181_000}, DeviceThresholds{}, DeviceConnectionIdle},
		{"camera never reachable", Device{Camera: camera, AgentLastSeen: now - 10_000}, DeviceThresholds{}, DeviceConnectionIdle},
		{"at the threshold", Device{Camera: camera, AgentLastSeen: now - 180_000, DeviceLastSeen: now - 180_000}, DeviceThresholds{}, DeviceConnectionConnected},
		{"no camera is never idle", Device{AgentLastSeen: now - 10_000}, DeviceThresholds{}, DeviceConnectionConnected},
		{"custom thresholds", Device{Camera: camera, AgentLastSeen: now - 30_000, DeviceLastSeen: now - 30_000}, DeviceThresholds{Idle: 20_000, Disconnected: 60_000}, DeviceConnectionIdle},
	}
	for _, tt := range tests {
		if got := tt.device.ConnectionStatus(now, tt.thresholds); got != tt.want {
			t.Errorf("%s: ConnectionStatus() = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestDevicePatchLocation(t *testing.T) {
	location := &Location{Address: "Bijlokehof 1, Gent", Latitude: 51.05, Longitude: 3.72}
	set := DevicePatch{Location: location}.ToSet()
	if !reflect.DeepEqual(set, map[string]interface{}{"location": location}) {
		t.Fatalf("ToSet() = %v", set)
	}

	// The API body uses the models field names.
	var patch DevicePatch
	if err := json.Unmarshal([]byte(`{"location":{"address":"Bijlokehof 1, Gent","latitude":51.05,"longitude":3.72}}`), &patch); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(patch.Location, location) {
		t.Fatalf("decoded location = %+v, want %+v", patch.Location, location)
	}
}

// Older documents can hold a location in another shape (e.g. a Google Places
// result); decoding must not fail on them.
func TestDeviceDecodesForeignLocationShape(t *testing.T) {
	raw, err := bson.Marshal(bson.M{
		"key": "cam-1",
		"location": bson.M{
			"formatted_address": "Bijlokehof 1, Gent",
			"geometry":          bson.M{"location": bson.M{"lat": 51.05, "lng": 3.72}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	var device Device
	if err := bson.Unmarshal(raw, &device); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if device.Location == nil || device.Location.Address != "" {
		t.Fatalf("Location = %+v, want an empty location", device.Location)
	}
}
