package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RoleAssignment represents the assignment of an organisation-specific role to a user.
// This allows users to have multiple roles assigned to them within an organisation.
type RoleAssignment struct {
	Id             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserId         primitive.ObjectID `json:"userId" bson:"userId,omitempty"`                     // Reference to the user
	RoleId         primitive.ObjectID `json:"roleId" bson:"roleId,omitempty"`                 // Reference to the organisation-specific Role
	OrganisationId primitive.ObjectID `json:"organisationId" bson:"organisationId,omitempty"` // Organisation context for this assignment
	ExpiresAt      time.Time          `json:"expiresAt" bson:"expiresAt,omitempty"`           // Optional expiration for temporary assignments
	IsActive       int                `json:"isActive" bson:"isActive"`
	Audit          *Audit              `json:"audit" bson:"audit,omitempty"`
}

// UserRoleAssignments is a helper struct to include role details with assignments
type UserRoleAssignments struct {
	UserId      primitive.ObjectID `json:"userId" bson:"userId,omitempty"`
	Assignments []RoleAssignment   `json:"assignments" bson:"assignments,omitempty"`
	Roles       []Role             `json:"roles" bson:"roles,omitempty"` // Populated role details
}
