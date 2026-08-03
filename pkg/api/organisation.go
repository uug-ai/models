package api

import "github.com/uug-ai/models/pkg/models"

// OrganisationStatus represents specific status codes for organisation operations.
type OrganisationStatus string

const (
	OrganisationBindingFailed  OrganisationStatus = "organisation_binding_failed"
	OrganisationMissingInfo    OrganisationStatus = "organisation_missing_info"
	OrganisationNameExists     OrganisationStatus = "organisation_name_exists"
	OrganisationFound          OrganisationStatus = "organisation_found"
	OrganisationNotFound       OrganisationStatus = "organisation_not_found"
	OrganisationGetAllSuccess  OrganisationStatus = "organisation_get_all_success"
	OrganisationGetAllFailed   OrganisationStatus = "organisation_get_all_failed"
	OrganisationCreateSuccess  OrganisationStatus = "organisation_create_success"
	OrganisationCreateFailed   OrganisationStatus = "organisation_create_failed"
	OrganisationUpdateSuccess  OrganisationStatus = "organisation_update_success"
	OrganisationUpdateFailed   OrganisationStatus = "organisation_update_failed"
	OrganisationDeleteSuccess  OrganisationStatus = "organisation_delete_success"
	OrganisationDeleteFailed   OrganisationStatus = "organisation_delete_failed"
	OrganisationValidationFail OrganisationStatus = "organisation_validation_failed"
)

// String returns the string representation of the organisation status.
func (s OrganisationStatus) String() string {
	return string(s)
}

// Translate returns the translated string representation of the organisation
// status in the specified language.
func (s OrganisationStatus) Translate(lang string) string {
	translations := map[string]map[OrganisationStatus]string{
		"en": {
			OrganisationBindingFailed:  "Organisation binding failed",
			OrganisationMissingInfo:    "Organisation is missing required information",
			OrganisationNameExists:     "Organisation name already exists",
			OrganisationFound:          "Organisation found",
			OrganisationNotFound:       "Organisation not found",
			OrganisationGetAllSuccess:  "Organisations retrieved successfully",
			OrganisationGetAllFailed:   "Failed to retrieve organisations",
			OrganisationCreateSuccess:  "Organisation created successfully",
			OrganisationCreateFailed:   "Failed to create organisation",
			OrganisationUpdateSuccess:  "Organisation updated successfully",
			OrganisationUpdateFailed:   "Failed to update organisation",
			OrganisationDeleteSuccess:  "Organisation deleted successfully",
			OrganisationDeleteFailed:   "Failed to delete organisation",
			OrganisationValidationFail: "Organisation validation failed",
		},
	}

	if langTranslations, exists := translations[lang]; exists {
		if translation, exists := langTranslations[s]; exists {
			return translation
		}
	}

	if enTranslations, exists := translations["en"]; exists {
		if translation, exists := enTranslations[s]; exists {
			return translation
		}
	}

	return s.String()
}

// ---------- Request / Response DTOs (organisation domain) ----------
//
// These are the typed request bodies and response envelopes for the versioned
// organisation CRUD endpoints, mirroring the workflow/videowall domains. The
// admin-facing organisation DTOs (loosely typed, see admin.go) are a separate,
// legacy surface and are intentionally left untouched.

// GetOrganisations
// @Router /organisations [get]
type GetOrganisationsResponse struct {
	Organisations []models.Organisation `json:"organisations"`
}
type GetOrganisationsSuccessResponse struct {
	SuccessResponse
	Data GetOrganisationsResponse `json:"data"`
}
type GetOrganisationsErrorResponse struct {
	ErrorResponse
}

// GetOrganisation resolves a single organisation, either by id
// (/organisations/{id}) or the caller's active one (/organisations/current).
type GetOrganisationResponse struct {
	Organisation models.Organisation `json:"organisation"`
}
type GetOrganisationSuccessResponse struct {
	SuccessResponse
	Data GetOrganisationResponse `json:"data"`
}
type GetOrganisationErrorResponse struct {
	ErrorResponse
}

// SetCurrentOrganisation selects the active organisation for the caller.
// @Router /organisations/current [patch]
type SetCurrentOrganisationRequest struct {
	OrganisationId string `json:"organisationId"`
}
type SetCurrentOrganisationResponse struct {
	Organisation models.Organisation `json:"organisation"`
}
type SetCurrentOrganisationSuccessResponse struct {
	SuccessResponse
	Data SetCurrentOrganisationResponse `json:"data"`
}
type SetCurrentOrganisationErrorResponse struct {
	ErrorResponse
}

// CreateOrganisation
// @Router /organisations [post]
type CreateOrganisationRequest struct {
	Organisation models.Organisation `json:"organisation"`
}
type CreateOrganisationResponse struct {
	Organisation models.Organisation `json:"organisation"`
}
type CreateOrganisationSuccessResponse struct {
	SuccessResponse
	Data CreateOrganisationResponse `json:"data"`
}
type CreateOrganisationErrorResponse struct {
	ErrorResponse
}

// UpdateOrganisation applies a partial update. The body is an OrganisationUpdate
// patch: only the fields present are changed (each field is optional).
// @Router /organisations/{id} [patch]
type UpdateOrganisationRequest struct {
	Organisation models.OrganisationUpdate `json:"organisation"`
}
type UpdateOrganisationResponse struct {
	Organisation models.Organisation `json:"organisation"`
}
type UpdateOrganisationSuccessResponse struct {
	SuccessResponse
	Data UpdateOrganisationResponse `json:"data"`
}
type UpdateOrganisationErrorResponse struct {
	ErrorResponse
}
