package models

// Media represents a media object.
// @Description Media model
type Media struct {
	// Time window of media file.
	StartTimestamp int64 `json:"startTimestamp,omitempty" bson:"startTimestamp,omitempty"`
	EndTimestamp   int64 `json:"endTimestamp,omitempty" bson:"endTimestamp,omitempty"`
	Duration       int64 `json:"duration,omitempty" bson:"duration,omitempty"`

	// Media file owner
	DeviceId       string `json:"deviceId,omitempty" bson:"deviceId,omitempty"`
	GroupId        string `json:"groupId,omitempty" bson:"groupId,omitempty"`
	UserId         string `json:"userId,omitempty" bson:"userId,omitempty"`
	OrganisationId string `json:"organisationId,omitempty" bson:"organisationId,omitempty"`

	// Media file information (by default Vault (=kstorage), however might change
	// in the future (integration with other storage solutions, next to Vault).
	StorageSolution string `json:"storageSolution,omitempty" bson:"storageSolution,omitempty"`

	// Media file information
	VideoFile         string `json:"videoFile,omitempty" bson:"videoFile,omitempty"`
	VideoUrl          string `json:"videoUrl,omitempty" bson:"videoUrl,omitempty"`
	VideoProvider     string `json:"videoProvider,omitempty" bson:"videoProvider,omitempty"`
	ThumbnailUrl      string `json:"thumbnailUrl,omitempty" bson:"thumbnailUrl,omitempty"`
	ThumbnailFile     string `json:"thumbnailFile,omitempty" bson:"thumbnailFile,omitempty"`
	ThumbnailProvider string `json:"thumbnailProvider,omitempty" bson:"thumbnailProvider,omitempty"`
	SpriteUrl         string `json:"spriteUrl,omitempty" bson:"spriteUrl,omitempty"`
	SpriteFile        string `json:"spriteFile,omitempty" bson:"spriteFile,omitempty"`
	SpriteProvider    string `json:"spriteProvider,omitempty" bson:"spriteProvider,omitempty"`
	SpriteInterval    int    `json:"spriteInterval,omitempty" bson:"spriteInterval,omitempty"`

	// Metadata
	Metadata []MediaMetadata `json:"metadata,omitempty" bson:"metadata,omitempty"`
}

// We can store additional metadata for media files, such as tags and classifications.
type MediaMetadata struct {
	Tags            []string `json:"tags,omitempty" bson:"tags,omitempty"`
	Classifications []string `json:"classifications,omitempty" bson:"classifications,omitempty"`
}
