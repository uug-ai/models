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

// GetMediaByDayAndDeviceFiltered
// @Router /media/day/{day}/device/{deviceId} [post]
type GetMediaByDayAndDeviceFilteredRequest struct {
}
type GetMediaByDayAndDeviceFilteredResponse struct {
}
type GetMediaByDayAndDeviceFilteredErrorResponse struct {
	ErrorResponse
}
type GetMediaByDayAndDeviceFilteredSuccessResponse struct {
	SuccessResponse
	Data GetMediaByDayAndDeviceFilteredResponse `json:"data"`
}

// GetMediaByDayAndDevice
// @Router /media/day/{day}/device/{deviceId} [get]
type GetMediaByDayAndDeviceRequest struct {
}

type GetMediaByDayAndDeviceResponse struct {
	Device models.Device  `json:"device"`
	Media  []models.Media `json:"media"`
}

type GetMediaByDayAndDeviceSuccessResponse struct {
	SuccessResponse
	Data GetMediaByDayAndDeviceResponse `json:"data"`
}

type GetMediaByDayAndDeviceErrorResponse struct {
	ErrorResponse
}
