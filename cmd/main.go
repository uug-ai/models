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

// GetMediaMetadata godoc
// @Summary Get MediaMetadata (schema generation only)
// @Description Internal endpoint used only to ensure MediaMetadata schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.MediaMetadata
// @Router /internal/mediametadata [get]
func GetMediaMetadata() {}

// GetTimelineValue godoc
// @Summary Get TimelineValue (schema generation only)
// @Description Internal endpoint used only to ensure TimelineValue schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.TimelineValue
// @Router /internal/timelinevalue [get]
func GetTimelineValue() {}

// GetTestFlowStruct godoc
// @Summary Get TestFlowStruct (schema generation only)
// @Description Internal endpoint used only to ensure TestFlowStruct schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} models.TestFlowStruct
// @Router /internal/testflowstruct [get]
func GetTestFlowStruct() {}
