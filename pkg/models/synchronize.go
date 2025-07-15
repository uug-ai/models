package models

type Status string

const (
	SynchronizationStatusPending    Status = "pending"
	SynchronizationStatusInProgress Status = "inProgress"
	SynchronizationStatusSkipped    Status = "skipped"
	SynchronizationStatusCompleted  Status = "completed"
	SynchronizationStatusFailed     Status = "failed"
)

type Synchronize struct {
	Status    Status `json:"status,omitempty" bson:"status,omitempty"`       // Status of synchronization with external systems
	WorkerId  string `json:"workerId,omitempty" bson:"workerId,omitempty"`   // ID of the worker handling synchronization
	Timestamp int64  `json:"timestamp,omitempty" bson:"timestamp,omitempty"` // Timestamp of the last synchronization attempt
	Message   string `json:"message,omitempty" bson:"message,omitempty"`     // Additional message or error description related to synchronization
}
