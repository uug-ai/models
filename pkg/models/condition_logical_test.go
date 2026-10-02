package models

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func logicalCondition(op ConditionOp, children ...WorkflowCondition) WorkflowCondition {
	return WorkflowCondition{Op: op, Conditions: children}
}

func TestLogicalConditionsPreserveIndependentAnyGroups(t *testing.T) {
	is := func(path, value string) WorkflowCondition {
		return WorkflowCondition{Path: path, Op: ConditionOpEq, Value: value}
	}
	condition := logicalCondition(ConditionOpAll,
		logicalCondition(ConditionOpAny, is("device.deviceKey", "front"), is("device.deviceKey", "back")),
		logicalCondition(ConditionOpAny, is("inputs.classify.properties.*", "person"), is("inputs.classify.properties.*", "car")),
	)
	for _, tc := range []struct {
		device, label string
		want          bool
	}{
		{"front", "person", true}, {"back", "car", true},
		{"other", "car", false}, {"front", "truck", false}, {"other", "truck", false},
	} {
		root := AutomaticTriggerRootWithInputs(WorkflowDevice{DeviceKey: tc.device}, WorkflowUser{},
			map[string]any{"classify": map[string]any{"properties": []any{tc.label}}})
		if err := ValidateWorkflowCondition(&condition); err != nil {
			t.Fatal(err)
		}
		if got := EvaluateCondition(&condition, root); got != tc.want {
			t.Fatalf("%+v = %v", tc, got)
		}
	}
	for name, marshal := range map[string]func(any) ([]byte, error){"json": json.Marshal, "bson": bson.Marshal} {
		t.Run(name, func(t *testing.T) {
			data, err := marshal(condition)
			if err != nil {
				t.Fatal(err)
			}
			var loaded WorkflowCondition
			if name == "json" {
				err = json.Unmarshal(data, &loaded)
			} else {
				err = bson.Unmarshal(data, &loaded)
			}
			if err != nil || !reflect.DeepEqual(loaded, condition) {
				t.Fatalf("round trip %+v: %v", loaded, err)
			}
		})
	}
}

func TestLogicalConditionsNestAnyMatchWithoutChangingSameElementSemantics(t *testing.T) {
	condition := logicalCondition(ConditionOpAll, sameObjectCondition(ConditionModeAll))
	if err := ValidateWorkflowCondition(&condition); err != nil {
		t.Fatal(err)
	}
	cross := detectionsRoot([]any{
		map[string]any{"classified": "car", "confidence": 0.5},
		map[string]any{"classified": "person", "confidence": 0.99},
	})
	if EvaluateCondition(&condition, cross) {
		t.Fatal("logical wrapper changed same-object semantics")
	}
	if !EvaluateCondition(&condition, detectionsRoot([]any{map[string]any{"classified": "car", "confidence": 0.99}})) {
		t.Fatal("valid anyMatch inside logical group failed")
	}
}

func TestLogicalConditionsRejectEmptyMixedAndUnboundedGroups(t *testing.T) {
	pass := WorkflowCondition{Path: "key", Op: ConditionOpExists}
	valid := logicalCondition(ConditionOpAll, pass)
	deep := pass
	for i := 0; i < MaxWorkflowConditionDepth; i++ {
		deep = logicalCondition(ConditionOpAll, deep)
	}
	if err := ValidateWorkflowCondition(&deep); err != nil {
		t.Fatalf("maximum valid depth: %v", err)
	}
	tooDeep := logicalCondition(ConditionOpAny, deep)
	tooMany := logicalCondition(ConditionOpAny, make([]WorkflowCondition, MaxWorkflowConditionCount)...)
	for i := range tooMany.Conditions {
		tooMany.Conditions[i] = pass
	}
	invalid := map[string]WorkflowCondition{
		"empty all":             {Op: ConditionOpAll},
		"empty any":             {Op: ConditionOpAny, Conditions: []WorkflowCondition{}},
		"group path":            {Op: ConditionOpAll, Path: "key", Conditions: valid.Conditions},
		"group value":           {Op: ConditionOpAny, Value: false, Conditions: valid.Conditions},
		"group match":           {Op: ConditionOpAll, Match: &WorkflowPredicateSet{}, Conditions: valid.Conditions},
		"scalar children":       {Op: ConditionOpExists, Path: "key", Conditions: valid.Conditions},
		"scalar empty children": {Op: ConditionOpExists, Path: "key", Conditions: []WorkflowCondition{}},
		"anyMatch children":     {Op: ConditionOpAnyMatch, Path: "inputs.classify.details", Conditions: valid.Conditions},
		"unknown child":         logicalCondition(ConditionOpAll, WorkflowCondition{Path: "key", Op: "unknown"}),
		"depth":                 tooDeep, "count": tooMany,
	}
	for name, condition := range invalid {
		t.Run(name, func(t *testing.T) {
			if err := ValidateWorkflowCondition(&condition); err == nil {
				t.Fatal("invalid group accepted")
			}
			// Invalid structure cannot hide behind an earlier successful OR arm.
			set := WorkflowConditionSet{ConditionMode: ConditionModeAny, Conditions: []WorkflowCondition{pass, condition}}
			if EvaluateCondition(&condition, map[string]any{"key": "recording"}) || EvaluateConditionSet(set, map[string]any{"key": "recording"}) {
				t.Fatal("invalid group did not fail closed")
			}
		})
	}
	leaves := make([]WorkflowCondition, MaxWorkflowConditionCount)
	for i := range leaves {
		leaves[i] = pass
	}
	if err := ValidateWorkflowConditionSet(WorkflowConditionSet{Conditions: leaves}); err != nil {
		t.Fatal(err)
	}
	if ValidateWorkflowConditionSet(WorkflowConditionSet{Conditions: append(leaves, pass)}) == nil {
		t.Fatal("set budget not enforced")
	}
}

func TestLogicalGroupsValidateEveryLeafAndAutomaticInputs(t *testing.T) {
	for _, leaf := range []WorkflowCondition{
		{Path: "storage.secret", Op: ConditionOpExists},
		{Path: "device.deviceKey", Op: ConditionOpMatches, Value: "("},
		{Path: "key", Op: ConditionOpGt, Value: "bad"},
	} {
		group := logicalCondition(ConditionOpAny, WorkflowCondition{Path: "key", Op: ConditionOpExists}, logicalCondition(ConditionOpAll, leaf))
		if ValidateWorkflowCondition(&group) == nil {
			t.Fatal("invalid nested leaf accepted")
		}
	}
	group := logicalCondition(ConditionOpAll, logicalCondition(ConditionOpAny, WorkflowCondition{Path: "results.worker.ready", Op: ConditionOpExists}))
	if ValidateWorkflowCondition(&group) != nil {
		t.Fatal("continuation result rejected")
	}
	trigger := WorkflowTrigger{Conditions: []WorkflowCondition{group}}
	if trigger.Validate() == nil {
		t.Fatal("nested future result accepted for automatic activation")
	}
	w := edgeWorkflow()
	w.Edges[0].Conditions = []WorkflowCondition{group}
	if w.ValidateStartNodes() == nil {
		t.Fatal("nested future result accepted on automatic Start")
	}
}

func TestStartLegacyMigrationToLogicalEdgePreservesActivation(t *testing.T) {
	w := edgeWorkflow()
	w.Nodes[0].Trigger.ConditionMode = ConditionModeAny
	w.Nodes[0].Trigger.Conditions = []WorkflowCondition{
		{Path: "device.deviceName", Op: ConditionOpEq, Value: "front"},
		{Path: "device.deviceName", Op: ConditionOpEq, Value: "back"},
	}
	w.Edges[0].ConditionMode = ConditionModeAny
	w.Edges[0].Conditions = append(w.Edges[0].Conditions,
		WorkflowCondition{Path: "device.deviceKey", Op: ConditionOpEq, Value: "special"})
	migrated := w
	migrated.Nodes = append([]WorkflowNode(nil), w.Nodes...)
	trigger := *w.Nodes[0].Trigger
	migrated.Nodes[0].Trigger = &trigger
	migrated.Edges = append([]WorkflowEdge(nil), w.Edges...)
	for i := range migrated.Edges {
		if migrated.Edges[i].Source != "start" {
			continue
		}
		set, err := migrated.Edges[i].ConditionSet()
		if err != nil {
			t.Fatal(err)
		}
		group := logicalCondition(ConditionOpAny, w.Nodes[0].Trigger.Conditions...)
		parts := []WorkflowCondition{group}
		if len(set.Conditions) > 0 {
			parts = append(parts, logicalCondition(ConditionOp(set.ConditionMode), set.Conditions...))
		}
		migrated.Edges[i].ConditionMode = ConditionModeAll
		migrated.Edges[i].Conditions = []WorkflowCondition{logicalCondition(ConditionOpAll, parts...)}
		migrated.Edges[i].Condition = nil
	}
	trigger.Conditions, trigger.ConditionMode = nil, ""
	if err := migrated.ValidateGraph(); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"front", "back", "other"} {
		for _, key := range []string{"camera", "special", "other"} {
			for _, count := range []int{0, 3} {
				root := AutomaticTriggerRootWithInputs(WorkflowDevice{DeviceKey: key, DeviceName: name}, WorkflowUser{},
					map[string]any{"classify": map[string]any{"objectCount": count}})
				before, err := w.MatchAutomaticTrigger(root, time.Now())
				if err != nil {
					t.Fatal(err)
				}
				after, err := migrated.MatchAutomaticTrigger(root, time.Now())
				if err != nil {
					t.Fatal(err)
				}
				if (before == nil) != (after == nil) {
					t.Fatal("migration changed activation")
				}
				if before != nil && !reflect.DeepEqual(before.MatchedEdgeIds, after.MatchedEdgeIds) {
					t.Fatal("migration changed entry branches")
				}
			}
		}
	}
}
