package api

// RedactionStatus represents specific status codes for redaction operations
type RedactionStatus string

const (
	// Queu status codes
	RedactionQueueSubscribed RedactionStatus = "redaction_queue_subscribed"
	RedactionQueueStarted    RedactionStatus = "redaction_queue_started"
	RedactionQueueCompleted  RedactionStatus = "redaction_queue_completed"
	RedactionQueueFailed     RedactionStatus = "redaction_queue_failed"

	// Trace status codes
	RedactionTracingStarted   RedactionStatus = "redaction_tracing_started"
	RedactionTracingCompleted RedactionStatus = "redaction_tracing_completed"
	RedactionTracingFailed    RedactionStatus = "redaction_tracing_failed"

	// Stage status codes
	RedactionStageStart         RedactionStatus = "redaction_stage_start"
	RedactionStageEnd           RedactionStatus = "redaction_stage_end"
	RedactionDownloadStart      RedactionStatus = "redaction_download_start"
	RedactionDownloadFailed     RedactionStatus = "redaction_download_failed"
	RedactionDownloadSuccess    RedactionStatus = "redaction_download_success"
	RedactionProcessingStart    RedactionStatus = "redaction_processing_start"
	RedactionProcessingPrepare  RedactionStatus = "redaction_processing_prepare"
	RedactionProcessingLoop     RedactionStatus = "redaction_processing_loop"
	RedactionProcessingRedact   RedactionStatus = "redaction_processing_redact"
	RedactionProcessingEnd      RedactionStatus = "redaction_processing_end"
	RedactionUploadStart        RedactionStatus = "redaction_upload_start"
	RedactionUploadFailed       RedactionStatus = "redaction_upload_failed"
	RedactionUploadSuccess      RedactionStatus = "redaction_upload_success"
	RedactionForwardingAnalysis RedactionStatus = "redaction_forwarding_analysis"
)

// String returns the string representation of the redaction status
func (rs RedactionStatus) String() string {
	return string(rs)
}

// Into returns the translated string representation of the redaction status in the specified language
func (rs RedactionStatus) Translate(lang string) string {
	translations := map[string]map[RedactionStatus]string{
		"en": {
			// Queue status codes
			RedactionQueueSubscribed: "Subscribed to redaction queue",
			RedactionQueueStarted:    "Redaction queue processing started",
			RedactionQueueCompleted:  "Redaction queue processing completed",
			RedactionQueueFailed:     "Redaction queue processing failed",

			// Tracing status codes
			RedactionTracingStarted:   "Redaction tracing started",
			RedactionTracingCompleted: "Redaction tracing completed",
			RedactionTracingFailed:    "Redaction tracing failed",

			// Stage status codes
			RedactionStageStart:         "Starting redaction stage",
			RedactionStageEnd:           "Redaction stage completed",
			RedactionDownloadStart:      "Starting download for redaction",
			RedactionDownloadFailed:     "Failed to download for redaction",
			RedactionDownloadSuccess:    "Download for redaction completed successfully",
			RedactionProcessingStart:    "Beginning redaction processing",
			RedactionProcessingPrepare:  "Preparing for redaction processing",
			RedactionProcessingLoop:     "Processing redaction loop",
			RedactionProcessingRedact:   "Applying redactions",
			RedactionProcessingEnd:      "Redaction processing finished",
			RedactionUploadStart:        "Starting upload of redacted file",
			RedactionUploadFailed:       "Failed to upload redacted file",
			RedactionUploadSuccess:      "Redacted file uploaded successfully",
			RedactionForwardingAnalysis: "Forwarding event to analysis pipeline",
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
