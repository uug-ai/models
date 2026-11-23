package api

import (
	"time"

	"github.com/uug-ai/models/pkg/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ---------- Request / Response DTOs (project access, RBAC spec 4.15) ----------
//
// A project's access is read and written per project: who holds which role
// in it, who reaches it, whom the caller may add, and one member's or team's
// assignments in it set in one request. The routes read and write the same
// role assignments as /role-assignments; nothing else is stored.
//
// Reading needs members.read, members.update, projects.manageAccess covering
// the project or project.manageAccess in it. Writing and the candidates need
// one of the last three. Every write is bounded like CreateRoleAssignment
// (models.Access.CanChange).

// ProjectAccessPrincipal is one member's or team's assignments in a project.
type ProjectAccessPrincipal struct {
	Principal RoleAssignmentPrincipal `json:"principal"`
	// Version changes whenever the principal's assignments in the project
	// change. SetProjectAccessRequest sends it back. It is empty when the
	// principal holds nothing in the project.
	Version     string                  `json:"version"`
	Assignments []RoleAssignmentDetails `json:"assignments"`
}

// ProjectAccessGrantable is what the caller may hand out in the project.
type ProjectAccessGrantable struct {
	// RoleIds are the active project roles the caller may grant here
	// (models.Access.CanGrant). Owner-only roles are listed for owners only.
	RoleIds []string `json:"roleIds"`
	// AllResources tells whether the caller may select the whole project;
	// otherwise only resources inside their own (GET
	// /projects/{id}/access-resources lists them).
	AllResources bool `json:"allResources"`
	// CanManage tells whether the caller may change access in the project.
	// RoleIds is empty when they may not.
	CanManage bool `json:"canManage"`
}

// GetProjectAccess lists the project's assignments per member and team, and
// what the caller may grant there. Pages are ordered by principal type, then
// id.
// @Router /projects/{id}/access [get]
type GetProjectAccessRequest struct{}
type GetProjectAccessResponse struct {
	Principals []ProjectAccessPrincipal `json:"principals"`
	Grantable  ProjectAccessGrantable   `json:"grantable"`
	Pagination *CursorPagination        `json:"pagination,omitempty"`
}
type GetProjectAccessSuccessResponse struct {
	SuccessResponse
	Data GetProjectAccessResponse `json:"data"`
}
type GetProjectAccessErrorResponse struct {
	ErrorResponse
}

// ProjectAccessTeamGrant is a team's assignments a member receives.
type ProjectAccessTeamGrant struct {
	Team        RoleAssignmentPrincipal `json:"team"`
	Assignments []RoleAssignmentDetails `json:"assignments"`
}

// ProjectAccessMember is a member who reaches the project and how: their own
// assignments and those of their teams that cover the project, organisation
// roles included. Only effectively active assignments of effectively active
// members count.
type ProjectAccessMember struct {
	User   RoleAssignmentPrincipal  `json:"user"`
	Direct []RoleAssignmentDetails  `json:"direct"`
	Teams  []ProjectAccessTeamGrant `json:"teams"`
}

// GetProjectAccessMembers lists everyone who reaches the project
// (models.ResolveAccess). Owners reach every project and are listed apart,
// in full, on the first page only. Members are paged by user id.
// @Router /projects/{id}/access/members [get]
type GetProjectAccessMembersRequest struct{}
type GetProjectAccessMembersResponse struct {
	Owners     []RoleAssignmentPrincipal `json:"owners"`
	Members    []ProjectAccessMember     `json:"members"`
	Pagination *CursorPagination         `json:"pagination,omitempty"`
}
type GetProjectAccessMembersSuccessResponse struct {
	SuccessResponse
	Data GetProjectAccessMembersResponse `json:"data"`
}
type GetProjectAccessMembersErrorResponse struct {
	ErrorResponse
}

// ProjectAccessCandidate is a member or team the caller may give access to:
// identity only.
type ProjectAccessCandidate struct {
	Principal RoleAssignmentPrincipal `json:"principal"`
	// Source is set for teams: manual, or the provider that syncs it.
	Source models.TeamSource `json:"source,omitempty"`
}

// GetProjectAccessCandidates lists the organisation's effectively active
// members and its teams, owners and the caller left out, matching search
// (any word of the name or email, ignoring case). Members come first, then
// teams, each ordered by id.
// @Router /projects/{id}/access/candidates [get]
type GetProjectAccessCandidatesRequest struct{}
type GetProjectAccessCandidatesResponse struct {
	Candidates []ProjectAccessCandidate `json:"candidates"`
	Pagination *CursorPagination        `json:"pagination,omitempty"`
}
type GetProjectAccessCandidatesSuccessResponse struct {
	SuccessResponse
	Data GetProjectAccessCandidatesResponse `json:"data"`
}
type GetProjectAccessCandidatesErrorResponse struct {
	ErrorResponse
}

// ProjectAccessAssignment is one role a principal holds in the project, with
// its resources. A field left out means none: no sites, no time window, no
// expiry.
type ProjectAccessAssignment struct {
	RoleId     primitive.ObjectID   `json:"roleId"`
	Resources  models.Selection     `json:"resources"`
	SiteIds    []primitive.ObjectID `json:"siteIds,omitempty"`
	GroupIds   []primitive.ObjectID `json:"groupIds,omitempty"`
	DeviceKeys []string             `json:"deviceKeys,omitempty"`
	TimeWindow *models.TimeWindow   `json:"timeWindow,omitempty"`
	ExpiresAt  *time.Time           `json:"expiresAt,omitempty"`
}

// Scope is the assignment's scope in the project.
func (a ProjectAccessAssignment) Scope(projectId primitive.ObjectID) models.RoleAssignmentScope {
	return models.RoleAssignmentScope{
		Level:      models.LevelProject,
		ProjectId:  projectId,
		Resources:  a.Resources,
		SiteIds:    a.SiteIds,
		GroupIds:   a.GroupIds,
		DeviceKeys: a.DeviceKeys,
	}
}

// SetProjectAccess sets one member's or team's assignments in the project to
// exactly Assignments, matched to the stored ones by role: a matched one is
// changed in place, an unmatched entry is created, a stored one without an
// entry is deleted. All of it is written in one transaction, or nothing is.
// Version is what GetProjectAccess returned for the principal (empty when it
// held nothing); when it no longer matches, the answer is 409
// RoleAssignmentChanged.
// @Router /projects/{id}/access/{principalType}/{principalId} [put]
type SetProjectAccessRequest struct {
	Version     string                    `json:"version"`
	Assignments []ProjectAccessAssignment `json:"assignments"`
}
type SetProjectAccessResponse struct {
	Access ProjectAccessPrincipal `json:"access"`
}
type SetProjectAccessSuccessResponse struct {
	SuccessResponse
	Data SetProjectAccessResponse `json:"data"`
}
type SetProjectAccessErrorResponse struct {
	ErrorResponse
}

// DeleteProjectAccess removes every assignment of the member or team in the
// project: SetProjectAccess with no assignments. The version is a query
// parameter.
// @Router /projects/{id}/access/{principalType}/{principalId} [delete]
type DeleteProjectAccessRequest struct{}
type DeleteProjectAccessResponse struct{}
type DeleteProjectAccessSuccessResponse struct {
	SuccessResponse
	Data DeleteProjectAccessResponse `json:"data"`
}
type DeleteProjectAccessErrorResponse struct {
	ErrorResponse
}
