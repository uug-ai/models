package models

type Status string

const (
	SynchronizeStatusPending    Status = "pending"
	SynchronizeStatusInProgress Status = "inProgress"
	SynchronizeStatusSkipped    Status = "skipped"
	SynchronizeStatusCompleted  Status = "completed"
	SynchronizeStatusFailed     Status = "failed"
)

type Synchronize struct {
	Status   Status              `json:"status,omitempty" bson:"status,omitempty"`     // Status of synchronization with external systems
	WorkerId string              `json:"workerId,omitempty" bson:"workerId,omitempty"` // ID of the worker handling synchronization
	Events   []SynchronizeEvents `json:"events,omitempty" bson:"events,omitempty"`     // Events related to synchronization, such as timestamps and messages
}

type SynchronizeEvents struct {
	Timestamp int64  `json:"timestamp,omitempty" bson:"timestamp,omitempty"` // Timestamp of the last synchronization attempt
	Message   string `json:"message,omitempty" bson:"message,omitempty"`     // Additional message or error description related to synchronization
}
