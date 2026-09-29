package models

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func TestWorkflowRunStageDecisionEvaluation(t *testing.T) {
	passed, failed := WorkflowRunDecisionPassed, WorkflowRunDecisionFailed
	waiting, skipped := WorkflowRunDecisionWaiting, WorkflowRunDecisionNotEvaluated
	pass := WorkflowCondition{Path: "results.pose.count", Op: ConditionOpGte, Value: 2}
	fail := WorkflowCondition{Path: "results.pose.count", Op: ConditionOpGt, Value: 99}
	root := map[string]any{"results": map[string]any{"pose": map[string]any{"count": 3}}}
	available := map[string]bool{"pose": true}
	at := time.UnixMilli(1_790_680_000_123)
	tests := []struct {
		name       string
		mode       NeedsMode
		needs      []StageDependency
		eligible   bool
		outcomes   []WorkflowRunDecisionOutcome
		ready      []bool
		conditions [][]WorkflowRunDecisionOutcome
	}{
		{
			name: "default any preserves gates and stops on first passing need",
			needs: []StageDependency{
				{Operation: "pending", Condition: &pass},
				{Operation: "pose", Condition: &fail},
				{Operation: "pose", Condition: &pass},
				{Operation: "pending", Condition: &fail},
			},
			eligible: true, outcomes: []WorkflowRunDecisionOutcome{waiting, failed, passed, skipped},
			ready: []bool{false, true, true, false},
			conditions: [][]WorkflowRunDecisionOutcome{
				{skipped}, {failed}, {passed}, {skipped},
			},
		},
		{
			name: "all stops at unavailable gate", mode: NeedsModeAll,
			needs: []StageDependency{
				{Operation: "pending", Conditions: []WorkflowCondition{pass}},
				{Operation: "pose", Conditions: []WorkflowCondition{pass}},
			},
			outcomes: []WorkflowRunDecisionOutcome{waiting, skipped}, ready: []bool{false, true},
			conditions: [][]WorkflowRunDecisionOutcome{{skipped}, {skipped}},
		},
		{
			name: "all stops at failed need", mode: NeedsModeAll,
			needs: []StageDependency{
				{Operation: "pose", Condition: &fail},
				{Operation: "pose", Condition: &pass},
			},
			outcomes: []WorkflowRunDecisionOutcome{failed, skipped}, ready: []bool{true, true},
			conditions: [][]WorkflowRunDecisionOutcome{{failed}, {skipped}},
		},
		{
			name: "all passes", mode: NeedsModeAll,
			needs:    []StageDependency{{Operation: "pose", Condition: &pass}, {Condition: &pass}},
			eligible: true, outcomes: []WorkflowRunDecisionOutcome{passed, passed}, ready: []bool{true, true},
			conditions: [][]WorkflowRunDecisionOutcome{{passed}, {passed}},
		},
		{
			name:     "condition all short circuit",
			needs:    []StageDependency{{Operation: "pose", Conditions: []WorkflowCondition{pass, fail, pass}}},
			outcomes: []WorkflowRunDecisionOutcome{failed}, ready: []bool{true},
			conditions: [][]WorkflowRunDecisionOutcome{{passed, failed, skipped}},
		},
		{
			name: "condition any short circuit",
			needs: []StageDependency{{
				Operation: "pose", ConditionMode: ConditionModeAny, Conditions: []WorkflowCondition{fail, pass, fail},
			}},
			eligible: true, outcomes: []WorkflowRunDecisionOutcome{passed}, ready: []bool{true},
			conditions: [][]WorkflowRunDecisionOutcome{{failed, passed, skipped}},
		},
		{
			name:     "legacy singular retains index zero",
			needs:    []StageDependency{{Condition: &pass, ConditionMode: ConditionModeAny}},
			eligible: true, outcomes: []WorkflowRunDecisionOutcome{passed}, ready: []bool{true},
			conditions: [][]WorkflowRunDecisionOutcome{{passed}},
		},
		{
			name:     "empty conditions impose no restriction even in any mode",
			needs:    []StageDependency{{ConditionMode: ConditionModeAny}},
			eligible: true, outcomes: []WorkflowRunDecisionOutcome{passed}, ready: []bool{true},
			conditions: [][]WorkflowRunDecisionOutcome{nil},
		},
		{
			name: "all needs miss in any mode", mode: NeedsModeAny,
			needs:    []StageDependency{{Operation: "pose", Condition: &fail}, {Operation: "pending"}},
			outcomes: []WorkflowRunDecisionOutcome{failed, waiting}, ready: []bool{true, false},
			conditions: [][]WorkflowRunDecisionOutcome{{failed}, nil},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stage := WorkflowRunStage{Operation: "report", Dispatch: DispatchConditional, NeedsMode: tt.mode, Needs: tt.needs}
			decision, err := stage.EvaluateDecision(root, available, at)
			if err != nil {
				t.Fatal(err)
			}
			if decision.Eligible != tt.eligible || decision.EvaluatedAtMs != at.UnixMilli() || len(decision.Needs) != len(tt.needs) {
				t.Fatalf("decision = %+v, want eligible %v and %d needs", decision, tt.eligible, len(tt.needs))
			}
			legacyEligible := tt.mode == NeedsModeAll
			for _, need := range stage.Needs {
				matches := need.Matches(root, available)
				if tt.mode == NeedsModeAll && !matches {
					legacyEligible = false
					break
				}
				if tt.mode != NeedsModeAll && matches {
					legacyEligible = true
					break
				}
			}
			if decision.Eligible != legacyEligible {
				t.Fatalf("decision diverges from existing dependency matching: %+v", decision)
			}
			for i, need := range decision.Needs {
				if need.Index != i || need.Ready != tt.ready[i] || need.Outcome != tt.outcomes[i] {
					t.Fatalf("need %d = %+v", i, need)
				}
				if len(need.Conditions) != len(tt.conditions[i]) {
					t.Fatalf("need %d condition checks = %+v", i, need.Conditions)
				}
				for j, condition := range need.Conditions {
					if condition.Index != j || condition.Outcome != tt.conditions[i][j] {
						t.Fatalf("need %d condition %d = %+v", i, j, condition)
					}
				}
			}
		})
	}
}

func TestWorkflowRunStageDecisionDefaultsAndValidation(t *testing.T) {
	for _, dispatch := range []Dispatch{"", DispatchAlways, DispatchConditional} {
		stage := WorkflowRunStage{Operation: "report", Dispatch: dispatch}
		decision, err := stage.EvaluateDecision(nil, nil, time.Now())
		if err != nil || !decision.Eligible || len(decision.Needs) != 0 {
			t.Fatalf("unconditional/default routing %q = %+v, %v", dispatch, decision, err)
		}
	}
	pass := WorkflowCondition{Path: "device.deviceKey", Op: ConditionOpExists}
	invalid := []WorkflowRunStage{
		{Operation: " "},
		{Operation: "report", Dispatch: "unknown"},
		{Operation: "report", Dispatch: DispatchConditional, NeedsMode: "unknown", Needs: []StageDependency{{}}},
		{Operation: "report", Dispatch: DispatchConditional, Needs: []StageDependency{{ConditionMode: "unknown"}}},
		{Operation: "report", Dispatch: DispatchConditional, Needs: []StageDependency{{Condition: &pass, Conditions: []WorkflowCondition{}}}},
		{Operation: "report", Dispatch: DispatchConditional, Needs: []StageDependency{{Conditions: []WorkflowCondition{{
			Path: "user.storage.password", Op: ConditionOpExists,
		}}}}},
		{Operation: "report", Dispatch: DispatchConditional, Needs: []StageDependency{{}, {Operation: "pending", Conditions: []WorkflowCondition{{
			Path: "device.deviceKey", Op: "unknown",
		}}}}},
		{Operation: "report", Dispatch: DispatchConditional, Needs: []StageDependency{{
			ConditionMode: ConditionModeAny, Conditions: []WorkflowCondition{pass, {
				Path: "device.deviceName", Op: ConditionOpMatches, Value: "[",
			}},
		}}},
	}
	root := AutomaticTriggerRoot(WorkflowDevice{DeviceKey: "cam-1"}, WorkflowUser{})
	for i, stage := range invalid {
		decision, err := stage.EvaluateDecision(root, nil, time.Now())
		if err == nil || decision != nil {
			t.Fatalf("invalid case %d produced a success-shaped decision: %+v, %v", i, decision, err)
		}
	}
}

func TestWorkflowRunStageDecisionAnyMatchStaysSameElement(t *testing.T) {
	stage := WorkflowRunStage{
		Operation: "report", Dispatch: DispatchConditional,
		Needs: []StageDependency{{Conditions: []WorkflowCondition{{
			Path: "results.pose.objects", Op: ConditionOpAnyMatch,
			Match: &WorkflowPredicateSet{Conditions: []WorkflowPredicate{
				{Path: "label", Op: ConditionOpEq, Value: "person"},
				{Path: "confidence", Op: ConditionOpGte, Value: 0.8},
			}},
		}}}},
	}
	objects := []any{
		map[string]any{"label": "person", "confidence": 0.2},
		map[string]any{"label": "car", "confidence": 0.9},
	}
	for _, eligible := range []bool{false, true} {
		if eligible {
			objects = append(objects, map[string]any{"label": "person", "confidence": 0.9})
		}
		root := map[string]any{"results": map[string]any{"pose": map[string]any{"objects": objects}}}
		decision, err := stage.EvaluateDecision(root, nil, time.Now())
		if err != nil {
			t.Fatal(err)
		}
		if decision.Eligible != eligible || len(decision.Needs[0].Conditions) != 1 ||
			decision.Needs[0].Conditions[0].Outcome != workflowDecisionOutcome(eligible) {
			t.Fatalf("anyMatch must stay one same-element check: %+v", decision)
		}
	}
}

func TestWorkflowRunStageDecisionDoesNotMutateOrCaptureInputs(t *testing.T) {
	previous := &WorkflowRunStageDecision{EvaluatedAtMs: 123}
	stage := WorkflowRunStage{
		Operation: "report", Dispatch: DispatchConditional,
		Needs: []StageDependency{{Operation: "pose", Conditions: []WorkflowCondition{{
			Path: "results.pose.label", Op: ConditionOpEq, Value: "person",
		}}}},
		Execution: &WorkflowRunStageExecutionDetails{DispatchedAtMs: 234, Decision: previous},
	}
	root := map[string]any{"results": map[string]any{"pose": map[string]any{"label": "person", "private": "do-not-store"}}}
	available := map[string]bool{"pose": true}
	before, err := json.Marshal(stage)
	if err != nil {
		t.Fatal(err)
	}
	decision, err := stage.EvaluateDecision(root, available, time.Now())
	if err != nil || !decision.Eligible {
		t.Fatalf("decision = %+v, %v", decision, err)
	}
	after, err := json.Marshal(stage)
	if err != nil || !bytes.Equal(before, after) || stage.Execution.Decision != previous {
		t.Fatal("evaluation overwrote the run's existing facts or rules")
	}
	encoded, err := json.Marshal(decision)
	if err != nil {
		t.Fatal(err)
	}
	for _, excluded := range []string{"results", "pose", "person", "private", "do-not-store"} {
		if bytes.Contains(encoded, []byte(excluded)) {
			t.Fatalf("decision duplicated rule/input data: %s", encoded)
		}
	}
	available["pose"] = false
	stage.Needs[0].Conditions[0].Value = "changed"
	root["results"] = nil
	if !decision.Eligible || !decision.Needs[0].Ready || decision.Needs[0].Conditions[0].Outcome != WorkflowRunDecisionPassed {
		t.Fatal("decision shares mutable evaluation context")
	}
}

func TestWorkflowRunStageDecisionSerializationAndLegacyCompatibility(t *testing.T) {
	decision := &WorkflowRunStageDecision{
		EvaluatedAtMs: 1_790_680_000_123,
		Needs: []WorkflowRunStageNeedDecision{{
			Index: 0, Ready: false, Outcome: WorkflowRunDecisionWaiting,
			Conditions: []WorkflowRunConditionDecision{{Index: 0, Outcome: WorkflowRunDecisionNotEvaluated}},
		}},
	}
	run := WorkflowRun{
		Stages: []WorkflowRunStage{{
			Operation: "report", Dispatch: DispatchConditional,
			Needs:     []StageDependency{{Operation: "pose", Condition: &WorkflowCondition{Path: "results.pose.label", Op: ConditionOpExists}}},
			Execution: &WorkflowRunStageExecutionDetails{Decision: decision},
		}},
		StageExecutions: []WorkflowRunStageExecution{{Operation: "report", DispatchAttempts: 99, ResolvedAtMs: 123}},
	}
	data, err := bson.Marshal(run)
	if err != nil {
		t.Fatal(err)
	}
	var decoded WorkflowRun
	if err := bson.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decoded.Stages[0].Execution.Decision, decision) {
		t.Fatal("BSON lost decision fields")
	}
	raw := bson.Raw(data).Lookup("stages").Array().Index(0).Value().Document().Lookup("execution").Document().Lookup("decision").Document()
	if raw.Lookup("evaluatedatms").Int64() != decision.EvaluatedAtMs || raw.Lookup("evaluatedAtMs").Type != 0 ||
		raw.Lookup("eligible").Type != bson.TypeBoolean || raw.Lookup("eligible").Boolean() {
		t.Fatalf("decision BSON spelling or false eligibility lost: %s", raw)
	}
	need := raw.Lookup("needs").Array().Index(0).Value().Document()
	if need.Lookup("index").Int32() != 0 || need.Lookup("ready").Type != bson.TypeBoolean || need.Lookup("ready").Boolean() {
		t.Fatalf("zero index or false readiness lost: %s", need)
	}
	if err := decoded.NormalizeStages(); err != nil {
		t.Fatal(err)
	}
	decoded.PopulateRuntimeFields(time.Now())
	execution := decoded.Stages[0].Execution
	if execution.DispatchAttempts != 0 || execution.ResolvedAtMs != 0 || !reflect.DeepEqual(execution.Decision, decision) {
		t.Fatalf("stale legacy lifecycle overwrote canonical decision authority: %+v", execution)
	}
	wire, err := json.Marshal(decoded)
	if err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{`"decision":`, `"evaluatedAtMs":1790680000123`, `"eligible":false`, `"index":0`, `"ready":false`} {
		if !bytes.Contains(wire, []byte(required)) {
			t.Fatalf("JSON missing %s: %s", required, wire)
		}
	}
	var fromJSON WorkflowRun
	if err := json.Unmarshal(wire, &fromJSON); err != nil || !reflect.DeepEqual(fromJSON.Stages[0].Execution.Decision, decision) {
		t.Fatalf("JSON round trip lost decision: %v", err)
	}
	routing, err := decoded.RoutingStages()
	if err != nil {
		t.Fatal(err)
	}
	routingJSON, err := json.Marshal(routing)
	if err != nil || bytes.Contains(routingJSON, []byte(`"decision"`)) {
		t.Fatalf("decision leaked into routing-only projection: %s, %v", routingJSON, err)
	}
	newStages, err := NewWorkflowRunStages(routing)
	if err != nil || newStages[0].Execution.Decision != nil {
		t.Fatalf("new plan inherited historical decision: %+v, %v", newStages, err)
	}
}

func TestWorkflowRunStageDecisionAbsentRemainsUnknown(t *testing.T) {
	for _, execution := range []bson.M{{}, {"decision": nil}, {"dispatchedatms": int64(123)}} {
		data, err := bson.Marshal(bson.M{
			"stages":  bson.A{bson.M{"operation": "report", "execution": execution}},
			"results": bson.M{"pose": bson.M{"label": "person"}},
		})
		if err != nil {
			t.Fatal(err)
		}
		var run WorkflowRun
		if err := bson.Unmarshal(data, &run); err != nil {
			t.Fatal(err)
		}
		if err := run.NormalizeStages(); err != nil {
			t.Fatal(err)
		}
		run.PopulateRuntimeFields(time.Now())
		wire, err := json.Marshal(run)
		if err != nil || run.Stages[0].Execution.Decision != nil || bytes.Contains(wire, []byte(`"decision"`)) {
			t.Fatalf("reader reconstructed uncaptured decision: %s, %v", wire, err)
		}
	}
}

func TestWorkflowRunStageDecisionLegacySummaryDoesNotInventDecision(t *testing.T) {
	run := WorkflowRun{
		Stages:          []WorkflowRunStage{{Operation: "report"}},
		StageExecutions: []WorkflowRunStageExecution{{Operation: "report", DispatchedAtMs: 123}},
	}
	if err := run.NormalizeStages(); err != nil {
		t.Fatal(err)
	}
	run.PopulateRuntimeFields(time.Now())
	execution := run.Stages[0].Execution
	if execution == nil || execution.Decision != nil || execution.DispatchedAtMs != 123 {
		t.Fatalf("legacy dispatch timestamp must not imply a captured decision: %+v", execution)
	}
}

func TestEvaluateConditionSetObserverPreservesShortCircuiting(t *testing.T) {
	pass := WorkflowCondition{Path: "device.deviceKey", Op: ConditionOpExists}
	fail := WorkflowCondition{Path: "device.deviceName", Op: ConditionOpEq, Value: "missing"}
	root := AutomaticTriggerRoot(WorkflowDevice{DeviceKey: "cam-1"}, WorkflowUser{})
	for _, mode := range []ConditionMode{ConditionModeAll, ConditionModeAny} {
		set := WorkflowConditionSet{ConditionMode: mode, Conditions: []WorkflowCondition{pass, fail, pass}}
		var reached []int
		result := evaluateConditionSet(set, root, func(index int, matched bool) {
			reached = append(reached, index)
			if matched != EvaluateCondition(&set.Conditions[index], root) {
				t.Fatal("observer did not receive the actual evaluated result")
			}
		})
		want := []int{0, 1}
		if mode == ConditionModeAny {
			want = []int{0}
		}
		if result != EvaluateConditionSet(set, root) || !reflect.DeepEqual(reached, want) {
			t.Fatalf("observer changed %s evaluation: result %v, reached %v", mode, result, reached)
		}
	}
	set := WorkflowConditionSet{ConditionMode: ConditionModeAny, Conditions: []WorkflowCondition{pass, {Op: "invalid"}}}
	called := false
	result := evaluateConditionSet(set, root, func(int, bool) { called = true })
	if result || called {
		t.Fatal("invalid groups must fail closed before recording any successful check")
	}
}
