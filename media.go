package main

type Media struct {
	FileName       string `json:"fileName,omitempty" bson:"fileName,omitempty"`
	StartTimestamp int64  `json:"startTimestamp,omitempty" bson:"startTimestamp,omitempty"`
	EndTimestamp   int64  `json:"endTimestamp,omitempty" bson:"endTimestamp,omitempty"`
	Duration       int64  `json:"duration,omitempty" bson:"duration,omitempty"`
	Provider       string `json:"provider,omitempty" bson:"provider,omitempty"`
	Storage        string `json:"storage,omitempty" bson:"storage,omitempty"`
	DeviceId       string `json:"deviceId,omitempty" bson:"deviceId,omitempty"`
	GroupId        string `json:"groupId,omitempty" bson:"groupId,omitempty"`
	UserId         string `json:"userId,omitempty" bson:"userId,omitempty"`

	VideoUrl          string `json:"video_url,omitempty" bson:"video_url,omitempty"`
	ThumbnailUrl      string `json:"thumbnail_url,omitempty" bson:"thumbnail_url,omitempty"`
	ThumbnailFile     string `json:"thumbnailFile,omitempty" bson:"thumbnailFile,omitempty"`
	ThumbnailProvider string `json:"thumbnailProvider,omitempty" bson:"thumbnailProvider,omitempty"`
	SpriteUrl         string `json:"sprite_url,omitempty" bson:"sprite_url,omitempty"`
	SpriteFile        string `json:"spriteFile,omitempty" bson:"spriteFile,omitempty"`
	SpriteProvider    string `json:"spriteProvider,omitempty" bson:"spriteProvider,omitempty"`
	SpriteInterval    int    `json:"spriteInterval,omitempty" bson:"spriteInterval,omitempty"`
}
