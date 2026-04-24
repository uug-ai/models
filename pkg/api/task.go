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
)

// String returns the string representation of the Task status
func (ms TaskStatus) String() string {
	return string(ms)
}

// Into returns the translated string representation of the Task status in the specified language
func (ms TaskStatus) Translate(lang string) string {
	translations := map[string]map[TaskStatus]string{
		"en": {
			TaskBindingFailed: "Task binding failed",
			TaskDuplicateName: "Task duplicate name",
			TaskMissingInfo:   "Task missing information",
			TaskFound:         "Task found",
			TaskNotFound:      "Task not found",
			TaskForbidden:     "You are not allowed to access this task",
			TaskAddSuccess:    "Task added successfully",
			TaskAddFailed:     "Task failed to add",
			TaskUpdateSuccess: "Task updated successfully",
			TaskUpdateFailed:  "Task failed to update",
			TaskDeleteSuccess: "Task deleted successfully",
			TaskDeleteFailed:  "Task failed to delete",
			TaskMediaAddSuccess: "Media was added to the task successfully",
			TaskMediaAddFailed:  "Failed to add media to the task",
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
	Title             string             `json:"title,omitempty" bson:"title,omitempty"`
	Status            string             `json:"status,omitempty" bson:"status,omitempty"`
	IsPrivate         bool               `json:"is_private,omitempty" bson:"is_private,omitempty"`
	ReporterId        string             `json:"reporter_id,omitempty" bson:"reporter_id,omitempty"`
	Assignees         []string           `json:"assignees,omitempty" bson:"assignees,omitempty"`
	Labels            []string           `json:"labels,omitempty" bson:"labels,omitempty"`
	Cameras           []string           `json:"cameras,omitempty" bson:"cameras,omitempty"`
	CameraNames       []string           `json:"camera_names,omitempty" bson:"camera_names,omitempty"`
	ThumbnailUrl      string             `json:"thumbnail_url,omitempty" bson:"thumbnail_url,omitempty"`
	MediaCount        int                `json:"mediaCount,omitempty" bson:"mediaCount,omitempty"`
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

// GetTaskMediaRequest captures URI + query parameters for GET /tasks/{id}/media.
// This endpoint is intended for on-demand media URL enrichment when a task is opened.
type GetTaskMediaRequest struct {
	Id     string `uri:"id" json:"id,omitempty" bson:"id,omitempty"`
	Cursor string `form:"cursor,omitempty" json:"cursor,omitempty" bson:"cursor,omitempty"`
	Limit  int64  `form:"limit,omitempty" json:"limit,omitempty" bson:"limit,omitempty"`
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
	TaskFilter  `bson:",inline"`
	Filter      *TaskFilter       `json:"filter,omitempty" bson:"filter,omitempty"`
	Pagination  *CursorPagination `json:"pagination,omitempty" bson:"pagination,omitempty"`
}

// GetTasksFilteredQuery captures query parameters for POST /tasks/filter.
type GetTasksFilteredQuery struct {
	Limit int `form:"limit,omitempty" json:"limit,omitempty" bson:"limit,omitempty"`
}

// TaskCompact is used by lightweight task pickers that only need summary fields.
type TaskCompact struct {
	Id               primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreationDate     int64         `json:"creation_date,omitempty" bson:"creation_date,omitempty"`
	CreationDateTime string        `json:"creation_datetime,omitempty" bson:"creation_datetime,omitempty"`
	Title            string        `json:"title,omitempty" bson:"title,omitempty"`
	Status           string        `json:"status,omitempty" bson:"status,omitempty"`
	IsPrivate        bool          `json:"is_private,omitempty" bson:"is_private,omitempty"`
	ThumbnailUrl     string        `json:"thumbnail_url,omitempty" bson:"thumbnail_url,omitempty"`
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
	Notes            *string   `json:"notes,omitempty" bson:"notes,omitempty"`
	Labels           *[]string `json:"labels,omitempty" bson:"labels,omitempty"`
	Assignees        *[]string `json:"assignees,omitempty" bson:"assignees,omitempty"`
	AssigneesProfile *[]string `json:"assignees_profile,omitempty" bson:"assignees_profile,omitempty"`
	NotifyAssignees  *bool     `json:"notify_assignees,omitempty" bson:"notify_assignees,omitempty"`
	IsPrivate        *bool     `json:"is_private,omitempty" bson:"is_private,omitempty"`
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
