package models

import (
	"slices"
	"strings"
	"testing"
)

func systemRoleFor(t *testing.T, key RoleKey) Role {
	t.Helper()
	role, ok := BuiltInRole(key)
	if !ok || role.Kind != RoleKindSystem || role.Grants != SelectionSelected || len(role.Permissions) == 0 {
		t.Fatalf("%s = %+v, %v", key, role, ok)
	}
	return role
}

func TestSystemProjectRolesNest(t *testing.T) {
	chain := []RoleKey{RoleKeyViewer, RoleKeyEditor, RoleKeyProjectAdmin}
	for index := 1; index < len(chain); index++ {
		lower, higher := systemRoleFor(t, chain[index-1]), systemRoleFor(t, chain[index])
		held := ExpandPermissions(higher.Permissions)
		for _, permission := range ExpandPermissions(lower.Permissions) {
			if !slices.Contains(held, permission) {
				t.Errorf("%s holds %s but %s does not", lower.Key, permission, higher.Key)
			}
		}
	}
}

func TestProjectAdminHoldsEveryProjectPermission(t *testing.T) {
	admin := ExpandPermissions(systemRoleFor(t, RoleKeyProjectAdmin).Permissions)
	for _, definition := range PermissionDefinitions() {
		if definition.Level == LevelProject && !slices.Contains(admin, definition.Permission) {
			t.Errorf("project admin lacks %s", definition.Permission)
		}
	}
	if !slices.Contains(admin, PermissionProjectManageAccess) {
		t.Error("project admin cannot grant access in the project")
	}
}

func TestViewerChangesNothing(t *testing.T) {
	for _, permission := range ExpandPermissions(systemRoleFor(t, RoleKeyViewer).Permissions) {
		action := string(permission)[strings.LastIndex(string(permission), ".")+1:]
		readOnly := action == "read" || action == "preview" || action == "stream" || strings.HasPrefix(string(permission), "pages.")
		if !readOnly {
			t.Errorf("viewer holds %s", permission)
		}
	}
}

func TestEditorDeletesAndAdministersNothing(t *testing.T) {
	editor := ExpandPermissions(systemRoleFor(t, RoleKeyEditor).Permissions)
	for _, permission := range editor {
		if strings.HasSuffix(string(permission), ".delete") {
			t.Errorf("editor holds %s", permission)
		}
	}
	for _, permission := range []Permission{
		PermissionDevicesConfigure, PermissionDevicesManageRecording, PermissionDevicesReadLogs, PermissionDevicesManagePermissions,
		PermissionMediaForward, PermissionCasesManageMembership, PermissionCasesModerateComments,
		PermissionChannelsRead, PermissionProjectUpdate, PermissionProjectManageCredentials, PermissionProjectManageAccess,
		PermissionAlertsManagePrivacy, PermissionVideowallsAssign,
	} {
		if slices.Contains(editor, permission) {
			t.Errorf("editor holds %s", permission)
		}
	}
	for _, permission := range []Permission{PermissionPTZZoom, PermissionPTZPanTilt, PermissionPTZControl, PermissionPTZPresets, PermissionMediaExport, PermissionIOTrigger} {
		if !slices.Contains(editor, permission) {
			t.Errorf("editor lacks %s", permission)
		}
	}
}

func TestSystemOrganisationRolesGrantNoProjectAccess(t *testing.T) {
	for _, key := range []RoleKey{RoleKeyOrganisationManager, RoleKeyBillingAdmin} {
		role := systemRoleFor(t, key)
		if role.AssignableAt != LevelOrganisation || !role.RequiresAllProjects() {
			t.Errorf("%s: assignable at %s, all projects %v", key, role.AssignableAt, role.RequiresAllProjects())
		}
		for _, permission := range role.Permissions {
			if definition, _ := LookupPermission(permission); definition.Level != LevelOrganisation {
				t.Errorf("%s holds project permission %s", key, permission)
			}
		}
		for _, ownerOnly := range []Permission{PermissionStorageUpdate, PermissionAuditRead, PermissionDevicesManageSystemTags} {
			if slices.Contains(role.Permissions, ownerOnly) {
				t.Errorf("%s holds owner-only %s", key, ownerOnly)
			}
		}
	}
	manager := systemRoleFor(t, RoleKeyOrganisationManager).Permissions
	for _, permission := range []Permission{PermissionMembersUpdate, PermissionRolesCreate, PermissionTeamsUpdate, PermissionProjectsManageAccess} {
		if !slices.Contains(manager, permission) {
			t.Errorf("organisation manager lacks %s", permission)
		}
	}
	if slices.Contains(manager, PermissionBillingManage) {
		t.Error("organisation manager manages billing")
	}
	if billing := systemRoleFor(t, RoleKeyBillingAdmin).Permissions; !slices.Contains(billing, PermissionBillingManage) || slices.Contains(billing, PermissionMembersUpdate) {
		t.Errorf("billing admin %v", billing)
	}
}

// Every system project role shows the pages its content lives on, and only
// those: a page without anything to show on it is a dead end.
func TestSystemRolesListTheirPages(t *testing.T) {
	pageNeeds := map[Permission]Permission{
		PermissionPagesLiveView: PermissionLivePreview, PermissionPagesRecordings: PermissionMediaRead,
		PermissionPagesCases: PermissionCasesRead, PermissionPagesDevices: PermissionDevicesRead,
		PermissionPagesSites: PermissionSitesRead, PermissionPagesAlerts: PermissionAlertsRead,
		PermissionPagesWorkflows: PermissionWorkflowsRead, PermissionPagesChannels: PermissionChannelsRead,
		PermissionPagesSettings: PermissionLabelsCreate,
	}
	for _, key := range []RoleKey{RoleKeyViewer, RoleKeyEditor, RoleKeyProjectAdmin} {
		held := ExpandPermissions(systemRoleFor(t, key).Permissions)
		for page, needs := range pageNeeds {
			if slices.Contains(held, page) != slices.Contains(held, needs) {
				t.Errorf("%s: page %s = %v but %s = %v", key, page, slices.Contains(held, page), needs, slices.Contains(held, needs))
			}
		}
	}
}
