package models

import "fmt"

// ConditionMode combines predicates within one condition set, independently of
// NeedsMode, which combines a stage's operation-gated dependencies.
type ConditionMode string

const (
	ConditionModeAll ConditionMode = "all"
	ConditionModeAny ConditionMode = "any"
)

// WorkflowConditionSet applies absolute conditions to the workflow root. An
// empty mode means all; an empty set imposes no additional restriction.
type WorkflowConditionSet struct {
	ConditionMode ConditionMode       `json:"conditionMode,omitempty" bson:"conditionMode,omitempty"`
	Conditions    []WorkflowCondition `json:"conditions,omitempty" bson:"conditions,omitempty"`
}

// WorkflowPredicate is a scalar predicate relative to one anyMatch array
// element. It deliberately cannot contain another predicate group.
type WorkflowPredicate struct {
	Path  string      `json:"path" bson:"path"`
	Op    ConditionOp `json:"op" bson:"op"`
	Value any         `json:"value" bson:"value"`
	// Field is the contract field ID the condition was authored with (for
	// example a Start or stage contract field), so editors can show the value
	// in that field's section and widget again. It is metadata only: evaluation
	// uses Path, Op and Value, and conditions without Field remain valid.
	Field string `json:"field,omitempty" bson:"field,omitempty"`
}

// WorkflowPredicateSet combines predicates against the same array object.
// Unlike the top-level set, an anyMatch predicate set must not be empty.
type WorkflowPredicateSet struct {
	ConditionMode ConditionMode       `json:"conditionMode,omitempty" bson:"conditionMode,omitempty"`
	Conditions    []WorkflowPredicate `json:"conditions" bson:"conditions"`
}

// NormalizeWorkflowConditions reads either the plural or legacy representation
// without modifying either. A nonnil plural slice, even if empty, conflicts
// with a legacy condition. Validation of predicates is a separate operation.
func NormalizeWorkflowConditions(mode ConditionMode, conditions []WorkflowCondition, legacy *WorkflowCondition) (WorkflowConditionSet, error) {
	if err := validateConditionMode(mode); err != nil {
		return WorkflowConditionSet{}, err
	}
	if legacy != nil {
		if conditions != nil {
			return WorkflowConditionSet{}, fmt.Errorf("legacy condition and plural conditions are ambiguous")
		}
		return WorkflowConditionSet{ConditionMode: ConditionModeAll, Conditions: []WorkflowCondition{*legacy}}, nil
	}
	if mode == "" {
		mode = ConditionModeAll
	}
	return WorkflowConditionSet{ConditionMode: mode, Conditions: conditions}, nil
}

func validateConditionMode(mode ConditionMode) error {
	switch mode {
	case "", ConditionModeAll, ConditionModeAny:
		return nil
	default:
		return fmt.Errorf("unknown condition mode %q (want all or any)", mode)
	}
}
