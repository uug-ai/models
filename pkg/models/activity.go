package models

type Activity struct {
	Day       string           `json:"day,omitempty"`
	Requests  int64            `json:"requests,omitempty"`
	Videos    int64            `json:"videos,omitempty"`
	Images    int64            `json:"images,omitempty"`
	Usage     int64            `json:"usage,omitempty"`
	Timestamp int64            `json:"timestamp,omitempty"`
	Devices   map[string]int64 `json:"devices,omitempty"`
}