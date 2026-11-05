package models

import (
	"context"
	"fmt"
	"slices"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AuthorizationMode says, per organisation, which data decides access (spec
// 7.2). It is stored on the organisation and read by every service that
// decides access, so they switch at the same moment.
type AuthorizationMode string

const (
	// AuthorizationModeLegacy: legacy data decides (ResolveLegacyAccess);
	// stored roles and assignments are not consulted.
	AuthorizationModeLegacy AuthorizationMode = "legacy"
	// AuthorizationModeShadow: legacy data still decides; access is also
	// resolved from the stored documents (ResolveAccess) and differences are
	// recorded (CompareAccess).
	AuthorizationModeShadow AuthorizationMode = "shadow"
	// AuthorizationModeRBAC: the stored documents decide. A missing
	// membership or a resolution error grants nothing; there is no fallback
	// to legacy data. Rolling back means setting the mode to legacy.
	AuthorizationModeRBAC AuthorizationMode = "rbac"
)

// IsValid reports whether the mode is one of the defined modes. The empty
// mode is valid and means legacy.
func (m AuthorizationMode) IsValid() bool {
	switch m {
	case "", AuthorizationModeLegacy, AuthorizationModeShadow, AuthorizationModeRBAC:
		return true
	}
	return false
}

// EffectiveAuthorizationMode is the organisation's mode as services apply it:
// a missing mode is legacy. An unknown value is also treated as legacy, which
// is how every organisation behaved before modes existed, and fails
// Organisation.ValidateAuthorizationMode so writers never store one.
func (o Organisation) EffectiveAuthorizationMode() AuthorizationMode {
	switch o.AuthorizationMode {
	case AuthorizationModeShadow, AuthorizationModeRBAC:
		return o.AuthorizationMode
	}
	return AuthorizationModeLegacy
}

// ValidateAuthorizationMode checks that the stored mode is a defined one.
func (o Organisation) ValidateAuthorizationMode() error {
	if !o.AuthorizationMode.IsValid() {
		return fmt.Errorf("unknown authorization mode %q", o.AuthorizationMode)
	}
	return nil
}

// AccessStore reads the stored documents access is resolved from: one
// member's membership and team memberships, the assignments to them and the
// roles those assignments grant. Every read is limited to one organisation.
type AccessStore interface {
	// Membership returns the user's membership of the organisation; found is
	// false when there is none.
	Membership(ctx context.Context, organisationId, userId primitive.ObjectID) (membership OrganisationUser, found bool, err error)
	// TeamMemberships returns the user's TeamMember documents in the
	// organisation.
	TeamMemberships(ctx context.Context, organisationId, userId primitive.ObjectID) ([]TeamMember, error)
	// Assignments returns the organisation's assignments to any of the
	// principals, active or not (NewAccess decides which are in force).
	Assignments(ctx context.Context, organisationId primitive.ObjectID, principals []Principal) ([]RoleAssignment, error)
	// Roles returns the organisation's roles with the given ids that exist.
	Roles(ctx context.Context, organisationId primitive.ObjectID, roleIds []primitive.ObjectID) ([]Role, error)
}

// ResolveAccess returns a member's access in an organisation at the given
// instant from the stored documents, the source of truth in rbac mode. A
// user without a membership has no access; any read error grants nothing and
// is returned, so the caller denies (spec 7.2: rbac fails closed).
func ResolveAccess(ctx context.Context, store AccessStore, organisationId, userId primitive.ObjectID, at time.Time) (Access, error) {
	if organisationId.IsZero() || userId.IsZero() {
		return Access{}, fmt.Errorf("resolve access: missing organisation or user")
	}
	membership, found, err := store.Membership(ctx, organisationId, userId)
	if err != nil {
		return Access{}, err
	}
	if !found || !membership.isEffectivelyActiveAt(at) {
		return Access{}, nil
	}
	teams, err := store.TeamMemberships(ctx, organisationId, userId)
	if err != nil {
		return Access{}, err
	}
	principals := []Principal{{Type: PrincipalTypeUser, Id: userId}}
	for _, teamId := range teamIdsOf(membership, teams) {
		principals = append(principals, Principal{Type: PrincipalTypeTeam, Id: teamId})
	}
	assignments, err := store.Assignments(ctx, organisationId, principals)
	if err != nil {
		return Access{}, err
	}
	if len(assignments) == 0 {
		return Access{}, nil
	}
	var roleIds []primitive.ObjectID
	for _, assignment := range assignments {
		if !slices.Contains(roleIds, assignment.RoleId) {
			roleIds = append(roleIds, assignment.RoleId)
		}
	}
	roles, err := store.Roles(ctx, organisationId, roleIds)
	if err != nil {
		return Access{}, err
	}
	return NewAccess(membership, teams, roles, assignments, at), nil
}

// AccessDifference is one way two resolutions of the same member differ, for
// the shadow storage comparison: Wider when the first (stored) access allows
// what the second (legacy) does not, a cutover blocker.
type AccessDifference struct {
	// Kind is "owner", "permission" or "devices".
	Kind       string
	Project    primitive.ObjectID // Zero for organisation-wide differences
	Permission Permission         // Empty for "owner"
	Wider      bool
}

// CompareAccess lists how stored access differs from legacy access for the
// given projects: ownership, the permissions held in each project (and on all
// projects), and for every device permission both hold, the devices it is
// held on. Device selections are compared as selected (all, or site, group
// and device ids), not expanded; both come from the same ids when the
// mapping is faithful. Nil means they agree.
func CompareAccess(stored, legacy Access, projects []primitive.ObjectID) []AccessDifference {
	var differences []AccessDifference
	if stored.owner != legacy.owner {
		differences = append(differences, AccessDifference{Kind: "owner", Wider: stored.owner})
	}
	for _, project := range append([]primitive.ObjectID{primitive.NilObjectID}, projects...) {
		storedPermissions, legacyPermissions := stored.Permissions(project), legacy.Permissions(project)
		for _, permission := range storedPermissions {
			if !slices.Contains(legacyPermissions, permission) {
				differences = append(differences, AccessDifference{Kind: "permission", Project: project, Permission: permission, Wider: true})
			}
		}
		for _, permission := range legacyPermissions {
			if !slices.Contains(storedPermissions, permission) {
				differences = append(differences, AccessDifference{Kind: "permission", Project: project, Permission: permission})
			}
		}
		if project.IsZero() {
			continue
		}
		for _, permission := range deviceScoped(storedPermissions) {
			if !slices.Contains(legacyPermissions, permission) {
				continue
			}
			if wider, narrower := compareSelections(stored.Devices(permission, project), legacy.Devices(permission, project)); wider || narrower {
				if wider {
					differences = append(differences, AccessDifference{Kind: "devices", Project: project, Permission: permission, Wider: true})
				}
				if narrower {
					differences = append(differences, AccessDifference{Kind: "devices", Project: project, Permission: permission})
				}
			}
		}
	}
	return differences
}

// compareSelections reports whether a selects something b does not (wider)
// and b something a does not (narrower).
func compareSelections(a, b DeviceSelection) (wider, narrower bool) {
	switch {
	case a.All && b.All:
		return false, false
	case a.All:
		return true, false
	case b.All:
		return false, true
	}
	return !subset(a.SiteIds, b.SiteIds) || !subset(a.GroupIds, b.GroupIds) || !subset(a.DeviceKeys, b.DeviceKeys),
		!subset(b.SiteIds, a.SiteIds) || !subset(b.GroupIds, a.GroupIds) || !subset(b.DeviceKeys, a.DeviceKeys)
}

func subset[T comparable](values, of []T) bool {
	for _, value := range values {
		if !slices.Contains(of, value) {
			return false
		}
	}
	return true
}
