package models

import (
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	PermissionRolesRead   Permission = "roles.read"
	PermissionRolesCreate Permission = "roles.create"
	PermissionRolesUpdate Permission = "roles.update"
	PermissionRolesDelete Permission = "roles.delete"
)

var rolePermissions = []Permission{
	PermissionRolesRead,
	PermissionRolesCreate,
	PermissionRolesUpdate,
	PermissionRolesDelete,
}

// RolePermissions returns the canonical role permission catalog.
func RolePermissions() []Permission {
	return append([]Permission(nil), rolePermissions...)
}

// AccessLevel defines the level of access for a feature or resource in the
// legacy profile permissions. Roles use catalog permissions instead.
type AccessLevel int

const (
	AccessLevelNone  AccessLevel = 0 // No access
	AccessLevelRead  AccessLevel = 1 // Read-only access
	AccessLevelWrite AccessLevel = 2 // Read and write access
	AccessLevelAdmin AccessLevel = 3 // Full administrative access
)

// FeaturePermissions is the legacy per-feature access level set, still served
// in the legacy profile permissions. Roles use catalog permissions instead.
type FeaturePermissions struct {
	PTZ                AccessLevel `json:"ptz" bson:"ptz"`                               // 0=none, 1=read, 2=write, 3=admin
	Liveview           AccessLevel `json:"liveview" bson:"liveview"`                     // 0=none, 1=read, 2=write, 3=admin
	RemoteConfig       AccessLevel `json:"remoteConfig" bson:"remoteConfig"`             // 0=none, 1=read, 2=write, 3=admin
	IO                 AccessLevel `json:"io" bson:"io"`                                 // 0=none, 1=read, 2=write, 3=admin
	FloorPlans         AccessLevel `json:"floorPlans" bson:"floorPlans"`                 // 0=none, 1=read, 2=write, 3=admin
	Playback           AccessLevel `json:"playback" bson:"playback"`                     // 0=none, 1=read, 2=write, 3=admin
	Export             AccessLevel `json:"export" bson:"export"`                         // 0=none, 1=read, 2=write, 3=admin
	Markers            AccessLevel `json:"markers" bson:"markers"`                       // 0=none, 1=read, 2=write, 3=admin
	Alerts             AccessLevel `json:"alerts" bson:"alerts"`                         // 0=none, 1=read, 2=write, 3=admin
	Users              AccessLevel `json:"users" bson:"users"`                           // 0=none, 1=read, 2=write, 3=admin
	Devices            AccessLevel `json:"devices" bson:"devices"`                       // 0=none, 1=read, 2=write, 3=admin
	Sites              AccessLevel `json:"sites" bson:"sites"`                           // 0=none, 1=read, 2=write, 3=admin
	Groups             AccessLevel `json:"groups" bson:"groups"`                         // 0=none, 1=read, 2=write, 3=admin
	Projects           AccessLevel `json:"projects" bson:"projects"`                     // 0=none, 1=read, 2=write, 3=admin
	Roles              AccessLevel `json:"roles" bson:"roles"`                           // 0=none, 1=read, 2=write, 3=admin
	Settings           AccessLevel `json:"settings" bson:"settings"`                     // 0=none, 1=read, 2=write, 3=admin
	DownloadRecordings AccessLevel `json:"downloadRecordings" bson:"downloadRecordings"` // 0=none, 1=read, 2=write, 3=admin
}

// BaseRole enumerates the legacy built-in role tiers. In the RBAC model a tier
// is provenance only (see RoleOrigin): it grants nothing and gates nothing.
type BaseRole string

const (
	BaseRoleGuest  BaseRole = "guest"
	BaseRoleEditor BaseRole = "editor"
	BaseRoleAdmin  BaseRole = "admin"
	BaseRoleOwner  BaseRole = "owner"
)

// RoleKind distinguishes who defines a role.
type RoleKind string

const (
	// RoleKindSystem roles are defined by the platform and read-only.
	RoleKindSystem RoleKind = "system"
	// RoleKindLegacy roles are defined by the platform and reproduce a legacy
	// built-in tier exactly, so migrated users keep their access. They are
	// read-only and can be retired once nothing is assigned to them.
	RoleKindLegacy RoleKind = "legacy"
	// RoleKindCustom roles are defined by the organisation, including legacy
	// custom roles copied by the migration.
	RoleKindCustom RoleKind = "custom"
)

// RoleKey identifies a system or legacy role. It is the same in every
// organisation, while each organisation stores its own copy of the role.
type RoleKey string

const (
	// RoleKeyOwner is the Owner role: every permission in every current and
	// future project (Grants=all). An organisation has at least one owner,
	// and Organisation.OwnerId names the primary one.
	RoleKeyOwner               RoleKey = "owner"
	RoleKeyOrganisationManager RoleKey = "organisation-manager"
	RoleKeyBillingAdmin        RoleKey = "billing-admin"
	RoleKeyProjectAdmin        RoleKey = "project-admin"
	RoleKeyEditor              RoleKey = "editor"
	RoleKeyViewer              RoleKey = "viewer"

	RoleKeyLegacyGuest  RoleKey = "legacy-guest"
	RoleKeyLegacyEditor RoleKey = "legacy-editor"
	// RoleKeyLegacyAdmin is the project part of the legacy admin tier and
	// RoleKeyLegacyAdminOrganisation its organisation part.
	RoleKeyLegacyAdmin             RoleKey = "legacy-admin"
	RoleKeyLegacyAdminOrganisation RoleKey = "legacy-admin-organisation"
)

// RoleOrigin records what a migrated role was created from. It is provenance
// only and grants nothing.
type RoleOrigin struct {
	BaseRole     BaseRole           `json:"baseRole,omitempty" bson:"baseRole,omitempty"`         // Legacy tier the role reproduces or extended
	LegacyRoleId primitive.ObjectID `json:"legacyRoleId,omitempty" bson:"legacyRoleId,omitempty"` // Legacy custom role it was copied from
}

var ErrRoleInvalid = errors.New("invalid role")

// Role is a named set of catalog permissions owned by an organisation. A role
// never carries resources or time windows: those belong to the assignment.
//
// A role is assignable at one level only. An organisation role holds
// organisation permissions and grants nothing inside projects; a project role
// holds project permissions and is assigned per project. The Owner role
// (RoleKeyOwner) is the one exception and spans every project.
type Role struct {
	Id             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	OrganisationId primitive.ObjectID `json:"organisationId" bson:"organisationId,omitempty"` // Organisation this role belongs to
	Name           string             `json:"name" bson:"name,omitempty"`
	Description    string             `json:"description" bson:"description,omitempty"`

	Kind         RoleKind        `json:"kind" bson:"kind,omitempty"`
	Key          RoleKey         `json:"key,omitempty" bson:"key,omitempty"` // System and legacy roles only; unique per organisation
	AssignableAt PermissionScope `json:"assignableAt" bson:"assignableAt,omitempty"`
	// Grants says whether the role grants every permission (all; the Owner
	// role only, with no Permissions) or exactly the listed Permissions
	// (selected; an empty list grants nothing). It is required: the zero value
	// is invalid, so an empty list never means "everything".
	Grants Selection `json:"grants" bson:"grants,omitempty"`
	// Permissions is the canonical set of grants carried by a Grants=selected
	// role, all of the AssignableAt level. Implied permissions (see
	// ExpandPermissions) are not stored.
	Permissions []Permission `json:"permissions,omitempty" bson:"permissions,omitempty"`
	Origin      *RoleOrigin  `json:"origin,omitempty" bson:"origin,omitempty"`
	IsActive    bool         `json:"isActive" bson:"isActive"`

	Audit Audit `json:"audit" bson:"audit"`
}

// ValidatePermissions verifies that the role contains only unique permissions
// from the canonical catalog.
func (r Role) ValidatePermissions() error {
	return ValidatePermissions(r.Permissions)
}

// IsOwner reports whether this is the organisation's Owner role.
func (r Role) IsOwner() bool {
	return r.Kind == RoleKindSystem && r.Key == RoleKeyOwner
}

// GrantsAll reports whether the role grants every permission, including ones
// added to the catalog later.
func (r Role) GrantsAll() bool {
	return r.Grants == SelectionAll
}

// Validate checks the role's own invariants: a known kind with a key exactly
// when the platform defines it, a known level, an explicit grants selection,
// and only catalog permissions of that level. Only the Owner role, a system
// organisation role, grants all.
func (r Role) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("%w: empty name", ErrRoleInvalid)
	}
	switch r.Kind {
	case RoleKindSystem, RoleKindLegacy:
		if r.Key == "" {
			return fmt.Errorf("%w: %s role without key", ErrRoleInvalid, r.Kind)
		}
	case RoleKindCustom:
		if r.Key != "" {
			return fmt.Errorf("%w: custom role with key %q", ErrRoleInvalid, r.Key)
		}
	default:
		return fmt.Errorf("%w: unknown kind %q", ErrRoleInvalid, r.Kind)
	}
	if r.AssignableAt != PermissionScopeOrganisation && r.AssignableAt != PermissionScopeProject {
		return fmt.Errorf("%w: unknown level %q", ErrRoleInvalid, r.AssignableAt)
	}
	switch r.Grants {
	case SelectionAll:
		if r.Key != RoleKeyOwner {
			return fmt.Errorf("%w: only the owner role grants all permissions", ErrRoleInvalid)
		}
		if len(r.Permissions) > 0 {
			return fmt.Errorf("%w: role granting all permissions lists permissions", ErrRoleInvalid)
		}
	case SelectionSelected:
		if r.Key == RoleKeyOwner {
			return fmt.Errorf("%w: owner role must grant all permissions", ErrRoleInvalid)
		}
	default:
		return fmt.Errorf("%w: unknown grants %q", ErrRoleInvalid, r.Grants)
	}
	if r.Key == RoleKeyOwner {
		if r.Kind != RoleKindSystem || r.AssignableAt != PermissionScopeOrganisation {
			return fmt.Errorf("%w: owner must be a system organisation role", ErrRoleInvalid)
		}
		return nil
	}
	if err := r.ValidatePermissions(); err != nil {
		return err
	}
	for _, permission := range r.Permissions {
		definition, _ := LookupPermission(permission)
		if definition.Scope != r.AssignableAt {
			return fmt.Errorf("%w: %s permission %q on %s role", ErrRoleInvalid, definition.Scope, permission, r.AssignableAt)
		}
	}
	return nil
}

// RequiresAllProjects reports whether the role is Owner or holds a tenant-wide
// permission, in which case it can only be assigned to cover all projects.
func (r Role) RequiresAllProjects() bool {
	if r.IsOwner() || r.GrantsAll() {
		return true
	}
	for _, permission := range r.Permissions {
		if definition, ok := LookupPermission(permission); ok && definition.TenantWide {
			return true
		}
	}
	return false
}
