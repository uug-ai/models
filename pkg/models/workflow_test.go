package models

import (
	"encoding/json"
	"strings"
	"testing"

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
	w := Workflow{Trigger: &WorkflowTrigger{Selection: "cam-1"}}
	w.NormalizeTriggers()
	if w.Trigger != nil {
		t.Fatal("legacy Trigger should be cleared after normalize")
	}
	if len(w.Triggers) != 1 || w.Triggers[0].Selection != "cam-1" {
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
		Trigger:  &WorkflowTrigger{Selection: "legacy"},
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
		{Type: WorkflowTriggerAutomatic, Selection: "all"},
		{Type: WorkflowTriggerManual, Surfaces: []WorkflowTriggerSurface{WorkflowSurfaceMedia}},
		{Type: WorkflowTriggerManual, Surfaces: []WorkflowTriggerSurface{WorkflowSurfaceCase}},
	}}
	got := w.ManualTriggersForSurface(WorkflowSurfaceCase)
	if len(got) != 1 {
		t.Fatalf("expected exactly one manual/case trigger, got %d", len(got))
	}
	// A legacy automatic-only workflow yields none.
	legacy := Workflow{Trigger: &WorkflowTrigger{Selection: "cam-1"}}
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
		"trigger": bson.M{"selection": "cam-1", "startAt": "08:00"},
	}
	raw, err := bson.Marshal(legacyDoc)
	if err != nil {
		t.Fatalf("marshal legacy doc: %v", err)
	}
	var w Workflow
	if err := bson.Unmarshal(raw, &w); err != nil {
		t.Fatalf("unmarshal legacy doc: %v", err)
	}
	if w.Trigger == nil || w.Trigger.Selection != "cam-1" {
		t.Fatalf("legacy trigger should decode into deprecated field, got %+v", w.Trigger)
	}
	w.NormalizeTriggers()
	if len(w.Triggers) != 1 || w.Triggers[0].StartAt != "08:00" {
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
	// Automatic-only scheduling fields must not leak when unset.
	if strings.Contains(s, `"selection"`) {
		t.Fatalf("unset selection should be omitted, got %s", s)
	}
}
