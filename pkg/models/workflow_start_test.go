package models

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func startWorkflow(mode WorkflowTriggerType) Workflow {
	w := vlmEditorWorkflow()
	w.Nodes[0].Type = WorkflowNodeStart
	w.Nodes[0].Trigger = &WorkflowTrigger{
		Type: mode, Surfaces: []WorkflowTriggerSurface{WorkflowSurfaceCase, WorkflowSurfaceMedia, WorkflowSurfaceRedaction},
		SiteIds: []string{"site"}, GroupIds: []string{"group"},
		Conditions: []WorkflowCondition{{Path: "device.deviceKey", Op: ConditionOpEq, Value: "cam-1"}},
	}
	w.Triggers = []WorkflowTrigger{{Type: WorkflowTriggerAutomatic}, {Type: WorkflowTriggerManual, Surfaces: []WorkflowTriggerSurface{WorkflowSurfaceCase}}}
	return w
}

func TestStartNodeRoundTripAndTriggerAuthority(t *testing.T) {
	for _, mode := range []WorkflowTriggerType{WorkflowTriggerAutomatic, WorkflowTriggerManual} {
		t.Run(string(mode), func(t *testing.T) {
			w := startWorkflow(mode)
			w.Nodes[0].Trigger.WeeklySchedule = []*WeeklySchedule{{
				Day: 1, Enabled: true, Timezone: "UTC", Segments: []DayTimeRange{{Start: 3600, End: 7200}},
			}}
			raw, err := json.Marshal(w)
			if err != nil {
				t.Fatal(err)
			}
			var decoded Workflow
			if err := json.Unmarshal(raw, &decoded); err != nil {
				t.Fatal(err)
			}
			data, err := bson.Marshal(decoded)
			if err != nil {
				t.Fatal(err)
			}
			var loaded Workflow
			if err := bson.Unmarshal(data, &loaded); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(w.Nodes[0], loaded.Nodes[0]) {
				t.Fatalf("node lost in JSON/BSON round trip: %#v", loaded.Nodes[0])
			}
			loaded.SyncGraphTriggers()
			first := append([]WorkflowTrigger(nil), loaded.Triggers...)
			loaded.SyncGraphTriggers()
			if !reflect.DeepEqual(first, loaded.Triggers) || len(first) != 1 || first[0].Type != mode {
				t.Fatalf("non-authoritative or non-idempotent triggers: %#v", loaded.Triggers)
			}
			if mode == WorkflowTriggerAutomatic {
				if len(first[0].Surfaces) != 0 || !reflect.DeepEqual(first[0].Conditions, w.Nodes[0].Trigger.Conditions) ||
					!reflect.DeepEqual(first[0].SiteIds, w.Nodes[0].Trigger.SiteIds) ||
					!reflect.DeepEqual(first[0].GroupIds, w.Nodes[0].Trigger.GroupIds) ||
					!reflect.DeepEqual(first[0].WeeklySchedule, w.Nodes[0].Trigger.WeeklySchedule) ||
					!reflect.DeepEqual(first[0].Devices, w.Nodes[0].Devices) {
					t.Fatalf("automatic trigger settings: %#v", first[0])
				}
			} else if len(first[0].Devices) != 0 || len(first[0].Conditions) != 0 || len(first[0].SiteIds) != 0 {
				t.Fatalf("inactive automatic settings leaked: %#v", first[0])
			}
			if len(loaded.Nodes[0].Trigger.Surfaces) != 3 || len(loaded.Nodes[0].Trigger.Conditions) != 1 {
				t.Fatal("inactive node settings discarded")
			}
		})
	}
}

func TestStartNodeModeSwitchAndDiscovery(t *testing.T) {
	w := startWorkflow(WorkflowTriggerManual)
	for _, surface := range []WorkflowTriggerSurface{WorkflowSurfaceCase, WorkflowSurfaceMedia, WorkflowSurfaceRedaction} {
		if len(w.ManualTriggersForSurface(surface)) != 1 {
			t.Fatalf("missing surface %s", surface)
		}
	}
	root := AutomaticTriggerRoot(WorkflowDevice{DeviceKey: "cam-1"}, WorkflowUser{})
	if w.AutomaticMatches(root, time.Now()) {
		t.Fatal("manual Start used stale automatic trigger")
	}
	w.Nodes[0].Trigger.Type = WorkflowTriggerAutomatic
	w.Nodes[0].Trigger.SiteIds, w.Nodes[0].Trigger.GroupIds = nil, nil
	if !w.AutomaticMatches(root, time.Now()) {
		t.Fatal("automatic Start did not activate")
	}
	for _, surface := range w.Nodes[0].Trigger.Surfaces {
		if len(w.ManualTriggersForSurface(surface)) != 0 {
			t.Fatalf("inactive surface advertised: %s", surface)
		}
	}
	w.Nodes[0].Trigger.Type = WorkflowTriggerManual
	if w.AutomaticMatches(root, time.Now()) {
		t.Fatal("switch back to manual retained automatic trigger")
	}
	w.Nodes[0].Trigger = nil
	if w.AutomaticMatches(root, time.Now()) || len(w.ManualTriggersForSurface(WorkflowSurfaceCase)) != 0 {
		t.Fatal("incomplete root used stale triggers")
	}
}

func TestStartNodeCompilesLikeDeviceRoot(t *testing.T) {
	device := vlmEditorWorkflow()
	start := startWorkflow(WorkflowTriggerManual)
	if !reflect.DeepEqual(device.CompileStages(), start.CompileStages()) {
		t.Fatal("Start outgoing classify gate changed")
	}
	start.Edges[0].Condition = nil
	if stage := start.CompileStages()[0]; stage.Dispatch != DispatchAlways || len(stage.Needs) != 0 {
		t.Fatalf("unconditional Start edge: %#v", stage)
	}
}

func TestStartNodeValidation(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Workflow)
		valid  bool
	}{
		{"manual", func(*Workflow) {}, true},
		{"automatic", func(w *Workflow) { w.Nodes[0].Trigger.Type = WorkflowTriggerAutomatic }, true},
		{"no trigger", func(w *Workflow) { w.Nodes[0].Trigger = nil }, false},
		{"empty type", func(w *Workflow) { w.Nodes[0].Trigger.Type = "" }, false},
		{"unknown type", func(w *Workflow) { w.Nodes[0].Trigger.Type = "scheduled" }, false},
		{"unknown surface", func(w *Workflow) { w.Nodes[0].Trigger.Surfaces = []WorkflowTriggerSurface{"unknown"} }, false},
		{"inactive unknown surface", func(w *Workflow) {
			w.Nodes[0].Trigger.Type = WorkflowTriggerAutomatic
			w.Nodes[0].Trigger.Surfaces = []WorkflowTriggerSurface{"unknown"}
		}, false},
		{"enabled no surfaces", func(w *Workflow) { w.Nodes[0].Trigger.Surfaces = nil }, false},
		{"draft no surfaces", func(w *Workflow) { w.Enabled = false; w.Nodes[0].Trigger.Surfaces = nil }, true},
		{"incoming edge", func(w *Workflow) {
			w.Edges = append(w.Edges, WorkflowEdge{Source: w.Nodes[1].Id, Target: w.Nodes[0].Id})
		}, false},
		{"second start", func(w *Workflow) { w.Nodes = append(w.Nodes, WorkflowNode{Id: "other", Type: WorkflowNodeStart}) }, false},
		{"second device", func(w *Workflow) { w.Nodes = append(w.Nodes, WorkflowNode{Id: "other", Type: WorkflowNodeDevice}) }, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := startWorkflow(WorkflowTriggerManual)
			tt.mutate(&w)
			if err := w.ValidateGraph(); (err == nil) != tt.valid {
				t.Fatalf("ValidateGraph() = %v, valid = %v", err, tt.valid)
			}
			if err := w.ValidateTriggers(); (err == nil) != tt.valid {
				t.Fatalf("ValidateTriggers() = %v, valid = %v", err, tt.valid)
			}
		})
	}
}
