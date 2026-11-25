package models

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func TestSupersededZipsAreStoredButNeverServed(t *testing.T) {
	at := time.Date(2026, 11, 24, 15, 0, 0, 0, time.UTC)
	task := Task{ExportSupersededZips: []TaskSupersededZip{
		{File: "owner/cases/1/zips/1_1700_0-0_user_0-0-0-0_0_0.zip", Provider: "archive", SupersededAt: at, Attempts: 1},
	}}

	raw, err := bson.Marshal(task)
	if err != nil {
		t.Fatalf("bson.Marshal() error = %v", err)
	}
	var stored Task
	if err := bson.Unmarshal(raw, &stored); err != nil {
		t.Fatalf("bson.Unmarshal() error = %v", err)
	}
	if len(stored.ExportSupersededZips) != 1 || stored.ExportSupersededZips[0] != task.ExportSupersededZips[0] {
		t.Fatalf("export_superseded_zips did not round-trip: %+v", stored.ExportSupersededZips)
	}

	served, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if strings.Contains(string(served), "superseded") || strings.Contains(string(served), "zips/") {
		t.Errorf("superseded zips sent to clients: %s", served)
	}
}

func TestTaskWithoutSupersededZipsOmitsTheField(t *testing.T) {
	raw, err := bson.Marshal(Task{})
	if err != nil {
		t.Fatalf("bson.Marshal() error = %v", err)
	}
	if _, err := bson.Raw(raw).LookupErr("export_superseded_zips"); err == nil {
		t.Errorf("export_superseded_zips stored on a task without any")
	}
}
