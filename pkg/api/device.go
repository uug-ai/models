package api

import "github.com/uug-ai/models/pkg/models"

// DeviceStatus represents specific status codes for device operations
type DeviceStatus string

const (
	DeviceBindingFailed    DeviceStatus = "device_binding_failed"
	DeviceDuplicateName    DeviceStatus = "device_duplicate_name"
	DeviceMissingInfo      DeviceStatus = "device_missing_info"
	DeviceRetrievalSuccess DeviceStatus = "device_retrieval_success"
	DeviceRetrievalFailed  DeviceStatus = "device_retrieval_failed"
	DeviceFound            DeviceStatus = "device_found"
	DeviceNotFound         DeviceStatus = "device_not_found"
	DeviceAddSuccess       DeviceStatus = "device_add_success"
	DeviceAddFailed        DeviceStatus = "device_add_failed"
	DeviceUpdateSuccess    DeviceStatus = "device_update_success"
	DeviceUpdateFailed     DeviceStatus = "device_update_failed"
	DeviceDeleteSuccess    DeviceStatus = "device_delete_success"
	DeviceDeleteFailed     DeviceStatus = "device_delete_failed"

	GetDeviceIdsForMarkersBindingFailed DeviceStatus = "get_device_ids_for_markers_binding_failed"
	GetDeviceIdsForMarkersError         DeviceStatus = "get_device_ids_for_markers_error"
	GetDeviceIdsForMarkersSuccess       DeviceStatus = "get_device_ids_for_markers_success"
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

			GetDeviceIdsForMarkersBindingFailed: "Get device IDs for markers binding failed",
			GetDeviceIdsForMarkersError:         "Get device IDs for markers error",
			GetDeviceIdsForMarkersSuccess:       "Get device IDs for markers succeeded",
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
	DeviceIds []*string `json:"deviceIds,omitempty" bson:"deviceIds,omitempty"`
	Name      *string   `json:"name,omitempty" bson:"name,omitempty"`
	Sites     []*string `json:"sites,omitempty" bson:"sites,omitempty"`
	Markers   []*string `json:"markers,omitempty" bson:"markers,omitempty"`
}

type GetDeviceOptionsRequest struct {
	Filter     *DeviceFilter     `json:"filter,omitempty" bson:"filter,omitempty"`
	Pagination *CursorPagination `json:"pagination,omitempty" bson:"pagination,omitempty"`
}
type GetDeviceOptionsResponse struct {
	Devices []models.DeviceOption `json:"devices,omitempty" bson:"devices,omitempty"`
}
type GetDeviceOptionsSuccessResponse struct {
	SuccessResponse
	Data GetDeviceOptionsResponse `json:"data,omitempty" bson:"data,omitempty"`
}
type GetDeviceOptionsErrorResponse struct {
	ErrorResponse
}

// GetDeviceIdsForMarkers
type GetDeviceIdsForMarkersRequest struct {
	MarkerNames []string `json:"markerIds,omitempty" bson:"markerIds,omitempty"`
}
type GetDeviceIdsForMarkersResponse struct {
	DeviceIds []string `json:"deviceIds,omitempty" bson:"deviceIds,omitempty"`
}
type GetDeviceIdsForMarkersSuccessResponse struct {
	SuccessResponse
	Data GetDeviceIdsForMarkersResponse `json:"data,omitempty" bson:"data,omitempty"`
}
type GetDeviceIdsForMarkersErrorResponse struct {
	ErrorResponse
}
