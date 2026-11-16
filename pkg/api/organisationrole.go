package api

import (
	"time"

	"github.com/uug-ai/models/pkg/models"
)

// OrganisationRoleStatus represents specific status codes for the
// organisation's RBAC roles (organisation_roles), as opposed to the legacy
// custom roles.
type OrganisationRoleStatus string

const (
	OrganisationRoleBindingFailed    OrganisationRoleStatus = "organisation_role_binding_failed"
	OrganisationRoleValidationFailed OrganisationRoleStatus = "organisation_role_validation_failed"
	OrganisationRoleFound            OrganisationRoleStatus = "organisation_role_found"
	OrganisationRoleNotFound         OrganisationRoleStatus = "organisation_role_not_found"
	OrganisationRoleGetAllSuccess    OrganisationRoleStatus = "organisation_role_get_all_success"
	OrganisationRoleGetAllFailed     OrganisationRoleStatus = "organisation_role_get_all_failed"
	OrganisationRoleCreateSuccess    OrganisationRoleStatus = "organisation_role_create_success"
	OrganisationRoleCreateFailed     OrganisationRoleStatus = "organisation_role_create_failed"
	OrganisationRoleUpdateSuccess    OrganisationRoleStatus = "organisation_role_update_success"
	OrganisationRoleUpdateFailed     OrganisationRoleStatus = "organisation_role_update_failed"
	OrganisationRoleDeleteSuccess    OrganisationRoleStatus = "organisation_role_delete_success"
	OrganisationRoleDeleteFailed     OrganisationRoleStatus = "organisation_role_delete_failed"
	// OrganisationRoleNameExists: another role of the organisation has the
	// same name, ignoring case.
	OrganisationRoleNameExists OrganisationRoleStatus = "organisation_role_name_exists"
	// OrganisationRoleReadOnly: system and legacy roles are defined by the
	// platform and cannot be changed or deleted.
	OrganisationRoleReadOnly OrganisationRoleStatus = "organisation_role_read_only"
	// OrganisationRoleInUse: the role is still assigned to a user or team, or
	// named by a pending invitation, so it cannot be deleted.
	OrganisationRoleInUse OrganisationRoleStatus = "organisation_role_in_use"
	// OrganisationRoleOwnerOnly: the role holds an owner-only permission,
	// before or after the change, so only an owner handles it (spec 4.5).
	OrganisationRoleOwnerOnly OrganisationRoleStatus = "organisation_role_owner_only"
	// OrganisationRoleSelf: nobody but an owner changes or deletes a role they
	// hold, directly or through a team (spec 4.5).
	OrganisationRoleSelf OrganisationRoleStatus = "organisation_role_self"
	// OrganisationRoleLegacyMode: the organisation has not been cut over to
	// RBAC, so its roles are still the legacy custom roles (spec 7.3).
	OrganisationRoleLegacyMode OrganisationRoleStatus = "organisation_role_legacy_mode"
)

// String returns the string representation of the organisation role status.
func (s OrganisationRoleStatus) String() string {
	return string(s)
}

// Translate returns the translated string representation of the
// organisation role status in the specified language, falling back to
// English and then to the raw value.
func (s OrganisationRoleStatus) Translate(lang string) string {
	translations := map[string]map[OrganisationRoleStatus]string{
		"en": {
			OrganisationRoleBindingFailed:    "Role binding failed",
			OrganisationRoleValidationFailed: "Role validation failed",
			OrganisationRoleFound:            "Role found",
			OrganisationRoleNotFound:         "Role not found",
			OrganisationRoleGetAllSuccess:    "Roles retrieved successfully",
			OrganisationRoleGetAllFailed:     "Failed to retrieve roles",
			OrganisationRoleCreateSuccess:    "Role created successfully",
			OrganisationRoleCreateFailed:     "Failed to create role",
			OrganisationRoleUpdateSuccess:    "Role updated successfully",
			OrganisationRoleUpdateFailed:     "Failed to update role",
			OrganisationRoleDeleteSuccess:    "Role deleted successfully",
			OrganisationRoleDeleteFailed:     "Failed to delete role",
			OrganisationRoleNameExists:       "A role with this name already exists",
			OrganisationRoleReadOnly:         "Built-in roles cannot be changed",
			OrganisationRoleInUse:            "This role is still assigned; remove its assignments and invitations first",
			OrganisationRoleOwnerOnly:        "Only an owner can manage this role",
			OrganisationRoleSelf:             "You cannot change a role you hold",
			OrganisationRoleLegacyMode:       "Access for this organisation is still managed through legacy roles",
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

// ---------- Request / Response DTOs (organisation role domain) ----------
//
// Roles belong to the caller's current organisation; no request names it.
// They are only served for organisations in rbac mode (otherwise
// organisation_role_legacy_mode); legacy organisations keep their custom
// roles on /roles.
//
// Reading needs roles.read or any permission that hands out access
// (members.read, members.invite, members.update, teams.update,
// projects.manageAccess, project.manageAccess), since choosing a role is part
// of granting it. Creating needs roles.create, changing roles.update and
// deleting roles.delete; only custom roles change, and
// models.Access.CanManageRole decides (owner-only permissions, never a role
// the caller holds). System and legacy roles are returned with the platform's
// current definition (models.Role.WithCurrentDefinition).

// OrganisationRoleFilter narrows a role listing. Set fields combine.
type OrganisationRoleFilter struct {
	Kind         *models.RoleKind `json:"kind,omitempty" bson:"kind,omitempty"`
	AssignableAt *models.Level    `json:"assignableAt,omitempty" bson:"assignableAt,omitempty"`
	// Name matches roles whose name contains it, ignoring case.
	Name *string `json:"name,omitempty" bson:"name,omitempty"`
}

// OrganisationRoleUsage counts what refers to a role: a role is deleted only
// when every count is zero.
type OrganisationRoleUsage struct {
	UserAssignments    int64 `json:"userAssignments"`    // Assignments to users, active or not
	TeamAssignments    int64 `json:"teamAssignments"`    // Assignments to teams, active or not
	PendingInvitations int64 `json:"pendingInvitations"` // Pending invitations naming the role
}

// InUse reports whether anything still refers to the role.
func (u OrganisationRoleUsage) InUse() bool {
	return u.UserAssignments > 0 || u.TeamAssignments > 0 || u.PendingInvitations > 0
}

// OrganisationRoleSummary is a role in a listing, with its usage.
type OrganisationRoleSummary struct {
	Role  models.Role           `json:"role"`
	Usage OrganisationRoleUsage `json:"usage"`
}

// OrganisationRoleInvitation is a pending invitation that grants the role
// when accepted.
type OrganisationRoleInvitation struct {
	InvitationId string                     `json:"invitationId"`
	Email        string                     `json:"email"`
	Scope        models.RoleAssignmentScope `json:"scope"`
	ExpiresAt    time.Time                  `json:"expiresAt"`
}

// GetOrganisationRoles lists the organisation's roles: Owner, the system and
// legacy roles, and its custom roles. Organisations have few roles, so the
// listing is not paged.
// @Router /organisation-roles [get]
type GetOrganisationRolesRequest struct {
	Filter *OrganisationRoleFilter `json:"filter,omitempty" bson:"filter,omitempty"`
}
type GetOrganisationRolesResponse struct {
	Roles []OrganisationRoleSummary `json:"roles"`
}
type GetOrganisationRolesSuccessResponse struct {
	SuccessResponse
	Data GetOrganisationRolesResponse `json:"data"`
}
type GetOrganisationRolesErrorResponse struct {
	ErrorResponse
}

// GetOrganisationRole returns one role with who holds it and where: its
// assignments the caller may read, and the pending invitations naming it.
// @Router /organisation-roles/{id} [get]
type GetOrganisationRoleRequest struct{}
type GetOrganisationRoleResponse struct {
	Role        models.Role                  `json:"role"`
	Usage       OrganisationRoleUsage        `json:"usage"`
	Assignments []RoleAssignmentDetails      `json:"assignments"`
	Invitations []OrganisationRoleInvitation `json:"invitations"`
}
type GetOrganisationRoleSuccessResponse struct {
	SuccessResponse
	Data GetOrganisationRoleResponse `json:"data"`
}
type GetOrganisationRoleErrorResponse struct {
	ErrorResponse
}

// CreateOrganisationRole creates a custom role granting exactly the listed
// permissions, all of the level it is assigned at (models.NewCustomRole).
// The server sets the id, organisation, kind and audit. Cloning a role is
// creating one with its permissions.
// @Router /organisation-roles [post]
type CreateOrganisationRoleRequest struct {
	Name         string              `json:"name"`
	Description  string              `json:"description,omitempty"`
	AssignableAt models.Level        `json:"assignableAt"`
	Permissions  []models.Permission `json:"permissions"`
}
type CreateOrganisationRoleResponse struct {
	Role models.Role `json:"role"`
}
type CreateOrganisationRoleSuccessResponse struct {
	SuccessResponse
	Data CreateOrganisationRoleResponse `json:"data"`
}
type CreateOrganisationRoleErrorResponse struct {
	ErrorResponse
}

// UpdateOrganisationRole changes a custom role's name, description or
// permissions; only the fields set on the patch change, and the change
// applies at once to everyone the role is assigned to.
// @Router /organisation-roles/{id} [patch]
type UpdateOrganisationRoleRequest struct {
	Role models.RolePatch `json:"role"`
}
type UpdateOrganisationRoleResponse struct {
	Role models.Role `json:"role"`
}
type UpdateOrganisationRoleSuccessResponse struct {
	SuccessResponse
	Data UpdateOrganisationRoleResponse `json:"data"`
}
type UpdateOrganisationRoleErrorResponse struct {
	ErrorResponse
}

// DeleteOrganisationRole deletes a custom role nothing refers to
// (OrganisationRoleUsage.InUse is false); otherwise
// organisation_role_in_use.
// @Router /organisation-roles/{id} [delete]
type DeleteOrganisationRoleRequest struct{}
type DeleteOrganisationRoleResponse struct{}
type DeleteOrganisationRoleSuccessResponse struct {
	SuccessResponse
	Data DeleteOrganisationRoleResponse `json:"data"`
}
type DeleteOrganisationRoleErrorResponse struct {
	ErrorResponse
}
