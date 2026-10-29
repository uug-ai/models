package models

import "go.mongodb.org/mongo-driver/bson/primitive"

const (
	PermissionIORead    Permission = "io.read"
	PermissionIOCreate  Permission = "io.create"
	PermissionIODelete  Permission = "io.delete"
	PermissionIOTrigger Permission = "io.trigger"
)

var ioPermissions = []Permission{
	PermissionIORead,
	PermissionIOCreate,
	PermissionIODelete,
	PermissionIOTrigger,
}

// IOPermissions returns the canonical IO permission catalog.
func IOPermissions() []Permission {
	return append([]Permission(nil), ioPermissions...)
}

type IO struct {
	Id             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	OrganisationId string             `json:"organisationId" bson:"organisationId,omitempty"`

	// ProjectId optionally places the IO in a project within its organisation.
	// A nil value places it in the organisation's default project (see ResolveProjectId).
	ProjectId *primitive.ObjectID `json:"projectId,omitempty" bson:"projectId,omitempty"`

	Hash              string `json:"hash" bson:"hash,omitempty"`
	DeviceId          string `json:"deviceId" bson:"deviceId,omitempty"` // device identifier
	Type              string `json:"type" bson:"type,omitempty"`         // input or output
	Key               string `json:"key" bson:"key,omitempty"`
	Value             string `json:"value" bson:"value,omitempty"`
	LastSeenTimestamp int64  `json:"lastSeenTimestamp" bson:"lastSeenTimestamp,omitempty"` // last time the IO was seen
	External          bool   `json:"external" bson:"external,omitempty"`

	// Audit information
	Audit *Audit `json:"audit,omitempty" bson:"audit,omitempty"`
}
