package api

import "github.com/uug-ai/models/pkg/models"

const (
	PipelineStarted         string = "pipeline_started"
	PipelineInProgress      string = "pipeline_in_progress"
	PipelineCompleted       string = "pipeline_completed"
	PipelineFailed          string = "pipeline_failed"
	PipelineCancelled       string = "pipeline_cancelled"
	PipelinePending         string = "pipeline_pending"
	PipelineRetrying        string = "pipeline_retrying"
	PipelinePaused          string = "pipeline_paused"
	PipelineResumed         string = "pipeline_resumed"
	PipelineValidationError string = "pipeline_validation_error"
	PipelineError           string = "pipeline_error"
	PipelineInfo            string = "pipeline_info"
	PipelineDebug           string = "pipeline_debug"
	PipelineWarning         string = "pipeline_warning"
)

type PipelineStatus string

const (
	UserMissing                PipelineStatus = "user_missing"
	TraceIdMissing             PipelineStatus = "trace_id_missing"
	UserEmailEmpty             PipelineStatus = "user_email_empty"
	MediaMissing               PipelineStatus = "media_missing"
	IONotFound                 PipelineStatus = "io_not_found"
	IODecodeError              PipelineStatus = "io_decode_error"
	SignedUrlFailed            PipelineStatus = "signed_url_failed"
	ThumbnailMissing           PipelineStatus = "thumbnail_missing"
	QueueCreationFailed        PipelineStatus = "queue_creation_failed"
	PanicRecovered             PipelineStatus = "panic_recovered"
	DeadLetterMarshalFailed    PipelineStatus = "dead_letter_marshal_failed"
	DeadLetterQueueSendSuccess PipelineStatus = "dead_letter_queue_send_success"
	DeadLetterQueueSendFailed  PipelineStatus = "dead_letter_queue_send_failed"
	QueueReadMessagesFailed    PipelineStatus = "queue_read_messages_failed"
	QueueReconnectionFailed    PipelineStatus = "queue_reconnection_failed"
)

func (ps PipelineStatus) String() string {
	return string(ps)
}

func (ps PipelineStatus) Translate(lang string) string {
	translations := map[string]map[PipelineStatus]string{
		"en": {
			UserMissing:                "User is missing",
			TraceIdMissing:             "Trace ID is missing",
			UserEmailEmpty:             "User email is empty",
			MediaMissing:               "Media is missing",
			IONotFound:                 "IO not found",
			IODecodeError:              "Error decoding IO information",
			SignedUrlFailed:            "Failed to generate signed URL",
			ThumbnailMissing:           "Thumbnail is missing",
			QueueCreationFailed:        "Failed to create queue",
			PanicRecovered:             "Panic recovered during pipeline execution",
			DeadLetterMarshalFailed:    "Failed to marshal message for dead letter queue",
			DeadLetterQueueSendSuccess: "Successfully sent message to dead letter queue",
			DeadLetterQueueSendFailed:  "Failed to send message to dead letter queue",
			QueueReadMessagesFailed:    "Failed to read messages from queue",
			QueueReconnectionFailed:    "Failed to reconnect to queue",
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
	RabbitMQMessagePublished     RabbitMQStatus = "rabbitmq_message_published"
	RabbitMQFailedToDeclareQueue RabbitMQStatus = "rabbitmq_failed_to_declare_queue"
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
			RabbitMQFailedToDeclareQueue: "RabbitMQ failed to declare queue",
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
	MonitorUserNotFound         MonitorStatus = "monitor_user_not_found"
	MonitorOrganizationNotFound MonitorStatus = "monitor_organization_not_found"
	MonitorProcessingStart      MonitorStatus = "monitor_processing_start"
	MonitorProcessingEnd        MonitorStatus = "monitor_processing_end"
	MonitorProcessingFailed     MonitorStatus = "monitor_processing_failed"
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
			MonitorUserNotFound:         "User not found during monitor stage",
			MonitorOrganizationNotFound: "Organization not found during monitor stage",
			MonitorProcessingStart:      "Starting monitor processing",
			MonitorProcessingEnd:        "Monitor processing completed",
			MonitorProcessingFailed:     "Monitor processing failed",
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

// --- THUMBNAILS ---

type ThumbnailStatus string

const (
	// Queue status codes
	ThumbnailQueueStarted    ThumbnailStatus = "thumbnail_queue_started"
	ThumbnailQueueSubscribed ThumbnailStatus = "thumbnail_queue_subscribed"
	ThumbnailQueueFailed     ThumbnailStatus = "thumbnail_queue_failed"
	ThumbnailQueueCompleted  ThumbnailStatus = "thumbnail_queue_completed"

	// Trace status codes
	ThumbnailTracingStarted   ThumbnailStatus = "thumbnail_tracing_started"
	ThumbnailTracingCompleted ThumbnailStatus = "thumbnail_tracing_completed"
	ThumbnailTracingFailed    ThumbnailStatus = "thumbnail_tracing_failed"

	// Stage status codes
	ThumbnailStageStart       ThumbnailStatus = "thumbnail_stage_start"
	ThumbnailStageEnd         ThumbnailStatus = "thumbnail_stage_end"
	ThumbnailCreationFailed   ThumbnailStatus = "thumbnail_creation_failed"
	ThumbnailProcessingFailed ThumbnailStatus = "thumbnail_processing_failed"
	ThumbnailGenerated        ThumbnailStatus = "thumbnail_generated"
)

// String returns the string representation of the Thumbnail status
func (ms ThumbnailStatus) String() string {
	return string(ms)
}

// Translate returns the translated string representation of the Thumbnail status in the specified language
func (ms ThumbnailStatus) Translate(lang string) string {
	translations := map[string]map[ThumbnailStatus]string{
		"en": {
			ThumbnailQueueStarted:     "Thumbnail queue processing started",
			ThumbnailQueueSubscribed:  "Subscribed to Thumbnail queue",
			ThumbnailQueueFailed:      "Thumbnail queue processing failed",
			ThumbnailQueueCompleted:   "Thumbnail queue processing completed",
			ThumbnailTracingStarted:   "Thumbnail tracing started",
			ThumbnailTracingCompleted: "Thumbnail tracing completed",
			ThumbnailTracingFailed:    "Thumbnail tracing failed",
			ThumbnailStageStart:       "Starting Thumbnail stage",
			ThumbnailStageEnd:         "Thumbnail stage completed",
			ThumbnailCreationFailed:   "Thumbnail creation failed",
			ThumbnailProcessingFailed: "Thumbnail processing failed",
			ThumbnailGenerated:        "Thumbnail generated successfully",
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

// --- DOMINANT COLORS ---

type DominantcolorsStatus string

const (
	// Queue status codes
	DominantcolorsQueueStarted    DominantcolorsStatus = "dominantcolors_queue_started"
	DominantcolorsQueueSubscribed DominantcolorsStatus = "dominantcolors_queue_subscribed"
	DominantcolorsQueueFailed     DominantcolorsStatus = "dominantcolors_queue_failed"
	DominantcolorsQueueCompleted  DominantcolorsStatus = "dominantcolors_queue_completed"

	// Trace status codes
	DominantcolorsTracingStarted   DominantcolorsStatus = "dominantcolors_tracing_started"
	DominantcolorsTracingCompleted DominantcolorsStatus = "dominantcolors_tracing_completed"
	DominantcolorsTracingFailed    DominantcolorsStatus = "dominantcolors_tracing_failed"

	// Stage status codes
	DominantcolorsStageStart       DominantcolorsStatus = "dominantcolors_stage_start"
	DominantcolorsStageEnd         DominantcolorsStatus = "dominantcolors_stage_end"
	DominantcolorsCreationFailed   DominantcolorsStatus = "dominantcolors_creation_failed"
	DominantcolorsProcessingFailed DominantcolorsStatus = "dominantcolors_processing_failed"
	DominantColorsCalculated       DominantcolorsStatus = "dominantcolors_calculated"
)

// String returns the string representation of the Dominantcolors status
func (ms DominantcolorsStatus) String() string {
	return string(ms)
}

// Translate returns the translated string representation of the Dominantcolors status in the specified language
func (ms DominantcolorsStatus) Translate(lang string) string {
	translations := map[string]map[DominantcolorsStatus]string{
		"en": {
			DominantcolorsQueueStarted:     "Dominantcolors queue processing started",
			DominantcolorsQueueSubscribed:  "Subscribed to Dominantcolors queue",
			DominantcolorsQueueFailed:      "Dominantcolors queue processing failed",
			DominantcolorsQueueCompleted:   "Dominantcolors queue processing completed",
			DominantcolorsTracingStarted:   "Dominantcolors tracing started",
			DominantcolorsTracingCompleted: "Dominantcolors tracing completed",
			DominantcolorsTracingFailed:    "Dominantcolors tracing failed",
			DominantcolorsStageStart:       "Starting Dominantcolors stage",
			DominantcolorsStageEnd:         "Dominantcolors stage completed",
			DominantcolorsCreationFailed:   "Dominantcolors creation failed",
			DominantcolorsProcessingFailed: "Dominantcolors processing failed",
			DominantColorsCalculated:       "Dominant colors calculated",
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

// --- NOTIFICATIONS ---

type NotificationStatus string

const (
	// Queue status codes
	NotificationQueueStarted    NotificationStatus = "notification_queue_started"
	NotificationQueueSubscribed NotificationStatus = "notification_queue_subscribed"
	NotificationQueueFailed     NotificationStatus = "notification_queue_failed"
	NotificationQueueCompleted  NotificationStatus = "notification_queue_completed"

	// Trace status codes
	NotificationTracingStarted   NotificationStatus = "notification_tracing_started"
	NotificationTracingCompleted NotificationStatus = "notification_tracing_completed"
	NotificationTracingFailed    NotificationStatus = "notification_tracing_failed"

	// Metrics status codes
	NotificationMetricsEnabled NotificationStatus = "notification_metrics_enabled"

	// Stage status codes
	NotificationStageStart           NotificationStatus = "notification_stage_start"
	NotificationStageEnd             NotificationStatus = "notification_stage_end"
	NotificationStageMissing         NotificationStatus = "notification_stage_missing"
	NotificationUserNotFound         NotificationStatus = "notification_user_not_found"
	NotificationOrganizationNotFound NotificationStatus = "notification_organization_not_found"
	NotificationMonitorStageMissing  NotificationStatus = "notification_monitor_stage_missing"

	// Internal status codes
	NotificationExpired                  NotificationStatus = "notification_expired"
	NotificationSequenceDecodeFailed     NotificationStatus = "notification_sequence_decode_failed"
	NotificationDecodeFailed             NotificationStatus = "notification_decode_failed"
	NotificationAlreadySent              NotificationStatus = "notification_already_sent"
	NotificationSiteNotFound             NotificationStatus = "notification_site_not_found"
	NotificationSiteDecodeFailed         NotificationStatus = "notification_site_decode_failed"
	NotificationGroupNotFound            NotificationStatus = "notification_group_not_found"
	NotificationGroupDecodeFailed        NotificationStatus = "notification_group_decode_failed"
	NotificationAlertNotFound            NotificationStatus = "notification_alert_not_found"
	NotificationAlertDecodeFailed        NotificationStatus = "notification_alert_decode_failed"
	NotificationFindingCustomAlerts      NotificationStatus = "notification_finding_custom_alerts"
	NotificationProcessingCustomAlert    NotificationStatus = "notification_processing_custom_alert"
	NotificationAlertDisabled            NotificationStatus = "notification_alert_disabled"
	NotificationMediaInSequenceNotFound  NotificationStatus = "notification_media_in_sequence_not_found"
	NotificationSelectedIONotActive      NotificationStatus = "notification_selected_io_not_active"
	NotificationInvalidClassification    NotificationStatus = "notification_invalid_classification"
	NotificationInvalidTimeInterval      NotificationStatus = "notification_invalid_time_interval"
	NotificationNoFrameDimensionsDefined NotificationStatus = "notification_no_frame_dimensions_defined"
	NotificationNotClassifyOperation     NotificationStatus = "notification_not_classify_operation"
	NotificationNoDeviceSelected         NotificationStatus = "notification_no_device_selected"
	NotificationNoRegionMatched          NotificationStatus = "notification_no_region_matched"
	NotificationNoValidCountingFound     NotificationStatus = "notification_no_valid_counting_found"
	NotificationSkippingInvalidCounting  NotificationStatus = "notification_skipping_invalid_counting"
	NotificationSendingNotification      NotificationStatus = "notification_sending_notification"
	NotificationSendingToChannels        NotificationStatus = "notification_sending_to_channels"
	NotificationUpdateSequenceFailed     NotificationStatus = "notification_update_sequence_failed"
	NotificationCustomAlertCompleted     NotificationStatus = "notification_custom_alert_completed"
	NotificationStartingGenericAlerts    NotificationStatus = "notification_starting_generic_alerts"
	NotificationGenericAlertNotEnabled   NotificationStatus = "notification_generic_alert_not_enabled"
	NotificationChannelsToBeTriggered    NotificationStatus = "notification_channels_to_be_triggered"
	NotificationComposedMessage          NotificationStatus = "notification_composed_message"
	NotificationProcessingStart          NotificationStatus = "notification_processing_start"
	NotificationProcessingEnd            NotificationStatus = "notification_processing_end"
	NotificationMarkerCreationFailed     NotificationStatus = "notification_marker_creation_failed"
	NotificationMarkerCreated            NotificationStatus = "notification_marker_created"
	NotificationNoChannelsToBeTriggered  NotificationStatus = "notification_no_channels_to_be_triggered"
	NotificationNoChannelsToSend         NotificationStatus = "notification_no_channels_to_send"
	NotificationSendNotificationFailed   NotificationStatus = "notification_send_notification_failed"
	NotificationCreateMarkerFailed       NotificationStatus = "notification_create_marker_failed"

	NotificationUserNotificationSettingsEmpty NotificationStatus = "notification_user_notification_settings_empty"
	NotificationUserChannelsEmpty             NotificationStatus = "notification_user_channels_empty"
)

// String returns the string representation of the Notification status
func (ms NotificationStatus) String() string {
	return string(ms)
}

// Translate returns the translated string representation of the Notification status in the specified language
func (ms NotificationStatus) Translate(lang string) string {
	translations := map[string]map[NotificationStatus]string{
		"en": {
			NotificationMetricsEnabled:       "Notification metrics enabled",
			NotificationStageStart:           "Starting Notification stage",
			NotificationStageEnd:             "Notification stage completed",
			NotificationStageMissing:         "Notification stage missing",
			NotificationUserNotFound:         "User not found during Notification stage",
			NotificationOrganizationNotFound: "Organization not found during Notification stage",

			NotificationMonitorStageMissing:      "Monitor stage missing during Notification stage",
			NotificationExpired:                  "Notification too old, has expired",
			NotificationSequenceDecodeFailed:     "Failed to decode notification sequence",
			NotificationDecodeFailed:             "Failed to decode notification",
			NotificationAlreadySent:              "Notification has already been sent",
			NotificationSiteNotFound:             "Site not found during Notification stage",
			NotificationSiteDecodeFailed:         "Failed to decode site information",
			NotificationGroupNotFound:            "Group not found during Notification stage",
			NotificationGroupDecodeFailed:        "Failed to decode group information",
			NotificationAlertNotFound:            "Alert not found during Notification stage",
			NotificationAlertDecodeFailed:        "Failed to decode alert information",
			NotificationFindingCustomAlerts:      "Finding custom alerts for notification",
			NotificationProcessingCustomAlert:    "Processing custom alert for notification",
			NotificationAlertDisabled:            "Alert is disabled, notification not sent",
			NotificationMediaInSequenceNotFound:  "Media in sequence not found during Notification stage",
			NotificationSelectedIONotActive:      "Selected IO is not active during Notification stage",
			NotificationInvalidClassification:    "Invalid classification for notification",
			NotificationInvalidTimeInterval:      "Invalid time interval for notification",
			NotificationNoFrameDimensionsDefined: "No frameWidth/frameHeight defined make sure the Kerberos Hub detector is running the right version.",
			NotificationNotClassifyOperation:     "Operation is not a classify operation for notification",
			NotificationNoDeviceSelected:         "No device selected or device was not in the selection.",
			NotificationNoRegionMatched:          "No region matched for notification",
			NotificationNoValidCountingFound:     "No valid counting found for notification",
			NotificationSkippingInvalidCounting:  "Skipping invalid counting for notification",
			NotificationSendingNotification:      "Sending notification",
			NotificationSendingToChannels:        "Sending notification to channels",
			NotificationUpdateSequenceFailed:     "Failed to update notification sequence",
			NotificationCustomAlertCompleted:     "Custom alert notification completed",
			NotificationStartingGenericAlerts:    "Starting generic alerts for notification",
			NotificationGenericAlertNotEnabled:   "Generic alert not enabled for notification",
			NotificationChannelsToBeTriggered:    "Channels to be triggered for notification",
			NotificationComposedMessage:          "Composed notification message",
			NotificationProcessingStart:          "Starting processing for notification",
			NotificationProcessingEnd:            "Notification processing completed",
			NotificationMarkerCreationFailed:     "Failed to create marker during Notification stage",
			NotificationMarkerCreated:            "Marker created during Notification stage",
			NotificationNoChannelsToBeTriggered:  "No channels to be triggered for notification",
			NotificationNoChannelsToSend:         "No channels to send notification to",
			NotificationSendNotificationFailed:   "Failed to send notification",
			NotificationCreateMarkerFailed:       "Failed to create marker for notification",

			NotificationUserNotificationSettingsEmpty: "User notification settings are empty",
			NotificationUserChannelsEmpty:             "User channels are empty",
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
