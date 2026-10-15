package models

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestNormalizeDeviceLabels(t *testing.T) {
	got, err := NormalizeDeviceLabels([]string{" ANPR-Pilot ", "zone:north", "anpr-pilot", "v1.2_beta"})
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	if want := []string{"anpr-pilot", "v1.2_beta", "zone:north"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("NormalizeDeviceLabels() = %v, want %v", got, want)
	}
	if got, err := NormalizeDeviceLabels(nil); err != nil || len(got) != 0 {
		t.Fatalf("empty labels = %v, %v; want none", got, err)
	}

	invalid := []string{"", "  ", "with space", "-leading", "trailing.", "émoji", "a/b", strings.Repeat("a", MaxDeviceLabelLength+1)}
	for _, label := range invalid {
		if _, err := NormalizeDeviceLabels([]string{label}); !errors.Is(err, ErrDeviceLabelInvalid) {
			t.Errorf("NormalizeDeviceLabels(%q) error = %v, want ErrDeviceLabelInvalid", label, err)
		}
	}
	tooMany := make([]string, MaxDeviceLabels+1)
	for i := range tooMany {
		tooMany[i] = "label-" + strings.Repeat("x", i+1)
	}
	if _, err := NormalizeDeviceLabels(tooMany); !errors.Is(err, ErrDeviceLabelInvalid) {
		t.Fatalf("too many labels error = %v, want ErrDeviceLabelInvalid", err)
	}
}

func TestDevicePatchLabels(t *testing.T) {
	labels := []string{"Zone:North", "anpr-pilot", "ANPR-PILOT"}
	patch := DevicePatch{Labels: &labels}
	if err := patch.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
	if got := patch.ToSet()["labels"]; !reflect.DeepEqual(got, []string{"anpr-pilot", "zone:north"}) {
		t.Fatalf("labels = %v, want the normalized form", got)
	}

	// An empty list clears all labels.
	empty := []string{}
	if got := (DevicePatch{Labels: &empty}).ToSet()["labels"]; !reflect.DeepEqual(got, []string{}) {
		t.Fatalf("clearing labels writes %#v, want an empty list", got)
	}

	bad := []string{"has space"}
	if err := (DevicePatch{Labels: &bad}).Validate(); !errors.Is(err, ErrDeviceLabelInvalid) {
		t.Fatalf("invalid label error = %v, want ErrDeviceLabelInvalid", err)
	}
}

func TestDeviceValidateRequiresNormalizedLabels(t *testing.T) {
	device := Device{Key: "cam-1", OrganisationId: "org-1", Labels: []string{"anpr-pilot", "zone:north"}}
	if err := device.Validate(); err != nil {
		t.Fatalf("normalized labels: %v", err)
	}
	device.Labels = []string{"Zone:North"}
	if err := device.Validate(); !errors.Is(err, ErrDeviceLabelInvalid) {
		t.Fatalf("unnormalized labels error = %v, want ErrDeviceLabelInvalid", err)
	}
}

func TestTriggerLabelsMatchDeviceLabels(t *testing.T) {
	trigger := WorkflowTrigger{Type: WorkflowTriggerAutomatic, Labels: []string{"zone:north", "anpr-pilot"}}
	if err := trigger.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
	match := AutomaticTriggerRoot(WorkflowDevice{DeviceKey: "cam-1", Labels: []string{"anpr-pilot", "other"}}, WorkflowUser{OrganisationId: "org-1"})
	if !trigger.MatchesEnvelope(match) {
		t.Fatal("device with a selected label does not match")
	}
	miss := AutomaticTriggerRoot(WorkflowDevice{DeviceKey: "cam-1", Labels: []string{"other"}}, WorkflowUser{OrganisationId: "org-1"})
	if trigger.MatchesEnvelope(miss) {
		t.Fatal("device without a selected label matches")
	}
	none := AutomaticTriggerRoot(WorkflowDevice{DeviceKey: "cam-1"}, WorkflowUser{OrganisationId: "org-1"})
	if trigger.MatchesEnvelope(none) {
		t.Fatal("device without labels matches")
	}

	// Labels combine with the other scope categories with AND.
	trigger.SiteIds = []string{"site-a"}
	labelledElsewhere := AutomaticTriggerRoot(WorkflowDevice{DeviceKey: "cam-1", Labels: []string{"anpr-pilot"}, SiteIds: []string{"site-b"}}, WorkflowUser{})
	if trigger.MatchesEnvelope(labelledElsewhere) {
		t.Fatal("labels bypassed the site scope")
	}
}

func TestTriggerLabelsMustBeStoredForm(t *testing.T) {
	for _, labels := range [][]string{{"ANPR"}, {" anpr"}, {"anpr", "anpr"}, {"has space"}} {
		trigger := WorkflowTrigger{Type: WorkflowTriggerAutomatic, Labels: labels}
		if err := trigger.Validate(); err == nil {
			t.Errorf("trigger labels %q validated, want an error", labels)
		}
	}
}

func TestStartEdgeTriggerCarriesLabels(t *testing.T) {
	edge := WorkflowEdge{Trigger: &WorkflowEdgeTrigger{Labels: []string{"anpr-pilot"}}}
	scope := edge.startTriggerScope(WorkflowNode{})
	if !reflect.DeepEqual(scope.Labels, []string{"anpr-pilot"}) {
		t.Fatalf("edge scope labels = %v", scope.Labels)
	}
	nodeScope := WorkflowEdge{}.startTriggerScope(WorkflowNode{Trigger: &WorkflowTrigger{Labels: []string{"zone:north"}}})
	if !reflect.DeepEqual(nodeScope.Labels, []string{"zone:north"}) {
		t.Fatalf("legacy node scope labels = %v", nodeScope.Labels)
	}
}

func TestWorkflowUsesDeviceLabels(t *testing.T) {
	tests := []struct {
		name     string
		workflow Workflow
		want     bool
	}{
		{"none", Workflow{Triggers: []WorkflowTrigger{{Type: WorkflowTriggerAutomatic, SiteIds: []string{"s"}}}}, false},
		{"trigger shorthand", Workflow{Triggers: []WorkflowTrigger{{Type: WorkflowTriggerAutomatic, Labels: []string{"x"}}}}, true},
		{"legacy trigger", Workflow{Trigger: &WorkflowTrigger{Labels: []string{"x"}}}, true},
		{"trigger condition", Workflow{Triggers: []WorkflowTrigger{{Conditions: []WorkflowCondition{{Path: "device.labels.*", Op: ConditionOpIn, Value: []any{"x"}}}}}}, true},
		{"nested condition", Workflow{Triggers: []WorkflowTrigger{{Conditions: []WorkflowCondition{{Conditions: []WorkflowCondition{{Path: "device.labels.*", Op: ConditionOpExists}}}}}}}, true},
		{"node trigger", Workflow{Nodes: []WorkflowNode{{Trigger: &WorkflowTrigger{Labels: []string{"x"}}}}}, true},
		{"edge trigger", Workflow{Edges: []WorkflowEdge{{Trigger: &WorkflowEdgeTrigger{Labels: []string{"x"}}}}}, true},
		{"edge condition", Workflow{Edges: []WorkflowEdge{{Conditions: []WorkflowCondition{{Path: "device.labels.*", Op: ConditionOpExists}}}}}, true},
		{"similar path", Workflow{Edges: []WorkflowEdge{{Conditions: []WorkflowCondition{{Path: "device.labelsCount", Op: ConditionOpExists}}}}}, false},
	}
	for _, tt := range tests {
		if got := tt.workflow.UsesDeviceLabels(); got != tt.want {
			t.Errorf("%s: UsesDeviceLabels() = %v, want %v", tt.name, got, tt.want)
		}
	}
}
