package api

import (
	"github.com/uug-ai/models/pkg/models"
)

// RuntimeConfigStatus represents specific status codes for runtime
// configuration operations.
type RuntimeConfigStatus string

const (
	RuntimeConfigFound      RuntimeConfigStatus = "runtime_config_found"
	RuntimeConfigFetchError RuntimeConfigStatus = "runtime_config_fetch_error"
)

// String returns the string representation of the RuntimeConfig status.
func (s RuntimeConfigStatus) String() string {
	return string(s)
}

// Translate returns the translated string representation of the
// RuntimeConfig status in the specified language.
func (s RuntimeConfigStatus) Translate(lang string) string {
	translations := map[string]map[RuntimeConfigStatus]string{
		"en": {
			RuntimeConfigFound:      "Runtime configuration retrieved",
			RuntimeConfigFetchError: "Failed to retrieve runtime configuration",
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

// GetRuntimeConfig response types
// @Router /runtime/config [get]
type GetRuntimeConfigResponse struct {
	RuntimeConfig models.RuntimeConfig `json:"runtimeConfig"`
}

type GetRuntimeConfigSuccessResponse struct {
	SuccessResponse
	Data GetRuntimeConfigResponse `json:"data"`
}

type GetRuntimeConfigErrorResponse struct {
	ErrorResponse
}
