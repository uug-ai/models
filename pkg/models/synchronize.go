package models

type Status string

const (
	SynchronizeStatusPending    Status = "pending"
	SynchronizeStatusSkipped    Status = "skipped"
	SynchronizeStatusAssigned   Status = "assigned"
	SynchronizeStatusInProgress Status = "inProgress"
	SynchronizeStatusCompleted  Status = "completed"
	SynchronizeStatusFailed     Status = "failed"
)

type Synchronize struct {
	SynchronizeEvent                    // This is the current event of synchronization
	Events           []SynchronizeEvent `json:"events,omitempty" bson:"events,omitempty"` // History of synchronization events
}

type SynchronizeEvent struct {
	Timestamp int64  `json:"timestamp,omitempty" bson:"timestamp,omitempty"` // Timestamp of the last synchronization attempt
	Status    Status `json:"status,omitempty" bson:"status,omitempty"`       // Status of synchronization with external systems
	WorkerId  string `json:"workerId,omitempty" bson:"workerId,omitempty"`   // ID of the worker handling synchronization
	Message   string `json:"message,omitempty" bson:"message,omitempty"`     // Additional message or error description related to synchronization

}
