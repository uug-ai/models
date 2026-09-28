package api

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/uug-ai/models/pkg/models"
)

func TestWorkflowRunStatus_DetailFields(t *testing.T) {
	t.Run("overview status omits detail fields", func(t *testing.T) {
		encoded, err := json.Marshal(WorkflowRunStatus{
			RunId: "run-1",
			State: string(models.WorkflowRunStateRunning),
		})
		if err != nil {
			t.Fatalf("marshal status: %v", err)
		}

		for _, field := range []string{"startedAtMs", "endedAtMs", "durationMs", "traceId", "stages"} {
			if strings.Contains(string(encoded), `"`+field+`"`) {
				t.Errorf("empty detail field %q appeared in overview status: %s", field, encoded)
			}
		}
	})

	t.Run("detail status includes execution timeline", func(t *testing.T) {
		encoded, err := json.Marshal(WorkflowRunStatus{
			RunId:       "run-1",
			State:       string(models.WorkflowRunStateCompleted),
			StartedAtMs: 1_000,
			EndedAtMs:   2_000,
			DurationMs:  1_000,
			TraceId:     "trace-1",
			Stages: []models.WorkflowRunStageExecution{{
				Operation:      "anpr",
				State:          models.WorkflowRunStageStateResolved,
				DispatchedAtMs: 1_100,
				ResolvedAtMs:   1_900,
				DurationMs:     800,
			}},
		})
		if err != nil {
			t.Fatalf("marshal status: %v", err)
		}

		var decoded map[string]any
		if err := json.Unmarshal(encoded, &decoded); err != nil {
			t.Fatalf("unmarshal status: %v", err)
		}
		if decoded["traceId"] != "trace-1" {
			t.Errorf("traceId = %v, want trace-1", decoded["traceId"])
		}
		stages, ok := decoded["stages"].([]any)
		if !ok || len(stages) != 1 {
			t.Fatalf("stages = %#v, want one stage", decoded["stages"])
		}
	})
}
