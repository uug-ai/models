package api

import (
	"encoding/json"
	"testing"
)

func TestDetectionBoxInputTimestampMsDistinguishesZeroFromOmitted(t *testing.T) {
	var omitted DetectionBoxInput
	if err := json.Unmarshal([]byte(`{"frame":0}`), &omitted); err != nil {
		t.Fatalf("unmarshal omitted timestamp: %v", err)
	}
	if omitted.TimestampMs != nil {
		t.Fatalf("omitted timestampMs = %d, want nil", *omitted.TimestampMs)
	}

	var zero DetectionBoxInput
	if err := json.Unmarshal([]byte(`{"frame":0,"timestampMs":0}`), &zero); err != nil {
		t.Fatalf("unmarshal zero timestamp: %v", err)
	}
	if zero.TimestampMs == nil || *zero.TimestampMs != 0 {
		t.Fatalf("explicit timestampMs = %v, want pointer to zero", zero.TimestampMs)
	}
}
