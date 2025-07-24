package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// --------------------------------------------------------------------
// Data model for Media
//
// This is the main struct for Media
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
