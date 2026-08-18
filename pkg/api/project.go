package api

import "github.com/uug-ai/models/pkg/models"

// ProjectStatus represents specific status codes for project operations.
type ProjectStatus string

const (
	ProjectBindingFailed    ProjectStatus = "project_binding_failed"
	ProjectMissingInfo      ProjectStatus = "project_missing_info"
	ProjectFound            ProjectStatus = "project_found"
	ProjectNotFound         ProjectStatus = "project_not_found"
	ProjectForbidden        ProjectStatus = "project_forbidden"
	ProjectNameExists       ProjectStatus = "project_name_exists"
	ProjectDefaultImmutable ProjectStatus = "project_default_immutable"
	ProjectGetAllSuccess    ProjectStatus = "project_get_all_success"
	ProjectGetAllFailed     ProjectStatus = "project_get_all_failed"
	ProjectCreateSuccess    ProjectStatus = "project_create_success"
	ProjectCreateFailed     ProjectStatus = "project_create_failed"
	ProjectUpdateSuccess    ProjectStatus = "project_update_success"
	ProjectUpdateFailed     ProjectStatus = "project_update_failed"
	ProjectDeleteSuccess    ProjectStatus = "project_delete_success"
	ProjectDeleteFailed     ProjectStatus = "project_delete_failed"
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
			ProjectBindingFailed:    "Project binding failed",
			ProjectMissingInfo:      "Project is missing required information",
			ProjectFound:            "Project found",
			ProjectNotFound:         "Project not found",
			ProjectForbidden:        "You do not have permission for this action",
			ProjectNameExists:       "Project slug already exists",
			ProjectDefaultImmutable: "The default project cannot be renamed or deleted",
			ProjectGetAllSuccess:    "Projects retrieved successfully",
			ProjectGetAllFailed:     "Failed to retrieve projects",
			ProjectCreateSuccess:    "Project created successfully",
			ProjectCreateFailed:     "Failed to create project",
			ProjectUpdateSuccess:    "Project updated successfully",
			ProjectUpdateFailed:     "Failed to update project",
			ProjectDeleteSuccess:    "Project deleted successfully",
			ProjectDeleteFailed:     "Failed to delete project",
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
// Soft-deleted projects are omitted unless ?includeInactive=true is supplied.
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

// GetProject resolves a single project, either the caller's active one
// (/projects/current) or one addressed by id (/projects/{id}).
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

// CreateProject creates a project inside the caller's active organisation. The
// organisation, id and audit stamps on the supplied project are ignored and
// filled in server-side; name and slug are required.
// @Router /projects [post]
type CreateProjectRequest struct {
	Project models.Project `json:"project"`
}
type CreateProjectResponse struct {
	Project models.Project `json:"project"`
}
type CreateProjectSuccessResponse struct {
	SuccessResponse
	Data CreateProjectResponse `json:"data"`
}
type CreateProjectErrorResponse struct {
	ErrorResponse
}

// UpdateProject applies a partial update. The body is a ProjectUpdate patch:
// only the fields present are changed (each field is optional).
// @Router /projects/{id} [patch]
type UpdateProjectRequest struct {
	Project models.ProjectUpdate `json:"project"`
}
type UpdateProjectResponse struct {
	Project models.Project `json:"project"`
}
type UpdateProjectSuccessResponse struct {
	SuccessResponse
	Data UpdateProjectResponse `json:"data"`
}
type UpdateProjectErrorResponse struct {
	ErrorResponse
}

// DeleteProject soft-deletes a project. The returned project is the stored
// document with its IsActive flag cleared; the organisation's default project
// cannot be deleted.
// @Router /projects/{id} [delete]
type DeleteProjectResponse struct {
	Project models.Project `json:"project"`
}
type DeleteProjectSuccessResponse struct {
	SuccessResponse
	Data DeleteProjectResponse `json:"data"`
}
type DeleteProjectErrorResponse struct {
	ErrorResponse
}
