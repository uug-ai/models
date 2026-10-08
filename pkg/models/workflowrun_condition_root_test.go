package models

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const conditionRootSecret = "never-matchable-secret"

// filledStruct sets every field of a struct to a recognisable non-zero value,
// so a field added later is exercised without updating this test.
func filledStruct(t *testing.T, target any) {
	t.Helper()
	value := reflect.ValueOf(target).Elem()
	for i := 0; i < value.NumField(); i++ {
		field, name := value.Field(i), value.Type().Field(i).Name
		switch field.Kind() {
		case reflect.String:
			field.SetString("value-" + name)
		case reflect.Slice:
			if field.Type().Elem().Kind() != reflect.String {
				t.Fatalf("unhandled slice field %s", name)
			}
			field.Set(reflect.ValueOf([]string{"first-" + name, "second-" + name}))
		case reflect.Ptr:
			if field.Type() != reflect.TypeOf(&primitive.ObjectID{}) {
				t.Fatalf("unhandled pointer field %s", name)
			}
			id := primitive.NewObjectID()
			field.Set(reflect.ValueOf(&id))
		case reflect.Struct:
			if field.Type() != reflect.TypeOf(Storage{}) {
				t.Fatalf("unhandled struct field %s", name)
			}
			field.Set(reflect.ValueOf(Storage{Uri: conditionRootSecret, AccessKey: conditionRootSecret, Secret: conditionRootSecret}))
		default:
			t.Fatalf("unhandled field %s of kind %s", name, field.Kind())
		}
	}
}

func filledConditionRun(t *testing.T) WorkflowRun {
	t.Helper()
	run := WorkflowRun{
		Id: primitive.NewObjectID(), RunId: "run-id", Key: "media-key", Operation: "classify", TraceId: "trace",
		Inputs:    map[string]any{"classify": map[string]any{"objectCount": 2}},
		Results:   map[string]any{"anpr": map[string]any{"detections": []any{}}},
		Storage:   &WorkflowStorage{AccessKey: conditionRootSecret, Secret: conditionRootSecret, VaultOverrideSecret: conditionRootSecret},
		SignedURL: conditionRootSecret, Payload: json.RawMessage(`"` + conditionRootSecret + `"`),
		Params: map[string]any{"password": conditionRootSecret},
	}
	filledStruct(t, &run.Device)
	filledStruct(t, &run.User)
	return run
}

// conditionRootPaths flattens a projection into path → type, describing array
// elements as ".*" and stopping at open namespaces.
func conditionRootPaths(t *testing.T, prefix string, value any, open map[string]bool, out map[string]string) {
	t.Helper()
	switch typed := value.(type) {
	case map[string]any:
		if open[prefix] {
			out[prefix] = "object"
			return
		}
		for key, child := range typed {
			path := key
			if prefix != "" {
				path = prefix + "." + key
			}
			conditionRootPaths(t, path, child, open, out)
		}
	case []any:
		if len(typed) == 0 {
			t.Fatalf("%s: fill arrays so their element type is checked", prefix)
		}
		for _, element := range typed {
			conditionRootPaths(t, prefix+".*", element, open, out)
		}
	case string:
		out[prefix] = "string"
	default:
		t.Fatalf("%s: unexpected projected type %T", prefix, value)
	}
}

func TestConditionRootMatchesSchema(t *testing.T) {
	run := filledConditionRun(t)
	for _, mode := range []WorkflowConditionRootMode{WorkflowConditionRootTrigger, WorkflowConditionRootRun} {
		t.Run(string(mode), func(t *testing.T) {
			want, open := map[string]string{}, map[string]bool{}
			for _, path := range WorkflowConditionRootSchema() {
				for _, allowed := range path.Modes {
					if allowed == mode && path.Open {
						want[path.Path], open[path.Path] = path.Type, true
					}
				}
			}
			for _, path := range WorkflowConditionRootSchema() {
				namespace, _, _ := strings.Cut(path.Path, ".")
				for _, allowed := range path.Modes {
					// Paths inside open namespaces are checked against worker output below.
					if allowed == mode && !path.Open && !open[namespace] {
						want[path.Path] = path.Type
					}
				}
			}
			got := map[string]string{}
			conditionRootPaths(t, "", run.ConditionRoot(mode), open, got)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("projection and schema differ:\n projected %v\n schema    %v", sortedPaths(got), sortedPaths(want))
			}
		})
	}
}

func sortedPaths(paths map[string]string) []string {
	out := make([]string, 0, len(paths))
	for path, kind := range paths {
		out = append(out, fmt.Sprintf("%s:%s", path, kind))
	}
	sort.Strings(out)
	return out
}

// Every envelope field must be projected or deliberately excluded, so adding a
// field to WorkflowDevice or WorkflowUser forces an explicit decision here.
func TestConditionRootCoversEnvelopeFields(t *testing.T) {
	excluded := map[string]bool{"user.projectId": true, "user.storage": true}
	root := filledConditionRun(t).ConditionRoot(WorkflowConditionRootRun)
	for prefix, kind := range map[string]reflect.Type{"device": reflect.TypeOf(WorkflowDevice{}), "user": reflect.TypeOf(WorkflowUser{})} {
		projected := root[prefix].(map[string]any)
		for i := 0; i < kind.NumField(); i++ {
			name := strings.Split(kind.Field(i).Tag.Get("json"), ",")[0]
			path := prefix + "." + name
			if _, ok := projected[name]; ok == excluded[path] {
				t.Errorf("%s must be either projected or listed as excluded, not both or neither", path)
			}
		}
	}
}

func TestConditionRootNeverExposesCredentials(t *testing.T) {
	run := filledConditionRun(t)
	for _, mode := range []WorkflowConditionRootMode{WorkflowConditionRootTrigger, WorkflowConditionRootRun} {
		encoded, err := json.Marshal(run.ConditionRoot(mode))
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(encoded), conditionRootSecret) {
			t.Fatalf("%s projection exposes credentials: %s", mode, encoded)
		}
	}
}

func TestConditionRootModes(t *testing.T) {
	run := WorkflowRun{Id: primitive.NewObjectID(), RunId: "ignored", Key: "media", Operation: "classify", TraceId: "trace",
		Device: WorkflowDevice{DeviceKey: "cam"}}
	trigger := run.ConditionRoot(WorkflowConditionRootTrigger)
	if _, ok := trigger["inputs"]; ok {
		t.Fatal("trigger mode must omit empty inputs")
	}
	if _, ok := trigger["runId"]; ok {
		t.Fatal("trigger mode has no run identity")
	}
	triggerRoot := AutomaticTriggerRootWithInputs(run.Device, run.User, nil)
	for _, key := range []string{"key", "operation", "traceId"} {
		delete(trigger, key)
	}
	if !reflect.DeepEqual(trigger, triggerRoot) {
		t.Fatalf("trigger mode %v differs from AutomaticTriggerRootWithInputs %v", trigger, triggerRoot)
	}
	root := run.ConditionRoot(WorkflowConditionRootRun)
	if root["runId"] != run.Id.Hex() || root["inputs"] == nil || root["results"] == nil {
		t.Fatalf("run mode = %v", root)
	}
	run.Id = primitive.NilObjectID
	if root := run.ConditionRoot(WorkflowConditionRootRun); root["runId"] != "ignored" {
		t.Fatalf("unsaved run identity = %v", root["runId"])
	}
}

// The projection is evaluated by the shared condition engine as before.
func TestConditionRootEvaluates(t *testing.T) {
	run := filledConditionRun(t)
	root := run.ConditionRoot(WorkflowConditionRootRun)
	for _, condition := range []WorkflowCondition{
		{Path: "device.siteIds", Op: ConditionOpContains, Value: "first-SiteIds"},
		{Path: "device.deviceName", Op: ConditionOpEq, Value: "value-DeviceName"},
		{Path: "inputs.classify.objectCount", Op: ConditionOpGte, Value: 2},
		{Path: "runId", Op: ConditionOpEq, Value: run.Id.Hex()},
	} {
		if !EvaluateConditionSet(WorkflowConditionSet{Conditions: []WorkflowCondition{condition}}, root) {
			t.Fatalf("condition %+v did not match %v", condition, root)
		}
	}
}
