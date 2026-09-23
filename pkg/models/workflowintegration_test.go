package models

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestWorkflowInvocationJSONContract(t *testing.T) {
	invocation := WorkflowInvocation{
		Schema:      WorkflowInvocationSchemaV1,
		ExecutionID: "execution-1",
		RunID:       "run-1",
		Workflow:    WorkflowInvocationWorkflow{ID: "workflow-1"},
		Stage:       WorkflowStageReference{Operation: "external"},
		Tenant: WorkflowTenantReference{
			OrganisationID: "organisation-1",
			ProjectID:      "project-1",
		},
		Media:  WorkflowMediaReference{Key: "recording.mp4"},
		Device: WorkflowDeviceReference{Key: "camera-1"},
		Callback: &WorkflowCallback{
			URL:          "https://api.example.com/workflows/runs/run-1",
			Method:       WorkflowCallbackMethodPost,
			ResultSchema: WorkflowResultSchemaV1,
		},
	}

	encoded, err := json.Marshal(invocation)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	payload := string(encoded)
	for _, expected := range []string{
		`"schema":"uug.ai/workflow-invocation/v1"`,
		`"executionId":"execution-1"`,
		`"runId":"run-1"`,
		`"callback":{"url":"https://api.example.com/workflows/runs/run-1","method":"POST","resultSchema":"uug.ai/workflow-result/v1"}`,
	} {
		if !strings.Contains(payload, expected) {
			t.Fatalf("workflow invocation = %s, missing %s", payload, expected)
		}
	}
	if strings.Contains(payload, "token") || strings.Contains(payload, "deliveryId") {
		t.Fatalf("workflow invocation contains unsupported credentials or delivery identity: %s", payload)
	}
}

func TestWorkflowResultJSONContract(t *testing.T) {
	result := WorkflowResult{
		Schema: WorkflowResultSchemaV1,
		Stage:  WorkflowStageReference{Operation: "external"},
		Result: map[string]any{"label": "person"},
	}

	encoded, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	const expected = `{"schema":"uug.ai/workflow-result/v1","stage":{"operation":"external"},"result":{"label":"person"}}`
	if string(encoded) != expected {
		t.Fatalf("workflow result = %s, want %s", encoded, expected)
	}
}
