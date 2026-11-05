package models

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// grantFixture is one organisation with two projects. Project one has site A
// (devices a1, a2, plus camera group X with x1) and site B (device b1).
type grantFixture struct {
	organisationId, userId     primitive.ObjectID
	projectOne, projectTwo     primitive.ObjectID
	siteA, siteB, cameraGroupX primitive.ObjectID
	emptySite                  primitive.ObjectID
	source                     *fakeDeviceSource
}

func newGrantFixture() grantFixture {
	f := grantFixture{
		organisationId: primitive.NewObjectID(), userId: primitive.NewObjectID(),
		projectOne: primitive.NewObjectID(), projectTwo: primitive.NewObjectID(),
		siteA: primitive.NewObjectID(), siteB: primitive.NewObjectID(), cameraGroupX: primitive.NewObjectID(),
		emptySite: primitive.NewObjectID(),
	}
	f.source = &fakeDeviceSource{
		sites: map[string]Site{
			f.siteA.Hex():     {Devices: []string{"a1", "a2"}, Groups: []string{f.cameraGroupX.Hex()}},
			f.siteB.Hex():     {Devices: []string{"b1"}},
			f.emptySite.Hex(): {},
		},
		groups: map[string]Group{
			f.cameraGroupX.Hex(): {GroupType: GroupTypeCamera, Devices: []string{"x1"}},
		},
	}
	return f
}

func (f grantFixture) access(t *testing.T, grants map[*Role]RoleAssignmentScope) Access {
	t.Helper()
	return accessWith(t, f.organisationId, f.userId, grants)
}

func (f grantFixture) projectRole(permissions ...Permission) Role {
	return Role{Id: primitive.NewObjectID(), OrganisationId: f.organisationId, Name: "Granted", Kind: RoleKindCustom,
		AssignableAt: LevelProject, Grants: SelectionSelected, Permissions: permissions, IsActive: true}
}

func (f grantFixture) organisationRole(permissions ...Permission) Role {
	role := f.projectRole(permissions...)
	role.AssignableAt = LevelOrganisation
	return role
}

func (f grantFixture) selected(project primitive.ObjectID, sites, groups []primitive.ObjectID, keys ...string) RoleAssignmentScope {
	return RoleAssignmentScope{Level: LevelProject, ProjectId: project, Resources: SelectionSelected, SiteIds: sites, GroupIds: groups, DeviceKeys: keys}
}

func (f grantFixture) allResources(project primitive.ObjectID) RoleAssignmentScope {
	return RoleAssignmentScope{Level: LevelProject, ProjectId: project, Resources: SelectionAll}
}

func selectedProjects(ids ...primitive.ObjectID) RoleAssignmentScope {
	return RoleAssignmentScope{Level: LevelOrganisation, Projects: SelectionSelected, ProjectIds: ids}
}

func mustGrant(t *testing.T, access Access, source DeviceSource, role Role, scope RoleAssignmentScope) GrantDecision {
	t.Helper()
	decision, err := access.CanGrant(context.Background(), source, role, scope)
	if err != nil {
		t.Fatalf("CanGrant: %v", err)
	}
	return decision
}

func ownerAccess(t *testing.T, f grantFixture) Access {
	t.Helper()
	owner, _ := BuiltInRole(RoleKeyOwner)
	owner.Id, owner.OrganisationId = primitive.NewObjectID(), f.organisationId
	membership := OrganisationUser{UserId: f.userId, OrganisationId: f.organisationId, Status: MembershipStatusActive}
	assignment := RoleAssignment{Id: primitive.NewObjectID(), OrganisationId: f.organisationId, Principal: Principal{Type: PrincipalTypeUser, Id: f.userId},
		RoleId: owner.Id, Scope: allProjects(), Source: AssignmentSourceManual, IsActive: true}
	return NewAccess(membership, nil, []Role{owner}, []RoleAssignment{assignment}, time.Now())
}

func TestCanGrantOwner(t *testing.T) {
	f := newGrantFixture()
	ownerRole, _ := BuiltInRole(RoleKeyOwner)
	ownerRole.Id, ownerRole.OrganisationId = primitive.NewObjectID(), f.organisationId

	owner := ownerAccess(t, f)
	for name, test := range map[string]struct {
		role  Role
		scope RoleAssignmentScope
	}{
		"Owner":                     {ownerRole, allProjects()},
		"organisation role":         {f.organisationRole(PermissionRolesCreate, PermissionBillingManage), allProjects()},
		"project role on all":       {f.projectRole(PermissionDevicesDelete), f.allResources(f.projectOne)},
		"project role on any sites": {f.projectRole(PermissionDevicesDelete), f.selected(f.projectOne, []primitive.ObjectID{f.siteB}, nil)},
	} {
		t.Run(name, func(t *testing.T) {
			if decision := mustGrant(t, owner, f.source, test.role, test.scope); !decision.Allowed {
				t.Errorf("owner denied: %+v", decision)
			}
		})
	}

	manager := f.access(t, map[*Role]RoleAssignmentScope{
		{Name: "Manager", AssignableAt: LevelOrganisation, Permissions: []Permission{PermissionMembersUpdate, PermissionProjectsManageAccess}}: allProjects(),
	})
	if decision := mustGrant(t, manager, f.source, ownerRole, allProjects()); decision.Allowed || decision.Reason != GrantDenialOwnerOnly {
		t.Errorf("non-owner granting Owner: %+v", decision)
	}
}

func TestCanGrantOrganisationRole(t *testing.T) {
	f := newGrantFixture()
	manager := f.access(t, map[*Role]RoleAssignmentScope{
		{Name: "Members", AssignableAt: LevelOrganisation, Permissions: []Permission{PermissionMembersRead, PermissionMembersUpdate}}: allProjects(),
	})
	tests := []struct {
		name       string
		access     Access
		role       Role
		scope      RoleAssignmentScope
		reason     GrantDenial
		permission Permission
	}{
		{name: "held", access: manager, role: f.organisationRole(PermissionMembersRead), scope: allProjects()},
		{name: "not held, on all projects", access: manager, role: f.organisationRole(PermissionRolesUpdate, PermissionTeamsUpdate), scope: allProjects()},
		{name: "not held, on selected projects", access: manager, role: f.organisationRole(PermissionProjectsManageAccess), scope: selectedProjects(f.projectOne, f.projectTwo)},
		{name: "owner-only", access: manager, role: f.organisationRole(PermissionMembersRead, PermissionStorageRead),
			scope: allProjects(), reason: GrantDenialOwnerOnly, permission: PermissionStorageRead},
		{name: "without members.update", access: f.access(t, map[*Role]RoleAssignmentScope{
			{Name: "Reader", AssignableAt: LevelOrganisation, Permissions: []Permission{PermissionMembersRead, PermissionProjectsManageAccess}}: allProjects(),
		}), role: f.organisationRole(PermissionMembersRead), scope: allProjects(), reason: GrantDenialNoAccessPermission},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			decision := mustGrant(t, test.access, f.source, test.role, test.scope)
			if test.reason == "" {
				if !decision.Allowed {
					t.Errorf("denied: %+v", decision)
				}
				return
			}
			if decision.Allowed || decision.Reason != test.reason || decision.Permission != test.permission {
				t.Errorf("decision %+v, want %s %s", decision, test.reason, test.permission)
			}
		})
	}
}

func TestCanGrantProjectRoleWithProjectsManageAccess(t *testing.T) {
	f := newGrantFixture()
	// An organisation delegate for project one, who sees no devices there
	// but holds devices.read and live.stream on site A.
	delegate := f.access(t, map[*Role]RoleAssignmentScope{
		{Name: "Delegate", AssignableAt: LevelOrganisation, Permissions: []Permission{PermissionProjectsManageAccess}}:       selectedProjects(f.projectOne),
		{Name: "Viewer", AssignableAt: LevelProject, Permissions: []Permission{PermissionDevicesRead, PermissionLiveStream}}: f.selected(f.projectOne, []primitive.ObjectID{f.siteA}, nil),
	})
	viewer := f.projectRole(PermissionDevicesRead)
	if decision := mustGrant(t, delegate, f.source, viewer, f.allResources(f.projectOne)); !decision.Allowed {
		t.Errorf("projects.manageAccess grants any resources: %+v", decision)
	}
	if decision := mustGrant(t, delegate, f.source, viewer, f.selected(f.projectOne, []primitive.ObjectID{f.siteB}, nil)); !decision.Allowed {
		t.Errorf("projects.manageAccess grants sites outside its own: %+v", decision)
	}
	if decision := mustGrant(t, delegate, f.source, f.projectRole(PermissionDevicesRead, PermissionDevicesUpdate, PermissionDevicesConfigure), f.allResources(f.projectOne)); !decision.Allowed {
		t.Errorf("permissions the delegate does not hold: %+v", decision)
	}
	if decision := mustGrant(t, delegate, f.source, f.projectRole(PermissionDevicesRead, PermissionDevicesDelete), f.allResources(f.projectOne)); decision.Allowed ||
		decision.Reason != GrantDenialOwnerOnly || decision.Permission != PermissionDevicesDelete {
		t.Errorf("owner-only permission: %+v", decision)
	}
	if decision := mustGrant(t, delegate, f.source, viewer, f.allResources(f.projectTwo)); decision.Allowed || decision.Reason != GrantDenialNoAccessPermission {
		t.Errorf("project not covered: %+v", decision)
	}
}

func TestCanGrantProjectRoleWithProjectManageAccess(t *testing.T) {
	f := newGrantFixture()
	siteAdmin := f.access(t, map[*Role]RoleAssignmentScope{
		{Name: "Site admin", AssignableAt: LevelProject, Permissions: []Permission{PermissionDevicesRead, PermissionLiveStream, PermissionProjectManageAccess}}: f.selected(f.projectOne, []primitive.ObjectID{f.siteA}, nil),
	})
	viewer := f.projectRole(PermissionDevicesRead, PermissionLivePreview) // live.preview is implied by live.stream
	sites := func(ids ...primitive.ObjectID) []primitive.ObjectID { return ids }
	tests := []struct {
		name   string
		scope  RoleAssignmentScope
		reason GrantDenial
	}{
		{name: "its own site", scope: f.selected(f.projectOne, sites(f.siteA), nil)},
		{name: "a device of its site", scope: f.selected(f.projectOne, nil, nil, "a1")},
		{name: "a camera group inside its site", scope: f.selected(f.projectOne, nil, sites(f.cameraGroupX))},
		{name: "no devices yet", scope: f.selected(f.projectOne, nil, nil)},
		{name: "a site without devices", scope: f.selected(f.projectOne, sites(f.emptySite), nil)},
		{name: "another site", scope: f.selected(f.projectOne, sites(f.siteB), nil), reason: GrantDenialResourcesNotCovered},
		{name: "a device outside", scope: f.selected(f.projectOne, nil, nil, "a1", "b1"), reason: GrantDenialResourcesNotCovered},
		{name: "a device that does not exist", scope: f.selected(f.projectOne, nil, nil, "ghost"), reason: GrantDenialResourcesNotCovered},
		{name: "all resources", scope: f.allResources(f.projectOne), reason: GrantDenialResourcesNotCovered},
		{name: "another project", scope: f.selected(f.projectTwo, nil, nil), reason: GrantDenialNoAccessPermission},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			decision := mustGrant(t, siteAdmin, f.source, viewer, test.scope)
			if test.reason == "" && !decision.Allowed {
				t.Errorf("denied: %+v", decision)
			}
			if test.reason != "" && (decision.Allowed || decision.Reason != test.reason) {
				t.Errorf("decision %+v, want %s", decision, test.reason)
			}
		})
	}

	if decision := mustGrant(t, siteAdmin, f.source, f.projectRole(PermissionDevicesRead, PermissionDevicesUpdate), f.selected(f.projectOne, sites(f.siteA), nil)); !decision.Allowed {
		t.Errorf("a permission the grantor does not hold, on its own site: %+v", decision)
	}
	if decision := mustGrant(t, siteAdmin, f.source, f.projectRole(PermissionDevicesRead, PermissionDevicesReadLogs), f.selected(f.projectOne, sites(f.siteA), nil)); decision.Allowed ||
		decision.Reason != GrantDenialOwnerOnly || decision.Permission != PermissionDevicesReadLogs {
		t.Errorf("owner-only permission: %+v", decision)
	}

	projectAdmin := f.access(t, map[*Role]RoleAssignmentScope{
		{Name: "Project admin", AssignableAt: LevelProject, Permissions: []Permission{PermissionDevicesRead, PermissionProjectManageAccess}}: f.allResources(f.projectOne),
	})
	if decision := mustGrant(t, projectAdmin, f.source, f.projectRole(PermissionDevicesRead), f.allResources(f.projectOne)); !decision.Allowed {
		t.Errorf("project.manageAccess on all resources grants all: %+v", decision)
	}

	noManage := f.access(t, map[*Role]RoleAssignmentScope{
		{Name: "Viewer", AssignableAt: LevelProject, Permissions: []Permission{PermissionDevicesRead}}: f.allResources(f.projectOne),
	})
	if decision := mustGrant(t, noManage, f.source, f.projectRole(PermissionDevicesRead), f.selected(f.projectOne, nil, nil)); decision.Allowed || decision.Reason != GrantDenialNoAccessPermission {
		t.Errorf("without manageAccess: %+v", decision)
	}
}

func TestCanGrantErrors(t *testing.T) {
	f := newGrantFixture()
	siteAdmin := f.access(t, map[*Role]RoleAssignmentScope{
		{Name: "Site admin", AssignableAt: LevelProject, Permissions: []Permission{PermissionDevicesRead, PermissionProjectManageAccess}}: f.selected(f.projectOne, []primitive.ObjectID{f.siteA}, nil),
	})
	if _, err := siteAdmin.CanGrant(context.Background(), f.source, f.projectRole(PermissionDevicesRead), allProjects()); !errors.Is(err, ErrRoleAssignmentInvalid) {
		t.Errorf("organisation scope for a project role: %v", err)
	}
	if _, err := siteAdmin.CanGrant(context.Background(), f.source, f.projectRole(PermissionMembersRead), f.allResources(f.projectOne)); !errors.Is(err, ErrRoleAssignmentInvalid) {
		t.Errorf("project role with an organisation permission: %v", err)
	}
	failing := &fakeDeviceSource{err: errors.New("read failed")}
	if _, err := siteAdmin.CanGrant(context.Background(), failing, f.projectRole(PermissionDevicesRead), f.selected(f.projectOne, []primitive.ObjectID{f.siteA}, nil)); err == nil {
		t.Error("an expansion read error must fail the check")
	}
}

func TestCanChange(t *testing.T) {
	f := newGrantFixture()
	siteAdmin := f.access(t, map[*Role]RoleAssignmentScope{
		{Name: "Site admin", AssignableAt: LevelProject, Permissions: []Permission{PermissionDevicesRead, PermissionProjectManageAccess}}: f.selected(f.projectOne, []primitive.ObjectID{f.siteA}, nil),
	})
	viewer := f.projectRole(PermissionDevicesRead)
	inside := f.selected(f.projectOne, nil, nil, "a1")
	outside := f.selected(f.projectOne, []primitive.ObjectID{f.siteB}, nil)
	change := func(current, next *RoleAssignmentScope) GrantDecision {
		t.Helper()
		decision, err := siteAdmin.CanChange(context.Background(), f.source, viewer, current, next)
		if err != nil {
			t.Fatal(err)
		}
		return decision
	}
	if !change(nil, &inside).Allowed {
		t.Error("create inside denied")
	}
	if !change(&inside, nil).Allowed {
		t.Error("remove inside denied")
	}
	if change(&outside, nil).Allowed {
		t.Error("removing an assignment the grantor could not grant")
	}
	if change(&outside, &inside).Allowed {
		t.Error("narrowing an assignment the grantor could not grant")
	}
	if change(&inside, &outside).Allowed {
		t.Error("widening beyond the grantor")
	}
	if _, err := siteAdmin.CanChange(context.Background(), f.source, viewer, nil, nil); !errors.Is(err, ErrRoleAssignmentInvalid) {
		t.Errorf("no scopes: %v", err)
	}
}

func builtIn(t *testing.T, f grantFixture, key RoleKey) Role {
	t.Helper()
	role, ok := BuiltInRole(key)
	if !ok {
		t.Fatalf("no built-in role %q", key)
	}
	role.Id, role.OrganisationId = primitive.NewObjectID(), f.organisationId
	return role
}

func TestOrganisationManagerGrants(t *testing.T) {
	f := newGrantFixture()
	manager := builtIn(t, f, RoleKeyOrganisationManager)
	access := f.access(t, map[*Role]RoleAssignmentScope{
		{Name: manager.Name, AssignableAt: manager.AssignableAt, Permissions: manager.Permissions}: allProjects(),
	})
	for _, test := range []struct {
		key        RoleKey
		scope      RoleAssignmentScope
		permission Permission // owner-only permission that denies it, if any
	}{
		{key: RoleKeyViewer, scope: f.allResources(f.projectOne)},
		{key: RoleKeyEditor, scope: f.selected(f.projectTwo, []primitive.ObjectID{f.siteB}, nil)},
		{key: RoleKeyOrganisationManager, scope: allProjects()},
		{key: RoleKeyProjectAdmin, scope: f.allResources(f.projectOne), permission: PermissionWorkflowsCreate},
		{key: RoleKeyBillingAdmin, scope: allProjects(), permission: PermissionBillingRead},
	} {
		t.Run(string(test.key), func(t *testing.T) {
			decision := mustGrant(t, access, f.source, builtIn(t, f, test.key), test.scope)
			if test.permission == "" && !decision.Allowed {
				t.Errorf("denied: %+v", decision)
			}
			if test.permission != "" && (decision.Allowed || decision.Reason != GrantDenialOwnerOnly || decision.Permission != test.permission) {
				t.Errorf("decision %+v, want owner-only %s", decision, test.permission)
			}
		})
	}
}

// Delegated administrators hand out at most what the legacy admin tier holds,
// so migrated admins can grant exactly what they could before.
func TestLegacyRolesHoldNoOwnerOnlyPermission(t *testing.T) {
	for _, key := range []RoleKey{RoleKeyLegacyGuest, RoleKeyLegacyEditor, RoleKeyLegacyAdmin, RoleKeyLegacyMemberOrganisation, RoleKeyLegacyAdminOrganisation,
		RoleKeyOrganisationManager, RoleKeyEditor, RoleKeyViewer} {
		role, _ := BuiltInRole(key)
		if permission, ok := role.OwnerOnlyPermission(); ok {
			t.Errorf("%s holds owner-only %q", key, permission)
		}
	}
	for _, key := range []RoleKey{RoleKeyOwner, RoleKeyProjectAdmin, RoleKeyBillingAdmin} {
		role, _ := BuiltInRole(key)
		if _, ok := role.OwnerOnlyPermission(); !ok {
			t.Errorf("%s holds no owner-only permission", key)
		}
	}
}

func TestCanManageRole(t *testing.T) {
	f := newGrantFixture()
	roleAdmin := f.access(t, map[*Role]RoleAssignmentScope{
		{Name: "Role admin", AssignableAt: LevelOrganisation, Permissions: []Permission{PermissionRolesCreate, PermissionRolesUpdate, PermissionRolesDelete}}: allProjects(),
	})
	updateOnly := f.access(t, map[*Role]RoleAssignmentScope{
		{Name: "Role editor", AssignableAt: LevelOrganisation, Permissions: []Permission{PermissionRolesUpdate}}: allProjects(),
	})
	teamId, otherUser := primitive.NewObjectID(), primitive.NewObjectID()
	myTeams := []TeamMember{{OrganisationId: f.organisationId, TeamId: teamId, UserId: f.userId}}
	operator := f.projectRole(PermissionDevicesRead)
	wider := operator
	wider.Permissions = []Permission{PermissionDevicesRead, PermissionDevicesConfigure}
	logs := operator
	logs.Permissions = []Permission{PermissionDevicesRead, PermissionDevicesReadLogs}
	assigned := func(principal Principal) []RoleAssignment {
		return []RoleAssignment{{OrganisationId: f.organisationId, Principal: principal, RoleId: operator.Id, Scope: f.allResources(f.projectOne), IsActive: true}}
	}
	toOther := assigned(Principal{Type: PrincipalTypeUser, Id: otherUser})

	tests := []struct {
		name          string
		access        Access
		current, next *Role
		assignments   []RoleAssignment
		reason        GrantDenial
	}{
		{name: "create a role with permissions not held", access: roleAdmin, next: &wider},
		{name: "widen a role assigned to others", access: roleAdmin, current: &operator, next: &wider, assignments: toOther},
		{name: "delete a role assigned to others", access: roleAdmin, current: &operator, assignments: toOther},
		{name: "create without roles.create", access: updateOnly, next: &wider, reason: GrantDenialNoAccessPermission},
		{name: "delete without roles.delete", access: updateOnly, current: &operator, reason: GrantDenialNoAccessPermission},
		{name: "add an owner-only permission", access: roleAdmin, current: &operator, next: &logs, reason: GrantDenialOwnerOnly},
		{name: "remove an owner-only permission", access: roleAdmin, current: &logs, next: &operator, reason: GrantDenialOwnerOnly},
		{name: "create with an owner-only permission", access: roleAdmin, next: &logs, reason: GrantDenialOwnerOnly},
		{name: "a role the actor holds", access: roleAdmin, current: &operator, next: &wider,
			assignments: assigned(Principal{Type: PrincipalTypeUser, Id: f.userId}), reason: GrantDenialSelf},
		{name: "a role the actor holds through a team", access: roleAdmin, current: &operator,
			assignments: assigned(Principal{Type: PrincipalTypeTeam, Id: teamId}), reason: GrantDenialSelf},
		{name: "owner", access: ownerAccess(t, f), current: &operator, next: &logs, assignments: assigned(Principal{Type: PrincipalTypeUser, Id: f.userId})},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			decision, err := test.access.CanManageRole(f.userId, myTeams, test.current, test.next, test.assignments)
			if err != nil {
				t.Fatal(err)
			}
			if test.reason == "" && !decision.Allowed {
				t.Errorf("denied: %+v", decision)
			}
			if test.reason != "" && (decision.Allowed || decision.Reason != test.reason) {
				t.Errorf("decision %+v, want %s", decision, test.reason)
			}
		})
	}

	viewer := builtIn(t, f, RoleKeyViewer)
	invalid := operator
	invalid.Name = ""
	for name, roles := range map[string][2]*Role{
		"system role":  {&viewer, &wider},
		"invalid next": {nil, &invalid},
		"no roles":     {nil, nil},
	} {
		if _, err := ownerAccess(t, f).CanManageRole(f.userId, nil, roles[0], roles[1], nil); !errors.Is(err, ErrRoleInvalid) {
			t.Errorf("%s: %v", name, err)
		}
	}
}

func TestCanChangeTeamMembers(t *testing.T) {
	f := newGrantFixture()
	teamAdmin := f.access(t, map[*Role]RoleAssignmentScope{
		{Name: "Team admin", AssignableAt: LevelOrganisation, Permissions: []Permission{PermissionTeamsUpdate}}: allProjects(),
	})
	reader := f.access(t, map[*Role]RoleAssignmentScope{
		{Name: "Team reader", AssignableAt: LevelOrganisation, Permissions: []Permission{PermissionTeamsRead}}: allProjects(),
	})
	team := Team{Id: primitive.NewObjectID(), OrganisationId: f.organisationId, Name: "Night shift", Source: TeamSourceManual}
	other := primitive.NewObjectID()
	projectAdmin := builtIn(t, f, RoleKeyProjectAdmin)
	editor := builtIn(t, f, RoleKeyEditor)

	tests := []struct {
		name    string
		access  Access
		roles   []Role
		members []primitive.ObjectID
		reason  GrantDenial
	}{
		{name: "team with access the actor does not hold", access: teamAdmin, roles: []Role{editor}, members: []primitive.ObjectID{other}},
		{name: "without teams.update", access: reader, roles: []Role{editor}, members: []primitive.ObjectID{other}, reason: GrantDenialNoAccessPermission},
		{name: "team assigned an owner-only role", access: teamAdmin, roles: []Role{editor, projectAdmin}, members: []primitive.ObjectID{other}, reason: GrantDenialOwnerOnly},
		{name: "the actor themselves", access: teamAdmin, roles: []Role{editor}, members: []primitive.ObjectID{other, f.userId}, reason: GrantDenialSelf},
		{name: "owner", access: ownerAccess(t, f), roles: []Role{projectAdmin}, members: []primitive.ObjectID{f.userId}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			decision, err := test.access.CanChangeTeamMembers(f.userId, team, test.roles, test.members)
			if err != nil {
				t.Fatal(err)
			}
			if test.reason == "" && !decision.Allowed {
				t.Errorf("denied: %+v", decision)
			}
			if test.reason != "" && (decision.Allowed || decision.Reason != test.reason) {
				t.Errorf("decision %+v, want %s", decision, test.reason)
			}
		})
	}

	synced := team
	synced.Source, synced.IdentityProvider = TeamSourceIdentityProvider, &TeamIdentityProvider{Provider: "entra", ExternalGroupId: "g-1"}
	if _, err := ownerAccess(t, f).CanChangeTeamMembers(f.userId, synced, nil, []primitive.ObjectID{other}); !errors.Is(err, ErrTeamInvalid) {
		t.Errorf("identity provider team: %v", err)
	}
}
