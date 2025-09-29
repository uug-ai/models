package api

// AuthenticationStatus represents specific status codes for authentication operations
type AuthenticationStatus string

const (
	AuthenticationFailed        AuthenticationStatus = "authentication_failed"
	AuthenticationSuccess       AuthenticationStatus = "authentication_success"
	AuthenticationExpired       AuthenticationStatus = "authentication_expired"
	AuthenticationRevoked       AuthenticationStatus = "authentication_revoked"
	AuthenticationNotFound      AuthenticationStatus = "authentication_not_found"
	AuthenticationInvalid       AuthenticationStatus = "authentication_invalid"
	AuthenticationUnknown       AuthenticationStatus = "authentication_unknown"
	AuthenticationScopesMissing AuthenticationStatus = "authentication_scopes_missing"
)

// String returns the string representation of the authentication status
func (as AuthenticationStatus) String() string {
	return string(as)
}

// Into returns the translated string representation of the authentication status in the specified language
func (as AuthenticationStatus) Translate(lang string) string {
	translations := map[string]map[AuthenticationStatus]string{
		"en": {
			AuthenticationFailed:        "Authentication failed",
			AuthenticationSuccess:       "Authentication successful",
			AuthenticationExpired:       "Authentication expired",
			AuthenticationRevoked:       "Authentication revoked",
			AuthenticationNotFound:      "Authentication not found",
			AuthenticationInvalid:       "Authentication invalid",
			AuthenticationUnknown:       "Authentication unknown",
			AuthenticationScopesMissing: "Authentication scopes missing",
		},
	}

	if langTranslations, exists := translations[lang]; exists {
		if translation, exists := langTranslations[as]; exists {
			return translation
		}
	}

	// Default to English if language not found or translation doesn't exist
	if enTranslations, exists := translations["en"]; exists {
		if translation, exists := enTranslations[as]; exists {
			return translation
		}
	}

	// Fallback to the string representation
	return as.String()
}
