package models

import "time"

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

	// DeviceReachable reports whether the agent could reach the device itself
	// (e.g. the camera stream is connected); it advances DeviceLastSeen.
	DeviceReachable bool `json:"deviceReachable" bson:"-"`

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
	if h.DeviceReachable {
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
	set["schemaVersion"] = DeviceSchemaVersion
	if h.Type != "" {
		set["type"] = h.Type
	}

	agent := h.Agent
	set["agent.cloudPublicKey"] = agent.CloudPublicKey
	set["agent.cameraName"] = agent.CameraName
	set["agent.enterprise"] = agent.Enterprise
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
// heartbeat: FullSet plus the creation time. Name is left empty so display
// falls back to Agent.CameraName until a user renames the device.
func (h DeviceHeartbeat) ProvisionSet() map[string]interface{} {
	set := h.FullSet()
	set["audit.createdAt"] = time.UnixMilli(h.ReceivedAt).UTC()
	return set
}

func setIfPositive(set map[string]interface{}, path string, value int64) {
	if value > 0 {
		set[path] = value
	}
}
