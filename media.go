package main

type Media struct {
	// Time window of media file.
	StartTimestamp int64 `json:"startTimestamp,omitempty" bson:"startTimestamp,omitempty"`
	EndTimestamp   int64 `json:"endTimestamp,omitempty" bson:"endTimestamp,omitempty"`
	Duration       int64 `json:"duration,omitempty" bson:"duration,omitempty"`

	// Media file owner
	DeviceId string `json:"deviceId,omitempty" bson:"deviceId,omitempty"`
	GroupId  string `json:"groupId,omitempty" bson:"groupId,omitempty"`
	UserId   string `json:"userId,omitempty" bson:"userId,omitempty"`

	// Media file information
	Storage           string `json:"storage,omitempty" bson:"storage,omitempty"`
	VideoFile         string `json:"videoFile,omitempty" bson:"videoFile,omitempty"`
	VideoUrl          string `json:"video_url,omitempty" bson:"video_url,omitempty"`
	VideoProvider     string `json:"videoProvider,omitempty" bson:"videoProvider,omitempty"`
	ThumbnailUrl      string `json:"thumbnail_url,omitempty" bson:"thumbnail_url,omitempty"`
	ThumbnailFile     string `json:"thumbnailFile,omitempty" bson:"thumbnailFile,omitempty"`
	ThumbnailProvider string `json:"thumbnailProvider,omitempty" bson:"thumbnailProvider,omitempty"`
	SpriteUrl         string `json:"sprite_url,omitempty" bson:"sprite_url,omitempty"`
	SpriteFile        string `json:"spriteFile,omitempty" bson:"spriteFile,omitempty"`
	SpriteProvider    string `json:"spriteProvider,omitempty" bson:"spriteProvider,omitempty"`
	SpriteInterval    int    `json:"spriteInterval,omitempty" bson:"spriteInterval,omitempty"`

	// Metadata
	Metadata []MediaMetadata `json:"metadata,omitempty" bson:"metadata,omitempty"`
}

type MediaMetadata struct {
	Tags            []string `json:"tags,omitempty" bson:"tags,omitempty"`
	Classifications []string `json:"classifications,omitempty" bson:"classifications,omitempty"`
}
