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
	tests := []struct {
		name     string
		result   WorkflowResult
		expected string
	}{
		{
			name: "routing result",
			result: WorkflowResult{
				Schema: WorkflowResultSchemaV1,
				Stage:  WorkflowStageReference{Operation: "external"},
				Result: map[string]any{"label": "person"},
			},
			expected: `{"schema":"uug.ai/workflow-result/v1","stage":{"operation":"external"},"result":{"label":"person"}}`,
		},
		{
			name: "ingest payload",
			result: WorkflowResult{
				Schema:  WorkflowResultSchemaV1,
				Stage:   WorkflowStageReference{Operation: "external"},
				Payload: json.RawMessage(`{"blocks":[{"type":"marker","data":{"name":"person"}}]}`),
			},
			expected: `{"schema":"uug.ai/workflow-result/v1","stage":{"operation":"external"},"payload":{"blocks":[{"type":"marker","data":{"name":"person"}}]}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := json.Marshal(tt.result)
			if err != nil {
				t.Fatalf("json.Marshal() error = %v", err)
			}
			if string(encoded) != tt.expected {
				t.Fatalf("workflow result = %s, want %s", encoded, tt.expected)
			}
		})
	}
}

func TestWorkflowOutputDigest(t *testing.T) {
	first, err := WorkflowOutputDigest(
		map[string]any{"confidence": 0.98, "label": "person"},
		json.RawMessage(`{"blocks":[{"type":"marker","data":{"name":"person"}}]}`),
	)
	if err != nil {
		t.Fatalf("WorkflowOutputDigest() error = %v", err)
	}
	second, err := WorkflowOutputDigest(
		map[string]any{"label": "person", "confidence": 0.98},
		json.RawMessage(`{ "blocks": [ { "type": "marker", "data": { "name": "person" } } ] }`),
	)
	if err != nil {
		t.Fatalf("WorkflowOutputDigest() error = %v", err)
	}
	if first != second {
		t.Fatalf("equivalent output digests differ: %q != %q", first, second)
	}
	const expected = "e198de7c5b4578fa8daed48ad1fdeb5eb807d425799da3617938419f3770dca3"
	payloadOnly, err := WorkflowOutputDigest(
		nil,
		json.RawMessage(`{"blocks":[{"type":"marker","data":{"name":"person"}}]}`),
	)
	if err != nil {
		t.Fatalf("WorkflowOutputDigest() error = %v", err)
	}
	if payloadOnly != expected {
		t.Fatalf("payload digest = %q, want %q", payloadOnly, expected)
	}

	different, err := WorkflowOutputDigest(
		map[string]any{"label": "vehicle", "confidence": 0.98},
		json.RawMessage(`{"blocks":[{"type":"marker","data":{"name":"person"}}]}`),
	)
	if err != nil {
		t.Fatalf("WorkflowOutputDigest() error = %v", err)
	}
	if first == different {
		t.Fatal("different outputs produced the same digest")
	}
}
