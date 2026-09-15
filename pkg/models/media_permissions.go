package models

const (
	PermissionMediaRead   Permission = "media.read"
	PermissionMediaUpdate Permission = "media.update"
	PermissionMediaExport Permission = "media.export"
	PermissionMediaShare  Permission = "media.share"
	PermissionMediaRedact Permission = "media.redact"
	PermissionMediaDelete Permission = "media.delete"
)

var mediaPermissions = []Permission{
	PermissionMediaRead,
	PermissionMediaUpdate,
	PermissionMediaExport,
	PermissionMediaShare,
	PermissionMediaRedact,
	PermissionMediaDelete,
}

// MediaPermissions returns the canonical media permission catalog.
func MediaPermissions() []Permission {
	return append([]Permission(nil), mediaPermissions...)
}
