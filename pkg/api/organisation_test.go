package api

import "testing"

func TestOrganisationStatus_String(t *testing.T) {
	if got := OrganisationFound.String(); got != "organisation_found" {
		t.Errorf("String() = %q, want %q", got, "organisation_found")
	}
}

func TestOrganisationStatus_Translate(t *testing.T) {
	tests := []struct {
		name   string
		status OrganisationStatus
		lang   string
		want   string
	}{
		{"english binding failed", OrganisationBindingFailed, "en", "Organisation binding failed"},
		{"english missing info", OrganisationMissingInfo, "en", "Organisation is missing required information"},
		{"english name exists", OrganisationNameExists, "en", "Organisation name already exists"},
		{"english found", OrganisationFound, "en", "Organisation found"},
		{"english not found", OrganisationNotFound, "en", "Organisation not found"},
		{"english get all success", OrganisationGetAllSuccess, "en", "Organisations retrieved successfully"},
		{"english get all failed", OrganisationGetAllFailed, "en", "Failed to retrieve organisations"},
		{"english create success", OrganisationCreateSuccess, "en", "Organisation created successfully"},
		{"english create failed", OrganisationCreateFailed, "en", "Failed to create organisation"},
		{"english update success", OrganisationUpdateSuccess, "en", "Organisation updated successfully"},
		{"english update failed", OrganisationUpdateFailed, "en", "Failed to update organisation"},
		{"english delete success", OrganisationDeleteSuccess, "en", "Organisation deleted successfully"},
		{"english delete failed", OrganisationDeleteFailed, "en", "Failed to delete organisation"},
		{"english validation fail", OrganisationValidationFail, "en", "Organisation validation failed"},
		{"unknown lang falls back to english", OrganisationFound, "xx", "Organisation found"},
		{"unknown status returns raw value", OrganisationStatus("organisation_unknown"), "en", "organisation_unknown"},
		{"unknown lang and status returns raw value", OrganisationStatus("organisation_unknown"), "xx", "organisation_unknown"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.status.Translate(tc.lang); got != tc.want {
				t.Errorf("Translate(%q) = %q, want %q", tc.lang, got, tc.want)
			}
		})
	}
}
