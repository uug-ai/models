package models

import (
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
)

// WorkflowRunStage combines one run's immutable execution rules with its mutable
// lifecycle facts. It is not a deployment/catalog WorkflowStage. Every planned
// stage is captured, including stages that never dispatch.
type WorkflowRunStage struct {
	Operation string            `json:"operation" bson:"operation"`
	Name      string            `json:"name,omitempty" bson:"name,omitempty"`
	Queue     string            `json:"queue,omitempty" bson:"queue,omitempty"`
	Dispatch  Dispatch          `json:"dispatch,omitempty" bson:"dispatch,omitempty"`
	Needs     []StageDependency `json:"needs,omitempty" bson:"needs,omitempty"`
	NeedsMode NeedsMode         `json:"needsMode,omitempty" bson:"needsMode,omitempty"`

	// A non-nil execution, including {}, takes precedence over the entire legacy
	// summary. Missing execution means lifecycle facts may still live there.
	Execution *WorkflowRunStageExecutionDetails `json:"execution,omitempty" bson:"execution,omitempty"`
}

// WorkflowRunStageExecutionDetails contains only lifecycle facts, without
// duplicating the containing stage's identity or dependency rules. Timestamps
// are Unix milliseconds; State and DurationMs are read projections.
type WorkflowRunStageExecutionDetails struct {
	DispatchAttempts         int    `json:"dispatchAttempts,omitempty" bson:"dispatchattempts,omitempty"`
	FirstDispatchAttemptAtMs int64  `json:"firstDispatchAttemptAtMs,omitempty" bson:"firstdispatchattemptatms,omitempty"`
	LastDispatchAttemptAtMs  int64  `json:"lastDispatchAttemptAtMs,omitempty" bson:"lastdispatchattemptatms,omitempty"`
	DispatchedAtMs           int64  `json:"dispatchedAtMs,omitempty" bson:"dispatchedatms,omitempty"`
	ResolvedAtMs             int64  `json:"resolvedAtMs,omitempty" bson:"resolvedatms,omitempty"`
	LastDispatchErrorCode    string `json:"lastDispatchErrorCode,omitempty" bson:"lastdispatcherrorcode,omitempty"`

	State      WorkflowRunStageState `json:"state,omitempty" bson:"-"`
	DurationMs int64                 `json:"durationMs,omitempty" bson:"-"`
}

// NewWorkflowRunStages snapshots already-validated, compiled routing. Callers
// resolve deployment queues and normalize dispatch defaults before calling.
// Deployment metadata is excluded and nested predicates are copied. Nil means
// no captured plan; an explicit empty slice means a captured plan with no stages.
func NewWorkflowRunStages(stages []WorkflowStage) ([]WorkflowRunStage, error) {
	if stages == nil {
		return nil, nil
	}
	out := make([]WorkflowRunStage, 0, len(stages))
	seen := make(map[string]struct{}, len(stages))
	for _, stage := range stages {
		if err := checkRunStageOperation(seen, stage.Operation, "stages"); err != nil {
			return nil, err
		}
		out = append(out, WorkflowRunStage{
			Operation: stage.Operation, Name: stage.Name, Queue: stage.Queue,
			Dispatch: stage.Dispatch, Needs: stage.Needs, NeedsMode: stage.NeedsMode,
			Execution: &WorkflowRunStageExecutionDetails{},
		})
	}
	return cloneWorkflowStageSlice(out)
}

// RoutingStages projects only the captured rules for registry compilation or
// worker dispatch. Legacy execution summaries never create routing stages.
// The returned rules are detached from the run, including predicate operands.
func (r WorkflowRun) RoutingStages() ([]WorkflowStage, error) {
	if r.Stages == nil {
		return nil, nil
	}
	out := make([]WorkflowStage, 0, len(r.Stages))
	seen := make(map[string]struct{}, len(r.Stages))
	for _, stage := range r.Stages {
		if err := checkRunStageOperation(seen, stage.Operation, "stages"); err != nil {
			return nil, err
		}
		out = append(out, WorkflowStage{
			Operation: stage.Operation, Name: stage.Name, Queue: stage.Queue,
			Dispatch: stage.Dispatch, Needs: stage.Needs, NeedsMode: stage.NeedsMode,
		})
	}
	return cloneWorkflowStageSlice(out)
}

// NormalizeStages folds legacy lifecycle summaries into existing routing stages
// by operation, never replacing an existing nested execution or non-empty name.
// Unmatched summaries are retained: old global runs can have a timeline without
// a captured plan, and dependency names cannot reconstruct the missing rules.
// Validation happens before mutation. Calling this again is a no-op.
func (r *WorkflowRun) NormalizeStages() error {
	seen := make(map[string]struct{}, len(r.Stages))
	for _, stage := range r.Stages {
		if err := checkRunStageOperation(seen, stage.Operation, "stages"); err != nil {
			return err
		}
	}
	legacy := make(map[string]WorkflowRunStageExecution, len(r.StageExecutions))
	legacySeen := make(map[string]struct{}, len(r.StageExecutions))
	for _, execution := range r.StageExecutions {
		if err := checkRunStageOperation(legacySeen, execution.Operation, "stageExecutions"); err != nil {
			return err
		}
		legacy[execution.Operation] = execution
	}
	var stages []WorkflowRunStage
	if r.Stages != nil {
		stages = make([]WorkflowRunStage, len(r.Stages))
		copy(stages, r.Stages)
	}
	for i := range stages {
		stage := &stages[i]
		if execution, ok := legacy[stage.Operation]; ok {
			if stage.Name == "" {
				stage.Name = execution.Name
			}
			if stage.Execution == nil {
				details := execution.details()
				stage.Execution = &details
			}
		}
	}
	var remaining []WorkflowRunStageExecution
	if r.StageExecutions != nil {
		remaining = make([]WorkflowRunStageExecution, 0, len(r.StageExecutions))
	}
	for _, execution := range r.StageExecutions {
		if _, matched := seen[execution.Operation]; !matched {
			remaining = append(remaining, execution)
		}
	}
	r.Stages, r.StageExecutions = stages, remaining
	return nil
}

// StageExecutionSummaries is a compatibility API projection for consumers of the
// old timeline shape. It derives dependencies from captured needs and includes
// unmatched legacy history, without modifying the run or inventing rules.
func (r WorkflowRun) StageExecutionSummaries() ([]WorkflowRunStageExecution, error) {
	if err := r.NormalizeStages(); err != nil {
		return nil, err
	}
	var summaries []WorkflowRunStageExecution
	for _, stage := range r.Stages {
		details := WorkflowRunStageExecutionDetails{}
		if stage.Execution != nil {
			details = *stage.Execution
		}
		summary := WorkflowRunStageExecution{
			Operation: stage.Operation, Name: stage.Name,
			DispatchAttempts: details.DispatchAttempts, FirstDispatchAttemptAtMs: details.FirstDispatchAttemptAtMs,
			LastDispatchAttemptAtMs: details.LastDispatchAttemptAtMs, DispatchedAtMs: details.DispatchedAtMs,
			ResolvedAtMs: details.ResolvedAtMs, LastDispatchErrorCode: details.LastDispatchErrorCode,
			State: details.State, DurationMs: details.DurationMs,
		}
		seen := make(map[string]struct{}, len(stage.Needs))
		for _, need := range stage.Needs {
			if _, duplicate := seen[need.Operation]; need.Operation != "" && !duplicate {
				summary.Dependencies = append(summary.Dependencies, need.Operation)
				seen[need.Operation] = struct{}{}
			}
		}
		summaries = append(summaries, summary)
	}
	for _, execution := range r.StageExecutions {
		execution.Dependencies = append([]string(nil), execution.Dependencies...)
		summaries = append(summaries, execution)
	}
	return summaries, nil
}

func checkRunStageOperation(seen map[string]struct{}, operation, source string) error {
	if strings.TrimSpace(operation) == "" {
		return fmt.Errorf("workflow run %s contains an empty operation", source)
	}
	if _, duplicate := seen[operation]; duplicate {
		return fmt.Errorf("workflow run %s contains duplicate operation %q", source, operation)
	}
	seen[operation] = struct{}{}
	return nil
}

func cloneWorkflowStageSlice[T any](stages []T) ([]T, error) {
	data, err := bson.Marshal(bson.M{"stages": stages})
	if err != nil {
		return nil, fmt.Errorf("encode workflow run stages: %w", err)
	}
	var snapshot struct {
		Stages []T `bson:"stages"`
	}
	if err := bson.Unmarshal(data, &snapshot); err != nil {
		return nil, fmt.Errorf("decode workflow run stages: %w", err)
	}
	return snapshot.Stages, nil
}

func (e WorkflowRunStageExecution) details() WorkflowRunStageExecutionDetails {
	return WorkflowRunStageExecutionDetails{
		DispatchAttempts: e.DispatchAttempts, FirstDispatchAttemptAtMs: e.FirstDispatchAttemptAtMs,
		LastDispatchAttemptAtMs: e.LastDispatchAttemptAtMs, DispatchedAtMs: e.DispatchedAtMs,
		ResolvedAtMs: e.ResolvedAtMs, LastDispatchErrorCode: e.LastDispatchErrorCode,
		State: e.State, DurationMs: e.DurationMs,
	}
}

func (e *WorkflowRunStageExecutionDetails) populateRuntimeFields(runEnded bool, endAtMs int64) {
	e.State = e.lifecycleState(runEnded)
	e.DurationMs = 0
	stageEndAtMs := e.ResolvedAtMs
	if stageEndAtMs == 0 {
		stageEndAtMs = endAtMs
	}
	if e.DispatchedAtMs > 0 && stageEndAtMs >= e.DispatchedAtMs {
		e.DurationMs = stageEndAtMs - e.DispatchedAtMs
	}
}

func (e WorkflowRunStageExecutionDetails) lifecycleState(runEnded bool) WorkflowRunStageState {
	if e.ResolvedAtMs > 0 {
		return WorkflowRunStageStateResolved
	}
	if e.DispatchedAtMs > 0 {
		if runEnded {
			return WorkflowRunStageStateTimedOut
		}
		return WorkflowRunStageStateDispatched
	}
	if e.DispatchAttempts > 0 {
		if runEnded {
			return WorkflowRunStageStateDispatchFailed
		}
		return WorkflowRunStageStateRetrying
	}
	if runEnded {
		return WorkflowRunStageStateSkipped
	}
	return WorkflowRunStageStateWaiting
}
