package models

// Timeline
type Timeline struct {
	Device  Device         `json:"device" bson:"device"`
	Media   *[]MediaGroup  `json:"media" bson:"media"`
	Markers *[]MarkerGroup `json:"markers" bson:"markers"`
}

type MediaTimeline struct {
	Device Device        `json:"device" bson:"device"`
	Media  *[]MediaGroup `json:"media" bson:"media"`
}

type MarkerTimeline struct {
	Device  Device         `json:"device" bson:"device"`
	Markers *[]MarkerGroup `json:"markers" bson:"markers"`
}

type MediaGroup struct {
	StartTimestamp int64   `json:"startTimestamp" bson:"startTimestamp"`
	EndTimestamp   int64   `json:"endTimestamp" bson:"endTimestamp"`
	Count          int64   `json:"count" bson:"count"`
	Media          []Media `json:"media" bson:"media"`
}

type MarkerGroup struct {
	StartTimestamp int64    `json:"startTimestamp" bson:"startTimestamp"`
	EndTimestamp   int64    `json:"endTimestamp" bson:"endTimestamp"`
	Count          int64    `json:"count" bson:"count"`
	Markers        []Marker `json:"markers" bson:"markers"`
}
