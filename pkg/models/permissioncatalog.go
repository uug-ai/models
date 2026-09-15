package models

import "fmt"

// Permission is a stable domain.action identifier granted through a role.
type Permission string

var allPermissions = appendPermissions(mediaPermissions, casePermissions)

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
		return "", fmt.Errorf("unknown permission %q", value)
	}
	return permission, nil
}

func appendPermissions(catalogs ...[]Permission) []Permission {
	var permissions []Permission
	for _, catalog := range catalogs {
		permissions = append(permissions, catalog...)
	}
	return permissions
}
