package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Group struct {
	Id          primitive.ObjectID `json:"id" bson:"_id,omitempty,omitempty"`
	Name        string             `json:"name" bson:"name,omitempty"`
	Description string             `json:"description" bson:"description,omitempty"`

	// RBAC information
	OrganisationId string   `json:"organisationId" bson:"organisationId,omitempty"`
	Devices        []string `json:"devices" bson:"devices"`
	Groups         []string `json:"groups" bson:"groups"` // Nested groups

	// Media file information (by default "vault", however might change
	// in the future (integration with other storage solutions, next to Vault).
	StorageSolution string `json:"storageSolution,omitempty" bson:"storageSolution,omitempty"`

	VaultAccessKey string `json:"vaultAccessKey" bson:"vaultAccessKey"`
	VaultSecretKey string `json:"vaultSecretKey" bson:"vaultSecretKey"`
	VaultUri       string `json:"vaultUri" bson:"vaultUri"`

	// Metadata
	Metadata *GroupMetadata `json:"metadata,omitempty" bson:"metadata,omitempty"`

	// Audit information
	Audit *Audit `json:"audit,omitempty" bson:"audit,omitempty"`
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
}
