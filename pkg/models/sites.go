package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Site struct {
	Id          primitive.ObjectID `json:"id" bson:"_id,omitempty,omitempty"`
	Name        string             `json:"name" bson:"name,omitempty"`
	Username    string             `json:"username" bson:"username,omitempty"` // added by accident should be removed!
	Description string             `json:"description" bson:"description,omitempty"`
	Initials    string             `json:"initials" bson:"initials,omitempty"`
	Color       string             `json:"color" bson:"color,omitempty"`

	UserId     string   `json:"user_id" bson:"user_id,omitempty"`
	AllDevices []string `json:"all_devices" bson:"all_devices"` // Calculated on the fly!
	Devices    []string `json:"devices" bson:"devices"`
	Groups     []string `json:"groups" bson:"groups"`

	AccessKey  string `json:"access_key" bson:"access_key"`
	SecretKey  string `json:"secret_key" bson:"secret_key"`
	StorageUri string `json:"storage_uri" bson:"storage_uri"`

	FloorPlans         []FloorPlan `json:"floor_plans" bson:"floor_plans,omitempty"`
	NumberOfFloorPlans int         `json:"numberOfFloorPlans" bson:"numberOfFloorPlans,omitempty"`

	// Location metadata
	LocationMetadata *SiteLocationMetadata `json:"locationMetadata,omitempty" bson:"locationMetadata,omitempty"`

	// Audit information
	Audit *Audit `json:"audit,omitempty" bson:"audit,omitempty"`
}

// SiteLocationMetadata contains metadata about the physical location of the site.
type SiteLocationMetadata struct {
	Location Location `json:"location" bson:"location,omitempty"`
}
