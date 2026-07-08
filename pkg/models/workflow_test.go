package models

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func TestWorkflow_EffectiveSourceAndIsGlobal(t *testing.T) {
	// Empty Source defaults to user and is never global.
	w := Workflow{}
	if got := w.EffectiveSource(); got != WorkflowSourceUser {
		t.Fatalf("empty Source should default to user, got %q", got)
	}
	if w.IsGlobal() {
		t.Fatal("a user workflow must not be global")
	}
	// A config workflow with no org is global.
	cfg := Workflow{Source: WorkflowSourceConfig}
	if !cfg.IsGlobal() {
		t.Fatal("a config workflow with empty org should be global")
	}
	// A config workflow pinned to an org is not global.
	scoped := Workflow{Source: WorkflowSourceConfig, OrganisationId: "org-1"}
	if scoped.IsGlobal() {
		t.Fatal("a config workflow scoped to an org must not be global")
	}
}

func TestWorkflow_SourceJSONOmitsWhenEmpty(t *testing.T) {
	b, err := json.Marshal(Workflow{Name: "w"})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(b), `"source"`) {
		t.Fatalf("empty Source should be omitted, got %s", b)
	}
	b, err = json.Marshal(Workflow{Name: "w", Source: WorkflowSourceConfig})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(b), `"source":"config"`) {
		t.Fatalf("expected source in %s", b)
	}
}

func TestWorkflow_CompileStages_FromGraph(t *testing.T) {
	// classify --(unconditional)--> anpr --(conditional)--> notify
	w := Workflow{
		Nodes: []WorkflowNode{
			{Id: "n1", StageRef: "classify"},
			{Id: "n2", StageRef: "anpr"},
			{Id: "n3", StageRef: "notify"},
		},
		Edges: []WorkflowEdge{
			{Id: "e1", Source: "n1", Target: "n2"},
			{Id: "e2", Source: "n2", Target: "n3", Condition: &StageCondition{Path: "results.anpr.tracks", Op: ConditionOpExists}},
		},
	}
	stages := w.CompileStages()
	if len(stages) != 3 {
		t.Fatalf("expected 3 stages, got %d", len(stages))
	}
	byOp := map[string]WorkflowStage{}
	for _, s := range stages {
		byOp[s.Operation] = s
	}
	// Root node has no incoming edges -> always.
	if byOp["classify"].Dispatch != DispatchAlways {
		t.Fatalf("classify should dispatch always, got %q", byOp["classify"].Dispatch)
	}
	// anpr has an unconditional incoming edge -> conditional with a gated, nil-condition need.
	anpr := byOp["anpr"]
	if anpr.Dispatch != DispatchConditional || len(anpr.Needs) != 1 {
		t.Fatalf("anpr should be conditional with one need, got %+v", anpr)
	}
	if anpr.Needs[0].Operation != "classify" || anpr.Needs[0].Condition != nil {
		t.Fatalf("anpr need should gate on classify with no condition, got %+v", anpr.Needs[0])
	}
	// notify has a conditional incoming edge -> its condition is projected onto the need.
	notify := byOp["notify"]
	if notify.Dispatch != DispatchConditional || len(notify.Needs) != 1 {
		t.Fatalf("notify should be conditional with one need, got %+v", notify)
	}
	if notify.Needs[0].Operation != "anpr" || notify.Needs[0].Condition == nil || notify.Needs[0].Condition.Path != "results.anpr.tracks" {
		t.Fatalf("notify need should gate on anpr with the edge condition, got %+v", notify.Needs[0])
	}
}

func TestWorkflow_CompileStages_PrefersStoredStages(t *testing.T) {
	stored := []WorkflowStage{{Operation: "loitering", Dispatch: DispatchAlways}}
	w := Workflow{
		Stages: stored,
		// A graph that would compile differently must be ignored when Stages is set.
		Nodes: []WorkflowNode{{Id: "n1", StageRef: "somethingelse"}},
	}
	stages := w.CompileStages()
	if len(stages) != 1 || stages[0].Operation != "loitering" {
		t.Fatalf("stored Stages should be returned as-is, got %+v", stages)
	}
}

func TestWorkflow_AutomaticMatches(t *testing.T) {
	loc, err := time.LoadLocation("Europe/Brussels")
	if err != nil {
		t.Fatalf("load location: %v", err)
	}
	wed := time.Date(2024, 1, 3, 9, 0, 0, 0, loc)
	w := Workflow{
		Enabled: true,
		Triggers: []WorkflowTrigger{
			// A manual trigger must never activate automatically.
			{Type: WorkflowTriggerManual, Surfaces: []WorkflowTriggerSurface{WorkflowSurfaceCase}},
			// An automatic trigger scoped to cam-1.
			{Type: WorkflowTriggerAutomatic, Devices: []DeviceKey{{Key: "cam-1"}}},
		},
	}
	if !w.AutomaticMatches("cam-1", wed) {
		t.Fatal("expected in-scope device to activate the automatic trigger")
	}
	if w.AutomaticMatches("cam-9", wed) {
		t.Fatal("out-of-scope device must not activate")
	}
	// A disabled workflow never activates, even with a matching trigger.
	disabled := w
	disabled.Enabled = false
	if disabled.AutomaticMatches("cam-1", wed) {
		t.Fatal("a disabled workflow must not activate")
	}
	// A manual-only workflow never activates automatically.
	manualOnly := Workflow{Enabled: true, Triggers: []WorkflowTrigger{
		{Type: WorkflowTriggerManual, Surfaces: []WorkflowTriggerSurface{WorkflowSurfaceCase}},
	}}
	if manualOnly.AutomaticMatches("cam-1", wed) {
		t.Fatal("a manual-only workflow must not activate automatically")
	}
}

// TestWorkflow_EffectiveID_GoldenVectors pins the deterministic id derivation so
// it cannot drift. Every service that reads WORKFLOW_DEFINITIONS (hub-api, the
// workflows engine) relies on this exact scheme to agree on a config workflow's
// identity without persisting it; if these fixed name→id vectors ever change,
// runs one service seeds would stop resolving to the workflow another loaded.
func TestWorkflow_EffectiveID_GoldenVectors(t *testing.T) {
	vectors := map[string]string{
		"tracking-workflow":        "ec7e72bc4cedd32018e1f92a",
		"objecttracking-on-demand": "1d07bbe3a979822d7a60b857",
	}
	for name, want := range vectors {
		got := (&Workflow{Name: name}).EffectiveID().Hex()
		if got != want {
			t.Fatalf("EffectiveID(%q) = %q, want %q (id scheme changed?)", name, got, want)
		}
	}
}

// TestWorkflow_EffectiveID_PrefersExplicitID confirms a workflow that already
// carries an id keeps it rather than deriving one from its name — so persisted
// user workflows resolve by their stored id.
func TestWorkflow_EffectiveID_PrefersExplicitID(t *testing.T) {
	id := primitive.NewObjectID()
	w := Workflow{Id: id, Name: "tracking-workflow"}
	if got := w.EffectiveID(); got != id {
		t.Fatalf("EffectiveID should return the explicit id %s, got %s", id.Hex(), got.Hex())
	}
}

