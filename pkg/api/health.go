package api

import "github.com/uug-ai/models/pkg/models"

// HealthStatus represents specific status codes for health operations
type HealthStatus string

const (
	NotHealthyLicense                    HealthStatus = "not_healthy_license"
	NotHealthyDatabase                   HealthStatus = "not_healthy_database"
	NotHealthyQueue                      HealthStatus = "not_healthy_queue"
	NotHealthyDatabaseAndQueue           HealthStatus = "not_healthy_database_and_queue"
	NotHealthyLicenseAndQueue            HealthStatus = "not_healthy_license_and_queue"
	NotHealthyLicenseAndDatabase         HealthStatus = "not_healthy_license_and_database"
	NotHealthyDatabaseAndLicenseAndQueue HealthStatus = "not_healthy_database_and_license_and_queue"
	Healthy                              HealthStatus = "healthy"
)

// String returns the string representation of the marker status
func (ms HealthStatus) String() string {
	return string(ms)
}

// Into returns the translated string representation of the marker status in the specified language
func (ms HealthStatus) Translate(lang string) string {
	translations := map[string]map[HealthStatus]string{
		"en": {
			NotHealthyDatabase:                   "Database is not healthy",
			NotHealthyQueue:                      "Queue is not healthy",
			NotHealthyLicense:                    "License is not healthy",
			NotHealthyDatabaseAndQueue:           "Database and queue are not healthy",
			NotHealthyLicenseAndQueue:            "License is not healthy and queue is not healthy",
			NotHealthyLicenseAndDatabase:         "License is not healthy and database is not healthy",
			NotHealthyDatabaseAndLicenseAndQueue: "Database is not healthy, license is not healthy, and queue is not healthy",
			Healthy:                              "Healthy",
		},
		"es": {
			NotHealthyDatabase:                   "La base de datos no está saludable",
			NotHealthyQueue:                      "La cola no está saludable",
			NotHealthyLicense:                    "La licencia no está saludable",
			NotHealthyDatabaseAndQueue:           "La base de datos y la cola no están saludables",
			NotHealthyLicenseAndQueue:            "La licencia no está saludable y la cola no está saludable",
			NotHealthyLicenseAndDatabase:         "La licencia no está saludable y la base de datos no está saludable",
			NotHealthyDatabaseAndLicenseAndQueue: "La base de datos no está saludable, la licencia no está saludable y la cola no está saludable",
			Healthy:                              "Saludable",
		},
		"fr": {
			NotHealthyDatabase:                   "La base de données n'est pas saine",
			NotHealthyQueue:                      "La file d'attente n'est pas saine",
			NotHealthyLicense:                    "La licence n'est pas saine",
			NotHealthyDatabaseAndQueue:           "La base de données et la file d'attente ne sont pas saines",
			NotHealthyLicenseAndQueue:            "La licence n'est pas saine et la file d'attente n'est pas saine",
			NotHealthyLicenseAndDatabase:         "La licence n'est pas saine et la base de données n'est pas saine",
			NotHealthyDatabaseAndLicenseAndQueue: "La base de données n'est pas saine, la licence n'est pas saine et la file d'attente n'est pas saine",
			Healthy:                              "Saine",
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
