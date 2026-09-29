package models

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestWorkflowRun_MarshalJSON_ProjectsRunIdFromId asserts the single-source-of-truth
// invariant: RunId is the wire projection of the persisted Id, derived automatically
// at marshal time so a producer only ever sets Id and the two representations can
// never drift.
func TestWorkflowRun_MarshalJSON_ProjectsRunIdFromId(t *testing.T) {
	id := primitive.NewObjectID()

	t.Run("set Id emits runId as the hex", func(t *testing.T) {
		b, err := json.Marshal(WorkflowRun{Operation: "anpr", Key: "media-1", Id: id})
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		var wire map[string]any
		if err := json.Unmarshal(b, &wire); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if got := wire["runId"]; got != id.Hex() {
			t.Errorf("runId = %v, want %s", got, id.Hex())
		}
		// The persisted identity itself never appears on the wire.
		if _, ok := wire["_id"]; ok {
			t.Errorf("wire carried _id; Id must stay persistence-only")
		}
		if _, ok := wire["id"]; ok {
			t.Errorf("wire carried id; Id must stay persistence-only")
		}
	})

	t.Run("zero Id omits runId", func(t *testing.T) {
		b, err := json.Marshal(WorkflowRun{Operation: "event", Key: "media-1"})
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		if strings.Contains(string(b), "runId") {
			t.Errorf("runId present on a run with no Id yet: %s", b)
		}
	})

	t.Run("pointer marshal path also projects", func(t *testing.T) {
		b, err := json.Marshal(&WorkflowRun{Id: id})
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		if !strings.Contains(string(b), id.Hex()) {
			t.Errorf("pointer marshal did not project runId: %s", b)
		}
	})

	t.Run("wire round-trip keeps RunId and leaves Id zero", func(t *testing.T) {
		var r WorkflowRun
		if err := json.Unmarshal([]byte(`{"operation":"anpr","runId":"`+id.Hex()+`"}`), &r); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if r.RunId != id.Hex() {
			t.Errorf("RunId = %q, want %s", r.RunId, id.Hex())
		}
		if !r.Id.IsZero() {
			t.Errorf("Id = %s, want zero (Id is json:\"-\", never read off the wire)", r.Id.Hex())
		}
	})

	t.Run("project ownership travels only in the workflow user", func(t *testing.T) {
		projectId := primitive.NewObjectID()
		b, err := json.Marshal(WorkflowRun{
			ProjectId: &projectId,
			User:      WorkflowUser{OrganisationId: "org-1", ProjectId: &projectId},
		})
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}

		var wire map[string]any
		if err := json.Unmarshal(b, &wire); err != nil {
			t.Fatalf("unmarshal wire map: %v", err)
		}
		if _, ok := wire["projectId"]; ok {
			t.Fatal("persisted projectId must not appear at the run wire root")
		}

		var decoded WorkflowRun
		if err := json.Unmarshal(b, &decoded); err != nil {
			t.Fatalf("unmarshal run: %v", err)
		}
		if decoded.User.ProjectId == nil || *decoded.User.ProjectId != projectId {
			t.Fatalf("wire user projectId = %v, want %s", decoded.User.ProjectId, projectId.Hex())
		}
		if decoded.ProjectId != nil {
			t.Fatal("wire projectId must not populate persistence ownership")
		}
	})

	t.Run("sparse dispatch omits API lifecycle fields", func(t *testing.T) {
		b, err := json.Marshal(WorkflowRun{
			Operation: "anpr",
			RunId:     "run-1",
			Key:       "media-1",
			TraceId:   "trace-1",
		})
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}

		for _, field := range []string{
			"start", "end", "startedAtMs", "endedAtMs", "durationMs",
			"stageExecutions", "state", "mediaId", "deviceKey", "deviceName",
			"dispatched", "resolved", "dispatchedOperations",
			"resolvedOperations", "operations", "hasResults",
		} {
			if strings.Contains(string(b), field) {
				t.Errorf("empty API lifecycle field %q appeared on queue wire: %s", field, b)
			}
		}
	})
}

func TestWorkflowRun_StageExecutionPersistence(t *testing.T) {
	run := WorkflowRun{
		StartedAtMs: 1_000,
		EndedAtMs:   2_000,
		StageExecutions: []WorkflowRunStageExecution{{
			Operation:                "anpr",
			Name:                     "Number plate recognition",
			Dependencies:             []string{"objecttracking"},
			DispatchAttempts:         2,
			FirstDispatchAttemptAtMs: 1_100,
			LastDispatchAttemptAtMs:  1_200,
			DispatchedAtMs:           1_250,
			ResolvedAtMs:             1_900,
			LastDispatchErrorCode:    "queue_unavailable",
			State:                    WorkflowRunStageStateResolved,
			DurationMs:               650,
		}},
	}

	encoded, err := bson.Marshal(run)
	if err != nil {
		t.Fatalf("marshal BSON: %v", err)
	}

	var document bson.M
	if err := bson.Unmarshal(encoded, &document); err != nil {
		t.Fatalf("unmarshal BSON: %v", err)
	}
	if document["startedatms"] != int64(1_000) {
		t.Errorf("startedatms = %v, want 1000", document["startedatms"])
	}
	if document["endedatms"] != int64(2_000) {
		t.Errorf("endedatms = %v, want 2000", document["endedatms"])
	}

	executions, ok := document["stageexecutions"].(bson.A)
	if !ok || len(executions) != 1 {
		t.Fatalf("stageexecutions = %#v, want one execution", document["stageexecutions"])
	}
	execution, ok := executions[0].(bson.M)
	if !ok {
		t.Fatalf("stage execution = %#v, want bson.M", executions[0])
	}
	if execution["operation"] != "anpr" {
		t.Errorf("operation = %v, want anpr", execution["operation"])
	}
	if _, ok := execution["state"]; ok {
		t.Error("derived state must not be persisted")
	}
	if _, ok := execution["durationMs"]; ok {
		t.Error("derived duration must not be persisted")
	}
}

func TestWorkflowRun_PopulateRuntimeFields(t *testing.T) {
	run := WorkflowRun{
		Start:                1,
		End:                  3,
		DispatchedOperations: []string{"anpr", "redaction"},
		ResolvedOperations:   []string{"anpr"},
		Results:              map[string]interface{}{"anpr": map[string]interface{}{"plate": "ABC"}},
		StageExecutions: []WorkflowRunStageExecution{
			{Operation: "anpr", DispatchedAtMs: 1_100, ResolvedAtMs: 1_900},
			{Operation: "redaction", DispatchAttempts: 2, DispatchedAtMs: 2_000},
			{Operation: "notification", DispatchAttempts: 1},
			{Operation: "export"},
		},
	}

	run.PopulateRuntimeFields(time.UnixMilli(4_000))

	if run.State != WorkflowRunStateCompleted {
		t.Errorf("State = %q, want completed", run.State)
	}
	if run.StartedAtMs != 1_000 || run.EndedAtMs != 3_000 || run.DurationMs != 2_000 {
		t.Errorf("timing = (%d, %d, %d), want (1000, 3000, 2000)", run.StartedAtMs, run.EndedAtMs, run.DurationMs)
	}
	if run.Dispatched != 2 || run.Resolved != 1 || !run.HasResults {
		t.Errorf("progress = (%d, %d, %t), want (2, 1, true)", run.Dispatched, run.Resolved, run.HasResults)
	}
	if len(run.Operations) != 2 ||
		run.Operations[0].Status != WorkflowRunOperationStateResolved ||
		run.Operations[1].Status != WorkflowRunOperationStateDispatched {
		t.Errorf("Operations = %#v, want resolved then dispatched", run.Operations)
	}

	wantStates := []WorkflowRunStageState{
		WorkflowRunStageStateResolved,
		WorkflowRunStageStateTimedOut,
		WorkflowRunStageStateDispatchFailed,
		WorkflowRunStageStateSkipped,
	}
	for i, want := range wantStates {
		if got := run.StageExecutions[i].State; got != want {
			t.Errorf("StageExecutions[%d].State = %q, want %q", i, got, want)
		}
	}
	if run.StageExecutions[0].DurationMs != 800 {
		t.Errorf("resolved stage duration = %d, want 800", run.StageExecutions[0].DurationMs)
	}
	if run.StageExecutions[1].DurationMs != 1_000 {
		t.Errorf("timed-out stage duration = %d, want 1000", run.StageExecutions[1].DurationMs)
	}
}

// TestAutomaticRunObjectID asserts the deterministic automatic run identity:
// stable for a given (key, org, workflow) triple, distinct across triples, never
// zero, and — via MarshalJSON — projected onto the wire RunId so producer and
// engine agree on the same runId.
func TestAutomaticRunObjectID(t *testing.T) {
	a := AutomaticRunObjectID("media-1", "org-1", "wf-1")

	t.Run("deterministic for the same triple", func(t *testing.T) {
		if b := AutomaticRunObjectID("media-1", "org-1", "wf-1"); a != b {
			t.Errorf("same triple produced different ids: %s vs %s", a.Hex(), b.Hex())
		}
	})

	t.Run("not the zero id", func(t *testing.T) {
		if a.IsZero() {
			t.Error("derived id should never be zero")
		}
	})

	t.Run("distinct per field, with unambiguous boundaries", func(t *testing.T) {
		cases := map[string]primitive.ObjectID{
			"other key":      AutomaticRunObjectID("media-2", "org-1", "wf-1"),
			"other org":      AutomaticRunObjectID("media-1", "org-2", "wf-1"),
			"other workflow": AutomaticRunObjectID("media-1", "org-1", "wf-2"),
			// Without a separator these two would hash the same bytes.
			"boundary shift": AutomaticRunObjectID("media", "1org", "1wf-1"),
		}
		for name, id := range cases {
			if id == a {
				t.Errorf("%s should differ from the base id", name)
			}
		}
	})

	t.Run("projects onto the wire runId via MarshalJSON", func(t *testing.T) {
		b, err := json.Marshal(WorkflowRun{Operation: "event", Key: "media-1", Id: a})
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		if want := `"runId":"` + a.Hex() + `"`; !strings.Contains(string(b), want) {
			t.Errorf("expected %s in %s", want, b)
		}
	})
}
