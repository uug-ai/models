package models

type Activity struct {
	Day       string           `json:"day,omitempty" bson:"day,omitempty"`
	Requests  int64            `json:"requests,omitempty" bson:"requests,omitempty"`
	Videos    int64            `json:"videos,omitempty" bson:"videos,omitempty"`
	Images    int64            `json:"images,omitempty" bson:"images,omitempty"`
	Usage     int64            `json:"usage,omitempty" bson:"usage,omitempty"`
	Timestamp int64            `json:"timestamp,omitempty" bson:"timestamp,omitempty"`
	Devices   map[string]int64 `json:"devices,omitempty" bson:"devices,omitempty"`
}
