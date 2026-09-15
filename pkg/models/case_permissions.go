package models

const (
	PermissionCasesRead   Permission = "cases.read"
	PermissionCasesCreate Permission = "cases.create"
	PermissionCasesUpdate Permission = "cases.update"
	PermissionCasesShare  Permission = "cases.share"
	PermissionCasesDelete Permission = "cases.delete"
)

var casePermissions = []Permission{
	PermissionCasesRead,
	PermissionCasesCreate,
	PermissionCasesUpdate,
	PermissionCasesShare,
	PermissionCasesDelete,
}

// CasePermissions returns the canonical case permission catalog.
func CasePermissions() []Permission {
	return append([]Permission(nil), casePermissions...)
}
