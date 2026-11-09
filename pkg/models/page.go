package models

// Page permissions show a page: its menu entry and its route. They decide
// navigation only. What a page then shows follows the resource permissions
// (which devices, which recordings), so a page can be shown with nothing on
// it, and holding a resource permission never shows a page by itself. The API
// does not check page permissions.
const (
	PermissionDashboardPage    Permission = "dashboard.page"
	PermissionLatestEventsPage Permission = "latest-events.page"
	PermissionLivePage         Permission = "live.page"
	PermissionVideowallsPage   Permission = "videowalls.page"
	PermissionMediaPage        Permission = "media.page"
	PermissionDaysPage         Permission = "days.page"
	PermissionAnalyticsPage    Permission = "analytics.page"
	PermissionWatchlistPage    Permission = "watchlist.page"
	PermissionCasesPage        Permission = "cases.page"
	PermissionAlertsPage       Permission = "alerts.page"
	PermissionChannelsPage     Permission = "channels.page"
	PermissionWorkflowsPage    Permission = "workflows.page"
	PermissionIOPage           Permission = "io.page"
	PermissionDevicesPage      Permission = "devices.page"
	PermissionSitesPage        Permission = "sites.page"
	PermissionFloorPlansPage   Permission = "floor-plans.page"
	PermissionGroupsPage       Permission = "groups.page"
	PermissionSettingsPage     Permission = "settings.page"
	PermissionOrganisationPage Permission = "organisation.page"
	PermissionMembersPage      Permission = "members.page"
	PermissionRolesPage        Permission = "roles.page"
	PermissionSubscriptionPage Permission = "subscription.page"
	PermissionAuditPage        Permission = "audit.page"
)

// Legacy page identifiers remain accepted in stored roles and access checks.
// New roles and permission catalogs use the domain.page constants above.
const (
	PermissionPagesDashboard    Permission = "pages.dashboard"
	PermissionPagesLatestEvents Permission = "pages.latestEvents"
	PermissionPagesLiveView     Permission = "pages.liveView"
	PermissionPagesVideowall    Permission = "pages.videowall"
	PermissionPagesRecordings   Permission = "pages.recordings"
	PermissionPagesDayOverview  Permission = "pages.dayOverview"
	PermissionPagesAnalytics    Permission = "pages.analytics"
	PermissionPagesWatchlist    Permission = "pages.watchlist"
	PermissionPagesCases        Permission = "pages.cases"
	PermissionPagesAlerts       Permission = "pages.alerts"
	PermissionPagesChannels     Permission = "pages.channels"
	PermissionPagesWorkflows    Permission = "pages.workflows"
	PermissionPagesIO           Permission = "pages.io"
	PermissionPagesDevices      Permission = "pages.devices"
	PermissionPagesSites        Permission = "pages.sites"
	PermissionPagesFloorPlans   Permission = "pages.floorPlans"
	PermissionPagesGroups       Permission = "pages.groups"
	PermissionPagesSettings     Permission = "pages.settings"

	PermissionPagesOrganisation Permission = "pages.organisation"
	PermissionPagesMembers      Permission = "pages.members"
	PermissionPagesRoles        Permission = "pages.roles"
	PermissionPagesSubscription Permission = "pages.subscription"
	PermissionPagesAudit        Permission = "pages.audit"
)

var legacyPageAliases = map[Permission]Permission{
	PermissionPagesDashboard:    PermissionDashboardPage,
	PermissionPagesLatestEvents: PermissionLatestEventsPage,
	PermissionPagesLiveView:     PermissionLivePage,
	PermissionPagesVideowall:    PermissionVideowallsPage,
	PermissionPagesRecordings:   PermissionMediaPage,
	PermissionPagesDayOverview:  PermissionDaysPage,
	PermissionPagesAnalytics:    PermissionAnalyticsPage,
	PermissionPagesWatchlist:    PermissionWatchlistPage,
	PermissionPagesCases:        PermissionCasesPage,
	PermissionPagesAlerts:       PermissionAlertsPage,
	PermissionPagesChannels:     PermissionChannelsPage,
	PermissionPagesWorkflows:    PermissionWorkflowsPage,
	PermissionPagesIO:           PermissionIOPage,
	PermissionPagesDevices:      PermissionDevicesPage,
	PermissionPagesSites:        PermissionSitesPage,
	PermissionPagesFloorPlans:   PermissionFloorPlansPage,
	PermissionPagesGroups:       PermissionGroupsPage,
	PermissionPagesSettings:     PermissionSettingsPage,
	PermissionPagesOrganisation: PermissionOrganisationPage,
	PermissionPagesMembers:      PermissionMembersPage,
	PermissionPagesRoles:        PermissionRolesPage,
	PermissionPagesSubscription: PermissionSubscriptionPage,
	PermissionPagesAudit:        PermissionAuditPage,
}

// CanonicalPagePermission maps a known legacy page identifier to its canonical
// name. Other permissions are unchanged; this does not validate or grant them.
func CanonicalPagePermission(permission Permission) Permission {
	if canonical, ok := legacyPageAliases[permission]; ok {
		return canonical
	}
	return permission
}

// pagePermissions are the pages inside a project.
var pagePermissions = []Permission{
	PermissionDashboardPage,
	PermissionLatestEventsPage,
	PermissionLivePage,
	PermissionVideowallsPage,
	PermissionMediaPage,
	PermissionDaysPage,
	PermissionAnalyticsPage,
	PermissionWatchlistPage,
	PermissionCasesPage,
	PermissionAlertsPage,
	PermissionChannelsPage,
	PermissionWorkflowsPage,
	PermissionIOPage,
	PermissionDevicesPage,
	PermissionSitesPage,
	PermissionFloorPlansPage,
	PermissionGroupsPage,
	PermissionSettingsPage,
}

// organisationPagePermissions are the pages that administer the organisation.
var organisationPagePermissions = []Permission{
	PermissionOrganisationPage,
	PermissionMembersPage,
	PermissionRolesPage,
	PermissionSubscriptionPage,
	PermissionAuditPage,
}

// PagePermissions returns the project page permission catalog.
func PagePermissions() []Permission {
	return append([]Permission(nil), pagePermissions...)
}

// OrganisationPagePermissions returns the organisation page permission catalog.
func OrganisationPagePermissions() []Permission {
	return append([]Permission(nil), organisationPagePermissions...)
}
