package models

import (
	"errors"
	"strings"
	"testing"
)

func TestWorkflow_ValidateGraph_AcceptsEditorGraphs(t *testing.T) {
	vlm := vlmEditorWorkflow()
	for name, w := range map[string]Workflow{
		"device to vlm to forwarder": vlm,
		"empty graph":                {},
		"stage-only graph":           {Nodes: []WorkflowNode{{Id: "n1", StageRef: "vlm"}}},
	} {
		if err := w.ValidateGraph(); err != nil {
			t.Errorf("%s: ValidateGraph() = %v, want nil", name, err)
		}
	}
}

func TestWorkflow_ValidateGraph_Rejects(t *testing.T) {
	device := WorkflowNode{Id: "device", Type: WorkflowNodeDevice}
	vlm := WorkflowNode{Id: "vlm", StageRef: "vlm"}
	forwarder := WorkflowNode{Id: "forwarder", StageRef: "forwarder"}

	tests := map[string]struct {
		workflow Workflow
		want     string
	}{
		"empty node id":         {Workflow{Nodes: []WorkflowNode{{StageRef: "vlm"}}}, "empty id"},
		"duplicate node id":     {Workflow{Nodes: []WorkflowNode{vlm, {Id: "vlm", StageRef: "forwarder"}}}, "duplicate node id"},
		"unknown node type":     {Workflow{Nodes: []WorkflowNode{{Id: "m", Type: "model"}}}, "unknown type"},
		"second device node":    {Workflow{Nodes: []WorkflowNode{device, {Id: "device-2", Type: WorkflowNodeDevice}}}, "only one device node"},
		"empty device key":      {Workflow{Nodes: []WorkflowNode{{Id: "device", Type: WorkflowNodeDevice, Devices: []DeviceKey{{Key: " "}}}}}, "empty key"},
		"stage without ref":     {Workflow{Nodes: []WorkflowNode{{Id: "s"}}}, "no stageRef"},
		"reserved operation":    {Workflow{Nodes: []WorkflowNode{{Id: "s", StageRef: WorkflowSeedOperation}}}, "reserved operation"},
		"duplicate operation":   {Workflow{Nodes: []WorkflowNode{vlm, {Id: "vlm-2", StageRef: "vlm"}}}, "placed more than once"},
		"dangling source":       {Workflow{Nodes: []WorkflowNode{vlm}, Edges: []WorkflowEdge{{Id: "e", Source: "gone", Target: "vlm"}}}, "unknown source"},
		"dangling target":       {Workflow{Nodes: []WorkflowNode{vlm}, Edges: []WorkflowEdge{{Id: "e", Source: "vlm", Target: "gone"}}}, "unknown target"},
		"self loop":             {Workflow{Nodes: []WorkflowNode{vlm}, Edges: []WorkflowEdge{{Id: "e", Source: "vlm", Target: "vlm"}}}, "itself"},
		"edge into device node": {Workflow{Nodes: []WorkflowNode{device, vlm}, Edges: []WorkflowEdge{{Id: "e", Source: "vlm", Target: "device"}}}, "feeds device node"},
		"cycle": {Workflow{Nodes: []WorkflowNode{vlm, forwarder}, Edges: []WorkflowEdge{
			{Id: "e1", Source: "vlm", Target: "forwarder"},
			{Id: "e2", Source: "forwarder", Target: "vlm"},
		}}, "cycle"},
		"unknown operator": {Workflow{Nodes: []WorkflowNode{device, vlm}, Edges: []WorkflowEdge{
			{Id: "e", Source: "device", Target: "vlm", Condition: &StageCondition{Path: "inputs.classify.objectCount", Op: "between"}},
		}}, "unknown condition operator"},
		"credential path": {Workflow{Nodes: []WorkflowNode{device, vlm}, Edges: []WorkflowEdge{
			{Id: "e", Source: "device", Target: "vlm", Condition: &StageCondition{Path: "user.storage.secret", Op: ConditionOpExists}},
		}}, "credentials"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := tt.workflow.ValidateGraph()
			if !errors.Is(err, ErrInvalidWorkflowGraph) || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("ValidateGraph() = %v, want ErrInvalidWorkflowGraph containing %q", err, tt.want)
			}
		})
	}
}

func TestValidateStageCondition(t *testing.T) {
	valid := []*StageCondition{
		nil,
		{Path: "inputs.classify.details.*.classified", Op: ConditionOpIn, Value: []any{"car"}},
		{Path: "inputs.classify.objectCount", Op: ConditionOpGt, Value: 0},
		{Path: "results.anpr.detections.*.tracks.*.confidence", Op: ConditionOpGte, Value: 0.5},
		{Path: "device.deviceKey", Op: ConditionOpMatches, Value: "^office-"},
		{Path: "user.organisationId", Op: ConditionOpEq, Value: "org-1"},
		{Path: "key", Op: ConditionOpExists},
	}
	for _, c := range valid {
		if err := ValidateStageCondition(c); err != nil {
			t.Errorf("ValidateStageCondition(%+v) = %v, want nil", c, err)
		}
	}

	invalid := map[string]*StageCondition{
		"empty path":            {Op: ConditionOpExists},
		"empty segment":         {Path: "inputs..classify", Op: ConditionOpExists},
		"storage root":          {Path: "storage.secret", Op: ConditionOpExists},
		"user storage":          {Path: "user.storage", Op: ConditionOpExists},
		"unknown root":          {Path: "secrets.x", Op: ConditionOpExists},
		"traversed scalar":      {Path: "key.x", Op: ConditionOpExists},
		"unknown device field":  {Path: "device.password", Op: ConditionOpExists},
		"unknown classify":      {Path: "inputs.classify.unknown", Op: ConditionOpExists},
		"array without fan-out": {Path: "inputs.classify.details.classified", Op: ConditionOpEq, Value: "car"},
		"non-string regex":      {Path: "device.deviceKey", Op: ConditionOpMatches, Value: 1},
		"invalid regex":         {Path: "device.deviceKey", Op: ConditionOpMatches, Value: "("},
	}
	for name, c := range invalid {
		if err := ValidateStageCondition(c); err == nil {
			t.Errorf("%s: ValidateStageCondition(%+v) = nil, want error", name, c)
		}
	}
}
