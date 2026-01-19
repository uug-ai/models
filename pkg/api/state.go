package api

import "github.com/uug-ai/models/pkg/models"

// StateStatus represents specific status codes for device operations
type StateStatus string

const (
	StateRetrievalSuccess StateStatus = "state_retrieval_success"
	StateBindingFailed    StateStatus = "state_binding_failed"
	StateDuplicateName    StateStatus = "state_duplicate_name"
	StateMissingInfo      StateStatus = "state_missing_info"
	StateRetrievalFailed  StateStatus = "state_retrieval_failed"
	StateFound            StateStatus = "state_found"
	StateNotFound         StateStatus = "state_not_found"
	StateAddSuccess       StateStatus = "state_add_success"
	StateAddFailed        StateStatus = "state_add_failed"
	StateUpdateSuccess    StateStatus = "state_update_success"
	StateUpdateFailed     StateStatus = "state_update_failed"
	StateDeleteSuccess    StateStatus = "state_delete_success"
	StateDeleteFailed     StateStatus = "state_delete_failed"
)

// String returns the string representation of the device status
func (ds StateStatus) String() string {
	return string(ds)
}

// Into returns the translated string representation of the device status in the specified language
func (ds StateStatus) Translate(lang string) string {
	translations := map[string]map[StateStatus]string{
		"en": {
			StateBindingFailed:    "State binding failed",
			StateDuplicateName:    "State duplicate name",
			StateMissingInfo:      "State missing information",
			StateRetrievalFailed:  "State retrieval failed",
			StateFound:            "State found",
			StateNotFound:         "State not found",
			StateAddSuccess:       "State added successfully",
			StateAddFailed:        "State failed to add",
			StateUpdateSuccess:    "State updated successfully",
			StateUpdateFailed:     "State failed to update",
			StateDeleteSuccess:    "State deleted successfully",
			StateDeleteFailed:     "State failed to delete",
			StateRetrievalSuccess: "State retrieved successfully",
		},
	}

	if langTranslations, exists := translations[lang]; exists {
		if translation, exists := langTranslations[ds]; exists {
			return translation
		}
	}

	// Default to English if language not found or translation doesn't exist
	if enTranslations, exists := translations["en"]; exists {
		if translation, exists := enTranslations[ds]; exists {
			return translation
		}
	}

	// Fallback to the string representation
	return ds.String()
}

// GetStatesRequest represents the request to get states
// @Router /states [get]
type GetStatesRequest struct {
}
type GetStatesResponse struct {
	States []models.State `json:"states"`
}
type GetStatesSuccessResponse struct {
	SuccessResponse
	Data GetStatesResponse `json:"data"`
}
type GetStatesErrorResponse struct {
	ErrorResponse
}

// AddStateRequest represents the request to create a state
// @Router /states [post]
type AddStateRequest struct {
	State models.State `json:"state" binding:"required"`
}
type AddStateResponse struct {
	State models.State `json:"state"`
}
type AddStateSuccessResponse struct {
	SuccessResponse
	Data AddStateResponse `json:"data"`
}
type AddStateErrorResponse struct {
	ErrorResponse
}

// UpdateStateRequest represents the request to update a state
// @Router /states/{stateId} [put]
type UpdateStateRequest struct {
	State models.State `json:"state" binding:"required"`
}
type UpdateStateResponse struct {
	State models.State `json:"state"`
}
type UpdateStateSuccessResponse struct {
	SuccessResponse
	Data UpdateStateResponse `json:"data"`
}
type UpdateStateErrorResponse struct {
	ErrorResponse
}

// DeleteStateRequest represents the request to delete a state
// @Router /states/{stateId} [delete]
type DeleteStateRequest struct {
}
type DeleteStateResponse struct {
}
type DeleteStateSuccessResponse struct {
	SuccessResponse
	Data DeleteStateResponse `json:"data"`
}
type DeleteStateErrorResponse struct {
	ErrorResponse
}
