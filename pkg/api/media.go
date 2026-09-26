package api

import "github.com/uug-ai/models/pkg/models"

// MediaStatus represents specific status codes for media operations
type MediaStatus string

const (
	MediaBindingFailed               MediaStatus = "media_binding_failed"
	MediaDuplicateName               MediaStatus = "media_duplicate_name"
	MediaMissingInfo                 MediaStatus = "media_missing_info"
	MediaFound                       MediaStatus = "media_found"
	MediaNotFound                    MediaStatus = "media_not_found"
	MediaAddSuccess                  MediaStatus = "media_add_success"
	MediaAddFailed                   MediaStatus = "media_add_failed"
	MediaUpdateSuccess               MediaStatus = "media_update_success"
	MediaUpdateFailed                MediaStatus = "media_update_failed"
	MediaDeleteSuccess               MediaStatus = "media_delete_success"
	MediaDeleteFailed                MediaStatus = "media_delete_failed"
	MediaIdMissing                   MediaStatus = "media_id_missing"
	MediaDownloadFailed              MediaStatus = "media_download_failed"
	MediaDownloadSuccess             MediaStatus = "media_download_success"
	MediaUploadFailed                MediaStatus = "media_upload_failed"
	MediaUploadSuccess               MediaStatus = "media_upload_success"
	MediaPublishFailed               MediaStatus = "media_publish_failed"
	MediaPublishSuccess              MediaStatus = "media_publish_success"
	MediaCleanupFailed               MediaStatus = "media_cleanup_failed"
	MediaVideoDurationExtracted      MediaStatus = "media_video_duration_extracted"
	MediaThumbnailLoaded             MediaStatus = "media_thumbnail_loaded"
	MediaFileCouldNotExtractUsername MediaStatus = "media_file_could_not_extract_username"
)

// String returns the string representation of the media status
func (ms MediaStatus) String() string {
	return string(ms)
}

// Translate returns the translated string representation of the media status in the specified language
func (ms MediaStatus) Translate(lang string) string {
	translations := map[string]map[MediaStatus]string{
		"en": {
			MediaBindingFailed:          "Media binding failed",
			MediaDuplicateName:          "Media duplicate name",
			MediaMissingInfo:            "Media missing information",
			MediaFound:                  "Media found",
			MediaNotFound:               "Media not found",
			MediaAddSuccess:             "Media added successfully",
			MediaAddFailed:              "Media failed to add",
			MediaUpdateSuccess:          "Media updated successfully",
			MediaUpdateFailed:           "Media failed to update",
			MediaDeleteSuccess:          "Media deleted successfully",
			MediaDeleteFailed:           "Media failed to delete",
			MediaIdMissing:              "Media ID is missing",
			MediaDownloadFailed:         "Media download failed",
			MediaDownloadSuccess:        "Media downloaded successfully",
			MediaUploadFailed:           "Media upload failed",
			MediaUploadSuccess:          "Media uploaded successfully",
			MediaPublishFailed:          "Media publish failed",
			MediaPublishSuccess:         "Media published successfully",
			MediaCleanupFailed:          "Media cleanup failed",
			MediaVideoDurationExtracted: "Media video duration extracted",
			MediaThumbnailLoaded:        "Media thumbnail loaded",
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
	TimeRanges      []*models.TimeRange `json:"timeRanges,omitempty" bson:"timeRanges,omitempty"`
	MediaIds        []*string           `json:"mediaIds,omitempty" bson:"mediaIds,omitempty"`
	Sites           []*string           `json:"sites,omitempty" bson:"sites,omitempty"`
	Groups          []*string           `json:"groups,omitempty" bson:"groups,omitempty"`
	Devices         []*string           `json:"devices,omitempty" bson:"devices,omitempty"`
	ExcludedDevices []*string           `json:"excludedDevices,omitempty" bson:"excludedDevices,omitempty"`
	ExcludedMedia   []*string           `json:"excludedMedia,omitempty" bson:"excludedMedia,omitempty"`
	Markers         []*string           `json:"markers,omitempty" bson:"markers,omitempty"`
	Events          []*string           `json:"events,omitempty" bson:"events,omitempty"`
	Tags            []*string           `json:"tags,omitempty" bson:"tags,omitempty"`
	Regions         []*models.Region    `json:"regions,omitempty" bson:"regions,omitempty"`
	Starred         *bool               `json:"starred,omitempty" bson:"starred,omitempty"`
	SortBy          *string             `json:"sortBy,omitempty" bson:"sortBy,omitempty"`
}

type MediaPatch struct {
	Metadata *MediaMetadataPatch `json:"metadata,omitempty" bson:"metadata,omitempty"`
	Star     *bool               `json:"star,omitempty" bson:"star,omitempty"`
}

type MediaMetadataPatch struct {
	Description *string `json:"description,omitempty" bson:"description,omitempty"`
}

// GetMedia
// @Router /media/ [post]
type GetMediaRequest struct {
	Filter     MediaFilter      `json:"filter" bson:"filter"`
	Pagination CursorPagination `json:"pagination" bson:"pagination"`
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

// CountMedia
// @Router /media/count [post]
type CountMediaRequest struct {
	Filter MediaFilter `json:"filter" bson:"filter"`
	Limit  int64       `json:"limit" bson:"limit"`
}
type CountMediaResponse struct {
	models.MediaCount
}
type CountMediaSuccessResponse struct {
	SuccessResponse
	Data CountMediaResponse `json:"data"`
}
type CountMediaErrorResponse struct {
	ErrorResponse
}

// GetMediaById
// @Router /media/{mediaId} [get]
type GetMediaByIdRequest struct {
	MediaId string `json:"mediaId" bson:"mediaId"`
}
type GetMediaByIdResponse struct {
	Media models.Media `json:"media"`
}
type GetMediaByIdSuccessResponse struct {
	SuccessResponse
	Data GetMediaByIdResponse `json:"data"`
}
type GetMediaByIdErrorResponse struct {
	ErrorResponse
}

// GetMediaByVideoFile
// @Router /media/video-file [get]
type GetMediaByVideoFileRequest struct {
}
type GetMediaByVideoFileResponse struct {
	Media models.Media `json:"media"`
}
type GetMediaByVideoFileSuccessResponse struct {
	SuccessResponse
	Data GetMediaByVideoFileResponse `json:"data"`
}
type GetMediaByVideoFileErrorResponse struct {
	ErrorResponse
}

// UpdateMedia
// @Router /media/{mediaId} [patch]
type UpdateMediaRequest struct {
	MediaPatch MediaPatch `json:"mediaPatch" bson:"mediaPatch"`
}
type UpdateMediaResponse struct {
	Media models.Media `json:"media"`
}
type UpdateMediaSuccessResponse struct {
	SuccessResponse
	Data UpdateMediaResponse `json:"data"`
}
type UpdateMediaErrorResponse struct {
	ErrorResponse
}

// GetMediaVLMContext
// @Router /media/{mediaId}/vlm [get]
type GetMediaVLMContextRequest struct {
	MediaId string `json:"mediaId" bson:"mediaId"`
}
type GetMediaVLMContextResponse struct {
	// Media is the VLM analysis of the recording; nil when the recording has not been analysed yet.
	Media *models.VLMMediaMetadata `json:"media,omitempty"`
	// StableStates are the stable scenes of the recording's device.
	StableStates []models.VLMStableState `json:"stableStates"`
}
type GetMediaVLMContextSuccessResponse struct {
	SuccessResponse
	Data GetMediaVLMContextResponse `json:"data"`
}
type GetMediaVLMContextErrorResponse struct {
	ErrorResponse
}

// PromoteMediaVLMStableState copies the VLM scene of a recording to its device as a named stable state.
// @Router /media/{mediaId}/vlm-baseline [post]
type PromoteMediaVLMStableStateRequest struct {
	// Name labels the stable state, e.g. "Daytime, gate closed". Defaults to the scene summary.
	Name string `json:"name,omitempty" bson:"name,omitempty"`
	// Scene optionally narrows the recording's scene before it is saved. It may only
	// drop the summary, static elements, zones or lighting; it cannot add values.
	Scene *models.VLMScene `json:"scene,omitempty" bson:"scene,omitempty"`
}
type PromoteMediaVLMStableStateResponse struct {
	StableState models.VLMStableState `json:"stableState"`
}
type PromoteMediaVLMStableStateSuccessResponse struct {
	SuccessResponse
	Data PromoteMediaVLMStableStateResponse `json:"data"`
}
type PromoteMediaVLMStableStateErrorResponse struct {
	ErrorResponse
}

// UpdateMediaVLMStableState renames or narrows a stable state of the recording's device.
// @Router /media/{mediaId}/vlm-baseline/{stableStateId} [patch]
type UpdateMediaVLMStableStateRequest struct {
	// Name relabels the stable state; an empty name keeps the current one.
	Name string `json:"name,omitempty" bson:"name,omitempty"`
	// Scene optionally narrows the stored scene. It may only drop the summary,
	// static elements, zones or lighting; it cannot add values.
	Scene *models.VLMScene `json:"scene,omitempty" bson:"scene,omitempty"`
}
type UpdateMediaVLMStableStateResponse struct {
	StableState models.VLMStableState `json:"stableState"`
}
type UpdateMediaVLMStableStateSuccessResponse struct {
	SuccessResponse
	Data UpdateMediaVLMStableStateResponse `json:"data"`
}
type UpdateMediaVLMStableStateErrorResponse struct {
	ErrorResponse
}

// DeleteMediaVLMStableState removes a stable state from the recording's device.
// @Router /media/{mediaId}/vlm-baseline/{stableStateId} [delete]
type DeleteMediaVLMStableStateRequest struct {
	MediaId       string `json:"mediaId" bson:"mediaId"`
	StableStateId string `json:"stableStateId" bson:"stableStateId"`
}
type DeleteMediaVLMStableStateResponse struct {
	StableStateId string `json:"stableStateId"`
}
type DeleteMediaVLMStableStateSuccessResponse struct {
	SuccessResponse
	Data DeleteMediaVLMStableStateResponse `json:"data"`
}
type DeleteMediaVLMStableStateErrorResponse struct {
	ErrorResponse
}
