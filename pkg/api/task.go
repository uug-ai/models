package api

import (
	"github.com/uug-ai/models/pkg/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TaskStatus represents specific status codes for task operations
type TaskStatus string

const (
	TaskBindingFailed   TaskStatus = "Task_binding_failed"
	TaskDuplicateName   TaskStatus = "Task_duplicate_name"
	TaskMissingInfo     TaskStatus = "Task_missing_info"
	TaskFound           TaskStatus = "Task_found"
	TaskNotFound        TaskStatus = "Task_not_found"
	TaskForbidden       TaskStatus = "Task_forbidden"
	TaskAddSuccess      TaskStatus = "Task_add_success"
	TaskAddFailed       TaskStatus = "Task_add_failed"
	TaskUpdateSuccess   TaskStatus = "Task_update_success"
	TaskUpdateFailed    TaskStatus = "Task_update_failed"
	TaskDeleteSuccess   TaskStatus = "Task_delete_success"
	TaskDeleteFailed    TaskStatus = "Task_delete_failed"
	TaskMediaAddSuccess TaskStatus = "Task_media_add_success"
	TaskMediaAddFailed  TaskStatus = "Task_media_add_failed"

	// Export generation trigger — set when the user explicitly
	// requests an export bundle build via POST /tasks/:id/export.
	TaskExportRequestSuccess TaskStatus = "Task_export_request_success"
	TaskExportRequestFailed  TaskStatus = "Task_export_request_failed"
	TaskExportAlreadyActive  TaskStatus = "Task_export_already_active"

	// Case attachments — auxiliary, non-pipeline files attached to a
	// case (PDFs, images, scanned documents, audio notes). See
	// models.CaseAttachment.
	TaskAttachmentAddSuccess    TaskStatus = "Task_attachment_add_success"
	TaskAttachmentAddFailed     TaskStatus = "Task_attachment_add_failed"
	TaskAttachmentUpdateSuccess TaskStatus = "Task_attachment_update_success"
	TaskAttachmentUpdateFailed  TaskStatus = "Task_attachment_update_failed"
	TaskAttachmentDeleteSuccess TaskStatus = "Task_attachment_delete_success"
	TaskAttachmentDeleteFailed  TaskStatus = "Task_attachment_delete_failed"
	TaskAttachmentNotFound      TaskStatus = "Task_attachment_not_found"
	TaskAttachmentTooLarge      TaskStatus = "Task_attachment_too_large"
	TaskAttachmentTypeRejected  TaskStatus = "Task_attachment_type_rejected"
)

// String returns the string representation of the Task status
func (ms TaskStatus) String() string {
	return string(ms)
}

// Into returns the translated string representation of the Task status in the specified language
func (ms TaskStatus) Translate(lang string) string {
	translations := map[string]map[TaskStatus]string{
		"en": {
			TaskBindingFailed:   "Task binding failed",
			TaskDuplicateName:   "Task duplicate name",
			TaskMissingInfo:     "Task missing information",
			TaskFound:           "Task found",
			TaskNotFound:        "Task not found",
			TaskForbidden:       "You are not allowed to access this task",
			TaskAddSuccess:      "Task added successfully",
			TaskAddFailed:       "Task failed to add",
			TaskUpdateSuccess:   "Task updated successfully",
			TaskUpdateFailed:    "Task failed to update",
			TaskDeleteSuccess:   "Task deleted successfully",
			TaskDeleteFailed:    "Task failed to delete",
			TaskMediaAddSuccess: "Media was added to the task successfully",
			TaskMediaAddFailed:  "Failed to add media to the task",

			TaskExportRequestSuccess: "Export bundle generation has been queued",
			TaskExportRequestFailed:  "Failed to queue export bundle generation",
			TaskExportAlreadyActive:  "An export is already in progress for this task",

			TaskAttachmentAddSuccess:    "Attachment added to the case successfully",
			TaskAttachmentAddFailed:     "Failed to add attachment to the case",
			TaskAttachmentUpdateSuccess: "Attachment updated successfully",
			TaskAttachmentUpdateFailed:  "Failed to update attachment",
			TaskAttachmentDeleteSuccess: "Attachment removed successfully",
			TaskAttachmentDeleteFailed:  "Failed to remove attachment",
			TaskAttachmentNotFound:      "Attachment not found on this case",
			TaskAttachmentTooLarge:      "Attachment exceeds the size limit",
			TaskAttachmentTypeRejected:  "Attachment file type is not allowed",
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

// AddTaskMediaRequest is used by POST /tasks/{id}/media to attach one or more media
// items to an existing task.
type AddTaskMediaRequest struct {
	MediaIds []string `json:"mediaIds,omitempty" bson:"mediaIds,omitempty"`
}

// AddTaskMediaResponse returns the updated task and which media IDs were added or skipped.
type AddTaskMediaResponse struct {
	Task            models.Task `json:"task,omitempty" bson:"task,omitempty"`
	AddedMediaIds   []string    `json:"addedMediaIds,omitempty" bson:"addedMediaIds,omitempty"`
	SkippedMediaIds []string    `json:"skippedMediaIds,omitempty" bson:"skippedMediaIds,omitempty"`
}

type AddTaskMediaSuccessResponse struct {
	SuccessResponse
	Data AddTaskMediaResponse `json:"data,omitempty" bson:"data,omitempty"`
}

type AddTaskMediaErrorResponse struct {
	ErrorResponse
}

// RequestTaskExportRequest is the optional body of POST /tasks/{id}/export.
// When ExportSelection is non-nil the server persists it on the task before
// queueing the export so users can include / exclude individual case_media
// items from the bundle. A nil value preserves the existing selection.
// ExportAttachmentSelection is the parallel knob for task.Attachments[];
// the two arrays are tracked independently so each side's "deselect
// everything" can be expressed without ambiguity.
type RequestTaskExportRequest struct {
	ExportSelection           *[]string `json:"export_selection,omitempty"            bson:"export_selection,omitempty"`
	ExportAttachmentSelection *[]string `json:"export_attachment_selection,omitempty" bson:"export_attachment_selection,omitempty"`
}

// RequestTaskExportResponse is returned by POST /tasks/{id}/export and
// carries the updated task whose export job has just been queued.
type RequestTaskExportResponse struct {
	Task models.Task `json:"task,omitempty" bson:"task,omitempty"`
}

type RequestTaskExportSuccessResponse struct {
	SuccessResponse
	Data RequestTaskExportResponse `json:"data,omitempty" bson:"data,omitempty"`
}

type RequestTaskExportErrorResponse struct {
	ErrorResponse
}

// TaskIdRequest represents task endpoints that identify a task by URI id.
type TaskIdRequest struct {
	Id string `uri:"id" json:"id,omitempty" bson:"id,omitempty"`
}

// TaskCommentIdRequest represents comment endpoints scoped to a task.
type TaskCommentIdRequest struct {
	Id        string `uri:"id" json:"id,omitempty" bson:"id,omitempty"`
	CommentId string `uri:"comment_id" json:"commentId,omitempty" bson:"commentId,omitempty"`
}

// GetTasksRequest captures query parameters for GET /tasks.
type GetTasksRequest struct {
	Limit  int    `form:"limit,omitempty" json:"limit,omitempty" bson:"limit,omitempty"`
	Offset int    `form:"offset,omitempty" json:"offset,omitempty" bson:"offset,omitempty"`
	Cursor string `form:"cursor,omitempty" json:"cursor,omitempty" bson:"cursor,omitempty"`
	View   string `form:"view,omitempty" json:"view,omitempty" bson:"view,omitempty"` // "full" (default), "compact", or "overview"
}

// TaskFilter defines filtering options for listing tasks.
type TaskFilter struct {
	TaskIds   []string `json:"taskIds,omitempty" bson:"taskIds,omitempty"`
	Title     string   `json:"title" bson:"title,omitempty"`
	View      string   `json:"view" bson:"view,omitempty"` // "full" (default), "compact", or "overview"
	Limit     int      `json:"limit" bson:"limit,omitempty"`
	Sites     []string `json:"sites" bson:"sites,omitempty"`
	Devices   []string `json:"devices" bson:"devices,omitempty"`
	Groups    []string `json:"groups" bson:"groups,omitempty"`
	Assignees []string `json:"assignees" bson:"assignees,omitempty"`
	Labels    []string `json:"labels" bson:"labels,omitempty"`
	Status    []string `json:"status" bson:"status,omitempty"`
	Offset    int      `json:"offset" bson:"offset,omitempty"`
}

const (
	TaskViewFull     = "full"
	TaskViewCompact  = "compact"
	TaskViewOverview = "overview"
)

// GetTasksResponse represents task list payloads returned by list endpoints.
type GetTasksResponse struct {
	Tasks []models.Task `json:"tasks,omitempty" bson:"tasks,omitempty"`
}

// TaskOverview is used for task list views that do not require media URL enrichment.
// It intentionally excludes heavy media payloads such as export_files.
type TaskOverview struct {
	Id                primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreationDate      int64              `json:"creation_date,omitempty" bson:"creation_date,omitempty"`
	CreationDateTime  string             `json:"creation_datetime,omitempty" bson:"creation_datetime,omitempty"`
	MediaTimestamp    int64              `json:"media_timestamp,omitempty" bson:"media_timestamp,omitempty"`
	MediaEndTimestamp int64              `json:"media_end_timestamp,omitempty" bson:"media_end_timestamp,omitempty"`
	MediaDateTime     string             `json:"media_datetime,omitempty" bson:"media_datetime,omitempty"`
	Title             string             `json:"title,omitempty" bson:"title,omitempty"`
	Notes             string             `json:"notes,omitempty" bson:"notes,omitempty"`
	NotesShort        string             `json:"notes_short,omitempty" bson:"notes_short,omitempty"`
	Status            string             `json:"status,omitempty" bson:"status,omitempty"`
	IsPrivate         bool               `json:"is_private,omitempty" bson:"is_private,omitempty"`
	ReporterId        string             `json:"reporter_id,omitempty" bson:"reporter_id,omitempty"`
	Reporter          string             `json:"reporter,omitempty" bson:"reporter,omitempty"`
	ReporterEmail     string             `json:"reporterEmail,omitempty" bson:"reporterEmail,omitempty"`
	Assignees         []string           `json:"assignees,omitempty" bson:"assignees,omitempty"`
	Labels            []string           `json:"labels,omitempty" bson:"labels,omitempty"`
	Cameras           []string           `json:"cameras,omitempty" bson:"cameras,omitempty"`
	CameraNames       []string           `json:"camera_names,omitempty" bson:"camera_names,omitempty"`
	ThumbnailUrl      string             `json:"thumbnail_url,omitempty" bson:"thumbnail_url,omitempty"`
	SequenceId        string             `json:"sequence_id,omitempty" bson:"sequence_id,omitempty"`
	CompressedUrl     string             `json:"compressed_url,omitempty" bson:"compressed_url,omitempty"`
	ExportStatus      string             `json:"export_status,omitempty" bson:"export_status,omitempty"`
	ExportFilesCount  int                `json:"export_files_count,omitempty" bson:"export_files_count,omitempty"`
	DownloadedFiles   []string           `json:"downloaded_files,omitempty" bson:"downloaded_files,omitempty"`
	MediaCount        int                `json:"mediaCount,omitempty" bson:"mediaCount,omitempty"`
	ExpiresAt         *int64             `json:"expires_at,omitempty" bson:"expires_at,omitempty"`
}

type GetTasksOverviewResponse struct {
	Tasks []TaskOverview `json:"tasks,omitempty" bson:"tasks,omitempty"`
}

type GetTasksOverviewSuccessResponse struct {
	SuccessResponse
	Data GetTasksOverviewResponse `json:"data,omitempty" bson:"data,omitempty"`
}

type GetTasksOverviewErrorResponse struct {
	ErrorResponse
}

type GetTasksSuccessResponse struct {
	SuccessResponse
	Data GetTasksResponse `json:"data,omitempty" bson:"data,omitempty"`
}

type GetTasksErrorResponse struct {
	ErrorResponse
}

const (
	TaskStatusOpen     TaskStatus = "open"
	TaskStatusApproved TaskStatus = "approved"
	TaskStatusRejected TaskStatus = "rejected"
)

type GetTaskByIdResponse struct {
	Task models.Task `json:"task,omitempty" bson:"task,omitempty"`
}

type GetTaskByIdSuccessResponse struct {
	SuccessResponse
	Data GetTaskByIdResponse `json:"data,omitempty" bson:"data,omitempty"`
}

type GetTaskByIdErrorResponse struct {
	ErrorResponse
}

// GetTaskMediaRequest captures URI + query parameters for legacy GET /tasks/{id}/media.
// This endpoint is intended for on-demand media URL enrichment when a task is opened.
type GetTaskMediaRequest struct {
	Id     string `uri:"id" json:"id,omitempty" bson:"id,omitempty"`
	Cursor string `form:"cursor,omitempty" json:"cursor,omitempty" bson:"cursor,omitempty"`
	Limit  int64  `form:"limit,omitempty" json:"limit,omitempty" bson:"limit,omitempty"`
}

// GetTaskMediaRequestBody matches POST /tasks/{id}/media/filter request body.
// The task id remains in the URI; filtering/pagination settings live in the body.
type GetTaskMediaRequestBody struct {
	Filter     map[string]interface{} `json:"filter,omitempty" bson:"filter,omitempty"`
	Pagination CursorPagination       `json:"pagination" bson:"pagination"`
}

// TaskMediaItem is the API representation of a media item attached to a task.
type TaskMediaItem = models.ExportFile

type GetTaskMediaResponse struct {
	TaskId string          `json:"taskId,omitempty" bson:"taskId,omitempty"`
	Media  []TaskMediaItem `json:"media,omitempty" bson:"media,omitempty"`
}

type GetTaskMediaSuccessResponse struct {
	SuccessResponse
	Data GetTaskMediaResponse `json:"data,omitempty" bson:"data,omitempty"`
}

type GetTaskMediaErrorResponse struct {
	ErrorResponse
}

// GetTasksFilteredRequest matches POST /tasks/filter request body.
// It supports both:
// - legacy direct filters: { title, status, limit, offset, ... }
// - preferred wrapped form: { filter: {...}, pagination: { cursor, limit } }
type GetTasksFilteredRequest struct {
	TaskFilter `bson:",inline"`
	Filter     *TaskFilter       `json:"filter,omitempty" bson:"filter,omitempty"`
	Pagination *CursorPagination `json:"pagination,omitempty" bson:"pagination,omitempty"`
}

// GetTasksFilteredQuery captures query parameters for POST /tasks/filter.
type GetTasksFilteredQuery struct {
	Limit int `form:"limit,omitempty" json:"limit,omitempty" bson:"limit,omitempty"`
}

// TaskCompact is used by lightweight task pickers that only need summary fields.
type TaskCompact struct {
	Id               primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreationDate     int64              `json:"creation_date,omitempty" bson:"creation_date,omitempty"`
	CreationDateTime string             `json:"creation_datetime,omitempty" bson:"creation_datetime,omitempty"`
	Title            string             `json:"title,omitempty" bson:"title,omitempty"`
	Status           string             `json:"status,omitempty" bson:"status,omitempty"`
	IsPrivate        bool               `json:"is_private,omitempty" bson:"is_private,omitempty"`
	ThumbnailUrl     string             `json:"thumbnail_url,omitempty" bson:"thumbnail_url,omitempty"`
}

type GetTasksFilteredResponse struct {
	Tasks []models.Task `json:"tasks,omitempty" bson:"tasks,omitempty"`
}

type GetTasksCompactResponse struct {
	Tasks []TaskCompact `json:"tasks,omitempty" bson:"tasks,omitempty"`
}

type GetTasksCompactSuccessResponse struct {
	SuccessResponse
	Data GetTasksCompactResponse `json:"data,omitempty" bson:"data,omitempty"`
}

type GetTasksCompactErrorResponse struct {
	ErrorResponse
}

type GetTasksFilteredSuccessResponse struct {
	SuccessResponse
	Data GetTasksFilteredResponse `json:"data,omitempty" bson:"data,omitempty"`
}

type GetTasksFilteredErrorResponse struct {
	ErrorResponse
}

type GetTaskStatisticsResponse struct {
	Statistics models.TaskStatistics `json:"statistics,omitempty" bson:"statistics,omitempty"`
}

type GetTaskStatisticsSuccessResponse struct {
	SuccessResponse
	Data GetTaskStatisticsResponse `json:"data,omitempty" bson:"data,omitempty"`
}

type GetTaskStatisticsErrorResponse struct {
	ErrorResponse
}

// AddTaskRequest wraps the task payload used by POST /tasks.
// The task payload can optionally include mediaFilter for server-side export file resolution.
type AddTaskRequest struct {
	Task AddTaskPayload `json:"task" bson:"task"`
}

// AddTaskPayload mirrors models.Task while allowing transport-level options.
type AddTaskPayload struct {
	models.Task `bson:",inline"`
	MediaFilter *MediaFilter `json:"mediaFilter,omitempty" bson:"mediaFilter,omitempty"`
}

type AddTaskResponse struct {
	Task  models.Task   `json:"task,omitempty" bson:"task,omitempty"`
	Tasks []models.Task `json:"tasks,omitempty" bson:"tasks,omitempty"`
}

type AddTaskSuccessResponse struct {
	SuccessResponse
	Data AddTaskResponse `json:"data,omitempty" bson:"data,omitempty"`
}

type AddTaskErrorResponse struct {
	ErrorResponse
}

// EditTaskRequest matches PATCH /tasks/{id} request payload.
// It includes only fields currently editable from the frontend task UI.
type EditTaskRequest struct {
	Status           *TaskStatus `json:"status,omitempty" bson:"status,omitempty"`
	Notes            *string     `json:"notes,omitempty" bson:"notes,omitempty"`
	Labels           *[]string   `json:"labels,omitempty" bson:"labels,omitempty"`
	Assignees        *[]string   `json:"assignees,omitempty" bson:"assignees,omitempty"`
	AssigneesProfile *[]string   `json:"assignees_profile,omitempty" bson:"assignees_profile,omitempty"`
	NotifyAssignees  *bool       `json:"notify_assignees,omitempty" bson:"notify_assignees,omitempty"`
	IsPrivate        *bool       `json:"is_private,omitempty" bson:"is_private,omitempty"`

	// Curation templates — pointer-to-slice so callers can distinguish
	// "field omitted" (no-op) from "field present with []" (empty
	// allow-list, which downstream readers interpret as "include all").
	// These mirror the per-side modal drafts and are persisted on every
	// inline checkbox toggle so the selection survives reloads. The
	// share-side arrays are templates only — once a share token is
	// created the snapshot lives on the CaseShare row itself.
	ExportSelection           *[]string `json:"export_selection,omitempty" bson:"export_selection,omitempty"`
	ExportAttachmentSelection *[]string `json:"export_attachment_selection,omitempty" bson:"export_attachment_selection,omitempty"`
	ShareSelection            *[]string `json:"share_selection,omitempty" bson:"share_selection,omitempty"`
	ShareAttachmentSelection  *[]string `json:"share_attachment_selection,omitempty" bson:"share_attachment_selection,omitempty"`
}

type EditTaskResponse struct {
	UpdatedFields map[string]interface{} `json:"updatedFields,omitempty" bson:"updatedFields,omitempty"`
	Task          models.Task            `json:"task,omitempty" bson:"task,omitempty"`
}

type EditTaskSuccessResponse struct {
	SuccessResponse
	Data EditTaskResponse `json:"data,omitempty" bson:"data,omitempty"`
}

type EditTaskErrorResponse struct {
	ErrorResponse
}

type DeleteTaskResponse struct {
	Task models.Task `json:"task,omitempty" bson:"task,omitempty"`
}

type DeleteTaskSuccessResponse struct {
	SuccessResponse
	Data DeleteTaskResponse `json:"data,omitempty" bson:"data,omitempty"`
}

type DeleteTaskErrorResponse struct {
	ErrorResponse
}

type GetTaskCommentsResponse struct {
	Comments []models.Comment `json:"comments,omitempty" bson:"comments,omitempty"`
}

type GetTaskCommentsSuccessResponse struct {
	SuccessResponse
	Data GetTaskCommentsResponse `json:"data,omitempty" bson:"data,omitempty"`
}

type GetTaskCommentsErrorResponse struct {
	ErrorResponse
}

type AddTaskCommentRequest struct {
	Comment models.Comment `json:"comment" bson:"comment"`
}

type AddTaskCommentResponse struct {
	Comment models.Comment `json:"comment,omitempty" bson:"comment,omitempty"`
}

type AddTaskCommentSuccessResponse struct {
	SuccessResponse
	Data AddTaskCommentResponse `json:"data,omitempty" bson:"data,omitempty"`
}

type AddTaskCommentErrorResponse struct {
	ErrorResponse
}

type EditTaskCommentRequest struct {
	Comment models.Comment `json:"comment" bson:"comment"`
}

type EditTaskCommentResponse struct {
	Comment models.Comment `json:"comment,omitempty" bson:"comment,omitempty"`
}

type EditTaskCommentSuccessResponse struct {
	SuccessResponse
	Data EditTaskCommentResponse `json:"data,omitempty" bson:"data,omitempty"`
}

type EditTaskCommentErrorResponse struct {
	ErrorResponse
}

type DeleteTaskCommentResponse struct {
	Message string `json:"message,omitempty" bson:"message,omitempty"`
}

type DeleteTaskCommentSuccessResponse struct {
	SuccessResponse
	Data DeleteTaskCommentResponse `json:"data,omitempty" bson:"data,omitempty"`
}

type DeleteTaskCommentErrorResponse struct {
	ErrorResponse
}

// ===== Case media (sources + edits) =====

// CreateMediaEditRequest is the body of POST /tasks/{taskId}/media-edits.
// It describes a single edit to apply to a source CaseMedia entry that
// already lives on the case. The server validates Action / EditType,
// allocates the next Version and enqueues to the matching worker.
//
// For Action = "composite" the Params map is expected to contain an
// "operations" array, each entry having an "op" discriminator matching
// one of the single-action CaseMediaAction values.
type CreateMediaEditRequest struct {
	SourceCaseMediaId string `json:"sourceCaseMediaId,omitempty"`
	// SourceVideoFile points at a legacy task.export_files entry by its
	// storage key. When SourceCaseMediaId is empty the API resolves
	// this key against the task's ExportFiles, lazily creates a
	// Role=source CaseMedia row for it (idempotent — reuses an
	// existing row with the same video_file when present), and then
	// applies the edit to that row. Lets pre-migration cases be
	// redacted without requiring a workspace-wide backfill.
	SourceVideoFile string                   `json:"sourceVideoFile,omitempty"`
	Action          models.CaseMediaAction   `json:"action"`
	EditType        models.CaseMediaEditType `json:"editType,omitempty"`
	Params          map[string]interface{}   `json:"params,omitempty"`
	SupersedesId    string                   `json:"supersedesId,omitempty"`
}

type CreateMediaEditResponse struct {
	CaseMedia models.CaseMedia `json:"caseMedia"`
}

type CreateMediaEditSuccessResponse struct {
	SuccessResponse
	Data CreateMediaEditResponse `json:"data"`
}

type CreateMediaEditErrorResponse struct {
	ErrorResponse
}

// ListCaseMediaResponse returns every case_media entry attached to a
// task (sources and edits) so the case view can render the inventory
// without further joins.
type ListCaseMediaResponse struct {
	CaseMedia []models.CaseMedia `json:"caseMedia"`
}

type ListCaseMediaSuccessResponse struct {
	SuccessResponse
	Data ListCaseMediaResponse `json:"data"`
}

type ListCaseMediaErrorResponse struct {
	ErrorResponse
}

// UpdateCaseMediaSelectedVersionRequest is the body of
// PATCH /tasks/{taskId}/media/{caseMediaId}/selected-version.
//
// It targets a Role = "source" case_media row and records which
// derivative the case should display and export. SelectedVersionId
// must reference an existing Role = "edit" CaseMedia entry that
// descends from the source (directly via ParentId or transitively
// via SupersedesId). Sending an empty SelectedVersionId clears the
// selection, restoring the default behaviour (latest completed edit
// if any, otherwise the source itself).
type UpdateCaseMediaSelectedVersionRequest struct {
	SelectedVersionId string `json:"selectedVersionId"`
}

type UpdateCaseMediaSelectedVersionResponse struct {
	CaseMedia models.CaseMedia `json:"caseMedia"`
}

type UpdateCaseMediaSelectedVersionSuccessResponse struct {
	SuccessResponse
	Data UpdateCaseMediaSelectedVersionResponse `json:"data"`
}

type UpdateCaseMediaSelectedVersionErrorResponse struct {
	ErrorResponse
}

// UpdateCaseMediaCurationRequest is the body of
// PATCH /tasks/{taskId}/media/{caseMediaId}/curation.
//
// Each field is a pointer so callers can patch one flag without
// having to round-trip the other. Both flags target Role = "source"
// rows only — edits inherit the source's inclusion state at resolve
// time. nil means "leave as-is".
type UpdateCaseMediaCurationRequest struct {
	IncludeInExport *bool `json:"includeInExport,omitempty"`
	IncludeInShare  *bool `json:"includeInShare,omitempty"`
}

type UpdateCaseMediaCurationResponse struct {
	CaseMedia models.CaseMedia `json:"caseMedia"`
}

type UpdateCaseMediaCurationSuccessResponse struct {
	SuccessResponse
	Data UpdateCaseMediaCurationResponse `json:"data"`
}

type UpdateCaseMediaCurationErrorResponse struct {
	ErrorResponse
}

// ===== Case attachments (auxiliary, non-pipeline files on a case) =====
//
// Attachments are PDFs, images, scanned documents etc. attached to a
// case. They are NOT part of the worker pipeline (no version chain,
// no worker-driven status). The collection lives embedded on
// Task.Attachments — see models.CaseAttachment for the schema and the
// per-case soft-cap rationale.
//
// Upload uses multipart/form-data (single-POST + TeeReader streaming
// inside hub-api). The bytes are forwarded to Vault and never persisted
// inside Mongo. Signed Url / ThumbnailUrl fields on the response are
// minted at fetch time and never persisted.

// UploadCaseAttachmentRequest documents the metadata accepted on a
// multipart upload to POST /tasks/{taskId}/attachments. The actual
// upload field is named `file`; all other fields are optional form
// values that override what hub-api would otherwise derive from the
// uploaded part.
//
// Defined as a struct (rather than free-floating form params) so swag
// can generate a consistent shape for OpenAPI / TS clients. The Go
// controller reads these via c.PostForm / c.FormFile.
type UploadCaseAttachmentRequest struct {
	// Name overrides the filename recorded on the attachment. Defaults
	// to the multipart part filename when omitted.
	Name string `json:"name,omitempty" form:"name"`

	// RelatedCaseMediaId optionally links the attachment to a specific
	// case_media entry it documents or annotates (annotated screenshot
	// of a redacted clip, etc.). Hex ObjectID; must belong to the same
	// task.
	RelatedCaseMediaId string `json:"relatedCaseMediaId,omitempty" form:"relatedCaseMediaId"`
}

type UploadCaseAttachmentResponse struct {
	Attachment models.CaseAttachment `json:"attachment"`
}

type UploadCaseAttachmentSuccessResponse struct {
	SuccessResponse
	Data UploadCaseAttachmentResponse `json:"data"`
}

type UploadCaseAttachmentErrorResponse struct {
	ErrorResponse
}

// ListCaseAttachmentsResponse returns every attachment embedded on the
// task. URLs are signed at fetch time so the case detail view can
// render the list without further round trips.
type ListCaseAttachmentsResponse struct {
	Attachments []models.CaseAttachment `json:"attachments"`
}

type ListCaseAttachmentsSuccessResponse struct {
	SuccessResponse
	Data ListCaseAttachmentsResponse `json:"data"`
}

type ListCaseAttachmentsErrorResponse struct {
	ErrorResponse
}

type GetCaseAttachmentResponse struct {
	Attachment models.CaseAttachment `json:"attachment"`
}

type GetCaseAttachmentSuccessResponse struct {
	SuccessResponse
	Data GetCaseAttachmentResponse `json:"data"`
}

type GetCaseAttachmentErrorResponse struct {
	ErrorResponse
}

// UpdateCaseAttachmentRequest covers in-place metadata edits that do
// not require re-uploading the bytes. Name is the original mutable
// field; IncludeInExport / IncludeInShare are the per-attachment
// curation flags. All fields are pointer-typed so a partial PATCH
// can target one without touching the others (nil = leave as-is).
type UpdateCaseAttachmentRequest struct {
	Name            *string `json:"name,omitempty"`
	IncludeInExport *bool   `json:"includeInExport,omitempty"`
	IncludeInShare  *bool   `json:"includeInShare,omitempty"`
}

type UpdateCaseAttachmentResponse struct {
	Attachment models.CaseAttachment `json:"attachment"`
}

type UpdateCaseAttachmentSuccessResponse struct {
	SuccessResponse
	Data UpdateCaseAttachmentResponse `json:"data"`
}

type UpdateCaseAttachmentErrorResponse struct {
	ErrorResponse
}

type DeleteCaseAttachmentResponse struct {
	// Id of the removed attachment, echoed back for client cache
	// invalidation.
	Id string `json:"id"`
}

type DeleteCaseAttachmentSuccessResponse struct {
	SuccessResponse
	Data DeleteCaseAttachmentResponse `json:"data"`
}

type DeleteCaseAttachmentErrorResponse struct {
	ErrorResponse
}
