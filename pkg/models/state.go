package models

import (
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// State represents the state pattern used for a specific device.
// This allows to achieve specific behaviors or configurations based on the defined state.
type State struct {
	Id          primitive.ObjectID `json:"id" bson:"_id,omitempty"`                                                                    // Unique identifier for the state
	Name        string             `json:"name" bson:"name" example:"Default State" required:"true"`                                   // Name of the state
	Description string             `json:"description,omitempty" bson:"description,omitempty" example:"Person forcably opened a door"` // Description of the status

	// Resource identification
	DeviceId       string   `json:"deviceId" bson:"deviceId" example:"686a906345c1df594939f9j25f4"`                                                         // DeviceId is used to identify the device associated with the state
	SiteId         string   `json:"siteId,omitempty" bson:"siteId,omitempty" example:"686a906345c1df594pcsr3r45"`                                           // SiteId is used to identify the site for which the state is relevant
	GroupId        string   `json:"groupId,omitempty" bson:"groupId,omitempty" example:"686a906345c1df594pmt41w4"`                                          // GroupId is used to identify the group for which the state is relevant
	OrganisationId string   `json:"organisationId" bson:"organisationId" example:"686a906345c1df594pad69f0"`                                                // OrganisationId is used to identify the organisation that owns the state.
	Devices        []string `json:"devices,omitempty" bson:"devices,omitempty" example:"[\"686a906345c1df594939f9j25f4\",\"686a906345c1df594939f9j25f5\"]"` // List of device IDs associated with the state

	// Timing information (all timestamps are in seconds)
	StartTimestamp int64  `json:"startTimestamp" bson:"startTimestamp" example:"1752482068" required:"true"` // Start timestamp of the marker in seconds since epoch
	EndTimestamp   int64  `json:"endTimestamp" bson:"endTimestamp" example:"1752482079" required:"true"`     // End timestamp of the marker in seconds since epoch
	Duration       int    `json:"duration" bson:"duration" example:"11" required:"true"`                     // Duration of the state in seconds
	State          string `json:"state" bson:"state" example:"active" required:"true"`                       // Current state, e.g., "active", "inactive"

	// Additional metadata
	Metadata *StateMetadata `json:"metadata,omitempty" bson:"metadata,omitempty"` // Metadata associated with the marker, such as comments and tags

	// AtRuntimeMetadata contains metadata that is generated at runtime, which can include
	// more verbose information about the device's current state, capabilities, or configuration.
	// for example the linked sites details, etc.
	AtRuntimeMetadata *StateAtRuntimeMetadata `json:"atRuntimeMetadata,omitempty" bson:"atRuntimeMetadata,omitempty"`

	// Audit information
	Audit *Audit `json:"audit,omitempty" bson:"audit,omitempty"` // Audit information for tracking changes to the marker
}

func (s *State) Validate() error {
	validate := validator.New()
	return validate.Struct(s)
}

type StateMetadata struct {
}

type StateAtRuntimeMetadata struct {
}
