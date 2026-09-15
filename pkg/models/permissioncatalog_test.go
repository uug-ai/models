package models

import "testing"

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
	if _, err := ParsePermission("media.unknown"); err == nil {
		t.Fatal("ParsePermission must reject an unknown permission")
	}
}

func TestPermissionCatalogAccessorsReturnCopies(t *testing.T) {
	all := AllPermissions()
	media := MediaPermissions()
	cases := CasePermissions()

	all[0] = "changed"
	media[0] = "changed"
	cases[0] = "changed"

	if AllPermissions()[0] != PermissionMediaRead {
		t.Fatal("AllPermissions returned mutable catalog storage")
	}
	if MediaPermissions()[0] != PermissionMediaRead {
		t.Fatal("MediaPermissions returned mutable catalog storage")
	}
	if CasePermissions()[0] != PermissionCasesRead {
		t.Fatal("CasePermissions returned mutable catalog storage")
	}
}
