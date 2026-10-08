package models

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

func classifySchemaPaths() []WorkflowConditionPath {
	var paths []WorkflowConditionPath
	for _, path := range WorkflowConditionRootSchema() {
		if strings.HasPrefix(path.Path, "inputs."+WorkflowClassifyOperation+".") {
			paths = append(paths, path)
		}
	}
	return paths
}

// classifyFixture is a classifier result shaped exactly like the worker's
// ReturnObject.py output, decoded strictly so a field the classifier adds or
// renames fails here until WorkflowClassifyInput is updated.
func classifyFixture(t *testing.T) (WorkflowClassifyInput, map[string]any) {
	t.Helper()
	raw, err := os.ReadFile("testdata/workflow_classify_input.json")
	if err != nil {
		t.Fatal(err)
	}
	var typed WorkflowClassifyInput
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&typed); err != nil {
		t.Fatalf("fixture does not match WorkflowClassifyInput: %v", err)
	}
	var generic map[string]any
	if err := json.Unmarshal(raw, &generic); err != nil {
		t.Fatal(err)
	}
	return typed, generic
}

func conditionValueType(value any) string {
	switch number := value.(type) {
	case string:
		return "string"
	case bool:
		return "boolean"
	case float64:
		if number == float64(int64(number)) {
			return "integer"
		}
		return "number"
	}
	return ""
}

func assertClassifyPaths(t *testing.T, classify map[string]any, include func(WorkflowConditionPath) bool) {
	t.Helper()
	root := WorkflowRun{Inputs: map[string]any{WorkflowClassifyOperation: classify}}.ConditionRoot(WorkflowConditionRootTrigger)
	for _, path := range classifySchemaPaths() {
		candidates, found := ResolveCandidates(root, path.Path)
		if !include(path) {
			if found {
				t.Errorf("%s resolves although it is marked automatic-only", path.Path)
			}
			continue
		}
		if !found || len(candidates) == 0 {
			t.Errorf("%s does not resolve in the classifier output", path.Path)
			continue
		}
		for _, candidate := range candidates {
			got := conditionValueType(candidate)
			if got != path.Type && !(path.Type == "number" && got == "integer") {
				t.Errorf("%s resolves to %T, schema says %s", path.Path, candidate, path.Type)
			}
		}
	}
}

func TestClassifySchemaMatchesClassifierOutput(t *testing.T) {
	_, classify := classifyFixture(t)
	assertClassifyPaths(t, classify, func(WorkflowConditionPath) bool { return true })
	if len(classifySchemaPaths()) != 15 {
		t.Fatalf("classify paths = %+v", classifySchemaPaths())
	}
}

// Manual launches seed inputs.classify from the stored analysis through
// Classify; every path not marked automatic-only must also resolve there.
func TestClassifySchemaMatchesManualSeed(t *testing.T) {
	typed, _ := classifyFixture(t)
	seed := Classify{Properties: typed.Properties}
	for _, detail := range typed.Details {
		seed.Details = append(seed.Details, ClassifyDetails{
			Id: detail.Id, Classified: detail.Classified, Distance: detail.Distance, StaticDistance: detail.StaticDistance,
			IsStatic: detail.IsStatic, FrameWidth: float64(detail.FrameWidth), FrameHeight: float64(detail.FrameHeight),
			Frame: float64(detail.Frame), Occurence: float64(detail.Occurence), ColorString: detail.ColorStr,
			Valid: detail.Valid, X: detail.X, Y: detail.Y,
		})
	}
	encoded, _ := json.Marshal(seed)
	var classify map[string]any
	if err := json.Unmarshal(encoded, &classify); err != nil {
		t.Fatal(err)
	}
	assertClassifyPaths(t, classify, func(path WorkflowConditionPath) bool { return !path.AutomaticOnly })
}

// The validator's path lists derive from the schema and must keep their
// previous meaning.
func TestConditionValidatorFieldsDeriveFromSchema(t *testing.T) {
	wantDevice := map[string]conditionFieldKind{
		"deviceKey": conditionFieldLeaf, "deviceName": conditionFieldLeaf,
		"provider": conditionFieldLeaf, "storageSolution": conditionFieldLeaf,
		"siteIds": conditionFieldArray, "groupIds": conditionFieldArray,
	}
	wantOperations := map[string]map[string]conditionFieldKind{
		"classify": {"properties": conditionFieldArray, "objectCount": conditionFieldLeaf, "details": conditionFieldArray},
	}
	if !reflect.DeepEqual(conditionDeviceFields, wantDevice) ||
		!reflect.DeepEqual(conditionUserFields, map[string]conditionFieldKind{"organisationId": conditionFieldLeaf}) ||
		!reflect.DeepEqual(conditionOperationFields, wantOperations) {
		t.Fatalf("derived fields: device %v user %v operations %v", conditionDeviceFields, conditionUserFields, conditionOperationFields)
	}
}
