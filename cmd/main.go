package main

import (
	"fmt"

	"github.com/uug-ai/models/pkg/models"
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
}

// GetMedia godoc
// @Summary Get a media item
// @Description Get a media item by ID
// @Tags media
// @Accept json
// @Produce json
// @Param id path string true "Media ID"
// @Success 200 {object} models.Media
// @Failure 400 {object} models.ErrorResponse
// @Router /media/{id} [get]
func GetMedia() {}

// CreateMedia godoc
// @Summary Create a new media item
// @Description Create a new media item
// @Tags media
// @Accept json
// @Produce json
// @Param media body models.Media true "Media object"
// @Success 201 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /media [post]
func CreateMedia() {}

// Dummy endpoints to ensure all models are included in OpenAPI spec
// These endpoints exist only to force swag to generate schemas for all models

// GetAPIMetadata godoc
// @Summary Get APIMetadata (schema generation only)
// @Description Internal endpoint used only to ensure APIMetadata schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.APIMetadata
// @Router /internal/apimetadata [get]
func GetAPIMetadata() {}

// GetDevicesWrapper godoc
// @Summary Get DevicesWrapper (schema generation only)
// @Description Internal endpoint used only to ensure DevicesWrapper schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.DevicesWrapper
// @Router /internal/deviceswrapper [get]
func GetDevicesWrapper() {}

// GetDeviceWrapper godoc
// @Summary Get DeviceWrapper (schema generation only)
// @Description Internal endpoint used only to ensure DeviceWrapper schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.DeviceWrapper
// @Router /internal/devicewrapper [get]
func GetDeviceWrapper() {}

// GetDevice godoc
// @Summary Get Device (schema generation only)
// @Description Internal endpoint used only to ensure Device schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.Device
// @Router /internal/device [get]
func GetDevice() {}

// GetPreset godoc
// @Summary Get Preset (schema generation only)
// @Description Internal endpoint used only to ensure Preset schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.Preset
// @Router /internal/preset [get]
func GetPreset() {}

// GetONVIFEvents godoc
// @Summary Get ONVIFEvents (schema generation only)
// @Description Internal endpoint used only to ensure ONVIFEvents schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.ONVIFEvents
// @Router /internal/onvifevents [get]
func GetONVIFEvents() {}

// GetTour godoc
// @Summary Get Tour (schema generation only)
// @Description Internal endpoint used only to ensure Tour schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.Tour
// @Router /internal/tour [get]
func GetTour() {}

// GetDeviceShort godoc
// @Summary Get DeviceShort (schema generation only)
// @Description Internal endpoint used only to ensure DeviceShort schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.DeviceShort
// @Router /internal/deviceshort [get]
func GetDeviceShort() {}

// GetDeviceMedia godoc
// @Summary Get DeviceMedia (schema generation only)
// @Description Internal endpoint used only to ensure DeviceMedia schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.DeviceMedia
// @Router /internal/devicemedia [get]
func GetDeviceMedia() {}

// GetHeartbeat godoc
// @Summary Get Heartbeat (schema generation only)
// @Description Internal endpoint used only to ensure Heartbeat schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.Heartbeat
// @Router /internal/heartbeat [get]
func GetHeartbeat() {}

// GetHeartbeatShort godoc
// @Summary Get HeartbeatShort (schema generation only)
// @Description Internal endpoint used only to ensure HeartbeatShort schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.HeartbeatShort
// @Router /internal/heartbeatshort [get]
func GetHeartbeatShort() {}

// GetHeartbeatOld godoc
// @Summary Get HeartbeatOld (schema generation only)
// @Description Internal endpoint used only to ensure HeartbeatOld schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.HeartbeatOld
// @Router /internal/heartbeatold [get]
func GetHeartbeatOld() {}

// GetMute godoc
// @Summary Get Mute (schema generation only)
// @Description Internal endpoint used only to ensure Mute schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.Mute
// @Router /internal/mute [get]
func GetMute() {}

// GetLocation godoc
// @Summary Get Location (schema generation only)
// @Description Internal endpoint used only to ensure Location schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.Location
// @Router /internal/location [get]
func GetLocation() {}

// GetLocationShort godoc
// @Summary Get LocationShort (schema generation only)
// @Description Internal endpoint used only to ensure LocationShort schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.LocationShort
// @Router /internal/locationshort [get]
func GetLocationShort() {}

// GetLocationGeometry godoc
// @Summary Get LocationGeometry (schema generation only)
// @Description Internal endpoint used only to ensure LocationGeometry schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.LocationGeometry
// @Router /internal/locationgeometry [get]
func GetLocationGeometry() {}

// GetLocationGeometryLocation godoc
// @Summary Get LocationGeometryLocation (schema generation only)
// @Description Internal endpoint used only to ensure LocationGeometryLocation schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.LocationGeometryLocation
// @Router /internal/locationgeometrylocation [get]
func GetLocationGeometryLocation() {}

// GetMediaMetadata godoc
// @Summary Get MediaMetadata (schema generation only)
// @Description Internal endpoint used only to ensure MediaMetadata schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.MediaMetadata
// @Router /internal/mediametadata [get]
func GetMediaMetadata() {}

// GetMediaShort godoc
// @Summary Get MediaShort (schema generation only)
// @Description Internal endpoint used only to ensure MediaShort schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.MediaShort
// @Router /internal/mediashort [get]
func GetMediaShort() {}

// GetTimelineValue godoc
// @Summary Get TimelineValue (schema generation only)
// @Description Internal endpoint used only to ensure TimelineValue schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.TimelineValue
// @Router /internal/timelinevalue [get]
func GetTimelineValue() {}

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

// GetSiteWrapper godoc
// @Summary Get SiteWrapper (schema generation only)
// @Description Internal endpoint used only to ensure SiteWrapper schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.SiteWrapper
// @Router /internal/sitewrapper [get]
func GetSiteWrapper() {}

// GetSite godoc
// @Summary Get Site (schema generation only)
// @Description Internal endpoint used only to ensure Site schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.Site
// @Router /internal/site [get]
func GetSite() {}

// GetSiteShort godoc
// @Summary Get SiteShort (schema generation only)
// @Description Internal endpoint used only to ensure SiteShort schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.SiteShort
// @Router /internal/siteshort [get]
func GetSiteShort() {}

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
