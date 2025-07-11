package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Site struct {
	Id                 primitive.ObjectID `json:"id" bson:"_id,omitempty,omitempty"`
	Name               string             `json:"name" bson:"name,omitempty"`
	Username           string             `json:"username" bson:"username,omitempty"` // added by accident should be removed!
	Description        string             `json:"description" bson:"description,omitempty"`
	Initials           string             `json:"initials" bson:"initials,omitempty"`
	Color              string             `json:"color" bson:"color,omitempty"`
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

	// Audit information
	Audit *Audit `json:"audit,omitempty" bson:"audit,omitempty"`
}
