package models

import (
	"slices"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func legacyAccessFor(t *testing.T, in LegacyUserAccess) (LegacyAccess, Access) {
	t.Helper()
	mapped, err := MapLegacyAccess(in)
	if err != nil {
		t.Fatal(err)
	}
	return mapped, NewAccess(mapped.Membership, mapped.Roles, mapped.Assignments, time.Now())
}

func TestAccessLegacyGuest(t *testing.T) {
	ids := legacyIds(6)
	organisationId, userId, siteId, groupId, deviceId, otherProject := ids[0], ids[1], ids[2], ids[3], ids[4], ids[5]
	project := DefaultProjectId(organisationId)
	_, access := legacyAccessFor(t, LegacyUserAccess{OrganisationId: organisationId, UserId: userId, Role: "guest",
		SiteIds: []primitive.ObjectID{siteId}, GroupIds: []primitive.ObjectID{groupId}, DeviceIds: []primitive.ObjectID{deviceId}})

	other := primitive.NewObjectID()
	checks := []struct {
		name string
		got  bool
		want bool
	}{
		{"device in a selected site", access.Device(PermissionDevicesRead, project, DeviceRef{Id: other, SiteIds: []primitive.ObjectID{siteId}}), true},
		{"device in a selected group", access.Device(PermissionLivePreview, project, DeviceRef{Id: other, GroupIds: []primitive.ObjectID{groupId}}), true},
		{"selected device", access.Device(PermissionLiveStream, project, DeviceRef{Id: deviceId}), true},
		{"unselected device", access.Device(PermissionDevicesRead, project, DeviceRef{Id: other}), false},
		{"permission the role lacks", access.Device(PermissionDevicesUpdate, project, DeviceRef{Id: deviceId}), false},
		{"another project", access.Device(PermissionDevicesRead, otherProject, DeviceRef{Id: deviceId}), false},
		{"project resource", access.Project(PermissionCasesRead, project), true},
		{"project resource the role lacks", access.Project(PermissionCasesDelete, project), false},
		{"organisation permission", access.Organisation(PermissionMembersRead, primitive.NilObjectID), true},
		{"organisation permission the role lacks", access.Organisation(PermissionMembersInvite, primitive.NilObjectID), false},
		{"project permission asked as organisation", access.Organisation(PermissionCasesRead, project), false},
		{"organisation permission asked as project", access.Project(PermissionMembersRead, project), false},
		{"not an owner", access.IsOwner(), false},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %v, want %v", check.name, check.got, check.want)
		}
	}

	selection := access.Devices(PermissionDevicesRead, project)
	if selection.All || !slices.Equal(selection.SiteIds, []primitive.ObjectID{siteId}) || !slices.Equal(selection.DeviceIds, []primitive.ObjectID{deviceId}) {
		t.Errorf("selection %+v", selection)
	}
	permissions := access.Permissions(project)
	if !slices.Contains(permissions, PermissionLivePreview) || !slices.Contains(permissions, PermissionMembersRead) || slices.Contains(permissions, PermissionDevicesUpdate) {
		t.Errorf("permissions %v", permissions)
	}
	if !slices.Equal(permissions, catalogOrder(permissions)) {
		t.Error("permissions not in catalog order")
	}
}

func TestAccessEmptySelectionKeepsProjectResources(t *testing.T) {
	ids := legacyIds(2)
	project := DefaultProjectId(ids[0])
	_, access := legacyAccessFor(t, LegacyUserAccess{OrganisationId: ids[0], UserId: ids[1], Role: "editor"})
	if !access.Project(PermissionCasesCreate, project) {
		t.Error("an editor without devices still works with cases, as today")
	}
	if access.Device(PermissionDevicesRead, project, DeviceRef{Id: primitive.NewObjectID()}) {
		t.Error("no devices selected")
	}
	if selection := access.Devices(PermissionDevicesRead, project); selection.All || len(selection.DeviceIds)+len(selection.SiteIds)+len(selection.GroupIds) > 0 {
		t.Errorf("selection %+v", selection)
	}
}

func TestAccessOwner(t *testing.T) {
	ids := legacyIds(2)
	_, access := legacyAccessFor(t, LegacyUserAccess{OrganisationId: ids[0], UserId: ids[1], Role: "owner"})
	project := primitive.NewObjectID()
	if !access.IsOwner() || !access.Device(PermissionDevicesDelete, project, DeviceRef{Id: primitive.NewObjectID()}) ||
		!access.Organisation(PermissionBillingManage, primitive.NilObjectID) || !access.Devices(PermissionDevicesRead, project).All {
		t.Error("owner holds everything everywhere")
	}
	if got := access.Permissions(project); !slices.Equal(got, AllPermissions()) {
		t.Errorf("owner permissions %d, want all %d", len(got), len(AllPermissions()))
	}
}

func TestAccessDenies(t *testing.T) {
	ids := legacyIds(2)
	project := DefaultProjectId(ids[0])
	tests := []struct {
		name   string
		mutate func(*LegacyAccess)
	}{
		{"suspended membership", func(a *LegacyAccess) { a.Membership.Status = MembershipStatusSuspended }},
		{"expired membership", func(a *LegacyAccess) { a.Membership.ExpiresAt = time.Now().Add(-time.Minute) }},
		{"inactive role", func(a *LegacyAccess) { a.Roles[0].IsActive = false }},
		{"role missing", func(a *LegacyAccess) { a.Roles = a.Roles[1:] }},
		{"role of another organisation", func(a *LegacyAccess) { a.Roles[0].OrganisationId = primitive.NewObjectID() }},
		{"assignment for another user", func(a *LegacyAccess) { a.Assignments[0].Principal.Id = primitive.NewObjectID() }},
		{"expired assignment", func(a *LegacyAccess) { a.Assignments[0].ExpiresAt = time.Now().Add(-time.Minute) }},
		{"disabled assignment", func(a *LegacyAccess) { a.Assignments[0].IsActive = false }},
		{"invalid scope", func(a *LegacyAccess) { a.Assignments[0].Scope.Resources = "" }},
		{"closed time window", func(a *LegacyAccess) {
			a.Assignments[0].TimeWindow = &TimeWindow{Timezone: "UTC", Ranges: []TimeWindowRange{{Days: []Weekday{weekdays[time.Now().UTC().Add(48 * time.Hour).Weekday()]}, StartMinute: 0, EndMinute: 1440}}}
		}},
		{"unknown time zone", func(a *LegacyAccess) {
			a.Assignments[0].TimeWindow = &TimeWindow{Timezone: "Nowhere/Nothing", Ranges: []TimeWindowRange{{Days: weekdayNames, StartMinute: 0, EndMinute: 1440}}}
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mapped, err := MapLegacyAccess(LegacyUserAccess{OrganisationId: ids[0], UserId: ids[1], Role: "guest", DeviceIds: ids[1:]})
			if err != nil {
				t.Fatal(err)
			}
			test.mutate(&mapped)
			access := NewAccess(mapped.Membership, mapped.Roles, mapped.Assignments, time.Now())
			if access.Project(PermissionCasesRead, project) || access.Device(PermissionDevicesRead, project, DeviceRef{Id: ids[1]}) {
				t.Error("granted")
			}
		})
	}
}

func TestAccessOrganisationProjectSelection(t *testing.T) {
	ids := legacyIds(4)
	organisationId, userId, covered, uncovered := ids[0], ids[1], ids[2], ids[3]
	role := Role{Id: primitive.NewObjectID(), OrganisationId: organisationId, Name: "Delegate", Kind: RoleKindCustom,
		AssignableAt: LevelOrganisation, Grants: SelectionSelected, Permissions: []Permission{PermissionProjectsManageAccess}, IsActive: true}
	assignment := RoleAssignment{Id: primitive.NewObjectID(), OrganisationId: organisationId, Principal: Principal{Type: PrincipalTypeUser, Id: userId},
		RoleId: role.Id, Scope: RoleAssignmentScope{Level: LevelOrganisation, Projects: SelectionSelected, ProjectIds: []primitive.ObjectID{covered}},
		Source: AssignmentSourceManual, IsActive: true}
	membership := OrganisationUser{UserId: userId, OrganisationId: organisationId, Status: MembershipStatusActive}
	access := NewAccess(membership, []Role{role}, []RoleAssignment{assignment}, time.Now())
	if !access.Organisation(PermissionProjectsManageAccess, covered) {
		t.Error("covered project denied")
	}
	if access.Organisation(PermissionProjectsManageAccess, uncovered) || access.Organisation(PermissionProjectsManageAccess, primitive.NilObjectID) {
		t.Error("uncovered project or whole organisation granted")
	}
}
