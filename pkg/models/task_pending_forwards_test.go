package models

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestPendingForwardsRoundTripAndHideTheIds(t *testing.T) {
	at := time.Date(2026, 11, 25, 10, 0, 0, 0, time.UTC)
	id := primitive.NewObjectID()
	task := Task{ExportPendingForwards: &TaskPendingForwards{
		Media: []primitive.ObjectID{id}, Count: 1, RequestedAt: at, Until: at.Add(24 * time.Hour),
	}}

	raw, err := bson.Marshal(task)
	if err != nil {
		t.Fatalf("bson.Marshal() error = %v", err)
	}
	var stored Task
	if err := bson.Unmarshal(raw, &stored); err != nil {
		t.Fatalf("bson.Unmarshal() error = %v", err)
	}
	got := stored.ExportPendingForwards
	if got == nil || len(got.Media) != 1 || got.Media[0] != id || got.Count != 1 || !got.Until.Equal(at.Add(24*time.Hour)) {
		t.Fatalf("export_pending_forwards did not round-trip: %+v", got)
	}

	served, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if !strings.Contains(string(served), `"export_pending_forwards":{"count":1`) {
		t.Errorf("JSON misses the count: %s", served)
	}
	if strings.Contains(string(served), id.Hex()) {
		t.Errorf("case_media ids sent to clients: %s", served)
	}
}

func TestTaskWithoutPendingForwardsOmitsTheField(t *testing.T) {
	served, err := json.Marshal(Task{})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(served), "export_pending_forwards") {
		t.Errorf("export_pending_forwards served without pending forwards: %s", served)
	}
}
