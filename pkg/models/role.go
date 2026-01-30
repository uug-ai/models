package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AccessLevel defines the level of access for a feature or resource
type AccessLevel int

const (
	AccessLevelNone  AccessLevel = 0 // No access
	AccessLevelRead  AccessLevel = 1 // Read-only access
	AccessLevelWrite AccessLevel = 2 // Read and write access
	AccessLevelAdmin AccessLevel = 3 // Full administrative access
)

type Role struct {
	Id                 primitive.ObjectID `json:"id" bson:"_id,omitempty,omitempty"`
	OrganisationId     primitive.ObjectID `json:"organisationId" bson:"organisationId,omitempty"` // Organisation this role belongs to
	Name               string             `json:"roleName" bson:"roleName,omitempty"`
	Description        string             `json:"description" bson:"description,omitempty"`
	Pages              []string           `json:"pages" bson:"pages"`
	TimeWindow         TimeWindow         `json:"timeWindow" bson:"timeWindow"`
	IsActive           int                `json:"isActive" bson:"isActive"`
	FeaturePermissions FeaturePermissions `json:"featurePermissions" bson:"featurePermissions"`
	TimeWindowActive   int                `json:"timeWindowActive" bson:"timeWindowActive"`
	Scope              RoleScope          `json:"scope" bson:"scope,omitempty"` // Optional granular scope within organisation
	Audit              Audit              `json:"audit" bson:"audit,omitempty"`
}

// RoleScope defines the scope/context where the role assignment applies.
// This allows for granular role assignments at different levels.
type RoleScope struct {
	Type      string   `json:"type" bson:"type,omitempty"`           // e.g., "global", "site", "group", "device"
	SiteIds   []string `json:"siteIds" bson:"siteIds,omitempty"`     // Sites where the role applies
	GroupIds  []string `json:"groupIds" bson:"groupIds,omitempty"`   // Groups where the role applies
	DeviceIds []string `json:"deviceIds" bson:"deviceIds,omitempty"` // Devices where the role applies
}

type FeaturePermissions struct {
	PTZ          AccessLevel `json:"ptz" bson:"ptz"`                   // 0=none, 1=read, 2=write, 3=admin
	Liveview     AccessLevel `json:"liveview" bson:"liveview"`         // 0=none, 1=read, 2=write, 3=admin
	RemoteConfig AccessLevel `json:"remoteConfig" bson:"remoteConfig"` // 0=none, 1=read, 2=write, 3=admin
	IO           AccessLevel `json:"io" bson:"io"`                     // 0=none, 1=read, 2=write, 3=admin
	FloorPlans   AccessLevel `json:"floorPlans" bson:"floorPlans"`     // 0=none, 1=read, 2=write, 3=admin
	Playback     AccessLevel `json:"playback" bson:"playback"`         // 0=none, 1=read, 2=write, 3=admin
	Export       AccessLevel `json:"export" bson:"export"`             // 0=none, 1=read, 2=write, 3=admin
	Markers      AccessLevel `json:"markers" bson:"markers"`           // 0=none, 1=read, 2=write, 3=admin
	Alerts       AccessLevel `json:"alerts" bson:"alerts"`             // 0=none, 1=read, 2=write, 3=admin
	Users        AccessLevel `json:"users" bson:"users"`               // 0=none, 1=read, 2=write, 3=admin
	Devices      AccessLevel `json:"devices" bson:"devices"`           // 0=none, 1=read, 2=write, 3=admin
	Sites        AccessLevel `json:"sites" bson:"sites"`               // 0=none, 1=read, 2=write, 3=admin
	Groups       AccessLevel `json:"groups" bson:"groups"`             // 0=none, 1=read, 2=write, 3=admin
	Roles        AccessLevel `json:"roles" bson:"roles"`               // 0=none, 1=read, 2=write, 3=admin
	Settings     AccessLevel `json:"settings" bson:"settings"`         // 0=none, 1=read, 2=write, 3=admin
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
