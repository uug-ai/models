package models

import (
	"errors"
	"fmt"
)

// Permission is a stable domain.action identifier granted through a role.
type Permission string

var (
	ErrPermissionUnknown   = errors.New("unknown permission")
	ErrPermissionDuplicate = errors.New("duplicate permission")
)

var allPermissions = appendPermissions(
	mediaPermissions,
	casePermissions,
	workflowPermissions,
	workflowRunPermissions,
)

var permissionSet = func() map[Permission]struct{} {
	permissions := make(map[Permission]struct{}, len(allPermissions))
	for _, permission := range allPermissions {
		permissions[permission] = struct{}{}
	}
	return permissions
}()

// AllPermissions returns the canonical permission catalog in stable order.
func AllPermissions() []Permission {
	return append([]Permission(nil), allPermissions...)
}

// ParsePermission validates and returns a canonical permission identifier.
func ParsePermission(value string) (Permission, error) {
	permission := Permission(value)
	if _, ok := permissionSet[permission]; !ok {
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

func appendPermissions(catalogs ...[]Permission) []Permission {
	var permissions []Permission
	for _, catalog := range catalogs {
		permissions = append(permissions, catalog...)
	}
	return permissions
}
