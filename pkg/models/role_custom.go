package models

import (
	"slices"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RolePatch is a partial update of a custom role: only the fields that are
// set change. The kind and the level a role is assigned at never change, so
// existing assignments stay valid; a role for another level is another role.
// A change applies at once to everyone the role is assigned to. The result of
// Apply still needs Validate (and models.Access.CanManageRole).
type RolePatch struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	// Permissions replaces the role's permissions; an empty list leaves a
	// role that grants nothing.
	Permissions *[]Permission `json:"permissions,omitempty"`
}

// IsEmpty reports whether the patch changes nothing.
func (p RolePatch) IsEmpty() bool {
	return p.Name == nil && p.Description == nil && p.Permissions == nil
}

// Apply returns the role with the patch applied, permissions in catalog
// order. It does not change the stored role's slices.
func (p RolePatch) Apply(role Role) Role {
	if p.Name != nil {
		role.Name = *p.Name
	}
	if p.Description != nil {
		role.Description = *p.Description
	}
	if p.Permissions != nil {
		role.Permissions = catalogOrder(*p.Permissions)
	} else {
		role.Permissions = slices.Clone(role.Permissions)
	}
	return role
}

// NewCustomRole returns an organisation's new custom role: active, granting
// exactly the given permissions (in catalog order) at one level. The writer
// sets the id and Audit; the result still needs Validate (and
// models.Access.CanManageRole).
func NewCustomRole(organisationId primitive.ObjectID, name, description string, level Level, permissions []Permission) Role {
	return Role{
		OrganisationId: organisationId,
		Name:           name,
		Description:    description,
		Kind:           RoleKindCustom,
		AssignableAt:   level,
		Grants:         SelectionSelected,
		Permissions:    catalogOrder(permissions),
		IsActive:       true,
	}
}
