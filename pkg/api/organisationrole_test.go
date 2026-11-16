package api

import (
	"strings"
	"testing"
)

var organisationRoleStatuses = []OrganisationRoleStatus{
	OrganisationRoleBindingFailed, OrganisationRoleValidationFailed, OrganisationRoleFound, OrganisationRoleNotFound,
	OrganisationRoleGetAllSuccess, OrganisationRoleGetAllFailed, OrganisationRoleCreateSuccess, OrganisationRoleCreateFailed,
	OrganisationRoleUpdateSuccess, OrganisationRoleUpdateFailed, OrganisationRoleDeleteSuccess, OrganisationRoleDeleteFailed,
	OrganisationRoleNameExists, OrganisationRoleReadOnly, OrganisationRoleInUse, OrganisationRoleOwnerOnly,
	OrganisationRoleSelf, OrganisationRoleLegacyMode,
}

func TestOrganisationRoleStatusTranslate(t *testing.T) {
	seen := map[OrganisationRoleStatus]bool{}
	for _, status := range organisationRoleStatuses {
		if seen[status] {
			t.Errorf("%s listed twice", status)
		}
		seen[status] = true
		// Distinct from the legacy custom role statuses (role_*).
		if !strings.HasPrefix(status.String(), "organisation_role_") {
			t.Errorf("%s: status codes start with organisation_role_", status)
		}
		if got := status.Translate("en"); got == status.String() || got == "" {
			t.Errorf("%s has no English translation", status)
		}
		if status.Translate("xx") != status.Translate("en") {
			t.Errorf("%s: unknown language does not fall back to English", status)
		}
	}
	if got := OrganisationRoleStatus("organisation_role_unknown").Translate("en"); got != "organisation_role_unknown" {
		t.Errorf("unknown status = %q, want the raw value", got)
	}
}

func TestOrganisationRoleUsageInUse(t *testing.T) {
	if (OrganisationRoleUsage{}).InUse() {
		t.Error("an unused role is in use")
	}
	for _, usage := range []OrganisationRoleUsage{{UserAssignments: 1}, {TeamAssignments: 1}, {PendingInvitations: 1}} {
		if !usage.InUse() {
			t.Errorf("%+v is not in use", usage)
		}
	}
}
