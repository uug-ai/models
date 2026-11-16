package models

import (
	"slices"
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestNewCustomRole(t *testing.T) {
	organisationId := primitive.NewObjectID()
	role := NewCustomRole(organisationId, "Night operator", "Watches at night", LevelProject,
		[]Permission{PermissionLiveStream, PermissionDevicesRead})
	if err := role.Validate(); err != nil {
		t.Fatal(err)
	}
	if role.Kind != RoleKindCustom || role.Grants != SelectionSelected || !role.IsActive || role.OrganisationId != organisationId || !role.Id.IsZero() {
		t.Errorf("role = %+v", role)
	}
	if !slices.Equal(role.Permissions, catalogOrder(role.Permissions)) {
		t.Errorf("permissions not in catalog order: %v", role.Permissions)
	}
	if empty := NewCustomRole(organisationId, "Nothing", "", LevelOrganisation, nil); empty.Validate() != nil || empty.GrantsAll() {
		t.Errorf("a role without permissions must be valid and grant nothing: %+v", empty)
	}
}

func TestRolePatch(t *testing.T) {
	role := NewCustomRole(primitive.NewObjectID(), "Operator", "Old", LevelProject, []Permission{PermissionDevicesRead})
	if !(RolePatch{}).IsEmpty() {
		t.Error("empty patch is not empty")
	}

	name := "Night operator"
	permissions := []Permission{PermissionLiveStream, PermissionDevicesRead}
	patched := RolePatch{Name: &name, Permissions: &permissions}.Apply(role)
	if patched.Name != name || patched.Description != "Old" || patched.AssignableAt != LevelProject || patched.Kind != RoleKindCustom {
		t.Errorf("patched = %+v", patched)
	}
	if !slices.Equal(patched.Permissions, catalogOrder(permissions)) {
		t.Errorf("permissions = %v", patched.Permissions)
	}
	permissions[0] = "changed"
	if slices.Contains(patched.Permissions, Permission("changed")) {
		t.Error("patched role shares the patch's slice")
	}

	description := ""
	kept := RolePatch{Description: &description}.Apply(role)
	kept.Permissions[0] = "changed"
	if role.Permissions[0] == "changed" {
		t.Error("patched role shares the stored role's slice")
	}

	none := []Permission{}
	if cleared := (RolePatch{Permissions: &none}).Apply(role); len(cleared.Permissions) != 0 || cleared.Validate() != nil {
		t.Errorf("cleared = %+v", cleared)
	}
	duplicate := []Permission{PermissionDevicesRead, PermissionDevicesRead}
	if (RolePatch{Permissions: &duplicate}).Apply(role).Validate() == nil {
		t.Error("duplicate permissions validate")
	}
}
