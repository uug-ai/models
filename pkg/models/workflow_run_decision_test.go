package models

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func TestWorkflowRunStageDecisionSummaryShape(t *testing.T) {
	for _, outcome := range []WorkflowRunDecisionOutcome{
		WorkflowRunDecisionPassed, WorkflowRunDecisionFailed,
		WorkflowRunDecisionWaiting, WorkflowRunDecisionNotEvaluated,
	} {
		t.Run(string(outcome), func(t *testing.T) {
			decision := WorkflowRunStageDecision{
				EvaluatedAtMs: 1_790_680_000_123,
				Needs:         []WorkflowRunStageNeedDecision{{Index: 0, Outcome: outcome}},
			}
			wire, err := json.Marshal(decision)
			if err != nil {
				t.Fatal(err)
			}
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(wire, &fields); err != nil {
				t.Fatal(err)
			}
			if len(fields) != 3 || string(fields["eligible"]) != "false" ||
				string(fields["evaluatedAtMs"]) != "1790680000123" {
				t.Fatalf("unexpected summary shape: %s", wire)
			}
			var needs []map[string]json.RawMessage
			if err := json.Unmarshal(fields["needs"], &needs); err != nil {
				t.Fatal(err)
			}
			if len(needs) != 1 || len(needs[0]) != 2 || string(needs[0]["index"]) != "0" ||
				string(needs[0]["outcome"]) != `"`+string(outcome)+`"` {
				t.Fatalf("edge summary must contain only index and outcome: %s", wire)
			}
			var decoded WorkflowRunStageDecision
			if err := json.Unmarshal(wire, &decoded); err != nil || !reflect.DeepEqual(decoded, decision) {
				t.Fatalf("JSON round trip = %+v, error = %v", decoded, err)
			}
			data, err := bson.Marshal(decision)
			if err != nil {
				t.Fatal(err)
			}
			raw := bson.Raw(data)
			if raw.Lookup("evaluatedatms").Int64() != decision.EvaluatedAtMs ||
				raw.Lookup("evaluatedAtMs").Type != 0 || raw.Lookup("eligible").Boolean() {
				t.Fatalf("BSON spelling or false eligibility lost: %s", raw)
			}
			if err := bson.Unmarshal(data, &decoded); err != nil || !reflect.DeepEqual(decoded, decision) {
				t.Fatalf("BSON round trip = %+v, error = %v", decoded, err)
			}
		})
	}
}

func TestWorkflowRunStageDecisionDoesNotControlLifecycleOrRouting(t *testing.T) {
	for _, eligible := range []bool{false, true} {
		decision := &WorkflowRunStageDecision{EvaluatedAtMs: 123, Eligible: eligible}
		run := WorkflowRun{
			Stages: []WorkflowRunStage{{
				Operation: "report", Dispatch: DispatchConditional,
				Needs:     []StageDependency{{Operation: "pose"}},
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
		if err := decoded.NormalizeStages(); err != nil {
			t.Fatal(err)
		}
		decoded.PopulateRuntimeFields(time.Now())
		execution := decoded.Stages[0].Execution
		if execution.State != WorkflowRunStageStateWaiting || execution.DispatchAttempts != 0 ||
			execution.ResolvedAtMs != 0 || !reflect.DeepEqual(execution.Decision, decision) {
			t.Fatalf("display summary or stale legacy fields changed execution: %+v", execution)
		}
		wire, err := json.Marshal(decoded)
		if err != nil {
			t.Fatal(err)
		}
		var fromJSON WorkflowRun
		if err := json.Unmarshal(wire, &fromJSON); err != nil ||
			!reflect.DeepEqual(fromJSON.Stages[0].Execution.Decision, decision) {
			t.Fatalf("nested summary JSON round trip failed: %v", err)
		}
		routing, err := decoded.RoutingStages()
		if err != nil {
			t.Fatal(err)
		}
		wire, err = json.Marshal(routing)
		if err != nil || bytes.Contains(wire, []byte(`"decision"`)) {
			t.Fatalf("display summary leaked into routing: %s, %v", wire, err)
		}
		newStages, err := NewWorkflowRunStages(routing)
		if err != nil || newStages[0].Execution.Decision != nil {
			t.Fatalf("new plan inherited a historical summary: %+v, %v", newStages, err)
		}
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
			t.Fatalf("reader reconstructed an uncaptured summary: %s, %v", wire, err)
		}
	}

	run := WorkflowRun{
		Stages:          []WorkflowRunStage{{Operation: "report"}},
		StageExecutions: []WorkflowRunStageExecution{{Operation: "report", DispatchedAtMs: 123}},
	}
	if err := run.NormalizeStages(); err != nil {
		t.Fatal(err)
	}
	run.PopulateRuntimeFields(time.Now())
	if execution := run.Stages[0].Execution; execution.Decision != nil || execution.DispatchedAtMs != 123 {
		t.Fatalf("legacy history must not imply a captured summary: %+v", execution)
	}
}
