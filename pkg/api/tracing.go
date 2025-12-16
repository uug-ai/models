package api

type TracingStatus string

const (
	TracingStatusConnected        TracingStatus = "tracing_status_connected"
	TracingStatusConnectionFailed TracingStatus = "tracing_status_connection_failed"
	TracingStatusDisconnected     TracingStatus = "tracing_status_disconnected"
	TracingStatusDataSent         TracingStatus = "tracing_status_data_sent"
	TracingStatusDataSendFailed   TracingStatus = "tracing_status_data_send_failed"
	TraceCreationFailed           TracingStatus = "trace_creation_failed"
)

// String returns the string representation of the Tracing status
func (rs TracingStatus) String() string {
	return string(rs)
}

// Translate returns the translated string representation of the Tracing status in the specified language
func (rs TracingStatus) Translate(lang string) string {
	translations := map[string]map[TracingStatus]string{
		"en": {
			TracingStatusConnected:        "Tracing connected",
			TracingStatusConnectionFailed: "Tracing connection failed",
			TracingStatusDisconnected:     "Tracing disconnected",
			TracingStatusDataSent:         "Tracing data sent",
			TracingStatusDataSendFailed:   "Tracing data send failed",
		},
	}

	if langTranslations, exists := translations[lang]; exists {
		if translation, exists := langTranslations[rs]; exists {
			return translation
		}
	}

	// Default to English if language not found or translation doesn't exist
	if enTranslations, exists := translations["en"]; exists {
		if translation, exists := enTranslations[rs]; exists {
			return translation
		}
	}

	// Fallback to the string representation
	return rs.String()
}
