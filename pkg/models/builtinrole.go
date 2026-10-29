package models

import "slices"

// Built-in roles are defined by the platform. Every organisation stores its own
// read-only copy, upserted by (organisationId, key) and referenced by its stored
// id, so the definitions here are the only source for what a copy contains.
//
// For now they are the Owner role and the legacy roles, which reproduce the
// legacy guest, editor and admin tiers so migrated users keep their access.
// The legacy sets follow the RBAC permission matrix: what each tier can do in
// the product today. Where the API allows a tier more than the product does,
// the set follows the product and a comment names the matrix decision that
// confirms it. Actions the product has no screen for (deleting recordings,
// creating IO, detections, moderating comments) are not granted either.
//
// Relationship rules are not role permissions; the resolver applies them next
// to roles. A creator may update and delete what they created when they hold
// the domain's create permission, so a tier that may only change its own sites,
// groups, videowalls or alerts holds create here, not update or delete.

// legacyGuestPermissions is the project part of the legacy guest tier.
//
// Not granted although the API allows it: device edit, delete and device
// permission limits, and PTZ presets (D1); site, group, floor plan and label
// writes and IO delete (D2); workflow writes and run results (D4); alerts and
// channels (D5); deleting recordings, creating IO and detections, which have
// no screen. The guest tier has no PTZ (level 1) and no remote config.
// Videowalls and private cases are visible through their assigned-user and
// participant rules only.
var legacyGuestPermissions = []Permission{
	PermissionDevicesRead,
	PermissionLiveStream,
	PermissionIORead,
	PermissionIOTrigger,
	PermissionSitesRead,
	PermissionGroupsRead,
	PermissionFloorPlansRead,
	PermissionMediaRead,
	PermissionMediaUpdate,
	PermissionMediaExport,
	PermissionMediaShare,
	PermissionMediaRedact,
	PermissionMediaForward,
	PermissionDaysRead,
	PermissionMarkersRead,
	PermissionMarkersWrite,
	PermissionCasesRead,
	PermissionCasesCreate,
	PermissionCasesUpdate,
	PermissionCasesShare,
	PermissionCasesExport,
	PermissionCasesRunWorkflow,
	PermissionWorkflowsRead,
	PermissionWorkflowRunsCreate,
	PermissionWorkflowRunsRead,
	PermissionNotificationsRead,
	PermissionNotificationsUpdate,
	PermissionLabelsRead,
	PermissionAnalyticsRead,
}

// legacyEditorPermissions is the project part of the legacy editor tier: the
// guest set plus remote config (but not the recording mode), every PTZ grade,
// floor plan editing and creating videowalls (and changing their own, by the
// creator rule).
//
// Not granted although the API allows it: as for guests, except PTZ and floor
// plans (D1, D2, D4, D5). Editors may not assign users to videowalls.
var legacyEditorPermissions = concatPermissions(legacyGuestPermissions, []Permission{
	PermissionDevicesConfigure,
	PermissionPTZPresets,
	PermissionFloorPlansCreate,
	PermissionFloorPlansUpdate,
	PermissionFloorPlansDelete,
	PermissionVideowallsCreate,
})

// legacyAdminPermissions is the project part of the legacy admin tier: the
// editor set plus device edits, sites, groups, IO, videowalls, case
// membership, alerts, labels and project administration.
//
// Admins delete their own sites, update their own groups, delete their own
// groups (D3) and delete their own alerts by the creator rule. Not granted
// although the API allows it: device delete and device permission limits,
// which only owners have (D1); workflow writes and run results (D4); channels
// (D5). Device recording mode and logs are owner only, as in the product.
var legacyAdminPermissions = concatPermissions(legacyEditorPermissions, []Permission{
	PermissionDevicesUpdate,
	PermissionSitesCreate,
	PermissionSitesUpdate,
	PermissionGroupsCreate,
	PermissionIODelete,
	PermissionVideowallsRead,
	PermissionVideowallsUpdate,
	PermissionVideowallsDelete,
	PermissionVideowallsAssign,
	PermissionCasesDelete,
	PermissionCasesManageMembership,
	PermissionAlertsRead,
	PermissionAlertsCreate,
	PermissionAlertsUpdate,
	PermissionAlertsManagePrivacy,
	PermissionLabelsCreate,
	PermissionLabelsUpdate,
	PermissionLabelsDelete,
	PermissionProjectUpdate,
	PermissionProjectManageCredentials,
	PermissionProjectManageAccess,
})

// legacyMemberOrganisationPermissions is the organisation part of the legacy
// guest and editor tiers. Not granted although the API allows it: reading
// roles (D7), and billing and storage (D8).
var legacyMemberOrganisationPermissions = []Permission{
	PermissionOrganisationRead,
	PermissionMembersRead,
}

// legacyAdminOrganisationPermissions is the organisation part of the legacy
// admin tier. Deleting roles stays with owners (D7), as do billing, storage,
// audit and device system tags.
var legacyAdminOrganisationPermissions = concatPermissions(legacyMemberOrganisationPermissions, []Permission{
	PermissionOrganisationUpdate,
	PermissionMembersInvite,
	PermissionMembersUpdate,
	PermissionMembersRemove,
	PermissionRolesRead,
	PermissionRolesCreate,
	PermissionRolesUpdate,
	PermissionProjectsCreate,
	PermissionProjectsDelete,
	PermissionProjectsManageAccess,
})

// BuiltInRoles returns the platform-defined roles in stable order, with their
// permissions in catalog order. The roles carry no Id, OrganisationId or
// Audit; the writer sets them on each organisation's copy.
func BuiltInRoles() []Role {
	return []Role{
		{
			Name:         "Owner",
			Description:  "Every permission in every current and future project.",
			Kind:         RoleKindSystem,
			Key:          RoleKeyOwner,
			AssignableAt: LevelOrganisation,
			Grants:       SelectionAll,
			IsActive:     true,
		},
		legacyRole(RoleKeyLegacyGuest, BaseRoleGuest, LevelProject, "Guest (legacy)",
			"The legacy guest tier inside a project.", legacyGuestPermissions),
		legacyRole(RoleKeyLegacyEditor, BaseRoleEditor, LevelProject, "Editor (legacy)",
			"The legacy editor tier inside a project.", legacyEditorPermissions),
		legacyRole(RoleKeyLegacyAdmin, BaseRoleAdmin, LevelProject, "Admin (legacy)",
			"The legacy admin tier inside a project.", legacyAdminPermissions),
		legacyRole(RoleKeyLegacyMemberOrganisation, "", LevelOrganisation, "Member (legacy), organisation",
			"What legacy guests and editors could see of the organisation.", legacyMemberOrganisationPermissions),
		legacyRole(RoleKeyLegacyAdminOrganisation, BaseRoleAdmin, LevelOrganisation, "Admin (legacy), organisation",
			"The legacy admin tier's organisation administration.", legacyAdminOrganisationPermissions),
	}
}

// BuiltInRole returns the definition of one built-in role.
func BuiltInRole(key RoleKey) (Role, bool) {
	for _, role := range BuiltInRoles() {
		if role.Key == key {
			return role, true
		}
	}
	return Role{}, false
}

func legacyRole(key RoleKey, baseRole BaseRole, level Level, name, description string, permissions []Permission) Role {
	role := Role{
		Name:         name,
		Description:  description,
		Kind:         RoleKindLegacy,
		Key:          key,
		AssignableAt: level,
		Grants:       SelectionSelected,
		Permissions:  catalogOrder(permissions),
		IsActive:     true,
	}
	if baseRole != "" {
		role.Origin = &RoleOrigin{BaseRole: baseRole}
	}
	return role
}

func concatPermissions(base, extra []Permission) []Permission {
	return append(slices.Clone(base), extra...)
}

// catalogOrder returns a copy of the permissions sorted in catalog order.
func catalogOrder(permissions []Permission) []Permission {
	sorted := slices.Clone(permissions)
	slices.SortFunc(sorted, func(a, b Permission) int {
		return permissionIndex[a] - permissionIndex[b]
	})
	return sorted
}
