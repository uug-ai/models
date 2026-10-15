package models

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestNormalizeDeviceTags(t *testing.T) {
	got, err := NormalizeDeviceTags([]string{" ANPR-Pilot ", "zone:north", "anpr-pilot", "v1.2_beta"})
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	if want := []string{"anpr-pilot", "v1.2_beta", "zone:north"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("NormalizeDeviceTags() = %v, want %v", got, want)
	}
	if got, err := NormalizeDeviceTags(nil); err != nil || len(got) != 0 {
		t.Fatalf("empty tags = %v, %v; want none", got, err)
	}

	invalid := []string{"", "  ", "with space", "-leading", "trailing.", "émoji", "a/b", strings.Repeat("a", MaxDeviceTagLength+1)}
	for _, tag := range invalid {
		if _, err := NormalizeDeviceTags([]string{tag}); !errors.Is(err, ErrDeviceTagInvalid) {
			t.Errorf("NormalizeDeviceTags(%q) error = %v, want ErrDeviceTagInvalid", tag, err)
		}
	}
	tooMany := make([]string, MaxDeviceTags+1)
	for i := range tooMany {
		tooMany[i] = "tag-" + strings.Repeat("x", i+1)
	}
	if _, err := NormalizeDeviceTags(tooMany); !errors.Is(err, ErrDeviceTagInvalid) {
		t.Fatalf("too many tags error = %v, want ErrDeviceTagInvalid", err)
	}
}

func TestDevicePatchSystemTags(t *testing.T) {
	tags := []string{"Zone:North", "anpr-pilot", "ANPR-PILOT"}
	patch := DevicePatch{SystemTags: &tags}
	if err := patch.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
	if got := patch.ToSet()["tags.system"]; !reflect.DeepEqual(got, []string{"anpr-pilot", "zone:north"}) {
		t.Fatalf("tags = %v, want the normalized form", got)
	}

	// An empty list clears all tags.
	empty := []string{}
	if got := (DevicePatch{SystemTags: &empty}).ToSet()["tags.system"]; !reflect.DeepEqual(got, []string{}) {
		t.Fatalf("clearing tags writes %#v, want an empty list", got)
	}

	bad := []string{"has space"}
	if err := (DevicePatch{SystemTags: &bad}).Validate(); !errors.Is(err, ErrDeviceTagInvalid) {
		t.Fatalf("invalid tag error = %v, want ErrDeviceTagInvalid", err)
	}
}

func TestDeviceValidateRequiresNormalizedSystemTags(t *testing.T) {
	device := Device{Key: "cam-1", OrganisationId: "org-1", Tags: &DeviceTags{System: []string{"anpr-pilot", "zone:north"}}}
	if err := device.Validate(); err != nil {
		t.Fatalf("normalized tags: %v", err)
	}
	device.Tags.System = []string{"Zone:North"}
	if err := device.Validate(); !errors.Is(err, ErrDeviceTagInvalid) {
		t.Fatalf("unnormalized tags error = %v, want ErrDeviceTagInvalid", err)
	}
}

func TestTriggerSystemTagsMatchDeviceSystemTags(t *testing.T) {
	trigger := WorkflowTrigger{Type: WorkflowTriggerAutomatic, SystemTags: []string{"zone:north", "anpr-pilot"}}
	if err := trigger.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
	match := AutomaticTriggerRoot(WorkflowDevice{DeviceKey: "cam-1", SystemTags: []string{"anpr-pilot", "other"}}, WorkflowUser{OrganisationId: "org-1"})
	if !trigger.MatchesEnvelope(match) {
		t.Fatal("device with a selected system tag does not match")
	}
	miss := AutomaticTriggerRoot(WorkflowDevice{DeviceKey: "cam-1", SystemTags: []string{"other"}}, WorkflowUser{OrganisationId: "org-1"})
	if trigger.MatchesEnvelope(miss) {
		t.Fatal("device without a selected system tag matches")
	}
	none := AutomaticTriggerRoot(WorkflowDevice{DeviceKey: "cam-1"}, WorkflowUser{OrganisationId: "org-1"})
	if trigger.MatchesEnvelope(none) {
		t.Fatal("device without tags matches")
	}

	// System tags combine with the other scope categories with AND.
	trigger.SiteIds = []string{"site-a"}
	taggedElsewhere := AutomaticTriggerRoot(WorkflowDevice{DeviceKey: "cam-1", SystemTags: []string{"anpr-pilot"}, SiteIds: []string{"site-b"}}, WorkflowUser{})
	if trigger.MatchesEnvelope(taggedElsewhere) {
		t.Fatal("system tags bypassed the site scope")
	}
}

func TestTriggerSystemTagsMustBeStoredForm(t *testing.T) {
	for _, tags := range [][]string{{"ANPR"}, {" anpr"}, {"anpr", "anpr"}, {"has space"}} {
		trigger := WorkflowTrigger{Type: WorkflowTriggerAutomatic, SystemTags: tags}
		if err := trigger.Validate(); err == nil {
			t.Errorf("trigger tags %q validated, want an error", tags)
		}
	}
}

func TestStartEdgeTriggerCarriesSystemTags(t *testing.T) {
	edge := WorkflowEdge{Trigger: &WorkflowEdgeTrigger{SystemTags: []string{"anpr-pilot"}}}
	scope := edge.startTriggerScope(WorkflowNode{})
	if !reflect.DeepEqual(scope.SystemTags, []string{"anpr-pilot"}) {
		t.Fatalf("edge scope tags = %v", scope.SystemTags)
	}
	nodeScope := WorkflowEdge{}.startTriggerScope(WorkflowNode{Trigger: &WorkflowTrigger{SystemTags: []string{"zone:north"}}})
	if !reflect.DeepEqual(nodeScope.SystemTags, []string{"zone:north"}) {
		t.Fatalf("legacy node scope tags = %v", nodeScope.SystemTags)
	}
}

func TestWorkflowUsesSystemTags(t *testing.T) {
	tests := []struct {
		name     string
		workflow Workflow
		want     bool
	}{
		{"none", Workflow{Triggers: []WorkflowTrigger{{Type: WorkflowTriggerAutomatic, SiteIds: []string{"s"}}}}, false},
		{"trigger shorthand", Workflow{Triggers: []WorkflowTrigger{{Type: WorkflowTriggerAutomatic, SystemTags: []string{"x"}}}}, true},
		{"legacy trigger", Workflow{Trigger: &WorkflowTrigger{SystemTags: []string{"x"}}}, true},
		{"trigger condition", Workflow{Triggers: []WorkflowTrigger{{Conditions: []WorkflowCondition{{Path: "device.systemTags.*", Op: ConditionOpIn, Value: []any{"x"}}}}}}, true},
		{"nested condition", Workflow{Triggers: []WorkflowTrigger{{Conditions: []WorkflowCondition{{Conditions: []WorkflowCondition{{Path: "device.systemTags.*", Op: ConditionOpExists}}}}}}}, true},
		{"node trigger", Workflow{Nodes: []WorkflowNode{{Trigger: &WorkflowTrigger{SystemTags: []string{"x"}}}}}, true},
		{"edge trigger", Workflow{Edges: []WorkflowEdge{{Trigger: &WorkflowEdgeTrigger{SystemTags: []string{"x"}}}}}, true},
		{"edge condition", Workflow{Edges: []WorkflowEdge{{Conditions: []WorkflowCondition{{Path: "device.systemTags.*", Op: ConditionOpExists}}}}}, true},
		{"similar path", Workflow{Edges: []WorkflowEdge{{Conditions: []WorkflowCondition{{Path: "device.systemTagsCount", Op: ConditionOpExists}}}}}, false},
	}
	for _, tt := range tests {
		if got := tt.workflow.UsesSystemTags(); got != tt.want {
			t.Errorf("%s: UsesSystemTags() = %v, want %v", tt.name, got, tt.want)
		}
	}
}
