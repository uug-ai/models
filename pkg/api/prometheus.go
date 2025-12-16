package api

type PrometheusStatus string

const (
	PrometheusServiceStarted PrometheusStatus = "prometheus_service_started"
	PrometheusServiceStopped PrometheusStatus = "prometheus_service_stopped"
)

func (ps PrometheusStatus) String() string {
	return string(ps)
}

func (ps PrometheusStatus) Translate(lang string) string {
	translations := map[string]map[PrometheusStatus]string{
		"en": {
			PrometheusServiceStarted: "Prometheus service started",
			PrometheusServiceStopped: "Prometheus service stopped",
		},
	}

	if langTranslations, exists := translations[lang]; exists {
		if translation, exists := langTranslations[ps]; exists {
			return translation
		}
	}

	// Default to English if language not found or translation doesn't exist
	if enTranslations, exists := translations["en"]; exists {
		if translation, exists := enTranslations[ps]; exists {
			return translation
		}
	}

	// Fallback to the string representation
	return ps.String()
}
