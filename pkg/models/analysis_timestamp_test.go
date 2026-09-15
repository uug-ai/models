package models

import (
	"encoding/json"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestTrackBoxTimestampMsPreservesExplicitZero(t *testing.T) {
	timestampMs := int64(0)
	box := TrackBox{TimestampMs: &timestampMs}

	jsonData, err := json.Marshal(box)
	if err != nil {
		t.Fatalf("marshal JSON: %v", err)
	}
	if string(jsonData) != `{"x1":0,"y1":0,"x2":0,"y2":0,"timestampMs":0,"smoothed":false,"edited":false}` {
		t.Fatalf("JSON = %s, want explicit timestampMs zero", jsonData)
	}

	bsonData, err := bson.Marshal(box)
	if err != nil {
		t.Fatalf("marshal BSON: %v", err)
	}
	var stored bson.M
	if err := bson.Unmarshal(bsonData, &stored); err != nil {
		t.Fatalf("unmarshal BSON: %v", err)
	}
	if got, ok := stored["timestampMs"]; !ok || got != int64(0) {
		t.Fatalf("BSON timestampMs = %#v (present %t), want int64(0)", got, ok)
	}
}

func TestTrackBoxTimestampMsOmittedWhenAbsent(t *testing.T) {
	jsonData, err := json.Marshal(TrackBox{})
	if err != nil {
		t.Fatalf("marshal JSON: %v", err)
	}

	var stored map[string]any
	if err := json.Unmarshal(jsonData, &stored); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	if _, ok := stored["timestampMs"]; ok {
		t.Fatalf("timestampMs should be omitted: %s", jsonData)
	}
}
