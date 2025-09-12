package models

// AccessTokenStatus represents specific status codes for access token operations
type AccessTokenStatus string

const (
	ACCESSTOKEN_BINDING_FAILED AccessTokenStatus = "accesstoken_binding_failed"
	ACCESSTOKEN_NAME_EXISTS    AccessTokenStatus = "accesstoken_name_exists"
	ACCESSTOKEN_MISSING_INFO   AccessTokenStatus = "accesstoken_missing_info"
	ACCESSTOKEN_FOUND          AccessTokenStatus = "accesstoken_found"
	ACCESSTOKEN_NOT_FOUND      AccessTokenStatus = "accesstoken_not_found"
	ACCESSTOKEN_ADD_SUCCESS    AccessTokenStatus = "accesstoken_add_success"
	ACCESSTOKEN_ADD_FAILED     AccessTokenStatus = "accesstoken_add_failed"
	ACCESSTOKEN_UPDATE_SUCCESS AccessTokenStatus = "accesstoken_update_success"
	ACCESSTOKEN_UPDATE_FAILED  AccessTokenStatus = "accesstoken_update_failed"
	ACCESSTOKEN_DELETE_SUCCESS AccessTokenStatus = "accesstoken_delete_success"
	ACCESSTOKEN_DELETE_FAILED  AccessTokenStatus = "accesstoken_delete_failed"
)

// String returns the string representation of the access token status
func (ds AccessTokenStatus) String() string {
	return string(ds)
}

// Into returns the translated string representation of the access token status in the specified language
func (ds AccessTokenStatus) Translate(lang string) string {
	translations := map[string]map[AccessTokenStatus]string{
		"en": {
			ACCESSTOKEN_BINDING_FAILED: "Access token binding failed",
			ACCESSTOKEN_NAME_EXISTS:    "Access token with the same name already exists",
			ACCESSTOKEN_MISSING_INFO:   "Access token is missing required information",
			ACCESSTOKEN_FOUND:          "Access token found",
			ACCESSTOKEN_NOT_FOUND:      "Access token not found",
			ACCESSTOKEN_ADD_SUCCESS:    "Access token added successfully",
			ACCESSTOKEN_ADD_FAILED:     "Failed to add access token",
			ACCESSTOKEN_UPDATE_SUCCESS: "Access token updated successfully",
			ACCESSTOKEN_UPDATE_FAILED:  "Failed to update access token",
			ACCESSTOKEN_DELETE_SUCCESS: "Access token deleted successfully",
			ACCESSTOKEN_DELETE_FAILED:  "Failed to delete access token",
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

