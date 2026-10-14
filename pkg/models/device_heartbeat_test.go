package models

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func fullHeartbeat() DeviceHeartbeat {
	return DeviceHeartbeat{
		ReceivedAt:      1791961313000,
		DeviceReachable: true,
		Type:            DeviceTypeCamera,
		Agent: DeviceAgent{
			CloudPublicKey: "AKIA5V6EGLUWSI42UQFH",
			CameraName:     "rtsp",
			Enterprise:     true,
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

	// An unreachable device keeps its DeviceLastSeen.
	heartbeat.DeviceReachable = false
	if _, ok := heartbeat.RefreshSet()["deviceLastSeen"]; ok {
		t.Fatal("RefreshSet advances deviceLastSeen for an unreachable device")
	}

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
	if full["schemaVersion"] != DeviceSchemaVersion || full["type"] != DeviceTypeCamera {
		t.Fatalf("schemaVersion/type = %v / %v", full["schemaVersion"], full["type"])
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
}
