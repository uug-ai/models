package models

// Audit holds bounded creation and last-update stamps embedded directly on an
// entity for cheap "who/when" display. The full, filterable change history
// lives in the separate AuditEvent collection (query it by target) so audited
// documents never accumulate an unbounded history array.
type Audit struct {
	Create AuditCreate `json:"create,omitempty" bson:"create,omitempty"`
	// Update records only the most recent modification. For the complete
	// history, query AuditEvent by TargetType/TargetId.
	Update AuditUpdate `json:"update,omitempty" bson:"update,omitempty"`
}

type AuditCreate struct {
	CreatedBy string `json:"createdBy,omitempty" bson:"createdBy,omitempty"`
	CreatedAt int64  `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
}

type AuditUpdate struct {
	UpdatedBy string `json:"updatedBy,omitempty" bson:"updatedBy,omitempty"`
	UpdatedAt int64  `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
	// Action is a short semantic label describing what changed in this update
	// (e.g. "role.assigned", "member.suspended", "device.renamed").
	Action string `json:"action,omitempty" bson:"action,omitempty"`
}
