package models

import (
	"encoding/json"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestMediaMetadataFPSAcceptsHistoricalIntegers(t *testing.T) {
	var fromJSON MediaMetadata
	if err := json.Unmarshal([]byte(`{"fps":18}`), &fromJSON); err != nil {
		t.Fatalf("decode integer JSON FPS: %v", err)
	}
	if fromJSON.FPS != 18 {
		t.Fatalf("integer JSON FPS = %v, want 18", fromJSON.FPS)
	}

	encoded, err := bson.Marshal(bson.M{"fps": int32(18)})
	if err != nil {
		t.Fatalf("encode historical BSON FPS: %v", err)
	}
	var fromBSON MediaMetadata
	if err := bson.Unmarshal(encoded, &fromBSON); err != nil {
		t.Fatalf("decode historical integer BSON FPS: %v", err)
	}
	if fromBSON.FPS != 18 {
		t.Fatalf("historical BSON FPS = %v, want 18", fromBSON.FPS)
	}
}

func TestMediaMetadataFPSPreservesFraction(t *testing.T) {
	const want = 17.35

	encoded, err := bson.Marshal(MediaMetadata{FPS: want})
	if err != nil {
		t.Fatalf("encode fractional BSON FPS: %v", err)
	}
	var decoded MediaMetadata
	if err := bson.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("decode fractional BSON FPS: %v", err)
	}
	if decoded.FPS != want {
		t.Fatalf("fractional BSON FPS = %v, want %v", decoded.FPS, want)
	}
}
