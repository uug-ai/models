package models

// Audit contains common audit fields for tracking creation and updates.
type Audit struct {
	CreatedBy string `json:"createdBy,omitempty" bson:"createdBy,omitempty"`
	CreatedAt int64  `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	UpdatedBy string `json:"updatedBy,omitempty" bson:"updatedBy,omitempty"`
	UpdatedAt int64  `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
}
