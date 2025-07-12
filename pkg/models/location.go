package models

type Location struct {
	Latitude  float64 `json:"latitude,omitempty" bson:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty" bson:"longitude,omitempty"`
	Altitude  float64 `json:"altitude,omitempty" bson:"altitude,omitempty"` // Altitude in meters
	Address   string  `json:"address" bson:"address,omitempty"`             // e.g. "123 Main St, Anytown, USA"
}
