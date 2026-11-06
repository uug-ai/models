package models

import "slices"

// The legacy tiers grant access through a page list and feature levels. The
// tables below record, per tier, which permissions each page gives, so a
// legacy custom role (a parent tier plus its own pages and feature levels)
// maps to exactly the permissions it has in the product today. The legacy
// setup is not consistent across tiers and pages; the tables keep it as it is
// rather than tidying it, since the legacy roles exist to preserve it.
//
// A page grants a permission when the product reaches that permission from
// the page. A permission reachable from several pages is listed with each of
// them, and a role keeps it when any of them is in its page list. A missing
// page also removes its permissions from the API (matrix decision D10), except
// devices.read: the device lists and device alerts never looked at the page
// list, so every role reads its devices whatever pages it is served.

// Legacy page keys, as stored in a role's page list. The owner-only pages
// (channels, io, subscription, plans-integrations) grant these tiers nothing.
const (
	legacyPageDashboard    = "dashboard"
	legacyPageLatestEvents = "latest-events"
	legacyPageLiveView     = "live-view"
	legacyPageVideowall    = "videowall"
	legacyPageRecordings   = "recordings"
	legacyPageDayOverview  = "day-overview"
	legacyPageAnalytics    = "analytics"
	legacyPageWatchlist    = "watchlist" // the notification feed
	legacyPageTasks        = "tasks"     // cases
	legacyPageAlerts       = "alerts"
	legacyPageIOs          = "ios"
	legacyPageCameras      = "cameras"
	legacyPageSites        = "sites"
	legacyPageFloorPlans   = "floor-plans"
	legacyPageGroups       = "groups"
	legacyPageAccounts     = "accounts"
	legacyPageRoles        = "roles"
	legacyPageSettings     = "settings"
	legacyPageAudit        = "audit"
)

// legacyPagePermissions is the page permission each legacy page key shows. A
// role holds it when the key is in its page list and its tier is served that
// page.
var legacyPagePermissions = map[string]Permission{
	legacyPageDashboard:    PermissionPagesDashboard,
	legacyPageLatestEvents: PermissionPagesLatestEvents,
	legacyPageLiveView:     PermissionPagesLiveView,
	legacyPageVideowall:    PermissionPagesVideowall,
	legacyPageRecordings:   PermissionPagesRecordings,
	legacyPageDayOverview:  PermissionPagesDayOverview,
	legacyPageAnalytics:    PermissionPagesAnalytics,
	legacyPageWatchlist:    PermissionPagesWatchlist,
	legacyPageTasks:        PermissionPagesCases,
	legacyPageAlerts:       PermissionPagesAlerts,
	legacyPageIOs:          PermissionPagesIO,
	legacyPageCameras:      PermissionPagesDevices,
	legacyPageSites:        PermissionPagesSites,
	legacyPageFloorPlans:   PermissionPagesFloorPlans,
	legacyPageGroups:       PermissionPagesGroups,
	legacyPageAccounts:     PermissionPagesMembers,
	legacyPageRoles:        PermissionPagesRoles,
	legacyPageSettings:     PermissionPagesSettings,
	legacyPageAudit:        PermissionPagesAudit,
}

// legacyAlways stands for "no page": its permissions are held whatever the
// page list says.
const legacyAlways = ""

// legacyTierPages is the page list each built-in tier is served today.
var legacyTierPages = map[BaseRole][]string{
	BaseRoleGuest: {
		legacyPageDashboard, legacyPageLatestEvents, legacyPageLiveView, legacyPageVideowall, legacyPageRecordings, legacyPageDayOverview,
		legacyPageAnalytics, legacyPageWatchlist, legacyPageTasks,
		legacyPageIOs, legacyPageCameras, legacyPageSites, legacyPageFloorPlans, legacyPageGroups,
	},
	BaseRoleEditor: {
		legacyPageDashboard, legacyPageLatestEvents, legacyPageLiveView, legacyPageVideowall, legacyPageRecordings, legacyPageDayOverview,
		legacyPageAnalytics, legacyPageWatchlist, legacyPageTasks,
		legacyPageIOs, legacyPageCameras, legacyPageSites, legacyPageFloorPlans, legacyPageGroups,
	},
	BaseRoleAdmin: {
		legacyPageDashboard, legacyPageLatestEvents, legacyPageLiveView, legacyPageVideowall, legacyPageRecordings, legacyPageDayOverview,
		legacyPageAnalytics, legacyPageWatchlist, legacyPageTasks,
		legacyPageAlerts,
		legacyPageIOs, legacyPageCameras, legacyPageSites, legacyPageFloorPlans, legacyPageGroups, legacyPageAccounts, legacyPageRoles, legacyPageSettings,
		legacyPageAudit,
	},
}

// legacyTierOptionalPages are pages a custom role on the tier may list
// although the built-in tier is not served them: the audit log is opt-in
// below the admin tier.
var legacyTierOptionalPages = map[BaseRole][]string{
	BaseRoleGuest:  {legacyPageAudit},
	BaseRoleEditor: {legacyPageAudit},
}

// legacyTierFeatures is the feature levels each built-in tier is served today:
// guests have no PTZ and no remote config, everything else is unset.
var legacyTierFeatures = map[BaseRole]FeaturePermissions{
	BaseRoleGuest:  {PTZ: 1, RemoteConfig: 1},
	BaseRoleEditor: {},
	BaseRoleAdmin:  {},
}

// Pages that share a reach: the recording pages, and the pages that show live
// video (with its PTZ and IO overlay; group rows show a live preview). The
// recording detail page (/media/:id) has no page key of its own and shows the
// device's live stream beside the recording; the watchlist, cases and
// analytics open it, so they reach live video too.
var (
	legacyMediaPages       = []string{legacyPageLatestEvents, legacyPageRecordings, legacyPageDayOverview}
	legacyMediaDetailPages = onPages(legacyMediaPages, []string{legacyPageWatchlist, legacyPageTasks, legacyPageAnalytics})
	legacyLivePages        = onPages([]string{legacyPageLiveView, legacyPageVideowall, legacyPageCameras, legacyPageSites, legacyPageFloorPlans, legacyPageGroups}, legacyMediaDetailPages)
)

// legacyPageGrant gives its permissions to a role listing any of its pages.
type legacyPageGrant struct {
	pages       []string
	permissions []Permission
}

func onPages(pages ...[]string) []string {
	return slices.Concat(pages...)
}

// legacyTierPageGrants lists the page grants of each tier, except graded
// permissions (live view, PTZ, remote config, download), which come from the
// feature levels; see legacyGradedPages. Markers are written by integrations,
// not on a page, and stay with every tier until access tokens are decided
// (D9).
var legacyTierPageGrants = map[BaseRole][]legacyPageGrant{
	// Not granted to guests although the API allows it: device edit, delete
	// and device permission limits, and PTZ presets (D1); site, group, floor
	// plan and label writes and IO delete (D2); workflow writes and run results
	// (D4); alerts and channels (D5); reading roles (D7); billing and storage
	// (D8); deleting recordings, creating IO, detections and moderating
	// comments, which have no screen (D13). Videowalls and private cases are
	// visible through their assigned-user and participant rules only.
	BaseRoleGuest: {
		{[]string{legacyAlways}, []Permission{PermissionOrganisationRead, PermissionMembersRead, PermissionMarkersWrite, PermissionDevicesRead}},
		{[]string{legacyPageIOs, legacyPageCameras}, []Permission{PermissionIORead}},
		{onPages(legacyLivePages, []string{legacyPageIOs}), []Permission{PermissionIOTrigger}},
		{[]string{legacyPageSites, legacyPageFloorPlans, legacyPageDashboard}, []Permission{PermissionSitesRead}},
		{[]string{legacyPageGroups}, []Permission{PermissionGroupsRead}},
		{[]string{legacyPageFloorPlans}, []Permission{PermissionFloorPlansRead}},
		{onPages(legacyMediaPages, []string{legacyPageWatchlist, legacyPageDashboard}), []Permission{PermissionMediaRead}},
		{legacyMediaPages, []Permission{PermissionMediaUpdate, PermissionMediaForward, PermissionDaysRead}},
		{onPages(legacyMediaPages, []string{legacyPageWatchlist}), []Permission{PermissionMediaShare}},
		{[]string{legacyPageAnalytics}, []Permission{PermissionAnalyticsRead, PermissionMarkersRead}},
		{[]string{legacyPageWatchlist}, []Permission{PermissionNotificationsRead, PermissionNotificationsUpdate}},
		{onPages(legacyMediaPages, []string{legacyPageWatchlist, legacyPageTasks}), []Permission{PermissionCasesCreate}},
		{[]string{legacyPageTasks}, []Permission{
			PermissionCasesRead, PermissionCasesUpdate, PermissionCasesShare, PermissionCasesExport, PermissionCasesRunWorkflow,
			PermissionMediaRedact, PermissionWorkflowsRead, PermissionWorkflowRunsCreate, PermissionWorkflowRunsRead, PermissionLabelsRead,
		}},
		{[]string{legacyPageAudit}, []Permission{PermissionAuditRead}},
	},
	// Editors add floor plan editing and creating videowalls (and changing
	// their own, by the creator rule); otherwise as guests. They may not assign
	// users to videowalls.
	BaseRoleEditor: {
		{[]string{legacyAlways}, []Permission{PermissionOrganisationRead, PermissionMembersRead, PermissionMarkersWrite, PermissionDevicesRead}},
		{[]string{legacyPageIOs, legacyPageCameras}, []Permission{PermissionIORead}},
		{onPages(legacyLivePages, []string{legacyPageIOs}), []Permission{PermissionIOTrigger}},
		{[]string{legacyPageSites, legacyPageFloorPlans, legacyPageDashboard}, []Permission{PermissionSitesRead}},
		{[]string{legacyPageGroups}, []Permission{PermissionGroupsRead}},
		{[]string{legacyPageFloorPlans}, []Permission{PermissionFloorPlansRead, PermissionFloorPlansCreate, PermissionFloorPlansUpdate, PermissionFloorPlansDelete}},
		{[]string{legacyPageVideowall}, []Permission{PermissionVideowallsCreate}},
		{onPages(legacyMediaPages, []string{legacyPageWatchlist, legacyPageDashboard}), []Permission{PermissionMediaRead}},
		{legacyMediaPages, []Permission{PermissionMediaUpdate, PermissionMediaForward, PermissionDaysRead}},
		{onPages(legacyMediaPages, []string{legacyPageWatchlist}), []Permission{PermissionMediaShare}},
		{[]string{legacyPageAnalytics}, []Permission{PermissionAnalyticsRead, PermissionMarkersRead}},
		{[]string{legacyPageWatchlist}, []Permission{PermissionNotificationsRead, PermissionNotificationsUpdate}},
		{onPages(legacyMediaPages, []string{legacyPageWatchlist, legacyPageTasks}), []Permission{PermissionCasesCreate}},
		{[]string{legacyPageTasks}, []Permission{
			PermissionCasesRead, PermissionCasesUpdate, PermissionCasesShare, PermissionCasesExport, PermissionCasesRunWorkflow,
			PermissionMediaRedact, PermissionWorkflowsRead, PermissionWorkflowRunsCreate, PermissionWorkflowRunsRead, PermissionLabelsRead,
		}},
		{[]string{legacyPageAudit}, []Permission{PermissionAuditRead}},
	},
	// Admins delete their own sites, update their own groups, delete their own
	// groups (D3) and delete their own alerts by the creator rule. Not granted
	// although the API allows it: device delete and device permission limits
	// (D1), workflow writes and run results (D4), channels (D5), deleting roles
	// (D7). Device recording mode and logs, billing, storage and device system
	// tags stay with owners.
	BaseRoleAdmin: {
		// Workspace settings have no page key: every admin reaches them.
		{[]string{legacyAlways}, []Permission{
			PermissionPagesOrganisation,
			PermissionOrganisationRead, PermissionOrganisationUpdate, PermissionMembersRead, PermissionMarkersWrite, PermissionDevicesRead,
			PermissionProjectsCreate, PermissionProjectsDelete, PermissionProjectUpdate, PermissionProjectManageCredentials,
		}},
		{[]string{legacyPageAccounts}, []Permission{
			PermissionMembersInvite, PermissionMembersUpdate, PermissionMembersRemove, PermissionProjectsManageAccess, PermissionProjectManageAccess,
		}},
		{[]string{legacyPageRoles}, []Permission{PermissionRolesRead, PermissionRolesCreate, PermissionRolesUpdate}},
		{[]string{legacyPageCameras}, []Permission{PermissionDevicesUpdate}},
		{[]string{legacyPageIOs, legacyPageCameras}, []Permission{PermissionIORead}},
		{[]string{legacyPageIOs}, []Permission{PermissionIODelete}},
		{onPages(legacyLivePages, []string{legacyPageIOs}), []Permission{PermissionIOTrigger}},
		{[]string{legacyPageSites, legacyPageFloorPlans, legacyPageDashboard}, []Permission{PermissionSitesRead}},
		{[]string{legacyPageSites}, []Permission{PermissionSitesCreate, PermissionSitesUpdate}},
		{[]string{legacyPageGroups}, []Permission{PermissionGroupsRead, PermissionGroupsCreate}},
		{[]string{legacyPageFloorPlans}, []Permission{PermissionFloorPlansRead, PermissionFloorPlansCreate, PermissionFloorPlansUpdate, PermissionFloorPlansDelete}},
		{[]string{legacyPageVideowall}, []Permission{
			PermissionVideowallsRead, PermissionVideowallsCreate, PermissionVideowallsUpdate, PermissionVideowallsDelete, PermissionVideowallsAssign,
		}},
		{onPages(legacyMediaPages, []string{legacyPageWatchlist, legacyPageDashboard}), []Permission{PermissionMediaRead}},
		{legacyMediaPages, []Permission{PermissionMediaUpdate, PermissionMediaForward, PermissionDaysRead}},
		{onPages(legacyMediaPages, []string{legacyPageWatchlist}), []Permission{PermissionMediaShare}},
		{[]string{legacyPageAnalytics}, []Permission{PermissionAnalyticsRead, PermissionMarkersRead}},
		{[]string{legacyPageWatchlist}, []Permission{PermissionNotificationsRead, PermissionNotificationsUpdate}},
		{onPages(legacyMediaPages, []string{legacyPageWatchlist, legacyPageTasks}), []Permission{PermissionCasesCreate}},
		{[]string{legacyPageTasks}, []Permission{
			PermissionCasesRead, PermissionCasesUpdate, PermissionCasesShare, PermissionCasesExport, PermissionCasesRunWorkflow,
			PermissionCasesDelete, PermissionCasesManageMembership,
			PermissionMediaRedact, PermissionWorkflowsRead, PermissionWorkflowRunsCreate, PermissionWorkflowRunsRead, PermissionLabelsRead,
		}},
		{[]string{legacyPageAlerts}, []Permission{PermissionAlertsRead, PermissionAlertsCreate, PermissionAlertsUpdate, PermissionAlertsManagePrivacy}},
		{[]string{legacyPageSettings}, []Permission{PermissionLabelsCreate, PermissionLabelsUpdate, PermissionLabelsDelete}},
		{[]string{legacyPageAudit}, []Permission{PermissionAuditRead}},
	},
}

// legacyGradedPages is where each graded capability is reached, for every
// tier: the feature level decides the grade, and the role needs one of these
// pages to hold it at all.
var (
	legacyGradedLivePages      = legacyLivePages
	legacyGradedConfigurePages = []string{legacyPageCameras}
	legacyGradedExportPages    = onPages(legacyMediaPages, []string{legacyPageWatchlist, legacyPageTasks})
)

// LegacyRolePermissions maps a legacy tier with a page list and feature levels
// (a built-in tier, or a custom role on that parent) to the permissions its
// legacy roles hold: the project part and the organisation part, each in
// catalog order without implied permissions. ok is false for a tier that has
// no legacy roles (owner maps to the Owner role).
//
// Feature levels follow the product: 0 (unset) allows the highest grade, 1
// denies, and higher values select a grade. Pages a tier is never served, nor
// offered as optional (legacyTierOptionalPages), grant nothing.
func LegacyRolePermissions(tier BaseRole, pages []string, features FeaturePermissions) (project, organisation []Permission, ok bool) {
	grants, ok := legacyTierPageGrants[tier]
	if !ok {
		return nil, nil, false
	}
	listed := func(reach []string) bool {
		for _, page := range reach {
			if page == legacyAlways || slices.Contains(pages, page) {
				return true
			}
		}
		return false
	}

	var held []Permission
	for _, page := range slices.Concat(legacyTierPages[tier], legacyTierOptionalPages[tier]) {
		if slices.Contains(pages, page) {
			held = append(held, legacyPagePermissions[page])
		}
	}
	for _, grant := range grants {
		if listed(grant.pages) {
			held = append(held, grant.permissions...)
		}
	}
	if listed(legacyGradedLivePages) {
		held = append(held, legacyLiveGrade(int(features.Liveview))...)
		held = append(held, legacyPTZGrade(int(features.PTZ))...)
	}
	if listed(legacyGradedConfigurePages) && legacyAllows(int(features.RemoteConfig)) {
		held = append(held, PermissionDevicesConfigure)
	}
	if listed(legacyGradedExportPages) && legacyAllows(int(features.DownloadRecordings)) {
		held = append(held, PermissionMediaExport)
	}

	for _, permission := range canonicalPermissions(held) {
		if definition, _ := LookupPermission(permission); definition.Level == LevelOrganisation {
			organisation = append(organisation, permission)
		} else {
			project = append(project, permission)
		}
	}
	return project, organisation, true
}

// legacyAllows reads an on/off feature level: unset or 2 and up allow.
func legacyAllows(level int) bool {
	return level == 0 || level >= 2
}

func legacyLiveGrade(level int) []Permission {
	switch {
	case level == 0 || level >= 3:
		return []Permission{PermissionLiveStream}
	case level == 2:
		return []Permission{PermissionLivePreview}
	}
	return nil
}

// legacyPTZGrade maps the PTZ scale: 2 zoom, 3 pan/tilt without zoom, 4 both,
// 5 (and unset) presets as well.
func legacyPTZGrade(level int) []Permission {
	switch {
	case level == 0 || level >= 5:
		return []Permission{PermissionPTZPresets}
	case level == 4:
		return []Permission{PermissionPTZControl}
	case level == 3:
		return []Permission{PermissionPTZPanTilt}
	case level == 2:
		return []Permission{PermissionPTZZoom}
	}
	return nil
}

// canonicalPermissions returns the distinct catalog permissions in catalog
// order, without any that another one in the set implies.
func canonicalPermissions(permissions []Permission) []Permission {
	var distinct []Permission
	for _, permission := range catalogOrder(permissions) {
		if _, ok := LookupPermission(permission); ok && !slices.Contains(distinct, permission) {
			distinct = append(distinct, permission)
		}
	}
	var canonical []Permission
	for _, permission := range distinct {
		others := slices.DeleteFunc(slices.Clone(distinct), func(other Permission) bool { return other == permission })
		if !slices.Contains(ExpandPermissions(others), permission) {
			canonical = append(canonical, permission)
		}
	}
	return canonical
}
