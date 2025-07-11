package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Role struct {
	Id                 primitive.ObjectID `json:"id" bson:"_id,omitempty,omitempty"`
	Name               string             `json:"roleName" bson:"roleName,omitempty"`
	ParentRole         string             `json:"role" bson:"role,omitempty"`
	Pages              []string           `json:"pages" bson:"pages"`
	TimeWindow         TimeWindow         `json:"timeWindow" bson:"timeWindow"`
	UserId             string             `json:"user_id" bson:"user_id,omitempty"`
	IsActive           int                `json:"isActive" bson:"isActive"`
	FeaturePermissions FeaturePermissions `json:"featurePermissions" bson:"featurePermissions"`
	TimeWindowActive   int                `json:"timeWindowActive" bson:"timeWindowActive"`
}

type FeaturePermissions struct {
	PTZ          int `json:"ptz" bson:"ptz"`
	Liveview     int `json:"liveview" bson:"liveview"`
	RemoteConfig int `json:"remote_config" bson:"remote_config"`
	IO           int `json:"io" bson:"io"`
	FloorPlans   int `json:"floorPlans" bson:"floorPlans"`
	// Talk..
	// ...
}

type TimeWindow struct {
	TimeRange1MinMonday    float64 `json:"timeRange1MinMonday" bson:"timeRange1MinMonday"`
	TimeRange1MaxMonday    float64 `json:"timeRange1MaxMonday" bson:"timeRange1MaxMonday"`
	TimeRange2MinMonday    float64 `json:"timeRange2MinMonday" bson:"timeRange2MinMonday"`
	TimeRange2MaxMonday    float64 `json:"timeRange2MaxMonday" bson:"timeRange2MaxMonday"`
	TimeRange1MinTuesday   float64 `json:"timeRange1MinTuesday" bson:"timeRange1MinTuesday"`
	TimeRange1MaxTuesday   float64 `json:"timeRange1MaxTuesday" bson:"timeRange1MaxTuesday"`
	TimeRange2MinTuesday   float64 `json:"timeRange2MinTuesday" bson:"timeRange2MinTuesday"`
	TimeRange2MaxTuesday   float64 `json:"timeRange2MaxTuesday" bson:"timeRange2MaxTuesday"`
	TimeRange1MinWednesday float64 `json:"timeRange1MinWednesday" bson:"timeRange1MinWednesday"`
	TimeRange1MaxWednesday float64 `json:"timeRange1MaxWednesday" bson:"timeRange1MaxWednesday"`
	TimeRange2MinWednesday float64 `json:"timeRange2MinWednesday" bson:"timeRange2MinWednesday"`
	TimeRange2MaxWednesday float64 `json:"timeRange2MaxWednesday" bson:"timeRange2MaxWednesday"`
	TimeRange1MinThursday  float64 `json:"timeRange1MinThursday" bson:"timeRange1MinThursday"`
	TimeRange1MaxThursday  float64 `json:"timeRange1MaxThursday" bson:"timeRange1MaxThursday"`
	TimeRange2MinThursday  float64 `json:"timeRange2MinThursday" bson:"timeRange2MinThursday"`
	TimeRange2MaxThursday  float64 `json:"timeRange2MaxThursday" bson:"timeRange2MaxThursday"`
	TimeRange1MinFriday    float64 `json:"timeRange1MinFriday" bson:"timeRange1MinFriday"`
	TimeRange1MaxFriday    float64 `json:"timeRange1MaxFriday" bson:"timeRange1MaxFriday"`
	TimeRange2MinFriday    float64 `json:"timeRange2MinFriday" bson:"timeRange2MinFriday"`
	TimeRange2MaxFriday    float64 `json:"timeRange2MaxFriday" bson:"timeRange2MaxFriday"`
	TimeRange1MinSaturday  float64 `json:"timeRange1MinSaturday" bson:"timeRange1MinSaturday"`
	TimeRange1MaxSaturday  float64 `json:"timeRange1MaxSaturday" bson:"timeRange1MaxSaturday"`
	TimeRange2MinSaturday  float64 `json:"timeRange2MinSaturday" bson:"timeRange2MinSaturday"`
	TimeRange2MaxSaturday  float64 `json:"timeRange2MaxSaturday" bson:"timeRange2MaxSaturday"`
	TimeRange1MinSunday    float64 `json:"timeRange1MinSunday" bson:"timeRange1MinSunday"`
	TimeRange1MaxSunday    float64 `json:"timeRange1MaxSunday" bson:"timeRange1MaxSunday"`
	TimeRange2MinSunday    float64 `json:"timeRange2MinSunday" bson:"timeRange2MinSunday"`
	TimeRange2MaxSunday    float64 `json:"timeRange2MaxSunday" bson:"timeRange2MaxSunday"`
}
