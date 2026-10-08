package models

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func float(value float64) *float64 { return &value }

func anprLikeStage() WorkflowStage {
	return WorkflowStage{Operation: "anpr", Params: []StageParam{
		{Name: "classes", Type: StageParamMultiSelect, Options: []string{"car", "lorry", "motorbike"}, Default: []any{"car"}},
		{Name: "confidence", Type: StageParamNumber, Minimum: float(0), Maximum: float(1), Default: 0.0},
		{Name: "frames", Type: StageParamNumber, Minimum: float(1), Maximum: float(5), Default: 1},
		{Name: "label", Type: StageParamString},
		{Name: "token", Type: StageParamSecret, Default: "never"},
	}}
}

func TestMultiSelectParams(t *testing.T) {
	stage := anprLikeStage()
	for name, tt := range map[string]struct {
		raw   any
		want  any
		valid bool
	}{
		"list":            {[]any{"lorry", " car ", "lorry"}, []string{"lorry", "car"}, true},
		"typed list":      {[]string{"motorbike"}, []string{"motorbike"}, true},
		"bson array":      {primitive.A{"car"}, []string{"car"}, true},
		"empty is unset":  {[]any{" ", ""}, nil, true},
		"outside options": {[]any{"bus"}, nil, false},
		"not a list":      {"car", nil, false},
		"mixed types":     {[]any{"car", 3}, nil, false},
	} {
		t.Run(name, func(t *testing.T) {
			got, err := stage.NormalizeParamValues(map[string]any{"classes": tt.raw})
			if (err == nil) != tt.valid {
				t.Fatalf("valid %v, error %v", tt.valid, err)
			}
			if tt.valid && !reflect.DeepEqual(got["classes"], tt.want) && !(tt.want == nil && got == nil) {
				t.Fatalf("classes = %#v, want %#v", got["classes"], tt.want)
			}
		})
	}
}

func TestNumberParamBounds(t *testing.T) {
	stage := anprLikeStage()
	for raw, valid := range map[float64]bool{0: true, 0.5: true, 1: true, -0.1: false, 1.01: false} {
		if _, err := stage.NormalizeParamValues(map[string]any{"confidence": raw}); (err == nil) != valid {
			t.Errorf("confidence %v: valid %v, error %v", raw, valid, err)
		}
	}
	if _, err := stage.NormalizeParamValues(map[string]any{"label": "unbounded"}); err != nil {
		t.Fatal(err)
	}
}

func TestApplyParamDefaults(t *testing.T) {
	stage := anprLikeStage()
	got, err := stage.ApplyParamDefaults(map[string]any{"frames": 3.0})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]any{"classes": []string{"car"}, "confidence": 0.0, "frames": 3.0}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("defaults = %#v, want %#v (existing values kept, secrets never defaulted)", got, want)
	}
	if got, _ := (WorkflowStage{}).ApplyParamDefaults(nil); got != nil {
		t.Fatalf("no params, no data = %#v", got)
	}
	bad := WorkflowStage{Params: []StageParam{{Name: "frames", Type: StageParamNumber, Maximum: float(2), Default: 9}}}
	if _, err := bad.ApplyParamDefaults(nil); err == nil {
		t.Fatal("a default outside its bounds must be reported")
	}
	if missing := stage.MissingRequiredParams(got); len(missing) != 0 {
		t.Fatalf("missing = %v", missing)
	}
}
