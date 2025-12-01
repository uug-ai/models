package models

// Strategy represents the strategy pattern used for a specific device.
// This allows to achieve specific behaviors or configurations based on the defined strategy.
type Strategy struct {
	Id          string `json:"id" bson:"_id" example:"strategy-123" required:"true"`                                       // Unique identifier for the strategy
	Name        string `json:"name" bson:"name" example:"Default Strategy" required:"true"`                                // Name of the strategy
	Description string `json:"description,omitempty" bson:"description,omitempty" example:"Person forcably opened a door"` // Description of the marker

	// RBAC information
	DeviceId       string `json:"deviceId" bson:"deviceId" example:"686a906345c1df594939f9j25f4" required:"true"` // DeviceId is used to identify the device associated with the marker
	SiteId         string `json:"siteId,omitempty" bson:"siteId,omitempty" example:"686a906345c1df594pcsr3r45"`   // SiteId is used to identify the site where the marker is located
	GroupId        string `json:"groupId,omitempty" bson:"groupId,omitempty" example:"686a906345c1df594pmt41w4"`  // GroupId is used to identify the group of markers
	OrganisationId string `json:"organisationId" bson:"organisationId" example:"686a906345c1df594pad69f0"`        // OrganisationId is used to identify the organisation that owns the marker, retrieved from the user's access token

	// Timing information (all timestamps are in seconds)
	StartTimestamp int64 `json:"startTimestamp" bson:"startTimestamp" example:"1752482068" required:"true"` // Start timestamp of the marker in seconds since epoch
	EndTimestamp   int64 `json:"endTimestamp" bson:"endTimestamp" example:"1752482079" required:"true"`     // End timestamp of the marker in seconds since epoch
	Duration       int   `json:"duration" bson:"duration" example:"11" required:"true"`                     // Duration of the strategy in seconds
	Active         bool  `json:"active" bson:"active" example:"true" required:"true"`                       // Indicates if the strategy is currently active

	// Additional metadata
	Metadata *StrategyMetadata `json:"metadata,omitempty" bson:"metadata,omitempty"` // Metadata associated with the marker, such as comments and tags

	// AtRuntimeMetadata contains metadata that is generated at runtime, which can include
	// more verbose information about the device's current state, capabilities, or configuration.
	// for example the linked sites details, etc.
	AtRuntimeMetadata *StrategyAtRuntimeMetadata `json:"atRuntimeMetadata,omitempty" bson:"atRuntimeMetadata,omitempty"`

	// Audit information
	Audit *Audit `json:"audit,omitempty" bson:"audit,omitempty"` // Audit information for tracking changes to the marker

}

type StrategyMetadata struct {
}

type StrategyAtRuntimeMetadata struct {
}
