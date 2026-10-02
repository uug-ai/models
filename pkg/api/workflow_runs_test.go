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
		Runs: []WorkflowRunOverview{{
			RunId:               "run-1",
			WorkflowId:          "workflow-1",
			WorkflowName:        "People",
			MatchedStartEdgeIds: []string{"entry-people", "entry-motion"},
			State:               models.WorkflowRunStateRunning,
			SourceAccess:        "available",
			SourceType:          "media",
			MediaId:             "media-1",
			Key:                 "recording.mp4",
			DeviceKey:           "camera-1",
			DeviceName:          "Lobby",
			RecordingTimestamp:  1699999990,
			Start:               1700000000,
			Operations: []models.WorkflowRunOperationStatus{{
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
		`"sourceAccess":"available"`,
		`"sourceType":"media"`,
		`"mediaId":"media-1"`,
		`"deviceKey":"camera-1"`,
		`"deviceName":"Lobby"`,
		`"recordingTimestamp":1699999990`,
		`"matchedStartEdgeIds":["entry-people","entry-motion"]`,
		`"operations":[{"operation":"forwarder","status":"resolved"}]`,
		`"summary":{"total":1,"running":1`,
		`"nextCursor":"cursor-2"`,
	} {
		if !strings.Contains(string(encoded), expected) {
			t.Fatalf("workflow run response = %s, missing %s", encoded, expected)
		}
	}

}

func TestWorkflowRunOverviewWithoutSource(t *testing.T) {
	for _, access := range []string{"restricted", "unavailable"} {
		t.Run(access, func(t *testing.T) {
			encoded, err := json.Marshal(WorkflowRunOverview{
				RunId: "run-1", WorkflowId: "workflow-1",
				State: models.WorkflowRunStateRunning, Start: 1700000000,
				SourceAccess: access, SourceType: "case",
			})
			if err != nil {
				t.Fatal(err)
			}
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(encoded, &fields); err != nil {
				t.Fatal(err)
			}
			for _, field := range []string{
				"sourceRef", "sourceLabel", "caseMediaId", "mediaId", "key",
				"deviceKey", "deviceName", "recordingTimestamp", "inputs",
				"results", "stages", "storage", "signedUrl", "triggerMatch", "conditions", "matchedStartEdgeIds",
			} {
				if _, exists := fields[field]; exists {
					t.Errorf("source-free overview exposes %s: %s", field, encoded)
				}
			}
			for _, field := range []string{"runId", "workflowId", "state", "start", "sourceAccess", "sourceType"} {
				if _, exists := fields[field]; !exists {
					t.Errorf("source-free overview omits %s: %s", field, encoded)
				}
			}
		})
	}
}
