package models

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type fakeLegacyAccessStore struct {
	roles  map[string][]LegacyRole
	owners map[primitive.ObjectID]primitive.ObjectID
	err    error
	reads  int
}

func (s *fakeLegacyAccessStore) LegacyRoles(_ context.Context, _ primitive.ObjectID, name string) ([]LegacyRole, error) {
	s.reads++
	return s.roles[name], s.err
}

func (s *fakeLegacyAccessStore) OrganisationOwner(_ context.Context, organisationId primitive.ObjectID) (primitive.ObjectID, error) {
	s.reads++
	return s.owners[organisationId], s.err
}

func TestResolveLegacyAccessHomeOrganisation(t *testing.T) {
	home, userId, siteId := primitive.NewObjectID(), primitive.NewObjectID(), primitive.NewObjectID()
	project := DefaultProjectId(home)
	store := &fakeLegacyAccessStore{}
	resolution, err := ResolveLegacyAccess(context.Background(), store, LegacySubject{
		UserId: userId, HomeOrganisationId: home, OrganisationId: home, Role: "admin",
		Sites: []string{siteId.Hex(), "not-an-id"}, Cameras: []string{"camera-1", "", "camera-1"},
	}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if len(resolution.Reasons) > 0 || store.reads != 0 {
		t.Errorf("reasons %v, %d store reads; a tier needs neither", resolution.Reasons, store.reads)
	}
	access := resolution.Access
	if !access.Organisation(PermissionMembersUpdate, primitive.NilObjectID) || !access.Project(PermissionCasesDelete, project) {
		t.Error("admin lacks admin permissions")
	}
	if !access.Device(PermissionDevicesRead, project, DeviceRef{Key: "camera-1"}) ||
		!access.Device(PermissionDevicesRead, project, DeviceRef{Key: "camera-2", SiteIds: []primitive.ObjectID{siteId}}) {
		t.Error("selected camera or site not covered")
	}
	if access.Device(PermissionDevicesRead, project, DeviceRef{Key: "camera-2"}) {
		t.Error("unselected device covered")
	}
}

func TestResolveLegacyAccessCustomRole(t *testing.T) {
	home, userId := primitive.NewObjectID(), primitive.NewObjectID()
	store := &fakeLegacyAccessStore{roles: map[string][]LegacyRole{
		"Operator": {{Id: primitive.NewObjectID(), Name: "Operator", ParentRole: "guest", Pages: []string{"tasks"}}},
	}}
	resolution, err := ResolveLegacyAccess(context.Background(), store, LegacySubject{UserId: userId, HomeOrganisationId: home, OrganisationId: home, Role: "Operator"}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	project := DefaultProjectId(home)
	if !resolution.Access.Project(PermissionCasesRead, project) || resolution.Access.Project(PermissionAnalyticsRead, project) {
		t.Error("custom role pages not applied")
	}
	if len(resolution.Reasons) > 0 {
		t.Errorf("reasons %v, want none", resolution.Reasons)
	}
}

func TestResolveLegacyAccessCustomRoleReasons(t *testing.T) {
	home, userId := primitive.NewObjectID(), primitive.NewObjectID()
	allDay := LegacyTimeWindow{TimeRange1MaxMonday: 24, TimeRange1MaxTuesday: 24, TimeRange1MaxWednesday: 24,
		TimeRange1MaxThursday: 24, TimeRange1MaxFriday: 24, TimeRange1MaxSaturday: 24, TimeRange1MaxSunday: 24}
	first := LegacyRole{Id: primitive.NewObjectID(), Name: "Operator", ParentRole: "guest", Pages: []string{"tasks"}, TimeWindowActive: 1, TimeWindow: allDay}
	second := LegacyRole{Id: primitive.NewObjectID(), Name: "Operator", ParentRole: "admin", Pages: []string{"tasks"}}
	store := &fakeLegacyAccessStore{roles: map[string][]LegacyRole{"Operator": {first, second}}}
	// No owner timezone: the window is evaluated in the legacy default.
	resolution, err := ResolveLegacyAccess(context.Background(), store, LegacySubject{UserId: userId, HomeOrganisationId: home, OrganisationId: home, Role: "Operator"}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(resolution.Reasons, []LegacyReason{LegacyReasonDuplicateCustomRole, LegacyReasonTimeWindow}) {
		t.Errorf("reasons %v", resolution.Reasons)
	}
	project := DefaultProjectId(home)
	if !resolution.Access.Project(PermissionCasesRead, project) {
		t.Error("an all-day window in the default timezone denied")
	}
	if resolution.Access.Project(PermissionCasesDelete, project) {
		t.Error("the second role was applied; legacy applies the first")
	}
}

func TestResolveLegacyAccessOtherOrganisations(t *testing.T) {
	home, owned, foreign, userId := primitive.NewObjectID(), primitive.NewObjectID(), primitive.NewObjectID(), primitive.NewObjectID()
	store := &fakeLegacyAccessStore{owners: map[primitive.ObjectID]primitive.ObjectID{owned: userId, foreign: primitive.NewObjectID()}}

	resolution, err := ResolveLegacyAccess(context.Background(), store, LegacySubject{UserId: userId, HomeOrganisationId: home, OrganisationId: owned, Role: "guest"}, time.Now())
	if err != nil || !resolution.Access.IsOwner() {
		t.Errorf("owner of a created organisation: owner %v, err %v", resolution.Access.IsOwner(), err)
	}
	resolution, err = ResolveLegacyAccess(context.Background(), store, LegacySubject{UserId: userId, HomeOrganisationId: home, OrganisationId: foreign, Role: "admin"}, time.Now())
	if err != nil || len(resolution.Access.Permissions(DefaultProjectId(foreign))) > 0 {
		t.Errorf("another organisation granted %v, err %v", resolution.Access.Permissions(DefaultProjectId(foreign)), err)
	}
	if !slices.Equal(resolution.Reasons, []LegacyReason{LegacyReasonOtherOrganisation}) {
		t.Errorf("reasons %v", resolution.Reasons)
	}
}

func TestResolveLegacyAccessFailuresGrantNothing(t *testing.T) {
	home, userId := primitive.NewObjectID(), primitive.NewObjectID()
	tests := []struct {
		name  string
		store *fakeLegacyAccessStore
		role  string
	}{
		{"store error", &fakeLegacyAccessStore{err: errors.New("down")}, "Operator"},
		{"missing custom role", &fakeLegacyAccessStore{}, "Operator"},
		{"application user", &fakeLegacyAccessStore{}, "application"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resolution, err := ResolveLegacyAccess(context.Background(), test.store, LegacySubject{UserId: userId, HomeOrganisationId: home, OrganisationId: home, Role: test.role}, time.Now())
			if err == nil {
				t.Error("no error")
			}
			if got := resolution.Access.Permissions(DefaultProjectId(home)); len(got) > 0 {
				t.Errorf("granted %v", got)
			}
			if !slices.Contains(resolution.Reasons, LegacyReasonUnresolved) {
				t.Errorf("reasons %v, want unresolved", resolution.Reasons)
			}
		})
	}
}

func TestResolveLegacyAccessNoRoleKeepsMembership(t *testing.T) {
	home, userId := primitive.NewObjectID(), primitive.NewObjectID()
	store := &fakeLegacyAccessStore{}
	subject := LegacySubject{UserId: userId, HomeOrganisationId: home, OrganisationId: home, Cameras: []string{"cam-1"}}
	resolution, err := ResolveLegacyAccess(context.Background(), store, subject, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if store.reads != 0 {
		t.Error("an empty role must not be looked up")
	}
	if got := resolution.Access.Permissions(DefaultProjectId(home)); len(got) > 0 {
		t.Errorf("granted %v", got)
	}
	if selection := resolution.Access.Devices(PermissionDevicesRead, DefaultProjectId(home)); selection.All || len(selection.DeviceKeys) > 0 {
		t.Errorf("devices %+v", selection)
	}
	if !slices.Equal(resolution.Reasons, []LegacyReason{LegacyReasonNoRole}) {
		t.Errorf("reasons %v, want noRole", resolution.Reasons)
	}
}

func TestNewLegacySubject(t *testing.T) {
	owner, sub, selected := primitive.NewObjectID(), primitive.NewObjectID(), primitive.NewObjectID()
	user := User{Id: sub, Role: "Operator", Timezone: "America/New_York", Sites: []string{"s"}, Groups: []string{"g"}, Cameras: []string{"c"}}
	user.MasterAccount = owner.Hex()
	user.Master = &User{Id: owner, Timezone: "Asia/Tokyo"}
	subject := NewLegacySubject(user)
	if subject.UserId != sub || subject.HomeOrganisationId != owner || subject.OrganisationId != owner || subject.Role != "Operator" ||
		subject.OwnerTimezone != "Asia/Tokyo" || !slices.Equal(subject.Cameras, []string{"c"}) {
		t.Errorf("sub-account subject %+v", subject)
	}
	user.OrganisationId = selected
	if subject := NewLegacySubject(user); subject.OrganisationId != selected || subject.HomeOrganisationId != owner {
		t.Errorf("selected organisation subject %+v", subject)
	}
	if subject := NewLegacySubject(User{Id: owner, Role: "owner", Timezone: "Europe/Paris"}); subject.HomeOrganisationId != owner ||
		subject.OrganisationId != owner || subject.OwnerTimezone != "Europe/Paris" {
		t.Errorf("owner subject %+v", subject)
	}
}
