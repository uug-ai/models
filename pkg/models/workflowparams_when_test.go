package models

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

// providerStage is the forwarder with a provider choice: shared connection
// settings, RabbitMQ-only and MQTT-only params, and a chained MQTT param.
func providerStage() WorkflowStage {
	rabbitmq := &StageParamCondition{Param: "broker", In: []any{"rabbitmq"}}
	mqtt := &StageParamCondition{Param: "broker", In: []any{"mqtt"}}
	return WorkflowStage{
		Operation: "forwarder",
		Params: []StageParam{
			{Name: "broker", Type: StageParamSelect, Required: true, Options: []string{"rabbitmq", "mqtt"}, Default: "rabbitmq"},
			{Name: "host", Type: StageParamString, Required: true},
			{Name: "password", Type: StageParamSecret, Required: true},
			{Name: "queue", Type: StageParamString, Required: true, When: rabbitmq},
			{Name: "vhost", Type: StageParamString, Default: "/", When: rabbitmq},
			{Name: "mqttPassword", Type: StageParamSecret, When: mqtt},
			{Name: "topic", Type: StageParamString, Required: true, When: mqtt},
			{Name: "qos", Type: StageParamSelect, Options: []string{"0", "1", "2"}, Default: "1", When: mqtt},
			{Name: "websocket", Type: StageParamBoolean, Default: false, When: mqtt},
			{Name: "path", Type: StageParamString, Required: true, When: &StageParamCondition{Param: "websocket", In: []any{true}}},
		},
	}
}

func activeNames(stage WorkflowStage, data map[string]any) []string {
	var names []string
	for _, param := range stage.ActiveParams(data) {
		names = append(names, param.Name)
	}
	return names
}

func TestActiveParamsFollowControllerValueOrDefault(t *testing.T) {
	stage := providerStage()
	cases := []struct {
		name string
		data map[string]any
		want []string
	}{
		{"new node uses the broker default", nil, []string{"broker", "host", "password", "queue", "vhost"}},
		{"rabbitmq", map[string]any{"broker": "rabbitmq"}, []string{"broker", "host", "password", "queue", "vhost"}},
		{"mqtt", map[string]any{"broker": "mqtt"}, []string{"broker", "host", "password", "mqttPassword", "topic", "qos", "websocket"}},
		{"mqtt over websocket chains to path", map[string]any{"broker": "mqtt", "websocket": true}, []string{"broker", "host", "password", "mqttPassword", "topic", "qos", "websocket", "path"}},
		{"websocket alone does not activate path under rabbitmq", map[string]any{"broker": "rabbitmq", "websocket": true}, []string{"broker", "host", "password", "queue", "vhost"}},
		{"invalid controller value falls back to its default", map[string]any{"broker": "kafka"}, []string{"broker", "host", "password", "queue", "vhost"}},
		{"controller with surrounding spaces is normalized", map[string]any{"broker": " mqtt "}, []string{"broker", "host", "password", "mqttPassword", "topic", "qos", "websocket"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := activeNames(stage, tc.data); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("ActiveParams = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestNormalizeParamValuesDropsAndSkipsInactiveParams(t *testing.T) {
	got, err := providerStage().NormalizeParamValues(map[string]any{
		"broker":    "mqtt",
		"host":      " broker.example.com ",
		"topic":     " plant/line-1 ",
		"queue":     "left-over-from-rabbitmq",
		"vhost":     42, // invalid, but inactive: not validated
		"websocket": false,
		"path":      "/mqtt",
	})
	if err != nil {
		t.Fatalf("NormalizeParamValues: %v", err)
	}
	want := map[string]any{"broker": "mqtt", "host": "broker.example.com", "topic": "plant/line-1", "websocket": false}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("NormalizeParamValues = %#v, want %#v", got, want)
	}

	if _, err := providerStage().NormalizeParamValues(map[string]any{"broker": "mqtt", "qos": "3"}); err == nil {
		t.Fatal("an active param with an invalid value must still fail")
	}
}

func TestApplyParamDefaultsOnlyForActiveParamsAndRemovesInactive(t *testing.T) {
	stage := providerStage()

	rabbit, err := stage.ApplyParamDefaults(map[string]any{"host": "h"})
	if err != nil {
		t.Fatal(err)
	}
	if want := map[string]any{"broker": "rabbitmq", "host": "h", "vhost": "/"}; !reflect.DeepEqual(rabbit, want) {
		t.Fatalf("rabbitmq defaults = %#v, want %#v", rabbit, want)
	}

	// A restored secret (or any kept value) of an inactive param is removed;
	// undeclared keys are left to the caller as before.
	mqtt, err := stage.ApplyParamDefaults(map[string]any{"broker": "mqtt", "password": "p", "queue": "q", "vhost": "/x", "extra": 1})
	if err != nil {
		t.Fatal(err)
	}
	if want := map[string]any{"broker": "mqtt", "password": "p", "qos": "1", "websocket": false, "extra": 1}; !reflect.DeepEqual(mqtt, want) {
		t.Fatalf("mqtt defaults = %#v, want %#v", mqtt, want)
	}
}

func TestMissingRequiredParamsIgnoresInactiveParams(t *testing.T) {
	stage := providerStage()
	if got := stage.MissingRequiredParams(map[string]any{"broker": "rabbitmq"}); !reflect.DeepEqual(got, []string{"host", "password", "queue"}) {
		t.Fatalf("rabbitmq missing = %v", got)
	}
	if got := stage.MissingRequiredParams(map[string]any{"broker": "mqtt", "websocket": true}); !reflect.DeepEqual(got, []string{"host", "password", "path", "topic"}) {
		t.Fatalf("mqtt missing = %v", got)
	}
	if got := stage.MissingRequiredParams(map[string]any{"broker": "mqtt", "host": "h", "password": "p", "topic": "t"}); got != nil {
		t.Fatalf("complete mqtt node missing = %v", got)
	}
}

func TestValidateParamConditions(t *testing.T) {
	if err := providerStage().ValidateParamConditions(); err != nil {
		t.Fatalf("provider stage: %v", err)
	}
	if err := forwarderStage().ValidateParamConditions(); err != nil {
		t.Fatalf("stage without conditions: %v", err)
	}

	when := func(param string, in ...any) *StageParamCondition { return &StageParamCondition{Param: param, In: in} }
	selectParam := StageParam{Name: "broker", Type: StageParamSelect, Options: []string{"rabbitmq", "mqtt"}}
	cases := []struct {
		name   string
		params []StageParam
		want   string
	}{
		{"later controller", []StageParam{{Name: "topic", Type: StageParamString, When: when("broker", "mqtt")}, selectParam}, "not declared before it"},
		{"unknown controller", []StageParam{{Name: "topic", Type: StageParamString, When: when("missing", "mqtt")}}, "not declared before it"},
		{"string controller", []StageParam{{Name: "host", Type: StageParamString}, {Name: "topic", Type: StageParamString, When: when("host", "x")}}, "not a select or boolean"},
		{"no values", []StageParam{selectParam, {Name: "topic", Type: StageParamString, When: when("broker")}}, "lists no values"},
		{"unknown option", []StageParam{selectParam, {Name: "topic", Type: StageParamString, When: when("broker", "kafka")}}, `"kafka" is not an option`},
		{"bool against select", []StageParam{selectParam, {Name: "topic", Type: StageParamString, When: when("broker", true)}}, "does not match select"},
		{"string against boolean", []StageParam{{Name: "tls", Type: StageParamBoolean}, {Name: "ca", Type: StageParamString, When: when("tls", "yes")}}, "is not an option"},
		{"number value", []StageParam{selectParam, {Name: "topic", Type: StageParamString, When: when("broker", 1)}}, "strings or booleans"},
		{"too deep", []StageParam{
			{Name: "a", Type: StageParamBoolean},
			{Name: "b", Type: StageParamBoolean, When: when("a", true)},
			{Name: "c", Type: StageParamBoolean, When: when("b", true)},
			{Name: "d", Type: StageParamBoolean, When: when("c", true)},
			{Name: "e", Type: StageParamString, When: when("d", true)},
		}, "nested deeper than 3"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := WorkflowStage{Params: tc.params}.ValidateParamConditions()
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("ValidateParamConditions() = %v, want error containing %q", err, tc.want)
			}
		})
	}

	// Exactly the maximum depth is allowed.
	maxDepth := WorkflowStage{Params: []StageParam{
		{Name: "a", Type: StageParamBoolean},
		{Name: "b", Type: StageParamBoolean, When: when("a", true)},
		{Name: "c", Type: StageParamBoolean, When: when("b", true)},
		{Name: "d", Type: StageParamString, When: when("c", true)},
	}}
	if err := maxDepth.ValidateParamConditions(); err != nil {
		t.Fatalf("depth %d rejected: %v", MaxStageParamConditionDepth, err)
	}
}

func TestStageParamConditionRoundTrips(t *testing.T) {
	param := StageParam{Name: "path", Type: StageParamString, When: &StageParamCondition{Param: "websocket", In: []any{true}}}

	encoded, err := json.Marshal(param)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(encoded), `"when":{"param":"websocket","in":[true]}`) {
		t.Fatalf("JSON = %s", encoded)
	}
	var fromJSON StageParam
	if err := json.Unmarshal(encoded, &fromJSON); err != nil || !reflect.DeepEqual(fromJSON, param) {
		t.Fatalf("JSON round trip = %#v, %v", fromJSON, err)
	}
	if plain, _ := json.Marshal(StageParam{Name: "host", Type: StageParamString}); strings.Contains(string(plain), "when") {
		t.Fatalf("unconditional param must omit when: %s", plain)
	}

	raw, err := bson.Marshal(param)
	if err != nil {
		t.Fatal(err)
	}
	var fromBSON StageParam
	if err := bson.Unmarshal(raw, &fromBSON); err != nil {
		t.Fatal(err)
	}
	if fromBSON.When == nil || fromBSON.When.Param != "websocket" || len(fromBSON.When.In) != 1 || fromBSON.When.In[0] != true {
		t.Fatalf("BSON round trip = %#v", fromBSON.When)
	}
	if !(WorkflowStage{Params: []StageParam{{Name: "websocket", Type: StageParamBoolean}, fromBSON}}).ActiveParams(map[string]any{"websocket": true})[1].When.In[0].(bool) {
		t.Fatal("BSON-decoded condition does not evaluate")
	}
}
