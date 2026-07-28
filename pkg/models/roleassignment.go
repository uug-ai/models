package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RoleAssignment represents the assignment of an organisation-specific role to a user.
// This allows users to have multiple roles assigned to them within an organisation.
type RoleAssignment struct {
	Id             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserId         primitive.ObjectID `json:"userId" bson:"userId,omitempty"`
	RoleId         primitive.ObjectID `json:"roleId" bson:"roleId,omitempty"`                 // Reference to the organisation-specific Role
	OrganisationId primitive.ObjectID `json:"organisationId" bson:"organisationId,omitempty"` // Organisation context for this assignment
	// ExpiresAt optionally bounds the assignment in time. It is an overlay on
	// IsActive: an assignment is only effectively active when IsActive is true
	// AND ExpiresAt is unset or in the future. Use IsEffectivelyActive().
	ExpiresAt time.Time           `json:"expiresAt" bson:"expiresAt,omitempty"`
	IsActive  bool                `json:"isActive" bson:"isActive"`     // Administrative enable/disable switch (independent of expiry)
	Scope     RoleAssignmentScope `json:"scope" bson:"scope,omitempty"` // Optional granular scope within organisation
	Audit     Audit               `json:"audit" bson:"audit,omitempty"`
}

// IsExpired reports whether the assignment has a set expiry that is in the past.
func (a RoleAssignment) IsExpired() bool {
	return !a.ExpiresAt.IsZero() && a.ExpiresAt.Before(time.Now())
}

// IsEffectivelyActive reports whether the assignment is currently in force:
// administratively active and not past its expiry. This is the single source
// of truth consumers should use for access decisions.
func (a RoleAssignment) IsEffectivelyActive() bool {
	return a.IsActive && !a.IsExpired()
}

// RoleAssignmentScope narrows a RoleAssignment to a subset of the organisation's
// resources. It follows the Kubernetes RBAC convention (resourceNames: an empty
// set means "everything"): each dimension is an independent allow-list where an
// empty list places NO restriction on that dimension (i.e. all sites / all groups
// / all devices). A zero-value scope therefore grants the role across the whole
// organisation; populate a list to restrict the assignment to those specific
// resources on that dimension.
//
// Deny-by-default lives one level up: a user only has access where a
// RoleAssignment exists and IsEffectivelyActive() is true. The scope only ever
// narrows an existing grant — it can widen nothing.
//
// The dimensions are orthogonal facets: the permission resolver applies SiteIds
// to site-level access, GroupIds to group-level access and DeviceIds to
// device-level access, each with "empty = all". This makes combinations such as
// "all sites, but only two devices" expressible (SiteIds empty, DeviceIds set).
type RoleAssignmentScope struct {
	SiteIds   []primitive.ObjectID `json:"siteIds" bson:"siteIds,omitempty"`     // Restrict to these sites (empty = all sites)
	GroupIds  []primitive.ObjectID `json:"groupIds" bson:"groupIds,omitempty"`   // Restrict to these groups (empty = all groups)
	DeviceIds []primitive.ObjectID `json:"deviceIds" bson:"deviceIds,omitempty"` // Restrict to these devices (empty = all devices)
}

// IsOrganisationWide reports whether the scope places no restriction on any
// dimension, i.e. the assignment applies to the entire organisation.
func (s RoleAssignmentScope) IsOrganisationWide() bool {
	return len(s.SiteIds) == 0 && len(s.GroupIds) == 0 && len(s.DeviceIds) == 0
}

// UserRoleAssignments is a helper struct to include role details with assignments
type UserRoleAssignments struct {
	UserId      primitive.ObjectID `json:"userId" bson:"userId,omitempty"`
	Assignments []RoleAssignment   `json:"assignments" bson:"assignments,omitempty"`
	Roles       []Role             `json:"roles" bson:"roles,omitempty"` // Populated role details
}
