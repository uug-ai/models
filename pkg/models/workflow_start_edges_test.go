package models

import (
	"reflect"
	"testing"
	"time"
)

func edgeWorkflow() Workflow {
	return Workflow{
		Enabled: true,
		Nodes: []WorkflowNode{
			{Id: "start", Type: WorkflowNodeStart, Trigger: &WorkflowTrigger{Type: WorkflowTriggerAutomatic}},
			{Id: "a", StageRef: "a"}, {Id: "b", StageRef: "b"}, {Id: "join", StageRef: "join"},
		},
		Edges: []WorkflowEdge{
			{Id: "cars", Source: "start", Target: "a", Conditions: []WorkflowCondition{{Path: "inputs.classify.objectCount", Op: ConditionOpGt, Value: 1}}},
			{Id: "camera", Source: "start", Target: "b", Conditions: []WorkflowCondition{{Path: "device.deviceKey", Op: ConditionOpEq, Value: "camera"}}},
			{Id: "a-join", Source: "a", Target: "join"}, {Id: "b-join", Source: "b", Target: "join"},
		},
	}
}

func TestAutomaticStartEdgeSelection(t *testing.T) {
	for _, tc := range []struct {
		name          string
		count         int
		camera        string
		unconditional bool
		want          []string
	}{
		{"none", 0, "other", false, nil},
		{"one", 2, "other", false, []string{"cars"}},
		{"multiple", 2, "camera", false, []string{"cars", "camera"}},
		{"unconditional", 0, "other", true, []string{"cars"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := edgeWorkflow()
			if tc.unconditional {
				w.Edges[0].Conditions = nil
			}
			root := AutomaticTriggerRootWithInputs(WorkflowDevice{DeviceKey: tc.camera}, WorkflowUser{},
				map[string]any{"classify": map[string]any{"objectCount": tc.count}})
			match, err := w.MatchAutomaticTrigger(root, time.Now())
			if err != nil {
				t.Fatal(err)
			}
			if len(tc.want) == 0 {
				if match != nil || w.AutomaticMatches(root, time.Now()) {
					t.Fatal("unmatched workflow activated")
				}
				return
			}
			if match == nil || !reflect.DeepEqual(match.MatchedEdgeIds, tc.want) ||
				match.Trigger.EdgeId != tc.want[0] {
				t.Fatalf("match = %+v", match)
			}
			stages, err := BindStartEdges(w.CompileStages(), match)
			if err != nil {
				t.Fatal(err)
			}
			for _, stage := range stages[:2] {
				want := stage.Operation == "a" || len(tc.want) == 2
				if stage.MatchesDependencies(root, nil) != want {
					t.Fatalf("stage %+v want %v", stage, want)
				}
			}
			// Replacing the envelope with matching worker data cannot open b.
			root["device"] = map[string]any{"deviceKey": "camera"}
			if stages[1].MatchesDependencies(root, map[string]bool{"a": true}) != (len(tc.want) == 2) {
				t.Fatal("late result changed entry selection")
			}
			w.Edges[0].Id = "edited"
			if match.MatchedEdgeIds[0] != "cars" {
				t.Fatal("snapshot changed after edit")
			}
		})
	}
}

func TestAutomaticStartSharedGatesAndLegacyGroups(t *testing.T) {
	w := edgeWorkflow()
	start := &w.Nodes[0]
	start.Devices = []DeviceKey{{Key: "camera"}}
	start.Trigger.SiteIds, start.Trigger.GroupIds = []string{"site"}, []string{"group"}
	start.Trigger.WeeklySchedule = []*WeeklySchedule{{Day: 1, Enabled: true, Timezone: "UTC", Segments: []DayTimeRange{{Start: 3600, End: 7200}}}}
	start.Trigger.ConditionMode = ConditionModeAny
	start.Trigger.Conditions = []WorkflowCondition{
		{Path: "device.deviceName", Op: ConditionOpEq, Value: "front"},
		{Path: "device.deviceName", Op: ConditionOpEq, Value: "back"},
	}
	w.Edges[0].ConditionMode = ConditionModeAny
	w.Edges[0].Conditions = append(w.Edges[0].Conditions, WorkflowCondition{Path: "device.deviceName", Op: ConditionOpEq, Value: "never"})
	at := time.Date(2026, 9, 28, 1, 30, 0, 0, time.UTC)
	device := WorkflowDevice{DeviceKey: "camera", DeviceName: "front", SiteIds: []string{"site"}, GroupIds: []string{"group"}}
	root := AutomaticTriggerRootWithInputs(device, WorkflowUser{}, map[string]any{"classify": map[string]any{"objectCount": 3}})
	match, err := w.MatchAutomaticTrigger(root, at)
	if err != nil || match == nil || len(match.MatchedEdgeIds) != 2 {
		t.Fatalf("match %+v: %v", match, err)
	}
	if match.Trigger.SharedConditions == nil || match.Trigger.SharedConditions.ConditionMode != ConditionModeAny {
		t.Fatal("legacy any-group was flattened")
	}
	for name, mutate := range map[string]func(){
		"device":            func() { device.DeviceKey = "other" },
		"site":              func() { device.SiteIds = nil },
		"group":             func() { device.GroupIds = nil },
		"shared predicates": func() { device.DeviceName = "other" },
	} {
		t.Run(name, func(t *testing.T) {
			saved := device
			mutate()
			root := AutomaticTriggerRootWithInputs(device, WorkflowUser{}, map[string]any{"classify": map[string]any{"objectCount": 3}})
			if w.AutomaticMatches(root, at) {
				t.Fatal("shared gate bypassed")
			}
			device = saved
		})
	}
	if w.AutomaticMatches(root, at.Add(2*time.Hour)) {
		t.Fatal("schedule bypassed")
	}
}

func TestAutomaticStartNoRunnablePathAndStaleStages(t *testing.T) {
	w := edgeWorkflow()
	w.Stages = []WorkflowStage{{Operation: "stale"}}
	if got := w.CompileStages(); len(got) != 3 || got[0].Operation != "a" {
		t.Fatalf("stale stages won: %+v", got)
	}
	w.Edges = nil
	w.SyncGraphTriggers()
	if len(w.Triggers) != 0 || len(w.CompileStages()) != 0 || w.AutomaticMatches(nil, time.Now()) {
		t.Fatal("disconnected stages activated")
	}
	w.Edges = []WorkflowEdge{{Id: "dangling", Source: "start", Target: "missing"}}
	if w.AutomaticMatches(nil, time.Now()) {
		t.Fatal("dangling Start edge activated")
	}
	legacy := Workflow{Enabled: true, Stages: []WorkflowStage{{Operation: "legacy"}}, Triggers: []WorkflowTrigger{{}}}
	if !legacy.AutomaticMatches(nil, time.Now()) || legacy.CompileStages()[0].Operation != "legacy" {
		t.Fatal("stage-only behavior changed")
	}
}

func TestStartAlternativesPreserveOrdinaryJoins(t *testing.T) {
	yes, no := true, false
	stage := WorkflowStage{NeedsMode: NeedsModeAll, Needs: []StageDependency{
		{StartEdgeId: "one", StartMatched: &yes}, {StartEdgeId: "two", StartMatched: &no},
		{Operation: "a"}, {Operation: "b"},
	}}
	if stage.MatchesDependencies(nil, map[string]bool{"a": true}) {
		t.Fatal("join ignored missing b")
	}
	if !stage.MatchesDependencies(nil, map[string]bool{"a": true, "b": true}) {
		t.Fatal("entry alternatives combined with AND")
	}
	stage.Needs[0].StartMatched = &no
	if stage.MatchesDependencies(nil, map[string]bool{"a": true, "b": true}) {
		t.Fatal("ordinary results opened unmatched entry")
	}
	stage.Needs = stage.Needs[2:]
	if !stage.MatchesDependencies(nil, map[string]bool{"a": true, "b": true}) {
		t.Fatal("ordinary all join changed")
	}
}

func TestAutomaticStartEdgeValidation(t *testing.T) {
	for name, mutate := range map[string]func(*Workflow){
		"empty id":              func(w *Workflow) { w.Edges[0].Id = "" },
		"duplicate id":          func(w *Workflow) { w.Edges[1].Id = w.Edges[0].Id },
		"future result":         func(w *Workflow) { w.Edges[0].Conditions[0].Path = "results.a.count" },
		"mixed predicate forms": func(w *Workflow) { w.Edges[0].Condition = &WorkflowCondition{Path: "key", Op: ConditionOpExists} },
	} {
		t.Run(name, func(t *testing.T) {
			w := edgeWorkflow()
			mutate(&w)
			if w.ValidateTriggers() == nil || w.ValidateGraph() == nil {
				t.Fatal("invalid Start edge accepted")
			}
		})
	}
	w := edgeWorkflow()
	w.Nodes[0].Trigger = &WorkflowTrigger{Type: WorkflowTriggerManual, Surfaces: []WorkflowTriggerSurface{WorkflowSurfaceMedia}}
	w.Edges[0].Conditions[0].Path = "results.a.count"
	if err := w.ValidateGraph(); err != nil {
		t.Fatal(err)
	}
	if w.AutomaticMatches(nil, time.Now()) {
		t.Fatal("manual edge activated automatically")
	}
	if w.CompileStages()[0].Needs[0].StartEdgeId != "" {
		t.Fatal("manual routing frozen as activation")
	}
}
