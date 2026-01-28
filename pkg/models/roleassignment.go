package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RoleAssignment represents the assignment of an organisation-specific role to a user.
// This allows users to have multiple roles assigned to them within an organisation.
type RoleAssignment struct {
	Id             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserId         primitive.ObjectID `json:"user_id" bson:"user_id,omitempty"`
	RoleId         primitive.ObjectID `json:"role_id" bson:"role_id,omitempty"`                 // Reference to the organisation-specific Role
	OrganisationId primitive.ObjectID `json:"organisation_id" bson:"organisation_id,omitempty"` // Organisation context for this assignment
	ExpiresAt      time.Time          `json:"expires_at" bson:"expires_at,omitempty"`           // Optional expiration for temporary assignments
	IsActive       int                `json:"is_active" bson:"is_active"`
	Scope          RoleScope          `json:"scope" bson:"scope,omitempty"` // Optional granular scope within organisation
	Audit          Audit              `json:"audit" bson:"audit,omitempty"`
}

// RoleScope defines the scope/context where the role assignment applies.
// This allows for granular role assignments at different levels.
type RoleScope struct {
	Type      string   `json:"type" bson:"type,omitempty"`             // e.g., "global", "site", "group", "device"
	SiteIds   []string `json:"site_ids" bson:"site_ids,omitempty"`     // Sites where the role applies
	GroupIds  []string `json:"group_ids" bson:"group_ids,omitempty"`   // Groups where the role applies
	DeviceIds []string `json:"device_ids" bson:"device_ids,omitempty"` // Devices where the role applies
}

// UserRoleAssignments is a helper struct to include role details with assignments
type UserRoleAssignments struct {
	UserId      primitive.ObjectID `json:"user_id" bson:"user_id,omitempty"`
	Assignments []RoleAssignment   `json:"assignments" bson:"assignments,omitempty"`
	Roles       []Role             `json:"roles" bson:"roles,omitempty"` // Populated role details
}
