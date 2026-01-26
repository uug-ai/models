package api

// LicenseStatus represents specific status codes for license operations
type LicenseStatus string

const (
	LicenseInvalid LicenseStatus = "license_invalid"
)

// String returns the string representation of the license status
func (ls LicenseStatus) String() string {
	return string(ls)
}

// Translate returns the translated string representation of the license status in the specified language
func (ls LicenseStatus) Translate(lang string) string {
	translations := map[string]map[LicenseStatus]string{
		"en": {
			LicenseInvalid: "License is invalid",
		},
		"es": {
			LicenseInvalid: "La licencia no es válida",
		},
		"fr": {
			LicenseInvalid: "La licence n'est pas valide",
		},
	}

	if langTranslations, exists := translations[lang]; exists {
		if translation, exists := langTranslations[ls]; exists {
			return translation
		}
	}

	// Default to English if language not found or translation doesn't exist
	if enTranslations, exists := translations["en"]; exists {
		if translation, exists := enTranslations[ls]; exists {
			return translation
		}
	}

	// Fallback to the string representation
	return ls.String()
}
