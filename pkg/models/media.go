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

type MediaShort struct {
	Key               string `json:"key" bson:"key"`
	Path              string `json:"path" bson:"path"`
	Url               string `json:"src" bson:"src"`
	Timestamp         int64  `json:"timestamp" bson:"timestamp"`
	CameraId          string `json:"camera_id" bson:"camera_id"`
	Date              string `json:"date" bson:"date"`
	Time              string `json:"time" bson:"time"`
	Type              string `json:"type" bson:"type"`
	Provider          string `json:"provider" bson:"provider"`
	Source            string `json:"source" bson:"source"`
	SpriteUrl         string `json:"sprite_url" bson:"sprite_url"`
	ThumbnailUrl      string `json:"thumbnail_url" bson:"thumbnail_url"`
	ThumbnailFile     string `json:"thumbnailFile" bson:"thumbnailFile"`
	ThumbnailProvider string `json:"thumbnailProvider" bson:"thumbnailProvider"`
	SpriteFile        string `json:"spriteFile" bson:"spriteFile"`
	SpriteProvider    string `json:"spriteProvider" bson:"spriteProvider"`
	SpriteInterval    int    `json:"spriteInterval" bson:"spriteInterval"`
}

// TimelineValue represents timeline state information.
// @Description Timeline value model
type TimelineValue struct {
	TimelineZero int64 `json:"timelineZero" bson:"timelineZero"`
	StartTime    int64 `json:"startTime" bson:"startTime"`
	CurrentTime  int64 `json:"currentTime" bson:"currentTime"`
	CurrentScale int64 `json:"currentScale" bson:"currentScale"`
}
