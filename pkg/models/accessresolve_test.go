package models

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// memoryAccessStore serves stored documents from memory, filtering the way a
// database store does: by organisation, user, principal and role id.
type memoryAccessStore struct {
	memberships []OrganisationUser
	teams       []TeamMember
	assignments []RoleAssignment
	roles       []Role
	err         error
	failOn      string
}

func (s *memoryAccessStore) fail(read string) error {
	if s.failOn == read {
		return s.err
	}
	return nil
}

func (s *memoryAccessStore) Membership(_ context.Context, organisationId, userId primitive.ObjectID) (OrganisationUser, bool, error) {
	if err := s.fail("membership"); err != nil {
		return OrganisationUser{}, false, err
	}
	for _, membership := range s.memberships {
		if membership.OrganisationId == organisationId && membership.UserId == userId {
			return membership, true, nil
		}
	}
	return OrganisationUser{}, false, nil
}

func (s *memoryAccessStore) TeamMemberships(_ context.Context, organisationId, userId primitive.ObjectID) ([]TeamMember, error) {
	if err := s.fail("teams"); err != nil {
		return nil, err
	}
	var found []TeamMember
	for _, member := range s.teams {
		if member.OrganisationId == organisationId && member.UserId == userId {
			found = append(found, member)
		}
	}
	return found, nil
}

func (s *memoryAccessStore) Assignments(_ context.Context, organisationId primitive.ObjectID, principals []Principal) ([]RoleAssignment, error) {
	if err := s.fail("assignments"); err != nil {
		return nil, err
	}
	var found []RoleAssignment
	for _, assignment := range s.assignments {
		if assignment.OrganisationId == organisationId && slices.Contains(principals, assignment.Principal) {
			found = append(found, assignment)
		}
	}
	return found, nil
}

func (s *memoryAccessStore) Roles(_ context.Context, organisationId primitive.ObjectID, roleIds []primitive.ObjectID) ([]Role, error) {
	if err := s.fail("roles"); err != nil {
		return nil, err
	}
	var found []Role
	for _, role := range s.roles {
		if role.OrganisationId == organisationId && slices.Contains(roleIds, role.Id) {
			found = append(found, role)
		}
	}
	return found, nil
}

// storeOf stores what MapLegacyAccess produced, as the migration will.
func storeOf(mapped LegacyAccess) *memoryAccessStore {
	return &memoryAccessStore{memberships: []OrganisationUser{mapped.Membership}, assignments: mapped.Assignments, roles: mapped.Roles}
}

func TestEffectiveAuthorizationMode(t *testing.T) {
	for stored, want := range map[AuthorizationMode]AuthorizationMode{
		"":                      AuthorizationModeLegacy,
		AuthorizationModeLegacy: AuthorizationModeLegacy,
		AuthorizationModeShadow: AuthorizationModeShadow,
		AuthorizationModeRBAC:   AuthorizationModeRBAC,
		"enforce":               AuthorizationModeLegacy,
	} {
		organisation := Organisation{AuthorizationMode: stored}
		if got := organisation.EffectiveAuthorizationMode(); got != want {
			t.Errorf("%q: got %q, want %q", stored, got, want)
		}
		if err := organisation.ValidateAuthorizationMode(); (err == nil) != stored.IsValid() {
			t.Errorf("%q: ValidateAuthorizationMode = %v", stored, err)
		}
	}
	if (AuthorizationMode("enforce")).IsValid() {
		t.Error("unknown mode is valid")
	}
}

// The storage comparison expects stored documents written from legacy data to
// resolve exactly as the legacy data does.
func TestResolveAccessMatchesLegacyResolution(t *testing.T) {
	ids := legacyIds(5)
	organisationId, userId, site, group := ids[0], ids[1], ids[2], ids[3]
	project := DefaultProjectId(organisationId)
	for _, in := range []LegacyUserAccess{
		{Role: "owner"},
		{Role: "admin", SiteIds: []primitive.ObjectID{site}},
		{Role: "editor", GroupIds: []primitive.ObjectID{group}, DeviceKeys: []string{"camera-1"}},
		{Role: "guest", DeviceKeys: []string{"camera-1", "camera-2"}},
		{Role: "guest"},
		{Role: "Operator", SiteIds: []primitive.ObjectID{site}, CustomRole: &LegacyRole{
			Id: ids[4], Name: "Operator", ParentRole: "editor", Pages: []string{"live-view", "recordings", "tasks"},
			FeaturePermissions: LegacyFeatureLevels{PTZ: 2, Liveview: 2}}},
	} {
		t.Run(in.Role, func(t *testing.T) {
			in.OrganisationId, in.UserId, in.Timezone = organisationId, userId, LegacyDefaultTimezone
			mapped, err := MapLegacyAccess(in)
			if err != nil {
				t.Fatal(err)
			}
			at := time.Now()
			legacy := NewAccess(mapped.Membership, nil, mapped.Roles, mapped.Assignments, at)
			stored, err := ResolveAccess(context.Background(), storeOf(mapped), organisationId, userId, at)
			if err != nil {
				t.Fatal(err)
			}
			if differences := CompareAccess(stored, legacy, []primitive.ObjectID{project}); differences != nil {
				t.Errorf("stored access differs: %+v", differences)
			}
		})
	}
}

func TestResolveAccessThroughTeams(t *testing.T) {
	f := newTeamFixture()
	store := &memoryAccessStore{
		memberships: []OrganisationUser{f.membership},
		teams:       f.teamMembers,
		assignments: []RoleAssignment{f.assignment},
		roles:       []Role{f.role},
	}
	access, err := ResolveAccess(context.Background(), store, f.organisationId, f.userId, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if !access.Project(PermissionLiveStream, f.project) {
		t.Error("team assignment not resolved")
	}
	store.teams = nil
	if access, _ := ResolveAccess(context.Background(), store, f.organisationId, f.userId, time.Now()); access.Project(PermissionLiveStream, f.project) {
		t.Error("team assignment granted without the team membership")
	}
}

func TestResolveAccessFailsClosed(t *testing.T) {
	f := newTeamFixture()
	full := func() *memoryAccessStore {
		return &memoryAccessStore{
			memberships: []OrganisationUser{f.membership},
			teams:       f.teamMembers,
			assignments: []RoleAssignment{f.assignment},
			roles:       []Role{f.role},
		}
	}
	for _, read := range []string{"membership", "teams", "assignments", "roles"} {
		t.Run(read+" read error", func(t *testing.T) {
			store := full()
			store.err, store.failOn = errors.New("read failed"), read
			access, err := ResolveAccess(context.Background(), store, f.organisationId, f.userId, time.Now())
			if err == nil || access.Project(PermissionLiveStream, f.project) {
				t.Errorf("err = %v, granted = %v", err, access.Project(PermissionLiveStream, f.project))
			}
		})
	}
	for name, mutate := range map[string]func(*memoryAccessStore){
		"no membership":        func(s *memoryAccessStore) { s.memberships = nil },
		"suspended membership": func(s *memoryAccessStore) { s.memberships[0].Status = MembershipStatusSuspended },
		"role missing":         func(s *memoryAccessStore) { s.roles = nil },
		"other organisation":   func(s *memoryAccessStore) { s.assignments[0].OrganisationId = primitive.NewObjectID() },
	} {
		t.Run(name, func(t *testing.T) {
			store := full()
			mutate(store)
			access, err := ResolveAccess(context.Background(), store, f.organisationId, f.userId, time.Now())
			if err != nil || access.Project(PermissionLiveStream, f.project) {
				t.Errorf("err = %v, granted = %v", err, access.Project(PermissionLiveStream, f.project))
			}
		})
	}
	if _, err := ResolveAccess(context.Background(), full(), primitive.NilObjectID, f.userId, time.Now()); err == nil {
		t.Error("zero organisation accepted")
	}
}

func TestCompareAccess(t *testing.T) {
	ids := legacyIds(4)
	organisationId, userId, site, otherSite := ids[0], ids[1], ids[2], ids[3]
	project := DefaultProjectId(organisationId)
	resolve := func(in LegacyUserAccess) Access {
		t.Helper()
		in.OrganisationId, in.UserId, in.Timezone = organisationId, userId, LegacyDefaultTimezone
		mapped, err := MapLegacyAccess(in)
		if err != nil {
			t.Fatal(err)
		}
		return NewAccess(mapped.Membership, nil, mapped.Roles, mapped.Assignments, time.Now())
	}
	guestOnSite := resolve(LegacyUserAccess{Role: "guest", SiteIds: []primitive.ObjectID{site}})
	has := func(differences []AccessDifference, kind string, permission Permission, wider bool) bool {
		return slices.ContainsFunc(differences, func(d AccessDifference) bool {
			return d.Kind == kind && (permission == "" || d.Permission == permission) && d.Wider == wider
		})
	}

	if differences := CompareAccess(guestOnSite, guestOnSite, []primitive.ObjectID{project}); differences != nil {
		t.Errorf("identical access differs: %+v", differences)
	}
	owner := resolve(LegacyUserAccess{Role: "owner"})
	if differences := CompareAccess(owner, guestOnSite, []primitive.ObjectID{project}); !has(differences, "owner", "", true) || !has(differences, "permission", PermissionDevicesDelete, true) {
		t.Errorf("owner vs guest: %+v", differences)
	}
	if differences := CompareAccess(guestOnSite, owner, []primitive.ObjectID{project}); !has(differences, "owner", "", false) || has(differences, "owner", "", true) {
		t.Errorf("guest vs owner: %+v", differences)
	}
	editorOnSite := resolve(LegacyUserAccess{Role: "editor", SiteIds: []primitive.ObjectID{site}})
	if differences := CompareAccess(editorOnSite, guestOnSite, []primitive.ObjectID{project}); !has(differences, "permission", "", true) || has(differences, "permission", "", false) {
		t.Errorf("editor vs guest: %+v", differences)
	}
	guestOnOtherSite := resolve(LegacyUserAccess{Role: "guest", SiteIds: []primitive.ObjectID{otherSite}})
	if differences := CompareAccess(guestOnOtherSite, guestOnSite, []primitive.ObjectID{project}); !has(differences, "devices", PermissionDevicesRead, true) || !has(differences, "devices", PermissionDevicesRead, false) {
		t.Errorf("other site: %+v", differences)
	}
	guestOnBoth := resolve(LegacyUserAccess{Role: "guest", SiteIds: []primitive.ObjectID{site, otherSite}})
	if differences := CompareAccess(guestOnBoth, guestOnSite, []primitive.ObjectID{project}); !has(differences, "devices", PermissionDevicesRead, true) || has(differences, "devices", PermissionDevicesRead, false) {
		t.Errorf("both sites vs one: %+v", differences)
	}
}
