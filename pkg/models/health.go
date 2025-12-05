package models


type Health struct {

	Description string `json:"description,omitempty" bson:"description,omitempty"`
	Database	string `json:"database,omitempty" bson:"database,omitempty"`
	Queue		string `json:"queue,omitempty" bson:"queue,omitempty"`
	License		string `json:"license,omitempty" bson:"license,omitempty"`

	// Additional metadata
	Metadata *HealthMetadata `json:"metadata,omitempty" bson:"metadata,omitempty"` // Metadata associated with the health, such as comments and tags

	// AtRuntimeMetadata contains metadata that is generated at runtime, which can include
	// more verbose information about the device's current state, capabilities, or configuration.
	// for example the linked sites details, etc.
	AtRuntimeMetadata *HealthAtRuntimeMetadata `json:"atRuntimeMetadata,omitempty" bson:"atRuntimeMetadata,omitempty"`

	// Synchronize
	Synchronize *Synchronize `json:"synchronize,omitempty" bson:"synchronize,omitempty"` // Synchronization status with external systems

	// Audit information
	Audit *Audit `json:"audit,omitempty" bson:"audit,omitempty"` // Audit information for tracking changes to the marker
}

type HealthMetadata struct {
	LicenseExpiryDate string `json:"licenseExpiryDate,omitempty" bson:"licenseExpiryDate,omitempty"`
}

type HealthAtRuntimeMetadata struct {
	
}