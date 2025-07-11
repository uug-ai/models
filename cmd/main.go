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

	// Reference the models so they are included in the spec
	var _ models.Media
	var _ models.MediaMetadata
	var _ models.APIMetadata
	var _ models.ErrorResponse
	var _ models.SuccessResponse
	var _ models.Tester
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

// GetTest godoc
// @Summary Get a test item
// @Description Get a test item by ID
// @Tags test
// @Accept json
// @Produce json
// @Param id path string true "Test ID"
// @Success 200 {object} models.Tester
// @Failure 400 {object} models.ErrorResponse
// @Router /test/{id} [get]
func GetTest() {}
