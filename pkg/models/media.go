package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Media struct {
	Id primitive.ObjectID `json:"id" bson:"_id,omitempty,omitempty"`

	// Time window of media file.
	StartTimestamp int64 `json:"startTimestamp,omitempty" bson:"startTimestamp,omitempty"`
	EndTimestamp   int64 `json:"endTimestamp,omitempty" bson:"endTimestamp,omitempty"`
	Duration       int64 `json:"duration,omitempty" bson:"duration,omitempty"`

	// RBAC information
	// DeviceId is a unique identifier for the device, it can be used to identify the device in the system.
	// OrganisationId is used to identify the organisation that owns the device.
	DeviceId       string `json:"deviceId" bson:"deviceId,omitempty"` // device identifier
	OrganisationId string `json:"organisationId" bson:"organisationId,omitempty"`

	// Media file information (by default "vault", however might change
	// in the future (integration with other storage solutions, next to Vault).
	StorageSolution string `json:"storageSolution,omitempty" bson:"storageSolution,omitempty"`

	// Vault provider information (contains where the media is stored on which underlaying cloud storage)
	VideoProvider     string `json:"videoProvider,omitempty" bson:"videoProvider,omitempty"`
	ThumbnailProvider string `json:"thumbnailProvider,omitempty" bson:"thumbnailProvider,omitempty"`
	SpriteProvider    string `json:"spriteProvider,omitempty" bson:"spriteProvider,omitempty"`

	// Media file information
	VideoFile      string `json:"videoFile,omitempty" bson:"videoFile,omitempty"`
	ThumbnailFile  string `json:"thumbnailFile,omitempty" bson:"thumbnailFile,omitempty"`
	SpriteFile     string `json:"spriteFile,omitempty" bson:"spriteFile,omitempty"`
	SpriteInterval int    `json:"spriteInterval,omitempty" bson:"spriteInterval,omitempty"`

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
	Tags            []string `json:"tags,omitempty" bson:"tags,omitempty"`
	Classifications []string `json:"classifications,omitempty" bson:"classifications,omitempty"`
}

// MediaAtRuntimeMetadata contains metadata that is generated at runtime, which can include
type MediaAtRuntimeMetadata struct {
	VideoUrl     string `json:"videoUrl,omitempty" bson:"videoUrl,omitempty"`
	ThumbnailUrl string `json:"thumbnailUrl,omitempty" bson:"thumbnailUrl,omitempty"`
	SpriteUrl    string `json:"spriteUrl,omitempty" bson:"spriteUrl,omitempty"`
}

type Region struct {
	Id           string  `json:"id" bson:"id"`
	Device       string  `json:"device" bson:"device"`
	Width        int     `json:"width" bson:"width"`
	Height       int     `json:"height" bson:"height"`
	RegionPoints []Point `json:"regionPoints" bson:"regionPoints"`
}

type Point struct {
	X float64 `json:"x" bson:"x"`
	Y float64 `json:"y" bson:"y"`
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

/*
Media object as used in Agent, should be more aligned with how we store in Hub
*/

type AgentMedia struct {
	FileName               string `json:"filename" bson:"filename"`
	Timestamp              int64  `json:"timestamp" bson:"timestamp"`
	MillisecondsWithLength string `json:"millisecondsWithLength" bson:"millisecondsWithLength"`
	DeviceName             string `json:"deviceName" bson:"deviceName"`
	RegionCoordinates      string `json:"regionCoordinates" bson:"regionCoordinates"`
	Token                  string `json:"token" bson:"token"`
	Duration               int64  `json:"duration" bson:"duration"`
	FramesPerSecond        int    `json:"framesPerSecond" bson:"framesPerSecond"`
	CameraResolution       string `json:"cameraResolution" bson:"cameraResolution"`
	Account                string `json:"account" bson:"account"`
}
