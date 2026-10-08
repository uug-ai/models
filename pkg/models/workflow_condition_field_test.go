package models

import (
	"encoding/json"
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestConditionFieldValidation(t *testing.T) {
	leaf := func(field string) WorkflowCondition {
		return WorkflowCondition{Path: "inputs.classify.objectCount", Op: ConditionOpGte, Value: 2, Field: field}
	}
	anyMatch := func(field, predicateField string) WorkflowCondition {
		return WorkflowCondition{Path: "results.anpr.detections", Op: ConditionOpAnyMatch, Field: field,
			Match: &WorkflowPredicateSet{Conditions: []WorkflowPredicate{{Path: "plate", Op: ConditionOpEq, Value: "AB", Field: predicateField}}}}
	}
	for name, tt := range map[string]struct {
		condition WorkflowCondition
		valid     bool
	}{
		"untagged leaf":          {leaf(""), true},
		"tagged leaf":            {leaf("objectCount"), true},
		"hyphen and digits":      {leaf("plate-2_list"), true},
		"leading digit":          {leaf("2plates"), false},
		"dotted":                 {leaf("start.objectCount"), false},
		"too long":               {leaf("a" + strings.Repeat("b", 80)), false},
		"tagged predicate":       {anyMatch("", "plate"), true},
		"invalid predicate":      {anyMatch("", "plate list"), false},
		"tagged anyMatch":        {anyMatch("detections", "plate"), false},
		"tagged group":           {WorkflowCondition{Op: ConditionOpAll, Field: "scope", Conditions: []WorkflowCondition{leaf("objectCount")}}, false},
		"group of tagged leaves": {WorkflowCondition{Op: ConditionOpAny, Conditions: []WorkflowCondition{leaf("objectCount"), leaf("")}}, true},
	} {
		t.Run(name, func(t *testing.T) {
			condition := tt.condition
			if err := ValidateWorkflowCondition(&condition); (err == nil) != tt.valid {
				t.Fatalf("valid %v, error %v", tt.valid, err)
			}
		})
	}
}

// Field is metadata: evaluation must not change with or without it.
func TestConditionFieldDoesNotAffectEvaluation(t *testing.T) {
	root := WorkflowRun{Inputs: map[string]any{"classify": map[string]any{"objectCount": 3}}}.ConditionRoot(WorkflowConditionRootTrigger)
	for _, value := range []any{2, 4} {
		untagged := WorkflowConditionSet{Conditions: []WorkflowCondition{{Path: "inputs.classify.objectCount", Op: ConditionOpGte, Value: value}}}
		tagged := WorkflowConditionSet{Conditions: []WorkflowCondition{{Path: "inputs.classify.objectCount", Op: ConditionOpGte, Value: value, Field: "objectCount"}}}
		if EvaluateConditionSet(untagged, root) != EvaluateConditionSet(tagged, root) {
			t.Fatalf("field changed evaluation for value %v", value)
		}
	}
}

func TestConditionFieldRoundTrips(t *testing.T) {
	condition := WorkflowCondition{Path: "results.anpr.detections", Op: ConditionOpAnyMatch,
		Match: &WorkflowPredicateSet{Conditions: []WorkflowPredicate{{Path: "plate", Op: ConditionOpEq, Value: "AB", Field: "plate"}}}}
	outer := WorkflowCondition{Path: "device.deviceName", Op: ConditionOpEq, Value: "gate", Field: "deviceName"}

	encoded, _ := json.Marshal([]WorkflowCondition{condition, outer})
	var fromJSON []WorkflowCondition
	if err := json.Unmarshal(encoded, &fromJSON); err != nil || fromJSON[0].Match.Conditions[0].Field != "plate" || fromJSON[1].Field != "deviceName" {
		t.Fatalf("JSON round trip = %+v, %v", fromJSON, err)
	}
	raw, _ := bson.Marshal(outer)
	var fromBSON WorkflowCondition
	if err := bson.Unmarshal(raw, &fromBSON); err != nil || fromBSON.Field != "deviceName" {
		t.Fatalf("BSON round trip = %+v, %v", fromBSON, err)
	}
	untagged, _ := json.Marshal(WorkflowCondition{Path: "key", Op: ConditionOpExists})
	if strings.Contains(string(untagged), "field") {
		t.Fatalf("untagged condition serialises a field: %s", untagged)
	}
}
