package models

import (
	"reflect"
	"testing"
	"time"
)

func TestStartEdgeClassificationsRequiredAlongsideAnyPredicates(t *testing.T) {
	for _, tc := range []struct {
		name, key, site, group string
		labels                 []string
		hour                   int
		want                   []string
	}{
		{"first classification", "camera", "site", "group", []string{"dog", "person"}, 1, []string{"cars"}},
		{"second selected value", "camera", "site", "group", []string{"bicycle"}, 1, []string{"cars"}},
		{"second edge", "camera", "site", "group", []string{"car"}, 1, []string{"camera"}},
		{"both edges", "camera", "site", "group", []string{"person", "car"}, 1, []string{"cars", "camera"}},
		{"missing initial input", "camera", "site", "group", nil, 1, nil},
		{"empty detections", "camera", "site", "group", []string{}, 1, nil},
		{"unrelated predicate cannot bypass", "camera", "site", "group", []string{"dog"}, 1, nil},
		{"device AND classifications", "other", "site", "group", []string{"person"}, 1, nil},
		{"site AND classifications", "camera", "other", "group", []string{"person"}, 1, nil},
		{"group AND classifications", "camera", "site", "other", []string{"person"}, 1, nil},
		{"schedule AND classifications", "camera", "site", "group", []string{"person"}, 3, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := edgeWorkflow()
			for i, labels := range [][]string{{"person", "bicycle"}, {"car"}} {
				w.Edges[i].Trigger = &WorkflowEdgeTrigger{
					Devices: []DeviceKey{{Key: "camera"}}, SiteIds: []string{"site"}, GroupIds: []string{"group"},
					Classifications: labels, WeeklySchedule: edgeSchedule(1, 3600, 7200),
				}
				w.Edges[i].ConditionMode = ConditionModeAny
				w.Edges[i].Conditions = []WorkflowCondition{
					{Path: "device.deviceName", Op: ConditionOpEq, Value: "allowed"},
					{Path: "inputs.classify.objectCount", Op: ConditionOpGt, Value: 999},
				}
			}
			var inputs map[string]any
			if tc.labels != nil {
				details := make([]any, 0, len(tc.labels))
				for _, label := range tc.labels {
					details = append(details, map[string]any{"classified": label})
				}
				inputs = map[string]any{"classify": map[string]any{"details": details}}
			}
			root := AutomaticTriggerRootWithInputs(WorkflowDevice{
				DeviceKey: tc.key, DeviceName: "allowed", SiteIds: []string{tc.site}, GroupIds: []string{tc.group},
			}, WorkflowUser{}, inputs)
			root["results"] = map[string]any{"classify": map[string]any{
				"details": []any{map[string]any{"classified": "person"}, map[string]any{"classified": "car"}},
			}}
			match, err := w.MatchAutomaticTrigger(root, time.Date(2026, 9, 28, tc.hour, 30, 0, 0, time.UTC))
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
			if match != nil {
				if len(match.Trigger.Classifications) == 0 {
					t.Fatal("classification scope missing from detached trigger snapshot")
				}
				saved := match.Trigger.Classifications[0]
				w.Edges[match.Index].Trigger.Classifications[0] = "edited"
				if match.Trigger.Classifications[0] != saved {
					t.Fatal("authored edit changed persisted classification selection")
				}
			}
		})
	}
}

func TestStartEdgeClassificationsValidationAndLegacyCompatibility(t *testing.T) {
	w := edgeWorkflow()
	w.Edges[0].Trigger = &WorkflowEdgeTrigger{Classifications: []string{" "}}
	for _, enabled := range []bool{true, false} {
		w.Enabled = enabled
		if w.ValidateStartNodes() == nil || w.ValidateGraph() == nil {
			t.Fatal("blank classification selection accepted")
		}
	}
	w.Nodes[0].Trigger.Type = WorkflowTriggerManual
	w.Nodes[0].Trigger.Surfaces = []WorkflowTriggerSurface{WorkflowSurfaceMedia}
	if err := w.ValidateGraph(); err != nil {
		t.Fatal(err)
	}
	w.NormalizeTriggers()
	if len(w.Triggers[0].Classifications) != 0 || w.Edges[0].Trigger.Classifications[0] != " " {
		t.Fatal("manual mode activated or rewrote dormant classification scope")
	}
	w.Enabled = true
	w.Nodes[0].Trigger.Type = WorkflowTriggerAutomatic
	w.Nodes[0].Trigger.Classifications = []string{"ignored-node-classification"}
	for i := range w.Edges {
		w.Edges[i].Trigger = nil
		w.Edges[i].Conditions = nil
	}
	if !w.AutomaticMatches(nil, time.Now()) {
		t.Fatal("nil edge trigger invented node classification fallback")
	}
	for _, trigger := range w.Triggers {
		if len(trigger.Classifications) != 0 {
			t.Fatal("node classification leaked into derived scope")
		}
	}
	w.Edges[0].Trigger = &WorkflowEdgeTrigger{}
	if !w.AutomaticMatches(nil, time.Now()) {
		t.Fatal("empty classification scope was not unrestricted")
	}
	legacy := Workflow{Enabled: true, Stages: []WorkflowStage{{Operation: "legacy"}}, Triggers: []WorkflowTrigger{{}}}
	if !legacy.AutomaticMatches(nil, time.Now()) || legacy.CompileStages()[0].Operation != "legacy" {
		t.Fatal("stage-only workflow behavior changed")
	}
}

func TestClassificationsDoNotReplaceOuterPredicates(t *testing.T) {
	root := AutomaticTriggerRootWithInputs(WorkflowDevice{DeviceName: "allowed"}, WorkflowUser{}, map[string]any{
		"classify": map[string]any{"details": []any{map[string]any{"classified": "person"}}},
	})
	trigger := WorkflowTrigger{
		Classifications: []string{"person"}, ConditionMode: ConditionModeAny,
		Conditions: []WorkflowCondition{{Path: "device.deviceName", Op: ConditionOpEq, Value: "blocked"}},
	}
	if trigger.MatchesEnvelope(root) {
		t.Fatal("matching classification bypassed outer predicates")
	}
	trigger.Conditions[0].Value = "allowed"
	if !trigger.MatchesEnvelope(root) {
		t.Fatal("classification and predicate conjunction did not match")
	}
	w := edgeWorkflow()
	w.Edges = w.Edges[:1]
	legacyCondition := WorkflowCondition{Path: "inputs.classify.details.*.classified", Op: ConditionOpEq, Value: "person"}
	w.Edges[0].Conditions = []WorkflowCondition{legacyCondition}
	match, err := w.MatchAutomaticTrigger(root, time.Now())
	if err != nil || match == nil || len(match.Trigger.Classifications) != 0 {
		t.Fatalf("legacy classification predicate behavior changed: %+v, %v", match, err)
	}
	if !reflect.DeepEqual(w.Edges[0].Conditions, []WorkflowCondition{legacyCondition}) {
		t.Fatal("legacy predicate was rewritten")
	}
}
