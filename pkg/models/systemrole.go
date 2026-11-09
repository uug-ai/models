package models

// System roles are the platform's designed roles (spec 4.1). Unlike the
// legacy roles they are not bound by legacy parity: migrated users receive
// legacy roles, and system roles are what new access is granted with. Each
// project role nests in the next: Viewer ⊂ Editor ⊂ Project admin. The
// organisation roles grant no project permissions, so they never widen what a
// holder can see. Project admin and Billing admin hold owner-only permissions
// (PermissionDefinition.OwnerOnly), so only owners grant them; an
// Organisation manager grants Viewer, Editor and custom roles without
// owner-only permissions in every project.

// viewerPermissions: see everything in the assignment's scope, change nothing.
// No PTZ (moving a camera affects everyone watching it) and no export or share
// (data leaving the system).
var viewerPermissions = []Permission{
	PermissionDevicesRead, PermissionLiveStream,
	PermissionMediaRead, PermissionDaysRead, PermissionMarkersRead, PermissionWorkflowRunsRead,
	PermissionSitesRead, PermissionGroupsRead, PermissionFloorPlansRead, PermissionVideowallsRead, PermissionIORead, PermissionLabelsRead,
	PermissionNotificationsRead, PermissionAnalyticsRead, PermissionCasesRead,
	PermissionDashboardPage, PermissionLatestEventsPage, PermissionLivePage, PermissionVideowallsPage,
	PermissionMediaPage, PermissionDaysPage, PermissionAnalyticsPage, PermissionWatchlistPage,
	PermissionCasesPage, PermissionIOPage, PermissionDevicesPage, PermissionSitesPage, PermissionFloorPlansPage, PermissionGroupsPage,
}

// editorPermissions adds day-to-day operation to Viewer: camera control,
// working with recordings and cases, and creating and updating the project's
// structure. No deletes (a creator still deletes what they created, a
// relationship rule), no device configuration, recording mode or logs, no
// forwarding, case moderation or membership, channels, project settings or
// access.
var editorPermissions = append(append([]Permission{}, viewerPermissions...),
	PermissionPTZPresets, PermissionIOTrigger, PermissionDevicesUpdate,
	PermissionMediaUpdate, PermissionMediaExport, PermissionMediaShare, PermissionMediaRedact,
	PermissionMarkersWrite, PermissionNotificationsUpdate,
	PermissionCasesCreate, PermissionCasesUpdate, PermissionCasesShare, PermissionCasesExport, PermissionCasesRunWorkflow,
	PermissionWorkflowsRead, PermissionWorkflowRunsCreate,
	PermissionSitesCreate, PermissionSitesUpdate, PermissionGroupsCreate, PermissionGroupsUpdate,
	PermissionFloorPlansCreate, PermissionFloorPlansUpdate, PermissionVideowallsCreate, PermissionVideowallsUpdate,
	PermissionLabelsCreate, PermissionLabelsUpdate,
	PermissionAlertsRead, PermissionAlertsCreate, PermissionAlertsUpdate,
	PermissionDetectionsRead,
	PermissionAlertsPage, PermissionWorkflowsPage, PermissionSettingsPage,
)

// organisationManagerPermissions administer members, roles, teams and the
// organisation's settings, and grant access in every project without seeing
// its resources. Billing, storage, audit and device system tags stay with
// owners.
var organisationManagerPermissions = []Permission{
	PermissionOrganisationRead, PermissionOrganisationUpdate,
	PermissionMembersRead, PermissionMembersInvite, PermissionMembersUpdate, PermissionMembersRemove,
	PermissionRolesRead, PermissionRolesCreate, PermissionRolesUpdate, PermissionRolesDelete,
	PermissionTeamsRead, PermissionTeamsCreate, PermissionTeamsUpdate, PermissionTeamsDelete,
	PermissionProjectsManageAccess,
	PermissionOrganisationPage, PermissionMembersPage, PermissionRolesPage,
}

// billingAdminPermissions manage the subscription and nothing else.
var billingAdminPermissions = []Permission{
	PermissionOrganisationRead, PermissionBillingRead, PermissionBillingManage,
	PermissionOrganisationPage, PermissionSubscriptionPage,
}

// allProjectPermissions returns every project permission of the catalog: Project
// admin's set, so it never goes stale as the catalog grows.
func allProjectPermissions() []Permission {
	var permissions []Permission
	for _, definition := range permissionDefinitions {
		if definition.Level == LevelProject {
			permissions = append(permissions, definition.Permission)
		}
	}
	return permissions
}

func systemRoles() []Role {
	return []Role{
		systemRole(RoleKeyOrganisationManager, LevelOrganisation, "Organisation manager",
			"Members, roles, teams and organisation settings; grants access in every project without seeing its resources.", organisationManagerPermissions),
		systemRole(RoleKeyBillingAdmin, LevelOrganisation, "Billing admin",
			"The subscription and billing.", billingAdminPermissions),
		systemRole(RoleKeyProjectAdmin, LevelProject, "Project admin",
			"Everything in the project, including device configuration, settings and who has access.", allProjectPermissions()),
		systemRole(RoleKeyEditor, LevelProject, "Editor",
			"Works with cameras, recordings and cases, and creates and updates sites, groups and other project content. No deletes, device configuration or access.", editorPermissions),
		systemRole(RoleKeyViewer, LevelProject, "Viewer",
			"Watches live and recorded video and reads everything in scope; changes nothing.", viewerPermissions),
	}
}

func systemRole(key RoleKey, level Level, name, description string, permissions []Permission) Role {
	return Role{
		Name:         name,
		Description:  description,
		Kind:         RoleKindSystem,
		Key:          key,
		AssignableAt: level,
		Grants:       SelectionSelected,
		Permissions:  canonicalPermissions(permissions),
		IsActive:     true,
	}
}
