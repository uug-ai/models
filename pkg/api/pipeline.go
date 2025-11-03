package api

import "github.com/uug-ai/models/pkg/models"

type RabbitMQStatus string

const (
	RabbitMQConnected            RabbitMQStatus = "rabbitmq_connected"
	RabbitMQConnectionFailed     RabbitMQStatus = "rabbitmq_connection_failed"
	RabbitMQDisconnected         RabbitMQStatus = "rabbitmq_disconnected"
	RabbitMQMessageBindingFailed RabbitMQStatus = "rabbitmq_message_binding_failed"
	RabbitMQAcknowledged         RabbitMQStatus = "rabbitmq_acknowledged"
	RabbitMQNotAcknowledged      RabbitMQStatus = "rabbitmq_not_acknowledged"
	RabbitMQChannelDoesNotExist  RabbitMQStatus = "rabbitmq_channel_does_not_exist"
	RabbitMQFailedToConsume      RabbitMQStatus = "rabbitmq_failed_to_consume"
	RabbitMQFailedToPublish      RabbitMQStatus = "rabbitmq_failed_to_publish"
	RabbitMMessagePublished      RabbitMQStatus = "rabbitmq_message_published"
)

// String returns the string representation of the RabbitMQ status
func (rs RabbitMQStatus) String() string {
	return string(rs)
}

// Translate returns the translated string representation of the RabbitMQ status in the specified language
func (rs RabbitMQStatus) Translate(lang string) string {
	translations := map[string]map[RabbitMQStatus]string{
		"en": {
			RabbitMQConnected:            "RabbitMQ connected",
			RabbitMQConnectionFailed:     "RabbitMQ connection failed",
			RabbitMQDisconnected:         "RabbitMQ disconnected",
			RabbitMQMessageBindingFailed: "RabbitMQ message binding failed",
			RabbitMQAcknowledged:         "RabbitMQ message acknowledged",
			RabbitMQNotAcknowledged:      "RabbitMQ message not acknowledged",
			RabbitMQChannelDoesNotExist:  "RabbitMQ channel does not exist",
			RabbitMQFailedToConsume:      "RabbitMQ failed to consume message",
			RabbitMQFailedToPublish:      "RabbitMQ failed to publish message",
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

// MonitorStage represents the initial stage of media processing
type MonitorStatus string

const (
	// Queue status codes
	MonitorQueueStarted    MonitorStatus = "monitor_queue_started"
	MonitorQueueSubscribed MonitorStatus = "monitor_queue_subscribed"
	MonitorQueueFailed     MonitorStatus = "monitor_queue_failed"
	MonitorQueueCompleted  MonitorStatus = "monitor_queue_completed"

	// Trace status codes
	MonitorTracingStarted   MonitorStatus = "monitor_tracing_started"
	MonitorTracingCompleted MonitorStatus = "monitor_tracing_completed"
	MonitorTracingFailed    MonitorStatus = "monitor_tracing_failed"

	// Stage status codes
	MonitorStageStart           MonitorStatus = "monitor_stage_start"
	MonitorStageEnd             MonitorStatus = "monitor_stage_end"
	MonitorStageMissing         MonitorStatus = "monitor_stage_missing"
	MonitorStageUserNotFound    MonitorStatus = "monitor_stage_user_not_found"
	MonitorOrganizationNotFound MonitorStatus = "monitor_stage_organization_not_found"
)

// String returns the string representation of the monitor status
func (ms MonitorStatus) String() string {
	return string(ms)
}

// Translate returns the translated string representation of the monitor status in the specified language
func (ms MonitorStatus) Translate(lang string) string {
	translations := map[string]map[MonitorStatus]string{
		"en": {
			MonitorStageStart:           "Starting monitor stage",
			MonitorStageEnd:             "Monitor stage completed",
			MonitorStageMissing:         "Monitor stage missing",
			MonitorStageUserNotFound:    "User not found during monitor stage",
			MonitorOrganizationNotFound: "Organization not found during monitor stage",
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

// RedactionStatus represents specific status codes for redaction operations
type RedactionStatus string

const (
	// Queu status codes
	RedactionQueueStarted    RedactionStatus = "redaction_queue_started"
	RedactionQueueSubscribed RedactionStatus = "redaction_queue_subscribed"
	RedactionQueueFailed     RedactionStatus = "redaction_queue_failed"
	RedactionQueueCompleted  RedactionStatus = "redaction_queue_completed"

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
			RedactionQueueStarted:    "Redaction queue processing started",
			RedactionQueueSubscribed: "Subscribed to redaction queue",
			RedactionQueueFailed:     "Redaction queue processing failed",
			RedactionQueueCompleted:  "Redaction queue processing completed",

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

type RedactionEvent struct {
	AllFrameCoordinates map[string][]models.TrackBox `json:"allFrameCoordinates"`
}
