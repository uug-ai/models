package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Site struct {
	Id          primitive.ObjectID `json:"id" bson:"_id,omitempty,omitempty"`
	Name        string             `json:"name" bson:"name,omitempty"`
	Description string             `json:"description" bson:"description,omitempty"`

	// RBAC information
	OrganisationId string   `json:"organisationId" bson:"organisationId,omitempty"`
	Devices        []string `json:"devices" bson:"devices"`
	Groups         []string `json:"groups" bson:"groups"`

	// Media file information (by default "vault", however might change
	// in the future (integration with other storage solutions, next to Vault).
	StorageSolution string `json:"storageSolution,omitempty" bson:"storageSolution,omitempty"`

	VaultAccessKey string `json:"vaultAccessKey" bson:"vaultAccessKey"`
	VaultSecretKey string `json:"vaultSecretKey" bson:"vaultSecretKey"`
	VaultUri       string `json:"vaultUri" bson:"vaultUri"`

	// Metadata
	Metadata *SiteMetadata `json:"metadata,omitempty" bson:"metadata,omitempty"`

	// Audit information
	Audit *Audit `json:"audit,omitempty" bson:"audit,omitempty"`

	// Deprecated
	UserId     string   `json:"user_id" bson:"user_id,omitempty"`
	Initials   string   `json:"initials" bson:"initials,omitempty"`
	Color      string   `json:"color" bson:"color,omitempty"`
	Address    Location `json:"address" bson:"address,omitempty"`
	AllDevices []string `json:"all_devices" bson:"all_devices"`
	AccessKey  string   `json:"access_key" bson:"access_key"`
	SecretKey  string   `json:"secret_key" bson:"secret_key"`
	StorageUri string   `json:"storage_uri" bson:"storage_uri"`
}

// SiteMetadata contains additional metadata for the site, such as tags and classifications.
type SiteMetadata struct {
	Initials           string      `json:"initials" bson:"initials,omitempty"`
	Color              string      `json:"color" bson:"color,omitempty"`
	FloorPlans         []FloorPlan `json:"floorPlans" bson:"floorPlans,omitempty"` // List of floor plans associated with the site
	NumberOfFloorPlans int         `json:"numberOfFloorPlans" bson:"numberOfFloorPlans,omitempty"`
	Location           Location    `json:"location" bson:"location,omitempty"`
}

type SiteOption struct {
	Value string `json:"value" bson:"value"`
	Text  string `json:"text" bson:"text"`

	Devices []string `json:"devices" bson:"devices"`
	Groups  []string `json:"groups" bson:"groups"`

	Metadata *SiteMetadata `json:"metadata,omitempty" bson:"metadata,omitempty"`
}
