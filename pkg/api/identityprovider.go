package api

import "github.com/uug-ai/models/pkg/models"

// IdentityProviderStatus represents specific status codes for identity
// provider connection operations (RBAC spec 4.16).
type IdentityProviderStatus string

const (
	IdentityProviderBindingFailed    IdentityProviderStatus = "identity_provider_binding_failed"
	IdentityProviderValidationFailed IdentityProviderStatus = "identity_provider_validation_failed"
	IdentityProviderFound            IdentityProviderStatus = "identity_provider_found"
	IdentityProviderNotFound         IdentityProviderStatus = "identity_provider_not_found"
	IdentityProviderGetAllSuccess    IdentityProviderStatus = "identity_provider_get_all_success"
	IdentityProviderGetAllFailed     IdentityProviderStatus = "identity_provider_get_all_failed"
	IdentityProviderCreateSuccess    IdentityProviderStatus = "identity_provider_create_success"
	IdentityProviderCreateFailed     IdentityProviderStatus = "identity_provider_create_failed"
	IdentityProviderUpdateSuccess    IdentityProviderStatus = "identity_provider_update_success"
	IdentityProviderUpdateFailed     IdentityProviderStatus = "identity_provider_update_failed"
	IdentityProviderDeleteSuccess    IdentityProviderStatus = "identity_provider_delete_success"
	IdentityProviderDeleteFailed     IdentityProviderStatus = "identity_provider_delete_failed"
	// IdentityProviderNameExists: the organisation has a connection with
	// this name already (names compare ignoring case).
	IdentityProviderNameExists IdentityProviderStatus = "identity_provider_name_exists"
	// IdentityProviderSlugRequired: enabling a connection needs the
	// organisation's slug, its sign-in name.
	IdentityProviderSlugRequired IdentityProviderStatus = "identity_provider_slug_required"
	// IdentityProviderInUse: synced teams name the connection; unlink or
	// delete them first.
	IdentityProviderInUse IdentityProviderStatus = "identity_provider_in_use"
	// IdentityProviderScimTokenCreated: a SCIM token was generated (shown
	// once).
	IdentityProviderScimTokenCreated IdentityProviderStatus = "identity_provider_scim_token_created"
	// IdentityProviderScimTokenDeleted: the SCIM token was removed.
	IdentityProviderScimTokenDeleted IdentityProviderStatus = "identity_provider_scim_token_deleted"
)

// String returns the string representation of the identity provider status.
func (s IdentityProviderStatus) String() string {
	return string(s)
}

// Translate returns the translated string representation of the identity
// provider status in the specified language, falling back to English and
// then to the raw value.
func (s IdentityProviderStatus) Translate(lang string) string {
	translations := map[string]map[IdentityProviderStatus]string{
		"en": {
			IdentityProviderBindingFailed:    "Identity provider binding failed",
			IdentityProviderValidationFailed: "Identity provider validation failed",
			IdentityProviderFound:            "Identity provider found",
			IdentityProviderNotFound:         "Identity provider not found",
			IdentityProviderGetAllSuccess:    "Identity providers retrieved successfully",
			IdentityProviderGetAllFailed:     "Failed to retrieve identity providers",
			IdentityProviderCreateSuccess:    "Identity provider connected successfully",
			IdentityProviderCreateFailed:     "Failed to connect the identity provider",
			IdentityProviderUpdateSuccess:    "Identity provider updated successfully",
			IdentityProviderUpdateFailed:     "Failed to update the identity provider",
			IdentityProviderDeleteSuccess:    "Identity provider removed successfully",
			IdentityProviderDeleteFailed:     "Failed to remove the identity provider",
			IdentityProviderNameExists:       "An identity provider with this name already exists",
			IdentityProviderSlugRequired:     "Set the organisation's slug before enabling sign-in",
			IdentityProviderInUse:            "Teams are synced from this identity provider",
			IdentityProviderScimTokenCreated: "SCIM token generated",
			IdentityProviderScimTokenDeleted: "SCIM token removed",
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

// ---------- Request / Response DTOs (identity provider connections) ----------
//
// An organisation's identity provider connections (RBAC spec 4.16), in the
// caller's current organisation. Only an owner manages them
// (organisation.manageIdentityProvider); session tokens only. The client
// secret is write-only: requests may set it, responses never return it. A
// connection is enabled only once the organisation has a slug
// (IdentityProviderSlugRequired).

// IdentityProviderConnectionDetails is a connection as returned: the stored
// connection without its secret, whether a secret is set, and the redirect
// URI to register at the provider (one per deployment).
type IdentityProviderConnectionDetails struct {
	Connection      models.IdentityProviderConnection `json:"connection"`
	HasClientSecret bool                              `json:"hasClientSecret"`
	RedirectUri     string                            `json:"redirectUri,omitempty"`
	// ScimBaseUri is the SCIM endpoint to give the provider.
	ScimBaseUri string `json:"scimBaseUri,omitempty"`
}

// CreateIdentityProviderScimToken generates the connection's SCIM token,
// replacing one it had. The token is in this response only.
// @Router /identity-providers/{id}/scim-token [post]
type CreateIdentityProviderScimTokenRequest struct{}
type CreateIdentityProviderScimTokenResponse struct {
	Token       string `json:"token"`
	ScimBaseUri string `json:"scimBaseUri"`
}
type CreateIdentityProviderScimTokenSuccessResponse struct {
	SuccessResponse
	Data CreateIdentityProviderScimTokenResponse `json:"data"`
}
type CreateIdentityProviderScimTokenErrorResponse struct {
	ErrorResponse
}

// DeleteIdentityProviderScimToken removes the connection's SCIM token: the
// provider can no longer provision.
// @Router /identity-providers/{id}/scim-token [delete]
type DeleteIdentityProviderScimTokenRequest struct{}
type DeleteIdentityProviderScimTokenResponse struct{}
type DeleteIdentityProviderScimTokenSuccessResponse struct {
	SuccessResponse
	Data DeleteIdentityProviderScimTokenResponse `json:"data"`
}
type DeleteIdentityProviderScimTokenErrorResponse struct {
	ErrorResponse
}

// GetIdentityProviderConnections lists the organisation's connections.
// @Router /identity-providers [get]
type GetIdentityProviderConnectionsRequest struct{}
type GetIdentityProviderConnectionsResponse struct {
	Connections []IdentityProviderConnectionDetails `json:"connections"`
}
type GetIdentityProviderConnectionsSuccessResponse struct {
	SuccessResponse
	Data GetIdentityProviderConnectionsResponse `json:"data"`
}
type GetIdentityProviderConnectionsErrorResponse struct {
	ErrorResponse
}

// GetIdentityProviderConnection returns one connection.
// @Router /identity-providers/{id} [get]
type GetIdentityProviderConnectionRequest struct{}
type GetIdentityProviderConnectionResponse struct {
	Connection IdentityProviderConnectionDetails `json:"connection"`
}
type GetIdentityProviderConnectionSuccessResponse struct {
	SuccessResponse
	Data GetIdentityProviderConnectionResponse `json:"data"`
}
type GetIdentityProviderConnectionErrorResponse struct {
	ErrorResponse
}

// CreateIdentityProviderConnection connects an identity provider. The server
// sets the id, organisation, protocol (oidc) and audit.
// @Router /identity-providers [post]
type CreateIdentityProviderConnectionRequest struct {
	Name         string                        `json:"name"`
	Issuer       string                        `json:"issuer"`
	ClientId     string                        `json:"clientId"`
	ClientSecret string                        `json:"clientSecret"`
	Scopes       []string                      `json:"scopes,omitempty"`
	Claims       models.IdentityProviderClaims `json:"claims,omitempty"`
	JoinOnSignIn bool                          `json:"joinOnSignIn"`
	Enabled      bool                          `json:"enabled"`
}
type CreateIdentityProviderConnectionResponse struct {
	Connection IdentityProviderConnectionDetails `json:"connection"`
}
type CreateIdentityProviderConnectionSuccessResponse struct {
	SuccessResponse
	Data CreateIdentityProviderConnectionResponse `json:"data"`
}
type CreateIdentityProviderConnectionErrorResponse struct {
	ErrorResponse
}

// IdentityProviderConnectionPatch changes a connection: only the fields that
// are set. The name never changes (synced teams name the connection by it).
// ClientSecret replaces the stored secret.
type IdentityProviderConnectionPatch struct {
	Issuer       *string                        `json:"issuer,omitempty"`
	ClientId     *string                        `json:"clientId,omitempty"`
	ClientSecret *string                        `json:"clientSecret,omitempty"`
	Scopes       *[]string                      `json:"scopes,omitempty"`
	Claims       *models.IdentityProviderClaims `json:"claims,omitempty"`
	JoinOnSignIn *bool                          `json:"joinOnSignIn,omitempty"`
	Enabled      *bool                          `json:"enabled,omitempty"`
}

// IsEmpty reports whether the patch changes nothing.
func (p IdentityProviderConnectionPatch) IsEmpty() bool {
	return p.Issuer == nil && p.ClientId == nil && p.ClientSecret == nil && p.Scopes == nil &&
		p.Claims == nil && p.JoinOnSignIn == nil && p.Enabled == nil
}

// UpdateIdentityProviderConnection changes a connection.
// @Router /identity-providers/{id} [patch]
type UpdateIdentityProviderConnectionRequest struct {
	Connection IdentityProviderConnectionPatch `json:"connection"`
}
type UpdateIdentityProviderConnectionResponse struct {
	Connection IdentityProviderConnectionDetails `json:"connection"`
}
type UpdateIdentityProviderConnectionSuccessResponse struct {
	SuccessResponse
	Data UpdateIdentityProviderConnectionResponse `json:"data"`
}
type UpdateIdentityProviderConnectionErrorResponse struct {
	ErrorResponse
}

// DeleteIdentityProviderConnection removes a connection and its account
// links. Refused (IdentityProviderInUse) while synced teams name it.
// @Router /identity-providers/{id} [delete]
type DeleteIdentityProviderConnectionRequest struct{}
type DeleteIdentityProviderConnectionResponse struct{}
type DeleteIdentityProviderConnectionSuccessResponse struct {
	SuccessResponse
	Data DeleteIdentityProviderConnectionResponse `json:"data"`
}
type DeleteIdentityProviderConnectionErrorResponse struct {
	ErrorResponse
}
