package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var SITE_ID_REQUIRED = "SITE_ID_REQUIRED"
var SITE_UPDATE_FAILED = "SITE_UPDATE_FAILED"
var SITE_UPDATE_SUCCESS = "SITE_UPDATE_SUCCESS"
var SITE_CREATE_FAILED = "SITE_CREATE_FAILED"
var SITE_CREATE_SUCCESS = "SITE_CREATE_SUCCESS"
var SITE_DELETE_FAILED = "SITE_DELETE_FAILED"
var SITE_DELETE_SUCCESS = "SITE_DELETE_SUCCESS"
var SITE_GET_ALL_FAILED = "SITE_GET_ALL_FAILED"
var SITE_GET_ALL_SUCCESS = "SITE_GET_ALL_SUCCESS"
var SITE_GET_FAILED = "SITE_GET_FAILED"
var SITE_GET_SUCCESS = "SITE_GET_SUCCESS"

var FLOOR_PLAN_ID_REQUIRED = "FLOOR_PLAN_ID_REQUIRED"
var FLOOR_PLAN_UPDATE_FAILED = "FLOOR_PLAN_UPDATE_FAILED"
var FLOOR_PLAN_UPDATE_SUCCESS = "FLOOR_PLAN_UPDATE_SUCCESS"
var FLOOR_PLAN_GET_FAILED = "FLOOR_PLAN_GET_FAILED"
var FLOOR_PLAN_GET_SUCCESS = "FLOOR_PLAN_GET_SUCCESS"
var FLOOR_PLAN_CREATE_FAILED = "FLOOR_PLAN_CREATE_FAILED"
var FLOOR_PLAN_CREATE_SUCCESS = "FLOOR_PLAN_CREATE_SUCCESS"
var FLOOR_PLAN_DELETE_FAILED = "FLOOR_PLAN_DELETE_FAILED"
var FLOOR_PLAN_DELETE_SUCCESS = "FLOOR_PLAN_DELETE_SUCCESS"
var DEVICE_KEY_REQUIRED = "DEVICE_KEY_REQUIRED"

type SiteWrapper struct {
	Site Site `json:"site" bson:"site"`
}

type Site struct {
	Id                 primitive.ObjectID `json:"id" bson:"_id,omitempty,omitempty"`
	Name               string             `json:"name" bson:"name,omitempty"`
	Username           string             `json:"username" bson:"username,omitempty"` // added by accident should be removed!
	Description        string             `json:"description" bson:"description,omitempty"`
	Initials           string             `json:"initials" bson:"initials,omitempty"`
	Color              string             `json:"color" bson:"color,omitempty"`
	Address            Location           `json:"address" bson:"address,omitempty"`
	UserId             string             `json:"user_id" bson:"user_id,omitempty"`
	AllDevices         []string           `json:"all_devices" bson:"all_devices"` // Calculated on the fly!
	Devices            []string           `json:"devices" bson:"devices"`
	Street             string             `json:"street" bson:"street"`
	Country            string             `json:"country" bson:"country"`
	Groups             []string           `json:"groups" bson:"groups"`
	AccessKey          string             `json:"access_key" bson:"access_key"`
	SecretKey          string             `json:"secret_key" bson:"secret_key"`
	StorageUri         string             `json:"storage_uri" bson:"storage_uri"`
	CreatedBy          string             `json:"created_by" bson:"created_by,omitempty"`
	CreatedTime        int64              `json:"created_time" bson:"created_time,omitempty"`
	UpdatedBy          string             `json:"updated_by" bson:"updated_by,omitempty"`
	UpdatedTime        int64              `json:"updated_time" bson:"updated_time,omitempty"`
	FloorPlans         []FloorPlan        `json:"floor_plans" bson:"floor_plans,omitempty"`
	NumberOfFloorPlans int                `json:"numberOfFloorPlans" bson:"numberOfFloorPlans,omitempty"`
}

type SiteShort struct {
	Id       primitive.ObjectID `json:"id" bson:"_id,omitempty,omitempty"`
	Name     string             `json:"name" bson:"name,omitempty"`
	Initials string             `json:"initials" bson:"initials,omitempty"`
	Color    string             `json:"color" bson:"color,omitempty"`
	Address  Location           `json:"address" bson:"address,omitempty"`
}

type FloorPlan struct {
	Id          primitive.ObjectID `json:"id" bson:"_id,omitempty,omitempty"`
	Name        string             `json:"name" bson:"name,omitempty"`
	Image       string             `json:"image" bson:"image,omitempty"`
	Width       int                `json:"width" bson:"width,omitempty"`
	Height      int                `json:"height" bson:"height,omitempty"`
	CreatedBy   string             `json:"created_by" bson:"created_by,omitempty"`
	CreatedTime int64              `json:"created_time" bson:"created_time,omitempty"`
	UpdatedBy   string             `json:"updated_by" bson:"updated_by,omitempty"`
	UpdatedTime int64              `json:"updated_time" bson:"updated_time,omitempty"`
	Devices     []DevicePlacement  `json:"devices" bson:"devices,omitempty"`
}

type DevicePlacement struct {
	Id               primitive.ObjectID `json:"id" bson:"_id"`
	DeviceKey        string             `json:"deviceKey" bson:"deviceKey"`
	DeviceName       string             `json:"deviceName" bson:"deviceName"`
	FabricKey        string             `json:"fabricKey" bson:"fabricKey"`
	X                float64            `json:"x" bson:"x"`                 // Absolute X coordinate
	Y                float64            `json:"y" bson:"y"`                 // Absolute Y coordinate
	RelativeX        float64            `json:"relativeX" bson:"relativeX"` // X relative to canvas width (0 to 1)
	RelativeY        float64            `json:"relativeY" bson:"relativeY"` // Y relative to canvas height (0 to 1)
	Radius           float64            `json:"radius" bson:"radius,omitempty"`
	FieldOfView      float64            `json:"fieldOfView" bson:"fieldOfView,omitempty"`
	Color            string             `json:"color" bson:"color,omitempty"`
	DeviceStatus     string             `json:"deviceStatus" bson:"deviceStatus,omitempty"`
	SliceStartAngle  float64            `json:"sliceStartAngle" bson:"sliceStartAngle,omitempty"`
	SliceEndAngle    float64            `json:"sliceEndAngle" bson:"sliceEndAngle,omitempty"`
	SliceMiddleAngle float64            `json:"sliceMiddleAngle" bson:"sliceMiddleAngle,omitempty"`
	CreatedBy        string             `json:"createdBy" bson:"createdBy,omitempty"`
	CreatedTime      int64              `json:"createdTime" bson:"createdTime,omitempty"`
	UpdatedBy        string             `json:"updatedBy" bson:"updatedBy,omitempty"`
	UpdatedTime      int64              `json:"updatedTime" bson:"updatedTime,omitempty"`
}
