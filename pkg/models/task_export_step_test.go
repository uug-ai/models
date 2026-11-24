package models

import (
	"encoding/json"
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestTaskExportStepRoundTrips(t *testing.T) {
	task := Task{ExportStep: &TaskExportStep{Name: ExportStepArchiving, Done: 80, Total: 426, Skipped: 217}}

	raw, err := bson.Marshal(task)
	if err != nil {
		t.Fatalf("bson.Marshal() error = %v", err)
	}
	var stored Task
	if err := bson.Unmarshal(raw, &stored); err != nil {
		t.Fatalf("bson.Unmarshal() error = %v", err)
	}
	if stored.ExportStep == nil || *stored.ExportStep != *task.ExportStep {
		t.Fatalf("export_step did not round-trip: %+v", stored.ExportStep)
	}

	served, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if !strings.Contains(string(served), `"export_step":{"name":"archiving","done":80,"total":426,"skipped":217}`) {
		t.Errorf("JSON misses export_step: %s", served)
	}
}

func TestTaskWithoutExportStepOmitsTheField(t *testing.T) {
	raw, err := bson.Marshal(Task{})
	if err != nil {
		t.Fatalf("bson.Marshal() error = %v", err)
	}
	if _, err := bson.Raw(raw).LookupErr("export_step"); err == nil {
		t.Errorf("export_step stored on a task without a step")
	}
	served, err := json.Marshal(Task{})
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if strings.Contains(string(served), "export_step") {
		t.Errorf("export_step served on a task without a step: %s", served)
	}
}
