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
		AtRuntimeMetadata: &DeviceAtRuntimeMetadata{Status: DeviceConnectionHealthy},
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
		{"negative threshold", Device{Key: "k", OrganisationId: "o", Thresholds: &DeviceThresholds{Idle: -1}}},
		{"healthy above disconnected", Device{Key: "k", OrganisationId: "o", Thresholds: &DeviceThresholds{Healthy: 10, Disconnected: 5}}},
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
