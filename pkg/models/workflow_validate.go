package models

import (
	"errors"
	"fmt"
	"strings"
)

// ErrInvalidWorkflowGraph wraps every ValidateGraph failure so API callers can
// map it to a client error with errors.Is.
var ErrInvalidWorkflowGraph = errors.New("invalid workflow graph")

var workflowConditionOps = map[ConditionOp]bool{
	ConditionOpEq: true, ConditionOpNe: true, ConditionOpContains: true, ConditionOpIn: true,
	ConditionOpExists: true, ConditionOpMatches: true, ConditionOpGt: true, ConditionOpGte: true,
	ConditionOpLt: true, ConditionOpLte: true,
}

// ValidateGraph checks that the node/edge graph compiles to a runnable stage
// set: unique node ids, known node types, at most one device node, one stage
// node per operation (never the reserved seed operation), edges between existing
// nodes that never feed a device node or form a cycle, and well-formed edge
// conditions. It does not check operations against a deployment catalog; callers
// own that allow-list.
func (w *Workflow) ValidateGraph() error {
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
		case WorkflowNodeDevice:
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
		if target.EffectiveType() == WorkflowNodeDevice {
			return invalid("edge %q feeds device node %q", e.Id, e.Target)
		}
		if e.Condition != nil {
			if !workflowConditionOps[e.Condition.Op] {
				return invalid("edge %q has unknown condition operator %q", e.Id, e.Condition.Op)
			}
			if err := ValidateStageCondition(e.Condition); err != nil {
				return invalid("edge %q: %v", e.Id, err)
			}
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
