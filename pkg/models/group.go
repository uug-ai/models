package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type GroupType struct {
	Id   primitive.ObjectID `json:"id" bson:"_id,omitempty,omitempty"`
	Name string             `json:"name" bson:"name,omitempty"`
}

type Group struct {
	Id          primitive.ObjectID `json:"id" bson:"_id,omitempty,omitempty"`
	Type        GroupType          `json:"type" bson:"type,omitempty"`
	Name        string             `json:"name" bson:"name,omitempty"`
	Description string             `json:"description" bson:"description,omitempty"`
	Initials    string             `json:"initials" bson:"initials,omitempty"`
	Color       string             `json:"color" bson:"color,omitempty"`

	// RBAC properties
	OrganisationId string `json:"organisationId" bson:"organisationId,omitempty"`
	UserId         string `json:"user_id" bson:"user_id,omitempty"`

	// Permissions and scope for group
	Devices []string `json:"devices" bson:"devices"`
	Groups  []string `json:"groups" bson:"groups"`

	// Location info for group (if applicable)
	Address Location `json:"address" bson:"address,omitempty"`
	Street  string   `json:"street" bson:"street"`
	Country string   `json:"country" bson:"country"`

	// Media file information (by default "vault", however might change
	// in the future (integration with other storage solutions, next to Vault).
	StorageSolution string `json:"storageSolution,omitempty" bson:"storageSolution,omitempty"`
	VaultAccessKey string `json:"vaultAccessKey" bson:"vaultAccessKey"`
	VaultSecretKey string `json:"vaultSecretKey" bson:"vaultSecretKey"`
	VaultUri       string `json:"vaultUri" bson:"vaultUri"`

	// @ Deprecated: should go into the Audit property
	CreatedBy   string `json:"created_by" bson:"created_by,omitempty"`
	CreatedTime int64  `json:"created_time" bson:"created_time,omitempty"`
	UpdatedBy   string `json:"updated_by" bson:"updated_by,omitempty"`
	UpdatedTime int64  `json:"updated_time" bson:"updated_time,omitempty"`

	// @ Deprecated: The idea is that we will create recursive groups instead of assigning sites to groups
	GroupType string   `json:"group_type" bson:"group_type,omitempty"`
	Sites     []string `json:"sites" bson:"sites"`
}


// GroupMetadata contains additional metadata for the group, such as tags and classifications.
type GroupMetadata struct {
	Initials           string      `json:"initials" bson:"initials,omitempty"`
	Color              string      `json:"color" bson:"color,omitempty"`
	FloorPlans         []FloorPlan `json:"floorPlans" bson:"floorPlans,omitempty"` // List of floor plans associated with the group
	NumberOfFloorPlans int         `json:"numberOfFloorPlans" bson:"numberOfFloorPlans,omitempty"`
	Location           Location    `json:"location" bson:"location,omitempty"`
}

type GroupOption struct {
	Value string `json:"value" bson:"value"`
	Text  string `json:"text" bson:"text"`

	Devices []string `json:"devices" bson:"devices"`
	Groups  []string `json:"groups" bson:"groups"`
	Sites   []string `json:"sites" bson:"sites"`
}
