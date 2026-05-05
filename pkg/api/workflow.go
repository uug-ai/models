package api

import "github.com/uug-ai/models/pkg/models"

// WorkflowStatus represents specific status codes for workflow operations.
type WorkflowStatus string

const (
	WorkflowBindingFailed    WorkflowStatus = "workflow_binding_failed"
	WorkflowMissingInfo      WorkflowStatus = "workflow_missing_info"
	WorkflowFound            WorkflowStatus = "workflow_found"
	WorkflowNotFound         WorkflowStatus = "workflow_not_found"
	WorkflowRetrievalSuccess WorkflowStatus = "workflow_retrieval_success"
	WorkflowRetrievalFailed  WorkflowStatus = "workflow_retrieval_failed"
	WorkflowAddSuccess       WorkflowStatus = "workflow_add_success"
	WorkflowAddFailed        WorkflowStatus = "workflow_add_failed"
	WorkflowUpdateSuccess    WorkflowStatus = "workflow_update_success"
	WorkflowUpdateFailed     WorkflowStatus = "workflow_update_failed"
	WorkflowDeleteSuccess    WorkflowStatus = "workflow_delete_success"
	WorkflowDeleteFailed     WorkflowStatus = "workflow_delete_failed"
	WorkflowDuplicateName    WorkflowStatus = "workflow_duplicate_name"
	WorkflowForbidden        WorkflowStatus = "workflow_forbidden"
)

// String returns the string representation of the workflow status.
func (cs WorkflowStatus) String() string {
	return string(cs)
}

// Translate returns the translated string representation of the workflow status in the specified language.
func (cs WorkflowStatus) Translate(lang string) string {
	translations := map[string]map[WorkflowStatus]string{
		"en": {
			WorkflowBindingFailed:    "Workflow binding failed",
			WorkflowMissingInfo:      "Workflow missing information",
			WorkflowFound:            "Workflow found",
			WorkflowNotFound:         "Workflow not found",
			WorkflowRetrievalSuccess: "Workflow retrieved successfully",
			WorkflowRetrievalFailed:  "Workflow retrieval failed",
			WorkflowAddSuccess:       "Workflow added successfully",
			WorkflowAddFailed:        "Workflow failed to add",
			WorkflowUpdateSuccess:    "Workflow updated successfully",
			WorkflowUpdateFailed:     "Workflow failed to update",
			WorkflowDeleteSuccess:    "Workflow deleted successfully",
			WorkflowDeleteFailed:     "Workflow failed to delete",
			WorkflowDuplicateName:    "Workflow with this name already exists",
			WorkflowForbidden:        "You do not have permission for this action",
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

// GetWorkflows
type GetWorkflowsRequest struct{}
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
