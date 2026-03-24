package api

import "github.com/uug-ai/models/pkg/models"

// VideowallStatus represents specific status codes for videowall operations.
type VideowallStatus string

const (
	VideowallBindingFailed    VideowallStatus = "videowall_binding_failed"
	VideowallMissingInfo      VideowallStatus = "videowall_missing_info"
	VideowallFound            VideowallStatus = "videowall_found"
	VideowallNotFound         VideowallStatus = "videowall_not_found"
	VideowallRetrievalSuccess VideowallStatus = "videowall_retrieval_success"
	VideowallRetrievalFailed  VideowallStatus = "videowall_retrieval_failed"
	VideowallAddSuccess       VideowallStatus = "videowall_add_success"
	VideowallAddFailed        VideowallStatus = "videowall_add_failed"
	VideowallUpdateSuccess    VideowallStatus = "videowall_update_success"
	VideowallUpdateFailed     VideowallStatus = "videowall_update_failed"
	VideowallDeleteSuccess    VideowallStatus = "videowall_delete_success"
	VideowallDeleteFailed     VideowallStatus = "videowall_delete_failed"
	VideowallDuplicateName    VideowallStatus = "videowall_duplicate_name"
	VideowallForbidden        VideowallStatus = "videowall_forbidden"
	VideowallDecryptSuccess   VideowallStatus = "videowall_decrypt_success"
	VideowallDecryptFailed    VideowallStatus = "videowall_decrypt_failed"
	VideowallInactive         VideowallStatus = "videowall_inactive"
)

// String returns the string representation of the videowall status.
func (cs VideowallStatus) String() string {
	return string(cs)
}

// Translate returns the translated string representation of the videowall status in the specified language.
func (cs VideowallStatus) Translate(lang string) string {
	translations := map[string]map[VideowallStatus]string{
		"en": {
			VideowallBindingFailed:    "Videowall binding failed",
			VideowallMissingInfo:      "Videowall missing information",
			VideowallFound:            "Videowall found",
			VideowallNotFound:         "Videowall not found",
			VideowallRetrievalSuccess: "Videowall retrieved successfully",
			VideowallRetrievalFailed:  "Videowall retrieval failed",
			VideowallAddSuccess:       "Videowall added successfully",
			VideowallAddFailed:        "Videowall failed to add",
			VideowallUpdateSuccess:    "Videowall updated successfully",
			VideowallUpdateFailed:     "Videowall failed to update",
			VideowallDeleteSuccess:    "Videowall deleted successfully",
			VideowallDeleteFailed:     "Videowall failed to delete",
			VideowallDuplicateName:    "Videowall with this name already exists",
			VideowallForbidden:        "You do not have permission for this action",
			VideowallDecryptSuccess:   "Videowall decrypted successfully",
			VideowallDecryptFailed:    "Videowall decryption failed",
			VideowallInactive:         "Videowall is not active",
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

// GetVideowalls
type GetVideowallsRequest struct{}
type GetVideowallsResponse struct {
	Videowalls []models.Videowall `json:"videowalls"`
}
type GetVideowallsSuccessResponse struct {
	SuccessResponse
	Data GetVideowallsResponse `json:"data"`
}
type GetVideowallsErrorResponse struct {
	ErrorResponse
}

// GetVideowall
type GetVideowallRequest struct{}
type GetVideowallResponse struct {
	Videowall models.Videowall `json:"videowall"`
}
type GetVideowallSuccessResponse struct {
	SuccessResponse
	Data GetVideowallResponse `json:"data"`
}
type GetVideowallErrorResponse struct {
	ErrorResponse
}

// CreateVideowall
type CreateVideowallRequest struct {
	Videowall models.Videowall `json:"videowall"`
}
type CreateVideowallResponse struct {
	Videowall models.Videowall `json:"videowall"`
}
type CreateVideowallSuccessResponse struct {
	SuccessResponse
	Data CreateVideowallResponse `json:"data"`
}
type CreateVideowallErrorResponse struct {
	ErrorResponse
}

// UpdateVideowall
type UpdateVideowallRequest struct {
	Videowall models.Videowall `json:"videowall"`
}
type UpdateVideowallResponse struct {
	Videowall models.Videowall `json:"videowall"`
}
type UpdateVideowallSuccessResponse struct {
	SuccessResponse
	Data UpdateVideowallResponse `json:"data"`
}
type UpdateVideowallErrorResponse struct {
	ErrorResponse
}

// PatchVideowall
type PatchVideowallRequest struct {
	Updates map[string]interface{} `json:"updates"`
}
type PatchVideowallResponse struct {
	Videowall models.Videowall `json:"videowall"`
}
type PatchVideowallSuccessResponse struct {
	SuccessResponse
	Data PatchVideowallResponse `json:"data"`
}
type PatchVideowallErrorResponse struct {
	ErrorResponse
}

// DeleteVideowall
type DeleteVideowallRequest struct{}
type DeleteVideowallResponse struct{}
type DeleteVideowallSuccessResponse struct {
	SuccessResponse
	Data DeleteVideowallResponse `json:"data"`
}
type DeleteVideowallErrorResponse struct {
	ErrorResponse
}
