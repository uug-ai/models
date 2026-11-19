package models

import (
	"errors"
	"slices"
	"testing"
)

func TestPlanGatedDomains(t *testing.T) {
	for _, domain := range []string{"live", "ptz", "videowalls", "analytics", "detections", "workflows", "workflow-runs", "cases", "storage"} {
		t.Run(domain, func(t *testing.T) {
			count := 0
			for _, definition := range PermissionDefinitions() {
				if PermissionDomain(definition.Permission) != domain {
					continue
				}
				count++
				if !definition.PlanGated {
					t.Errorf("%s is not plan gated", definition.Permission)
				}
			}
			if count == 0 {
				t.Fatalf("domain %q has no permissions", domain)
			}
		})
	}
}

func TestPlanGatedPermissions(t *testing.T) {
	for _, permission := range []Permission{
		PermissionMediaPlayHD,
		PermissionLivePage, PermissionVideowallsPage, PermissionAnalyticsPage, PermissionCasesPage, PermissionWorkflowsPage,
		PermissionPagesLiveView, PermissionPagesVideowall, PermissionPagesAnalytics, PermissionPagesCases, PermissionPagesWorkflows,
		PermissionStorageUpdate, PermissionPTZControl, PermissionWorkflowRunsCreate, PermissionDetectionsRead,
	} {
		if !IsPlanGated(permission) {
			t.Errorf("%s is not plan gated", permission)
		}
	}
	for _, permission := range []Permission{
		PermissionMediaRead, PermissionMediaExport, PermissionDevicesRead, PermissionDevicesPage, PermissionDashboardPage,
		PermissionWatchlistPage, PermissionMediaPage, PermissionSitesRead, PermissionChannelsRead,
		PermissionBillingManage, PermissionBillingRead, PermissionOrganisationUpdate, PermissionMembersInvite,
		PermissionRolesUpdate, PermissionTeamsCreate, PermissionSubscriptionPage, PermissionPagesSubscription,
		PermissionOrganisationPage, PermissionMembersPage, PermissionRolesPage, "unknown.permission",
	} {
		if IsPlanGated(permission) {
			t.Errorf("%s is plan gated", permission)
		}
	}
	gated, ungated := PlanGatedPermissions(), UngatedPermissions()
	if len(gated)+len(ungated) != len(AllPermissions()) || len(gated) == 0 {
		t.Errorf("gated %d + ungated %d != %d", len(gated), len(ungated), len(AllPermissions()))
	}
	if !slices.Contains(AllPermissions(), PermissionMediaPlayHD) {
		t.Error("media.playHD is not in the catalog")
	}
}

func TestValidatePermissionCatalog(t *testing.T) {
	if err := ValidatePermissionCatalog(); err != nil {
		t.Fatal(err)
	}
}

func TestPlanGatingNeverGated(t *testing.T) {
	gated := func(permission Permission) []PermissionDefinition {
		return []PermissionDefinition{{Permission: permission, Level: LevelOrganisation, PlanGated: true}}
	}
	for _, permission := range []Permission{
		PermissionBillingManage, PermissionOrganisationRead, PermissionMembersInvite, PermissionRolesCreate,
		PermissionTeamsUpdate, PermissionSubscriptionPage, PermissionPagesSubscription,
	} {
		if err := validatePlanGating(gated(permission)); !errors.Is(err, ErrPlanGatingInvalid) {
			t.Errorf("gating %s: error = %v", permission, err)
		}
	}
	if err := validatePlanGating(gated(PermissionLiveStream)); err != nil {
		t.Errorf("gating live.stream: %v", err)
	}

	definitions := PermissionDefinitions()
	for _, test := range []struct {
		name    string
		domains []string
		singles []Permission
	}{
		{"billing domain", []string{"billing"}, nil},
		{"organisation domain", []string{"organisation"}, nil},
		{"members domain", []string{"members"}, nil},
		{"roles domain", []string{"roles"}, nil},
		{"teams domain", []string{"teams"}, nil},
		{"subscription page", nil, []Permission{PermissionSubscriptionPage}},
		{"legacy subscription page", nil, []Permission{PermissionPagesSubscription}},
		{"billing permission", nil, []Permission{PermissionBillingRead}},
		{"unknown domain", []string{"liveview"}, nil},
		{"unknown permission", nil, []Permission{"media.play4K"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := validatePlanGatingLists(definitions, test.domains, test.singles); !errors.Is(err, ErrPlanGatingInvalid) {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func TestResolvePlanPermissions(t *testing.T) {
	ptz := []Permission{PermissionPTZZoom, PermissionPTZPanTilt, PermissionPTZControl, PermissionPTZPresets}
	tests := []struct {
		name               string
		patterns, add, rmv []string
		want               []Permission
	}{
		{name: "domain", patterns: []string{"ptz.*"}, want: ptz},
		{name: "implies", patterns: []string{"ptz.control"}, want: []Permission{PermissionPTZZoom, PermissionPTZPanTilt, PermissionPTZControl}},
		{name: "legacy page name", patterns: []string{"pages.liveView"}, want: []Permission{PermissionLivePage}},
		{name: "remove", patterns: []string{"ptz.*"}, rmv: []string{"ptz.presets"}, want: ptz[:3]},
		{name: "remove implied removes implier", patterns: []string{"ptz.*"}, rmv: []string{"ptz.panTilt"}, want: []Permission{PermissionPTZZoom}},
		{name: "add", patterns: []string{"ptz.zoom"}, add: []string{"media.playHD"}, want: []Permission{PermissionMediaPlayHD, PermissionPTZZoom}},
		{name: "remove wins over add", add: []string{"storage.*"}, rmv: []string{"storage.*"}},
		{name: "ungated and unknown patterns", patterns: []string{"media.read", "nothing.*", "billing.*"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := ResolvePlanPermissions(test.patterns, test.add, test.rmv); !slices.Equal(got, test.want) {
				t.Fatalf("got %v, want %v", got, test.want)
			}
		})
	}
}

func TestValidatePlanPermissionPatterns(t *testing.T) {
	if err := ValidatePlanPermissionPatterns([]string{"live.*", "ptz.control", "media.playHD", "pages.liveView"}); err != nil {
		t.Fatal(err)
	}
	for _, pattern := range []string{"media.read", "devices.*", "billing.*", "*", ".*", "live", "live.unknown"} {
		if err := ValidatePlanPermissionPatterns([]string{pattern}); !errors.Is(err, ErrPlanPermissionPatternInvalid) {
			t.Errorf("%q: error = %v", pattern, err)
		}
	}
}
