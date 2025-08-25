package models

type HighUpload struct {
	Requests       int64 `json:"requests,omitempty"`
	StartTimestamp int64 `json:"start_timestamp" bson:"start_timestamp,omitempty"`
	Notification   int64 `json:"notification,omitempty"`
}