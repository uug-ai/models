package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// error codes
var CREATE_TASK_EMPTY = "CREATE_TASK_EMPTY"
var CREATE_TASK_ALREADY_EXISTS = "CREATE_TASK_ALREADY_EXISTS"
var CREATED_TASKS_SUCCESSFULLY = "CREATED_TASKS_SUCCESSFULLY"
var CREATE_TASK_FAILED = "CREATE_TASK_FAILED"
var FILTER_TASKS_FAILED = "FILTER_TASKS_FAILED"

type TaskWrapper struct {
	Task Task `json:"task" bson:"task"`
}

type TaskStatistics struct {
	Open     int64 `json:"open" bson:"open"`
	Rejected int64 `json:"rejected" bson:"rejected"`
	Approved int64 `json:"approved" bson:"approved"`
	Total    int64 `json:"total" bson:"total"`
}

type Task struct {
	Id                primitive.ObjectID `json:"id" bson:"_id,omitempty,omitempty"`
	CreationDate      int64              `json:"creation_date" bson:"creation_date,omitempty"`
	CreationDateTime  string             `json:"creation_datetime" bson:"creation_datetime,omitempty"`
	Date              int64              `json:"date" bson:"date,omitempty"`
	MediaTimestamp    int64              `json:"media_timestamp" bson:"media_timestamp,omitempty"`
	MediaDate         string             `json:"media_date" bson:"media_date,omitempty"`
	MediaDateTime     string             `json:"media_datetime" bson:"media_datetime,omitempty"`
	MediaEndTimestamp int64              `json:"media_end_timestamp" bson:"media_end_timestamp,omitempty"`
	MediaEndDateTime  string             `json:"media_end_datetime" bson:"media_end_datetime,omitempty"`
	UserId            string             `json:"user_id" bson:"user_id,omitempty"`
	Username          string             `json:"username" bson:"username,omitempty"`
	ReporterId        string             `json:"reporter_id" bson:"reporter_id,omitempty"`
	Title             string             `json:"title" bson:"title,omitempty"`
	Notes             string             `json:"notes" bson:"notes,omitempty"`
	NotesShort        string             `json:"notes_short" bson:"notes_short,omitempty"`
	Status            string             `json:"status" bson:"status,omitempty"` // open, approved, rejected
	SequenceId        string             `json:"sequenceId" bson:"sequenceId,omitempty"`
	IsPrivate         bool               `json:"is_private" bson:"is_private"`

	// A task can be assigned to a single camera or multiple cameras (depending of the export)
	Camera      string   `json:"camera" bson:"camera"`             // this is legacy we know use the array object.
	Cameras     []string `json:"cameras" bson:"cameras"`           // these are camera ids
	CameraNames []string `json:"camera_names" bson:"camera_names"` // this is for the camera names (is computed on demand)

	// Users associated with this task
	Assignees        []string `json:"assignees" bson:"assignees,omitempty"`
	ReporterProfile  string   `json:"reporter_profile" bson:"reporter_profile,omitempty"`
	ReporterEmail    string   `json:"reporter_email" bson:"reporter_email,omitempty"`
	AssigneesProfile []string `json:"assignees_profile" bson:"assignees_profile,omitempty"`
	NotifyAssignees  bool     `json:"notify_assignees" bson:"notify_assignees"`
	AssigneesSentTo  []string `json:"assignees_sent_to" bson:"assignees_sent_to,omitempty"`

	// This is used for a single export, for a bulk export, we will use the first video as thumbnail and sprite.
	// Referencing to some media: video, thumbnail and sprite
	VideoUrl           string `json:"video_url" bson:"video_url,omitempty"`
	VideoFile          string `json:"videoFile" bson:"videoFile"`
	VideoProvider      string `json:"videoProvider" bson:"videoProvider"`
	ThumbnailUrl       string `json:"thumbnail_url" bson:"thumbnail_url,omitempty"`
	Thumbnail          string `json:"thumbnail" bson:"thumbnail,omitempty"` // base 64 encoded
	ThumbnailFile      string `json:"thumbnailFile" bson:"thumbnailFile"`
	ThumbnailProvider  string `json:"thumbnailProvider" bson:"thumbnailProvider"`
	SpriteUrl          string `json:"sprite_url" bson:"sprite_url,omitempty"`
	SpriteFile         string `json:"spriteFile" bson:"spriteFile"`
	SpriteProvider     string `json:"spriteProvider" bson:"spriteProvider"`
	SpriteInterval     int    `json:"spriteInterval" bson:"spriteInterval"`
	CompressedUrl      string `json:"compressed_url" bson:"compressed_url,omitempty"`
	CompressedFile     string `json:"compressedFile" bson:"compressedFile"`
	CompressedProvider string `json:"compressedProvider" bson:"compressedProvider"`

	// An export task, is containing multiple video in a compressed file format (.zip)
	ExportStatus     string       `json:"export_status" bson:"export_status,omitempty"`     // "", "idle", "start", "progress", "success", "error"
	ExportProgress   int          `json:"export_progress" bson:"export_progress,omitempty"` // "0% -> 100%"
	ExportFiles      []ExportFile `json:"export_files" bson:"export_files,omitempty"`       // legacy: read by v20130101 only
	ExportFilesCount int          `json:"export_files_count" bson:"export_files_count,omitempty"`
	DownloadedFiles  []string     `json:"downloaded_files" bson:"downloaded_files,omitempty"`
	ExportRequested  bool         `json:"export_requested" bson:"export_requested,omitempty"`
	ExportInProgress bool         `json:"export_in_progress" bson:"export_in_progress,omitempty"`
	ExportRevision   int64        `json:"export_revision" bson:"export_revision,omitempty"`

	// Number of source case_media rows attached to this task. Mirrors
	// case_media documents where role == "source" and task_id matches.
	MediaCount int `json:"media_count" bson:"media_count,omitempty"`

	// Per-purpose case_media selections. Each slice is an ordered list
	// of case_media ids picked for that purpose. The full inventory
	// lives in the case_media collection (queryable by task_id); these
	// fields only record which subset participates in the export /
	// share bundle and in what order.
	//
	// An empty selection means "default rule": consumers fall back to
	// every source's latest completed edit (or the source itself when
	// no completed edit exists).
	//
	// NOTE on the share fields specifically: ShareSelection and
	// ShareAttachmentSelection are the owner-side TEMPLATE used to
	// pre-fill the next share modal ("start where the last share left
	// off"). They are NOT the source of truth for what an active
	// recipient sees \u2014 each CaseShare row carries its own Selection /
	// AttachmentSelection snapshot captured at CreateShare time, so
	// later edits to this template do not retroactively change what
	// already-issued tokens resolve to.
	ExportSelection []primitive.ObjectID `json:"export_selection,omitempty" bson:"export_selection,omitempty"`
	ShareSelection  []primitive.ObjectID `json:"share_selection,omitempty"  bson:"share_selection,omitempty"`

	// ExportAttachmentSelection is the parallel to ExportSelection for
	// task.Attachments[]. Kept as its own array because attachments
	// live on a different storage path (and a different mongo
	// document shape) than case_media — mixing both into a single
	// selection array would force the export pipeline to consult two
	// collections to disambiguate every id, and would also make
	// "deselect every attachment" indistinguishable from a legacy
	// (media-only) selection. Same semantics as ExportSelection: nil
	// or empty means "default rule" (include every attachment), a
	// non-empty slice is the literal allow-list.
	ExportAttachmentSelection []primitive.ObjectID `json:"export_attachment_selection,omitempty" bson:"export_attachment_selection,omitempty"`

	// ShareAttachmentSelection mirrors ExportAttachmentSelection for
	// the share flow. Same semantics: nil/empty = include every
	// attachment in the recipient view, non-empty = literal
	// allow-list.
	ShareAttachmentSelection []primitive.ObjectID `json:"share_attachment_selection,omitempty" bson:"share_attachment_selection,omitempty"`

	// Attachments are auxiliary, non-pipeline files attached to the
	// case (PDFs, hi-res images, scanned documents, audio notes, …).
	// They are embedded directly here under the assumption that the
	// per-case cardinality stays bounded (soft cap ~100). Only
	// metadata is stored; bytes live in Vault. List-cases endpoints
	// SHOULD project this field out to keep the list view light.
	Attachments []CaseAttachment `json:"attachments,omitempty" bson:"attachments,omitempty"`

	// Retention / lifecycle.
	//
	// RetentionDays captures the *policy intent*: how many days the
	// case should be kept after its retention anchor (typically the
	// moment the case is closed, falling back to CreationDate when
	// the case is still open). Nil means "use the workspace/tenant
	// default policy"; 0 is a valid explicit "delete on next sweep".
	//
	// ExpiresAt is the *materialized* date at which the cleanup
	// worker is allowed to purge the case (and its attachments,
	// media, comments, …). It is recomputed from RetentionDays on
	// write, EXCEPT when ExpiresAtOverridden is true. Nil means
	// "never expires". An index on this field powers the cleanup
	// sweep; do not use a Mongo TTL index — deletion must cascade
	// through the API so downstream storage (Vault, search, audit)
	// stays consistent.
	//
	// ExpiresAtOverridden is set to true the moment a user manually
	// edits ExpiresAt. Once set, policy changes will NOT silently
	// re-materialize ExpiresAt, so a granted extension cannot be
	// undone by a tenant-wide policy tweak. Clearing the override
	// (e.g. "reset to policy") flips this back to false.
	//
	// LegalHold, when true, suppresses cleanup unconditionally
	// regardless of ExpiresAt. It is intentionally a separate flag
	// from the retention date so audits can distinguish "kept longer
	// because investigator extended" from "kept because under legal
	// hold". Managed by a dedicated permission.
	RetentionDays       *int   `json:"retention_days,omitempty"        bson:"retention_days,omitempty"`
	ExpiresAt           *int64 `json:"expires_at,omitempty"            bson:"expires_at,omitempty"`
	ExpiresAtOverridden bool   `json:"expires_at_overridden,omitempty" bson:"expires_at_overridden,omitempty"`
	LegalHold           bool   `json:"legal_hold,omitempty"            bson:"legal_hold,omitempty"`

	// Related collections
	Comments []Comment `json:"comments" bson:"comments,omitempty"`
	Labels   []string  `json:"labels" bson:"labels,omitempty"`
}

// ExportFile is the legacy media descriptor embedded on Task.ExportFiles.
// Kept only for the v20130101 API surface; new code uses CaseMedia.
type ExportFile struct {
	Timestamp         int64  `json:"timestamp" bson:"timestamp"`
	Key               string `json:"key" bson:"key,omitempty"`
	CameraId          string `json:"camera_id" bson:"camera_id,omitempty"`
	Provider          string `json:"provider" bson:"provider,omitempty"`
	Source            string `json:"source" bson:"source,omitempty"`
	VideoUrl          string `json:"video_url,omitempty" bson:"video_url,omitempty"`
	SpriteFile        string `json:"spriteFile" bson:"spriteFile"`
	SpriteProvider    string `json:"spriteProvider" bson:"spriteProvider"`
	SpriteUrl         string `json:"sprite_url,omitempty" bson:"sprite_url,omitempty"`
	SpriteInterval    int    `json:"spriteInterval" bson:"spriteInterval"`
	ThumbnailFile     string `json:"thumbnailFile" bson:"thumbnailFile"`
	ThumbnailProvider string `json:"thumbnailProvider" bson:"thumbnailProvider"`
	ThumbnailUrl      string `json:"thumbnail_url,omitempty" bson:"thumbnail_url,omitempty"`
}

type MediaWrapper struct {
	Time         string   `json:"time" bson:"time"`
	Timestamp    int64    `json:"timestamp" bson:"timestamp"`
	Description  string   `json:"description" bson:"description"`
	Index        int      `json:"index" bson:"index"`
	InstanceName string   `json:"instanceName" bson:"instanceName"`
	Labels       []string `json:"labels" bson:"labels"`
	Properties   []string `json:"properties" bson:"properties"`
	MetaData     Media    `json:"metadata" bson:"metadata"`
	FileUrl      string   `json:"src" bson:"src"`
	ThumbnailUrl string   `json:"thumbnailUrl" bson:"thumbnailUrl"`
	SpriteUrl    string   `json:"spriteUrl" bson:"spriteUrl"`
	Type         string   `json:"type" bson:"type"`
	Vault        bool     `json:"vault" bson:"vault"`
	TaskCreated  bool     `json:"task_created" bson:"task_created"`
	Persisted    bool     `json:"persisted" bson:"persisted"`
}
