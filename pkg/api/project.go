package api

import "github.com/uug-ai/models/pkg/models"

// ProjectStatus represents specific status codes for project operations.
type ProjectStatus string

const (
	ProjectBindingFailed ProjectStatus = "project_binding_failed"
	ProjectMissingInfo   ProjectStatus = "project_missing_info"
	ProjectFound         ProjectStatus = "project_found"
	ProjectNotFound      ProjectStatus = "project_not_found"
	ProjectGetAllSuccess ProjectStatus = "project_get_all_success"
	ProjectGetAllFailed  ProjectStatus = "project_get_all_failed"
	ProjectUpdateSuccess ProjectStatus = "project_update_success"
	ProjectUpdateFailed  ProjectStatus = "project_update_failed"
)

// String returns the string representation of the project status.
func (s ProjectStatus) String() string {
	return string(s)
}

// Translate returns the translated string representation of the project status
// in the specified language.
func (s ProjectStatus) Translate(lang string) string {
	translations := map[string]map[ProjectStatus]string{
		"en": {
			ProjectBindingFailed: "Project binding failed",
			ProjectMissingInfo:   "Project is missing required information",
			ProjectFound:         "Project found",
			ProjectNotFound:      "Project not found",
			ProjectGetAllSuccess: "Projects retrieved successfully",
			ProjectGetAllFailed:  "Failed to retrieve projects",
			ProjectUpdateSuccess: "Project updated successfully",
			ProjectUpdateFailed:  "Failed to update project",
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

// ---------- Request / Response DTOs (project domain) ----------
//
// These are the typed request bodies and response envelopes for the versioned
// project selection endpoints, mirroring the organisation domain. A project is
// an optional sub-scope within an organisation; a nil active project means the
// caller is scoped organisation-wide.

// GetProjects lists the projects in the caller's active organisation.
// @Router /projects [get]
type GetProjectsResponse struct {
	Projects []models.Project `json:"projects"`
}
type GetProjectsSuccessResponse struct {
	SuccessResponse
	Data GetProjectsResponse `json:"data"`
}
type GetProjectsErrorResponse struct {
	ErrorResponse
}

// GetProject resolves a single project, currently the caller's active one
// (/projects/current).
type GetProjectResponse struct {
	Project models.Project `json:"project"`
}
type GetProjectSuccessResponse struct {
	SuccessResponse
	Data GetProjectResponse `json:"data"`
}
type GetProjectErrorResponse struct {
	ErrorResponse
}

// SetCurrentProject selects the active project sub-scope for the caller.
// @Router /projects/current [patch]
type SetCurrentProjectRequest struct {
	ProjectId string `json:"projectId"`
}
type SetCurrentProjectResponse struct {
	Project models.Project `json:"project"`
}
type SetCurrentProjectSuccessResponse struct {
	SuccessResponse
	Data SetCurrentProjectResponse `json:"data"`
}
type SetCurrentProjectErrorResponse struct {
	ErrorResponse
}
