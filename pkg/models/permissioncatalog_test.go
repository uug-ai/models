package models

import (
	"errors"
	"slices"
	"testing"
)

func TestPermissionCatalogContainsUniqueValidPermissions(t *testing.T) {
	seen := make(map[Permission]struct{})
	for _, permission := range AllPermissions() {
		if permission == "" {
			t.Fatal("permission catalog contains an empty permission")
		}
		if _, exists := seen[permission]; exists {
			t.Fatalf("permission catalog contains duplicate %q", permission)
		}
		seen[permission] = struct{}{}

		parsed, err := ParsePermission(string(permission))
		if err != nil {
			t.Fatalf("ParsePermission(%q) returned error: %v", permission, err)
		}
		if parsed != permission {
			t.Fatalf("ParsePermission(%q) = %q", permission, parsed)
		}
	}
}

func TestParsePermissionRejectsUnknownPermission(t *testing.T) {
	if _, err := ParsePermission("media.unknown"); !errors.Is(err, ErrPermissionUnknown) {
		t.Fatalf("ParsePermission error = %v, want ErrPermissionUnknown", err)
	}
}

func TestValidatePermissions(t *testing.T) {
	tests := []struct {
		name        string
		permissions []Permission
		wantErr     error
	}{
		{name: "empty"},
		{name: "valid", permissions: []Permission{PermissionMediaRead, PermissionCasesExport}},
		{name: "unknown", permissions: []Permission{"media.unknown"}, wantErr: ErrPermissionUnknown},
		{name: "duplicate", permissions: []Permission{PermissionMediaRead, PermissionMediaRead}, wantErr: ErrPermissionDuplicate},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidatePermissions(test.permissions)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("ValidatePermissions() error = %v, want %v", err, test.wantErr)
			}
		})
	}
}

func TestRoleValidatePermissions(t *testing.T) {
	role := Role{Permissions: []Permission{PermissionCasesRead, "cases.unknown"}}
	if err := role.ValidatePermissions(); !errors.Is(err, ErrPermissionUnknown) {
		t.Fatalf("Role.ValidatePermissions() error = %v, want ErrPermissionUnknown", err)
	}
}

func TestCasePermissionsStableOrder(t *testing.T) {
	want := []Permission{
		PermissionCasesRead,
		PermissionCasesCreate,
		PermissionCasesUpdate,
		PermissionCasesShare,
		PermissionCasesExport,
		PermissionCasesRunWorkflow,
		PermissionCasesDelete,
	}
	if got := CasePermissions(); !slices.Equal(got, want) {
		t.Fatalf("CasePermissions() = %v, want %v", got, want)
	}
}

func TestMarkerPermissionsStableOrder(t *testing.T) {
	want := []Permission{
		PermissionMarkersRead,
		PermissionMarkersWrite,
	}
	if got := MarkerPermissions(); !slices.Equal(got, want) {
		t.Fatalf("MarkerPermissions() = %v, want %v", got, want)
	}
}

func TestDayPermissionsStableOrder(t *testing.T) {
	want := []Permission{PermissionDaysRead}
	if got := DayPermissions(); !slices.Equal(got, want) {
		t.Fatalf("DayPermissions() = %v, want %v", got, want)
	}
}

func TestWorkflowPermissionsStableOrder(t *testing.T) {
	want := []Permission{
		PermissionWorkflowsRead,
		PermissionWorkflowsCreate,
		PermissionWorkflowsUpdate,
		PermissionWorkflowsDelete,
	}
	if got := WorkflowPermissions(); !slices.Equal(got, want) {
		t.Fatalf("WorkflowPermissions() = %v, want %v", got, want)
	}
}

func TestWorkflowRunPermissionsStableOrder(t *testing.T) {
	want := []Permission{
		PermissionWorkflowRunsCreate,
		PermissionWorkflowRunsRead,
		PermissionWorkflowRunsUpdate,
	}
	if got := WorkflowRunPermissions(); !slices.Equal(got, want) {
		t.Fatalf("WorkflowRunPermissions() = %v, want %v", got, want)
	}
}

func TestChannelPermissionsStableOrder(t *testing.T) {
	want := []Permission{
		PermissionChannelsRead,
		PermissionChannelsCreate,
		PermissionChannelsUpdate,
		PermissionChannelsDelete,
	}
	if got := ChannelPermissions(); !slices.Equal(got, want) {
		t.Fatalf("ChannelPermissions() = %v, want %v", got, want)
	}
}

func TestPermissionCatalogAccessorsReturnCopies(t *testing.T) {
	all := AllPermissions()
	markers := MarkerPermissions()
	media := MediaPermissions()
	days := DayPermissions()
	cases := CasePermissions()
	workflows := WorkflowPermissions()
	workflowRuns := WorkflowRunPermissions()
	channels := ChannelPermissions()

	all[0] = "changed"
	markers[0] = "changed"
	media[0] = "changed"
	days[0] = "changed"
	cases[0] = "changed"
	workflows[0] = "changed"
	workflowRuns[0] = "changed"
	channels[0] = "changed"

	if AllPermissions()[0] != PermissionMarkersRead {
		t.Fatal("AllPermissions returned mutable catalog storage")
	}
	if MarkerPermissions()[0] != PermissionMarkersRead {
		t.Fatal("MarkerPermissions returned mutable catalog storage")
	}
	if MediaPermissions()[0] != PermissionMediaRead {
		t.Fatal("MediaPermissions returned mutable catalog storage")
	}
	if DayPermissions()[0] != PermissionDaysRead {
		t.Fatal("DayPermissions returned mutable catalog storage")
	}
	if CasePermissions()[0] != PermissionCasesRead {
		t.Fatal("CasePermissions returned mutable catalog storage")
	}
	if WorkflowPermissions()[0] != PermissionWorkflowsRead {
		t.Fatal("WorkflowPermissions returned mutable catalog storage")
	}
	if WorkflowRunPermissions()[0] != PermissionWorkflowRunsCreate {
		t.Fatal("WorkflowRunPermissions returned mutable catalog storage")
	}
	if ChannelPermissions()[0] != PermissionChannelsRead {
		t.Fatal("ChannelPermissions returned mutable catalog storage")
	}
}

func TestAccessTokenScopesAreProjectCatalogPermissions(t *testing.T) {
	scopes := AccessTokenScopes()
	if len(scopes) == 0 {
		t.Fatal("AccessTokenScopes() is empty")
	}
	for _, scope := range scopes {
		definition, ok := LookupPermission(scope)
		if !ok {
			t.Fatalf("AccessTokenScopes() contains non-catalog permission %q", scope)
		}
		if definition.Scope != PermissionScopeProject {
			t.Fatalf("AccessTokenScopes() contains %s permission %q", definition.Scope, scope)
		}
	}
	if err := ValidateAccessTokenScopes(scopes); err != nil {
		t.Fatalf("ValidateAccessTokenScopes(AccessTokenScopes()) error = %v", err)
	}
}

func TestAccessTokenScopesReturnCopy(t *testing.T) {
	scopes := AccessTokenScopes()
	want := scopes[0]
	scopes[0] = "changed"
	if AccessTokenScopes()[0] != want {
		t.Fatal("AccessTokenScopes returned mutable storage")
	}
}

func TestValidateAccessTokenScopes(t *testing.T) {
	tests := []struct {
		name    string
		scopes  []AccessTokenScope
		wantErr error
	}{
		{name: "empty"},
		{name: "valid", scopes: []AccessTokenScope{PermissionMediaRead, PermissionMarkersWrite}},
		{name: "unknown", scopes: []AccessTokenScope{"media.unknown"}, wantErr: ErrPermissionUnknown},
		{name: "application", scopes: []AccessTokenScope{PermissionApplication}, wantErr: ErrPermissionUnknown},
		{name: "duplicate", scopes: []AccessTokenScope{PermissionMediaRead, PermissionMediaRead}, wantErr: ErrPermissionDuplicate},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateAccessTokenScopes(test.scopes)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("ValidateAccessTokenScopes() error = %v, want %v", err, test.wantErr)
			}
		})
	}
}

func TestValidateAccessTokenScopesRejectsCatalogPermissionOutsideAllowList(t *testing.T) {
	original := accessTokenScopeSet
	t.Cleanup(func() { accessTokenScopeSet = original })
	accessTokenScopeSet = map[AccessTokenScope]struct{}{PermissionMediaRead: {}}

	err := ValidateAccessTokenScopes([]AccessTokenScope{PermissionMediaRead, PermissionMediaDelete})
	if !errors.Is(err, ErrAccessTokenScopeNotGrantable) {
		t.Fatalf("ValidateAccessTokenScopes() error = %v, want ErrAccessTokenScopeNotGrantable", err)
	}
}

func TestPermissionCatalogDefinitionsAreValid(t *testing.T) {
	if err := validatePermissionGroups(permissionGroups); err != nil {
		t.Fatalf("validatePermissionGroups(permissionGroups) error = %v", err)
	}
}

func TestPermissionDefinitionsMatchCatalogOrder(t *testing.T) {
	definitions := PermissionDefinitions()
	all := AllPermissions()
	if len(definitions) != len(all) {
		t.Fatalf("len(PermissionDefinitions()) = %d, want %d", len(definitions), len(all))
	}
	for index, definition := range definitions {
		if definition.Permission != all[index] {
			t.Fatalf("PermissionDefinitions()[%d] = %q, want %q", index, definition.Permission, all[index])
		}
	}
}

func TestExistingPermissionsAreProjectScoped(t *testing.T) {
	for _, permission := range appendPermissions(
		MarkerPermissions(),
		MediaPermissions(),
		CasePermissions(),
		WorkflowPermissions(),
		WorkflowRunPermissions(),
		DayPermissions(),
		ChannelPermissions(),
	) {
		definition, ok := LookupPermission(permission)
		if !ok {
			t.Fatalf("LookupPermission(%q) not found", permission)
		}
		if definition.Scope != PermissionScopeProject {
			t.Fatalf("LookupPermission(%q).Scope = %q, want %q", permission, definition.Scope, PermissionScopeProject)
		}
	}
}

func TestLookupPermission(t *testing.T) {
	if _, ok := LookupPermission("media.unknown"); ok {
		t.Fatal("LookupPermission(unknown) found a definition")
	}
	if _, ok := LookupPermission(PermissionApplication); ok {
		t.Fatal("LookupPermission(application) found a definition")
	}
}

func TestValidatePermissionGroups(t *testing.T) {
	const (
		a Permission = "test.a"
		b Permission = "test.b"
		c Permission = "test.c"
	)
	project := PermissionScopeProject
	organisation := PermissionScopeOrganisation

	tests := []struct {
		name    string
		groups  []permissionGroup
		wantErr error
	}{
		{
			name: "chain",
			groups: []permissionGroup{{scope: project, permissions: []Permission{a, b, c}, implies: map[Permission][]Permission{
				a: {b},
				b: {c},
			}}},
		},
		{
			name:    "empty permission",
			groups:  []permissionGroup{{scope: project, permissions: []Permission{""}}},
			wantErr: ErrPermissionInvalid,
		},
		{
			name:    "unknown scope",
			groups:  []permissionGroup{{scope: "tenant", permissions: []Permission{a}}},
			wantErr: ErrPermissionInvalid,
		},
		{
			name: "duplicate across groups",
			groups: []permissionGroup{
				{scope: project, permissions: []Permission{a}},
				{scope: organisation, permissions: []Permission{a}},
			},
			wantErr: ErrPermissionDuplicate,
		},
		{
			name:    "implies unknown",
			groups:  []permissionGroup{{scope: project, permissions: []Permission{a}, implies: map[Permission][]Permission{a: {b}}}},
			wantErr: ErrPermissionInvalid,
		},
		{
			name: "implies across scopes",
			groups: []permissionGroup{
				{scope: organisation, permissions: []Permission{a}, implies: map[Permission][]Permission{a: {b}}},
				{scope: project, permissions: []Permission{b}},
			},
			wantErr: ErrPermissionInvalid,
		},
		{
			name: "implication source in another group",
			groups: []permissionGroup{
				{scope: project, permissions: []Permission{a}, implies: map[Permission][]Permission{b: {a}}},
				{scope: project, permissions: []Permission{b}},
			},
			wantErr: ErrPermissionInvalid,
		},
		{
			name:    "self implication",
			groups:  []permissionGroup{{scope: project, permissions: []Permission{a}, implies: map[Permission][]Permission{a: {a}}}},
			wantErr: ErrPermissionInvalid,
		},
		{
			name: "cycle",
			groups: []permissionGroup{{scope: project, permissions: []Permission{a, b, c}, implies: map[Permission][]Permission{
				a: {b},
				b: {c},
				c: {a},
			}}},
			wantErr: ErrPermissionInvalid,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validatePermissionGroups(test.groups)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("validatePermissionGroups() error = %v, want %v", err, test.wantErr)
			}
		})
	}
}

func TestExpandPermissions(t *testing.T) {
	const (
		presets Permission = "test.presets"
		control Permission = "test.control"
		zoom    Permission = "test.zoom"
		other   Permission = "test.other"
	)
	definitions := definePermissions([]permissionGroup{{
		scope:       PermissionScopeProject,
		permissions: []Permission{zoom, control, presets, other},
		implies: map[Permission][]Permission{
			presets: {control},
			control: {zoom},
		},
	}})
	index := indexPermissions(definitions)

	tests := []struct {
		name        string
		permissions []Permission
		want        []Permission
	}{
		{name: "empty"},
		{name: "transitive in catalog order", permissions: []Permission{presets}, want: []Permission{zoom, control, presets}},
		{name: "duplicates collapse", permissions: []Permission{control, zoom, control}, want: []Permission{zoom, control}},
		{name: "unknown dropped", permissions: []Permission{"test.unknown", other}, want: []Permission{other}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := expandPermissions(definitions, index, test.permissions); !slices.Equal(got, test.want) {
				t.Fatalf("expandPermissions() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestExpandPermissionsUsesCatalog(t *testing.T) {
	got := ExpandPermissions([]Permission{PermissionMediaDelete, PermissionApplication, PermissionMarkersRead})
	want := []Permission{PermissionMarkersRead, PermissionMediaDelete}
	if !slices.Equal(got, want) {
		t.Fatalf("ExpandPermissions() = %v, want %v", got, want)
	}
}

func TestPermissionDefinitionsReturnCopies(t *testing.T) {
	definitions := PermissionDefinitions()
	definitions[0].Permission = "changed"
	definitions[0].Implies = append(definitions[0].Implies, "changed")
	if got := PermissionDefinitions()[0]; got.Permission != PermissionMarkersRead || len(got.Implies) != 0 {
		t.Fatalf("PermissionDefinitions returned mutable catalog storage: %+v", got)
	}
}

func TestPrivilegedPermissionsAreNotGrantableAccessTokenScopes(t *testing.T) {
	scopes := AccessTokenScopes()
	for _, permission := range []Permission{PermissionApplication, PermissionMarkersAll} {
		if slices.Contains(scopes, permission) {
			t.Fatalf("AccessTokenScopes() contains privileged permission %q", permission)
		}
	}
}
