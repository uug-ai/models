package api

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestOnboardingStatusTranslate(t *testing.T) {
	statuses := []OnboardingStatus{
		OrganisationRequired, OnboardingFound, OnboardingGetFailed, OnboardingBindingFailed, OnboardingMissingInfo,
		OnboardingOrganisationCreateSuccess, OnboardingOrganisationCreateFailed, OnboardingOrganisationCreateDisabled,
		OnboardingOrganisationExists,
	}
	seen := map[OnboardingStatus]bool{}
	for _, status := range statuses {
		if seen[status] {
			t.Errorf("%s listed twice", status)
		}
		seen[status] = true
		if got := status.Translate("en"); got == status.String() || got == "" {
			t.Errorf("%s has no English translation", status)
		}
		if status.Translate("xx") != status.Translate("en") {
			t.Errorf("%s: unknown language does not fall back to English", status)
		}
	}
	if OrganisationRequired.String() != "organisation_required" {
		t.Errorf("OrganisationRequired = %q; clients match on organisation_required", OrganisationRequired)
	}
	if got := OnboardingStatus("onboarding_unknown").Translate("en"); got != "onboarding_unknown" {
		t.Errorf("unknown status = %q, want the raw value", got)
	}
}

func TestGetOnboardingResponseEncodesEmptyLists(t *testing.T) {
	encoded, err := json.Marshal(GetOnboardingResponse{OrganisationRequired: true,
		Invitations: []PendingInvitation{}, Organisations: []OnboardingOrganisation{}})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"organisationRequired":true`, `"canCreateOrganisation":false`, `"invitations":[]`, `"organisations":[]`} {
		if !strings.Contains(string(encoded), want) {
			t.Errorf("%s: missing %s", encoded, want)
		}
	}
}
