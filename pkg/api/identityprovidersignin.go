package api

import "time"

// IdentityProviderSignInStatus represents status codes of signing in through
// an organisation's identity provider (RBAC spec 4.16).
type IdentityProviderSignInStatus string

const (
	IdentityProviderSignInFound         IdentityProviderSignInStatus = "identity_provider_sign_in_found"
	IdentityProviderSignInFailed        IdentityProviderSignInStatus = "identity_provider_sign_in_failed"
	IdentityProviderSignInBindingFailed IdentityProviderSignInStatus = "identity_provider_sign_in_binding_failed"
	// IdentityProviderSignInOrganisationNotFound: no organisation has the
	// slug.
	IdentityProviderSignInOrganisationNotFound IdentityProviderSignInStatus = "identity_provider_sign_in_organisation_not_found"
	// IdentityProviderSignInNotFound: the code is unknown, used or expired.
	IdentityProviderSignInNotFound IdentityProviderSignInStatus = "identity_provider_sign_in_not_found"
	// IdentityProviderSignInEmailSent: the email proof link was sent.
	IdentityProviderSignInEmailSent IdentityProviderSignInStatus = "identity_provider_sign_in_email_sent"
	// IdentityProviderSignInEmailUnavailable: this deployment sends no email;
	// prove the address by signing in to the account instead.
	IdentityProviderSignInEmailUnavailable IdentityProviderSignInStatus = "identity_provider_sign_in_email_unavailable"
	// IdentityProviderSignInEmailMismatch: the signed-in account's email is
	// not the address the identity provider sent.
	IdentityProviderSignInEmailMismatch IdentityProviderSignInStatus = "identity_provider_sign_in_email_mismatch"
	// IdentityProviderSignInAccountsAmbiguous: several accounts have the
	// address; sign in to the one you mean to link it.
	IdentityProviderSignInAccountsAmbiguous IdentityProviderSignInStatus = "identity_provider_sign_in_accounts_ambiguous"
	// IdentityProviderSignInJoinDisabled: no account has the address and the
	// connection creates none; ask the organisation for an invitation.
	IdentityProviderSignInJoinDisabled IdentityProviderSignInStatus = "identity_provider_sign_in_join_disabled"
	// IdentityProviderSignInMembershipClosed: the person's membership of the
	// organisation is suspended or revoked.
	IdentityProviderSignInMembershipClosed IdentityProviderSignInStatus = "identity_provider_sign_in_membership_closed"
	// IdentityProviderSignInLinked: the account is linked to the connection.
	IdentityProviderSignInLinked IdentityProviderSignInStatus = "identity_provider_sign_in_linked"
	// SessionSignInRequired: the session was started through an
	// organisation's identity provider and stays in that organisation; this
	// needs the account's own sign-in (password or passkey), or the other
	// organisation's.
	SessionSignInRequired IdentityProviderSignInStatus = "session_sign_in_required"
)

// String returns the string representation of the status.
func (s IdentityProviderSignInStatus) String() string {
	return string(s)
}

// Translate returns the translated string representation of the status in
// the specified language, falling back to English and then to the raw value.
func (s IdentityProviderSignInStatus) Translate(lang string) string {
	translations := map[string]map[IdentityProviderSignInStatus]string{
		"en": {
			IdentityProviderSignInFound:                "Sign-in found",
			IdentityProviderSignInFailed:               "Sign-in failed",
			IdentityProviderSignInBindingFailed:        "Sign-in request is not valid",
			IdentityProviderSignInOrganisationNotFound: "Organisation not found",
			IdentityProviderSignInNotFound:             "This sign-in has expired; start again",
			IdentityProviderSignInEmailSent:            "Check your email for the link",
			IdentityProviderSignInEmailUnavailable:     "Email is not available; sign in to your account instead",
			IdentityProviderSignInEmailMismatch:        "Your account's email is not the one your identity provider sent",
			IdentityProviderSignInAccountsAmbiguous:    "Several accounts use this email; sign in to the one you mean",
			IdentityProviderSignInJoinDisabled:         "No account uses this email; ask the organisation for an invitation",
			IdentityProviderSignInMembershipClosed:     "Your membership of this organisation is suspended or revoked",
			IdentityProviderSignInLinked:               "Your account is linked",
			SessionSignInRequired:                      "Sign in again to do this",
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

// ---------- Request / Response DTOs (identity provider sign-in) ----------
//
// Public routes, except linking, which needs the account's own session. A
// sign-in goes: GET /sso/organisations/{slug}/{connectionId} (browser, to the
// provider), the provider's callback (one per deployment), then the frontend
// with a single-use code: exchanged for a session (POST
// /sso/organisations/exchange), or, while the email still needs its proof,
// standing for the pending link (the proof routes below). A session started
// this way stays in the connection's organisation (SessionSignInRequired).

// OrganisationSignInConnection is an enabled connection people may sign in
// through.
type OrganisationSignInConnection struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

// GetOrganisationSignIn returns how people sign in to the organisation with
// the slug: its name and enabled connections.
// @Router /sign-in/organisations/{slug} [get]
type GetOrganisationSignInRequest struct{}
type GetOrganisationSignInResponse struct {
	OrganisationId string                         `json:"organisationId"`
	Name           string                         `json:"name"`
	Slug           string                         `json:"slug"`
	Connections    []OrganisationSignInConnection `json:"connections"`
}
type GetOrganisationSignInSuccessResponse struct {
	SuccessResponse
	Data GetOrganisationSignInResponse `json:"data"`
}
type GetOrganisationSignInErrorResponse struct {
	ErrorResponse
}

// ExchangeIdentityProviderSignIn trades a sign-in code for a session. The
// answer is the one a password sign-in gets.
// @Router /sso/organisations/exchange [post]
type ExchangeIdentityProviderSignInRequest struct {
	Code string `json:"code"`
}

// IdentityProviderEmailProof is a pending link: the provider signed the
// person in, and their email still needs its proof before Hub links or
// creates an account.
type IdentityProviderEmailProof struct {
	OrganisationName string `json:"organisationName"`
	ConnectionName   string `json:"connectionName"`
	// Email is the address the provider sent.
	Email string `json:"email"`
	// EmailAvailable says whether this deployment can send the proof link.
	EmailAvailable bool      `json:"emailAvailable"`
	ExpiresAt      time.Time `json:"expiresAt"`
}

// GetIdentityProviderEmailProof returns a pending link.
// @Router /sso/proofs/{code} [get]
type GetIdentityProviderEmailProofRequest struct{}
type GetIdentityProviderEmailProofResponse struct {
	Proof IdentityProviderEmailProof `json:"proof"`
}
type GetIdentityProviderEmailProofSuccessResponse struct {
	SuccessResponse
	Data GetIdentityProviderEmailProofResponse `json:"data"`
}
type GetIdentityProviderEmailProofErrorResponse struct {
	ErrorResponse
}

// SendIdentityProviderEmailProof emails a single-use link to the address.
// @Router /sso/proofs/{code}/email [post]
type SendIdentityProviderEmailProofRequest struct{}
type SendIdentityProviderEmailProofResponse struct{}
type SendIdentityProviderEmailProofSuccessResponse struct {
	SuccessResponse
	Data SendIdentityProviderEmailProofResponse `json:"data"`
}
type SendIdentityProviderEmailProofErrorResponse struct {
	ErrorResponse
}

// ConfirmIdentityProviderEmailProof redeems the emailed link: the account
// with the address is linked (or created, when the connection allows it),
// and a sign-in code is returned for the exchange.
// @Router /sso/proofs/confirm [post]
type ConfirmIdentityProviderEmailProofRequest struct {
	Token string `json:"token"`
}
type ConfirmIdentityProviderEmailProofResponse struct {
	Code string `json:"code"`
}
type ConfirmIdentityProviderEmailProofSuccessResponse struct {
	SuccessResponse
	Data ConfirmIdentityProviderEmailProofResponse `json:"data"`
}
type ConfirmIdentityProviderEmailProofErrorResponse struct {
	ErrorResponse
}

// LinkIdentityProviderAccount links the signed-in account (its own
// session, whose email must be the address) and returns a sign-in code for
// the exchange.
// @Router /sso/proofs/{code}/link [post]
type LinkIdentityProviderAccountRequest struct{}
type LinkIdentityProviderAccountResponse struct {
	Code string `json:"code"`
}
type LinkIdentityProviderAccountSuccessResponse struct {
	SuccessResponse
	Data LinkIdentityProviderAccountResponse `json:"data"`
}
type LinkIdentityProviderAccountErrorResponse struct {
	ErrorResponse
}
