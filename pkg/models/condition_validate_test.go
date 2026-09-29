package models

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestValidateWorkflowConditionScalars(t *testing.T) {
	for _, c := range []WorkflowCondition{
		{Path: "key", Op: ConditionOpEq, Value: nil},
		{Path: "key", Op: ConditionOpNe, Value: true},
		{Path: "key", Op: ConditionOpContains, Value: "camera"},
		{Path: "key", Op: ConditionOpIn, Value: []any{nil, true, "camera", 42}},
		{Path: "key", Op: ConditionOpIn, Value: primitive.A{"camera"}},
		{Path: "key", Op: ConditionOpIn, Value: []string{}},
		{Path: "key", Op: ConditionOpExists, Value: "legacy ignored value"},
		{Path: "key", Op: ConditionOpMatches, Value: "^camera-[0-9]+$"},
		{Path: "inputs.classify.objectCount", Op: ConditionOpGt, Value: int64(2)},
		{Path: "inputs.classify.objectCount", Op: ConditionOpGte, Value: float32(2)},
		{Path: "inputs.classify.objectCount", Op: ConditionOpLt, Value: uint16(2)},
		{Path: "inputs.classify.objectCount", Op: ConditionOpLte, Value: 2.5},
		{Path: "device.siteIds", Op: ConditionOpContains, Value: "site"},
		{Path: "device.siteIds.*", Op: ConditionOpIn, Value: []string{"site"}},
		{Path: "device.groupIds", Op: ConditionOpContains, Value: "group"},
		{Path: "device.groupIds.*", Op: ConditionOpIn, Value: []string{"group"}},
	} {
		if err := ValidateWorkflowCondition(&c); err != nil {
			t.Errorf("%+v: %v", c, err)
		}
		if err := ValidateStageCondition(&c); err != nil {
			t.Errorf("legacy validator %+v: %v", c, err)
		}
	}
	if err := ValidateWorkflowCondition(nil); err != nil {
		t.Fatal(err)
	}
}

func TestValidateWorkflowConditionRejectsInvalidOperands(t *testing.T) {
	for name, c := range map[string]WorkflowCondition{
		"unknown operator": {Path: "key", Op: "overlap"},
		"missing operator": {Path: "key"},
		"eq object":        {Path: "key", Op: ConditionOpEq, Value: map[string]any{"key": "value"}},
		"ne array":         {Path: "key", Op: ConditionOpNe, Value: []string{"value"}},
		"contains array":   {Path: "key", Op: ConditionOpContains, Value: []string{"value"}},
		"in scalar":        {Path: "key", Op: ConditionOpIn, Value: "value"},
		"in null":          {Path: "key", Op: ConditionOpIn},
		"in bytes":         {Path: "key", Op: ConditionOpIn, Value: []byte("value")},
		"in nested list":   {Path: "key", Op: ConditionOpIn, Value: []any{[]string{"value"}}},
		"in object":        {Path: "key", Op: ConditionOpIn, Value: []any{primitive.M{"key": "value"}}},
		"numeric string":   {Path: "key", Op: ConditionOpGt, Value: "42"},
		"numeric boolean":  {Path: "key", Op: ConditionOpGte, Value: true},
		"numeric null":     {Path: "key", Op: ConditionOpLt},
		"numeric array":    {Path: "key", Op: ConditionOpLte, Value: []int{42}},
		"regex number":     {Path: "key", Op: ConditionOpMatches, Value: 42},
		"regex syntax":     {Path: "key", Op: ConditionOpMatches, Value: "("},
		"regex lookahead":  {Path: "key", Op: ConditionOpMatches, Value: "(?=car)"},
	} {
		t.Run(name, func(t *testing.T) {
			if err := ValidateWorkflowCondition(&c); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestValidateWorkflowConditionAnyMatch(t *testing.T) {
	for _, mode := range []ConditionMode{"", ConditionModeAll, ConditionModeAny} {
		c := sameObjectCondition(mode)
		if err := ValidateWorkflowCondition(&c); err != nil {
			t.Fatal(err)
		}
	}
	for _, path := range []string{"classified", "confidence", "attributes.color", "attributes.*.color", "object-key"} {
		c := sameObjectCondition("")
		c.Match.Conditions = []WorkflowPredicate{{Path: path, Op: ConditionOpExists}}
		if err := ValidateWorkflowCondition(&c); err != nil {
			t.Errorf("relative path %q: %v", path, err)
		}
	}
	for _, path := range []string{
		"inputs.classify.details", "inputs.classify.properties", "results.anpr.detections",
		"results.anpr.batches.*.detections", "device.siteIds", "device.groupIds",
	} {
		c := sameObjectCondition("")
		c.Path = path
		if err := ValidateWorkflowCondition(&c); err != nil {
			t.Errorf("array root %q: %v", path, err)
		}
	}
}

func TestValidateWorkflowConditionRejectsMalformedGroups(t *testing.T) {
	base := sameObjectCondition("")
	for name, c := range map[string]WorkflowCondition{
		"missing match":       {Path: base.Path, Op: ConditionOpAnyMatch},
		"empty match":         {Path: base.Path, Op: ConditionOpAnyMatch, Match: &WorkflowPredicateSet{}},
		"empty any match":     {Path: base.Path, Op: ConditionOpAnyMatch, Match: &WorkflowPredicateSet{ConditionMode: ConditionModeAny}},
		"unknown match mode":  {Path: base.Path, Op: ConditionOpAnyMatch, Match: &WorkflowPredicateSet{ConditionMode: "none", Conditions: base.Match.Conditions}},
		"anyMatch value":      {Path: base.Path, Op: ConditionOpAnyMatch, Value: "operand", Match: base.Match},
		"match on scalar op":  {Path: "key", Op: ConditionOpEq, Value: "key", Match: base.Match},
		"nested anyMatch":     {Path: base.Path, Op: ConditionOpAnyMatch, Match: &WorkflowPredicateSet{Conditions: []WorkflowPredicate{{Path: "objects", Op: ConditionOpAnyMatch}}}},
		"unknown inner op":    {Path: base.Path, Op: ConditionOpAnyMatch, Match: &WorkflowPredicateSet{Conditions: []WorkflowPredicate{{Path: "label", Op: "between"}}}},
		"invalid inner type":  {Path: base.Path, Op: ConditionOpAnyMatch, Match: &WorkflowPredicateSet{Conditions: []WorkflowPredicate{{Path: "confidence", Op: ConditionOpGt, Value: "0.9"}}}},
		"invalid inner regex": {Path: base.Path, Op: ConditionOpAnyMatch, Match: &WorkflowPredicateSet{Conditions: []WorkflowPredicate{{Path: "label", Op: ConditionOpMatches, Value: "("}}}},
	} {
		t.Run(name, func(t *testing.T) {
			if err := ValidateWorkflowCondition(&c); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
	for _, path := range []string{
		"", ".classified", "classified.", "attributes..color", " classified", "attributes. color",
		"objects.0.color", "objects[0].color", "*", "class*", "$where",
		"inputs.classify.details", "results.anpr.detections", "device.deviceKey",
		"user.organisationId", "storage.secret", "user.storage.secret",
	} {
		c := sameObjectCondition("")
		c.Match.Conditions = []WorkflowPredicate{{Path: path, Op: ConditionOpExists}}
		if err := ValidateWorkflowCondition(&c); err == nil {
			t.Errorf("invalid relative path %q accepted", path)
		}
	}
	for _, path := range []string{
		"operation", "key", "traceId", "runId", "user.organisationId",
		"device.deviceKey", "device.deviceName", "device.siteIds.*", "device.groupIds.*",
		"inputs.classify.objectCount", "inputs.classify", "results.anpr",
	} {
		c := sameObjectCondition("")
		c.Path = path
		if err := ValidateWorkflowCondition(&c); err == nil {
			t.Errorf("non-array root %q accepted", path)
		}
	}
}

func TestValidateWorkflowConditionCredentialAndNamespacePaths(t *testing.T) {
	for _, path := range []string{
		"storage", "storage.secret", "user.storage", "user.storage.secret",
		"device.password", "device.*", "device.groupIds.name", "device.groupIds.*.name",
		"device.siteIds.0", "device.deviceKey.*", "device.deviceName.extra",
		"inputs.classify.details.classified", "inputs.classify.objectCount.*",
		"results.classify.unknown", "inputs..classify", "results.custom.array[0]",
	} {
		c := WorkflowCondition{Path: path, Op: ConditionOpExists}
		if err := ValidateWorkflowCondition(&c); err == nil {
			t.Errorf("invalid path %q accepted", path)
		}
	}
}

func TestValidateWorkflowConditionSetModes(t *testing.T) {
	for _, mode := range []ConditionMode{"", ConditionModeAll, ConditionModeAny} {
		if err := ValidateWorkflowConditionSet(WorkflowConditionSet{ConditionMode: mode}); err != nil {
			t.Errorf("empty set mode %q: %v", mode, err)
		}
	}
	if err := ValidateWorkflowConditionSet(WorkflowConditionSet{ConditionMode: "bad"}); err == nil {
		t.Fatal("invalid mode on an empty set must be rejected")
	}
	set := WorkflowConditionSet{
		ConditionMode: ConditionModeAny,
		Conditions: []WorkflowCondition{
			{Path: "key", Op: ConditionOpExists},
			{Path: "key", Op: "unknown"},
		},
	}
	if err := ValidateWorkflowConditionSet(set); err == nil {
		t.Fatal("all members must be validated, regardless of mode or order")
	}
}
