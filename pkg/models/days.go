package models

const PermissionDaysRead Permission = "days.read"

var dayPermissions = []Permission{
	PermissionDaysRead,
}

// DayPermissions returns the canonical days permission catalog.
func DayPermissions() []Permission {
	return append([]Permission(nil), dayPermissions...)
}
