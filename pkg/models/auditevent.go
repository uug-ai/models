package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AuditEvent is a single, append-only record of an action taken against an
// entity. Audit events live in their own collection (rather than in an array on
// the audited document) so the full history can be filtered, sorted and
// paginated efficiently — for example every event for a case, an organisation
// or an actor — and so retention/archival can be applied independently.
type AuditEvent struct {
	Id primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	// OrganisationId scopes the event to a tenant for org-wide audit queries.
	OrganisationId primitive.ObjectID `json:"organisationId" bson:"organisationId,omitempty"`
	// ActorId is the user who performed the action; ActorName is denormalised so
	// history can be rendered without a lookup.
	ActorId   primitive.ObjectID `json:"actorId" bson:"actorId,omitempty"`
	ActorName string             `json:"actorName,omitempty" bson:"actorName,omitempty"`
	// Action is a short semantic label, e.g. "case.commented", "device.renamed",
	// "member.suspended".
	Action string `json:"action" bson:"action,omitempty"`
	// TargetType and TargetId identify the entity the action was performed on
	// (e.g. TargetType "case" with the case id). Query these to render an
	// entity's audit history. TargetId is a string so it can hold either an
	// ObjectID hex or a non-ObjectID key (such as a media key).
	TargetType string `json:"targetType" bson:"targetType,omitempty"`
	TargetId   string `json:"targetId" bson:"targetId,omitempty"`
	// Changes optionally captures the field-level diff for this event.
	Changes []AuditFieldChange `json:"changes,omitempty" bson:"changes,omitempty"`
	// Metadata holds contextual key/value pairs (ip, userAgent, requestId, ...).
	Metadata map[string]string `json:"metadata,omitempty" bson:"metadata,omitempty"`
	// Timestamp is when the action occurred.
	Timestamp time.Time `json:"timestamp" bson:"timestamp,omitempty"`
}

// AuditFieldChange records a single field mutation captured by an AuditEvent.
type AuditFieldChange struct {
	Field    string `json:"field" bson:"field,omitempty"`
	OldValue string `json:"oldValue,omitempty" bson:"oldValue,omitempty"`
	NewValue string `json:"newValue,omitempty" bson:"newValue,omitempty"`
}
