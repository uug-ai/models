package api

import (
	"time"

	"github.com/uug-ai/models/pkg/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RoleAssignmentStatus represents specific status codes for role assignment
// operations.
type RoleAssignmentStatus string

const (
	RoleAssignmentBindingFailed    RoleAssignmentStatus = "role_assignment_binding_failed"
	RoleAssignmentValidationFailed RoleAssignmentStatus = "role_assignment_validation_failed"
	RoleAssignmentFound            RoleAssignmentStatus = "role_assignment_found"
	RoleAssignmentNotFound         RoleAssignmentStatus = "role_assignment_not_found"
	RoleAssignmentGetAllSuccess    RoleAssignmentStatus = "role_assignment_get_all_success"
	RoleAssignmentGetAllFailed     RoleAssignmentStatus = "role_assignment_get_all_failed"
	RoleAssignmentCreateSuccess    RoleAssignmentStatus = "role_assignment_create_success"
	RoleAssignmentCreateFailed     RoleAssignmentStatus = "role_assignment_create_failed"
	RoleAssignmentUpdateSuccess    RoleAssignmentStatus = "role_assignment_update_success"
	RoleAssignmentUpdateFailed     RoleAssignmentStatus = "role_assignment_update_failed"
	RoleAssignmentDeleteSuccess    RoleAssignmentStatus = "role_assignment_delete_success"
	RoleAssignmentDeleteFailed     RoleAssignmentStatus = "role_assignment_delete_failed"
	// RoleAssignmentExists: the principal already holds the role in that
	// project (or at organisation level). Widen that assignment instead.
	RoleAssignmentExists RoleAssignmentStatus = "role_assignment_exists"
	// RoleAssignmentRoleNotFound: the role is not one of the organisation's
	// active roles.
	RoleAssignmentRoleNotFound RoleAssignmentStatus = "role_assignment_role_not_found"
	// RoleAssignmentPrincipalNotFound: the user is not an effectively active
	// member of the organisation, or the team does not exist in it.
	RoleAssignmentPrincipalNotFound RoleAssignmentStatus = "role_assignment_principal_not_found"
	// RoleAssignmentResourceNotFound: a selected project, site or group is not
	// the organisation's.
	RoleAssignmentResourceNotFound RoleAssignmentStatus = "role_assignment_resource_not_found"
	// RoleAssignmentSelf: callers never create, change or remove an assignment
	// that affects themselves, directly or through a team; an owner may only
	// step down from Owner (spec 4.5, 4.10).
	RoleAssignmentSelf RoleAssignmentStatus = "role_assignment_self"
	// RoleAssignmentExceedsAccess: the selected resources go beyond the
	// caller's own scope in the project; only a caller with nothing but
	// project.manageAccess there is bounded this way (spec 4.5).
	RoleAssignmentExceedsAccess RoleAssignmentStatus = "role_assignment_exceeds_access"
	// RoleAssignmentOwnerOnly: the role is Owner or holds an owner-only
	// permission, so only an owner grants, changes or removes it (spec 4.5).
	RoleAssignmentOwnerOnly RoleAssignmentStatus = "role_assignment_owner_only"
	// RoleAssignmentLastOwner: the change would leave the organisation without
	// an effectively active owner.
	RoleAssignmentLastOwner RoleAssignmentStatus = "role_assignment_last_owner"
	// RoleAssignmentLegacyMode: the organisation has not been cut over to RBAC,
	// so access is still changed through the legacy role and scope fields
	// (spec 7.3).
	RoleAssignmentLegacyMode RoleAssignmentStatus = "role_assignment_legacy_mode"
)

// String returns the string representation of the role assignment status.
func (s RoleAssignmentStatus) String() string {
	return string(s)
}

// Translate returns the translated string representation of the role
// assignment status in the specified language, falling back to English and
// then to the raw value.
func (s RoleAssignmentStatus) Translate(lang string) string {
	translations := map[string]map[RoleAssignmentStatus]string{
		"en": {
			RoleAssignmentBindingFailed:     "Role assignment binding failed",
			RoleAssignmentValidationFailed:  "Role assignment validation failed",
			RoleAssignmentFound:             "Role assignment found",
			RoleAssignmentNotFound:          "Role assignment not found",
			RoleAssignmentGetAllSuccess:     "Role assignments retrieved successfully",
			RoleAssignmentGetAllFailed:      "Failed to retrieve role assignments",
			RoleAssignmentCreateSuccess:     "Role assigned successfully",
			RoleAssignmentCreateFailed:      "Failed to assign role",
			RoleAssignmentUpdateSuccess:     "Role assignment updated successfully",
			RoleAssignmentUpdateFailed:      "Failed to update role assignment",
			RoleAssignmentDeleteSuccess:     "Role assignment removed successfully",
			RoleAssignmentDeleteFailed:      "Failed to remove role assignment",
			RoleAssignmentExists:            "This role is already assigned here; change that assignment instead",
			RoleAssignmentRoleNotFound:      "Role not found",
			RoleAssignmentPrincipalNotFound: "User or team not found in this organisation",
			RoleAssignmentResourceNotFound:  "A selected project, site or group was not found",
			RoleAssignmentSelf:              "You cannot change your own access",
			RoleAssignmentExceedsAccess:     "You cannot grant resources outside your own access",
			RoleAssignmentOwnerOnly:         "Only an owner can grant this role",
			RoleAssignmentLastOwner:         "An organisation needs at least one owner",
			RoleAssignmentLegacyMode:        "Access for this organisation is still managed through legacy roles",
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

// ---------- Request / Response DTOs (role assignment domain) ----------
//
// Role assignments grant a role to a user or a team (principal) within an
// explicit scope, in the caller's current organisation; no request names it.
// Project assignments need project.manageAccess in that project or
// projects.manageAccess covering it; organisation assignments need
// members.update; reading needs the same or members.read. The caller need not
// hold the role's permissions, but only an owner grants Owner or a role
// holding an owner-only permission, a caller with only project.manageAccess
// selects resources inside their own scope, and no write affects the caller
// (models.Access.CanChange, spec 4.5).

// RoleAssignmentFilter narrows an assignment listing. Set fields combine.
type RoleAssignmentFilter struct {
	PrincipalType *models.PrincipalType    `json:"principalType,omitempty" bson:"principalType,omitempty"`
	PrincipalId   *string                  `json:"principalId,omitempty" bson:"principalId,omitempty"`
	RoleId        *string                  `json:"roleId,omitempty" bson:"roleId,omitempty"`
	Level         *models.Level            `json:"level,omitempty" bson:"level,omitempty"`
	ProjectId     *string                  `json:"projectId,omitempty" bson:"projectId,omitempty"` // Project assignments in this project
	Source        *models.AssignmentSource `json:"source,omitempty" bson:"source,omitempty"`
	// IncludeTeams, with a user PrincipalId, also lists the assignments of
	// the teams the user belongs to: everything that grants the user access.
	IncludeTeams bool `json:"includeTeams,omitempty" bson:"includeTeams,omitempty"`
}

// RoleAssignmentPrincipal is who an assignment grants to, for display:
// identity only, never credentials.
type RoleAssignmentPrincipal struct {
	Type models.PrincipalType `json:"type"`
	Id   string               `json:"id"`
	// Name is a user's username or a team's name.
	Name string `json:"name"`
	// Email is set for users.
	Email string `json:"email,omitempty"`
	// MemberCount is set for teams.
	MemberCount int64 `json:"memberCount,omitempty"`
}

// RoleAssignmentDetails is an assignment with its role and principal.
type RoleAssignmentDetails struct {
	Assignment models.RoleAssignment   `json:"assignment"`
	Role       models.Role             `json:"role"`
	Principal  RoleAssignmentPrincipal `json:"principal"`
	// EffectivelyActive is RoleAssignment.IsEffectivelyActive at response
	// time: enabled, not expired and inside its time window.
	EffectivelyActive bool `json:"effectivelyActive"`
}

// GetRoleAssignments lists the organisation's assignments the caller may see.
// @Router /role-assignments [get]
type GetRoleAssignmentsRequest struct {
	Filter     *RoleAssignmentFilter `json:"filter,omitempty" bson:"filter,omitempty"`
	Pagination *CursorPagination     `json:"pagination,omitempty" bson:"pagination,omitempty"`
}
type GetRoleAssignmentsResponse struct {
	RoleAssignments []RoleAssignmentDetails `json:"roleAssignments"`
	Pagination      *CursorPagination       `json:"pagination,omitempty"`
}
type GetRoleAssignmentsSuccessResponse struct {
	SuccessResponse
	Data GetRoleAssignmentsResponse `json:"data"`
}
type GetRoleAssignmentsErrorResponse struct {
	ErrorResponse
}

// GetRoleAssignment returns one assignment.
// @Router /role-assignments/{id} [get]
type GetRoleAssignmentRequest struct{}
type GetRoleAssignmentResponse struct {
	RoleAssignment RoleAssignmentDetails `json:"roleAssignment"`
}
type GetRoleAssignmentSuccessResponse struct {
	SuccessResponse
	Data GetRoleAssignmentResponse `json:"data"`
}
type GetRoleAssignmentErrorResponse struct {
	ErrorResponse
}

// CreateRoleAssignment grants a role to a user or team. The server sets the
// id, organisation, source (manual), audit and IsActive (true). When the
// principal already holds the role in that project it answers
// RoleAssignmentExists; widen that assignment with UpdateRoleAssignment.
// @Router /role-assignments [post]
type CreateRoleAssignmentRequest struct {
	Principal  models.Principal           `json:"principal"`
	RoleId     primitive.ObjectID         `json:"roleId"`
	Scope      models.RoleAssignmentScope `json:"scope"`
	TimeWindow *models.TimeWindow         `json:"timeWindow,omitempty"`
	ExpiresAt  *time.Time                 `json:"expiresAt,omitempty"`
}
type CreateRoleAssignmentResponse struct {
	RoleAssignment RoleAssignmentDetails `json:"roleAssignment"`
}
type CreateRoleAssignmentSuccessResponse struct {
	SuccessResponse
	Data CreateRoleAssignmentResponse `json:"data"`
}
type CreateRoleAssignmentErrorResponse struct {
	ErrorResponse
}

// UpdateRoleAssignment changes an assignment's resources, time window, expiry
// or enabled state (models.RoleAssignmentPatch). The principal, role, level
// and project never change. Narrowing is bounded like granting.
// @Router /role-assignments/{id} [patch]
type UpdateRoleAssignmentRequest struct {
	RoleAssignment models.RoleAssignmentPatch `json:"roleAssignment"`
}
type UpdateRoleAssignmentResponse struct {
	RoleAssignment RoleAssignmentDetails `json:"roleAssignment"`
}
type UpdateRoleAssignmentSuccessResponse struct {
	SuccessResponse
	Data UpdateRoleAssignmentResponse `json:"data"`
}
type UpdateRoleAssignmentErrorResponse struct {
	ErrorResponse
}

// DeleteRoleAssignment removes an assignment. Removing the organisation's
// last effectively active Owner assignment is refused.
// @Router /role-assignments/{id} [delete]
type DeleteRoleAssignmentRequest struct{}
type DeleteRoleAssignmentResponse struct{}
type DeleteRoleAssignmentSuccessResponse struct {
	SuccessResponse
	Data DeleteRoleAssignmentResponse `json:"data"`
}
type DeleteRoleAssignmentErrorResponse struct {
	ErrorResponse
}
