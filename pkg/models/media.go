package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Media struct {
	// Unique identifier for the media file
	Id primitive.ObjectID `json:"id" bson:"_id,omitempty,omitempty"`

	// Time window of media file.
	StartTimestamp int64 `json:"startTimestamp,omitempty" bson:"startTimestamp,omitempty"`
	EndTimestamp   int64 `json:"endTimestamp,omitempty" bson:"endTimestamp,omitempty"`
	Duration       int   `json:"duration,omitempty" bson:"duration,omitempty"`

	// RBAC information
	// DeviceId is a unique identifier for the device, it can be used to identify the device in the system.
	// OrganisationId is used to identify the organisation that owns the device.
	DeviceId       string `json:"deviceId" bson:"deviceId,omitempty"` // device identifier
	GroupId        string `json:"groupId" bson:"groupId,omitempty"`
	SiteId         string `json:"siteId" bson:"siteId,omitempty"`
	OrganisationId string `json:"organisationId" bson:"organisationId,omitempty"`

	// Media file information (by default "vault", however might change
	// in the future (integration with other storage solutions, next to Vault).
	StorageSolution   string `json:"storageSolution,omitempty" bson:"storageSolution,omitempty"`
	VideoFile         string `json:"videoFile,omitempty" bson:"videoFile,omitempty"`
	VideoProvider     string `json:"videoProvider,omitempty" bson:"videoProvider,omitempty"`
	ThumbnailFile     string `json:"thumbnailFile,omitempty" bson:"thumbnailFile,omitempty"`
	ThumbnailProvider string `json:"thumbnailProvider,omitempty" bson:"thumbnailProvider,omitempty"`
	SpriteFile        string `json:"spriteFile,omitempty" bson:"spriteFile,omitempty"`
	SpriteProvider    string `json:"spriteProvider,omitempty" bson:"spriteProvider,omitempty"`
	RedactionFile     string `json:"redactionFile,omitempty" bson:"redactionFile,omitempty"`
	RedactionProvider string `json:"redactionProvider,omitempty" bson:"redactionProvider,omitempty"`

	ClassificationSummary []ClassificationSummary `json:"classificationSummary,omitempty" bson:"classificationSummary,omitempty"`

	// Name of the device that uploaded media
	DeviceName string `json:"deviceName,omitempty" bson:"deviceName,omitempty"`

	// Metadata
	Metadata *MediaMetadata `json:"metadata,omitempty" bson:"metadata,omitempty"`

	// AtRuntimeMetadata contains metadata that is generated at runtime, which can include
	// more verbose information about the device's current state, capabilities, or configuration.
	// for example the linked sites details, etc.
	AtRuntimeMetadata *MediaAtRuntimeMetadata `json:"atRuntimeMetadata,omitempty" bson:"atRuntimeMetadata,omitempty"`

	// Audit information
	Audit *Audit `json:"audit,omitempty" bson:"audit,omitempty"`
}

// We can store additional metadata for media files, such as tags and classifications.
type MediaMetadata struct {
	// Media containers related information
	Container  string `json:"containerType,omitempty" bson:"containerType,omitempty"` // e.g., mp4, mkv, avi
	Resolution string `json:"resolution,omitempty" bson:"resolution,omitempty"`       // e.g., 1920x1080
	Width      int    `json:"width,omitempty" bson:"width,omitempty"`                 // in pixels
	Height     int    `json:"height,omitempty" bson:"height,omitempty"`               // in pixels
	Codec      string `json:"codec,omitempty" bson:"codec,omitempty"`                 // e.g., H.264, VP9
	Bitrate    int    `json:"bitrate,omitempty" bson:"bitrate,omitempty"`             // in kbps
	FPS        int    `json:"fps,omitempty" bson:"fps,omitempty"`                     // frames per second

	// Tags associated to give some context about the media file
	Tags []string `json:"tags,omitempty" bson:"tags,omitempty"`

	// Sprite interval in seconds
	SpriteInterval int `json:"spriteInterval,omitempty" bson:"spriteInterval,omitempty"`

	// Motion information
	MotionPixels     int     `json:"motionPixels,omitempty" bson:"motionPixels,omitempty"`
	MotionPercentage float64 `json:"motionPercentage,omitempty" bson:"motionPercentage,omitempty"`

	// Analysis data (we keep a reference to the original analysis, and cache some data here)
	AnalysisId      string           `json:"analysisId,omitempty" bson:"analysisId,omitempty"`
	Classifications []Classification `json:"classifications,omitempty" bson:"classifications,omitempty"`
	Description     string           `json:"description,omitempty" bson:"description,omitempty"`
	Detections      []string         `json:"detections,omitempty" bson:"detections,omitempty"`
	DominantColors  []string         `json:"dominantColors,omitempty" bson:"dominantColors,omitempty"`
	Count           int              `json:"count,omitempty" bson:"count,omitempty"`
	Embedding       []int            `json:"embedding,omitempty" bson:"embedding,omitempty"`
}

// MediaAtRuntimeMetadata contains metadata that is generated at runtime, which can include
type MediaAtRuntimeMetadata struct {
	CachedTimestamp int64            `json:"cachedTimestamp,omitempty" bson:"cachedTimestamp,omitempty"` // Timestamp when the runtime metadata was cached.
	VideoUrl        string           `json:"videoUrl,omitempty" bson:"videoUrl,omitempty"`
	ThumbnailUrl    string           `json:"thumbnailUrl,omitempty" bson:"thumbnailUrl,omitempty"`
	SpriteUrl       string           `json:"spriteUrl,omitempty" bson:"spriteUrl,omitempty"`
	RedactionUrl    string           `json:"redactionUrl,omitempty" bson:"redactionUrl,omitempty"`
	Analysis        *AnalysisWrapper `json:"analysis,omitempty" bson:"analysis,omitempty"`
	Device          *Device          `json:"device,omitempty" bson:"device,omitempty"`
}

type Region struct {
	Id           string  `json:"id" bson:"id"`
	Device       string  `json:"device" bson:"device"`
	Width        int     `json:"width" bson:"width"`
	Height       int     `json:"height" bson:"height"`
	RegionPoints []Point `json:"regionPoints" bson:"regionPoints"`
}

type HourRange struct {
	Start int64 `json:"start" bson:"start"`
	End   int64 `json:"end" bson:"end"`
}

/*
Media object as used in Vault, should be more aligned with how we store in Hub
*/

type VaultMedia struct {
	Timestamp         int64              `json:"timestamp" bson:"timestamp"`
	FileName          string             `json:"filename" bson:"filename"`
	FileSize          int64              `json:"filesize" bson:"filesize"`
	Device            string             `json:"device" bson:"device"`
	Account           string             `json:"account" bson:"account"`
	Provider          string             `json:"provider" bson:"provider"`
	Status            string             `json:"status" bson:"status"`
	Finished          bool               `json:"finished" bson:"finished"`
	Temporary         bool               `json:"temporary" bson:"temporary"`
	Forwarded         bool               `json:"forwarded" bson:"forwarded"`
	ToBeForwarded     bool               `json:"to_be_forwarded" bson:"to_be_forwarded"`
	Uploaded          bool               `json:"uploaded" bson:"uploaded"`
	ForwarderId       string             `json:"forwarder_id" bson:"forwarder_id"`
	ForwarderType     string             `json:"forwarder_type" bson:"forwarder_type"`
	ForwarderWorker   string             `json:"forwarder_worker" bson:"forwarder_worker"`
	ForwardTimestamp  int64              `json:"forward_timestamp" bson:"forward_timestamp"`
	Events            []VaultMediaEvent  `json:"events" bson:"events"`
	MainProvider      bool               `json:"main_provider" bson:"main_provider"`
	SecondaryProvider bool               `json:"secondary_provider" bson:"secondary_provider"`
	Metadata          VaultMediaMetadata `json:"metadata" bson:"metadata"`
	UriExpiryTime     string             `json:"uriExpiryTime" bson:"uriExpiryTime"`
}

type VaultMediaMetadata struct {
	BytesRanges      string                       `json:"bytes_ranges" bson:"bytes_ranges"`
	BytesRangeOnTime []FragmentedBytesRangeOnTime `json:"bytes_range_on_time" bson:"bytes_range_on_time"`
	IsFragmented     bool                         `json:"is_fragmented" bson:"is_fragmented"`
	Duration         uint64                       `json:"duration" bson:"duration"`
	Timescale        uint32                       `json:"timescale" bson:"timescale"`
}

type VaultMediaEvent struct {
	Timestamp   int64  `json:"timestamp" bson:"timestamp"`
	Title       string `json:"title" bson:"title"`
	Description string `json:"description" bson:"description"`
}

type VaultMediaFragmentCollection struct {
	Key              string                       `json:"key" bson:"key"`
	FileName         string                       `json:"filename" bson:"filename"`
	CameraId         string                       `json:"camera_id" bson:"camera_id"`
	Timestamp        int64                        `json:"timestamp" bson:"timestamp"`
	Url              string                       `json:"url" bson:"url"`
	Start            float64                      `json:"start" bson:"start"`
	End              float64                      `json:"end" bson:"end"`
	Duration         float64                      `json:"duration" bson:"duration"`
	BytesRanges      string                       `json:"bytes_ranges" bson:"bytes_ranges"`
	BytesRangeOnTime []FragmentedBytesRangeOnTime `json:"bytes_range_on_time" bson:"bytes_range_on_time"`
}

type TimeRange struct {
	Start int64 `json:"start,omitempty" bson:"start,omitempty"`
	End   int64 `json:"end,omitempty" bson:"end,omitempty"`
}

type ClassificationSummary struct {
	Key   string `json:"key" bson:"key"`
	Count int    `json:"count" bson:"count"`
	// Additional attributes can be added as needed to be shown in the front end
}

type Classification struct {
	Key       string       `json:"key" bson:"key"`
	Centroids [][2]float64 `json:"centroids" bson:"centroids"` // e.g., [[x1, y1], [x2, y2], ...]
	// Additional attributes can be added as needed
}
