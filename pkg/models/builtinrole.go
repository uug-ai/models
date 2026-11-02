package models

import "slices"

// Built-in roles are defined by the platform. Every organisation stores its own
// read-only copy, upserted by (organisationId, key) and referenced by its stored
// id, so the definitions here are the only source for what a copy contains.
//
// For now they are the Owner role and the legacy roles, which reproduce the
// legacy guest, editor and admin tiers so migrated users keep their access.
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
// Audit; the writer sets them on each organisation's copy.
func BuiltInRoles() []Role {
	guest, member := legacyTierPermissions(BaseRoleGuest)
	editor, _ := legacyTierPermissions(BaseRoleEditor)
	admin, adminOrganisation := legacyTierPermissions(BaseRoleAdmin)
	return []Role{
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
}

// BuiltInRole returns the definition of one built-in role.
func BuiltInRole(key RoleKey) (Role, bool) {
	for _, role := range BuiltInRoles() {
		if role.Key == key {
			return role, true
		}
	}
	return Role{}, false
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
