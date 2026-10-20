package api

import "github.com/uug-ai/models/pkg/models"

// ChannelStatus represents specific status codes for channel operations.
type ChannelStatus string

const (
	ChannelBindingFailed    ChannelStatus = "channel_binding_failed"
	ChannelMissingInfo      ChannelStatus = "channel_missing_info"
	ChannelInvalidConfig    ChannelStatus = "channel_invalid_config"
	ChannelFound            ChannelStatus = "channel_found"
	ChannelNotFound         ChannelStatus = "channel_not_found"
	ChannelRetrievalSuccess ChannelStatus = "channel_retrieval_success"
	ChannelRetrievalFailed  ChannelStatus = "channel_retrieval_failed"
	ChannelAddSuccess       ChannelStatus = "channel_add_success"
	ChannelAddFailed        ChannelStatus = "channel_add_failed"
	ChannelUpdateSuccess    ChannelStatus = "channel_update_success"
	ChannelUpdateFailed     ChannelStatus = "channel_update_failed"
	ChannelDeleteSuccess    ChannelStatus = "channel_delete_success"
	ChannelDeleteFailed     ChannelStatus = "channel_delete_failed"
	ChannelTestSuccess      ChannelStatus = "channel_test_success"
	ChannelTestFailed       ChannelStatus = "channel_test_failed"
	ChannelForbidden        ChannelStatus = "channel_forbidden"
)

// String returns the string representation of the channel status.
func (cs ChannelStatus) String() string {
	return string(cs)
}

// Translate returns the translated string representation of the channel status in the specified language.
func (cs ChannelStatus) Translate(lang string) string {
	translations := map[string]map[ChannelStatus]string{
		"en": {
			ChannelBindingFailed:    "Channel binding failed",
			ChannelMissingInfo:      "Channel missing information",
			ChannelInvalidConfig:    "Channel settings are invalid or do not match the channel type",
			ChannelFound:            "Channel found",
			ChannelNotFound:         "Channel not found",
			ChannelRetrievalSuccess: "Channel retrieved successfully",
			ChannelRetrievalFailed:  "Channel retrieval failed",
			ChannelAddSuccess:       "Channel added successfully",
			ChannelAddFailed:        "Channel failed to add",
			ChannelUpdateSuccess:    "Channel updated successfully",
			ChannelUpdateFailed:     "Channel failed to update",
			ChannelDeleteSuccess:    "Channel deleted successfully",
			ChannelDeleteFailed:     "Channel failed to delete",
			ChannelTestSuccess:      "Test notification sent successfully",
			ChannelTestFailed:       "Test notification failed to send",
			ChannelForbidden:        "You do not have permission for this action",
		},
	}

	if langTranslations, exists := translations[lang]; exists {
		if translation, exists := langTranslations[cs]; exists {
			return translation
		}
	}

	if enTranslations, exists := translations["en"]; exists {
		if translation, exists := enTranslations[cs]; exists {
			return translation
		}
	}

	return cs.String()
}

// Channel responses carry models.Channel, whose Config never serialises, so no
// read path returns recipient addresses or credentials. Config is only ever
// sent by the client, on create and update.

// GetChannels
type GetChannelsRequest struct{}
type GetChannelsResponse struct {
	Channels []models.Channel `json:"channels"`
}
type GetChannelsSuccessResponse struct {
	SuccessResponse
	Data GetChannelsResponse `json:"data"`
}
type GetChannelsErrorResponse struct {
	ErrorResponse
}

// GetChannel
type GetChannelRequest struct{}
type GetChannelResponse struct {
	Channel models.Channel `json:"channel"`
}
type GetChannelSuccessResponse struct {
	SuccessResponse
	Data GetChannelResponse `json:"data"`
}
type GetChannelErrorResponse struct {
	ErrorResponse
}

// CreateChannel
type CreateChannelRequest struct {
	Channel models.Channel       `json:"channel"`
	Config  models.ChannelConfig `json:"config"`
}
type CreateChannelResponse struct {
	Channel models.Channel `json:"channel"`
}
type CreateChannelSuccessResponse struct {
	SuccessResponse
	Data CreateChannelResponse `json:"data"`
}
type CreateChannelErrorResponse struct {
	ErrorResponse
}

// UpdateChannel
type UpdateChannelRequest struct {
	Channel models.Channel `json:"channel"`
	// Config replaces the stored config when set; nil keeps it unchanged.
	Config *models.ChannelConfig `json:"config,omitempty"`
}
type UpdateChannelResponse struct {
	Channel models.Channel `json:"channel"`
}
type UpdateChannelSuccessResponse struct {
	SuccessResponse
	Data UpdateChannelResponse `json:"data"`
}
type UpdateChannelErrorResponse struct {
	ErrorResponse
}

// DeleteChannel
type DeleteChannelRequest struct{}
type DeleteChannelResponse struct{}
type DeleteChannelSuccessResponse struct {
	SuccessResponse
	Data DeleteChannelResponse `json:"data"`
}
type DeleteChannelErrorResponse struct {
	ErrorResponse
}

// TestChannel sends a test notification through a stored channel and records
// the outcome on Channel.Valid.
type TestChannelRequest struct{}
type TestChannelResponse struct {
	Channel models.Channel `json:"channel"`
}
type TestChannelSuccessResponse struct {
	SuccessResponse
	Data TestChannelResponse `json:"data"`
}
type TestChannelErrorResponse struct {
	ErrorResponse
}
