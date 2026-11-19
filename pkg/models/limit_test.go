package models

import (
	"encoding/json"
	"errors"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"gopkg.in/yaml.v3"
)

func fullLimits() Limits {
	return Limits{
		RetentionDays:    NewLimit(30),
		UploadGBPerDay:   NewLimit(10),
		UploadGBPerMonth: NewLimit(300),
		WorkflowCredits:  NewLimit(0),
		Devices:          NewUnlimitedLimit(),
		Projects:         NewUnlimitedLimit(),
		Members:          NewLimit(5),
	}
}

func equalLimits(t *testing.T, got, want Limits) {
	t.Helper()
	gotFields, wantFields := got.limitFields(), want.limitFields()
	for index, field := range wantFields {
		g, w := *gotFields[index].field, *field.field
		if (g == nil) != (w == nil) || (g != nil && *g != *w) {
			t.Errorf("%s = %v, want %v", field.name, g, w)
		}
	}
}

func TestLimitAllows(t *testing.T) {
	tests := []struct {
		limit Limit
		used  int64
		want  bool
	}{
		{Limit{Value: 3}, 0, true},
		{Limit{Value: 3}, 3, true},
		{Limit{Value: 3}, 4, false},
		{Limit{Value: 0}, 0, true},
		{Limit{Value: 0}, 1, false},
		{Limit{Unlimited: true}, 1 << 62, true},
	}
	for _, test := range tests {
		if got := test.limit.Allows(test.used); got != test.want {
			t.Errorf("%v.Allows(%d) = %v, want %v", test.limit, test.used, got, test.want)
		}
	}
}

func TestLimitsJSONRoundTrip(t *testing.T) {
	data, err := json.Marshal(fullLimits())
	if err != nil {
		t.Fatal(err)
	}
	want := `{"retentionDays":30,"uploadGBPerDay":10,"uploadGBPerMonth":300,"workflowCredits":0,"devices":"unlimited","projects":"unlimited","members":5}`
	if string(data) != want {
		t.Fatalf("json = %s\nwant %s", data, want)
	}
	var decoded Limits
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	equalLimits(t, decoded, fullLimits())
}

func TestLimitsJSONMissingAndNull(t *testing.T) {
	var limits Limits
	if err := json.Unmarshal([]byte(`{"devices":"unlimited","members":null}`), &limits); err != nil {
		t.Fatal(err)
	}
	if limits.Devices == nil || !limits.Devices.Unlimited {
		t.Errorf("devices = %v", limits.Devices)
	}
	if limits.Members != nil || limits.RetentionDays != nil {
		t.Error("missing and null limits stay unset")
	}
	if data, _ := json.Marshal(Limits{}); string(data) != "{}" {
		t.Errorf("empty limits = %s", data)
	}
}

func TestLimitJSONRejectsInvalid(t *testing.T) {
	for _, input := range []string{`-1`, `1.5`, `"lots"`, `"100"`, `"Unlimited"`, `true`, `{}`} {
		var limit Limit
		if err := json.Unmarshal([]byte(input), &limit); err == nil {
			t.Errorf("Unmarshal(%s) accepted %v", input, limit)
		}
	}
	if _, err := json.Marshal(Limit{Value: -1}); !errors.Is(err, ErrLimitInvalid) {
		t.Errorf("Marshal(-1) error = %v", err)
	}
}

func TestLimitsYAMLRoundTrip(t *testing.T) {
	data, err := yaml.Marshal(fullLimits())
	if err != nil {
		t.Fatal(err)
	}
	var decoded Limits
	if err := yaml.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("%v\n%s", err, data)
	}
	equalLimits(t, decoded, fullLimits())

	var file Limits
	if err := yaml.Unmarshal([]byte("devices: unlimited\nuploadGBPerDay: 100\n"), &file); err != nil {
		t.Fatal(err)
	}
	if !file.Devices.Unlimited || file.UploadGBPerDay.Value != 100 || file.Members != nil {
		t.Errorf("limits = %+v", file)
	}
}

func TestLimitYAMLRejectsInvalid(t *testing.T) {
	for _, input := range []string{`devices: -1`, `devices: 1.5`, `devices: lots`, `devices: "100"`, `devices: [1]`, `devices: {a: 1}`, `devices: true`} {
		var limits Limits
		if err := yaml.Unmarshal([]byte(input), &limits); err == nil {
			t.Errorf("Unmarshal(%s) accepted %v", input, limits.Devices)
		}
	}
}

func TestLimitsBSONRoundTrip(t *testing.T) {
	data, err := bson.Marshal(fullLimits())
	if err != nil {
		t.Fatal(err)
	}
	raw := bson.Raw(data)
	if value := raw.Lookup("devices"); value.StringValue() != "unlimited" {
		t.Errorf("devices stored as %v", value)
	}
	if value := raw.Lookup("retentionDays"); value.Int64() != 30 {
		t.Errorf("retentionDays stored as %v", value)
	}
	var decoded Limits
	if err := bson.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	equalLimits(t, decoded, fullLimits())

	data, _ = bson.Marshal(Limits{})
	if elements, _ := bson.Raw(data).Elements(); len(elements) != 0 {
		t.Error("unset limits are omitted")
	}
}

func TestLimitBSONAcceptsShellNumbers(t *testing.T) {
	data, _ := bson.Marshal(bson.M{"devices": int32(3), "members": float64(7), "projects": nil})
	var limits Limits
	if err := bson.Unmarshal(data, &limits); err != nil {
		t.Fatal(err)
	}
	if limits.Devices.Value != 3 || limits.Members.Value != 7 || limits.Projects != nil {
		t.Errorf("limits = %v %v %v", limits.Devices, limits.Members, limits.Projects)
	}
}

func TestLimitBSONRejectsInvalid(t *testing.T) {
	for _, value := range []any{"lots", 1.5, int64(-1), true} {
		data, _ := bson.Marshal(bson.M{"devices": value})
		var limits Limits
		if err := bson.Unmarshal(data, &limits); err == nil {
			t.Errorf("Unmarshal(%v) accepted %v", value, limits.Devices)
		}
	}
}

func TestLimitsWithOverrides(t *testing.T) {
	plan := fullLimits()
	overrides := Limits{Devices: NewLimit(50), UploadGBPerDay: NewUnlimitedLimit(), RetentionDays: NewLimit(90)}
	merged := plan.WithOverrides(overrides)

	want := fullLimits()
	want.Devices, want.UploadGBPerDay, want.RetentionDays = NewLimit(50), NewUnlimitedLimit(), NewLimit(90)
	equalLimits(t, merged, want)
	equalLimits(t, plan, fullLimits())

	merged.Members.Value = 99
	overrides.Devices.Value = 1
	if plan.Members.Value != 5 || merged.Devices.Value != 50 {
		t.Error("merged limits share pointers with their inputs")
	}
	equalLimits(t, plan.WithOverrides(Limits{}), fullLimits())
}

func TestLimitsMissing(t *testing.T) {
	if missing := fullLimits().Missing(); len(missing) != 0 {
		t.Errorf("missing = %v", missing)
	}
	limits := fullLimits()
	limits.Devices, limits.Members = nil, nil
	if missing := limits.Missing(); len(missing) != 2 || missing[0] != "devices" || missing[1] != "members" {
		t.Errorf("missing = %v", missing)
	}
	if len((Limits{}).Missing()) != len(LimitNames()) {
		t.Error("empty limits miss every limit")
	}
}
