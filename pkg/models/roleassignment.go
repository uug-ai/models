package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RoleAssignment represents the assignment of an organisation-specific role to a user.
// This allows users to have multiple roles assigned to them within an organisation.
type RoleAssignment struct {
	Id             primitive.ObjectID  `json:"id" bson:"_id,omitempty"`
	UserId         primitive.ObjectID  `json:"userId" bson:"userId,omitempty"`
	RoleId         primitive.ObjectID  `json:"roleId" bson:"roleId,omitempty"`                 // Reference to the organisation-specific Role
	OrganisationId primitive.ObjectID  `json:"organisationId" bson:"organisationId,omitempty"` // Organisation context for this assignment
	ExpiresAt      time.Time           `json:"expiresAt" bson:"expiresAt,omitempty"`           // Optional expiration for temporary assignments
	IsActive       int                 `json:"isActive" bson:"isActive"`
	Scope          RoleAssignmentScope `json:"scope" bson:"scope,omitempty"` // Optional granular scope within organisation
	Audit          Audit               `json:"audit" bson:"audit,omitempty"`
}

// RoleAssignmentScope defines the scope/context where the role assignment applies.
// This allows for granular role assignments at different levels.
//
// Scope resolution: if AllOrganisation is true, the assignment applies to the
// entire organisation and the id lists are ignored. Otherwise the assignment is
// limited to the listed sites/groups/devices; empty lists mean NO access (never
// "everything").
type RoleAssignmentScope struct {
	AllOrganisation bool     `json:"allOrganisation" bson:"allOrganisation,omitempty"` // Applies org-wide (ignores the id lists)
	SiteIds         []string `json:"siteIds" bson:"siteIds,omitempty"`                 // Sites where the role applies
	GroupIds        []string `json:"groupIds" bson:"groupIds,omitempty"`               // Groups where the role applies
	DeviceIds       []string `json:"deviceIds" bson:"deviceIds,omitempty"`             // Devices where the role applies
}

// UserRoleAssignments is a helper struct to include role details with assignments
type UserRoleAssignments struct {
	UserId      primitive.ObjectID `json:"userId" bson:"userId,omitempty"`
	Assignments []RoleAssignment   `json:"assignments" bson:"assignments,omitempty"`
	Roles       []Role             `json:"roles" bson:"roles,omitempty"` // Populated role details
}
