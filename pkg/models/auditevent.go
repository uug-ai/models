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
	// ProjectId optionally narrows the event to a project within the organisation.
	// Organisation-wide actions leave it nil.
	ProjectId *primitive.ObjectID `json:"projectId,omitempty" bson:"projectId,omitempty"`
	// ActorId is the user who performed the action; ActorName is denormalised so
	// history can be rendered without a lookup.
	ActorId   primitive.ObjectID `json:"actorId" bson:"actorId,omitempty"`
	ActorName string             `json:"actorName,omitempty" bson:"actorName,omitempty"`
	// ActorType identifies the principal kind, e.g. "user", "apikey" or "system".
	ActorType string `json:"actorType,omitempty" bson:"actorType,omitempty"`
	// Action is a short semantic label, e.g. "case.commented", "device.renamed",
	// "member.suspended".
	Action string `json:"action" bson:"action,omitempty"`
	// Operation is the coarse CRUD-style action class.
	Operation string `json:"operation,omitempty" bson:"operation,omitempty"`
	// RequiredPermission is the canonical RBAC capability associated with the
	// action, expressed as "resource.action" (for example, "media.read").
	// Events such as authentication that do not require a resource permission
	// leave it empty.
	RequiredPermission string `json:"requiredPermission,omitempty" bson:"requiredPermission,omitempty"`
	// Category groups the action by subsystem.
	Category string `json:"category,omitempty" bson:"category,omitempty"`
	// Outcome records whether the action succeeded.
	Outcome string `json:"outcome,omitempty" bson:"outcome,omitempty"`
	// TargetType and TargetId identify the entity the action was performed on
	// (e.g. TargetType "case" with the case id). Query these to render an
	// entity's audit history. TargetId is a string so it can hold either an
	// ObjectID hex or a non-ObjectID key (such as a media key). TargetName is the
	// denormalised point-in-time label of the target.
	TargetType string `json:"targetType" bson:"targetType,omitempty"`
	TargetId   string `json:"targetId" bson:"targetId,omitempty"`
	TargetName string `json:"targetName,omitempty" bson:"targetName,omitempty"`
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
