package models

import (
	"slices"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Access answers what one member may do in one organisation at one instant.
// It is built once per request from the member's membership, team
// memberships, roles and assignments (their own and their teams'), whether
// those are stored or mapped from legacy access (see MapLegacyAccess), and
// holds no I/O: the caller resolves which sites and groups a device belongs
// to.
//
// Access only covers role permissions. Relationship rules (creators, case
// participants, assigned users) and per-device feature ceilings are applied
// by the caller next to it.
type Access struct {
	owner  bool
	grants []accessGrant
}

type accessGrant struct {
	permissions []Permission // Expanded with everything they imply
	scope       RoleAssignmentScope
}

// DeviceRef is a device together with everything that selects it in an
// assignment scope.
type DeviceRef struct {
	Key string
	// SiteIds are the sites containing the device, directly or through a
	// camera group inside the site.
	SiteIds []primitive.ObjectID
	// GroupIds are the camera groups containing the device and the site
	// groups containing one of its sites.
	GroupIds []primitive.ObjectID
}

// DeviceSelection is the union of devices a permission is held on in one
// project: every device when All is set, otherwise the devices in SiteIds,
// in GroupIds and the devices with a key in DeviceKeys.
type DeviceSelection struct {
	All        bool
	SiteIds    []primitive.ObjectID
	GroupIds   []primitive.ObjectID
	DeviceKeys []string
}

// NewAccess resolves a member's access at the given instant. Nothing is
// granted unless the membership is effectively active, team-derived access
// included. An assignment counts only when it is for this member, directly
// or through one of the member's teams (teams: the member's TeamMember
// documents), effectively active (see RoleAssignment.IsEffectivelyActive)
// and valid for an active role of the same organisation among roles;
// anything else is skipped, so incomplete data denies. Grants from several
// assignments add up.
func NewAccess(membership OrganisationUser, teams []TeamMember, roles []Role, assignments []RoleAssignment, at time.Time) Access {
	var access Access
	if !membership.isEffectivelyActiveAt(at) || membership.OrganisationId.IsZero() {
		return access
	}
	teamIds := teamIdsOf(membership, teams)
	for _, assignment := range assignments {
		if assignment.OrganisationId != membership.OrganisationId ||
			!grantsTo(assignment.Principal, membership.UserId, teamIds) ||
			!assignment.isEffectivelyActiveAt(at) {
			continue
		}
		index := slices.IndexFunc(roles, func(role Role) bool { return role.Id == assignment.RoleId })
		if index < 0 {
			continue
		}
		role := roles[index]
		if !role.IsActive || role.Validate() != nil || assignment.Validate(role) != nil {
			continue
		}
		if role.IsOwner() {
			access.owner = true
			continue
		}
		access.grants = append(access.grants, accessGrant{
			permissions: ExpandPermissions(role.Permissions),
			scope:       assignment.Scope,
		})
	}
	return access
}

// grantsTo reports whether a principal is the user or one of their teams.
func grantsTo(principal Principal, userId primitive.ObjectID, teamIds []primitive.ObjectID) bool {
	switch principal.Type {
	case PrincipalTypeUser:
		return principal.Id == userId
	case PrincipalTypeTeam:
		return slices.Contains(teamIds, principal.Id)
	}
	return false
}

// IsOwner reports whether the member holds the Owner role.
func (a Access) IsOwner() bool {
	return a.owner
}

// Organisation reports whether the member holds an organisation permission.
// With a zero projectId it asks about the organisation as a whole, which
// only an assignment covering all projects grants; otherwise about the
// organisation permission within that project.
func (a Access) Organisation(permission Permission, projectId primitive.ObjectID) bool {
	permission = CanonicalPagePermission(permission)
	if definition, ok := LookupPermission(permission); !ok || definition.Level != LevelOrganisation {
		return false
	}
	if a.owner {
		return true
	}
	for _, grant := range a.grants {
		if grant.scope.Level == LevelOrganisation && slices.Contains(grant.permissions, permission) && grant.coversProject(projectId) {
			return true
		}
	}
	return false
}

// Project reports whether the member holds a project permission on the
// project's own resources that have no device, such as cases, labels and
// workflows. Any assignment in the project counts, including one whose
// selection has no devices yet.
func (a Access) Project(permission Permission, projectId primitive.ObjectID) bool {
	return a.projectGrant(permission, projectId, func(accessGrant) bool { return true })
}

// Device reports whether the member holds a project permission on a device
// of the project.
func (a Access) Device(permission Permission, projectId primitive.ObjectID, device DeviceRef) bool {
	return a.projectGrant(permission, projectId, func(grant accessGrant) bool { return grant.scope.coversDevice(device) })
}

// Devices returns the devices of a project the member holds a permission on,
// for filtering lists; the caller expands the sites and groups.
func (a Access) Devices(permission Permission, projectId primitive.ObjectID) DeviceSelection {
	var selection DeviceSelection
	if a.owner {
		selection.All = true
		return selection
	}
	for _, grant := range a.grants {
		if !grant.inProject(permission, projectId) {
			continue
		}
		if grant.scope.Resources == SelectionAll {
			return DeviceSelection{All: true}
		}
		selection.SiteIds = appendMissing(selection.SiteIds, grant.scope.SiteIds)
		selection.GroupIds = appendMissing(selection.GroupIds, grant.scope.GroupIds)
		selection.DeviceKeys = appendMissing(selection.DeviceKeys, grant.scope.DeviceKeys)
	}
	return selection
}

// Permissions returns every permission the member holds in a project, in
// catalog order with implied permissions included: the organisation
// permissions covering the project and the project permissions on any of
// its resources. With a zero projectId it returns the organisation
// permissions held on all projects.
func (a Access) Permissions(projectId primitive.ObjectID) []Permission {
	if a.owner {
		return AllPermissions()
	}
	var held []Permission
	for _, grant := range a.grants {
		switch {
		case grant.scope.Level == LevelOrganisation && grant.coversProject(projectId):
			held = append(held, grant.permissions...)
		case grant.scope.Level == LevelProject && !projectId.IsZero() && grant.scope.ProjectId == projectId:
			held = append(held, grant.permissions...)
		}
	}
	return ExpandPermissions(held)
}

func (a Access) projectGrant(permission Permission, projectId primitive.ObjectID, covers func(accessGrant) bool) bool {
	if definition, ok := LookupPermission(permission); !ok || definition.Level != LevelProject || projectId.IsZero() {
		return false
	}
	if a.owner {
		return true
	}
	for _, grant := range a.grants {
		if grant.inProject(permission, projectId) && covers(grant) {
			return true
		}
	}
	return false
}

func (g accessGrant) inProject(permission Permission, projectId primitive.ObjectID) bool {
	permission = CanonicalPagePermission(permission)
	return g.scope.Level == LevelProject && g.scope.ProjectId == projectId && slices.Contains(g.permissions, permission)
}

func (g accessGrant) coversProject(projectId primitive.ObjectID) bool {
	if g.scope.Projects == SelectionAll {
		return true
	}
	return !projectId.IsZero() && slices.Contains(g.scope.ProjectIds, projectId)
}

func (s RoleAssignmentScope) coversDevice(device DeviceRef) bool {
	if s.Resources == SelectionAll {
		return true
	}
	return (device.Key != "" && slices.Contains(s.DeviceKeys, device.Key)) ||
		slices.ContainsFunc(s.SiteIds, func(id primitive.ObjectID) bool { return slices.Contains(device.SiteIds, id) }) ||
		slices.ContainsFunc(s.GroupIds, func(id primitive.ObjectID) bool { return slices.Contains(device.GroupIds, id) })
}

func appendMissing[T comparable](values, more []T) []T {
	var zero T
	for _, value := range more {
		if value != zero && !slices.Contains(values, value) {
			values = append(values, value)
		}
	}
	return values
}
