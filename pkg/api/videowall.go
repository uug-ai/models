package api

// VideowallStatus represents specific status codes for videowall operations.
type VideowallStatus string

const (
	VideowallBindingFailed VideowallStatus = "videowall_binding_failed"
	VideowallMissingInfo   VideowallStatus = "videowall_missing_info"
	VideowallFound         VideowallStatus = "videowall_found"
	VideowallNotFound      VideowallStatus = "videowall_not_found"
)

// String returns the string representation of the videowall status.
func (cs VideowallStatus) String() string {
	return string(cs)
}

// Translate returns the translated string representation of the videowall status in the specified language.
func (cs VideowallStatus) Translate(lang string) string {
	translations := map[string]map[VideowallStatus]string{
		"en": {
			VideowallBindingFailed: "Videowall binding failed",
			VideowallMissingInfo:   "Videowall missing information",
			VideowallFound:         "Videowall found",
			VideowallNotFound:      "Videowall not found",
		},
	}

	if langTranslations, exists := translations[lang]; exists {
		if translation, exists := langTranslations[cs]; exists {
			return translation
		}
	}

	if enTranslations, exists := translations["en"]; exists {
		if translation, exists := enTranslations[cs]; exists {
			return translation
		}
	}

	return cs.String()
}
