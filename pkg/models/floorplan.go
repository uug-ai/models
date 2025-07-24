package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type FloorPlan struct {
	Id               primitive.ObjectID `json:"id" bson:"_id,omitempty,omitempty"`
	Name             string             `json:"name" bson:"name,omitempty"`                         // Name of the floor plan
	Image            string             `json:"image" bson:"image,omitempty"`                       // Base64 encoded image of the floor plan
	Width            int                `json:"width" bson:"width,omitempty"`                       // Dimensions of the floor plan in pixels
	Height           int                `json:"height" bson:"height,omitempty"`                     // Dimensions of the floor plan in pixels
	DevicePlacements []DevicePlacement  `json:"devicePlacements" bson:"devicePlacements,omitempty"` // List of devices placed on the floor plan

	// Audit information
	Audit *Audit `json:"audit,omitempty" bson:"audit,omitempty"`
}

type DevicePlacement struct {
	Id       primitive.ObjectID `json:"id" bson:"_id"`
	DeviceId string             `json:"deviceId" bson:"deviceId"` // ID of the device being placed

	FabricKey        string  `json:"fabricKey" bson:"fabricKey"`
	X                float64 `json:"x" bson:"x"`                                         // Absolute X coordinate
	Y                float64 `json:"y" bson:"y"`                                         // Absolute Y coordinate
	RelativeX        float64 `json:"relativeX" bson:"relativeX"`                         // X relative to canvas width (0 to 1)
	RelativeY        float64 `json:"relativeY" bson:"relativeY"`                         // Y relative to canvas height (0 to 1)
	Radius           float64 `json:"radius" bson:"radius,omitempty"`                     // Radius of the device placement circle
	FieldOfView      float64 `json:"fieldOfView" bson:"fieldOfView,omitempty"`           // Field of view in degrees, if applicable
	SliceStartAngle  float64 `json:"sliceStartAngle" bson:"sliceStartAngle,omitempty"`   // Start angle of the slice in degrees, if applicable
	SliceEndAngle    float64 `json:"sliceEndAngle" bson:"sliceEndAngle,omitempty"`       // End angle of the slice in degrees, if applicable
	SliceMiddleAngle float64 `json:"sliceMiddleAngle" bson:"sliceMiddleAngle,omitempty"` // Middle angle of the slice in degrees, if applicable

	Color string `json:"color" bson:"color,omitempty"` // Color to represent the device on the floor plan
	Icon  string `json:"icon" bson:"icon,omitempty"`   // Icon to represent the device on the floor plan

	// Audit information
	Audit *Audit `json:"audit,omitempty" bson:"audit,omitempty"`
}
