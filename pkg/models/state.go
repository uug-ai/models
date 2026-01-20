package models

import (
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// State enumeration
type StateEnum string

const (
	StateDefault     StateEnum = "default"      // Default state, no specific behavior, returns to default when no desired state is valid or available.
	StateDebug       StateEnum = "debug"        // Debug state, device operates in debug mode with verbose logging
	StatePaused      StateEnum = "paused"       // Paused state, device operations are temporarily halted
	StateNoRecording StateEnum = "no_recording" // No recording state, device does not record any media
	StateNoLiveView  StateEnum = "no_live_view" // No live view state, device does not provide live video feed
)

// State represents the state pattern used for a specific device.
// This allows to achieve specific behaviors or configurations based on the defined state.
type State struct {
	Id          primitive.ObjectID `json:"id" bson:"_id,omitempty"`                                                                    // Unique identifier for the state
	Name        string             `json:"name" bson:"name" example:"Default State" required:"true"`                                   // Name of the state
	Description string             `json:"description,omitempty" bson:"description,omitempty" example:"Person forcably opened a door"` // Description of the status

	// Resource identification, to which this state applies
	DeviceId       string   `json:"deviceId" bson:"deviceId" example:"686a906345c1df594939f9j25f4"`                                                         // DeviceId is used to identify the device associated with the state
	SiteId         string   `json:"siteId,omitempty" bson:"siteId,omitempty" example:"686a906345c1df594pcsr3r45"`                                           // SiteId is used to identify the site for which the state is relevant
	GroupId        string   `json:"groupId,omitempty" bson:"groupId,omitempty" example:"686a906345c1df594pmt41w4"`                                          // GroupId is used to identify the group for which the state is relevant
	OrganisationId string   `json:"organisationId" bson:"organisationId" example:"686a906345c1df594pad69f0"`                                                // OrganisationId is used to identify the organisation that owns the state.
	Devices        []string `json:"devices,omitempty" bson:"devices,omitempty" example:"[\"686a906345c1df594939f9j25f4\",\"686a906345c1df594939f9j25f5\"]"` // List of device IDs associated with the state

	// Timing information (all timestamps are in seconds)
	DesiredState               StateEnum `json:"desiredState,omitempty" bson:"desiredState,omitempty" example:"active"`                                 // The desired state to be applied to the device
	DesiredStateStartTimestamp int64     `json:"desiredStateStartTimestamp,omitempty" bson:"desiredStateStartTimestamp,omitempty" example:"1752482068"` // Timestamp when the desired state should start being applied
	DesiredStateEndTimestamp   int64     `json:"desiredStateEndTimestamp,omitempty" bson:"desiredStateEndTimestamp,omitempty" example:"1784018068"`     // Timestamp when the desired state should stop being applied

	// Conditions for the state to be applied
	TimeSchedule TimeSchedule `json:"timeSchedule,omitempty" bson:"timeSchedule,omitempty"` // Time schedule for when the state is active (if applicable)

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

// Check if the state needs to be considered ( time > desiredStateStartTimestamp and time < desiredStateEndTimestamp )
func (s *State) IsActiveAt(timestamp int64) bool {
	if s.DesiredStateStartTimestamp != 0 && timestamp < s.DesiredStateStartTimestamp {
		return false
	}
	if s.DesiredStateEndTimestamp != 0 && timestamp > s.DesiredStateEndTimestamp {
		return false
	}
	return true
}

type StateMetadata struct {
}

type StateAtRuntimeMetadata struct {
}

// TimeSchedule represents a cron-like schedule for when a state should be active.
// This allows defining recurring time windows using familiar cron patterns.
type TimeSchedule struct {
	Enabled    bool   `json:"enabled" bson:"enabled" example:"true"`                                   // Whether the time schedule is enabled
	Cron       string `json:"cron,omitempty" bson:"cron,omitempty" example:"0 9 * * 1-5"`              // Cron expression (minute, hour, day of month, month, day of week)
	Duration   int64  `json:"duration,omitempty" bson:"duration,omitempty" example:"28800"`            // Duration in seconds for how long the state remains active after trigger
	Timezone   string `json:"timezone,omitempty" bson:"timezone,omitempty" example:"Europe/Amsterdam"` // Timezone for the schedule (IANA format)
	StartDate  int64  `json:"startDate,omitempty" bson:"startDate,omitempty" example:"1752482068"`     // Optional start date (epoch seconds) from when the schedule is valid
	EndDate    int64  `json:"endDate,omitempty" bson:"endDate,omitempty" example:"1784018068"`         // Optional end date (epoch seconds) until when the schedule is valid
	DaysOfWeek []int  `json:"daysOfWeek,omitempty" bson:"daysOfWeek,omitempty" example:"[1,2,3,4,5]"`  // Days of week (0=Sunday, 1=Monday, ..., 6=Saturday) - alternative to cron
	StartTime  string `json:"startTime,omitempty" bson:"startTime,omitempty" example:"09:00"`          // Start time in HH:MM format - alternative to cron
	EndTime    string `json:"endTime,omitempty" bson:"endTime,omitempty" example:"17:00"`              // End time in HH:MM format - alternative to cron
}
