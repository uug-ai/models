package models

import (
	"errors"
	"fmt"
	"strings"
)

// ErrInvalidWorkflowGraph wraps every ValidateGraph failure so API callers can
// map it to a client error with errors.Is.
var ErrInvalidWorkflowGraph = errors.New("invalid workflow graph")

// ValidateGraph checks that the node/edge graph compiles to a runnable stage
// set: unique node ids, known node types, at most one Start/device root, one stage
// node per operation (never the reserved seed operation), edges between existing
// nodes that never feed a root or form a cycle, and well-formed edge
// conditions. It does not check operations against a deployment catalog; callers
// own that allow-list.
func (w *Workflow) ValidateGraph() error {
	if err := w.ValidateStartNodes(); err != nil {
		return err
	}
	invalid := func(format string, args ...any) error {
		return fmt.Errorf("%w: %s", ErrInvalidWorkflowGraph, fmt.Sprintf(format, args...))
	}

	nodes := make(map[string]WorkflowNode, len(w.Nodes))
	operations := make(map[string]bool, len(w.Nodes))
	deviceNodes := 0
	for _, n := range w.Nodes {
		if strings.TrimSpace(n.Id) == "" {
			return invalid("node has an empty id")
		}
		if _, duplicate := nodes[n.Id]; duplicate {
			return invalid("duplicate node id %q", n.Id)
		}
		nodes[n.Id] = n
		switch n.EffectiveType() {
		case WorkflowNodeDevice, WorkflowNodeStart:
			if deviceNodes++; deviceNodes > 1 {
				return invalid("a workflow may have only one device node")
			}
			for _, d := range n.Devices {
				if strings.TrimSpace(d.Key) == "" {
					return invalid("device node %q selects a device with an empty key", n.Id)
				}
			}
		case WorkflowNodeStage:
			if strings.TrimSpace(n.StageRef) == "" {
				return invalid("stage node %q has no stageRef", n.Id)
			}
			if n.StageRef == WorkflowSeedOperation {
				return invalid("stage node %q uses the reserved operation %q", n.Id, n.StageRef)
			}
			if operations[n.StageRef] {
				return invalid("operation %q is placed more than once", n.StageRef)
			}
			operations[n.StageRef] = true
		default:
			return invalid("node %q has unknown type %q", n.Id, n.Type)
		}
	}

	outgoing := make(map[string][]string, len(w.Edges))
	for _, e := range w.Edges {
		if _, ok := nodes[e.Source]; !ok {
			return invalid("edge %q references unknown source node %q", e.Id, e.Source)
		}
		target, ok := nodes[e.Target]
		if !ok {
			return invalid("edge %q references unknown target node %q", e.Id, e.Target)
		}
		if e.Source == e.Target {
			return invalid("edge %q connects node %q to itself", e.Id, e.Source)
		}
		if target.IsRoot() {
			return invalid("edge %q feeds %s node %q", e.Id, target.EffectiveType(), e.Target)
		}

		if err := e.ValidateConditions(); err != nil {
			return invalid("edge %q: %v", e.Id, err)
		}
		outgoing[e.Source] = append(outgoing[e.Source], e.Target)
	}

	const (
		visiting = 1
		visited  = 2
	)
	state := make(map[string]int, len(nodes))
	var acyclic func(id string) bool
	acyclic = func(id string) bool {
		switch state[id] {
		case visiting:
			return false
		case visited:
			return true
		}
		state[id] = visiting
		for _, next := range outgoing[id] {
			if !acyclic(next) {
				return false
			}
		}
		state[id] = visited
		return true
	}
	for _, n := range w.Nodes {
		if !acyclic(n.Id) {
			return invalid("edges form a cycle")
		}
	}
	return nil
}

// ValidateStartNodes validates the explicit activation contract, including for
// drafts. A disabled manual Start may omit surfaces while it is being edited.
// Legacy graphs retain their existing draft validation behavior.
func (w *Workflow) ValidateStartNodes() error {
	roots := 0
	for _, node := range w.Nodes {
		if node.IsRoot() {
			roots++
		}
	}
	for _, node := range w.Nodes {
		if node.EffectiveType() != WorkflowNodeStart {
			continue
		}
		invalid := func(message string) error {
			return fmt.Errorf("%w: start node %q: %s", ErrInvalidWorkflowGraph, node.Id, message)
		}
		if roots > 1 {
			return invalid("a workflow may have only one start or device node")
		}
		for _, edge := range w.Edges {
			if edge.Target == node.Id {
				return invalid("cannot have incoming edges")
			}
		}
		if node.Trigger == nil {
			return invalid("requires a trigger")
		}
		trigger := *node.Trigger
		if trigger.Type != WorkflowTriggerAutomatic && trigger.Type != WorkflowTriggerManual {
			return invalid(fmt.Sprintf("unknown trigger type %q", trigger.Type))
		}
		for _, surface := range trigger.Surfaces {
			switch surface {
			case WorkflowSurfaceCase, WorkflowSurfaceMedia, WorkflowSurfaceRedaction:
			default:
				return invalid(fmt.Sprintf("unknown trigger surface %q", surface))
			}
		}
		if !w.Enabled && trigger.Type == WorkflowTriggerManual && len(trigger.Surfaces) == 0 {
			continue
		}
		if trigger.Type == WorkflowTriggerAutomatic {
			trigger.Devices = node.Devices
		}
		if err := trigger.Validate(); err != nil {
			return invalid(err.Error())
		}
	}
	return nil
}

// Validate checks activation settings without changing legacy documents.
// Manual triggers still ignore automatic-only fields, including conditions.
// Membership access and deployment-specific input schemas belong to callers.
func (t WorkflowTrigger) Validate() error {
	switch t.EffectiveType() {
	case WorkflowTriggerManual:
		if len(t.Surfaces) == 0 {
			return fmt.Errorf("manual trigger requires at least one surface")
		}
		for _, surface := range t.Surfaces {
			switch surface {
			case WorkflowSurfaceCase, WorkflowSurfaceMedia, WorkflowSurfaceRedaction:
			default:
				return fmt.Errorf("unknown manual trigger surface %q", surface)
			}
		}
		return nil
	case WorkflowTriggerAutomatic:
	default:
		return fmt.Errorf("unknown trigger type %q", t.Type)
	}
	for _, device := range t.Devices {
		if strings.TrimSpace(device.Key) == "" {
			return fmt.Errorf("trigger selects a device with an empty key")
		}
	}
	for name, ids := range map[string][]string{"siteIds": t.SiteIds, "groupIds": t.GroupIds} {
		for _, id := range ids {
			if strings.TrimSpace(id) == "" {
				return fmt.Errorf("trigger %s contains an empty id", name)
			}
		}
	}
	for _, c := range t.Conditions {
		if c.Path == "results" || strings.HasPrefix(c.Path, "results.") {
			return fmt.Errorf("automatic trigger cannot reference future results: %q", c.Path)
		}
	}
	return ValidateWorkflowConditionSet(t.ConditionSet())
}

// ValidateTriggers validates an explicit Start, otherwise the canonical list or
// legacy single trigger. It does not normalize or rewrite persisted data.
func (w *Workflow) ValidateTriggers() error {
	for _, node := range w.Nodes {
		if node.EffectiveType() == WorkflowNodeStart {
			return w.ValidateStartNodes()
		}
	}
	triggers := w.Triggers
	if len(triggers) == 0 && w.Trigger != nil {
		triggers = []WorkflowTrigger{*w.Trigger}
	}
	for i, trigger := range triggers {
		if err := trigger.Validate(); err != nil {
			return fmt.Errorf("trigger %d: %w", i, err)
		}
	}
	return nil
}
