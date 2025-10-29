package api

// ApplicationStatus represents specific status codes for application operations
type ApplicationStatus string

const (
	PongSuccess ApplicationStatus = "pong_success"
	DatabaseSuccess ApplicationStatus = "database_success"
	DatabaseError ApplicationStatus = "database_error"
	QueueSuccess ApplicationStatus = "queue_success"
	QueueError ApplicationStatus = "queue_error"
	CacheSuccess ApplicationStatus = "cache_success"
	CacheError ApplicationStatus = "cache_error"
)

// String returns the string representation of the application status
func (as ApplicationStatus) String() string {
	return string(as)
}

// Translate returns the translated string representation of the application status in the specified language
func (as ApplicationStatus) Translate(lang string) string {
	translations := map[string]map[ApplicationStatus]string{
		"en": {
			PongSuccess: "Pong successful",
		},
		"es": {
			PongSuccess: "Pong exitoso",
		},
		"fr": {
			PongSuccess: "Pong réussi",
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
