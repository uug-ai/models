package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// VLMSchemaVersion identifies the structure of VLM metadata stored on media
// (metadata.vlm) and devices (metadata.vlm).
const VLMSchemaVersion = 1

// VLMMaxStableStates bounds the stable states embedded on a single device.
// Promoting a new state beyond this limit drops the oldest entry.
const VLMMaxStableStates = 32

// VLMScene describes the persistent, mostly static content of a camera view as
// observed by a vision-language model. It is compared against a device's stable
// states so that future analysis reports what changed instead of repeating the
// scene.
type VLMScene struct {
	// Summary is a short description of the scene, e.g. "Parking lot with two parked cars".
	Summary string `json:"summary" bson:"summary"`

	// StaticElements lists stable objects in view, e.g. "gate", "bench".
	StaticElements []string `json:"staticElements,omitempty" bson:"staticElements,omitempty"`

	// Zones lists named areas in view, e.g. "entrance", "driveway".
	Zones []string `json:"zones,omitempty" bson:"zones,omitempty"`

	// Lighting describes the lighting condition, e.g. "daylight", "infrared".
	Lighting string `json:"lighting,omitempty" bson:"lighting,omitempty"`
}

// VLMObservation is a single action or change detected by a vision-language
// model. It maps onto a marker with category "vlm": Marker is the marker name,
// Event the marker event and Tags the marker tags.
type VLMObservation struct {
	// Marker is the subject of the observation, e.g. "person", "car".
	Marker string `json:"marker" bson:"marker"`

	// Event is the action or change, e.g. "walking", "entering".
	Event string `json:"event" bson:"event"`

	// Tags are short attributes of the subject, e.g. "red", "sweater".
	Tags []string `json:"tags,omitempty" bson:"tags,omitempty"`
}

// VLMMediaMetadata is the VLM analysis result stored on a media document at
// metadata.vlm.
type VLMMediaMetadata struct {
	SchemaVersion int              `json:"schemaVersion" bson:"schemaVersion"`
	Scene         VLMScene         `json:"scene" bson:"scene"`
	Observations  []VLMObservation `json:"observations,omitempty" bson:"observations,omitempty"`
}

// VLMStableState is a named "normal" scene for a device, promoted from the VLM
// scene of a recording. Promoting the same recording again replaces its entry.
type VLMStableState struct {
	Id    primitive.ObjectID `json:"id" bson:"id"`
	Name  string             `json:"name" bson:"name"`
	Scene VLMScene           `json:"scene" bson:"scene"`

	// SourceMediaId and SourceMediaKey reference the recording the state was promoted from.
	SourceMediaId  primitive.ObjectID `json:"sourceMediaId" bson:"sourceMediaId"`
	SourceMediaKey string             `json:"sourceMediaKey,omitempty" bson:"sourceMediaKey,omitempty"`

	// UpdatedAt is the promotion time (seconds since Unix epoch).
	UpdatedAt int64 `json:"updatedAt" bson:"updatedAt"`
}

// DeviceVLMMetadata holds the stable states of a device at metadata.vlm. The
// StableStates array is bounded by VLMMaxStableStates.
type DeviceVLMMetadata struct {
	SchemaVersion int              `json:"schemaVersion" bson:"schemaVersion"`
	StableStates  []VLMStableState `json:"stableStates,omitempty" bson:"stableStates,omitempty"`
}

// VLMContext combines the VLM analysis of a recording with the stable states
// of the device that produced it.
type VLMContext struct {
	Media        *VLMMediaMetadata `json:"media,omitempty" bson:"media,omitempty"`
	StableStates []VLMStableState  `json:"stableStates,omitempty" bson:"stableStates,omitempty"`
}
