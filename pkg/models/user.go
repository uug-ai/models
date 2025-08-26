package models

import "go.mongodb.org/mongo-driver/bson/primitive"

const (
	USER_FOUND     string = "One or more users were found"
	USER_NOT_FOUND string = "One or more users not found, returning empty list"
)

type User struct {
	Id                    primitive.ObjectID                `json:"id" bson:"_id,omitempty"`
	Username              string                            `json:"username,omitempty" bson:"username,omitempty"`
	Email                 string                            `json:"email,omitempty" bson:"email,omitempty"`
	Error                 bool                              `json:"error,omitempty" bson:"error,omitempty"`
	ReachedLimit          bool                              `json:"reachedLimit" bson:"reachedLimit,omitempty"`
	ReachedLimitTimestamp int64                             `json:"reachedLimitTimestamp" bson:"reachedLimitTimestamp,omitempty"`
	Timezone              string                            `json:"timezone,omitempty" bson:"timezone,omitempty"`
	Dates                 []string                          `json:"dates,omitempty" bson:"dates,omitempty"`
	Instances             []string                          `json:"instances,omitempty" bson:"instances,omitempty"`
	PublicKey             string                            `json:"amazon_access_key_id" bson:"amazon_access_key_id,omitempty"`
	PrivateKey            string                            `json:"amazon_secret_access_key" bson:"amazon_secret_access_key,omitempty"`
	Pushbullet            string                            `json:"pushbullet_api_key" bson:"pushbullet_api_key,omitempty"`
	Settings              map[string]interface{}            `json:"settings,omitempty" bson:"settings,omitempty"`
	Throttler             map[string]interface{}            `json:"throttler,omitempty" bson:"throttler,omitempty"`
	Activity              []map[string]interface{}          `json:"activity,omitempty" bson:"activity,omitempty"`
	HighUpload            HighUpload                        `json:"highupload,omitempty" bson:"highupload,omitempty"`
	Devices               []map[string]interface{}          `json:"devices,omitempty" bson:"devices,omitempty"`
	NotificationSettings  map[string]map[string]interface{} `json:"notificationSettings" bson:"notificationSettings,omitempty"`
	Channels              map[string]map[string]interface{} `json:"channels,omitempty" bson:"channels,omitempty"`
	Storage               Storage                           `json:"storage,omitempty" bson:"storage,omitempty"`

	// We can override the subscription settings if needed.
	CustomUsageLimit    int `json:"custom_usage_limit" bson:"custom_usage_limit,omitempty"`
	CustomDayLimit      int `json:"custom_day_limit" bson:"custom_day_limit,omitempty"`
	CustomAnalysisLimit int `json:"custom_analysis_limit" bson:"custom_analysis_limit,omitempty"`
}
