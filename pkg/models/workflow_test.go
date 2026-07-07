package models

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func TestWorkflowTrigger_EffectiveType(t *testing.T) {
	if got := (WorkflowTrigger{}).EffectiveType(); got != WorkflowTriggerAutomatic {
		t.Fatalf("empty Type should default to automatic, got %q", got)
	}
	if got := (WorkflowTrigger{Type: WorkflowTriggerManual}).EffectiveType(); got != WorkflowTriggerManual {
		t.Fatalf("explicit Type should be preserved, got %q", got)
	}
}

func TestWorkflowTrigger_HasSurface(t *testing.T) {
	trg := WorkflowTrigger{Type: WorkflowTriggerManual, Surfaces: []WorkflowTriggerSurface{WorkflowSurfaceCase}}
	if !trg.HasSurface(WorkflowSurfaceCase) {
		t.Fatal("expected case surface to match")
	}
	if trg.HasSurface(WorkflowSurfaceMedia) {
		t.Fatal("did not expect media surface to match")
	}
}

func TestWorkflow_NormalizeTriggers_FoldsLegacySingle(t *testing.T) {
	w := Workflow{Trigger: &WorkflowTrigger{Devices: []DeviceKey{{Key: "cam-1"}}}}
	w.NormalizeTriggers()
	if w.Trigger != nil {
		t.Fatal("legacy Trigger should be cleared after normalize")
	}
	if len(w.Triggers) != 1 || len(w.Triggers[0].Devices) != 1 || w.Triggers[0].Devices[0].Key != "cam-1" {
		t.Fatalf("legacy trigger should fold into Triggers, got %+v", w.Triggers)
	}
	// Idempotent.
	w.NormalizeTriggers()
	if len(w.Triggers) != 1 {
		t.Fatalf("normalize should be idempotent, got %d triggers", len(w.Triggers))
	}
}

func TestWorkflow_NormalizeTriggers_PrefersListOverLegacy(t *testing.T) {
	w := Workflow{
		Triggers: []WorkflowTrigger{{Type: WorkflowTriggerManual, Surfaces: []WorkflowTriggerSurface{WorkflowSurfaceCase}}},
		Trigger:  &WorkflowTrigger{Devices: []DeviceKey{{Key: "legacy"}}},
	}
	w.NormalizeTriggers()
	if w.Trigger != nil {
		t.Fatal("legacy Trigger should be dropped when Triggers is populated")
	}
	if len(w.Triggers) != 1 || w.Triggers[0].Type != WorkflowTriggerManual {
		t.Fatalf("existing Triggers should win over legacy, got %+v", w.Triggers)
	}
}

func TestWorkflow_ManualTriggersForSurface(t *testing.T) {
	w := Workflow{Triggers: []WorkflowTrigger{
		{Type: WorkflowTriggerAutomatic, Devices: []DeviceKey{{Key: "cam-1"}}},
		{Type: WorkflowTriggerManual, Surfaces: []WorkflowTriggerSurface{WorkflowSurfaceMedia}},
		{Type: WorkflowTriggerManual, Surfaces: []WorkflowTriggerSurface{WorkflowSurfaceCase}},
	}}
	got := w.ManualTriggersForSurface(WorkflowSurfaceCase)
	if len(got) != 1 {
		t.Fatalf("expected exactly one manual/case trigger, got %d", len(got))
	}
	// A legacy automatic-only workflow yields none.
	legacy := Workflow{Trigger: &WorkflowTrigger{Devices: []DeviceKey{{Key: "cam-1"}}}}
	if got := legacy.ManualTriggersForSurface(WorkflowSurfaceCase); len(got) != 0 {
		t.Fatalf("legacy automatic-only workflow should have no manual triggers, got %d", len(got))
	}
}

// Legacy workflow_runs / workflows stored a single "trigger" subdocument. Decoding
// such a document must still populate the deprecated field so NormalizeTriggers
// can fold it forward rather than silently dropping the data on re-save.
func TestWorkflow_LegacyTriggerBSONRoundTrip(t *testing.T) {
	legacyDoc := bson.M{
		"name":    "legacy",
		"enabled": true,
		"trigger": bson.M{"type": "automatic", "devices": bson.A{bson.M{"key": "cam-1"}}},
	}
	raw, err := bson.Marshal(legacyDoc)
	if err != nil {
		t.Fatalf("marshal legacy doc: %v", err)
	}
	var w Workflow
	if err := bson.Unmarshal(raw, &w); err != nil {
		t.Fatalf("unmarshal legacy doc: %v", err)
	}
	if w.Trigger == nil || len(w.Trigger.Devices) != 1 || w.Trigger.Devices[0].Key != "cam-1" {
		t.Fatalf("legacy trigger should decode into deprecated field, got %+v", w.Trigger)
	}
	w.NormalizeTriggers()
	if len(w.Triggers) != 1 || len(w.Triggers[0].Devices) != 1 || w.Triggers[0].Devices[0].Key != "cam-1" {
		t.Fatalf("legacy trigger should fold forward on normalize, got %+v", w.Triggers)
	}
}

func TestWorkflowTrigger_JSONOmitsEmptyMode(t *testing.T) {
	b, err := json.Marshal(WorkflowTrigger{Type: WorkflowTriggerManual, Surfaces: []WorkflowTriggerSurface{WorkflowSurfaceCase}})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(b)
	if want := `"type":"manual"`; !strings.Contains(s, want) {
		t.Fatalf("expected %s in %s", want, s)
	}
	if want := `"surfaces":["case"]`; !strings.Contains(s, want) {
		t.Fatalf("expected %s in %s", want, s)
	}
	// Automatic-only scoping fields must not leak when unset.
	if strings.Contains(s, `"devices"`) {
		t.Fatalf("unset devices should be omitted, got %s", s)
	}
	if strings.Contains(s, `"weeklySchedule"`) {
		t.Fatalf("unset weeklySchedule should be omitted, got %s", s)
	}
}

func TestWorkflowTrigger_MatchesDevice(t *testing.T) {
	// Empty Devices matches every device.
	if !(WorkflowTrigger{}).MatchesDevice("cam-9") {
		t.Fatal("empty device list should match any device")
	}
	trg := WorkflowTrigger{Devices: []DeviceKey{{Key: "cam-1"}, {Key: "cam-2"}}}
	if !trg.MatchesDevice("cam-2") {
		t.Fatal("expected listed device to match")
	}
	if trg.MatchesDevice("cam-9") {
		t.Fatal("did not expect unlisted device to match")
	}
}

func TestWorkflowTrigger_IsScheduledAt(t *testing.T) {
	// Empty schedule matches any time.
	if !(WorkflowTrigger{}).IsScheduledAt(time.Now()) {
		t.Fatal("empty schedule should match any time")
	}
	loc, err := time.LoadLocation("Europe/Brussels")
	if err != nil {
		t.Fatalf("load location: %v", err)
	}
	// Wednesday 2024-01-03, 09:00 local. time.Weekday: Wednesday == 3.
	wed := time.Date(2024, 1, 3, 9, 0, 0, 0, loc)
	trg := WorkflowTrigger{WeeklySchedule: []*WeeklySchedule{{
		Day:      int(time.Wednesday),
		Enabled:  true,
		Timezone: "Europe/Brussels",
		Segments: []DayTimeRange{{Start: 8 * 3600, End: 18 * 3600}},
	}}}
	if !trg.IsScheduledAt(wed) {
		t.Fatal("expected in-window Wednesday time to match")
	}
	// Same clock time on Thursday should not match the Wednesday schedule.
	thu := time.Date(2024, 1, 4, 9, 0, 0, 0, loc)
	if trg.IsScheduledAt(thu) {
		t.Fatal("did not expect Thursday to match a Wednesday-only schedule")
	}
	// Outside the daily window on the right day should not match.
	wedEarly := time.Date(2024, 1, 3, 7, 0, 0, 0, loc)
	if trg.IsScheduledAt(wedEarly) {
		t.Fatal("did not expect pre-window time to match")
	}
}

func TestWorkflowTrigger_Matches(t *testing.T) {
	loc, err := time.LoadLocation("Europe/Brussels")
	if err != nil {
		t.Fatalf("load location: %v", err)
	}
	wed := time.Date(2024, 1, 3, 9, 0, 0, 0, loc)
	trg := WorkflowTrigger{
		Devices: []DeviceKey{{Key: "cam-1"}},
		WeeklySchedule: []*WeeklySchedule{{
			Day:      int(time.Wednesday),
			Enabled:  true,
			Timezone: "Europe/Brussels",
			Segments: []DayTimeRange{{Start: 8 * 3600, End: 18 * 3600}},
		}},
	}
	if !trg.Matches("cam-1", wed) {
		t.Fatal("expected in-scope device and time to match")
	}
	if trg.Matches("cam-9", wed) {
		t.Fatal("wrong device should not match even when time is in window")
	}
	if trg.Matches("cam-1", wed.Add(24*time.Hour)) {
		t.Fatal("right device but out-of-schedule time should not match")
	}
	// A bare automatic trigger (no scoping) matches everything.
	if !(WorkflowTrigger{}).Matches("anything", wed) {
		t.Fatal("empty automatic trigger should match any device/time")
	}
}
