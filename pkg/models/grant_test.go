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
		{Name: "Delegate", AssignableAt: LevelOrganisation, Permissions: []Permission{PermissionProjectsManageAccess}}:                selectedProjects(f.projectOne),
	})
	tests := []struct {
		name       string
		access     Access
		role       Role
		scope      RoleAssignmentScope
		reason     GrantDenial
		permission Permission
		project    primitive.ObjectID
	}{
		{name: "held on all projects", access: manager, role: f.organisationRole(PermissionMembersRead), scope: allProjects()},
		{name: "held on the selected project", access: manager, role: f.organisationRole(PermissionProjectsManageAccess), scope: selectedProjects(f.projectOne)},
		{name: "not held on another selected project", access: manager, role: f.organisationRole(PermissionProjectsManageAccess),
			scope: selectedProjects(f.projectOne, f.projectTwo), reason: GrantDenialPermissionNotHeld, permission: PermissionProjectsManageAccess, project: f.projectTwo},
		{name: "not held on all projects", access: manager, role: f.organisationRole(PermissionProjectsManageAccess),
			scope: allProjects(), reason: GrantDenialPermissionNotHeld, permission: PermissionProjectsManageAccess},
		{name: "not held anywhere", access: manager, role: f.organisationRole(PermissionMembersRead, PermissionRolesRead),
			scope: allProjects(), reason: GrantDenialPermissionNotHeld, permission: PermissionRolesRead},
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
			if decision.Allowed || decision.Reason != test.reason || decision.Permission != test.permission || decision.Project != test.project {
				t.Errorf("decision %+v, want %s %s %s", decision, test.reason, test.permission, test.project.Hex())
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
	if decision := mustGrant(t, delegate, f.source, f.projectRole(PermissionDevicesRead, PermissionDevicesDelete), f.allResources(f.projectOne)); decision.Allowed ||
		decision.Reason != GrantDenialPermissionNotHeld || decision.Permission != PermissionDevicesDelete {
		t.Errorf("permission not held: %+v", decision)
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

	if decision := mustGrant(t, siteAdmin, f.source, f.projectRole(PermissionDevicesRead, PermissionDevicesUpdate), f.selected(f.projectOne, sites(f.siteA), nil)); decision.Allowed ||
		decision.Reason != GrantDenialPermissionNotHeld || decision.Permission != PermissionDevicesUpdate {
		t.Errorf("permission not held: %+v", decision)
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
