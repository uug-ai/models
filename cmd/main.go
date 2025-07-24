package main

import (
	"fmt"

	"github.com/uug-ai/models/pkg/models"
	"github.com/uug-ai/models/pkg/api"
)

// @title Models API
// @version 1.0
// @description API for media models and related types
// @host localhost
// @BasePath /

func main() {
	// This main function exists to allow swag to generate OpenAPI spec
	// from the models in pkg/models
	fmt.Println("Models API")
	
	// Keep the import valid - models are used in the API endpoint annotations below
	_ = models.Media{}
	_ = api.SuccessResponse{}
}

// GetMedia godoc
// @Summary Get a media item
// @Description Get a media item by ID
// @Tags media
// @Accept json
// @Produce json
// @Param id path string true "Media ID"
// @Success 200 {object} api.SuccessResponse
// @Failure 400 {object} api.ErrorResponse
// @Router /media/{id} [get]
func GetMedia() {}

// CreateMedia godoc
// @Summary Create a new media item
// @Description Create a new media item
// @Tags media
// @Accept json
// @Produce json
// @Param media body models.Media true "Media object"
// @Success 201 {object} api.SuccessResponse
// @Failure 400 {object} api.ErrorResponse
// @Router /media [post]
func CreateMedia() {}

// Dummy endpoints to ensure all models are included in OpenAPI spec
// These endpoints exist only to force swag to generate schemas for all models

// GetAudit godoc
// @Summary Get Audit (schema generation only)
// @Description Internal endpoint used only to ensure Audit schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.Audit
// @Router /internal/audit [get]
func GetAudit() {}

// GetComment godoc
// @Summary Get Comment (schema generation only)
// @Description Internal endpoint used only to ensure Comment schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.Comment
// @Router /internal/comment [get]
func GetComment() {}

// GetDevice godoc
// @Summary Get Device (schema generation only)
// @Description Internal endpoint used only to ensure Device schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.Device
// @Router /internal/device [get]
func GetDevice() {}

// GetDeviceMetadata godoc
// @Summary Get DeviceMetadata (schema generation only)
// @Description Internal endpoint used only to ensure DeviceMetadata schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.DeviceMetadata
// @Router /internal/devicemetadata [get]
func GetDeviceMetadata() {}

// GetDeviceCameraMetadata godoc
// @Summary Get DeviceCameraMetadata (schema generation only)
// @Description Internal endpoint used only to ensure DeviceCameraMetadata schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.DeviceCameraMetadata
// @Router /internal/devicecamerametadata [get]
func GetDeviceCameraMetadata() {}

// GetDeviceLocationMetadata godoc
// @Summary Get DeviceLocationMetadata (schema generation only)
// @Description Internal endpoint used only to ensure DeviceLocationMetadata schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.DeviceLocationMetadata
// @Router /internal/devicelocationmetadata [get]
func GetDeviceLocationMetadata() {}

// GetDeviceFeaturePermissions godoc
// @Summary Get DeviceFeaturePermissions (schema generation only)
// @Description Internal endpoint used only to ensure DeviceFeaturePermissions schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.DeviceFeaturePermissions
// @Router /internal/devicefeaturepermissions [get]
func GetDeviceFeaturePermissions() {}

// GetDeviceAtRuntimeMetadata godoc
// @Summary Get DeviceAtRuntimeMetadata (schema generation only)
// @Description Internal endpoint used only to ensure DeviceAtRuntimeMetadata schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.DeviceAtRuntimeMetadata
// @Router /internal/deviceatruntimemetadata [get]
func GetDeviceAtRuntimeMetadata() {}

// GetFloorPlan godoc
// @Summary Get FloorPlan (schema generation only)
// @Description Internal endpoint used only to ensure FloorPlan schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.FloorPlan
// @Router /internal/floorplan [get]
func GetFloorPlan() {}

// GetDevicePlacement godoc
// @Summary Get DevicePlacement (schema generation only)
// @Description Internal endpoint used only to ensure DevicePlacement schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.DevicePlacement
// @Router /internal/deviceplacement [get]
func GetDevicePlacement() {}

// GetIO godoc
// @Summary Get IO (schema generation only)
// @Description Internal endpoint used only to ensure IO schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.IO
// @Router /internal/io [get]
func GetIO() {}

// GetLocation godoc
// @Summary Get Location (schema generation only)
// @Description Internal endpoint used only to ensure Location schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.Location
// @Router /internal/location [get]
func GetLocation() {}

// GetMarker godoc
// @Summary Get Marker (schema generation only)
// @Description Internal endpoint used only to ensure Marker schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.Marker
// @Router /internal/marker [get]
func GetMarker() {}

// GetMarkerMetadata godoc
// @Summary Get MarkerMetadata (schema generation only)
// @Description Internal endpoint used only to ensure MarkerMetadata schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.MarkerMetadata
// @Router /internal/markermetadata [get]
func GetMarkerMetadata() {}

// GetMediaMetadata godoc
// @Summary Get MediaMetadata (schema generation only)
// @Description Internal endpoint used only to ensure MediaMetadata schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.MediaMetadata
// @Router /internal/mediametadata [get]
func GetMediaMetadata() {}

// GetMediaAtRuntimeMetadata godoc
// @Summary Get MediaAtRuntimeMetadata (schema generation only)
// @Description Internal endpoint used only to ensure MediaAtRuntimeMetadata schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.MediaAtRuntimeMetadata
// @Router /internal/mediaatruntimemetadata [get]
func GetMediaAtRuntimeMetadata() {}

// GetPreset godoc
// @Summary Get Preset (schema generation only)
// @Description Internal endpoint used only to ensure Preset schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.Preset
// @Router /internal/preset [get]
func GetPreset() {}

// GetTour godoc
// @Summary Get Tour (schema generation only)
// @Description Internal endpoint used only to ensure Tour schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.Tour
// @Router /internal/tour [get]
func GetTour() {}

// GetRole godoc
// @Summary Get Role (schema generation only)
// @Description Internal endpoint used only to ensure Role schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.Role
// @Router /internal/role [get]
func GetRole() {}

// GetFeaturePermissions godoc
// @Summary Get FeaturePermissions (schema generation only)
// @Description Internal endpoint used only to ensure FeaturePermissions schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.FeaturePermissions
// @Router /internal/featurepermissions [get]
func GetFeaturePermissions() {}

// GetTimeWindow godoc
// @Summary Get TimeWindow (schema generation only)
// @Description Internal endpoint used only to ensure TimeWindow schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.TimeWindow
// @Router /internal/timewindow [get]
func GetTimeWindow() {}

// GetSite godoc
// @Summary Get Site (schema generation only)
// @Description Internal endpoint used only to ensure Site schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.Site
// @Router /internal/site [get]
func GetSite() {}

// GetSiteMetadata godoc
// @Summary Get SiteMetadata (schema generation only)
// @Description Internal endpoint used only to ensure SiteMetadata schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.SiteMetadata
// @Router /internal/sitemetadata [get]
func GetSiteMetadata() {}

// GetSiteLocationMetadata godoc
// @Summary Get SiteLocationMetadata (schema generation only)
// @Description Internal endpoint used only to ensure SiteLocationMetadata schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.SiteLocationMetadata
// @Router /internal/sitelocationmetadata [get]
func GetSiteLocationMetadata() {}

// GetSynchronize godoc
// @Summary Get Synchronize (schema generation only)
// @Description Internal endpoint used only to ensure Synchronize schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.Synchronize
// @Router /internal/synchronize [get]
func GetSynchronize() {}

// GetSynchronizeEvent godoc
// @Summary Get SynchronizeEvent (schema generation only)
// @Description Internal endpoint used only to ensure SynchronizeEvent schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.SynchronizeEvent
// @Router /internal/synchronizeevent [get]
func GetSynchronizeEvent() {}
