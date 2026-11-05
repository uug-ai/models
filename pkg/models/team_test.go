package models

import (
	"errors"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestTeamValidate(t *testing.T) {
	organisationId := primitive.NewObjectID()
	valid := func(mutate func(*Team)) Team {
		team := Team{Id: primitive.NewObjectID(), OrganisationId: organisationId, Name: "Control room", Source: TeamSourceManual}
		mutate(&team)
		return team
	}
	synced := func(team *Team) {
		team.Source = TeamSourceIdentityProvider
		team.IdentityProvider = &TeamIdentityProvider{Provider: "entra-id", ExternalGroupId: "5f1c0e7a-0000-4000-8000-000000000001"}
	}
	for _, test := range []struct {
		name  string
		team  Team
		valid bool
	}{
		{"manual", valid(func(*Team) {}), true},
		{"synced", valid(synced), true},
		{"missing organisation", valid(func(team *Team) { team.OrganisationId = primitive.NilObjectID }), false},
		{"missing name", valid(func(team *Team) { team.Name = "  " }), false},
		{"name with spaces around", valid(func(team *Team) { team.Name = " Control room" }), false},
		{"name too long", valid(func(team *Team) { team.Name = strings.Repeat("é", maxTeamNameLength+1) }), false},
		{"name at the limit", valid(func(team *Team) { team.Name = strings.Repeat("é", maxTeamNameLength) }), true},
		{"unknown source", valid(func(team *Team) { team.Source = "scim" }), false},
		{"missing source", valid(func(team *Team) { team.Source = "" }), false},
		{"manual with a group", valid(func(team *Team) {
			team.IdentityProvider = &TeamIdentityProvider{Provider: "okta", ExternalGroupId: "00g1"}
		}), false},
		{"synced without a group", valid(func(team *Team) { team.Source = TeamSourceIdentityProvider }), false},
		{"synced without a provider", valid(func(team *Team) {
			synced(team)
			team.IdentityProvider.Provider = " "
		}), false},
		{"synced without a group id", valid(func(team *Team) {
			synced(team)
			team.IdentityProvider.ExternalGroupId = ""
		}), false},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := test.team.Validate()
			if test.valid && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !test.valid && !errors.Is(err, ErrTeamInvalid) {
				t.Fatalf("expected ErrTeamInvalid, got %v", err)
			}
		})
	}
}

func TestTeamManagesMembersInHub(t *testing.T) {
	if !(Team{Source: TeamSourceManual}).ManagesMembersInHub() {
		t.Error("manual team members are managed in Hub")
	}
	if (Team{Source: TeamSourceIdentityProvider}).ManagesMembersInHub() {
		t.Error("synced team members come from the identity provider only")
	}
}

func TestTeamMemberValidate(t *testing.T) {
	team := Team{Id: primitive.NewObjectID(), OrganisationId: primitive.NewObjectID(), Name: "Control room", Source: TeamSourceManual}
	valid := func(mutate func(*TeamMember)) TeamMember {
		member := TeamMember{OrganisationId: team.OrganisationId, TeamId: team.Id, UserId: primitive.NewObjectID()}
		mutate(&member)
		return member
	}
	for _, test := range []struct {
		name   string
		member TeamMember
		valid  bool
	}{
		{"member", valid(func(*TeamMember) {}), true},
		{"missing user", valid(func(member *TeamMember) { member.UserId = primitive.NilObjectID }), false},
		{"missing team", valid(func(member *TeamMember) { member.TeamId = primitive.NilObjectID }), false},
		{"another team", valid(func(member *TeamMember) { member.TeamId = primitive.NewObjectID() }), false},
		{"missing organisation", valid(func(member *TeamMember) { member.OrganisationId = primitive.NilObjectID }), false},
		{"another organisation", valid(func(member *TeamMember) { member.OrganisationId = primitive.NewObjectID() }), false},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := test.member.Validate(team)
			if test.valid && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !test.valid && !errors.Is(err, ErrTeamInvalid) {
				t.Fatalf("expected ErrTeamInvalid, got %v", err)
			}
		})
	}
}

func TestTeamPermissionsInCatalog(t *testing.T) {
	for _, permission := range TeamPermissions() {
		definition, ok := LookupPermission(permission)
		if !ok {
			t.Fatalf("%s missing from the catalog", permission)
		}
		if definition.Level != LevelOrganisation {
			t.Errorf("%s: level %s, want organisation", permission, definition.Level)
		}
		// Reading teams can be limited to selected projects; changing a team
		// changes access in every project its assignments cover.
		if want := permission != PermissionTeamsRead; definition.TenantWide != want {
			t.Errorf("%s: tenantWide %v, want %v", permission, definition.TenantWide, want)
		}
	}
}

// teamFixture is an organisation with one member in one team, the team's
// project role and assignment, and nothing granted to the user directly.
type teamFixture struct {
	organisationId, userId, project primitive.ObjectID
	team                            Team
	membership                      OrganisationUser
	teamMembers                     []TeamMember
	role                            Role
	assignment                      RoleAssignment
}

func newTeamFixture() teamFixture {
	f := teamFixture{organisationId: primitive.NewObjectID(), userId: primitive.NewObjectID(), project: primitive.NewObjectID()}
	f.team = Team{Id: primitive.NewObjectID(), OrganisationId: f.organisationId, Name: "Control room", Source: TeamSourceManual}
	f.membership = OrganisationUser{UserId: f.userId, OrganisationId: f.organisationId, Status: MembershipStatusActive}
	f.teamMembers = []TeamMember{{OrganisationId: f.organisationId, TeamId: f.team.Id, UserId: f.userId}}
	f.role = Role{Id: primitive.NewObjectID(), OrganisationId: f.organisationId, Name: "Operator", Kind: RoleKindCustom,
		AssignableAt: LevelProject, Grants: SelectionSelected, Permissions: []Permission{PermissionDevicesRead, PermissionLiveStream}, IsActive: true}
	f.assignment = RoleAssignment{Id: primitive.NewObjectID(), OrganisationId: f.organisationId,
		Principal: Principal{Type: PrincipalTypeTeam, Id: f.team.Id}, RoleId: f.role.Id,
		Scope:  RoleAssignmentScope{Level: LevelProject, ProjectId: f.project, Resources: SelectionSelected, DeviceKeys: []string{"camera-1"}},
		Source: AssignmentSourceManual, IsActive: true}
	return f
}

func (f teamFixture) access() Access {
	return NewAccess(f.membership, f.teamMembers, []Role{f.role}, []RoleAssignment{f.assignment}, time.Now())
}

func TestAccessThroughTeam(t *testing.T) {
	f := newTeamFixture()
	access := f.access()
	if !access.Project(PermissionLiveStream, f.project) {
		t.Fatal("team member denied the team's permission")
	}
	if got := access.Devices(PermissionDevicesRead, f.project); got.All || len(got.DeviceKeys) != 1 || got.DeviceKeys[0] != "camera-1" {
		t.Errorf("devices %+v, want the team's selection", got)
	}
	if access.Project(PermissionLiveStream, primitive.NewObjectID()) {
		t.Error("team grant leaked into another project")
	}
}

func TestAccessThroughTeamAddsUpWithDirectAssignments(t *testing.T) {
	f := newTeamFixture()
	direct := f.assignment
	direct.Id = primitive.NewObjectID()
	direct.Principal = Principal{Type: PrincipalTypeUser, Id: f.userId}
	direct.Scope.DeviceKeys = []string{"camera-2"}
	access := NewAccess(f.membership, f.teamMembers, []Role{f.role}, []RoleAssignment{f.assignment, direct}, time.Now())
	got := access.Devices(PermissionDevicesRead, f.project)
	if len(got.DeviceKeys) != 2 {
		t.Errorf("devices %+v, want the union of the team's and the user's", got)
	}
}

func TestAccessThroughTeamDenies(t *testing.T) {
	for _, test := range []struct {
		name   string
		mutate func(*teamFixture)
	}{
		{"not in the team", func(f *teamFixture) { f.teamMembers = nil }},
		{"membership suspended", func(f *teamFixture) { f.membership.Status = MembershipStatusSuspended }},
		{"membership expired", func(f *teamFixture) { f.membership.ExpiresAt = time.Now().Add(-time.Minute) }},
		{"team member of another user", func(f *teamFixture) { f.teamMembers[0].UserId = primitive.NewObjectID() }},
		{"team member in another organisation", func(f *teamFixture) { f.teamMembers[0].OrganisationId = primitive.NewObjectID() }},
		{"another team", func(f *teamFixture) { f.teamMembers[0].TeamId = primitive.NewObjectID() }},
		{"assignment disabled", func(f *teamFixture) { f.assignment.IsActive = false }},
		{"assignment in another organisation", func(f *teamFixture) { f.assignment.OrganisationId = primitive.NewObjectID() }},
		{"user id used as a team id", func(f *teamFixture) { f.assignment.Principal = Principal{Type: PrincipalTypeTeam, Id: f.userId} }},
		{"team id used as a user id", func(f *teamFixture) { f.assignment.Principal = Principal{Type: PrincipalTypeUser, Id: f.team.Id} }},
	} {
		t.Run(test.name, func(t *testing.T) {
			f := newTeamFixture()
			test.mutate(&f)
			if f.access().Project(PermissionLiveStream, f.project) {
				t.Error("granted")
			}
		})
	}
}

func TestOwnerNeverThroughTeam(t *testing.T) {
	f := newTeamFixture()
	owner, _ := BuiltInRole(RoleKeyOwner)
	owner.Id, owner.OrganisationId = primitive.NewObjectID(), f.organisationId
	assignment := RoleAssignment{OrganisationId: f.organisationId, Principal: Principal{Type: PrincipalTypeTeam, Id: f.team.Id},
		RoleId: owner.Id, Scope: RoleAssignmentScope{Level: LevelOrganisation, Projects: SelectionAll},
		Source: AssignmentSourceManual, IsActive: true}
	if err := assignment.Validate(owner); !errors.Is(err, ErrRoleAssignmentInvalid) {
		t.Fatalf("owner assignment to a team: %v, want ErrRoleAssignmentInvalid", err)
	}
	access := NewAccess(f.membership, f.teamMembers, []Role{owner}, []RoleAssignment{assignment}, time.Now())
	if access.IsOwner() {
		t.Error("team member became owner")
	}
}

func TestTeamAssignmentValidate(t *testing.T) {
	f := newTeamFixture()
	if err := f.assignment.Validate(f.role); err != nil {
		t.Fatalf("team assignment: %v", err)
	}
	organisation := Role{Id: primitive.NewObjectID(), OrganisationId: f.organisationId, Name: "Monitoring desk", Kind: RoleKindCustom,
		AssignableAt: LevelOrganisation, Grants: SelectionSelected, Permissions: []Permission{PermissionMembersRead}, IsActive: true}
	assignment := RoleAssignment{OrganisationId: f.organisationId, Principal: Principal{Type: PrincipalTypeTeam, Id: f.team.Id},
		RoleId: organisation.Id, Scope: RoleAssignmentScope{Level: LevelOrganisation, Projects: SelectionAll},
		Source: AssignmentSourceManual, IsActive: true}
	if err := assignment.Validate(organisation); err != nil {
		t.Fatalf("organisation role for a team: %v", err)
	}
	assignment.Principal.Id = primitive.NilObjectID
	if err := assignment.Validate(organisation); !errors.Is(err, ErrRoleAssignmentInvalid) {
		t.Errorf("team assignment without a team: %v", err)
	}
}

func TestRoleAssignmentAffectsUser(t *testing.T) {
	f := newTeamFixture()
	other := primitive.NewObjectID()
	direct := RoleAssignment{OrganisationId: f.organisationId, Principal: Principal{Type: PrincipalTypeUser, Id: f.userId}}
	for _, test := range []struct {
		name       string
		assignment RoleAssignment
		userId     primitive.ObjectID
		teams      []TeamMember
		want       bool
	}{
		{"own direct assignment", direct, f.userId, nil, true},
		{"someone else's direct assignment", direct, other, nil, false},
		{"zero user", RoleAssignment{Principal: Principal{Type: PrincipalTypeUser}}, primitive.NilObjectID, nil, false},
		{"own team", f.assignment, f.userId, f.teamMembers, true},
		{"team without the user", f.assignment, other, f.teamMembers, false},
		{"team membership in another organisation", f.assignment, f.userId,
			[]TeamMember{{OrganisationId: primitive.NewObjectID(), TeamId: f.team.Id, UserId: f.userId}}, false},
		{"unknown principal type", RoleAssignment{Principal: Principal{Type: "group", Id: f.userId}}, f.userId, nil, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := test.assignment.AffectsUser(test.userId, test.teams); got != test.want {
				t.Errorf("AffectsUser = %v, want %v", got, test.want)
			}
		})
	}
}

func TestTeamPatch(t *testing.T) {
	team := Team{Id: primitive.NewObjectID(), OrganisationId: primitive.NewObjectID(), Name: "Control room", Description: "Night shift",
		Source: TeamSourceIdentityProvider, IdentityProvider: &TeamIdentityProvider{Provider: "okta", ExternalGroupId: "00g1"}}
	if !(TeamPatch{}).IsEmpty() || (TeamPatch{Description: new(string)}).IsEmpty() {
		t.Fatal("IsEmpty")
	}
	if got := (TeamPatch{}).Apply(team); got.Name != team.Name || got.Description != team.Description {
		t.Errorf("empty patch changed the team: %+v", got)
	}
	name, description := "Control room Ghent", ""
	got := TeamPatch{Name: &name, Description: &description}.Apply(team)
	if got.Name != name || got.Description != "" {
		t.Errorf("patched %+v", got)
	}
	if got.Id != team.Id || got.OrganisationId != team.OrganisationId || got.Source != team.Source || got.IdentityProvider != team.IdentityProvider {
		t.Error("patch changed a fixed field")
	}
}
