package api

import "github.com/uug-ai/models/pkg/models"

// AlertStatus represents specific status codes for alert operations
type AlertStatus string

const (
	AlertBindingFailed    AlertStatus = "alert_binding_failed"
	AlertDuplicateName    AlertStatus = "alert_duplicate_name"
	AlertMissingInfo      AlertStatus = "alert_missing_info"
	AlertRetrievalSuccess AlertStatus = "alert_retrieval_success"
	AlertRetrievalFailed  AlertStatus = "alert_retrieval_failed"
	AlertFound            AlertStatus = "alert_found"
	AlertNotFound         AlertStatus = "alert_not_found"
	AlertAddSuccess       AlertStatus = "alert_add_success"
	AlertAddFailed        AlertStatus = "alert_add_failed"
	AlertUpdateSuccess    AlertStatus = "alert_update_success"
	AlertUpdateFailed     AlertStatus = "alert_update_failed"
	AlertDeleteSuccess    AlertStatus = "alert_delete_success"
	AlertDeleteFailed     AlertStatus = "alert_delete_failed"
	AlertValidationFailed AlertStatus = "alert_validation_failed"
)

// String returns the string representation of the alert status
func (as AlertStatus) String() string {
	return string(as)
}

// Translate returns the translated string representation of the alert status in the specified language
func (as AlertStatus) Translate(lang string) string {
	translations := map[string]map[AlertStatus]string{
		"en": {
			AlertBindingFailed:    "Alert binding failed",
			AlertDuplicateName:    "Alert duplicate name",
			AlertMissingInfo:      "Alert missing information",
			AlertRetrievalSuccess: "Alert retrieved successfully",
			AlertRetrievalFailed:  "Alert retrieval failed",
			AlertFound:            "Alert found",
			AlertNotFound:         "Alert not found",
			AlertAddSuccess:       "Alert added successfully",
			AlertAddFailed:        "Alert failed to add",
			AlertUpdateSuccess:    "Alert updated successfully",
			AlertUpdateFailed:     "Alert failed to update",
			AlertDeleteSuccess:    "Alert deleted successfully",
			AlertDeleteFailed:     "Alert failed to delete",
			AlertValidationFailed: "Alert validation failed",
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

type CustomAlertPatch struct {
	Title               *string                  `json:"title,omitempty" bson:"title,omitempty"`
	Enabled             *bool                    `json:"enabled,omitempty" bson:"enabled,omitempty"`
	Description         *string                  `json:"description,omitempty" bson:"description,omitempty"`
	TimeRange1Min       *int32                   `json:"timeRange1Min,omitempty" bson:"timeRange1Min,omitempty"`
	TimeRange1Max       *int32                   `json:"timeRange1Max,omitempty" bson:"timeRange1Max,omitempty"`
	TimeRange2Min       *int32                   `json:"timeRange2Min,omitempty" bson:"timeRange2Min,omitempty"`
	TimeRange2Max       *int32                   `json:"timeRange2Max,omitempty" bson:"timeRange2Max,omitempty"`
	TimeAdvanced        *bool                    `json:"timeAdvanced,omitempty" bson:"timeAdvanced,omitempty"`
	WeeklySchedule      []*models.WeeklySchedule `json:"weeklySchedule,omitempty" bson:"weeklySchedule,omitempty"`
	ChannelsAll         *bool                    `json:"channelsAll,omitempty" bson:"channelsAll,omitempty"`
	ChannelsList        []string                 `json:"channelsList,omitempty" bson:"channelsList,omitempty"`
	DevicesAll          *bool                    `json:"devicesAll,omitempty" bson:"devicesAll,omitempty"`
	DevicesList         []models.DeviceKey       `json:"devicesList,omitempty" bson:"devicesList,omitempty"`
	CountingDevicesAll  *bool                    `json:"countingDevicesAll,omitempty" bson:"countingDevicesAll,omitempty"`
	CountingDevicesList []models.DeviceKey       `json:"countingDevicesList,omitempty" bson:"countingDevicesList,omitempty"`
	ClassificationAll   *bool                    `json:"classificationAll,omitempty" bson:"classificationAll,omitempty"`
	ClassificationList  []string                 `json:"classificationList,omitempty" bson:"classificationList,omitempty"`
	MotionRegions       []models.Region          `json:"motionRegions,omitempty" bson:"motionRegions,omitempty"`
	CountingRegions     []models.Region          `json:"countingRegions,omitempty" bson:"countingRegions,omitempty"`
	CountingLines       []models.Region          `json:"countingLines,omitempty" bson:"countingLines,omitempty"`
	InputList           []string                 `json:"inputList,omitempty" bson:"inputList,omitempty"`
	OutputList          []string                 `json:"outputList,omitempty" bson:"outputList,omitempty"`
	InputsAND           *bool                    `json:"inputsAND,omitempty" bson:"inputsAND,omitempty"`
	EmailEmail          *string                  `json:"email_email,omitempty" bson:"email_email,omitempty"`
	SlackHook           *string                  `json:"slack_hook,omitempty" bson:"slack_hook,omitempty"`
	SlackBotname        *string                  `json:"slack_botname,omitempty" bson:"slack_botname,omitempty"`
	PushbulletApikey    *string                  `json:"pushbullet_apikey,omitempty" bson:"pushbullet_apikey,omitempty"`
	TelegramToken       *string                  `json:"telegram_token,omitempty" bson:"telegram_token,omitempty"`
	TelegramChannel     *string                  `json:"telegram_channel,omitempty" bson:"telegram_channel,omitempty"`
	AlexaToken          *string                  `json:"alexa_token,omitempty" bson:"alexa_token,omitempty"`
	WebhookUrl          *string                  `json:"webhook_url,omitempty" bson:"webhook_url,omitempty"`
	IftttToken          *string                  `json:"ifttt_token,omitempty" bson:"ifttt_token,omitempty"`
	SMSAccountsid       *string                  `json:"sms_accountsid,omitempty" bson:"sms_accountsid,omitempty"`
	SMSAuthtoken        *string                  `json:"sms_authtoken,omitempty" bson:"sms_authtoken,omitempty"`
	SMSTelfrom          *string                  `json:"sms_telfrom,omitempty" bson:"sms_telfrom,omitempty"`
	SMSTelto            *string                  `json:"sms_telto,omitempty" bson:"sms_telto,omitempty"`
	PushoverApikey      *string                  `json:"pushover_apikey,omitempty" bson:"pushover_apikey,omitempty"`
	PushoverSendto      *string                  `json:"pushover_sendto,omitempty" bson:"pushover_sendto,omitempty"`
}

// GetCustomAlerts
type GetCustomAlertsRequest struct {
}
type GetCustomAlertsResponse struct {
	Alerts []models.CustomAlert `json:"alerts"`
}
type GetCustomAlertsSuccessResponse struct {
	SuccessResponse
	Data GetCustomAlertsResponse `json:"data"`
}
type GetCustomAlertsErrorResponse struct {
	ErrorResponse
}

// CreateCustomAlert
type CreateCustomAlertRequest struct {
	Alert models.CustomAlert `json:"alert"`
}
type CreateCustomAlertResponse struct {
	Alert models.CustomAlert `json:"alert"`
}
type CreateCustomAlertSuccessResponse struct {
	SuccessResponse
	Data CreateCustomAlertResponse `json:"data"`
}
type CreateCustomAlertErrorResponse struct {
	ErrorResponse
}

// UpdateCustomAlert
type UpdateCustomAlertRequest struct {
	AlertPatch models.AlertPatch `json:"alertPatch"`
}
type UpdateCustomAlertResponse struct {
	Alert models.CustomAlert `json:"alert"`
}
type UpdateCustomAlertSuccessResponse struct {
	SuccessResponse
	Data UpdateCustomAlertResponse `json:"data"`
}
type UpdateCustomAlertErrorResponse struct {
	ErrorResponse
}

// RemoveCustomAlert
type RemoveCustomAlertRequest struct {
}
type RemoveCustomAlertResponse struct {
}
type RemoveCustomAlertSuccessResponse struct {
	SuccessResponse
	Data RemoveCustomAlertResponse `json:"data"`
}
type RemoveCustomAlertErrorResponse struct {
	ErrorResponse
}
