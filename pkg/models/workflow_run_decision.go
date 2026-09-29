package models

// WorkflowRunDecisionOutcome summarizes an edge check for display only.
// Waiting means its upstream operation was unavailable; notEvaluated means the
// check was skipped, for example by short-circuiting.
type WorkflowRunDecisionOutcome string

const (
	WorkflowRunDecisionPassed       WorkflowRunDecisionOutcome = "passed"
	WorkflowRunDecisionFailed       WorkflowRunDecisionOutcome = "failed"
	WorkflowRunDecisionWaiting      WorkflowRunDecisionOutcome = "waiting"
	WorkflowRunDecisionNotEvaluated WorkflowRunDecisionOutcome = "notEvaluated"
)

// WorkflowRunStageDecision is a read-only summary of the engine's existing
// routing evaluation. It must never be used to decide whether to dispatch.
type WorkflowRunStageDecision struct {
	EvaluatedAtMs int64                          `json:"evaluatedAtMs" bson:"evaluatedatms"`
	Eligible      bool                           `json:"eligible" bson:"eligible"`
	Needs         []WorkflowRunStageNeedDecision `json:"needs,omitempty" bson:"needs,omitempty"`
}

// WorkflowRunStageNeedDecision references Needs[Index] on the containing stage.
// It summarizes the whole edge, not its individual predicates or input values.
type WorkflowRunStageNeedDecision struct {
	Index   int                        `json:"index" bson:"index"`
	Outcome WorkflowRunDecisionOutcome `json:"outcome" bson:"outcome"`
}
