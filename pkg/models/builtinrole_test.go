package models

import (
	"slices"
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestBuiltInRolesValidate(t *testing.T) {
	seen := map[RoleKey]bool{}
	for _, role := range BuiltInRoles() {
		if err := role.Validate(); err != nil {
			t.Errorf("%s: %v", role.Key, err)
		}
		if role.Kind == RoleKindCustom || !role.IsActive || !role.Id.IsZero() || !role.OrganisationId.IsZero() {
			t.Errorf("%s: want an active system or legacy role without ids, got %+v", role.Key, role)
		}
		if seen[role.Key] {
			t.Errorf("duplicate key %s", role.Key)
		}
		seen[role.Key] = true
	}
	if owner, ok := BuiltInRole(RoleKeyOwner); !ok || !owner.IsOwner() || !owner.GrantsAll() {
		t.Errorf("owner role = %+v, %v", owner, ok)
	}
}

func TestBuiltInRolePermissionsAreCanonical(t *testing.T) {
	for _, role := range BuiltInRoles() {
		if !slices.Equal(role.Permissions, catalogOrder(role.Permissions)) {
			t.Errorf("%s: permissions not in catalog order", role.Key)
		}
		// A stored permission implied by another stored one is redundant.
		for _, permission := range role.Permissions {
			others := slices.DeleteFunc(slices.Clone(role.Permissions), func(other Permission) bool { return other == permission })
			if slices.Contains(ExpandPermissions(others), permission) {
				t.Errorf("%s: %s is implied by another permission", role.Key, permission)
			}
		}
	}
}

func TestLegacyTiersNest(t *testing.T) {
	chains := [][]RoleKey{
		{RoleKeyLegacyGuest, RoleKeyLegacyEditor, RoleKeyLegacyAdmin},
		{RoleKeyLegacyMemberOrganisation, RoleKeyLegacyAdminOrganisation},
	}
	for _, chain := range chains {
		for index := 1; index < len(chain); index++ {
			lower, _ := BuiltInRole(chain[index-1])
			higher, _ := BuiltInRole(chain[index])
			for _, permission := range lower.Permissions {
				if !slices.Contains(higher.Permissions, permission) {
					t.Errorf("%s holds %s but %s does not", lower.Key, permission, higher.Key)
				}
			}
		}
	}
}

func TestBuiltInRoleReturnsCopies(t *testing.T) {
	role, ok := BuiltInRole(RoleKeyLegacyGuest)
	if !ok {
		t.Fatal("legacy guest role missing")
	}
	role.Permissions[0] = PermissionDevicesDelete
	again, _ := BuiltInRole(RoleKeyLegacyGuest)
	if again.Permissions[0] == PermissionDevicesDelete {
		t.Error("BuiltInRole shares its permission slice")
	}
	if _, ok := BuiltInRole(RoleKeyViewer); ok {
		t.Error("viewer is not defined yet")
	}
}

func TestBuiltInRolesReturnsCopies(t *testing.T) {
	roles := BuiltInRoles()
	for i := range roles {
		if len(roles[i].Permissions) > 0 {
			roles[i].Permissions[0] = PermissionDevicesDelete
		}
		if roles[i].Origin != nil {
			roles[i].Origin.BaseRole = BaseRoleOwner
		}
		roles[i].Id = primitive.NewObjectID()
	}
	for i, role := range BuiltInRoles() {
		if len(role.Permissions) > 0 && role.Permissions[0] == PermissionDevicesDelete {
			t.Errorf("%s: BuiltInRoles shares its permission slice", role.Key)
		}
		if role.Origin != nil && role.Origin.BaseRole == BaseRoleOwner {
			t.Errorf("%s: BuiltInRoles shares its origin", role.Key)
		}
		if !role.Id.IsZero() || role.Key != roles[i].Key {
			t.Errorf("%s: BuiltInRoles returned a changed role", role.Key)
		}
	}
}

func BenchmarkBuiltInRole(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		BuiltInRole(RoleKeyLegacyAdminOrganisation)
	}
}

func BenchmarkMapLegacyAccessCustomRole(b *testing.B) {
	in := LegacyUserAccess{
		OrganisationId: primitive.NewObjectID(),
		UserId:         primitive.NewObjectID(),
		Role:           "Operator",
		Timezone:       LegacyDefaultTimezone,
		CustomRole: &LegacyRole{
			Id:         primitive.NewObjectID(),
			Name:       "Operator",
			ParentRole: string(BaseRoleEditor),
			Pages:      []string{legacyPageLiveView, legacyPageRecordings, legacyPageTasks},
		},
	}
	b.ReportAllocs()
	for b.Loop() {
		if _, err := MapLegacyAccess(in); err != nil {
			b.Fatal(err)
		}
	}
}
