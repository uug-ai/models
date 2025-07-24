package api

// DeviceStatus represents specific status codes for device operations
type DeviceStatus string

const (
	DeviceBindingFailed DeviceStatus = "device_binding_failed"
	DeviceDuplicateName DeviceStatus = "device_duplicate_name"
	DeviceMissingInfo   DeviceStatus = "device_missing_info"
	DeviceFound         DeviceStatus = "device_found"
	DeviceNotFound      DeviceStatus = "device_not_found"
	DeviceAddSuccess    DeviceStatus = "device_add_success"
	DeviceAddFailed     DeviceStatus = "device_add_failed"
	DeviceUpdateSuccess DeviceStatus = "device_update_success"
	DeviceUpdateFailed  DeviceStatus = "device_update_failed"
	DeviceDeleteSuccess DeviceStatus = "device_delete_success"
	DeviceDeleteFailed  DeviceStatus = "device_delete_failed"
)

// String returns the string representation of the device status
func (ds DeviceStatus) String() string {
	return string(ds)
}

// Into returns the translated string representation of the device status in the specified language
func (ds DeviceStatus) Translate(lang string) string {
	translations := map[string]map[DeviceStatus]string{
		"en": {
			DeviceBindingFailed: "Device binding failed",
			DeviceDuplicateName: "Device duplicate name",
			DeviceMissingInfo:   "Device missing information",
			DeviceFound:         "Device found",
			DeviceNotFound:      "Device not found",
			DeviceAddSuccess:    "Device added successfully",
			DeviceAddFailed:     "Device failed to add",
			DeviceUpdateSuccess: "Device updated successfully",
			DeviceUpdateFailed:  "Device failed to update",
			DeviceDeleteSuccess: "Device deleted successfully",
			DeviceDeleteFailed:  "Device failed to delete",
		},
	}

	if langTranslations, exists := translations[lang]; exists {
		if translation, exists := langTranslations[ds]; exists {
			return translation
		}
	}

	// Default to English if language not found or translation doesn't exist
	if enTranslations, exists := translations["en"]; exists {
		if translation, exists := enTranslations[ds]; exists {
			return translation
		}
	}

	// Fallback to the string representation
	return ds.String()
}
