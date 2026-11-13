package models

import (
	"slices"
	"testing"
)

// verifiedLegacyRoles are the legacy role sets checked against the product's
// gates (hub-frontend 26446132, hub-api cd9eaedf). The page tables must keep
// reproducing them for the built-in tiers. One deliberate change since: the
// member list (members.read) left guests and editors, who now read the member
// directory instead.
var verifiedLegacyRoles = map[RoleKey][]Permission{
	"legacy-guest":               {"markers.read", "markers.write", "media.read", "media.update", "media.export", "media.share", "media.redact", "media.forward", "cases.read", "cases.create", "cases.update", "cases.share", "cases.export", "cases.runWorkflow", "workflows.read", "workflow-runs.create", "workflow-runs.read", "days.read", "devices.read", "live.stream", "sites.read", "groups.read", "floor-plans.read", "io.read", "io.trigger", "notifications.read", "notifications.update", "labels.read", "analytics.read", "dashboard.page", "latest-events.page", "live.page", "videowalls.page", "media.page", "days.page", "analytics.page", "watchlist.page", "cases.page", "io.page", "devices.page", "sites.page", "floor-plans.page", "groups.page"},
	"legacy-editor":              {"markers.read", "markers.write", "media.read", "media.update", "media.export", "media.share", "media.redact", "media.forward", "cases.read", "cases.create", "cases.update", "cases.share", "cases.export", "cases.runWorkflow", "workflows.read", "workflow-runs.create", "workflow-runs.read", "days.read", "devices.read", "devices.configure", "live.stream", "ptz.presets", "sites.read", "groups.read", "floor-plans.read", "floor-plans.create", "floor-plans.update", "floor-plans.delete", "io.read", "io.trigger", "videowalls.create", "notifications.read", "notifications.update", "labels.read", "analytics.read", "dashboard.page", "latest-events.page", "live.page", "videowalls.page", "media.page", "days.page", "analytics.page", "watchlist.page", "cases.page", "io.page", "devices.page", "sites.page", "floor-plans.page", "groups.page"},
	"legacy-admin":               {"markers.read", "markers.write", "media.read", "media.update", "media.export", "media.share", "media.redact", "media.forward", "cases.read", "cases.create", "cases.update", "cases.share", "cases.export", "cases.runWorkflow", "cases.delete", "cases.manageMembership", "workflows.read", "workflow-runs.create", "workflow-runs.read", "days.read", "devices.read", "devices.update", "devices.configure", "live.stream", "ptz.presets", "sites.read", "sites.create", "sites.update", "groups.read", "groups.create", "floor-plans.read", "floor-plans.create", "floor-plans.update", "floor-plans.delete", "io.read", "io.delete", "io.trigger", "videowalls.read", "videowalls.create", "videowalls.update", "videowalls.delete", "videowalls.assign", "alerts.read", "alerts.create", "alerts.update", "alerts.managePrivacy", "notifications.read", "notifications.update", "labels.read", "labels.create", "labels.update", "labels.delete", "analytics.read", "project.update", "project.manageCredentials", "project.manageAccess", "dashboard.page", "latest-events.page", "live.page", "videowalls.page", "media.page", "days.page", "analytics.page", "watchlist.page", "cases.page", "alerts.page", "io.page", "devices.page", "sites.page", "floor-plans.page", "groups.page", "settings.page"},
	"legacy-member-organisation": {"organisation.read"},
	"legacy-admin-organisation":  {"organisation.read", "organisation.update", "members.read", "members.invite", "members.update", "members.remove", "roles.read", "roles.create", "roles.update", "projects.create", "projects.delete", "projects.manageAccess", "audit.read", "organisation.page", "members.page", "roles.page", "audit.page"},
}

func TestBuiltInLegacyRolesMatchVerifiedSets(t *testing.T) {
	for key, want := range verifiedLegacyRoles {
		role, ok := BuiltInRole(key)
		if !ok {
			t.Fatalf("%s missing", key)
		}
		if !slices.Equal(role.Permissions, want) {
			t.Errorf("%s:\n got %v\nwant %v", key, role.Permissions, want)
		}
	}
}

func TestLegacyGuestAndEditorShareOrganisationPart(t *testing.T) {
	_, guest := legacyTierPermissions(BaseRoleGuest)
	_, editor := legacyTierPermissions(BaseRoleEditor)
	if !slices.Equal(guest, editor) {
		t.Errorf("guest %v, editor %v", guest, editor)
	}
}

func TestLegacyRolePermissions(t *testing.T) {
	// devices.read follows no page: the device lists and device alerts never
	// read the page list.
	always := []Permission{PermissionMarkersWrite, PermissionDevicesRead, PermissionSitesRead, PermissionGroupsRead}
	tests := []struct {
		name             string
		tier             BaseRole
		pages            []string
		features         FeaturePermissions
		wantProject      []Permission
		wantOrganisation []Permission
	}{
		{
			name:             "no pages keeps only what no page gates",
			tier:             BaseRoleGuest,
			wantProject:      always,
			wantOrganisation: []Permission{PermissionOrganisationRead},
		},
		{
			name:             "owner-only pages grant other tiers nothing",
			tier:             BaseRoleAdmin,
			pages:            []string{"channels", "plans-integrations", "subscription", "io"},
			wantProject:      []Permission{PermissionMarkersWrite, PermissionDevicesRead, PermissionSitesRead, PermissionGroupsRead, PermissionProjectUpdate, PermissionProjectManageCredentials},
			wantOrganisation: []Permission{PermissionOrganisationRead, PermissionOrganisationUpdate, PermissionMembersRead, PermissionProjectsCreate, PermissionProjectsDelete, PermissionOrganisationPage},
		},
		{
			name:             "live view follows the feature levels, not the tier",
			tier:             BaseRoleGuest,
			pages:            []string{legacyPageLiveView},
			wantProject:      []Permission{PermissionMarkersWrite, PermissionDevicesRead, PermissionLiveStream, PermissionPTZPresets, PermissionSitesRead, PermissionGroupsRead, PermissionIOTrigger, PermissionLivePage},
			wantOrganisation: []Permission{PermissionOrganisationRead},
		},
		{
			name:             "a role without device pages still reads its devices, and cases open live video",
			tier:             BaseRoleGuest,
			pages:            []string{legacyPageTasks},
			wantProject:      []Permission{PermissionMarkersWrite, PermissionMediaExport, PermissionMediaRedact, PermissionCasesRead, PermissionCasesCreate, PermissionCasesUpdate, PermissionCasesShare, PermissionCasesExport, PermissionCasesRunWorkflow, PermissionWorkflowsRead, PermissionWorkflowRunsCreate, PermissionWorkflowRunsRead, PermissionDevicesRead, PermissionLiveStream, PermissionPTZPresets, PermissionSitesRead, PermissionGroupsRead, PermissionIOTrigger, PermissionLabelsRead, PermissionCasesPage},
			wantOrganisation: []Permission{PermissionOrganisationRead},
		},
		{
			name:             "group rows show live video",
			tier:             BaseRoleGuest,
			pages:            []string{legacyPageGroups},
			wantProject:      []Permission{PermissionMarkersWrite, PermissionDevicesRead, PermissionLiveStream, PermissionPTZPresets, PermissionSitesRead, PermissionGroupsRead, PermissionIOTrigger, PermissionGroupsPage},
			wantOrganisation: []Permission{PermissionOrganisationRead},
		},
		{
			name:             "the watchlist opens recordings beside live video",
			tier:             BaseRoleGuest,
			pages:            []string{legacyPageWatchlist},
			wantProject:      []Permission{PermissionMarkersWrite, PermissionMediaRead, PermissionMediaExport, PermissionMediaShare, PermissionCasesCreate, PermissionDevicesRead, PermissionLiveStream, PermissionPTZPresets, PermissionSitesRead, PermissionGroupsRead, PermissionIOTrigger, PermissionNotificationsRead, PermissionNotificationsUpdate, PermissionWatchlistPage},
			wantOrganisation: []Permission{PermissionOrganisationRead},
		},
		{
			name:             "analytics opens recordings beside live video",
			tier:             BaseRoleGuest,
			pages:            []string{legacyPageAnalytics},
			features:         FeaturePermissions{Liveview: 2, PTZ: 1},
			wantProject:      []Permission{PermissionMarkersRead, PermissionMarkersWrite, PermissionDevicesRead, PermissionLivePreview, PermissionSitesRead, PermissionGroupsRead, PermissionIOTrigger, PermissionAnalyticsRead, PermissionAnalyticsPage},
			wantOrganisation: []Permission{PermissionOrganisationRead},
		},
		{
			name:             "graded levels select a grade or deny",
			tier:             BaseRoleEditor,
			pages:            []string{legacyPageCameras, legacyPageRecordings},
			features:         FeaturePermissions{Liveview: 2, PTZ: 3, RemoteConfig: 1, DownloadRecordings: 1},
			wantProject:      []Permission{PermissionMarkersWrite, PermissionMediaRead, PermissionMediaUpdate, PermissionMediaShare, PermissionMediaForward, PermissionCasesCreate, PermissionDaysRead, PermissionDevicesRead, PermissionLivePreview, PermissionPTZPanTilt, PermissionSitesRead, PermissionGroupsRead, PermissionIORead, PermissionIOTrigger, PermissionMediaPage, PermissionDevicesPage},
			wantOrganisation: []Permission{PermissionOrganisationRead},
		},
		{
			name:             "guest custom role opts into the audit log",
			tier:             BaseRoleGuest,
			pages:            []string{legacyPageAudit},
			wantProject:      []Permission{PermissionMarkersWrite, PermissionDevicesRead, PermissionSitesRead, PermissionGroupsRead},
			wantOrganisation: []Permission{PermissionOrganisationRead, PermissionAuditRead, PermissionAuditPage},
		},
		{
			name:             "editor custom role opts into the audit log",
			tier:             BaseRoleEditor,
			pages:            []string{legacyPageAudit},
			wantProject:      []Permission{PermissionMarkersWrite, PermissionDevicesRead, PermissionSitesRead, PermissionGroupsRead},
			wantOrganisation: []Permission{PermissionOrganisationRead, PermissionAuditRead, PermissionAuditPage},
		},
		{
			name:             "admin without accounts and roles keeps no member or role administration",
			tier:             BaseRoleAdmin,
			pages:            []string{legacyPageTasks},
			wantProject:      []Permission{PermissionMarkersWrite, PermissionMediaExport, PermissionMediaRedact, PermissionCasesRead, PermissionCasesCreate, PermissionCasesUpdate, PermissionCasesShare, PermissionCasesExport, PermissionCasesRunWorkflow, PermissionCasesDelete, PermissionCasesManageMembership, PermissionWorkflowsRead, PermissionWorkflowRunsCreate, PermissionWorkflowRunsRead, PermissionDevicesRead, PermissionLiveStream, PermissionPTZPresets, PermissionSitesRead, PermissionGroupsRead, PermissionIOTrigger, PermissionLabelsRead, PermissionProjectUpdate, PermissionProjectManageCredentials, PermissionCasesPage},
			wantOrganisation: []Permission{PermissionOrganisationRead, PermissionOrganisationUpdate, PermissionMembersRead, PermissionProjectsCreate, PermissionProjectsDelete, PermissionOrganisationPage},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			project, organisation, ok := LegacyRolePermissions(test.tier, test.pages, test.features)
			if !ok {
				t.Fatal("tier not mapped")
			}
			if !slices.Equal(project, test.wantProject) {
				t.Errorf("project:\n got %v\nwant %v", project, test.wantProject)
			}
			if !slices.Equal(organisation, test.wantOrganisation) {
				t.Errorf("organisation:\n got %v\nwant %v", organisation, test.wantOrganisation)
			}
		})
	}
	if _, _, ok := LegacyRolePermissions(BaseRoleOwner, nil, FeaturePermissions{}); ok {
		t.Error("owner has no legacy roles")
	}
}

func TestLegacyAdminCustomRoleOptsIntoAudit(t *testing.T) {
	_, without, _ := LegacyRolePermissions(BaseRoleAdmin, []string{legacyPageSettings}, FeaturePermissions{})
	if slices.Contains(without, PermissionAuditRead) || slices.Contains(without, PermissionAuditPage) {
		t.Errorf("admin custom role without the audit page holds audit: %v", without)
	}
	_, with, _ := LegacyRolePermissions(BaseRoleAdmin, []string{legacyPageSettings, legacyPageAudit}, FeaturePermissions{})
	if !slices.Contains(with, PermissionAuditRead) || !slices.Contains(with, PermissionAuditPage) {
		t.Errorf("admin custom role listing the audit page lacks audit: %v", with)
	}
}

func TestLegacyPageTables(t *testing.T) {
	for tier, grants := range legacyTierPageGrants {
		for _, grant := range grants {
			for _, page := range grant.pages {
				if page != legacyAlways && !slices.Contains(legacyTierPages[BaseRoleAdmin], page) {
					t.Errorf("%s: unknown page %q", tier, page)
				}
			}
			for _, permission := range grant.permissions {
				if _, ok := LookupPermission(permission); !ok {
					t.Errorf("%s: unknown permission %q", tier, permission)
				}
			}
		}
	}
}

func TestLegacyPagesShowOnlyThemselves(t *testing.T) {
	for _, page := range legacyTierPages[BaseRoleAdmin] {
		if _, ok := legacyPagePermissions[page]; !ok {
			t.Errorf("page %q has no page permission", page)
		}
	}
	// A role on the dashboard and recordings reads devices for them, but is
	// not shown the devices page.
	project, _, _ := LegacyRolePermissions(BaseRoleAdmin, []string{legacyPageDashboard, legacyPageRecordings}, FeaturePermissions{})
	for _, want := range []Permission{PermissionDashboardPage, PermissionMediaPage, PermissionDevicesRead} {
		if !slices.Contains(project, want) {
			t.Errorf("missing %s in %v", want, project)
		}
	}
	for _, page := range []Permission{PermissionDevicesPage, PermissionLivePage, PermissionSettingsPage} {
		if slices.Contains(project, page) {
			t.Errorf("%s shown without its page", page)
		}
	}
	// A page the tier is never served shows nothing.
	if project, _, _ := LegacyRolePermissions(BaseRoleGuest, []string{legacyPageAlerts, legacyPageSettings}, FeaturePermissions{}); slices.Contains(project, PermissionAlertsPage) || slices.Contains(project, PermissionSettingsPage) {
		t.Errorf("guest shown admin pages: %v", project)
	}
}
