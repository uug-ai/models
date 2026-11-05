package models

import (
	"fmt"
	"time"
)

// RoleAssignmentPatch is a partial update of a role assignment: only the
// fields that are set change. Who it grants to, the role and where it applies
// (the scope's level and project) never change; granting another role or in
// another project is another assignment. The result of Apply still needs
// Validate against the role.
type RoleAssignmentPatch struct {
	// Scope replaces the scope, e.g. to widen or narrow the selected sites,
	// groups and devices, or to switch between all and selected resources or
	// projects. Its level and project must stay the same.
	Scope *RoleAssignmentScope `json:"scope,omitempty"`
	// TimeWindow replaces the time window; RemoveTimeWindow removes it.
	TimeWindow       *TimeWindow `json:"timeWindow,omitempty"`
	RemoveTimeWindow bool        `json:"removeTimeWindow,omitempty"`
	// ExpiresAt replaces the expiry; RemoveExpiry removes it.
	ExpiresAt    *time.Time `json:"expiresAt,omitempty"`
	RemoveExpiry bool       `json:"removeExpiry,omitempty"`
	// IsActive enables or disables the assignment without removing it.
	IsActive *bool `json:"isActive,omitempty"`
}

// IsEmpty reports whether the patch changes nothing.
func (p RoleAssignmentPatch) IsEmpty() bool {
	return p.Scope == nil && p.TimeWindow == nil && !p.RemoveTimeWindow &&
		p.ExpiresAt == nil && !p.RemoveExpiry && p.IsActive == nil
}

// Apply returns the assignment with the patch applied. It refuses a patch
// that sets and removes the same field, or that moves the assignment to
// another level or project.
func (p RoleAssignmentPatch) Apply(assignment RoleAssignment) (RoleAssignment, error) {
	if p.TimeWindow != nil && p.RemoveTimeWindow {
		return RoleAssignment{}, fmt.Errorf("%w: time window both set and removed", ErrRoleAssignmentInvalid)
	}
	if p.ExpiresAt != nil && p.RemoveExpiry {
		return RoleAssignment{}, fmt.Errorf("%w: expiry both set and removed", ErrRoleAssignmentInvalid)
	}
	if p.Scope != nil {
		if p.Scope.Level != assignment.Scope.Level || p.Scope.ProjectId != assignment.Scope.ProjectId {
			return RoleAssignment{}, fmt.Errorf("%w: a patch cannot move an assignment to another level or project", ErrRoleAssignmentInvalid)
		}
		assignment.Scope = *p.Scope
	}
	switch {
	case p.RemoveTimeWindow:
		assignment.TimeWindow = nil
	case p.TimeWindow != nil:
		assignment.TimeWindow = p.TimeWindow
	}
	switch {
	case p.RemoveExpiry:
		assignment.ExpiresAt = time.Time{}
	case p.ExpiresAt != nil:
		assignment.ExpiresAt = *p.ExpiresAt
	}
	if p.IsActive != nil {
		assignment.IsActive = *p.IsActive
	}
	return assignment, nil
}

// SameSlot reports whether two assignments occupy the same slot: the same
// organisation, principal, role and project (organisation-level assignments
// have no project). An organisation holds at most one assignment per slot;
// granting the same role on more resources widens that assignment instead.
func (a RoleAssignment) SameSlot(other RoleAssignment) bool {
	return a.OrganisationId == other.OrganisationId &&
		a.Principal == other.Principal &&
		a.RoleId == other.RoleId &&
		a.Scope.Level == other.Scope.Level &&
		a.Scope.ProjectId == other.Scope.ProjectId
}
