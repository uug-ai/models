package models

const (
	WorkflowInvocationSchemaV1 = "uug.ai/workflow-invocation/v1"
	WorkflowResultSchemaV1     = "uug.ai/workflow-result/v1"
	WorkflowCallbackMethodPost = "POST"
)

// WorkflowInvocation is the versioned customer-facing projection of a
// WorkflowRun stage dispatch.
type WorkflowInvocation struct {
	Schema      string                     `json:"schema" bson:"-"`
	ExecutionID string                     `json:"executionId" bson:"-"`
	RunID       string                     `json:"runId" bson:"-"`
	Workflow    WorkflowInvocationWorkflow `json:"workflow" bson:"-"`
	Stage       WorkflowStageReference     `json:"stage" bson:"-"`
	Tenant      WorkflowTenantReference    `json:"tenant" bson:"-"`
	Media       WorkflowMediaReference     `json:"media" bson:"-"`
	Device      WorkflowDeviceReference    `json:"device" bson:"-"`
	Data        map[string]any             `json:"data,omitempty" bson:"-"`
	Callback    *WorkflowCallback          `json:"callback,omitempty" bson:"-"`
	TraceID     string                     `json:"traceId,omitempty" bson:"-"`
}

// WorkflowInvocationWorkflow identifies the workflow definition being run.
type WorkflowInvocationWorkflow struct {
	ID   string `json:"id" bson:"-"`
	Name string `json:"name,omitempty" bson:"-"`
}

// WorkflowStageReference identifies a stage by its trusted operation.
type WorkflowStageReference struct {
	Operation string `json:"operation" bson:"-"`
}

// WorkflowTenantReference identifies the owning organisation and project.
type WorkflowTenantReference struct {
	OrganisationID string `json:"organisationId" bson:"-"`
	ProjectID      string `json:"projectId" bson:"-"`
}

// WorkflowMediaReference identifies the media processed by a workflow run.
type WorkflowMediaReference struct {
	Key       string `json:"key" bson:"-"`
	SignedURL string `json:"signedUrl,omitempty" bson:"-"`
}

// WorkflowDeviceReference carries the safe device context for an invocation.
type WorkflowDeviceReference struct {
	Key     string   `json:"key,omitempty" bson:"-"`
	Name    string   `json:"name,omitempty" bson:"-"`
	SiteIDs []string `json:"siteIds,omitempty" bson:"-"`
}

// WorkflowCallback describes how an external stage returns its result. The
// caller authenticates with its own Hub bearer token.
type WorkflowCallback struct {
	URL          string `json:"url" bson:"-"`
	Method       string `json:"method" bson:"-"`
	ResultSchema string `json:"resultSchema" bson:"-"`
}

// WorkflowResult is the versioned request body submitted by an external stage.
type WorkflowResult struct {
	Schema string                 `json:"schema" bson:"-"`
	Stage  WorkflowStageReference `json:"stage" bson:"-"`
	Result map[string]any         `json:"result" bson:"-"`
}
