package models

import "time"

// Audit holds bounded "who/when" stamps embedded directly on an entity for
// cheap display and sorting (e.g. sort by audit.updatedAt). The full, filterable
// change history lives in the separate AuditEvent collection (query it by
// target) so audited documents never accumulate an unbounded history array.
//
// The fields are kept flat (a single level under audit) so they map to simple,
// indexable paths such as "audit.createdAt" / "audit.updatedAt".
type Audit struct {
	CreatedBy string    `json:"createdBy,omitempty" bson:"createdBy,omitempty"`
	CreatedAt time.Time `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	UpdatedBy string    `json:"updatedBy,omitempty" bson:"updatedBy,omitempty"`
	UpdatedAt time.Time `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
	// LastAction is a short semantic label describing the most recent change
	// (e.g. "role.assigned", "member.suspended", "device.renamed"). For the
	// complete history, query AuditEvent by TargetType/TargetId.
	LastAction string `json:"lastAction,omitempty" bson:"lastAction,omitempty"`
}
