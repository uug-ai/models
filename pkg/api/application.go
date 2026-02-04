package api

// ApplicationStatus represents specific status codes for application operations
type ApplicationStatus string

const (
	PongSuccess     ApplicationStatus = "pong_success"
	DatabaseSuccess ApplicationStatus = "database_success"
	DatabaseError   ApplicationStatus = "database_error"
	QueueSuccess    ApplicationStatus = "queue_success"
	QueueError      ApplicationStatus = "queue_error"
	CacheSuccess    ApplicationStatus = "cache_success"
	CacheError      ApplicationStatus = "cache_error"
	TypeCastFailed  ApplicationStatus = "type_cast_failed"
)

// String returns the string representation of the application status
func (as ApplicationStatus) String() string {
	return string(as)
}

// Translate returns the translated string representation of the application status in the specified language
func (as ApplicationStatus) Translate(lang string) string {
	translations := map[string]map[ApplicationStatus]string{
		"en": {
			PongSuccess:     "Pong successful",
			DatabaseSuccess: "Database operation successful",
			DatabaseError:   "Database operation failed",
			QueueSuccess:    "Queue operation successful",
			QueueError:      "Queue operation failed",
			CacheSuccess:    "Cache operation successful",
			CacheError:      "Cache operation failed",
			TypeCastFailed:  "Type casting failed",
		},
		"es": {
			PongSuccess:     "Pong exitoso",
			DatabaseSuccess: "Operación de base de datos exitosa",
			DatabaseError:   "Error en la operación de base de datos",
			QueueSuccess:    "Operación de cola exitosa",
			QueueError:      "Error en la operación de cola",
			CacheSuccess:    "Operación de caché exitosa",
			CacheError:      "Error en la operación de caché",
			TypeCastFailed:  "Error de conversión de tipo",
		},
		"fr": {
			PongSuccess:     "Pong réussi",
			DatabaseSuccess: "Opération de base de données réussie",
			DatabaseError:   "Échec de l'opération de base de données",
			QueueSuccess:    "Opération de file d'attente réussie",
			QueueError:      "Échec de l'opération de file d'attente",
			CacheSuccess:    "Opération de cache réussie",
			CacheError:      "Échec de l'opération de cache",
			TypeCastFailed:  "Échec de la conversion de type",
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
