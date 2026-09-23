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
		PermissionWorkflowRunsUpdate,
	}
	if got := WorkflowRunPermissions(); !slices.Equal(got, want) {
		t.Fatalf("WorkflowRunPermissions() = %v, want %v", got, want)
	}
}

func TestPermissionCatalogAccessorsReturnCopies(t *testing.T) {
	all := AllPermissions()
	markers := MarkerPermissions()
	media := MediaPermissions()
	cases := CasePermissions()
	workflows := WorkflowPermissions()
	workflowRuns := WorkflowRunPermissions()

	all[0] = "changed"
	markers[0] = "changed"
	media[0] = "changed"
	cases[0] = "changed"
	workflows[0] = "changed"
	workflowRuns[0] = "changed"

	if AllPermissions()[0] != PermissionMarkersRead {
		t.Fatal("AllPermissions returned mutable catalog storage")
	}
	if MarkerPermissions()[0] != PermissionMarkersRead {
		t.Fatal("MarkerPermissions returned mutable catalog storage")
	}
	if MediaPermissions()[0] != PermissionMediaRead {
		t.Fatal("MediaPermissions returned mutable catalog storage")
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
}

func TestAccessTokenScopesContainCanonicalPermissions(t *testing.T) {
	want := AllPermissions()
	if got := AccessTokenScopes(); !slices.Equal(got, want) {
		t.Fatalf("AccessTokenScopes() = %v, want %v", got, want)
	}
}
