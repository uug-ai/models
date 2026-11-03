package models

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestApplyDeviceCeilings(t *testing.T) {
	everything := deviceScoped(AllPermissions())
	without := func(removed ...Permission) []Permission {
		var kept []Permission
		for _, permission := range everything {
			if !slices.Contains(removed, permission) {
				kept = append(kept, permission)
			}
		}
		return kept
	}
	tests := []struct {
		name     string
		ceilings *DeviceFeaturePermissions
		want     []Permission
	}{
		{"no ceilings", nil, everything},
		{"unset ceilings", &DeviceFeaturePermissions{}, everything},
		{"no PTZ", &DeviceFeaturePermissions{PTZ: 1}, without(PermissionPTZZoom, PermissionPTZPanTilt, PermissionPTZControl, PermissionPTZPresets)},
		{"PTZ zoom", &DeviceFeaturePermissions{PTZ: 2}, without(PermissionPTZPanTilt, PermissionPTZControl, PermissionPTZPresets)},
		{"PTZ pan/tilt without zoom", &DeviceFeaturePermissions{PTZ: 3}, without(PermissionPTZZoom, PermissionPTZControl, PermissionPTZPresets)},
		{"PTZ without presets", &DeviceFeaturePermissions{PTZ: 4}, without(PermissionPTZPresets)},
		{"PTZ presets", &DeviceFeaturePermissions{PTZ: 5}, everything},
		{"no live view", &DeviceFeaturePermissions{Liveview: 1}, without(PermissionLivePreview, PermissionLiveStream)},
		{"live preview", &DeviceFeaturePermissions{Liveview: 2}, without(PermissionLiveStream)},
		{"no remote config", &DeviceFeaturePermissions{RemoteConfig: 1}, without(PermissionDevicesConfigure, PermissionDevicesManageRecording)},
		{"remote config", &DeviceFeaturePermissions{RemoteConfig: 2}, everything},
		{"no IO", &DeviceFeaturePermissions{IO: 1}, without(PermissionIOTrigger)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := ApplyDeviceCeilings(everything, test.ceilings); !slices.Equal(got, test.want) {
				t.Errorf("got %v, want %v", got, test.want)
			}
		})
	}
}

// accessWith resolves a member holding each role on its scope in one
// organisation.
func accessWith(t *testing.T, organisationId, userId primitive.ObjectID, grants map[*Role]RoleAssignmentScope) Access {
	t.Helper()
	var roles []Role
	var assignments []RoleAssignment
	for role, scope := range grants {
		role.Id, role.OrganisationId, role.Kind, role.Grants, role.IsActive = primitive.NewObjectID(), organisationId, RoleKindCustom, SelectionSelected, true
		roles = append(roles, *role)
		assignments = append(assignments, RoleAssignment{
			Id: primitive.NewObjectID(), OrganisationId: organisationId, Principal: Principal{Type: PrincipalTypeUser, Id: userId},
			RoleId: role.Id, Scope: scope, Source: AssignmentSourceManual, IsActive: true,
		})
	}
	membership := OrganisationUser{UserId: userId, OrganisationId: organisationId, Status: MembershipStatusActive}
	return NewAccess(membership, roles, assignments, time.Now())
}

func TestDeviceAccessSplitsPermissionsPerSite(t *testing.T) {
	organisationId, userId := primitive.NewObjectID(), primitive.NewObjectID()
	project := DefaultProjectId(organisationId)
	siteA, siteB := primitive.NewObjectID(), primitive.NewObjectID()
	source := &fakeDeviceSource{sites: map[string]Site{
		siteA.Hex(): {Devices: []string{"cam-a", "cam-shared"}},
		siteB.Hex(): {Devices: []string{"cam-b", "cam-shared"}},
	}}
	editor := &Role{Name: "Site editor", AssignableAt: LevelProject,
		Permissions: []Permission{PermissionDevicesRead, PermissionDevicesUpdate, PermissionLiveStream, PermissionPTZPresets, PermissionCasesUpdate}}
	viewer := &Role{Name: "Site viewer", AssignableAt: LevelProject,
		Permissions: []Permission{PermissionDevicesRead, PermissionLivePreview}}
	access := accessWith(t, organisationId, userId, map[*Role]RoleAssignmentScope{
		editor: {Level: LevelProject, ProjectId: project, Resources: SelectionSelected, SiteIds: []primitive.ObjectID{siteA}},
		viewer: {Level: LevelProject, ProjectId: project, Resources: SelectionSelected, SiteIds: []primitive.ObjectID{siteB}},
	})

	devices, err := access.CompileDeviceAccess(context.Background(), source, project)
	if err != nil {
		t.Fatal(err)
	}
	editing := ExpandPermissions([]Permission{PermissionDevicesRead, PermissionDevicesUpdate, PermissionLiveStream, PermissionPTZPresets})
	viewing := ExpandPermissions([]Permission{PermissionDevicesRead, PermissionLivePreview})
	for _, check := range []struct {
		device string
		want   []Permission
	}{
		{"cam-a", editing},
		{"cam-b", viewing},
		{"cam-shared", editing}, // Both grants cover it; the editor's includes the viewer's
		{"cam-elsewhere", nil},
	} {
		if got := devices.Permissions(check.device, nil); !slices.Equal(got, check.want) {
			t.Errorf("%s: got %v, want %v", check.device, got, check.want)
		}
	}
	if !devices.Allows(PermissionDevicesUpdate, "cam-a", nil) || devices.Allows(PermissionDevicesUpdate, "cam-b", nil) {
		t.Error("devices.update must hold on site A's camera only")
	}
	// Ceilings narrow the grant on that device only.
	if got := devices.Permissions("cam-a", &DeviceFeaturePermissions{PTZ: 2, Liveview: 2}); slices.Contains(got, PermissionPTZPresets) ||
		slices.Contains(got, PermissionLiveStream) || !slices.Contains(got, PermissionPTZZoom) || !slices.Contains(got, PermissionLivePreview) {
		t.Errorf("ceilings not applied: %v", got)
	}
	// Permissions without a device dimension are not device permissions.
	if slices.Contains(devices.Permissions("cam-a", nil), PermissionCasesUpdate) {
		t.Error("cases.update listed on a device")
	}
}

func TestDeviceAccessScopes(t *testing.T) {
	organisationId, userId := primitive.NewObjectID(), primitive.NewObjectID()
	project, otherProject := DefaultProjectId(organisationId), primitive.NewObjectID()
	reader := &Role{Name: "Reader", AssignableAt: LevelProject, Permissions: []Permission{PermissionDevicesRead}}
	deleter := &Role{Name: "Deleter", AssignableAt: LevelProject, Permissions: []Permission{PermissionDevicesDelete}}
	manager := &Role{Name: "Manager", AssignableAt: LevelOrganisation, Permissions: []Permission{PermissionDevicesManageSystemTags}}
	access := accessWith(t, organisationId, userId, map[*Role]RoleAssignmentScope{
		reader:  {Level: LevelProject, ProjectId: project, Resources: SelectionAll},
		deleter: {Level: LevelProject, ProjectId: otherProject, Resources: SelectionAll},
		manager: {Level: LevelOrganisation, Projects: SelectionAll},
	})
	devices, err := access.CompileDeviceAccess(context.Background(), &fakeDeviceSource{}, project)
	if err != nil {
		t.Fatal(err)
	}
	if got := devices.Permissions("any-camera", nil); !slices.Equal(got, []Permission{PermissionDevicesRead}) {
		t.Errorf("got %v; want only the read grant of this project (all resources), no other project's grant, no organisation permission", got)
	}
}

func TestDeviceAccessOwnerAndLegacyGuest(t *testing.T) {
	organisationId, userId := primitive.NewObjectID(), primitive.NewObjectID()
	project := DefaultProjectId(organisationId)
	siteId := primitive.NewObjectID()
	source := &fakeDeviceSource{sites: map[string]Site{siteId.Hex(): {Devices: []string{"site-cam"}}}}

	mapped, err := MapLegacyAccess(LegacyUserAccess{OrganisationId: organisationId, UserId: userId, Role: "owner"})
	if err != nil {
		t.Fatal(err)
	}
	owner, err := NewAccess(mapped.Membership, mapped.Roles, mapped.Assignments, time.Now()).CompileDeviceAccess(context.Background(), source, project)
	if err != nil {
		t.Fatal(err)
	}
	if got := owner.Permissions("any-camera", nil); !slices.Equal(got, deviceScoped(AllPermissions())) {
		t.Errorf("owner got %v", got)
	}
	if owner.Allows(PermissionPTZPresets, "any-camera", &DeviceFeaturePermissions{PTZ: 1}) {
		t.Error("ceilings must apply to owners too")
	}

	mapped, err = MapLegacyAccess(LegacyUserAccess{OrganisationId: organisationId, UserId: userId, Role: "guest",
		SiteIds: []primitive.ObjectID{siteId}, DeviceKeys: []string{"own-cam"}})
	if err != nil {
		t.Fatal(err)
	}
	guest, err := NewAccess(mapped.Membership, mapped.Roles, mapped.Assignments, time.Now()).CompileDeviceAccess(context.Background(), source, project)
	if err != nil {
		t.Fatal(err)
	}
	for _, device := range []string{"site-cam", "own-cam"} {
		if got := guest.Permissions(device, nil); !slices.Contains(got, PermissionDevicesRead) || !slices.Contains(got, PermissionLiveStream) ||
			slices.Contains(got, PermissionDevicesUpdate) || slices.Contains(got, PermissionPTZZoom) {
			t.Errorf("guest on %s got %v", device, got)
		}
	}
	if got := guest.Permissions("unselected-cam", nil); len(got) > 0 {
		t.Errorf("guest on an unselected camera got %v", got)
	}
}

func TestCompileDeviceAccessFailsOnReadError(t *testing.T) {
	organisationId, userId := primitive.NewObjectID(), primitive.NewObjectID()
	project := DefaultProjectId(organisationId)
	mapped, err := MapLegacyAccess(LegacyUserAccess{OrganisationId: organisationId, UserId: userId, Role: "guest", SiteIds: []primitive.ObjectID{primitive.NewObjectID()}})
	if err != nil {
		t.Fatal(err)
	}
	want := errors.New("down")
	_, err = NewAccess(mapped.Membership, mapped.Roles, mapped.Assignments, time.Now()).CompileDeviceAccess(context.Background(), &fakeDeviceSource{err: want}, project)
	if !errors.Is(err, want) {
		t.Errorf("err = %v, want %v", err, want)
	}
}
