package models

import (
	"context"
	"slices"
	"strings"
	"sync"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Built-in roles are defined by the platform. Every organisation stores its own
// read-only copy, upserted by (organisationId, key) and referenced by its stored
// id, so the definitions here are the only source for what a copy contains.
//
// They are the Owner role, the system roles new access is granted with
// (systemrole.go: Organisation manager, Billing admin, Project admin, Editor,
// Viewer), and the legacy roles, which reproduce the legacy guest, editor and
// admin tiers so migrated users keep their access.
// The legacy permission sets are generated from the legacy page tables (see
// LegacyRolePermissions) for each tier's own pages and feature levels, so the
// built-in tiers and custom roles on them come from one definition.
//
// Relationship rules are not role permissions; the resolver applies them next
// to roles. A creator may update and delete what they created when they hold
// the domain's create permission, so a tier that may only change its own sites,
// groups, videowalls or alerts holds create here, not update or delete.

// BuiltInRoles returns the platform-defined roles in stable order, with their
// permissions in catalog order. The roles carry no Id, OrganisationId or
// Audit; the writer sets them on each organisation's copy. Each call returns
// its own copies, which callers may change.
func BuiltInRoles() []Role {
	defined := builtInRoles()
	roles := make([]Role, len(defined))
	for i, role := range defined {
		roles[i] = role.clone()
	}
	return roles
}

// BuiltInRole returns the definition of one built-in role, as a copy the
// caller may change.
func BuiltInRole(key RoleKey) (Role, bool) {
	for _, role := range builtInRoles() {
		if role.Key == key {
			return role.clone(), true
		}
	}
	return Role{}, false
}

// OrganisationBuiltInRoles returns the copies of every built-in role that an
// organisation stores: each with its id from BuiltInRoleId and the
// organisation set, without Audit. Writers upsert them by (organisationId,
// key) when the organisation enters rbac mode, so assignments and invitations
// can name any built-in role by its stored id.
func OrganisationBuiltInRoles(organisationId primitive.ObjectID) []Role {
	roles := BuiltInRoles()
	for index := range roles {
		roles[index].Id = BuiltInRoleId(organisationId, roles[index].Key)
		roles[index].OrganisationId = organisationId
	}
	return roles
}

// BuiltInRoleById returns the release's built-in role that has this id in the
// organisation, with the id and organisation set. A built-in role's id is
// BuiltInRoleId(organisation, key): a stable reference to a definition that
// lives in the release, so nothing has to be stored or read to resolve it.
// Built-in roles are always active.
func BuiltInRoleById(organisationId, id primitive.ObjectID) (Role, bool) {
	if organisationId.IsZero() || id.IsZero() {
		return Role{}, false
	}
	for _, role := range builtInRoles() {
		if BuiltInRoleId(organisationId, role.Key) == id {
			found := role.clone()
			found.Id, found.OrganisationId = id, organisationId
			return found, true
		}
	}
	return Role{}, false
}

// IsBuiltInRoleName reports whether a name is that of a built-in role in
// this release, ignoring case and surrounding space: a custom role may not
// take it. A unique role name index need not cover built-in roles.
func IsBuiltInRoleName(name string) bool {
	name = strings.TrimSpace(name)
	for _, role := range builtInRoles() {
		if strings.EqualFold(role.Name, name) {
			return true
		}
	}
	return false
}

// StoredRoles reads the stored roles with the given ids in an organisation.
type StoredRoles func(ctx context.Context, organisationId primitive.ObjectID, roleIds []primitive.ObjectID) ([]Role, error)

// ResolveRoles returns the roles with the given ids in an organisation. The
// ids of built-in roles are resolved from the release (BuiltInRoleById)
// without reading anything, whether or not the organisation stores a copy
// of them. The other ids are read with stored: the organisation's custom
// roles. stored is not called when every id is a built-in role's.
//
// Every reader of an organisation's roles resolves them this way, so that
// built-in roles need no stored document (RBAC spec 4.14).
func ResolveRoles(ctx context.Context, organisationId primitive.ObjectID, roleIds []primitive.ObjectID, stored StoredRoles) ([]Role, error) {
	roles := make([]Role, 0, len(roleIds))
	var rest []primitive.ObjectID
	for _, id := range roleIds {
		if role, ok := BuiltInRoleById(organisationId, id); ok {
			roles = append(roles, role)
		} else {
			rest = append(rest, id)
		}
	}
	if len(rest) == 0 || stored == nil {
		return roles, nil
	}
	found, err := stored(ctx, organisationId, rest)
	if err != nil {
		return nil, err
	}
	for _, role := range found {
		// A system or legacy document under another id follows the release
		// as well; custom roles are returned as stored.
		roles = append(roles, role.WithCurrentDefinition())
	}
	return roles, nil
}

// Deprecated: built-in roles are not stored (ResolveRoles, RBAC spec 4.14);
// nothing writes this form. Kept until its last caller is gone.
//
// OrganisationBuiltInRoleCopies returns the copies of every built-in role in
// the form that leaves the definition out (Role.StoredCopy): what an
// organisation stores once every reader of its roles applies
// Role.WithCurrentDefinition.
func OrganisationBuiltInRoleCopies(organisationId primitive.ObjectID) []Role {
	roles := OrganisationBuiltInRoles(organisationId)
	for index := range roles {
		roles[index] = roles[index].StoredCopy()
	}
	return roles
}

// Deprecated: built-in roles are not stored (ResolveRoles, RBAC spec 4.14);
// nothing writes this form. Kept until its last caller is gone.
//
// StoredCopy returns a system or legacy role without its definition: only
// the id, organisation, kind, key, name, IsActive and Audit. The description,
// level, grants, permissions and origin are the platform's and come from
// WithCurrentDefinition when the role is read, so a stored copy cannot
// disagree with the release that reads it. The name stays: it is what the
// unique role name index of an organisation is built on, and it tells a
// reader of the database which role the document is.
//
// A reader that does not apply WithCurrentDefinition sees a role that grants
// nothing. Store copies in this form only once every service reading the
// organisation's roles applies it. Custom roles, and built-in keys this
// release does not know, are returned unchanged.
func (r Role) StoredCopy() Role {
	if r.Kind != RoleKindSystem && r.Kind != RoleKindLegacy {
		return r
	}
	if current, ok := BuiltInRole(r.Key); !ok || current.Kind != r.Kind {
		return r
	}
	return Role{
		Id:             r.Id,
		OrganisationId: r.OrganisationId,
		Name:           r.Name,
		Kind:           r.Kind,
		Key:            r.Key,
		IsActive:       r.IsActive,
		Audit:          r.Audit,
	}
}

// IsStoredCopy reports whether a system or legacy role is in the form
// StoredCopy returns: it names its key and leaves the definition out.
func (r Role) IsStoredCopy() bool {
	return (r.Kind == RoleKindSystem || r.Kind == RoleKindLegacy) && r.Key != "" &&
		r.AssignableAt == "" && r.Grants == "" && len(r.Permissions) == 0 && r.Origin == nil && r.Description == ""
}

// WithCurrentDefinition returns a stored system or legacy role with the
// platform's current definition: name, description, level, grants,
// permissions and origin come from BuiltInRole, while the id, organisation,
// IsActive and Audit stay as stored. A copy stored before a models release
// therefore never grants an older permission set, and a copy stored without
// its definition (StoredCopy) gets it here. Custom roles, and built-in keys
// this release does not know, are returned unchanged: such a copy without a
// definition grants nothing.
func (r Role) WithCurrentDefinition() Role {
	if r.Kind != RoleKindSystem && r.Kind != RoleKindLegacy {
		return r
	}
	current, ok := BuiltInRole(r.Key)
	if !ok || current.Kind != r.Kind {
		return r
	}
	current.Id = r.Id
	current.OrganisationId = r.OrganisationId
	current.IsActive = r.IsActive
	current.Audit = r.Audit
	return current
}

// builtInRoles are the definitions, built once per process: the legacy page
// tables are fixed, and generating the permission sets is costly, while
// access resolution looks built-in roles up on every request. Never hand
// these out without clone.
var builtInRoles = sync.OnceValue(defineBuiltInRoles)

// clone copies a role's slice and pointer fields, so changing the copy never
// changes the definition it came from.
func (r Role) clone() Role {
	r.Permissions = slices.Clone(r.Permissions)
	if r.Origin != nil {
		origin := *r.Origin
		r.Origin = &origin
	}
	return r
}

func defineBuiltInRoles() []Role {
	guest, member := legacyTierPermissions(BaseRoleGuest)
	editor, _ := legacyTierPermissions(BaseRoleEditor)
	admin, adminOrganisation := legacyTierPermissions(BaseRoleAdmin)
	roles := []Role{
		{
			Name:         "Owner",
			Description:  "Every permission in every current and future project.",
			Kind:         RoleKindSystem,
			Key:          RoleKeyOwner,
			AssignableAt: LevelOrganisation,
			Grants:       SelectionAll,
			IsActive:     true,
		},
		legacyRole(RoleKeyLegacyGuest, BaseRoleGuest, LevelProject, "Guest (legacy)",
			"The legacy guest tier inside a project.", guest),
		legacyRole(RoleKeyLegacyEditor, BaseRoleEditor, LevelProject, "Editor (legacy)",
			"The legacy editor tier inside a project.", editor),
		legacyRole(RoleKeyLegacyAdmin, BaseRoleAdmin, LevelProject, "Admin (legacy)",
			"The legacy admin tier inside a project.", admin),
		legacyRole(RoleKeyLegacyMemberOrganisation, "", LevelOrganisation, "Member (legacy), organisation",
			"What legacy guests and editors could see of the organisation.", member),
		legacyRole(RoleKeyLegacyAdminOrganisation, BaseRoleAdmin, LevelOrganisation, "Admin (legacy), organisation",
			"The legacy admin tier's organisation administration.", adminOrganisation),
	}
	return append(roles[:1], append(systemRoles(), roles[1:]...)...)
}

func legacyRole(key RoleKey, baseRole BaseRole, level Level, name, description string, permissions []Permission) Role {
	role := Role{
		Name:         name,
		Description:  description,
		Kind:         RoleKindLegacy,
		Key:          key,
		AssignableAt: level,
		Grants:       SelectionSelected,
		Permissions:  permissions,
		IsActive:     true,
	}
	if baseRole != "" {
		role.Origin = &RoleOrigin{BaseRole: baseRole}
	}
	return role
}

// legacyTierPermissions returns a built-in tier's project and organisation
// permissions, from its own pages and feature levels.
func legacyTierPermissions(tier BaseRole) (project, organisation []Permission) {
	project, organisation, _ = LegacyRolePermissions(tier, legacyTierPages[tier], legacyTierFeatures[tier])
	return project, organisation
}

// catalogOrder returns a copy of the permissions sorted in catalog order.
func catalogOrder(permissions []Permission) []Permission {
	sorted := slices.Clone(permissions)
	slices.SortFunc(sorted, func(a, b Permission) int {
		return permissionIndex[a] - permissionIndex[b]
	})
	return sorted
}
