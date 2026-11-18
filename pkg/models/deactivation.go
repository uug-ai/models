package models

import "time"

// Reasons of a Deactivation.
const (
	DeactivationReasonDeactivated = "deactivated"
	DeactivationReasonDeletion    = "deletion"
)

// Deactivation is the lock on an owner account (User.Deactivation) or on an
// organisation (Organisation.Deactivation). While it is set, the account with
// its sub-accounts, or every member of the organisation, is refused, and so
// are their tokens. With reason "deletion" the account or organisation is
// erased once DeleteAfter has passed; PurgeStartedAt marks that the erasure
// began, after which the lock can no longer be lifted.
type Deactivation struct {
	At             time.Time  `json:"at" bson:"at"`
	Reason         string     `json:"reason" bson:"reason"`
	DeleteAfter    *time.Time `json:"deleteAfter,omitempty" bson:"deleteAfter,omitempty"`
	PurgeStartedAt *time.Time `json:"purgeStartedAt,omitempty" bson:"purgeStartedAt,omitempty"`
}
