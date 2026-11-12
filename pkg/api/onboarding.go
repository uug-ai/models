package api

import (
	"time"

	"github.com/uug-ai/models/pkg/models"
)

// OnboardingStatus represents specific status codes for the onboarding step:
// what an account without an organisation sees after signing in.
type OnboardingStatus string

const (
	// OrganisationRequired: the account has no organisation to work in. Every
	// organisation-scoped route answers it with HTTP 409 until the account
	// accepts an invitation, selects an organisation it belongs to or creates
	// one.
	OrganisationRequired OnboardingStatus = "organisation_required"

	OnboardingFound                     OnboardingStatus = "onboarding_found"
	OnboardingGetFailed                 OnboardingStatus = "onboarding_get_failed"
	OnboardingBindingFailed             OnboardingStatus = "onboarding_binding_failed"
	OnboardingMissingInfo               OnboardingStatus = "onboarding_missing_info"
	OnboardingOrganisationCreateSuccess OnboardingStatus = "onboarding_organisation_create_success"
	OnboardingOrganisationCreateFailed  OnboardingStatus = "onboarding_organisation_create_failed"
	// OnboardingOrganisationCreateDisabled: this deployment does not let
	// accounts create an organisation in the onboarding step.
	OnboardingOrganisationCreateDisabled OnboardingStatus = "onboarding_organisation_create_disabled"
	// OnboardingOrganisationExists: the account already has a home
	// organisation; further organisations are created with POST /organisations.
	OnboardingOrganisationExists OnboardingStatus = "onboarding_organisation_exists"
)

// String returns the string representation of the onboarding status.
func (s OnboardingStatus) String() string {
	return string(s)
}

// Translate returns the translated string representation of the onboarding
// status in the specified language, falling back to English and then to the
// raw value.
func (s OnboardingStatus) Translate(lang string) string {
	translations := map[string]map[OnboardingStatus]string{
		"en": {
			OrganisationRequired:                 "Join or create an organisation to continue",
			OnboardingFound:                      "Onboarding status retrieved successfully",
			OnboardingGetFailed:                  "Failed to retrieve the onboarding status",
			OnboardingBindingFailed:              "Onboarding binding failed",
			OnboardingMissingInfo:                "The organisation needs a name",
			OnboardingOrganisationCreateSuccess:  "Organisation created successfully",
			OnboardingOrganisationCreateFailed:   "Failed to create organisation",
			OnboardingOrganisationCreateDisabled: "Creating an organisation is not available; ask for an invitation",
			OnboardingOrganisationExists:         "This account already has its own organisation",
		},
	}
	if langTranslations, exists := translations[lang]; exists {
		if translation, exists := langTranslations[s]; exists {
			return translation
		}
	}
	if enTranslations, exists := translations["en"]; exists {
		if translation, exists := enTranslations[s]; exists {
			return translation
		}
	}
	return s.String()
}

// ---------- Request / Response DTOs (onboarding domain) ----------

// PendingInvitation is an open invitation to the signed-in account's own
// email address, as shown in the onboarding step and the organisation
// switcher. It carries no token: the account's address is what lets it accept
// (POST /invitations/{id}/accept).
type PendingInvitation struct {
	Id               string    `json:"id"`
	OrganisationId   string    `json:"organisationId"`
	OrganisationName string    `json:"organisationName"`
	InvitedByName    string    `json:"invitedByName"`
	ExpiresAt        time.Time `json:"expiresAt"`
}

// OnboardingOrganisation names an organisation the account belongs to and
// can select (PATCH /organisations/current).
type OnboardingOrganisation struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

// GetOnboarding tells a signed-in account whether it needs an organisation
// and what it can do about it. It answers for every account, also one that
// already works in an organisation.
// @Router /onboarding [get]
type GetOnboardingRequest struct{}
type GetOnboardingResponse struct {
	// OrganisationRequired is true while organisation-scoped routes refuse
	// the account with OrganisationRequired.
	OrganisationRequired bool `json:"organisationRequired"`
	// CanCreateOrganisation is true when the account may create its home
	// organisation here (POST /onboarding/organisation).
	CanCreateOrganisation bool `json:"canCreateOrganisation"`
	// Invitations are the open invitations to the account's address.
	Invitations []PendingInvitation `json:"invitations"`
	// Organisations are the organisations the account is a member of.
	Organisations []OnboardingOrganisation `json:"organisations"`
	Username      string                   `json:"username"`
	Email         string                   `json:"email"`
}
type GetOnboardingSuccessResponse struct {
	SuccessResponse
	Data GetOnboardingResponse `json:"data"`
}
type GetOnboardingErrorResponse struct {
	ErrorResponse
}

// CreateOnboardingOrganisation creates the account's home organisation: the
// account becomes its owner, gets the default project and works in it from
// the returned token on.
// @Router /onboarding/organisation [post]
type CreateOnboardingOrganisationRequest struct {
	Name string `json:"name"`
}
type CreateOnboardingOrganisationResponse struct {
	Organisation models.Organisation `json:"organisation"`
	// Token is a session token scoped to the new organisation; Expire is its
	// expiry (RFC 3339).
	Token  string `json:"token"`
	Expire string `json:"expire"`
}
type CreateOnboardingOrganisationSuccessResponse struct {
	SuccessResponse
	Data CreateOnboardingOrganisationResponse `json:"data"`
}
type CreateOnboardingOrganisationErrorResponse struct {
	ErrorResponse
}
