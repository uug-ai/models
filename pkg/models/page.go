package models

// Page permissions show a page: its menu entry and its route. They decide
// navigation only. What a page then shows follows the resource permissions
// (which devices, which recordings), so a page can be shown with nothing on
// it, and holding a resource permission never shows a page by itself. The API
// does not check page permissions.
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

// pagePermissions are the pages inside a project.
var pagePermissions = []Permission{
	PermissionPagesDashboard,
	PermissionPagesLatestEvents,
	PermissionPagesLiveView,
	PermissionPagesVideowall,
	PermissionPagesRecordings,
	PermissionPagesDayOverview,
	PermissionPagesAnalytics,
	PermissionPagesWatchlist,
	PermissionPagesCases,
	PermissionPagesAlerts,
	PermissionPagesChannels,
	PermissionPagesWorkflows,
	PermissionPagesIO,
	PermissionPagesDevices,
	PermissionPagesSites,
	PermissionPagesFloorPlans,
	PermissionPagesGroups,
	PermissionPagesSettings,
}

// organisationPagePermissions are the pages that administer the organisation.
var organisationPagePermissions = []Permission{
	PermissionPagesOrganisation,
	PermissionPagesMembers,
	PermissionPagesRoles,
	PermissionPagesSubscription,
	PermissionPagesAudit,
}

// PagePermissions returns the project page permission catalog.
func PagePermissions() []Permission {
	return append([]Permission(nil), pagePermissions...)
}

// OrganisationPagePermissions returns the organisation page permission catalog.
func OrganisationPagePermissions() []Permission {
	return append([]Permission(nil), organisationPagePermissions...)
}
