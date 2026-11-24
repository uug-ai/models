package models

import (
	"encoding/json"
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
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

func TestNewWorkflowInvocationProjectsOnlySafeRunContext(t *testing.T) {
	projectID, err := primitive.ObjectIDFromHex("507f1f77bcf86cd799439011")
	if err != nil {
		t.Fatal(err)
	}
	run := &WorkflowRun{
		RunId:        "run-1",
		WorkflowId:   "workflow-1",
		WorkflowName: "Gate after hours",
		Operation:    "forwarder",
		Key:          "recording.mp4",
		SignedURL:    "https://vault.example.com/signed",
		TraceId:      "trace-1",
		Storage:      &WorkflowStorage{AccessKey: "internal-access", Secret: "internal-secret"},
		User: WorkflowUser{
			OrganisationId: "org-1",
			ProjectId:      &projectID,
			Storage:        Storage{AccessKey: "account-access", Secret: "account-secret"},
		},
		Device:  WorkflowDevice{DeviceKey: "camera-1", DeviceName: "Gate", SiteIds: []string{"site-1"}},
		Inputs:  map[string]interface{}{"classify": map[string]interface{}{"label": "person"}},
		Results: map[string]interface{}{"anpr": map[string]interface{}{"plate": "1-ABC-123"}},
	}

	invocation, err := NewWorkflowInvocation(run)
	if err != nil {
		t.Fatalf("NewWorkflowInvocation() error = %v", err)
	}
	// Golden value: the forwarder's execution ID for this run before the
	// constructor moved here. Changing it breaks consumer deduplication.
	if want := "wfx_059989ff8e10ce09b8e803b2b77e497535d3b8f22f8c1e86125059e5174dbb21"; invocation.ExecutionID != want {
		t.Fatalf("ExecutionID = %q, want %q", invocation.ExecutionID, want)
	}
	if invocation.Schema != WorkflowInvocationSchemaV1 || invocation.RunID != "run-1" ||
		invocation.Workflow != (WorkflowInvocationWorkflow{ID: "workflow-1", Name: "Gate after hours"}) ||
		invocation.Stage.Operation != "forwarder" || invocation.TraceID != "trace-1" {
		t.Fatalf("identity = %#v", invocation)
	}
	if invocation.Tenant != (WorkflowTenantReference{OrganisationID: "org-1", ProjectID: projectID.Hex()}) {
		t.Fatalf("Tenant = %#v", invocation.Tenant)
	}
	if invocation.Media != (WorkflowMediaReference{Key: "recording.mp4", SignedURL: run.SignedURL}) {
		t.Fatalf("Media = %#v", invocation.Media)
	}
	if invocation.Data != nil || invocation.Callback != nil {
		t.Fatalf("Data/Callback must be left to the caller: %#v %#v", invocation.Data, invocation.Callback)
	}

	encoded, err := json.Marshal(invocation)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	for _, leaked := range []string{"internal-access", "internal-secret", "account-access", "account-secret", "1-ABC-123", "person"} {
		if strings.Contains(string(encoded), leaked) {
			t.Fatalf("invocation leaked %q: %s", leaked, encoded)
		}
	}

	// The device site list is copied, not shared with the run.
	invocation.Device.SiteIDs[0] = "changed"
	if run.Device.SiteIds[0] != "site-1" {
		t.Fatal("invocation shares the run's site slice")
	}
}

func TestNewWorkflowInvocationLegacyTenantAndStableID(t *testing.T) {
	run := &WorkflowRun{
		RunId:      "run-1",
		WorkflowId: "workflow-1",
		Operation:  "mqtt",
		SignedURL:  "https://vault.example.com/first",
		User:       WorkflowUser{OrganisationId: " owner-1 "},
	}
	first, err := NewWorkflowInvocation(run)
	if err != nil {
		t.Fatalf("NewWorkflowInvocation() error = %v", err)
	}
	if first.Tenant != (WorkflowTenantReference{OrganisationID: "owner-1", ProjectID: "owner-1"}) {
		t.Fatalf("legacy tenant = %#v", first.Tenant)
	}
	if want := "wfx_e73be8feaf919e438728761f1bc0813e59f875f13526487e93d63f0d23de8611"; first.ExecutionID != want {
		t.Fatalf("ExecutionID = %q, want %q", first.ExecutionID, want)
	}

	run.SignedURL = "https://vault.example.com/second"
	second, err := NewWorkflowInvocation(run)
	if err != nil {
		t.Fatalf("NewWorkflowInvocation() error = %v", err)
	}
	if second.ExecutionID != first.ExecutionID {
		t.Fatal("execution identity changed with the signed URL")
	}

	projectOnly, _ := primitive.ObjectIDFromHex("507f1f77bcf86cd799439011")
	tenant := WorkflowRunTenant(&WorkflowRun{User: WorkflowUser{ProjectId: &projectOnly}})
	if tenant != (WorkflowTenantReference{OrganisationID: projectOnly.Hex(), ProjectID: projectOnly.Hex()}) {
		t.Fatalf("project-only tenant = %#v", tenant)
	}
	if (WorkflowRunTenant(nil) != WorkflowTenantReference{}) {
		t.Fatal("nil run must yield an empty tenant")
	}
}

func TestNewWorkflowInvocationRequiresIdentity(t *testing.T) {
	if _, err := NewWorkflowInvocation(nil); err == nil {
		t.Fatal("nil run accepted")
	}
	for _, run := range []*WorkflowRun{
		{WorkflowId: "workflow-1", Operation: "mqtt"},
		{RunId: "run-1", Operation: "mqtt"},
		{RunId: "run-1", WorkflowId: "workflow-1"},
	} {
		if _, err := NewWorkflowInvocation(run); err == nil {
			t.Fatalf("run without full identity accepted: %#v", run)
		}
	}
}
