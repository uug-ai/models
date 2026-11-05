package api

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/uug-ai/models/pkg/models"
)

var roleAssignmentStatuses = []RoleAssignmentStatus{
	RoleAssignmentBindingFailed, RoleAssignmentValidationFailed, RoleAssignmentFound, RoleAssignmentNotFound,
	RoleAssignmentGetAllSuccess, RoleAssignmentGetAllFailed, RoleAssignmentCreateSuccess, RoleAssignmentCreateFailed,
	RoleAssignmentUpdateSuccess, RoleAssignmentUpdateFailed, RoleAssignmentDeleteSuccess, RoleAssignmentDeleteFailed,
	RoleAssignmentExists, RoleAssignmentRoleNotFound, RoleAssignmentPrincipalNotFound, RoleAssignmentResourceNotFound,
	RoleAssignmentSelf, RoleAssignmentExceedsAccess, RoleAssignmentLastOwner, RoleAssignmentLegacyMode,
}

func TestRoleAssignmentStatusTranslate(t *testing.T) {
	seen := map[RoleAssignmentStatus]bool{}
	for _, status := range roleAssignmentStatuses {
		if seen[status] {
			t.Errorf("%s listed twice", status)
		}
		seen[status] = true
		if !strings.HasPrefix(status.String(), "role_assignment_") {
			t.Errorf("%s: status codes start with role_assignment_", status)
		}
		if got := status.Translate("en"); got == status.String() || got == "" {
			t.Errorf("%s has no English translation", status)
		}
		if status.Translate("xx") != status.Translate("en") {
			t.Errorf("%s: unknown language does not fall back to English", status)
		}
	}
	if got := RoleAssignmentStatus("role_assignment_unknown").Translate("en"); got != "role_assignment_unknown" {
		t.Errorf("unknown status = %q, want the raw value", got)
	}
}

func TestCreateRoleAssignmentRequestDecodes(t *testing.T) {
	body := `{
		"principal": {"type": "user", "id": "6720f0a1b2c3d4e5f6a7b8c9"},
		"roleId": "6720f0a1b2c3d4e5f6a7b8ca",
		"scope": {"level": "project", "projectId": "6720f0a1b2c3d4e5f6a7b8cb", "resources": "selected", "siteIds": ["6720f0a1b2c3d4e5f6a7b8cc"]},
		"expiresAt": "2027-01-01T00:00:00Z"
	}`
	var request CreateRoleAssignmentRequest
	if err := json.Unmarshal([]byte(body), &request); err != nil {
		t.Fatal(err)
	}
	if request.Principal.Type != models.PrincipalTypeUser || request.Principal.Id.Hex() != "6720f0a1b2c3d4e5f6a7b8c9" {
		t.Errorf("principal %+v", request.Principal)
	}
	if request.RoleId.IsZero() || request.Scope.Resources != models.SelectionSelected || len(request.Scope.SiteIds) != 1 {
		t.Errorf("decoded %+v", request)
	}
	if request.ExpiresAt == nil || request.ExpiresAt.Year() != 2027 || request.TimeWindow != nil {
		t.Errorf("expiry %v, window %v", request.ExpiresAt, request.TimeWindow)
	}
}

func TestUpdateRoleAssignmentRequestDecodesRemovals(t *testing.T) {
	var request UpdateRoleAssignmentRequest
	if err := json.Unmarshal([]byte(`{"roleAssignment": {"removeTimeWindow": true, "isActive": false}}`), &request); err != nil {
		t.Fatal(err)
	}
	patch := request.RoleAssignment
	if !patch.RemoveTimeWindow || patch.IsActive == nil || *patch.IsActive || patch.Scope != nil || patch.ExpiresAt != nil {
		t.Errorf("decoded %+v", patch)
	}
}

func TestRoleAssignmentPrincipalCarriesIdentityOnly(t *testing.T) {
	body, err := json.Marshal(RoleAssignmentPrincipal{Type: models.PrincipalTypeUser, Id: "u1", Name: "kim", Email: "kim@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatal(err)
	}
	for field := range decoded {
		switch field {
		case "type", "id", "name", "email", "memberCount":
		default:
			t.Errorf("principal exposes %q", field)
		}
	}
}
