package api

import "github.com/uug-ai/models/pkg/models"

// MediaStatus represents specific status codes for media operations
type MediaStatus string

const (
	MediaBindingFailed MediaStatus = "media_binding_failed"
	MediaDuplicateName MediaStatus = "media_duplicate_name"
	MediaMissingInfo   MediaStatus = "media_missing_info"
	MediaFound         MediaStatus = "media_found"
	MediaNotFound      MediaStatus = "media_not_found"
	MediaAddSuccess    MediaStatus = "media_add_success"
	MediaAddFailed     MediaStatus = "media_add_failed"
	MediaUpdateSuccess MediaStatus = "media_update_success"
	MediaUpdateFailed  MediaStatus = "media_update_failed"
	MediaDeleteSuccess MediaStatus = "media_delete_success"
	MediaDeleteFailed  MediaStatus = "media_delete_failed"
)

// String returns the string representation of the media status
func (ms MediaStatus) String() string {
	return string(ms)
}

// Into returns the translated string representation of the media status in the specified language
func (ms MediaStatus) Translate(lang string) string {
	translations := map[string]map[MediaStatus]string{
		"en": {
			MediaBindingFailed: "Media binding failed",
			MediaDuplicateName: "Media duplicate name",
			MediaMissingInfo:   "Media missing information",
			MediaFound:         "Media found",
			MediaNotFound:      "Media not found",
			MediaAddSuccess:    "Media added successfully",
			MediaAddFailed:     "Media failed to add",
			MediaUpdateSuccess: "Media updated successfully",
			MediaUpdateFailed:  "Media failed to update",
			MediaDeleteSuccess: "Media deleted successfully",
			MediaDeleteFailed:  "Media failed to delete",
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

type MediaFilter struct {
	LastMedia              int64            `json:"lastMedia" bson:"lastMedia"`
	GlobalSearch           bool             `json:"globalSearch" bson:"globalSearch"`
	Dates                  []string         `json:"dates" bson:"dates"`
	Devices                []string         `json:"devices" bson:"devices"`
	Regions                []models.Region  `json:"regions" bson:"regions"`
	Classifications        []string         `json:"classifications" bson:"classifications"`
	Sort                   string           `json:"sort" bson:"sort"`
	Favourite              bool             `json:"favourite" bson:"favourite"`
	HasLabel               bool             `json:"hasLabel" bson:"hasLabel"`
	HourRange              models.HourRange `json:"hourRange" bson:"hourRange"`
	Markers                []string         `json:"markers" bson:"markers"`
	Events                 []string         `json:"events" bson:"events"`
	ViewStyle              string           `json:"viewStyle" bson:"viewStyle"`
	Offset                 int64            `json:"offset" bson:"offset"`
	Limit                  int64            `json:"limit" bson:"limit"`
	TimelineStartTimestamp int64            `json:"timelineStartTimestamp" bson:"timelineStartTimestamp"`
	TimelineEndTimestamp   int64            `json:"timelineEndTimestamp" bson:"timelineEndTimestamp"`
}

type MediaGroup struct {
	StartTimestamp int64          `json:"startTimestamp" bson:"startTimestamp"`
	EndTimestamp   int64          `json:"endTimestamp" bson:"endTimestamp"`
	Count          int64          `json:"count" bson:"count"`
	Media          []models.Media `json:"media" bson:"media"`
}

// GetTimeline
// @Router /timeline/{deviceId} [post]
type GetTimelineRequest struct {
	Filter MediaFilter `json:"filter" bson:"filter"`
}
type GetTimelineResponse struct {
	Device models.Device `json:"device"`
	Media  []MediaGroup  `json:"media"`
}
type GetTimelineErrorResponse struct {
	ErrorResponse
}
type GetTimelineSuccessResponse struct {
	SuccessResponse
	Data GetTimelineResponse `json:"data"`
}

// GetMedia
// @Router /media/ [post]
type GetMediaRequest struct {
	Filter MediaFilter `json:"filter" bson:"filter"`
}
type GetMediaResponse struct {
	Media []models.Media `json:"media"`
}
type GetMediaSuccessResponse struct {
	SuccessResponse
	Data GetMediaResponse `json:"data"`
}
type GetMediaErrorResponse struct {
	ErrorResponse
}
