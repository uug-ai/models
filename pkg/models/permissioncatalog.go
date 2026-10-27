package models

import (
	"errors"
	"fmt"
	"slices"
)

// Permission is a stable domain.action identifier granted through a role.
type Permission string

const PermissionApplication Permission = "application"

// PermissionScope is the level at which a permission takes effect.
type PermissionScope string

const (
	// PermissionScopeOrganisation permissions govern the tenant itself (members,
	// roles, billing, settings). They are only effective through an
	// organisation-level role assignment.
	PermissionScopeOrganisation PermissionScope = "organisation"
	// PermissionScopeProject permissions govern operational resources. They are
	// effective within the projects and resources a role assignment covers.
	PermissionScopeProject PermissionScope = "project"
)

// PermissionDefinition describes one permission in the canonical catalog.
type PermissionDefinition struct {
	Permission Permission      `json:"permission"`
	Scope      PermissionScope `json:"scope"`
	// Implies lists the permissions that holding this one also grants, e.g. a
	// graded capability implies its lower grades. Implied permissions always
	// share the same scope.
	Implies []Permission `json:"implies,omitempty"`
}

var (
	ErrPermissionUnknown   = errors.New("unknown permission")
	ErrPermissionDuplicate = errors.New("duplicate permission")
	ErrPermissionInvalid   = errors.New("invalid permission definition")
)

// permissionGroup registers one domain's permissions in the catalog. Every
// permission of a domain shares the group's scope.
type permissionGroup struct {
	scope       PermissionScope
	permissions []Permission
	implies     map[Permission][]Permission
}

var permissionGroups = []permissionGroup{
	{scope: PermissionScopeProject, permissions: markerPermissions},
	{scope: PermissionScopeProject, permissions: mediaPermissions},
	{scope: PermissionScopeProject, permissions: casePermissions},
	{scope: PermissionScopeProject, permissions: workflowPermissions},
	{scope: PermissionScopeProject, permissions: workflowRunPermissions},
	{scope: PermissionScopeProject, permissions: dayPermissions},
	{scope: PermissionScopeProject, permissions: channelPermissions},
}

var permissionDefinitions = definePermissions(permissionGroups)

var allPermissions = func() []Permission {
	permissions := make([]Permission, len(permissionDefinitions))
	for index, definition := range permissionDefinitions {
		permissions[index] = definition.Permission
	}
	return permissions
}()

var permissionIndex = indexPermissions(permissionDefinitions)

// AllPermissions returns the canonical permission catalog in stable order.
func AllPermissions() []Permission {
	return append([]Permission(nil), allPermissions...)
}

// PermissionDefinitions returns every catalog definition in stable order.
func PermissionDefinitions() []PermissionDefinition {
	definitions := make([]PermissionDefinition, len(permissionDefinitions))
	for index, definition := range permissionDefinitions {
		definition.Implies = append([]Permission(nil), definition.Implies...)
		definitions[index] = definition
	}
	return definitions
}

// LookupPermission returns the catalog definition of a permission.
func LookupPermission(permission Permission) (PermissionDefinition, bool) {
	index, ok := permissionIndex[permission]
	if !ok {
		return PermissionDefinition{}, false
	}
	definition := permissionDefinitions[index]
	definition.Implies = append([]Permission(nil), definition.Implies...)
	return definition, true
}

// ParsePermission validates and returns a canonical permission identifier.
func ParsePermission(value string) (Permission, error) {
	permission := Permission(value)
	if _, ok := permissionIndex[permission]; !ok {
		return "", fmt.Errorf("%w %q", ErrPermissionUnknown, value)
	}
	return permission, nil
}

// ValidatePermissions verifies that every permission is canonical and appears
// at most once.
func ValidatePermissions(permissions []Permission) error {
	seen := make(map[Permission]struct{}, len(permissions))
	for _, permission := range permissions {
		if _, err := ParsePermission(string(permission)); err != nil {
			return err
		}
		if _, exists := seen[permission]; exists {
			return fmt.Errorf("%w %q", ErrPermissionDuplicate, permission)
		}
		seen[permission] = struct{}{}
	}
	return nil
}

// ExpandPermissions returns the given permissions together with everything
// they imply, in catalog order. Unknown permissions are dropped: they grant
// nothing.
func ExpandPermissions(permissions []Permission) []Permission {
	return expandPermissions(permissionDefinitions, permissionIndex, permissions)
}

func expandPermissions(definitions []PermissionDefinition, index map[Permission]int, permissions []Permission) []Permission {
	granted := make([]bool, len(definitions))
	pending := append([]Permission(nil), permissions...)
	for len(pending) > 0 {
		permission := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		position, ok := index[permission]
		if !ok || granted[position] {
			continue
		}
		granted[position] = true
		pending = append(pending, definitions[position].Implies...)
	}

	var expanded []Permission
	for position, definition := range definitions {
		if granted[position] {
			expanded = append(expanded, definition.Permission)
		}
	}
	return expanded
}

func definePermissions(groups []permissionGroup) []PermissionDefinition {
	var definitions []PermissionDefinition
	for _, group := range groups {
		for _, permission := range group.permissions {
			definitions = append(definitions, PermissionDefinition{
				Permission: permission,
				Scope:      group.scope,
				Implies:    append([]Permission(nil), group.implies[permission]...),
			})
		}
	}
	return definitions
}

func indexPermissions(definitions []PermissionDefinition) map[Permission]int {
	index := make(map[Permission]int, len(definitions))
	for position, definition := range definitions {
		index[definition.Permission] = position
	}
	return index
}

// validatePermissionGroups checks the catalog invariants: identifiers are
// non-empty and unique, scopes are known, implications only reference the
// group's own permissions, stay within one scope, and contain no cycles.
func validatePermissionGroups(groups []permissionGroup) error {
	definitions := definePermissions(groups)
	index := make(map[Permission]int, len(definitions))
	for position, definition := range definitions {
		if definition.Permission == "" {
			return fmt.Errorf("%w: empty permission", ErrPermissionInvalid)
		}
		if definition.Scope != PermissionScopeOrganisation && definition.Scope != PermissionScopeProject {
			return fmt.Errorf("%w: %q has unknown scope %q", ErrPermissionInvalid, definition.Permission, definition.Scope)
		}
		if _, exists := index[definition.Permission]; exists {
			return fmt.Errorf("%w %q", ErrPermissionDuplicate, definition.Permission)
		}
		index[definition.Permission] = position
	}

	for _, group := range groups {
		for permission := range group.implies {
			if !slices.Contains(group.permissions, permission) {
				return fmt.Errorf("%w: implication source %q is not a permission of its group", ErrPermissionInvalid, permission)
			}
		}
	}

	for _, definition := range definitions {
		for _, implied := range definition.Implies {
			position, ok := index[implied]
			if !ok {
				return fmt.Errorf("%w: %q implies unknown %q", ErrPermissionInvalid, definition.Permission, implied)
			}
			if definitions[position].Scope != definition.Scope {
				return fmt.Errorf("%w: %q implies %q across scopes", ErrPermissionInvalid, definition.Permission, implied)
			}
		}
	}

	const (
		unvisited = iota
		visiting
		done
	)
	state := make([]int, len(definitions))
	var visit func(position int) error
	visit = func(position int) error {
		switch state[position] {
		case visiting:
			return fmt.Errorf("%w: implication cycle through %q", ErrPermissionInvalid, definitions[position].Permission)
		case done:
			return nil
		}
		state[position] = visiting
		for _, implied := range definitions[position].Implies {
			if err := visit(index[implied]); err != nil {
				return err
			}
		}
		state[position] = done
		return nil
	}
	for position := range definitions {
		if err := visit(position); err != nil {
			return err
		}
	}
	return nil
}
