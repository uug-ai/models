package models

// Audit contains common audit fields for tracking creation and updates.
type Audit struct {
	Create AuditCreate   `json:"create,omitempty" bson:"create,omitempty"`
	Update []AuditUpdate `json:"update,omitempty" bson:"update,omitempty"`
}

type AuditCreate struct {
	CreatedBy string `json:"createdBy,omitempty" bson:"createdBy,omitempty"`
	CreatedAt int64  `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
}

type AuditUpdate struct {
	UpdatedBy string `json:"updatedBy,omitempty" bson:"updatedBy,omitempty"`
	UpdatedAt int64  `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
}
