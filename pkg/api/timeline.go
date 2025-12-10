package api

import "github.com/uug-ai/models/pkg/models"

// MediaStatus represents specific status codes for media operations
type TimelineStatus string

const (
	TimelineBindingFailed TimelineStatus = "Timeline_binding_failed"
	TimelineDuplicateName TimelineStatus = "Timeline_duplicate_name"
	TimelineMissingInfo   TimelineStatus = "Timeline_missing_info"
	TimelineFound         TimelineStatus = "Timeline_found"
	TimelineNotFound      TimelineStatus = "Timeline_not_found"
	TimelineAddSuccess    TimelineStatus = "Timeline_add_success"
	TimelineAddFailed     TimelineStatus = "Timeline_add_failed"
	TimelineUpdateSuccess TimelineStatus = "Timeline_update_success"
	TimelineUpdateFailed  TimelineStatus = "Timeline_update_failed"
	TimelineDeleteSuccess TimelineStatus = "Timeline_delete_success"
	TimelineDeleteFailed  TimelineStatus = "Timeline_delete_failed"
)

// String returns the string representation of the Timeline status
func (ms TimelineStatus) String() string {
	return string(ms)
}

// Into returns the translated string representation of the Timeline status in the specified language
func (ms TimelineStatus) Translate(lang string) string {
	translations := map[string]map[TimelineStatus]string{
		"en": {
			TimelineBindingFailed: "Timeline binding failed",
			TimelineDuplicateName: "Timeline duplicate name",
			TimelineMissingInfo:   "Timeline missing information",
			TimelineFound:         "Timeline found",
			TimelineNotFound:      "Timeline not found",
			TimelineAddSuccess:    "Timeline added successfully",
			TimelineAddFailed:     "Timeline failed to add",
			TimelineUpdateSuccess: "Timeline updated successfully",
			TimelineUpdateFailed:  "Timeline failed to update",
			TimelineDeleteSuccess: "Timeline deleted successfully",
			TimelineDeleteFailed:  "Timeline failed to delete",
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

// GetTimelineMedia
// @Router /timeline/{deviceId} [post]
type GetTimelineMediaRequest struct {
	Filter MediaFilter `json:"filter" bson:"filter"`
}
type GetTimelineMediaResponse struct {
	Device models.Device `json:"device"`
	Media  []MediaGroup  `json:"media"`
}
type GetTimelineMediaErrorResponse struct {
	ErrorResponse
}
type GetTimelineMediaSuccessResponse struct {
	SuccessResponse
	Data GetTimelineMediaResponse `json:"data"`
}

// GetTimelineMarkers
// @Router /timeline/{deviceId}/markers [post]
type GetTimelineMarkersRequest struct {
	Filter MarkerFilter `json:"filter" bson:"filter"`
}

type GetTimelineMarkersResponse struct {
	Device  models.Device                  `json:"device"`
	Markers []models.MarkerOptionTimeRange `json:"markers"`
}

type GetTimelineMarkersSuccessResponse struct {
	SuccessResponse
	Data GetTimelineMarkersResponse `json:"data"`
}

type GetTimelineMarkersErrorResponse struct {
	ErrorResponse
}

// GetTimelineEvents
// @Router /timeline/{deviceId}/events [post]
type GetTimelineEventsRequest struct {
	Filter MarkerEventFilter `json:"filter" bson:"filter"`
}

type GetTimelineEventsResponse struct {
	Device models.Device                 `json:"device"`
	Events []models.MarkerEventTimeRange `json:"events"`
}

type GetTimelineEventsSuccessResponse struct {
	SuccessResponse
	Data GetTimelineEventsResponse `json:"data"`
}

type GetTimelineEventsErrorResponse struct {
	ErrorResponse
}
