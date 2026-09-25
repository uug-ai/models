package models

import (
	"encoding/json"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMediaMetadataVLMRoundTripsBSON(t *testing.T) {
	metadata := MediaMetadata{
		Description: "Person walks to the gate",
		VLM: &VLMMediaMetadata{
			SchemaVersion: VLMSchemaVersion,
			Scene:         VLMScene{Summary: "Driveway with a gate", StaticElements: []string{"gate"}},
			Observations:  []VLMObservation{{Marker: "person", Event: "walking", Tags: []string{"red", "sweater"}}},
		},
	}

	encoded, err := bson.Marshal(metadata)
	if err != nil {
		t.Fatalf("encode media metadata: %v", err)
	}
	var raw bson.M
	if err := bson.Unmarshal(encoded, &raw); err != nil {
		t.Fatalf("decode raw media metadata: %v", err)
	}
	vlm, ok := raw["vlm"].(bson.M)
	if !ok {
		t.Fatalf("metadata.vlm = %#v, want document", raw["vlm"])
	}
	if _, ok := vlm["observations"]; !ok {
		t.Fatalf("metadata.vlm.observations missing: %#v", vlm)
	}

	var decoded MediaMetadata
	if err := bson.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("decode media metadata: %v", err)
	}
	if decoded.VLM == nil || decoded.VLM.Scene.Summary != "Driveway with a gate" {
		t.Fatalf("decoded VLM = %#v", decoded.VLM)
	}
	if got := decoded.VLM.Observations[0]; got.Marker != "person" || got.Event != "walking" || len(got.Tags) != 2 {
		t.Fatalf("decoded observation = %#v", got)
	}
}

func TestMetadataWithoutVLMOmitsField(t *testing.T) {
	media, err := bson.Marshal(MediaMetadata{Description: "legacy"})
	if err != nil {
		t.Fatalf("encode media metadata: %v", err)
	}
	device, err := bson.Marshal(DeviceMetadata{Brand: "legacy"})
	if err != nil {
		t.Fatalf("encode device metadata: %v", err)
	}
	for name, encoded := range map[string][]byte{"media": media, "device": device} {
		var raw bson.M
		if err := bson.Unmarshal(encoded, &raw); err != nil {
			t.Fatalf("decode %s metadata: %v", name, err)
		}
		if _, ok := raw["vlm"]; ok {
			t.Fatalf("%s metadata wrote vlm for legacy document: %#v", name, raw)
		}
	}
}

func TestDeviceMetadataVLMStableStatesRoundTripBSON(t *testing.T) {
	sourceMediaId := primitive.NewObjectID()
	metadata := DeviceMetadata{
		VLM: &DeviceVLMMetadata{
			SchemaVersion: VLMSchemaVersion,
			StableStates: []VLMStableState{{
				Id:            primitive.NewObjectID(),
				Name:          "Daytime",
				Scene:         VLMScene{Summary: "Empty driveway"},
				SourceMediaId: sourceMediaId,
				UpdatedAt:     1700000000,
			}},
		},
	}

	encoded, err := bson.Marshal(metadata)
	if err != nil {
		t.Fatalf("encode device metadata: %v", err)
	}
	var decoded DeviceMetadata
	if err := bson.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("decode device metadata: %v", err)
	}
	if decoded.VLM == nil || len(decoded.VLM.StableStates) != 1 {
		t.Fatalf("decoded device VLM = %#v", decoded.VLM)
	}
	if got := decoded.VLM.StableStates[0]; got.Name != "Daytime" || got.SourceMediaId != sourceMediaId {
		t.Fatalf("decoded stable state = %#v", got)
	}
}

func TestVLMContextJSONOmitsMissingMedia(t *testing.T) {
	encoded, err := json.Marshal(VLMContext{StableStates: []VLMStableState{{Name: "Night"}}})
	if err != nil {
		t.Fatalf("encode VLM context: %v", err)
	}
	var raw map[string]any
	if err := json.Unmarshal(encoded, &raw); err != nil {
		t.Fatalf("decode VLM context: %v", err)
	}
	if _, ok := raw["media"]; ok {
		t.Fatalf("VLM context wrote media without analysis: %s", encoded)
	}
	if states, ok := raw["stableStates"].([]any); !ok || len(states) != 1 {
		t.Fatalf("VLM context stableStates = %#v", raw["stableStates"])
	}
}
