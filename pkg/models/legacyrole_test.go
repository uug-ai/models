package models

import (
	"slices"
	"testing"
)

// verifiedLegacyRoles are the legacy role sets checked against the product's
// gates (hub-frontend 26446132, hub-api cd9eaedf). The page tables must keep
// reproducing them for the built-in tiers.
var verifiedLegacyRoles = map[RoleKey][]Permission{
	"legacy-guest":               {"markers.read", "markers.write", "media.read", "media.update", "media.export", "media.share", "media.redact", "media.forward", "cases.read", "cases.create", "cases.update", "cases.share", "cases.export", "cases.runWorkflow", "workflows.read", "workflow-runs.create", "workflow-runs.read", "days.read", "devices.read", "live.stream", "sites.read", "groups.read", "floor-plans.read", "io.read", "io.trigger", "notifications.read", "notifications.update", "labels.read", "analytics.read", "pages.dashboard", "pages.latestEvents", "pages.liveView", "pages.videowall", "pages.recordings", "pages.dayOverview", "pages.analytics", "pages.watchlist", "pages.cases", "pages.io", "pages.devices", "pages.sites", "pages.floorPlans", "pages.groups"},
	"legacy-editor":              {"markers.read", "markers.write", "media.read", "media.update", "media.export", "media.share", "media.redact", "media.forward", "cases.read", "cases.create", "cases.update", "cases.share", "cases.export", "cases.runWorkflow", "workflows.read", "workflow-runs.create", "workflow-runs.read", "days.read", "devices.read", "devices.configure", "live.stream", "ptz.presets", "sites.read", "groups.read", "floor-plans.read", "floor-plans.create", "floor-plans.update", "floor-plans.delete", "io.read", "io.trigger", "videowalls.create", "notifications.read", "notifications.update", "labels.read", "analytics.read", "pages.dashboard", "pages.latestEvents", "pages.liveView", "pages.videowall", "pages.recordings", "pages.dayOverview", "pages.analytics", "pages.watchlist", "pages.cases", "pages.io", "pages.devices", "pages.sites", "pages.floorPlans", "pages.groups"},
	"legacy-admin":               {"markers.read", "markers.write", "media.read", "media.update", "media.export", "media.share", "media.redact", "media.forward", "cases.read", "cases.create", "cases.update", "cases.share", "cases.export", "cases.runWorkflow", "cases.delete", "cases.manageMembership", "workflows.read", "workflow-runs.create", "workflow-runs.read", "days.read", "devices.read", "devices.update", "devices.configure", "live.stream", "ptz.presets", "sites.read", "sites.create", "sites.update", "groups.read", "groups.create", "floor-plans.read", "floor-plans.create", "floor-plans.update", "floor-plans.delete", "io.read", "io.delete", "io.trigger", "videowalls.read", "videowalls.create", "videowalls.update", "videowalls.delete", "videowalls.assign", "alerts.read", "alerts.create", "alerts.update", "alerts.managePrivacy", "notifications.read", "notifications.update", "labels.read", "labels.create", "labels.update", "labels.delete", "analytics.read", "project.update", "project.manageCredentials", "project.manageAccess", "pages.dashboard", "pages.latestEvents", "pages.liveView", "pages.videowall", "pages.recordings", "pages.dayOverview", "pages.analytics", "pages.watchlist", "pages.cases", "pages.alerts", "pages.io", "pages.devices", "pages.sites", "pages.floorPlans", "pages.groups", "pages.settings"},
	"legacy-member-organisation": {"organisation.read", "members.read"},
	"legacy-admin-organisation":  {"organisation.read", "organisation.update", "members.read", "members.invite", "members.update", "members.remove", "roles.read", "roles.create", "roles.update", "projects.create", "projects.delete", "projects.manageAccess", "audit.read", "pages.organisation", "pages.members", "pages.roles", "pages.audit"},
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
	always := []Permission{PermissionMarkersWrite, PermissionDevicesRead}
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
			wantOrganisation: []Permission{PermissionOrganisationRead, PermissionMembersRead},
		},
		{
			name:             "owner-only pages grant other tiers nothing",
			tier:             BaseRoleAdmin,
			pages:            []string{"channels", "plans-integrations", "subscription", "io"},
			wantProject:      []Permission{PermissionMarkersWrite, PermissionDevicesRead, PermissionProjectUpdate, PermissionProjectManageCredentials},
			wantOrganisation: []Permission{PermissionOrganisationRead, PermissionOrganisationUpdate, PermissionMembersRead, PermissionProjectsCreate, PermissionProjectsDelete, PermissionPagesOrganisation},
		},
		{
			name:             "live view follows the feature levels, not the tier",
			tier:             BaseRoleGuest,
			pages:            []string{legacyPageLiveView},
			wantProject:      []Permission{PermissionMarkersWrite, PermissionDevicesRead, PermissionLiveStream, PermissionPTZPresets, PermissionIOTrigger, PermissionPagesLiveView},
			wantOrganisation: []Permission{PermissionOrganisationRead, PermissionMembersRead},
		},
		{
			name:             "a role without device pages still reads its devices, and cases open live video",
			tier:             BaseRoleGuest,
			pages:            []string{legacyPageTasks},
			wantProject:      []Permission{PermissionMarkersWrite, PermissionMediaExport, PermissionMediaRedact, PermissionCasesRead, PermissionCasesCreate, PermissionCasesUpdate, PermissionCasesShare, PermissionCasesExport, PermissionCasesRunWorkflow, PermissionWorkflowsRead, PermissionWorkflowRunsCreate, PermissionWorkflowRunsRead, PermissionDevicesRead, PermissionLiveStream, PermissionPTZPresets, PermissionIOTrigger, PermissionLabelsRead, PermissionPagesCases},
			wantOrganisation: []Permission{PermissionOrganisationRead, PermissionMembersRead},
		},
		{
			name:             "group rows show live video",
			tier:             BaseRoleGuest,
			pages:            []string{legacyPageGroups},
			wantProject:      []Permission{PermissionMarkersWrite, PermissionDevicesRead, PermissionLiveStream, PermissionPTZPresets, PermissionGroupsRead, PermissionIOTrigger, PermissionPagesGroups},
			wantOrganisation: []Permission{PermissionOrganisationRead, PermissionMembersRead},
		},
		{
			name:             "the watchlist opens recordings beside live video",
			tier:             BaseRoleGuest,
			pages:            []string{legacyPageWatchlist},
			wantProject:      []Permission{PermissionMarkersWrite, PermissionMediaRead, PermissionMediaExport, PermissionMediaShare, PermissionCasesCreate, PermissionDevicesRead, PermissionLiveStream, PermissionPTZPresets, PermissionIOTrigger, PermissionNotificationsRead, PermissionNotificationsUpdate, PermissionPagesWatchlist},
			wantOrganisation: []Permission{PermissionOrganisationRead, PermissionMembersRead},
		},
		{
			name:             "analytics opens recordings beside live video",
			tier:             BaseRoleGuest,
			pages:            []string{legacyPageAnalytics},
			features:         FeaturePermissions{Liveview: 2, PTZ: 1},
			wantProject:      []Permission{PermissionMarkersRead, PermissionMarkersWrite, PermissionDevicesRead, PermissionLivePreview, PermissionIOTrigger, PermissionAnalyticsRead, PermissionPagesAnalytics},
			wantOrganisation: []Permission{PermissionOrganisationRead, PermissionMembersRead},
		},
		{
			name:             "graded levels select a grade or deny",
			tier:             BaseRoleEditor,
			pages:            []string{legacyPageCameras, legacyPageRecordings},
			features:         FeaturePermissions{Liveview: 2, PTZ: 3, RemoteConfig: 1, DownloadRecordings: 1},
			wantProject:      []Permission{PermissionMarkersWrite, PermissionMediaRead, PermissionMediaUpdate, PermissionMediaShare, PermissionMediaForward, PermissionCasesCreate, PermissionDaysRead, PermissionDevicesRead, PermissionLivePreview, PermissionPTZPanTilt, PermissionIORead, PermissionIOTrigger, PermissionPagesRecordings, PermissionPagesDevices},
			wantOrganisation: []Permission{PermissionOrganisationRead, PermissionMembersRead},
		},
		{
			name:             "guest custom role opts into the audit log",
			tier:             BaseRoleGuest,
			pages:            []string{legacyPageAudit},
			wantProject:      []Permission{PermissionMarkersWrite, PermissionDevicesRead},
			wantOrganisation: []Permission{PermissionOrganisationRead, PermissionMembersRead, PermissionAuditRead, PermissionPagesAudit},
		},
		{
			name:             "editor custom role opts into the audit log",
			tier:             BaseRoleEditor,
			pages:            []string{legacyPageAudit},
			wantProject:      []Permission{PermissionMarkersWrite, PermissionDevicesRead},
			wantOrganisation: []Permission{PermissionOrganisationRead, PermissionMembersRead, PermissionAuditRead, PermissionPagesAudit},
		},
		{
			name:             "admin without accounts and roles keeps no member or role administration",
			tier:             BaseRoleAdmin,
			pages:            []string{legacyPageTasks},
			wantProject:      []Permission{PermissionMarkersWrite, PermissionMediaExport, PermissionMediaRedact, PermissionCasesRead, PermissionCasesCreate, PermissionCasesUpdate, PermissionCasesShare, PermissionCasesExport, PermissionCasesRunWorkflow, PermissionCasesDelete, PermissionCasesManageMembership, PermissionWorkflowsRead, PermissionWorkflowRunsCreate, PermissionWorkflowRunsRead, PermissionDevicesRead, PermissionLiveStream, PermissionPTZPresets, PermissionIOTrigger, PermissionLabelsRead, PermissionProjectUpdate, PermissionProjectManageCredentials, PermissionPagesCases},
			wantOrganisation: []Permission{PermissionOrganisationRead, PermissionOrganisationUpdate, PermissionMembersRead, PermissionProjectsCreate, PermissionProjectsDelete, PermissionPagesOrganisation},
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
	for _, want := range []Permission{PermissionPagesDashboard, PermissionPagesRecordings, PermissionDevicesRead} {
		if !slices.Contains(project, want) {
			t.Errorf("missing %s in %v", want, project)
		}
	}
	for _, page := range []Permission{PermissionPagesDevices, PermissionPagesLiveView, PermissionPagesSettings} {
		if slices.Contains(project, page) {
			t.Errorf("%s shown without its page", page)
		}
	}
	// A page the tier is never served shows nothing.
	if project, _, _ := LegacyRolePermissions(BaseRoleGuest, []string{legacyPageAlerts, legacyPageSettings}, FeaturePermissions{}); slices.Contains(project, PermissionPagesAlerts) || slices.Contains(project, PermissionPagesSettings) {
		t.Errorf("guest shown admin pages: %v", project)
	}
}

func TestLegacyCustomRolePages(t *testing.T) {
	tests := []struct {
		name     string
		tier     BaseRole
		pages    []string
		disabled []string
		want     []string
	}{
		{"admin role from before the audit page holds it", BaseRoleAdmin, []string{legacyPageSettings}, nil, []string{legacyPageSettings, legacyPageAudit}},
		{"admin role switched off from the audit page", BaseRoleAdmin, []string{legacyPageSettings}, []string{legacyPageAudit}, []string{legacyPageSettings}},
		{"admin role listing the audit page keeps it once", BaseRoleAdmin, []string{legacyPageAudit}, nil, []string{legacyPageAudit}},
		{"guest role gets no default page", BaseRoleGuest, []string{legacyPageDashboard}, nil, []string{legacyPageDashboard}},
		{"guest role opts in by listing it", BaseRoleGuest, []string{legacyPageAudit}, nil, []string{legacyPageAudit}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := LegacyCustomRolePages(test.tier, test.pages, test.disabled); !slices.Equal(got, test.want) {
				t.Errorf("got %v, want %v", got, test.want)
			}
		})
	}
}
