package api

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/uug-ai/models/pkg/models"
)

func TestDetectionRunNameJSONContract(t *testing.T) {
	var request PostDetectionsRequest
	if err := json.Unmarshal([]byte(`{"name":"Front entrance"}`), &request); err != nil {
		t.Fatalf("unmarshal request: %v", err)
	}
	if request.Name != "Front entrance" {
		t.Fatalf("request name = %q, want %q", request.Name, "Front entrance")
	}

	encoded, err := json.Marshal(models.DetectionRun{Name: request.Name})
	if err != nil {
		t.Fatalf("marshal detection run: %v", err)
	}
	if !strings.Contains(string(encoded), `"name":"Front entrance"`) {
		t.Fatalf("encoded run %s does not contain name", encoded)
	}

	empty, err := json.Marshal(models.DetectionRun{})
	if err != nil {
		t.Fatalf("marshal empty detection run: %v", err)
	}
	var emptyObject map[string]json.RawMessage
	if err := json.Unmarshal(empty, &emptyObject); err != nil {
		t.Fatalf("unmarshal empty detection run: %v", err)
	}
	if _, exists := emptyObject["name"]; exists {
		t.Fatalf("empty run unexpectedly contains name: %s", empty)
	}
}