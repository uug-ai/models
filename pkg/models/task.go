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
	ExportFiles      []ExportFile `json:"export_files" bson:"export_files,omitempty"`
	ExportFilesCount int          `json:"export_files_count" bson:"export_files_count,omitempty"`
	DownloadedFiles  []string     `json:"downloaded_files" bson:"downloaded_files,omitempty"`

	// Related collections
	Comments []Comment `json:"comments" bson:"comments,omitempty"`
	Labels   []string  `json:"labels" bson:"labels,omitempty"`
}

type TaskFilter struct {
	Title     string   `json:"title" bson:"title,omitempty"`
	Limit     int      `json:"limit" bson:"limit,omitempty"`
	Sites     []string `json:"sites" bson:"sites,omitempty"`
	Devices   []string `json:"devices" bson:"devices,omitempty"`
	Groups    []string `json:"groups" bson:"groups,omitempty"`
	Assignees []string `json:"assignees" bson:"assignees,omitempty"`
	Labels    []string `json:"labels" bson:"labels,omitempty"`
	Status    []string `json:"status" bson:"status,omitempty"`
	Offset    int      `json:"offset" bson:"offset,omitempty"`
}

type ExportFile struct {
	Timestamp         int64  `json:"timestamp" bson:"timestamp"`
	Key               string `json:"key" bson:"key,omitempty"`
	CameraId          string `json:"camera_id" bson:"camera_id,omitempty"`
	Provider          string `json:"provider" bson:"provider,omitempty"`
	Source            string `json:"source" bson:"source,omitempty"`
	SpriteFile        string `json:"spriteFile" bson:"spriteFile"`
	SpriteProvider    string `json:"spriteProvider" bson:"spriteProvider"`
	SpriteInterval    int    `json:"spriteInterval" bson:"spriteInterval"`
	ThumbnailFile     string `json:"thumbnailFile" bson:"thumbnailFile"`
	ThumbnailProvider string `json:"thumbnailProvider" bson:"thumbnailProvider"`
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
