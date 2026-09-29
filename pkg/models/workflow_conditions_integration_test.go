package models

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func TestWorkflowTrigger_ConditionGroupCannotBypassScope(t *testing.T) {
	trigger := WorkflowTrigger{
		Devices:       []DeviceKey{{Key: "cam-1"}, {Key: "cam-2"}},
		SiteIds:       []string{"site-1", "site-2"},
		GroupIds:      []string{"group-1", "group-2"},
		ConditionMode: ConditionModeAny,
		Conditions: []WorkflowCondition{
			{Path: "device.deviceName", Op: ConditionOpEq, Value: "Front"},
			{Path: "device.deviceName", Op: ConditionOpEq, Value: "Back"},
		},
	}
	if err := trigger.Validate(); err != nil {
		t.Fatal(err)
	}
	device := WorkflowDevice{
		DeviceKey: "cam-2", DeviceName: "Back",
		SiteIds: []string{"other-site", "site-2"}, GroupIds: []string{"group-1"},
	}
	for _, tc := range []struct {
		name   string
		mutate func(*WorkflowDevice)
		want   bool
	}{
		{"matching alternative in every category", func(*WorkflowDevice) {}, true},
		{"wrong device", func(d *WorkflowDevice) { d.DeviceKey = "cam-3" }, false},
		{"wrong site", func(d *WorkflowDevice) { d.SiteIds = []string{"other-site"} }, false},
		{"missing group", func(d *WorkflowDevice) { d.GroupIds = nil }, false},
		{"wrong group", func(d *WorkflowDevice) { d.GroupIds = []string{"group-3"} }, false},
		{"failed predicates", func(d *WorkflowDevice) { d.DeviceName = "Side" }, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			candidate := device
			tc.mutate(&candidate)
			if got := trigger.MatchesEnvelope(AutomaticTriggerRoot(candidate, WorkflowUser{})); got != tc.want {
				t.Fatalf("MatchesEnvelope = %v, want %v", got, tc.want)
			}
		})
	}
	trigger.Conditions = nil
	if !trigger.MatchesEnvelope(AutomaticTriggerRoot(device, WorkflowUser{})) {
		t.Fatal("empty any group must impose no additional restriction")
	}
	device.DeviceKey = "unselected"
	if trigger.MatchesEnvelope(AutomaticTriggerRoot(device, WorkflowUser{})) {
		t.Fatal("empty any group must still honor scope")
	}
}

func TestWorkflowTrigger_InputConditionsAndSchedule(t *testing.T) {
	trigger := WorkflowTrigger{
		ConditionMode: ConditionModeAny,
		Conditions: []WorkflowCondition{
			{Path: "inputs.classify.objectCount", Op: ConditionOpGt, Value: 0},
			{Path: "device.deviceKey", Op: ConditionOpEq, Value: "another-camera"},
		},
		WeeklySchedule: []*WeeklySchedule{{
			Enabled: true, Day: int(time.Monday), Timezone: "UTC",
			Segments: []DayTimeRange{{Start: 9 * 3600, End: 17 * 3600}},
		}},
	}
	if err := trigger.Validate(); err != nil {
		t.Fatal(err)
	}
	inputs := map[string]any{"classify": map[string]any{"objectCount": 1}}
	root := AutomaticTriggerRootWithInputs(WorkflowDevice{DeviceKey: "cam-1"}, WorkflowUser{}, inputs)
	monday := time.Date(2026, time.June, 1, 12, 0, 0, 0, time.UTC)
	if !trigger.Matches(root, monday) || trigger.Matches(root, monday.Add(8*time.Hour)) {
		t.Fatal("condition group must be ANDed with active recording hours")
	}
	withoutInputs := AutomaticTriggerRoot(WorkflowDevice{DeviceKey: "cam-1"}, WorkflowUser{})
	if trigger.Matches(withoutInputs, monday) {
		t.Fatal("future/missing inputs must not be fabricated")
	}
	for _, forbidden := range []string{"results", "storage", "runId"} {
		if _, exists := root[forbidden]; exists {
			t.Fatalf("pre-run root contains %q", forbidden)
		}
	}
	if !reflect.DeepEqual(root["inputs"], inputs) {
		t.Fatal("supplied sanitized inputs were not exposed unchanged")
	}
	workflow := Workflow{Enabled: true, Triggers: []WorkflowTrigger{trigger, trigger}}
	if !workflow.AutomaticMatches(root, monday) {
		t.Fatal("matching alternative triggers should activate once via a boolean gate")
	}
}

func TestWorkflowTrigger_Validation(t *testing.T) {
	for _, tc := range []struct {
		name    string
		trigger WorkflowTrigger
		valid   bool
	}{
		{"default automatic", WorkflowTrigger{}, true},
		{"manual surfaces", WorkflowTrigger{Type: WorkflowTriggerManual, Surfaces: []WorkflowTriggerSurface{WorkflowSurfaceCase, WorkflowSurfaceMedia, WorkflowSurfaceRedaction}}, true},
		{"manual without surface", WorkflowTrigger{Type: WorkflowTriggerManual}, false},
		{"unknown surface", WorkflowTrigger{Type: WorkflowTriggerManual, Surfaces: []WorkflowTriggerSurface{"unsupported"}}, false},
		{"unknown type", WorkflowTrigger{Type: "scheduled"}, false},
		{"empty device", WorkflowTrigger{Devices: []DeviceKey{{Key: " "}}}, false},
		{"empty site", WorkflowTrigger{SiteIds: []string{""}}, false},
		{"empty group", WorkflowTrigger{GroupIds: []string{""}}, false},
		{"invalid mode", WorkflowTrigger{ConditionMode: "not"}, false},
		{"unknown op", WorkflowTrigger{Conditions: []WorkflowCondition{{Path: "device.deviceKey", Op: "unknown"}}}, false},
		{"future results", WorkflowTrigger{Conditions: []WorkflowCondition{{Path: "results.classify.objectCount", Op: ConditionOpExists}}}, false},
		{"storage credentials", WorkflowTrigger{Conditions: []WorkflowCondition{{Path: "user.storage", Op: ConditionOpExists}}}, false},
		{"manual ignores legacy automatic fields", WorkflowTrigger{
			Type: WorkflowTriggerManual, Surfaces: []WorkflowTriggerSurface{WorkflowSurfaceMedia},
			Devices: []DeviceKey{{Key: ""}}, ConditionMode: "ignored",
			Conditions: []WorkflowCondition{{Path: "legacy.path", Op: "ignored"}},
		}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.trigger.Validate(); (err == nil) != tc.valid {
				t.Fatalf("Validate = %v, valid = %v", err, tc.valid)
			}
		})
	}
	invalid := WorkflowTrigger{Type: "unknown"}
	workflow := Workflow{Trigger: &invalid}
	if workflow.ValidateTriggers() == nil || workflow.Trigger == nil {
		t.Fatal("legacy trigger must be validated without mutation")
	}
	workflow.Triggers = []WorkflowTrigger{{}}
	if err := workflow.ValidateTriggers(); err != nil || workflow.Trigger == nil {
		t.Fatal("canonical list must win without rewriting legacy data")
	}
}

func TestWorkflow_CompilePluralEdgeConditions(t *testing.T) {
	w := vlmEditorWorkflow()
	w.Edges[0].Condition = nil
	w.Edges[0].ConditionMode = ConditionModeAny
	w.Edges[0].Conditions = []WorkflowCondition{
		{Path: "inputs.classify.objectCount", Op: ConditionOpGt, Value: 0},
		{Path: "device.deviceName", Op: ConditionOpEq, Value: "Front"},
	}
	w.Edges[1].ConditionMode = ConditionModeAll
	w.Edges[1].Conditions = []WorkflowCondition{
		{Path: "results.vlm.description", Op: ConditionOpExists},
	}
	if err := w.ValidateGraph(); err != nil {
		t.Fatal(err)
	}
	stages := w.CompileStages()
	for i, operation := range []string{WorkflowDeviceGateOperation, "vlm"} {
		stage := stages[i]
		if stage.Dispatch != DispatchConditional || len(stage.Needs) != 1 {
			t.Fatalf("compiled stage = %+v", stage)
		}
		need := stage.Needs[0]
		if need.Operation != operation || need.ConditionMode != w.Edges[i].ConditionMode ||
			!reflect.DeepEqual(need.Conditions, w.Edges[i].Conditions) || need.Condition != nil {
			t.Fatalf("edge conditions/readiness lost: %+v", need)
		}
		if stage.NeedsMode != "" {
			t.Fatal("per-edge condition mode must not change default fan-in")
		}
	}
	w.Edges[0].Conditions = nil
	if stage := w.CompileStages()[0]; stage.Dispatch != DispatchAlways || len(stage.Needs) != 0 {
		t.Fatal("empty device edge group should still start the run without a gate")
	}
	w.Edges[0].ConditionMode = "invalid"
	if err := w.ValidateGraph(); !errors.Is(err, ErrInvalidWorkflowGraph) {
		t.Fatalf("invalid mode not rejected: %v", err)
	}
	w.Edges[0].ConditionMode = ""
	w.Edges[0].Condition = &StageCondition{Path: "device.deviceKey", Op: ConditionOpExists}
	w.Edges[0].Conditions = []WorkflowCondition{}
	if err := w.ValidateGraph(); !errors.Is(err, ErrInvalidWorkflowGraph) {
		t.Fatalf("ambiguous legacy plus explicit empty plural not rejected: %v", err)
	}
}

func TestWorkflow_ConditionContractRoundTrip(t *testing.T) {
	workflow := vlmEditorWorkflow()
	workflow.Triggers = []WorkflowTrigger{{
		Devices: []DeviceKey{{Key: "cam-1"}}, SiteIds: []string{"site-1"}, GroupIds: []string{"group-1"},
		ConditionMode: ConditionModeAll,
		Conditions: []WorkflowCondition{{
			Path: "inputs.classify.details", Op: ConditionOpAnyMatch,
			Match: &WorkflowPredicateSet{Conditions: []WorkflowPredicate{
				{Path: "classified", Op: ConditionOpEq, Value: "pedestrian"},
			}},
		}},
	}}
	workflow.Edges[0].Condition = nil
	workflow.Edges[0].ConditionMode = ConditionModeAny
	workflow.Edges[0].Conditions = workflow.Triggers[0].Conditions
	for _, tc := range []struct {
		name   string
		encode func(any) ([]byte, error)
		decode func([]byte, any) error
	}{
		{"JSON", json.Marshal, json.Unmarshal},
		{"BSON", bson.Marshal, bson.Unmarshal},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data, err := tc.encode(workflow)
			if err != nil {
				t.Fatal(err)
			}
			var decoded Workflow
			if err := tc.decode(data, &decoded); err != nil {
				t.Fatal(err)
			}
			if err := decoded.ValidateGraph(); err != nil {
				t.Fatal(err)
			}
			if err := decoded.ValidateTriggers(); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(decoded.Triggers, workflow.Triggers) || !reflect.DeepEqual(decoded.Edges, workflow.Edges) {
				t.Fatal("condition contract changed during serialization")
			}
			root := AutomaticTriggerRootWithInputs(WorkflowDevice{
				DeviceKey: "cam-1", SiteIds: []string{"site-1"}, GroupIds: []string{"group-1"},
			}, WorkflowUser{}, map[string]any{
				"classify": bson.M{"details": bson.A{bson.M{"classified": "pedestrian"}}},
			})
			if !decoded.AutomaticMatches(root, time.Now()) {
				t.Fatal("roundtripped trigger did not match")
			}
			need := decoded.CompileStages()[0].Needs[0]
			if need.Matches(root, nil) || !need.Matches(root, map[string]bool{"classify": true}) {
				t.Fatal("roundtripped edge did not honor readiness and conditions")
			}
		})
	}
}
