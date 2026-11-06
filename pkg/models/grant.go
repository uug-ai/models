package models

import (
	"context"
	"fmt"
	"slices"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Administering access is gated by permissions, not by holding what is
// handed out (spec 4.5): whoever holds members.update, projects.manageAccess,
// project.manageAccess, roles.* or teams.update administers that area, through
// a system or a custom role. Three rules bound them:
//
//   - Owner-only permissions (PermissionDefinition.OwnerOnly) and the Owner
//     role are handled by owners only, so a delegated administrator hands out
//     at most what the legacy admin tier holds.
//   - A holder of only project.manageAccess selects resources inside their own
//     scope in that project.
//   - Nobody changes their own access: an assignment, a role or a team
//     membership that applies to them (owners excepted, they hold everything).

// GrantDenial says why an access change is refused.
type GrantDenial string

const (
	// GrantDenialOwnerOnly: only an owner grants, changes or revokes the Owner
	// role or a role holding an owner-only permission, and only an owner
	// changes the members of a team that is assigned such a role.
	// GrantDecision.Permission names the owner-only permission (empty for the
	// Owner role).
	GrantDenialOwnerOnly GrantDenial = "ownerOnly"
	// GrantDenialNoAccessPermission: the actor lacks the permission for the
	// change: members.update for organisation roles, projects.manageAccess
	// covering the project or project.manageAccess in it for project roles,
	// roles.create, roles.update or roles.delete for roles, teams.update for
	// team members.
	GrantDenialNoAccessPermission GrantDenial = "noAccessPermission"
	// GrantDenialResourcesNotCovered: the selected resources reach devices
	// outside the grantor's own scope in the project.
	GrantDenialResourcesNotCovered GrantDenial = "resourcesNotCovered"
	// GrantDenialSelf: the change applies to the actor's own access, a role
	// they hold (directly or through a team) or their own team membership.
	GrantDenialSelf GrantDenial = "self"
)

// GrantDecision is the outcome of an access-change check. When it is not
// allowed, Reason says why and, for GrantDenialOwnerOnly, Permission names the
// first owner-only permission of the role in catalog order.
type GrantDecision struct {
	Allowed    bool
	Reason     GrantDenial
	Permission Permission
}

func allowGrant() GrantDecision { return GrantDecision{Allowed: true} }

func denyGrant(reason GrantDenial) GrantDecision { return GrantDecision{Reason: reason} }

// OwnerOnlyPermission returns the role's first owner-only permission in
// catalog order, implied permissions included. The Owner role, which grants
// all, reports true with no permission.
func (r Role) OwnerOnlyPermission() (Permission, bool) {
	if r.GrantsAll() {
		return "", true
	}
	for _, permission := range ExpandPermissions(r.Permissions) {
		if definition, _ := LookupPermission(permission); definition.OwnerOnly {
			return permission, true
		}
	}
	return "", false
}

// ownerOnlyDenial returns the denial for a non-owner handling role, when the
// role is the Owner role or holds an owner-only permission.
func ownerOnlyDenial(role Role) (GrantDecision, bool) {
	if role.IsOwner() {
		return denyGrant(GrantDenialOwnerOnly), true
	}
	if permission, ok := role.OwnerOnlyPermission(); ok {
		return GrantDecision{Reason: GrantDenialOwnerOnly, Permission: permission}, true
	}
	return GrantDecision{}, false
}

// CanGrant reports whether the member may grant role on scope to someone
// else (spec 4.5):
//
//   - Owners may grant anything; only owners grant the Owner role or a role
//     holding an owner-only permission.
//   - An organisation role needs members.update, on any project selection.
//   - A project role needs projects.manageAccess covering the project, which
//     selects any of its resources, or project.manageAccess in it, which
//     selects resources inside the devices that assignment covers: all
//     resources only when it covers all, otherwise the selected sites, groups
//     and devices, expanded now (ExpandDeviceKeys), must be a subset of the
//     grantor's own, expanded the same way. A selection without devices ("no
//     devices yet") is always inside.
//
// The grantor need not hold the role's permissions. Removing or narrowing an
// assignment is bounded the same way (CanChange). Whether the assignment
// applies to the grantor (RoleAssignment.AffectsUser) is checked by the writer
// on the assignment's principal.
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
	if a.owner {
		return allowGrant(), nil
	}
	if denial, ok := ownerOnlyDenial(role); ok {
		return denial, nil
	}
	if scope.Level == LevelOrganisation {
		if !a.Organisation(PermissionMembersUpdate, primitive.NilObjectID) {
			return denyGrant(GrantDenialNoAccessPermission), nil
		}
		return allowGrant(), nil
	}
	return a.canGrantProjectResources(ctx, source, scope)
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

func (a Access) canGrantProjectResources(ctx context.Context, source DeviceSource, scope RoleAssignmentScope) (GrantDecision, error) {
	project := scope.ProjectId
	if a.Organisation(PermissionProjectsManageAccess, project) {
		return allowGrant(), nil
	}
	if !a.Project(PermissionProjectManageAccess, project) {
		return denyGrant(GrantDenialNoAccessPermission), nil
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

// CanManageRole reports whether the member (userId, with their TeamMember
// documents teams) may create, change or delete a custom role: current is the
// role as stored (nil when creating), next what it becomes (nil when
// deleting), and assignments the organisation's assignments of current. A
// change applies at once to everyone the role is assigned to.
//
//   - Creating needs roles.create, changing roles.update, deleting
//     roles.delete; the author need not hold the role's permissions.
//   - Only owners handle a role holding an owner-only permission, before or
//     after the change.
//   - Nobody but an owner changes or deletes a role they hold, directly or
//     through a team, in force or not.
//
// The error is ErrRoleInvalid for a system or legacy role, which are
// read-only, for next failing Role.Validate, or for neither role given; the
// caller must deny on any error.
func (a Access) CanManageRole(userId primitive.ObjectID, teams []TeamMember, current, next *Role, assignments []RoleAssignment) (GrantDecision, error) {
	if current == nil && next == nil {
		return GrantDecision{}, fmt.Errorf("%w: a change needs a current or a next role", ErrRoleInvalid)
	}
	for _, role := range []*Role{current, next} {
		if role != nil && role.Kind != RoleKindCustom {
			return GrantDecision{}, fmt.Errorf("%w: %s role %q is read-only", ErrRoleInvalid, role.Kind, role.Key)
		}
	}
	if next != nil {
		if err := next.Validate(); err != nil {
			return GrantDecision{}, err
		}
	}
	if a.owner {
		return allowGrant(), nil
	}
	permission := PermissionRolesUpdate
	switch {
	case current == nil:
		permission = PermissionRolesCreate
	case next == nil:
		permission = PermissionRolesDelete
	}
	if !a.Organisation(permission, primitive.NilObjectID) {
		return denyGrant(GrantDenialNoAccessPermission), nil
	}
	for _, role := range []*Role{current, next} {
		if role == nil {
			continue
		}
		if denial, ok := ownerOnlyDenial(*role); ok {
			return denial, nil
		}
	}
	if current != nil {
		for _, assignment := range assignments {
			if assignment.RoleId == current.Id && assignment.AffectsUser(userId, teams) {
				return denyGrant(GrantDenialSelf), nil
			}
		}
	}
	return allowGrant(), nil
}

// CanChangeTeamMembers reports whether the member (userId) may add or remove
// members (user ids) of team. Adding a member grants them every assignment of
// the team at once, so teamRoles are the roles of the team's assignments.
//
//   - It needs teams.update; the actor need not hold the team's access.
//   - Only owners change the members of a team assigned a role holding an
//     owner-only permission.
//   - Nobody but an owner adds or removes themselves.
//
// The error is ErrTeamInvalid for a team synced from an identity provider,
// whose members change only through the provider's sync; the caller must
// deny on any error. Inviting someone into a team is the same change.
func (a Access) CanChangeTeamMembers(userId primitive.ObjectID, team Team, teamRoles []Role, members []primitive.ObjectID) (GrantDecision, error) {
	if !team.ManagesMembersInHub() {
		return GrantDecision{}, fmt.Errorf("%w: members of team %q change through its identity provider", ErrTeamInvalid, team.Name)
	}
	if a.owner {
		return allowGrant(), nil
	}
	if !a.Organisation(PermissionTeamsUpdate, primitive.NilObjectID) {
		return denyGrant(GrantDenialNoAccessPermission), nil
	}
	for _, role := range teamRoles {
		if denial, ok := ownerOnlyDenial(role); ok {
			return denial, nil
		}
	}
	if !userId.IsZero() && slices.Contains(members, userId) {
		return denyGrant(GrantDenialSelf), nil
	}
	return allowGrant(), nil
}
