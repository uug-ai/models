package api

import "github.com/uug-ai/models/pkg/models"

// HealthStatus represents specific status codes for health operations
type HealthStatus string

const (
	NotHealtyDatabase                   HealthStatus = "no_healthy_database"
	NotHealthyQueue                     HealthStatus = "no_healthy_queue"
	NotHealtyLicense                    HealthStatus = "no_valid_license"
	NotHealthDatabaseAndQueue           HealthStatus = "no_healthy_database_and_queue"
	NotHealthyLicenseAndQueue           HealthStatus = "no_valid_license_and_queue"
	NotHealthyLicenseAndDatabase        HealthStatus = "no_valid_license_and_healthy_database"
	NotHealthDatabaseAndLicenseAndQueue HealthStatus = "no_healthy_database_no_valid_license_and_queue"
	Healthy                             HealthStatus = "healthy"
)

// String returns the string representation of the marker status
func (ms HealthStatus) String() string {
	return string(ms)
}

// Into returns the translated string representation of the marker status in the specified language
func (ms HealthStatus) Translate(lang string) string {
	translations := map[string]map[HealthStatus]string{
		"en": {
			NotHealtyDatabase:                   "Database is not healthy",
			NotHealthyQueue:                     "Queue is not healthy",
			NotHealtyLicense:                    "License is not valid",
			NotHealthDatabaseAndQueue:           "Database and queue are not healthy",
			NotHealthyLicenseAndQueue:           "License is not valid and queue is not healthy",
			NotHealthyLicenseAndDatabase:        "License is not valid and database is not healthy",
			NotHealthDatabaseAndLicenseAndQueue: "Database is not healthy, license is not valid, and queue is not healthy",
			Healthy:                             "Healthy",
		},
		"es": {
			NotHealtyDatabase:                   "La base de datos no está saludable",
			NotHealthyQueue:                     "La cola no está saludable",
			NotHealtyLicense:                    "La licencia no es válida",
			NotHealthDatabaseAndQueue:           "La base de datos y la cola no están saludables",
			NotHealthyLicenseAndQueue:           "La licencia no es válida y la cola no está saludable",
			NotHealthyLicenseAndDatabase:        "La licencia no es válida y la base de datos no está saludable",
			NotHealthDatabaseAndLicenseAndQueue: "La base de datos no está saludable, la licencia no es válida y la cola no está saludable",
			Healthy:                             "Saludable",
		},
		"fr": {
			NotHealtyDatabase:                   "La base de données n'est pas saine",
			NotHealthyQueue:                     "La file d'attente n'est pas saine",
			NotHealtyLicense:                    "La licence n'est pas valide",
			NotHealthDatabaseAndQueue:           "La base de données et la file d'attente ne sont pas saines",
			NotHealthyLicenseAndQueue:           "La licence n'est pas valide et la file d'attente n'est pas saine",
			NotHealthyLicenseAndDatabase:        "La licence n'est pas valide et la base de données n'est pas saine",
			NotHealthDatabaseAndLicenseAndQueue: "La base de données n'est pas saine, la licence n'est pas valide et la file d'attente n'est pas saine",
			Healthy:                             "Saine",
		},
	}

	if langTranslations, exists := translations[lang]; exists {
		if translation, exists := langTranslations[ms]; exists {
			return translation
		}
	}

	// Default to English if language not found or translation doesn't exist
	if enTranslations, exists := translations["en"]; exists {
		if translation, exists := enTranslations[ms]; exists {
			return translation
		}
	}

	// Fallback to the string representation
	return ms.String()
}

// GetHealth
// @Router /health [get]
type GetHealthRequest struct {
}
type GetHealthResponse struct {
	Health models.Health `json:"health"`
}
type GetHealthSuccessResponse struct {
	SuccessResponse
	Data GetHealthResponse `json:"data"`
}
type GetHealthErrorResponse struct {
	ErrorResponse
}
