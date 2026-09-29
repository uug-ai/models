package models

import (
	"fmt"
	"time"
)

// WorkflowRunDecisionOutcome describes an evaluation, not a stage's lifecycle.
// Waiting is used only for a dependency whose upstream operation is unavailable.
type WorkflowRunDecisionOutcome string

const (
	WorkflowRunDecisionPassed       WorkflowRunDecisionOutcome = "passed"
	WorkflowRunDecisionFailed       WorkflowRunDecisionOutcome = "failed"
	WorkflowRunDecisionWaiting      WorkflowRunDecisionOutcome = "waiting"
	WorkflowRunDecisionNotEvaluated WorkflowRunDecisionOutcome = "notEvaluated"
)

// WorkflowRunStageDecision is one bounded evaluation snapshot, not an event log.
// It references the containing stage's frozen rules and contains no input values.
type WorkflowRunStageDecision struct {
	EvaluatedAtMs int64                          `json:"evaluatedAtMs" bson:"evaluatedatms"`
	Eligible      bool                           `json:"eligible" bson:"eligible"`
	Needs         []WorkflowRunStageNeedDecision `json:"needs,omitempty" bson:"needs,omitempty"`
}

// WorkflowRunStageNeedDecision references Needs[Index] on the containing stage.
// Ready records gate availability even when stage-level short-circuiting means
// this need was not evaluated. An ungated need is always ready.
type WorkflowRunStageNeedDecision struct {
	Index      int                            `json:"index" bson:"index"`
	Ready      bool                           `json:"ready" bson:"ready"`
	Outcome    WorkflowRunDecisionOutcome     `json:"outcome" bson:"outcome"`
	Conditions []WorkflowRunConditionDecision `json:"conditions,omitempty" bson:"conditions,omitempty"`
}

// WorkflowRunConditionDecision references the dependency's normalized condition
// list: a legacy singular condition has index zero. anyMatch is one aggregate
// check, not independent successes across different objects or a per-object log.
type WorkflowRunConditionDecision struct {
	Index   int                        `json:"index" bson:"index"`
	Outcome WorkflowRunDecisionOutcome `json:"outcome" bson:"outcome"`
}

// EvaluateDecision evaluates this run stage's routing without mutating the run,
// rules, or inputs. It uses the same condition evaluator as dependency matching,
// recording only checks actually reached; skipped checks stay notEvaluated.
// Empty dispatch and conditional stages without needs are unconditional, matching
// registry normalization. Other conditional stages default to needsMode any.
//
// This is explanation data, not a dispatch claim or persistence operation. A
// writer must guard updates against newer evaluations, run closure and dispatch,
// retain the last evaluation for stages that never dispatch, and freeze the
// decision used for dispatch rather than recomputing it after the fact.
func (s WorkflowRunStage) EvaluateDecision(root map[string]any, available map[string]bool, at time.Time) (*WorkflowRunStageDecision, error) {
	if err := checkRunStageOperation(make(map[string]struct{}), s.Operation, "stages"); err != nil {
		return nil, err
	}
	decision := &WorkflowRunStageDecision{EvaluatedAtMs: at.UnixMilli()}
	switch s.Dispatch {
	case "", DispatchAlways:
		decision.Eligible = true
		return decision, nil
	case DispatchConditional:
		if len(s.Needs) == 0 {
			decision.Eligible = true
			return decision, nil
		}
	default:
		return nil, fmt.Errorf("stage %q has unknown dispatch %q", s.Operation, s.Dispatch)
	}
	switch s.NeedsMode {
	case "", NeedsModeAny, NeedsModeAll:
	default:
		return nil, fmt.Errorf("stage %q has unknown needsMode %q", s.Operation, s.NeedsMode)
	}

	sets := make([]WorkflowConditionSet, len(s.Needs))
	decision.Needs = make([]WorkflowRunStageNeedDecision, len(s.Needs))
	for i, need := range s.Needs {
		set, err := need.ConditionSet()
		if err != nil {
			return nil, fmt.Errorf("stage %q need %d: %w", s.Operation, i, err)
		}
		if err := ValidateWorkflowConditionSet(set); err != nil {
			return nil, fmt.Errorf("stage %q need %d: %w", s.Operation, i, err)
		}
		sets[i] = set
		check := WorkflowRunStageNeedDecision{
			Index: i, Ready: need.Operation == "" || available[need.Operation],
			Outcome: WorkflowRunDecisionNotEvaluated,
		}
		if len(set.Conditions) > 0 {
			check.Conditions = make([]WorkflowRunConditionDecision, len(set.Conditions))
			for j := range set.Conditions {
				check.Conditions[j] = WorkflowRunConditionDecision{Index: j, Outcome: WorkflowRunDecisionNotEvaluated}
			}
		}
		decision.Needs[i] = check
	}

	all := s.NeedsMode == NeedsModeAll
	for i := range decision.Needs {
		need := &decision.Needs[i]
		matched := false
		if !need.Ready {
			need.Outcome = WorkflowRunDecisionWaiting
		} else {
			matched = evaluateConditionSet(sets[i], root, func(index int, passed bool) {
				need.Conditions[index].Outcome = workflowDecisionOutcome(passed)
			})
			need.Outcome = workflowDecisionOutcome(matched)
		}
		if !all && matched {
			decision.Eligible = true
			return decision, nil
		}
		if all && !matched {
			return decision, nil
		}
	}
	decision.Eligible = all
	return decision, nil
}

func workflowDecisionOutcome(passed bool) WorkflowRunDecisionOutcome {
	if passed {
		return WorkflowRunDecisionPassed
	}
	return WorkflowRunDecisionFailed
}
