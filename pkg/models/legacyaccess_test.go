package models

import (
	"errors"
	"math"
	"reflect"
	"slices"
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func legacyIds(n int) []primitive.ObjectID {
	ids := make([]primitive.ObjectID, n)
	for index := range ids {
		ids[index] = primitive.NewObjectID()
	}
	return ids
}

func roleKeys(access LegacyAccess) []RoleKey {
	var keys []RoleKey
	for _, role := range access.Roles {
		keys = append(keys, role.Key)
	}
	return keys
}

func TestMapLegacyAccessTiers(t *testing.T) {
	ids := legacyIds(4)
	organisationId, userId, siteId, deviceId := ids[0], ids[1], ids[2], ids[3]
	tests := []struct {
		role     string
		wantKeys []RoleKey
	}{
		{"owner", []RoleKey{RoleKeyOwner}},
		{"admin", []RoleKey{RoleKeyLegacyAdmin, RoleKeyLegacyAdminOrganisation}},
		{"editor", []RoleKey{RoleKeyLegacyEditor, RoleKeyLegacyMemberOrganisation}},
		{"guest", []RoleKey{RoleKeyLegacyGuest, RoleKeyLegacyMemberOrganisation}},
	}
	for _, test := range tests {
		t.Run(test.role, func(t *testing.T) {
			in := LegacyUserAccess{OrganisationId: organisationId, UserId: userId, Role: test.role,
				SiteIds: []primitive.ObjectID{siteId, siteId}, DeviceIds: []primitive.ObjectID{deviceId}}
			access, err := MapLegacyAccess(in)
			if err != nil {
				t.Fatal(err)
			}
			if got := roleKeys(access); !slices.Equal(got, test.wantKeys) {
				t.Fatalf("roles %v, want %v", got, test.wantKeys)
			}
			for index, assignment := range access.Assignments {
				role := access.Roles[index]
				if role.Id != BuiltInRoleId(organisationId, role.Key) || role.OrganisationId != organisationId {
					t.Errorf("%s: role id %s not derived", role.Key, role.Id.Hex())
				}
				if err := assignment.Validate(role); err != nil {
					t.Errorf("%s: %v", role.Key, err)
				}
				if role.AssignableAt == LevelProject {
					scope := assignment.Scope
					if scope.ProjectId != DefaultProjectId(organisationId) || scope.Resources != SelectionSelected ||
						!slices.Equal(scope.SiteIds, []primitive.ObjectID{siteId}) || !slices.Equal(scope.DeviceIds, []primitive.ObjectID{deviceId}) {
						t.Errorf("%s: project scope %+v", role.Key, scope)
					}
				} else if assignment.Scope.Projects != SelectionAll {
					t.Errorf("%s: organisation scope %+v", role.Key, assignment.Scope)
				}
			}
			again, _ := MapLegacyAccess(in)
			if !reflect.DeepEqual(access, again) {
				t.Error("mapping is not deterministic")
			}
		})
	}
}

func TestMapLegacyAccessCustomRole(t *testing.T) {
	ids := legacyIds(3)
	organisationId, userId, legacyRoleId := ids[0], ids[1], ids[2]
	custom := LegacyRole{
		Id: legacyRoleId, Name: "Operator", ParentRole: "guest", IsActive: 0,
		Pages:              []string{legacyPageLiveView},
		FeaturePermissions: LegacyFeatureLevels{PTZ: 2},
		TimeWindowActive:   1,
		TimeWindow:         LegacyTimeWindow{TimeRange1MinMonday: 8, TimeRange1MaxMonday: 17.5, TimeRange2MinMonday: 20.01, TimeRange2MaxMonday: 24},
	}
	access, err := MapLegacyAccess(LegacyUserAccess{OrganisationId: organisationId, UserId: userId, Role: "Operator", CustomRole: &custom, Timezone: "Europe/Brussels"})
	if err != nil {
		t.Fatal(err)
	}
	project, member := access.Roles[0], access.Roles[1]
	if project.Kind != RoleKindCustom || project.Name != "Operator" || !project.IsActive || project.Origin.LegacyRoleId != legacyRoleId || project.Origin.BaseRole != BaseRoleGuest {
		t.Errorf("project role %+v", project)
	}
	if want := []Permission{PermissionMarkersWrite, PermissionDevicesRead, PermissionLiveStream, PermissionPTZZoom, PermissionIOTrigger}; !slices.Equal(project.Permissions, want) {
		t.Errorf("project permissions %v, want %v", project.Permissions, want)
	}
	if member.Key != RoleKeyLegacyMemberOrganisation {
		t.Errorf("organisation part %s, want the member role", member.Key)
	}
	want := &TimeWindow{Timezone: "Europe/Brussels", Ranges: []TimeWindowRange{
		{Days: []Weekday{WeekdayMonday}, StartMinute: 480, EndMinute: 1050},
		{Days: []Weekday{WeekdayMonday}, StartMinute: 1201, EndMinute: 1440},
	}}
	for _, assignment := range access.Assignments {
		if !reflect.DeepEqual(assignment.TimeWindow, want) || !assignment.IsActive {
			t.Errorf("assignment window %+v", assignment.TimeWindow)
		}
	}
}

func TestMapLegacyAccessCustomOrganisationPart(t *testing.T) {
	ids := legacyIds(3)
	custom := LegacyRole{Id: ids[2], Name: "Site manager", ParentRole: "admin", Pages: []string{legacyPageAccounts, legacyPageSites}}
	access, err := MapLegacyAccess(LegacyUserAccess{OrganisationId: ids[0], UserId: ids[1], Role: custom.Name, CustomRole: &custom})
	if err != nil {
		t.Fatal(err)
	}
	organisation := access.Roles[1]
	if organisation.Kind != RoleKindCustom || organisation.AssignableAt != LevelOrganisation {
		t.Fatalf("organisation part %+v", organisation)
	}
	if !slices.Contains(organisation.Permissions, PermissionMembersInvite) || slices.Contains(organisation.Permissions, PermissionRolesRead) {
		t.Errorf("organisation permissions %v", organisation.Permissions)
	}
}

func TestMapLegacyAccessClosedWindowDisablesAssignments(t *testing.T) {
	ids := legacyIds(3)
	custom := LegacyRole{Id: ids[2], Name: "Nobody", ParentRole: "editor", TimeWindowActive: 1}
	access, err := MapLegacyAccess(LegacyUserAccess{OrganisationId: ids[0], UserId: ids[1], Role: custom.Name, CustomRole: &custom, Timezone: "Europe/Brussels"})
	if err != nil {
		t.Fatal(err)
	}
	for _, assignment := range access.Assignments {
		if assignment.IsActive || assignment.TimeWindow != nil {
			t.Errorf("assignment %+v should be inactive", assignment)
		}
	}
}

func TestMapLegacyAccessRejects(t *testing.T) {
	ids := legacyIds(3)
	role := func(mutate func(*LegacyRole)) *LegacyRole {
		custom := &LegacyRole{Id: ids[2], Name: "Custom", ParentRole: "editor"}
		mutate(custom)
		return custom
	}
	tests := []struct {
		name     string
		role     string
		custom   *LegacyRole
		timezone string
	}{
		{"application user", "application", nil, ""},
		{"missing custom role", "Custom", nil, ""},
		{"custom role with another name", "Other", role(func(*LegacyRole) {}), ""},
		{"owner parent", "Custom", role(func(r *LegacyRole) { r.ParentRole = "owner" }), ""},
		{"feature out of scale", "Custom", role(func(r *LegacyRole) { r.FeaturePermissions.PTZ = 6 }), ""},
		{"negative feature", "Custom", role(func(r *LegacyRole) { r.FeaturePermissions.Liveview = -1 }), ""},
		{"window value not a number", "Custom", role(func(r *LegacyRole) {
			r.TimeWindowActive, r.TimeWindow.TimeRange1MaxFriday = 1, math.NaN()
		}), "Europe/Brussels"},
		{"window without timezone", "Custom", role(func(r *LegacyRole) {
			r.TimeWindowActive, r.TimeWindow.TimeRange1MaxFriday = 1, 12
		}), ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := MapLegacyAccess(LegacyUserAccess{OrganisationId: ids[0], UserId: ids[1], Role: test.role, CustomRole: test.custom, Timezone: test.timezone})
			if !errors.Is(err, ErrLegacyAccessInvalid) && !errors.Is(err, ErrRoleAssignmentInvalid) {
				t.Errorf("err = %v", err)
			}
		})
	}
}

func TestLegacyTimeWindowNeverWidens(t *testing.T) {
	tests := []struct {
		name       string
		min, max   float64
		wantRanges []TimeWindowRange
	}{
		{"whole hours", 8, 17, []TimeWindowRange{{Days: []Weekday{WeekdayTuesday}, StartMinute: 480, EndMinute: 1020}}},
		{"fractions round inwards", 8.51, 9.249, []TimeWindowRange{{Days: []Weekday{WeekdayTuesday}, StartMinute: 511, EndMinute: 554}}},
		{"less than a whole minute", 8.999, 9.001, nil},
		{"end past midnight", 22, 30, []TimeWindowRange{{Days: []Weekday{WeekdayTuesday}, StartMinute: 1320, EndMinute: 1440}}},
		{"start after end", 17, 8, nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			custom := LegacyRole{Name: "Shift", TimeWindowActive: 1, TimeWindow: LegacyTimeWindow{TimeRange1MinTuesday: test.min, TimeRange1MaxTuesday: test.max}}
			window, open, err := custom.timeWindow("Europe/Brussels")
			if err != nil {
				t.Fatal(err)
			}
			if test.wantRanges == nil {
				if open || window != nil {
					t.Errorf("window %+v open %v, want closed", window, open)
				}
				return
			}
			if !open || !reflect.DeepEqual(window.Ranges, test.wantRanges) {
				t.Errorf("ranges %+v, want %+v", window, test.wantRanges)
			}
		})
	}
}
