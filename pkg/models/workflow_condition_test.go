package models

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestNormalizeWorkflowConditions(t *testing.T) {
	legacy := &WorkflowCondition{Path: "key", Op: ConditionOpEq, Value: "camera"}
	for _, mode := range []ConditionMode{"", ConditionModeAll, ConditionModeAny} {
		set, err := NormalizeWorkflowConditions(mode, nil, legacy)
		if err != nil || set.ConditionMode != ConditionModeAll || len(set.Conditions) != 1 || !reflect.DeepEqual(set.Conditions[0], *legacy) {
			t.Fatalf("legacy mode %q: set=%+v, err=%v", mode, set, err)
		}
		if legacy.Path != "key" || legacy.Op != ConditionOpEq {
			t.Fatal("normalizing must not mutate the legacy representation")
		}
		for _, plural := range [][]WorkflowCondition{{}, {*legacy}} {
			if _, err := NormalizeWorkflowConditions(mode, plural, legacy); err == nil {
				t.Fatal("nonnil plural, even empty, must conflict with legacy")
			}
		}
		set, err = NormalizeWorkflowConditions(mode, []WorkflowCondition{*legacy}, nil)
		if err != nil || len(set.Conditions) != 1 {
			t.Fatalf("plural mode %q: set=%+v, err=%v", mode, set, err)
		}
		want := mode
		if want == "" {
			want = ConditionModeAll
		}
		if set.ConditionMode != want {
			t.Errorf("got mode %q, want %q", set.ConditionMode, want)
		}
	}
	for _, c := range []*WorkflowCondition{nil, legacy} {
		if _, err := NormalizeWorkflowConditions("bad", nil, c); err == nil {
			t.Fatal("invalid mode must be rejected with and without legacy")
		}
	}
	for _, conditions := range [][]WorkflowCondition{nil, {}} {
		set, err := NormalizeWorkflowConditions("", conditions, nil)
		if err != nil || (set.Conditions == nil) != (conditions == nil) {
			t.Fatalf("normalization changed nil/empty presence: %+v, %v", set, err)
		}
	}
	invalid := []WorkflowCondition{{Path: "key", Op: ConditionOpMatches, Value: "("}}
	if _, err := NormalizeWorkflowConditions("", invalid, nil); err != nil {
		t.Fatal("normalization must not perform predicate validation")
	}
}

func TestStageDependencyConditionsAndReadiness(t *testing.T) {
	root := map[string]any{"key": "recording"}
	predicates := []WorkflowCondition{
		{Path: "key", Op: ConditionOpEq, Value: "other"},
		{Path: "key", Op: ConditionOpEq, Value: "recording"},
	}
	dependencies := []StageDependency{
		{Operation: "classify"},
		{Operation: "classify", Condition: &predicates[1]},
		{Operation: "classify", ConditionMode: ConditionModeAny, Conditions: predicates},
		{Operation: "classify", ConditionMode: ConditionModeAny, Conditions: []WorkflowCondition{}},
	}
	for _, d := range dependencies {
		if err := d.ValidateConditions(); err != nil {
			t.Fatal(err)
		}
		for _, available := range []map[string]bool{nil, {}, {"other": true}, {"classify": false}} {
			if d.Matches(root, available) {
				t.Errorf("unsatisfied operation gate was bypassed: %+v", d)
			}
		}
		if !d.Matches(root, map[string]bool{"classify": true}) {
			t.Errorf("ready matching dependency failed: %+v", d)
		}
	}
	d := StageDependency{ConditionMode: ConditionModeAll, Conditions: predicates}
	if d.Matches(root, nil) {
		t.Fatal("condition all is independent from the default NeedsModeAny")
	}
	d.ConditionMode = ConditionModeAny
	if !d.Matches(root, nil) {
		t.Fatal("ungated any predicates should match immediately")
	}
	for _, invalid := range []StageDependency{
		{ConditionMode: "bad"},
		{Condition: &predicates[1], Conditions: []WorkflowCondition{}},
		{Conditions: []WorkflowCondition{{Path: "key", Op: "bad"}}},
	} {
		if invalid.ValidateConditions() == nil || invalid.Matches(root, nil) {
			t.Errorf("invalid dependency did not fail closed: %+v", invalid)
		}
	}
}

func TestWorkflowConditionJSONAndBSONRoundTrips(t *testing.T) {
	c := sameObjectCondition(ConditionModeAll)
	c.Match.Conditions = append(c.Match.Conditions,
		WorkflowPredicate{Path: "labels.*", Op: ConditionOpIn, Value: []string{"vehicle"}},
		WorkflowPredicate{Path: "labels", Op: ConditionOpMatches, Value: "^vehicle$"},
	)
	original := StageDependency{
		Operation: "anpr", ConditionMode: ConditionModeAny,
		Conditions: []WorkflowCondition{
			{Path: "key", Op: ConditionOpEq, Value: "other"},
			c,
		},
	}
	root := detectionsRoot(primitive.A{primitive.M{
		"classified": "car", "confidence": 0.95, "labels": primitive.A{"vehicle"},
	}})
	for _, codec := range []struct {
		name      string
		marshal   func(any) ([]byte, error)
		unmarshal func([]byte, any) error
	}{
		{"json", json.Marshal, json.Unmarshal},
		{"bson", bson.Marshal, bson.Unmarshal},
	} {
		t.Run(codec.name, func(t *testing.T) {
			raw, err := codec.marshal(original)
			if err != nil {
				t.Fatal(err)
			}
			var decoded StageDependency
			if err := codec.unmarshal(raw, &decoded); err != nil {
				t.Fatal(err)
			}
			if err := decoded.ValidateConditions(); err != nil {
				t.Fatal(err)
			}
			if decoded.Condition != nil || decoded.ConditionMode != ConditionModeAny ||
				len(decoded.Conditions) != 2 || decoded.Conditions[1].Match.ConditionMode != ConditionModeAll {
				t.Fatalf("condition wire fields lost: %+v", decoded)
			}
			rootRaw, err := codec.marshal(root)
			if err != nil {
				t.Fatal(err)
			}
			var decodedRoot map[string]any
			if err := codec.unmarshal(rootRaw, &decodedRoot); err != nil {
				t.Fatal(err)
			}
			if !decoded.Matches(decodedRoot, map[string]bool{"anpr": true}) || decoded.Matches(decodedRoot, nil) {
				t.Fatal("round-trip must preserve same-object predicates and readiness")
			}
		})
	}
}

func TestWorkflowConditionLegacyWireCompatibility(t *testing.T) {
	legacy := StageDependency{Operation: "classify", Condition: &StageCondition{Path: "key", Op: ConditionOpExists}}
	raw, err := json.Marshal(legacy)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"operation":"classify","condition":{"path":"key","op":"exists","value":null}}`
	if string(raw) != want {
		t.Fatalf("legacy JSON changed:\ngot %s\nwant %s", raw, want)
	}
	bsonRaw, err := bson.Marshal(legacy)
	if err != nil {
		t.Fatal(err)
	}
	var doc bson.M
	if err := bson.Unmarshal(bsonRaw, &doc); err != nil {
		t.Fatal(err)
	}
	condition := doc["condition"].(bson.M)
	if value, present := condition["value"]; !present || value != nil {
		t.Fatal("legacy nil value must remain an explicit BSON null")
	}
	if _, present := condition["match"]; present {
		t.Fatal("unused match must be omitted")
	}
	if _, present := doc["conditionMode"]; present {
		t.Fatal("new empty fields must not rewrite legacy BSON")
	}
	for name, payload := range map[string]string{
		"absent": `{"condition":{"path":"key","op":"exists"}}`,
		"null":   `{"condition":{"path":"key","op":"exists"},"conditions":null}`,
		"empty":  `{"condition":{"path":"key","op":"exists"},"conditions":[]}`,
	} {
		var d StageDependency
		if err := json.Unmarshal([]byte(payload), &d); err != nil {
			t.Fatal(err)
		}
		before, _ := json.Marshal(d)
		_, err := d.ConditionSet()
		if (err != nil) != (name == "empty") {
			t.Errorf("%s: got error %v", name, err)
		}
		after, _ := json.Marshal(d)
		if !bytes.Equal(before, after) {
			t.Fatal("normalization must not persist an automatic rewrite")
		}
	}
	for name, plural := range map[string]any{"null": nil, "empty": primitive.A{}} {
		doc := bson.M{"condition": legacy.Condition, "conditions": plural}
		raw, err := bson.Marshal(doc)
		if err != nil {
			t.Fatal(err)
		}
		var d StageDependency
		if err := bson.Unmarshal(raw, &d); err != nil {
			t.Fatal(err)
		}
		if _, err := d.ConditionSet(); (err != nil) != (name == "empty") {
			t.Errorf("BSON %s: got error %v", name, err)
		}
	}
}
