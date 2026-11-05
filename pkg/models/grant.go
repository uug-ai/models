package models

import (
	"context"
	"fmt"
	"slices"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GrantDenial says why a grant goes beyond the grantor's access.
type GrantDenial string

const (
	// GrantDenialOwnerOnly: only an owner grants or revokes Owner.
	GrantDenialOwnerOnly GrantDenial = "ownerOnly"
	// GrantDenialNoAccessPermission: the grantor may not manage access there
	// at all: members.update for organisation roles, project.manageAccess or
	// projects.manageAccess for project roles.
	GrantDenialNoAccessPermission GrantDenial = "noAccessPermission"
	// GrantDenialPermissionNotHeld: the role holds a permission the grantor
	// does not hold where it would apply (GrantDecision.Permission).
	GrantDenialPermissionNotHeld GrantDenial = "permissionNotHeld"
	// GrantDenialResourcesNotCovered: the selected resources reach devices
	// outside the grantor's own access scope in the project.
	GrantDenialResourcesNotCovered GrantDenial = "resourcesNotCovered"
)

// GrantDecision is the outcome of a grant check. When it is not allowed,
// Reason says why and, for GrantDenialPermissionNotHeld, Permission names the
// first permission (in the role's order) and Project the project it is not
// held in (zero for "all projects").
type GrantDecision struct {
	Allowed    bool
	Reason     GrantDenial
	Permission Permission
	Project    primitive.ObjectID
}

func allowGrant() GrantDecision { return GrantDecision{Allowed: true} }

func denyGrant(reason GrantDenial) GrantDecision { return GrantDecision{Reason: reason} }

// CanGrant reports whether the member may grant role on scope to someone
// else (spec 4.5). Permissions and resources are bounded separately:
//
//   - Owner is granted only by owners; owners may grant anything.
//   - An organisation role needs members.update, and the grantor must hold
//     every permission of the role on every project the scope covers (all
//     projects, or each selected one): its project selection stays inside
//     the grantor's own.
//   - A project role needs projects.manageAccess covering the project or
//     project.manageAccess in it, and the grantor must hold every permission
//     of the role in that project, on any of its resources.
//   - With projects.manageAccess any resources of the project may be
//     selected. With only project.manageAccess the selection must stay
//     inside the devices that assignment covers: all resources only when it
//     covers all, otherwise the selected sites, groups and devices, expanded
//     now (ExpandDeviceKeys), must be a subset of the grantor's own, expanded
//     the same way. A selection without devices ("no devices yet") is always
//     inside.
//
// Removing or narrowing an assignment is bounded the same way (CanChange).
// Whether the grant touches the grantor's own access (directly or through a
// team) is a separate rule the writer checks on the assignment's principal.
//
// The error is ErrRoleAssignmentInvalid for a role or scope that is invalid
// or do not fit together, or the expansion's read error; the caller must deny
// on any error. source must be limited to the organisation (and project).
func (a Access) CanGrant(ctx context.Context, source DeviceSource, role Role, scope RoleAssignmentScope) (GrantDecision, error) {
	if err := role.Validate(); err != nil {
		return GrantDecision{}, fmt.Errorf("%w: %v", ErrRoleAssignmentInvalid, err)
	}
	if err := scope.ValidateFor(role); err != nil {
		return GrantDecision{}, err
	}
	if role.IsOwner() {
		if a.owner {
			return allowGrant(), nil
		}
		return denyGrant(GrantDenialOwnerOnly), nil
	}
	if a.owner {
		return allowGrant(), nil
	}
	if scope.Level == LevelOrganisation {
		return a.canGrantOrganisationRole(role, scope), nil
	}
	return a.canGrantProjectRole(ctx, source, role, scope)
}

// CanChange is CanGrant for a change to an assignment: current is the scope
// it has (nil when creating it), next the scope it gets (nil when removing
// it). A change is allowed only when the grantor could grant both, so nobody
// narrows or removes access they could not have granted.
func (a Access) CanChange(ctx context.Context, source DeviceSource, role Role, current, next *RoleAssignmentScope) (GrantDecision, error) {
	if current == nil && next == nil {
		return GrantDecision{}, fmt.Errorf("%w: a change needs a current or a next scope", ErrRoleAssignmentInvalid)
	}
	for _, scope := range []*RoleAssignmentScope{current, next} {
		if scope == nil {
			continue
		}
		decision, err := a.CanGrant(ctx, source, role, *scope)
		if err != nil || !decision.Allowed {
			return decision, err
		}
	}
	return allowGrant(), nil
}

func (a Access) canGrantOrganisationRole(role Role, scope RoleAssignmentScope) GrantDecision {
	if !a.Organisation(PermissionMembersUpdate, primitive.NilObjectID) {
		return denyGrant(GrantDenialNoAccessPermission)
	}
	projects := []primitive.ObjectID{primitive.NilObjectID}
	if scope.Projects == SelectionSelected {
		projects = scope.ProjectIds
	}
	for _, permission := range role.Permissions {
		for _, project := range projects {
			if !a.Organisation(permission, project) {
				return GrantDecision{Reason: GrantDenialPermissionNotHeld, Permission: permission, Project: project}
			}
		}
	}
	return allowGrant()
}

func (a Access) canGrantProjectRole(ctx context.Context, source DeviceSource, role Role, scope RoleAssignmentScope) (GrantDecision, error) {
	project := scope.ProjectId
	anyResources := a.Organisation(PermissionProjectsManageAccess, project)
	if !anyResources && !a.Project(PermissionProjectManageAccess, project) {
		return denyGrant(GrantDenialNoAccessPermission), nil
	}
	held := a.Permissions(project)
	for _, permission := range role.Permissions {
		if !slices.Contains(held, permission) {
			return GrantDecision{Reason: GrantDenialPermissionNotHeld, Permission: permission, Project: project}, nil
		}
	}
	if anyResources {
		return allowGrant(), nil
	}
	own := a.Devices(PermissionProjectManageAccess, project)
	if own.All {
		return allowGrant(), nil
	}
	if scope.Resources == SelectionAll {
		return denyGrant(GrantDenialResourcesNotCovered), nil
	}
	if !scope.hasResourceIds() {
		return allowGrant(), nil
	}
	selected, err := ExpandDeviceKeys(ctx, source, hexIds(scope.SiteIds), hexIds(scope.GroupIds), scope.DeviceKeys)
	if err != nil {
		return GrantDecision{}, err
	}
	if len(selected) == 0 {
		return allowGrant(), nil
	}
	covered, err := ExpandDeviceKeys(ctx, source, hexIds(own.SiteIds), hexIds(own.GroupIds), own.DeviceKeys)
	if err != nil {
		return GrantDecision{}, err
	}
	for _, key := range selected {
		if _, found := slices.BinarySearch(covered, key); !found {
			return denyGrant(GrantDenialResourcesNotCovered), nil
		}
	}
	return allowGrant(), nil
}
