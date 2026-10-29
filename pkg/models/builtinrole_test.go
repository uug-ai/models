package models

import (
	"slices"
	"testing"
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
