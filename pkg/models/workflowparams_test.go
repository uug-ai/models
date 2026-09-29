package models

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func forwarderStage() WorkflowStage {
	return WorkflowStage{
		Operation: "forwarder",
		Params: []StageParam{
			{Name: "host", Type: StageParamString},
			{Name: "password", Type: StageParamSecret},
			{Name: "port", Type: StageParamNumber},
			{Name: "tls", Type: StageParamBoolean},
			{Name: "mode", Type: StageParamSelect, Options: []string{"callback", "delivered"}},
			{Name: "queue", Type: StageParamString, Required: true},
			{Name: "vhost", Type: StageParamString, Required: true, Default: "/"},
		},
	}
}

func TestNormalizeParamValuesCanonicalizesDeclaredValues(t *testing.T) {
	got, err := forwarderStage().NormalizeParamValues(map[string]any{
		"host":     "  rabbit.example.com:5671 ",
		"password": " keep spaces ",
		"port":     int32(5671),
		"tls":      false,
		"mode":     "callback",
		"queue":    "",
		"unknown":  "dropped",
	})
	if err != nil {
		t.Fatalf("NormalizeParamValues: %v", err)
	}
	want := map[string]any{
		"host":     "rabbit.example.com:5671",
		"password": " keep spaces ",
		"port":     float64(5671),
		"tls":      false,
		"mode":     "callback",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("NormalizeParamValues = %#v, want %#v", got, want)
	}
}

func TestNormalizeParamValuesReturnsNilWithoutValues(t *testing.T) {
	got, err := forwarderStage().NormalizeParamValues(map[string]any{"host": " ", "other": 1})
	if err != nil || got != nil {
		t.Fatalf("NormalizeParamValues = %#v, %v; want nil, nil", got, err)
	}
}

func TestNormalizeParamValuesRejectsInvalidValues(t *testing.T) {
	for name, data := range map[string]map[string]any{
		"string type":   {"host": 42},
		"secret type":   {"password": true},
		"number type":   {"port": "5671"},
		"boolean type":  {"tls": "true"},
		"select option": {"mode": "sometimes"},
		"length":        {"host": strings.Repeat("a", MaxStageParamStringLength+1)},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := forwarderStage().NormalizeParamValues(data); err == nil {
				t.Fatalf("NormalizeParamValues(%v) succeeded, want error", data)
			}
		})
	}
}

func TestMissingRequiredParamsHonoursDefaults(t *testing.T) {
	if got := forwarderStage().MissingRequiredParams(map[string]any{"queue": " "}); !reflect.DeepEqual(got, []string{"queue"}) {
		t.Fatalf("MissingRequiredParams = %v, want [queue]", got)
	}
	if got := forwarderStage().MissingRequiredParams(map[string]any{"queue": "invocations"}); len(got) != 0 {
		t.Fatalf("MissingRequiredParams = %v, want none", got)
	}
}

func TestIsSecretParam(t *testing.T) {
	stage := forwarderStage()
	if !stage.IsSecretParam("password") || stage.IsSecretParam("host") || stage.IsSecretParam("missing") {
		t.Fatal("IsSecretParam did not match the declared secret parameter only")
	}
}

func TestWorkflowRunParamsAreWireOnly(t *testing.T) {
	run := WorkflowRun{Key: "media", Params: map[string]interface{}{"password": "secret"}}
	wire, err := json.Marshal(run)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	if !strings.Contains(string(wire), `"params":{"password":"secret"}`) {
		t.Fatalf("dispatch JSON %s does not carry params", wire)
	}
	document, err := bson.Marshal(run)
	if err != nil {
		t.Fatalf("bson.Marshal: %v", err)
	}
	var stored bson.M
	if err := bson.Unmarshal(document, &stored); err != nil {
		t.Fatalf("bson.Unmarshal: %v", err)
	}
	if _, ok := stored["params"]; ok {
		t.Fatalf("persisted run %v contains params", stored)
	}
}
