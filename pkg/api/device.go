package api

import "github.com/uug-ai/models/pkg/models"

// DeviceStatus represents specific status codes for device operations
type DeviceStatus string

const (
	DeviceBindingFailed   DeviceStatus = "device_binding_failed"
	DeviceDuplicateName   DeviceStatus = "device_duplicate_name"
	DeviceMissingInfo     DeviceStatus = "device_missing_info"
	DeviceRetrievalFailed DeviceStatus = "device_retrieval_failed"
	DeviceFound           DeviceStatus = "device_found"
	DeviceNotFound        DeviceStatus = "device_not_found"
	DeviceAddSuccess      DeviceStatus = "device_add_success"
	DeviceAddFailed       DeviceStatus = "device_add_failed"
	DeviceUpdateSuccess   DeviceStatus = "device_update_success"
	DeviceUpdateFailed    DeviceStatus = "device_update_failed"
	DeviceDeleteSuccess   DeviceStatus = "device_delete_success"
	DeviceDeleteFailed    DeviceStatus = "device_delete_failed"
)

// String returns the string representation of the device status
func (ds DeviceStatus) String() string {
	return string(ds)
}

// Into returns the translated string representation of the device status in the specified language
func (ds DeviceStatus) Translate(lang string) string {
	translations := map[string]map[DeviceStatus]string{
		"en": {
			DeviceBindingFailed:   "Device binding failed",
			DeviceDuplicateName:   "Device duplicate name",
			DeviceMissingInfo:     "Device missing information",
			DeviceRetrievalFailed: "Device retrieval failed",
			DeviceFound:           "Device found",
			DeviceNotFound:        "Device not found",
			DeviceAddSuccess:      "Device added successfully",
			DeviceAddFailed:       "Device failed to add",
			DeviceUpdateSuccess:   "Device updated successfully",
			DeviceUpdateFailed:    "Device failed to update",
			DeviceDeleteSuccess:   "Device deleted successfully",
			DeviceDeleteFailed:    "Device failed to delete",
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

type DeviceFilter struct {
	DeviceIds []string `json:"deviceIds,omitempty" bson:"deviceIds,omitempty"`
	Sites     []string `json:"sites,omitempty" bson:"sites,omitempty"`
}

type GetDeviceSummariesRequest struct {
	Filter     *DeviceFilter     `json:"filter,omitempty" bson:"filter,omitempty"`
	Pagination *CursorPagination `json:"pagination,omitempty" bson:"pagination,omitempty"`
}
type GetDeviceSummariesResponse struct {
	Devices []models.DeviceSummary `json:"devices,omitempty" bson:"devices,omitempty"`
}
type GetDeviceSummariesSuccessResponse struct {
	SuccessResponse
	Data GetDeviceSummariesResponse `json:"data,omitempty" bson:"data,omitempty"`
}
type GetDeviceSummariesErrorResponse struct {
	ErrorResponse
}
