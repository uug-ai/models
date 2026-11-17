package models

import (
	"context"
	"errors"
	"reflect"
	"slices"
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
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
	if _, ok := BuiltInRole(RoleKey("auditor")); ok {
		t.Error("an undefined key resolved")
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

func TestOrganisationBuiltInRoles(t *testing.T) {
	organisationId := primitive.NewObjectID()
	roles := OrganisationBuiltInRoles(organisationId)
	if len(roles) != len(BuiltInRoles()) {
		t.Fatalf("%d copies, want one per built-in role (%d)", len(roles), len(BuiltInRoles()))
	}
	for _, role := range roles {
		if role.Id != BuiltInRoleId(organisationId, role.Key) || role.OrganisationId != organisationId {
			t.Errorf("%s: id %s organisation %s", role.Key, role.Id.Hex(), role.OrganisationId.Hex())
		}
		if err := role.Validate(); err != nil {
			t.Errorf("%s: %v", role.Key, err)
		}
	}
	// Copies are the caller's own.
	roles[1].Permissions[0] = "changed"
	if again := OrganisationBuiltInRoles(organisationId); again[1].Permissions[0] == "changed" {
		t.Error("changing a copy changed the definition")
	}
}

func TestWithCurrentDefinition(t *testing.T) {
	organisationId := primitive.NewObjectID()
	current, _ := BuiltInRole(RoleKeyViewer)
	stale := current
	stale.Id = BuiltInRoleId(organisationId, RoleKeyViewer)
	stale.OrganisationId = organisationId
	stale.Name = "Viewer (old)"
	stale.Permissions = []Permission{PermissionDevicesRead}
	stale.IsActive = false
	stale.Audit = Audit{CreatedBy: "migration"}

	got := stale.WithCurrentDefinition()
	if got.Name != current.Name || !slices.Equal(got.Permissions, current.Permissions) {
		t.Errorf("definition not current: %q %v", got.Name, got.Permissions)
	}
	if got.Id != stale.Id || got.OrganisationId != organisationId || got.IsActive || got.Audit.CreatedBy != "migration" {
		t.Errorf("stored fields not kept: %+v", got)
	}

	custom := NewCustomRole(organisationId, "Viewer", "", LevelProject, []Permission{PermissionDevicesRead})
	if got := custom.WithCurrentDefinition(); got.Name != "Viewer" || !slices.Equal(got.Permissions, custom.Permissions) {
		t.Errorf("custom role changed: %+v", got)
	}
	unknown := stale
	unknown.Key = "retired-system-role"
	if got := unknown.WithCurrentDefinition(); got.Name != stale.Name {
		t.Errorf("unknown key changed: %+v", got)
	}
	wrongKind := stale
	wrongKind.Kind = RoleKindLegacy
	if got := wrongKind.WithCurrentDefinition(); got.Name != stale.Name {
		t.Errorf("legacy role with a system key took the system definition: %+v", got)
	}
}

func TestStoredCopyLeavesTheDefinitionOut(t *testing.T) {
	organisationId := primitive.NewObjectID()
	full := OrganisationBuiltInRoles(organisationId)
	copies := OrganisationBuiltInRoleCopies(organisationId)
	if len(copies) != len(full) {
		t.Fatalf("%d copies for %d built-in roles", len(copies), len(full))
	}
	for i, role := range full {
		role.IsActive = true
		role.Audit = Audit{CreatedBy: "owner"}
		stored := role.StoredCopy()
		if stored.Id != role.Id || stored.OrganisationId != organisationId || stored.Kind != role.Kind || stored.Key != role.Key ||
			stored.Name != role.Name || !stored.IsActive || stored.Audit.CreatedBy != "owner" {
			t.Errorf("%s: identity not kept: %+v", role.Key, stored)
		}
		if !stored.IsStoredCopy() || role.IsStoredCopy() {
			t.Errorf("%s: IsStoredCopy stored=%v full=%v", role.Key, stored.IsStoredCopy(), role.IsStoredCopy())
		}
		// Read back, it is the role the full copy is.
		if got, want := stored.WithCurrentDefinition(), role.WithCurrentDefinition(); !reflect.DeepEqual(got, want) {
			t.Errorf("%s: stored copy resolves to %+v, want %+v", role.Key, got, want)
		}
		if copies[i].Id != role.Id || !copies[i].IsStoredCopy() {
			t.Errorf("%s: OrganisationBuiltInRoleCopies gave %+v", role.Key, copies[i])
		}
		// The document holds no definition.
		raw, err := bson.Marshal(stored)
		if err != nil {
			t.Fatal(err)
		}
		var document bson.M
		if err := bson.Unmarshal(raw, &document); err != nil {
			t.Fatal(err)
		}
		for _, field := range []string{"permissions", "grants", "assignableAt", "description", "origin"} {
			if _, present := document[field]; present {
				t.Errorf("%s: stored copy has %s", role.Key, field)
			}
		}
		for _, field := range []string{"_id", "organisationId", "kind", "key", "name", "isActive"} {
			if _, present := document[field]; !present {
				t.Errorf("%s: stored copy lacks %s", role.Key, field)
			}
		}
	}

	// A custom role is the organisation's own: nothing is left out.
	custom := NewCustomRole(organisationId, "Night operator", "Watches live view", LevelProject, []Permission{PermissionDevicesRead})
	if got := custom.StoredCopy(); !reflect.DeepEqual(got, custom) || got.IsStoredCopy() {
		t.Errorf("custom role changed: %+v", got)
	}
	// A key this release does not know keeps what it stores, and a copy of it
	// without a definition grants nothing.
	unknown := full[0]
	unknown.Key = "retired-system-role"
	if got := unknown.StoredCopy(); !reflect.DeepEqual(got, unknown) {
		t.Errorf("unknown key trimmed: %+v", got)
	}
	orphan := Role{Id: primitive.NewObjectID(), OrganisationId: organisationId, Kind: RoleKindSystem, Key: "retired-system-role", Name: "Retired", IsActive: true}
	if got := orphan.WithCurrentDefinition(); len(got.Permissions) != 0 || got.GrantsAll() {
		t.Errorf("unknown stored copy grants something: %+v", got)
	}
}

func TestBuiltInRoleById(t *testing.T) {
	organisationId, other := primitive.NewObjectID(), primitive.NewObjectID()
	for _, definition := range BuiltInRoles() {
		id := BuiltInRoleId(organisationId, definition.Key)
		role, ok := BuiltInRoleById(organisationId, id)
		if !ok || role.Key != definition.Key || role.Id != id || role.OrganisationId != organisationId || !role.IsActive ||
			role.Name != definition.Name || !slices.Equal(role.Permissions, definition.Permissions) || role.Grants != definition.Grants {
			t.Errorf("%s: resolved %+v", definition.Key, role)
		}
		// The id belongs to one organisation.
		if _, ok := BuiltInRoleById(other, id); ok {
			t.Errorf("%s: the id of one organisation resolves in another", definition.Key)
		}
	}
	for name, id := range map[string]primitive.ObjectID{"custom": primitive.NewObjectID(), "zero": primitive.NilObjectID} {
		if _, ok := BuiltInRoleById(organisationId, id); ok {
			t.Errorf("%s id resolves to a built-in role", name)
		}
	}
	if _, ok := BuiltInRoleById(primitive.NilObjectID, BuiltInRoleId(primitive.NilObjectID, RoleKeyOwner)); ok {
		t.Error("resolved without an organisation")
	}
}

func TestIsBuiltInRoleName(t *testing.T) {
	for _, role := range BuiltInRoles() {
		for _, name := range []string{role.Name, strings.ToUpper(role.Name), "  " + strings.ToLower(role.Name) + " "} {
			if !IsBuiltInRoleName(name) {
				t.Errorf("%q is not recognised as the name of %s", name, role.Key)
			}
		}
	}
	for _, name := range []string{"", "Night operator", "Viewer 2", "Owners"} {
		if IsBuiltInRoleName(name) {
			t.Errorf("%q is taken for a built-in name", name)
		}
	}
}

func TestResolveRoles(t *testing.T) {
	organisationId := primitive.NewObjectID()
	viewer, editor := BuiltInRoleId(organisationId, RoleKeyViewer), BuiltInRoleId(organisationId, RoleKeyEditor)
	custom := NewCustomRole(organisationId, "Night operator", "", LevelProject, []Permission{PermissionDevicesRead})
	custom.Id = primitive.NewObjectID()
	var asked [][]primitive.ObjectID
	stored := func(_ context.Context, _ primitive.ObjectID, ids []primitive.ObjectID) ([]Role, error) {
		asked = append(asked, ids)
		return []Role{custom}, nil
	}

	// Only built-in ids: nothing is read.
	roles, err := ResolveRoles(context.Background(), organisationId, []primitive.ObjectID{viewer, editor}, stored)
	if err != nil || len(roles) != 2 || roles[0].Key != RoleKeyViewer || roles[1].Key != RoleKeyEditor || len(asked) != 0 {
		t.Fatalf("built-in only: %+v %v, read %v", roles, err, asked)
	}
	// Mixed: only the custom id is read.
	roles, err = ResolveRoles(context.Background(), organisationId, []primitive.ObjectID{viewer, custom.Id}, stored)
	if err != nil || len(roles) != 2 || roles[1].Name != "Night operator" || len(asked) != 1 || !slices.Equal(asked[0], []primitive.ObjectID{custom.Id}) {
		t.Fatalf("mixed: %+v %v, read %v", roles, err, asked)
	}
	// A read error is the caller's.
	failing := func(context.Context, primitive.ObjectID, []primitive.ObjectID) ([]Role, error) {
		return nil, errors.New("database unavailable")
	}
	if _, err := ResolveRoles(context.Background(), organisationId, []primitive.ObjectID{custom.Id}, failing); err == nil {
		t.Error("read error swallowed")
	}
	if roles, err := ResolveRoles(context.Background(), organisationId, nil, stored); err != nil || len(roles) != 0 {
		t.Errorf("no ids: %+v %v", roles, err)
	}
}
