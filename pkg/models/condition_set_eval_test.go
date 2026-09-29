package models

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func sameObjectCondition(mode ConditionMode) WorkflowCondition {
	return WorkflowCondition{
		Path: "results.anpr.detections",
		Op:   ConditionOpAnyMatch,
		Match: &WorkflowPredicateSet{
			ConditionMode: mode,
			Conditions: []WorkflowPredicate{
				{Path: "classified", Op: ConditionOpEq, Value: "car"},
				{Path: "confidence", Op: ConditionOpGte, Value: 0.9},
			},
		},
	}
}

func detectionsRoot(detections any) map[string]any {
	return map[string]any{"results": primitive.M{"anpr": primitive.M{"detections": detections}}}
}

func TestEvaluateConditionSetModes(t *testing.T) {
	pass := WorkflowCondition{Path: "key", Op: ConditionOpEq, Value: "recording"}
	fail := WorkflowCondition{Path: "key", Op: ConditionOpEq, Value: "other"}
	root := map[string]any{"key": "recording"}
	for _, mode := range []ConditionMode{"", ConditionModeAll, ConditionModeAny, "invalid"} {
		for _, tt := range []struct {
			name       string
			conditions []WorkflowCondition
			all, any   bool
		}{
			{"nil", nil, true, true},
			{"empty", []WorkflowCondition{}, true, true},
			{"one true", []WorkflowCondition{pass}, true, true},
			{"one false", []WorkflowCondition{fail}, false, false},
			{"true false", []WorkflowCondition{pass, fail}, false, true},
			{"false true", []WorkflowCondition{fail, pass}, false, true},
			{"both true", []WorkflowCondition{pass, pass}, true, true},
			{"both false", []WorkflowCondition{fail, fail}, false, false},
		} {
			t.Run(string(mode)+"/"+tt.name, func(t *testing.T) {
				want := tt.all
				if mode == ConditionModeAny {
					want = tt.any
				} else if mode == "invalid" {
					want = false
				}
				if got := EvaluateConditionSet(WorkflowConditionSet{ConditionMode: mode, Conditions: tt.conditions}, root); got != want {
					t.Fatalf("got %v, want %v", got, want)
				}
			})
		}
	}
}

func TestEvaluateConditionAnyMatchSameObject(t *testing.T) {
	crossObject := primitive.A{
		primitive.M{"classified": "car", "confidence": 0.5},
		primitive.M{"classified": "pedestrian", "confidence": 0.99},
	}
	for _, mode := range []ConditionMode{"", ConditionModeAll, ConditionModeAny} {
		c := sameObjectCondition(mode)
		if got := EvaluateCondition(&c, detectionsRoot(crossObject)); got != (mode == ConditionModeAny) {
			t.Errorf("mode %q cross-object match = %v", mode, got)
		}
	}
	legacy := WorkflowConditionSet{Conditions: []WorkflowCondition{
		{Path: "results.anpr.detections.*.classified", Op: ConditionOpEq, Value: "car"},
		{Path: "results.anpr.detections.*.confidence", Op: ConditionOpGte, Value: 0.9},
	}}
	if !EvaluateConditionSet(legacy, detectionsRoot(crossObject)) {
		t.Fatal("absolute wildcard predicates must retain independent, cross-object matching")
	}

	c := sameObjectCondition("")
	for name, objects := range map[string]any{
		"bson":       primitive.A{primitive.M{"classified": "car", "confidence": 0.95}},
		"typed maps": []map[string]any{{"classified": "car", "confidence": 1}},
		"array":      [1]map[string]any{{"classified": "car", "confidence": float32(0.95)}},
		"mixed":      []any{"car", nil, primitive.M{"classified": "car", "confidence": 0.95}},
	} {
		t.Run(name, func(t *testing.T) {
			if !EvaluateCondition(&c, detectionsRoot(objects)) {
				t.Fatal("same-object match should succeed")
			}
		})
	}
}

func TestEvaluateConditionAnyMatchRequiresObjectArray(t *testing.T) {
	c := sameObjectCondition("")
	for name, root := range map[string]map[string]any{
		"missing": nil,
		"null":    detectionsRoot(nil),
		"empty":   detectionsRoot([]any{}),
		"string":  detectionsRoot("car"),
		"bytes":   detectionsRoot([]byte("car")),
		"object":  detectionsRoot(map[string]any{"classified": "car", "confidence": 1}),
		"scalars": detectionsRoot([]any{"car", 1, true, nil}),
	} {
		t.Run(name, func(t *testing.T) {
			if EvaluateCondition(&c, root) {
				t.Fatal("anyMatch must match an object inside an array")
			}
		})
	}

	c.Match.Conditions = []WorkflowPredicate{{Path: "absent", Op: ConditionOpNe, Value: "blocked"}}
	if !EvaluateCondition(&c, detectionsRoot([]any{map[string]any{}})) {
		t.Fatal("ne on a missing relative field must retain vacuous truth")
	}
	if EvaluateCondition(&c, detectionsRoot([]any{"not an object"})) {
		t.Fatal("scalar array elements cannot satisfy even a negative predicate")
	}
}

func TestEvaluateConditionAnyMatchRelativeWildcards(t *testing.T) {
	c := sameObjectCondition("")
	c.Match.Conditions = []WorkflowPredicate{
		{Path: "classified", Op: ConditionOpEq, Value: "car"},
		{Path: "attributes.*.color", Op: ConditionOpIn, Value: primitive.A{"red"}},
	}
	root := detectionsRoot(primitive.A{primitive.M{
		"classified": "car", "attributes": primitive.A{primitive.M{"color": "red"}},
	}})
	if !EvaluateCondition(&c, root) {
		t.Fatal("relative scalar predicates may traverse nested arrays with wildcards")
	}

	c.Path = "results.anpr.batches.*.detections"
	root = map[string]any{"results": primitive.M{"anpr": primitive.M{
		"batches": primitive.A{primitive.M{"detections": primitive.A{primitive.M{
			"classified": "car", "attributes": primitive.A{primitive.M{"color": "red"}},
		}}}},
	}}}
	if !EvaluateCondition(&c, root) {
		t.Fatal("an absolute anyMatch path may resolve multiple array candidates")
	}
}

func TestEvaluateConditionSetMalformedGroupsFailClosed(t *testing.T) {
	valid := sameObjectCondition("")
	pass := WorkflowCondition{Path: "key", Op: ConditionOpExists}
	malformed := []WorkflowCondition{
		{Path: valid.Path, Op: ConditionOpAnyMatch},
		{Path: valid.Path, Op: ConditionOpAnyMatch, Match: &WorkflowPredicateSet{}},
		{Path: valid.Path, Op: ConditionOpAnyMatch, Match: &WorkflowPredicateSet{ConditionMode: "bad", Conditions: valid.Match.Conditions}},
		{Path: valid.Path, Op: ConditionOpAnyMatch, Match: &WorkflowPredicateSet{Conditions: []WorkflowPredicate{{Path: "nested", Op: ConditionOpAnyMatch}}}},
		{Path: valid.Path, Op: ConditionOpAnyMatch, Match: &WorkflowPredicateSet{ConditionMode: ConditionModeAny, Conditions: []WorkflowPredicate{
			{Path: "classified", Op: ConditionOpEq, Value: "car"},
			{Path: "nested", Op: ConditionOpAnyMatch},
		}}},
		{Path: valid.Path, Op: ConditionOpAnyMatch, Match: &WorkflowPredicateSet{Conditions: []WorkflowPredicate{{Path: "results.anpr.detections", Op: ConditionOpExists}}}},
		{Path: valid.Path, Op: ConditionOpAnyMatch, Match: &WorkflowPredicateSet{Conditions: []WorkflowPredicate{{Path: "classified", Op: "unknown"}}}},
		{Path: valid.Path, Op: ConditionOpAnyMatch, Value: "ignored", Match: valid.Match},
		{Path: "key", Op: ConditionOpAnyMatch, Match: valid.Match},
		{Path: "key", Op: ConditionOpExists, Match: valid.Match},
	}
	root := detectionsRoot([]any{map[string]any{"classified": "car", "confidence": 1}})
	root["key"] = "recording"
	for i, c := range malformed {
		if EvaluateCondition(&c, root) {
			t.Errorf("malformed condition %d matched directly", i)
		}
		for _, conditions := range [][]WorkflowCondition{{pass, c}, {c, pass}} {
			if EvaluateConditionSet(WorkflowConditionSet{ConditionMode: ConditionModeAny, Conditions: conditions}, root) {
				t.Errorf("malformed condition %d bypassed by outer any short-circuit", i)
			}
		}
	}
}

func TestEvaluateConditionScalarCompatibility(t *testing.T) {
	root := map[string]any{
		"key":    "camera-42",
		"device": primitive.M{"siteIds": []string{"site-1", "site-2"}, "groupIds": primitive.A{"group-1"}},
		"results": primitive.M{"custom": primitive.M{
			"number": int64(4), "nil": nil, "bool": true,
			"objects": primitive.A{primitive.M{"tag": "a"}, primitive.M{"tag": "b"}},
		}},
	}
	for _, c := range []WorkflowCondition{
		{Path: "key", Op: ConditionOpEq, Value: "camera-42"},
		{Path: "key", Op: ConditionOpNe, Value: "other"},
		{Path: "key", Op: ConditionOpContains, Value: "camera"},
		{Path: "key", Op: ConditionOpIn, Value: []string{"camera-42"}},
		{Path: "key", Op: ConditionOpExists},
		{Path: "key", Op: ConditionOpMatches, Value: "^camera-[0-9]+$"},
		{Path: "device.siteIds", Op: ConditionOpContains, Value: "site-1"},
		{Path: "device.siteIds.*", Op: ConditionOpIn, Value: primitive.A{"site-2"}},
		{Path: "device.groupIds.*", Op: ConditionOpIn, Value: []string{"group-1"}},
		{Path: "results.custom.number", Op: ConditionOpEq, Value: float64(4)},
		{Path: "results.custom.number", Op: ConditionOpGt, Value: 3},
		{Path: "results.custom.number", Op: ConditionOpGte, Value: 4},
		{Path: "results.custom.number", Op: ConditionOpLt, Value: 5},
		{Path: "results.custom.number", Op: ConditionOpLte, Value: 4},
		{Path: "results.custom.nil", Op: ConditionOpEq, Value: nil},
		{Path: "results.custom.nil", Op: ConditionOpExists},
		{Path: "results.custom.bool", Op: ConditionOpEq, Value: true},
		{Path: "results.custom.missing", Op: ConditionOpNe, Value: nil},
		{Path: "results.custom.objects.*.tag", Op: ConditionOpNe, Value: "c"},
	} {
		if !EvaluateCondition(&c, root) {
			t.Errorf("legacy scalar predicate did not match: %+v", c)
		}
	}
	for _, c := range []WorkflowCondition{
		{Path: "results.custom.missing", Op: ConditionOpEq, Value: nil},
		{Path: "results.custom.missing", Op: ConditionOpExists},
		{Path: "results.custom.objects.*.tag", Op: ConditionOpNe, Value: "a"},
		{Path: "results.custom.number", Op: ConditionOpEq, Value: "4"},
		{Path: "key", Op: ConditionOpMatches, Value: "("},
		{Path: "key", Op: ConditionOpMatches, Value: 1},
		{Path: "key", Op: "unknown"},
	} {
		if EvaluateCondition(&c, root) {
			t.Errorf("legacy scalar predicate unexpectedly matched: %+v", c)
		}
	}
	if !EvaluateCondition(nil, nil) {
		t.Fatal("a nil legacy condition remains unconditional")
	}
}

func TestEvaluateConditionMatchesTypedCollections(t *testing.T) {
	c := WorkflowCondition{Path: "results.custom.labels", Op: ConditionOpMatches, Value: "^car$"}
	for name, labels := range map[string]any{
		"string": "car", "slice": []string{"lorry", "car"}, "array": [2]string{"lorry", "car"},
		"bson": primitive.A{7, "car"}, "plain": []any{false, "car"},
	} {
		if !EvaluateCondition(&c, map[string]any{"results": primitive.M{"custom": primitive.M{"labels": labels}}}) {
			t.Errorf("%s should match", name)
		}
	}
	for _, labels := range []any{[]byte("car"), []any{nil, 42}, map[string]string{"label": "car"}} {
		if EvaluateCondition(&c, map[string]any{"results": primitive.M{"custom": primitive.M{"labels": labels}}}) {
			t.Errorf("%T should not match", labels)
		}
	}
}

func TestEvaluateConditionIncomparableValuesDoNotPanic(t *testing.T) {
	for _, value := range []any{[]any{"car"}, map[string]any{"type": "car"}} {
		root := map[string]any{"results": map[string]any{"custom": map[string]any{"value": value}}}
		for _, op := range []ConditionOp{ConditionOpEq, ConditionOpNe, ConditionOpContains, ConditionOpIn} {
			c := WorkflowCondition{Path: "results.custom.value", Op: op, Value: value}
			if op == ConditionOpIn {
				c.Value = []any{value}
			}
			_ = EvaluateCondition(&c, root)
		}
	}
}
