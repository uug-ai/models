package models

// Strategy represents the strategy pattern used for a specific device.
// This allows to achieve specific behaviors or configurations based on the defined strategy.
type Strategy struct {
	Id   string `json:"id" bson:"_id" example:"strategy-123" required:"true"` // Unique identifier for the strategy
	Name string `json:"name" bson:"name" example:"Default Strategy" required:"true"` // Name of the strategy
	DeviceId string `json:"deviceId" bson:"deviceId" example:"device-456" required:"true"` // Associated device identifier
}