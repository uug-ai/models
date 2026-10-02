package models

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func edgeSchedule(day int, start, end int64) []*WeeklySchedule {
	return []*WeeklySchedule{{Day: day, Enabled: true, Timezone: "UTC", Segments: []DayTimeRange{{Start: start, End: end}}}}
}

func TestStartEdgeTriggerScopeRoundTrip(t *testing.T) {
	for _, codec := range []struct {
		name      string
		marshal   func(any) ([]byte, error)
		unmarshal func([]byte, any) error
	}{
		{"json", json.Marshal, json.Unmarshal},
		{"bson", bson.Marshal, bson.Unmarshal},
	} {
		t.Run(codec.name, func(t *testing.T) {
			w := edgeWorkflow()
			w.Nodes[0].Devices = []DeviceKey{{Key: "legacy"}}
			w.Nodes[0].Trigger.Devices = []DeviceKey{{Key: "ignored-trigger-device"}}
			w.Nodes[0].Trigger.SiteIds = []string{"legacy-site"}
			w.Nodes[0].Trigger.GroupIds = []string{"legacy-group"}
			w.Nodes[0].Trigger.WeeklySchedule = edgeSchedule(1, 3600, 7200)
			w.Edges[0].Conditions, w.Edges[1].Conditions = nil, nil
			w.Edges[0].Trigger = &WorkflowEdgeTrigger{}
			raw, err := codec.marshal(w)
			if err != nil {
				t.Fatal(err)
			}
			var decoded Workflow
			if err := codec.unmarshal(raw, &decoded); err != nil {
				t.Fatal(err)
			}
			if decoded.Edges[0].Trigger == nil || decoded.Edges[1].Trigger != nil {
				t.Fatal("empty and absent edge triggers lost their distinct meaning")
			}
			at := time.Date(2026, 9, 29, 12, 0, 0, 0, time.UTC)
			match, err := decoded.MatchAutomaticTrigger(AutomaticTriggerRoot(WorkflowDevice{DeviceKey: "other"}, WorkflowUser{}), at)
			if err != nil || match == nil || !reflect.DeepEqual(match.MatchedEdgeIds, []string{"cars"}) {
				t.Fatalf("explicit empty scope did not override every legacy field: %+v, %v", match, err)
			}
			legacy := decoded.Triggers[1]
			if legacy.Devices[0].Key != "legacy" || !reflect.DeepEqual(legacy.SiteIds, []string{"legacy-site"}) ||
				!reflect.DeepEqual(legacy.GroupIds, []string{"legacy-group"}) || len(legacy.WeeklySchedule) != 1 {
				t.Fatalf("absent edge trigger lost legacy scope: %+v", legacy)
			}
			decoded.Edges[0].Trigger = &WorkflowEdgeTrigger{Devices: []DeviceKey{{Key: "new"}}}
			decoded.NormalizeTriggers()
			scope := decoded.Triggers[0]
			if scope.Devices[0].Key != "new" || len(scope.SiteIds)+len(scope.GroupIds)+len(scope.WeeklySchedule) != 0 {
				t.Fatalf("partial edge scope merged legacy fields: %+v", scope)
			}
			decoded.Triggers[0].Devices[0].Key = "runtime-edit"
			if decoded.Edges[0].Trigger.Devices[0].Key != "new" {
				t.Fatal("runtime projection mutated authored edge scope")
			}
			authored := &WorkflowEdgeTrigger{
				Devices: []DeviceKey{{Key: "new"}}, SiteIds: []string{"new-site"}, GroupIds: []string{"new-group"},
				Classifications: []string{"person", "car"},
				WeeklySchedule:  edgeSchedule(2, 7200, 10800),
			}
			decoded.Edges[0].Trigger = authored
			raw, err = codec.marshal(decoded)
			if err != nil {
				t.Fatal(err)
			}
			var populated Workflow
			if err := codec.unmarshal(raw, &populated); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(populated.Edges[0].Trigger, authored) {
				t.Fatal("populated edge source/schedule did not round-trip")
			}
		})
	}
}

func TestStartEdgeTriggerIndependentSourceSchedules(t *testing.T) {
	w := edgeWorkflow()
	w.Nodes[0].Devices = []DeviceKey{{Key: "obsolete"}}
	w.Nodes[0].Trigger.SiteIds = []string{"obsolete"}
	w.Nodes[0].Trigger.WeeklySchedule = edgeSchedule(0, 0, 1)
	w.Nodes[0].Trigger.Conditions = []WorkflowCondition{{Path: "device.provider", Op: ConditionOpEq, Value: "vault"}}
	for i, key := range []string{"a", "b"} {
		w.Edges[i].Trigger = &WorkflowEdgeTrigger{
			Devices: []DeviceKey{{Key: key}}, SiteIds: []string{"site-" + key}, GroupIds: []string{"group-" + key},
			WeeklySchedule: edgeSchedule(1+i, 3600, 7200),
		}
		w.Edges[i].ConditionMode = ConditionModeAny
		w.Edges[i].Conditions = []WorkflowCondition{
			{Path: "inputs.classify.objectCount", Op: ConditionOpGt, Value: 1},
			{Path: "device.deviceName", Op: ConditionOpEq, Value: "alternate"},
		}
	}
	for _, tc := range []struct {
		name, key, site, group, provider string
		day, count                       int
		want                             []string
	}{
		{"first branch", "a", "site-a", "group-a", "vault", 28, 3, []string{"cars"}},
		{"second branch", "b", "site-b", "group-b", "vault", 29, 3, []string{"camera"}},
		{"wrong schedule", "b", "site-b", "group-b", "vault", 28, 3, nil},
		{"wrong device", "a", "site-b", "group-b", "vault", 29, 3, nil},
		{"wrong site", "a", "site-b", "group-a", "vault", 28, 3, nil},
		{"wrong group", "a", "site-a", "group-b", "vault", 28, 3, nil},
		{"edge predicates", "a", "site-a", "group-a", "vault", 28, 0, nil},
		{"legacy predicates", "a", "site-a", "group-a", "other", 28, 3, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := AutomaticTriggerRootWithInputs(WorkflowDevice{
				DeviceKey: tc.key, SiteIds: []string{tc.site}, GroupIds: []string{tc.group}, Provider: tc.provider,
			}, WorkflowUser{}, map[string]any{"classify": map[string]any{"objectCount": tc.count}})
			match, err := w.MatchAutomaticTrigger(root, time.Date(2026, 9, tc.day, 1, 30, 0, 0, time.UTC))
			if err != nil {
				t.Fatal(err)
			}
			var got []string
			if match != nil {
				got = match.MatchedEdgeIds
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("matched %v, want %v", got, tc.want)
			}
		})
	}
}

func TestStartEdgeTriggerValidationAndModeSwitch(t *testing.T) {
	for name, mutate := range map[string]func(*Workflow){
		"stage edge":   func(w *Workflow) { w.Edges[2].Trigger = &WorkflowEdgeTrigger{} },
		"device edge":  func(w *Workflow) { w.Nodes[0].Type = WorkflowNodeDevice; w.Edges[0].Trigger = &WorkflowEdgeTrigger{} },
		"empty device": func(w *Workflow) { w.Edges[0].Trigger = &WorkflowEdgeTrigger{Devices: []DeviceKey{{Key: " "}}} },
		"empty site":   func(w *Workflow) { w.Edges[0].Trigger = &WorkflowEdgeTrigger{SiteIds: []string{""}} },
		"empty group":  func(w *Workflow) { w.Edges[0].Trigger = &WorkflowEdgeTrigger{GroupIds: []string{""}} },
	} {
		t.Run(name, func(t *testing.T) {
			w := edgeWorkflow()
			mutate(&w)
			if w.ValidateTriggers() == nil || w.ValidateGraph() == nil || w.ValidateStartNodes() == nil {
				t.Fatal("invalid edge trigger accepted")
			}

		})
	}
	w := edgeWorkflow()
	w.Edges[0].Trigger = &WorkflowEdgeTrigger{Devices: []DeviceKey{{Key: "camera"}}, WeeklySchedule: edgeSchedule(1, 3600, 7200)}
	w.Edges[1].Trigger = &WorkflowEdgeTrigger{}
	w.Nodes[0].Trigger.Surfaces = []WorkflowTriggerSurface{WorkflowSurfaceMedia}
	original, _ := json.Marshal(w.Edges)
	for _, mode := range []WorkflowTriggerType{WorkflowTriggerManual, WorkflowTriggerAutomatic, WorkflowTriggerManual} {
		w.Nodes[0].Trigger.Type = mode
		if err := w.ValidateGraph(); err != nil {
			t.Fatal(err)
		}
		w.NormalizeTriggers()
		after, _ := json.Marshal(w.Edges)
		if string(after) != string(original) {
			t.Fatal("mode switch rewrote edge configuration")
		}
		if mode == WorkflowTriggerManual && (len(w.Triggers) != 1 || len(w.Triggers[0].Devices)+len(w.Triggers[0].WeeklySchedule) != 0) {
			t.Fatal("manual mode exposed dormant source/schedule")
		}
		if mode == WorkflowTriggerAutomatic && (len(w.Triggers) != 2 || w.Triggers[0].Devices[0].Key != "camera") {
			t.Fatal("automatic mode did not restore edge scope")
		}
	}
	// Dormant edge scope is ignored, just like legacy automatic node settings.
	w.Edges[0].Trigger.Devices[0].Key = ""
	if err := w.ValidateGraph(); err != nil {
		t.Fatal(err)
	}
	w.Nodes[0].Trigger.Type = WorkflowTriggerAutomatic
	if w.ValidateGraph() == nil {
		t.Fatal("automatic mode accepted invalid restored scope")
	}
	w.Edges[0].Trigger = &WorkflowEdgeTrigger{}
	w.Nodes[0].Devices = []DeviceKey{{Key: ""}}
	w.Nodes[0].Trigger.SiteIds = []string{""}
	if err := w.ValidateGraph(); err != nil {
		t.Fatalf("overridden legacy scope blocked authored edge scope: %v", err)
	}
	w.Edges[0].Trigger = nil
	if w.ValidateGraph() == nil {
		t.Fatal("inherited legacy scope escaped validation")
	}
}
