package api

import "github.com/uug-ai/models/pkg/models"

// AccessTokenStatus represents specific status codes for access token operations
type AccessTokenStatus string

const (
	AccessTokenBindingFailed    AccessTokenStatus = "accesstoken_binding_failed"
	AccessTokenPermissionsFound AccessTokenStatus = "accesstoken_permissions_found"
	AccessTokenNameExists       AccessTokenStatus = "accesstoken_name_exists"
	AccessTokenMissingInfo      AccessTokenStatus = "accesstoken_missing_info"
	AccessTokenFound            AccessTokenStatus = "accesstoken_found"
	AccessTokenNotFound         AccessTokenStatus = "accesstoken_not_found"
	AccessTokenAddSuccess       AccessTokenStatus = "accesstoken_add_success"
	AccessTokenAddFailed        AccessTokenStatus = "accesstoken_add_failed"
	AccessTokenUpdateSuccess    AccessTokenStatus = "accesstoken_update_success"
	AccessTokenUpdateFailed     AccessTokenStatus = "accesstoken_update_failed"
	AccessTokenDeleteSuccess    AccessTokenStatus = "accesstoken_delete_success"
	AccessTokenDeleteFailed     AccessTokenStatus = "accesstoken_1delete_failed"
	AccessTokenRotateSuccess    AccessTokenStatus = "accesstoken_rotate_success"
	AccessTokenRotateFailed     AccessTokenStatus = "accesstoken_rotate_failed"
	AccessTokenExpired          AccessTokenStatus = "accesstoken_expired"
	AccessTokenForbidden        AccessTokenStatus = "accesstoken_forbidden"
)

// String returns the string representation of the access token status
func (ds AccessTokenStatus) String() string {
	return string(ds)
}

// Into returns the translated string representation of the access token status in the specified language
func (ds AccessTokenStatus) Translate(lang string) string {
	translations := map[string]map[AccessTokenStatus]string{
		"en": {
			AccessTokenBindingFailed:    "Access token binding failed",
			AccessTokenPermissionsFound: "Access token permissions found",
			AccessTokenNameExists:       "Access token with the same name already exists",
			AccessTokenMissingInfo:      "Access token is missing required information",
			AccessTokenFound:            "Access token found",
			AccessTokenNotFound:         "Access token not found",
			AccessTokenAddSuccess:       "Access token added successfully",
			AccessTokenAddFailed:        "Failed to add access token",
			AccessTokenUpdateSuccess:    "Access token updated successfully",
			AccessTokenUpdateFailed:     "Failed to update access token",
			AccessTokenDeleteSuccess:    "Access token deleted successfully",
			AccessTokenDeleteFailed:     "Failed to delete access token",
			AccessTokenRotateSuccess:    "Access token rotated successfully",
			AccessTokenRotateFailed:     "Failed to rotate access token",
			AccessTokenExpired:          "Access token has expired",
			AccessTokenForbidden:        "You are not allowed to manage this access token",
		},
	}

	if langTranslations, exists := translations[lang]; exists {
		if translation, exists := langTranslations[ds]; exists {
			return translation
		}
	}

	// Default to English if language not found or translation doesn't exist
	if enTranslations, exists := translations["en"]; exists {
		if translation, exists := enTranslations[ds]; exists {
			return translation
		}
	}

	// Fallback to the string representation
	return ds.String()
}

// GetAccessTokens
// @Router /profile/token [get]
type GetAccessTokensRequest struct {
}
type GetAccessTokensResponse struct {
	AccessTokens []models.AccessToken `json:"access_tokens"`
}
type GetAccessTokensSuccessResponse struct {
	SuccessResponse
	Data GetAccessTokensResponse `json:"data"`
}
type GetAccessTokensErrorResponse struct {
	ErrorResponse
}

// GetAccessTokenPermissions
// @Router /profile/tokens/permissions [get]
type GetAccessTokenPermissionsRequest struct {
}
type GetAccessTokenPermissionsResponse struct {
	Permissions []models.AccessTokenScope `json:"permissions"`
}
type GetAccessTokenPermissionsSuccessResponse struct {
	SuccessResponse
	Data GetAccessTokenPermissionsResponse `json:"data"`
}
type GetAccessTokenPermissionsErrorResponse struct {
	ErrorResponse
}

// AddAccessToken
// @Router /profile/token [post]
type AddAccessTokenRequest struct {
	Token models.AccessToken `json:"token"`
}
type AddAccessTokenResponse struct {
	Token models.AccessToken `json:"token"`
}
type AddAccessTokenSuccessResponse struct {
	SuccessResponse
	Data AddAccessTokenResponse `json:"data"`
}
type AddAccessTokenErrorResponse struct {
	ErrorResponse
}

// UpdateAccessToken
// @Router /profile/token/{id} [put]
type UpdateAccessTokenRequest struct {
	Token models.AccessToken `json:"token"`
}
type UpdateAccessTokenResponse struct {
	Token models.AccessToken `json:"token"`
}
type UpdateAccessTokenSuccessResponse struct {
	SuccessResponse
	Data UpdateAccessTokenResponse `json:"data"`
}
type UpdateAccessTokenErrorResponse struct {
	ErrorResponse
}

// DeleteAccessToken
// @Router /profile/token/{id} [delete]
type DeleteAccessTokenRequest struct {
}
type DeleteAccessTokenResponse struct {
}
type DeleteAccessTokenSuccessResponse struct {
	SuccessResponse
	Data DeleteAccessTokenResponse `json:"data"`
}
type DeleteAccessTokenErrorResponse struct {
	ErrorResponse
}

// RotateAccessToken
// @Router /profile/tokens/{id}/rotate [post]
// Rotation issues a new secret with the same name, description, scopes and
// expiration, and revokes the previous secret immediately. The returned token
// contains the full secret, which is only shown once.
type RotateAccessTokenRequest struct {
}
type RotateAccessTokenResponse struct {
	Token models.AccessToken `json:"token"`
}
type RotateAccessTokenSuccessResponse struct {
	SuccessResponse
	Data RotateAccessTokenResponse `json:"data"`
}
type RotateAccessTokenErrorResponse struct {
	ErrorResponse
}
