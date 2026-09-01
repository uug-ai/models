package api

import "github.com/uug-ai/models/pkg/models"

// WorkflowStatus represents specific status codes for workflow operations.
type WorkflowStatus string

const (
	WorkflowBindingFailed       WorkflowStatus = "workflow_binding_failed"
	WorkflowMissingInfo         WorkflowStatus = "workflow_missing_info"
	WorkflowFound               WorkflowStatus = "workflow_found"
	WorkflowNotFound            WorkflowStatus = "workflow_not_found"
	WorkflowRetrievalSuccess    WorkflowStatus = "workflow_retrieval_success"
	WorkflowRetrievalFailed     WorkflowStatus = "workflow_retrieval_failed"
	WorkflowAddSuccess          WorkflowStatus = "workflow_add_success"
	WorkflowAddFailed           WorkflowStatus = "workflow_add_failed"
	WorkflowUpdateSuccess       WorkflowStatus = "workflow_update_success"
	WorkflowUpdateFailed        WorkflowStatus = "workflow_update_failed"
	WorkflowDeleteSuccess       WorkflowStatus = "workflow_delete_success"
	WorkflowDeleteFailed        WorkflowStatus = "workflow_delete_failed"
	WorkflowDuplicateName       WorkflowStatus = "workflow_duplicate_name"
	WorkflowForbidden           WorkflowStatus = "workflow_forbidden"
	WorkflowRunSuccess          WorkflowStatus = "workflow_run_success"
	WorkflowRunFailed           WorkflowStatus = "workflow_run_failed"
	WorkflowRunNoMedia          WorkflowStatus = "workflow_run_no_media"
	WorkflowRunsFound           WorkflowStatus = "workflow_runs_found"
	WorkflowRunsRetrievalFailed WorkflowStatus = "workflow_runs_retrieval_failed"
)

// String returns the string representation of the workflow status.
func (cs WorkflowStatus) String() string {
	return string(cs)
}

// Translate returns the translated string representation of the workflow status in the specified language.
func (cs WorkflowStatus) Translate(lang string) string {
	translations := map[string]map[WorkflowStatus]string{
		"en": {
			WorkflowBindingFailed:       "Workflow binding failed",
			WorkflowMissingInfo:         "Workflow missing information",
			WorkflowFound:               "Workflow found",
			WorkflowNotFound:            "Workflow not found",
			WorkflowRetrievalSuccess:    "Workflow retrieved successfully",
			WorkflowRetrievalFailed:     "Workflow retrieval failed",
			WorkflowAddSuccess:          "Workflow added successfully",
			WorkflowAddFailed:           "Workflow failed to add",
			WorkflowUpdateSuccess:       "Workflow updated successfully",
			WorkflowUpdateFailed:        "Workflow failed to update",
			WorkflowDeleteSuccess:       "Workflow deleted successfully",
			WorkflowDeleteFailed:        "Workflow failed to delete",
			WorkflowDuplicateName:       "Workflow with this name already exists",
			WorkflowForbidden:           "You do not have permission for this action",
			WorkflowRunSuccess:          "Workflow run(s) launched successfully",
			WorkflowRunFailed:           "Workflow run(s) failed to launch",
			WorkflowRunNoMedia:          "No eligible media to run the workflow on",
			WorkflowRunsFound:           "Workflow runs retrieved successfully",
			WorkflowRunsRetrievalFailed: "Workflow runs retrieval failed",
		},
	}

	if langTranslations, exists := translations[lang]; exists {
		if translation, exists := langTranslations[cs]; exists {
			return translation
		}
	}

	if enTranslations, exists := translations["en"]; exists {
		if translation, exists := enTranslations[cs]; exists {
			return translation
		}
	}

	return cs.String()
}

// WorkflowFilter narrows a workflow listing. Every field is optional; an unset
// field does not constrain the result. Scalars are pointers so "not provided" is
// distinct from a zero value, and DeviceKeys is a set matched against automatic
// triggers' device scope. It is the workflow counterpart to MediaFilter, posted
// to /workflows/filter for criteria (notably a set of device keys) that GET
// query params fit poorly.
type WorkflowFilter struct {
	Source      *models.WorkflowSource         `json:"source,omitempty" bson:"source,omitempty"`
	Surface     *models.WorkflowTriggerSurface `json:"surface,omitempty" bson:"surface,omitempty"`
	TriggerType *models.WorkflowTriggerType    `json:"triggerType,omitempty" bson:"triggerType,omitempty"`
	Enabled     *bool                          `json:"enabled,omitempty" bson:"enabled,omitempty"`
	DeviceKeys  []string                       `json:"deviceKeys,omitempty" bson:"deviceKeys,omitempty"`
}

// GetWorkflows
// @Router /workflows/filter [post]
type GetWorkflowsRequest struct {
	Filter WorkflowFilter `json:"filter" bson:"filter"`
}
type GetWorkflowsResponse struct {
	Workflows []models.Workflow `json:"workflows"`
}
type GetWorkflowsSuccessResponse struct {
	SuccessResponse
	Data GetWorkflowsResponse `json:"data"`
}
type GetWorkflowsErrorResponse struct {
	ErrorResponse
}

// GetWorkflow
type GetWorkflowRequest struct{}
type GetWorkflowResponse struct {
	Workflow models.Workflow `json:"workflow"`
}
type GetWorkflowSuccessResponse struct {
	SuccessResponse
	Data GetWorkflowResponse `json:"data"`
}
type GetWorkflowErrorResponse struct {
	ErrorResponse
}

// CreateWorkflow
type CreateWorkflowRequest struct {
	Workflow models.Workflow `json:"workflow"`
}
type CreateWorkflowResponse struct {
	Workflow models.Workflow `json:"workflow"`
}
type CreateWorkflowSuccessResponse struct {
	SuccessResponse
	Data CreateWorkflowResponse `json:"data"`
}
type CreateWorkflowErrorResponse struct {
	ErrorResponse
}

// UpdateWorkflow
type UpdateWorkflowRequest struct {
	Workflow models.Workflow `json:"workflow"`
}
type UpdateWorkflowResponse struct {
	Workflow models.Workflow `json:"workflow"`
}
type UpdateWorkflowSuccessResponse struct {
	SuccessResponse
	Data UpdateWorkflowResponse `json:"data"`
}
type UpdateWorkflowErrorResponse struct {
	ErrorResponse
}

// DeleteWorkflow
type DeleteWorkflowRequest struct{}
type DeleteWorkflowResponse struct{}
type DeleteWorkflowSuccessResponse struct {
	SuccessResponse
	Data DeleteWorkflowResponse `json:"data"`
}
type DeleteWorkflowErrorResponse struct {
	ErrorResponse
}

// RunWorkflow launches a workflow on demand over a set of a case's source
// media. It is the manual counterpart to the automatic analysis hand-off: the
// caller picks a workflow and the media to send through, and the server fans
// out one run per selected recording onto the workflows queue.
//
// @Router /tasks/{taskId}/workflows [post]
//
// WorkflowId is the id of the workflow to run (a config or user workflow that
// exposes a manual trigger on the case surface). MediaIds is the set of
// case_media source-row ids to run it over; an empty MediaIds means "every
// source media on the case". AttachmentIds additionally selects video
// CaseAttachments to run it over — each is materialised into a linked
// Role=source case_media row on demand so attached videos flow through the
// same run machinery as device recordings.
type RunWorkflowRequest struct {
	WorkflowId    string   `json:"workflowId" bson:"workflowId"`
	MediaIds      []string `json:"mediaIds,omitempty" bson:"mediaIds,omitempty"`
	AttachmentIds []string `json:"attachmentIds,omitempty" bson:"attachmentIds,omitempty"`
}

// RunWorkflowResponse reports the runs opened by the launch: the freshly minted
// run ids (one per selected media) and their count.
type RunWorkflowResponse struct {
	RunIds []string `json:"runIds"`
	Count  int      `json:"count"`
}
type RunWorkflowSuccessResponse struct {
	SuccessResponse
	Data RunWorkflowResponse `json:"data"`
}
type RunWorkflowErrorResponse struct {
	ErrorResponse
}

// WorkflowRunStatus is the slim, client-facing status of a single workflow run,
// projected from a persisted models.WorkflowRun. It exists because the run's
// lifecycle fields (start/end, dispatched/resolved) are persistence-only and
// never cross the wire on the run itself, so the state a surface needs to render
// "still working" vs "results are in" is derived server-side (via
// WorkflowRun.LifecycleState) and carried here instead. It is surface-agnostic:
// the same shape serves a case today and any future launch surface.
type WorkflowRunStatus struct {
	RunId        string `json:"runId"`
	WorkflowId   string `json:"workflowId,omitempty"`
	WorkflowName string `json:"workflowName,omitempty"`
	Origin       string `json:"origin,omitempty"`
	SourceRef    string `json:"sourceRef,omitempty"`
	Key          string `json:"key,omitempty"`
	// State is the derived lifecycle: running | completed | noResult
	// (models.WorkflowRunState).
	State string `json:"state"`
	// Start / End are unix seconds; End is 0 while the run is still open.
	Start int64 `json:"start,omitempty"`
	End   int64 `json:"end,omitempty"`
	// Dispatched / Resolved are the sizes of the run's dispatched and resolved
	// operation sets, exposed as a coarse progress hint.
	Dispatched int `json:"dispatched"`
	Resolved   int `json:"resolved"`
	// DispatchedOperations / ResolvedOperations name the stages behind the
	// Dispatched / Resolved counts, in dispatch/resolution order, so a surface
	// can render per-stage progress (e.g. "pose done, redaction running")
	// instead of only a run-level running/completed flip for multi-stage runs.
	DispatchedOperations []string `json:"dispatchedOperations,omitempty"`
	ResolvedOperations   []string `json:"resolvedOperations,omitempty"`
	// HasResults is true when the run accumulated any stage output.
	HasResults bool `json:"hasResults"`
}

// WorkflowRunStatusSummary aggregates a run set by state so a surface can render
// a headline ("3 running / 1 done") without re-tallying client-side.
type WorkflowRunStatusSummary struct {
	Total     int `json:"total"`
	Running   int `json:"running"`
	Completed int `json:"completed"`
	NoResult  int `json:"noResult"`
}

// GetWorkflowRuns reports the status of the workflow runs launched from a
// surface. Runs are grouped by the launch's SourceRef (e.g. a case id) and
// scoped to the caller's organisation; an optional RunIds filter narrows to a
// specific launch.
//
// @Router /tasks/{taskId}/workflow-runs [get]
type GetWorkflowRunsResponse struct {
	Runs    []WorkflowRunStatus      `json:"runs"`
	Summary WorkflowRunStatusSummary `json:"summary"`
}
type GetWorkflowRunsSuccessResponse struct {
	SuccessResponse
	Data GetWorkflowRunsResponse `json:"data"`
}
type GetWorkflowRunsErrorResponse struct {
	ErrorResponse
}
