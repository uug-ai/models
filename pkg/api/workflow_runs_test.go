package api

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/uug-ai/models/pkg/models"
)

func TestListWorkflowRunsJSONContract(t *testing.T) {
	request := ListWorkflowRunsRequest{
		Filter: WorkflowRunFilter{
			WorkflowIds: []string{"workflow-1"},
			States:      []models.WorkflowRunState{models.WorkflowRunStateRunning},
			Origins:     []models.WorkflowRunOrigin{models.WorkflowOriginAutomatic},
			From:        1700000000,
			To:          1700003600,
			Search:      "lobby",
		},
		Pagination: CursorPagination{Cursor: "cursor-1", Limit: 25},
	}
	encoded, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	for _, expected := range []string{
		`"workflowIds":["workflow-1"]`,
		`"states":["running"]`,
		`"origins":["automatic"]`,
		`"from":1700000000`,
		`"to":1700003600`,
		`"search":"lobby"`,
		`"pagination":{"cursor":"cursor-1","limit":25`,
	} {
		if !strings.Contains(string(encoded), expected) {
			t.Fatalf("workflow run request = %s, missing %s", encoded, expected)
		}
	}

	response := ListWorkflowRunsResponse{
		Runs: []WorkflowRunStatus{{
			RunId:              "run-1",
			WorkflowId:         "workflow-1",
			WorkflowName:       "People",
			State:              string(models.WorkflowRunStateRunning),
			MediaId:            "media-1",
			Key:                "recording.mp4",
			DeviceKey:          "camera-1",
			DeviceName:         "Lobby",
			RecordingTimestamp: 1699999990,
			Start:              1700000000,
			Operations: []WorkflowRunOperationStatus{{
				Operation: "forwarder",
				Status:    models.WorkflowRunOperationStateResolved,
			}},
		}},
		Summary:    WorkflowRunStatusSummary{Total: 1, Running: 1},
		Pagination: CursorPagination{NextCursor: "cursor-2", HasMore: true, PageSize: 25},
	}
	encoded, err = json.Marshal(response)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	for _, expected := range []string{
		`"mediaId":"media-1"`,
		`"deviceKey":"camera-1"`,
		`"deviceName":"Lobby"`,
		`"recordingTimestamp":1699999990`,
		`"operations":[{"operation":"forwarder","status":"resolved"}]`,
		`"summary":{"total":1,"running":1`,
		`"nextCursor":"cursor-2"`,
	} {
		if !strings.Contains(string(encoded), expected) {
			t.Fatalf("workflow run response = %s, missing %s", encoded, expected)
		}
	}
}
