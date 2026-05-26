package models

// RuntimeConfig contains runtime configuration values that must not be
// exposed publicly (e.g. via the SPA's /assets/env.js). The hub-api
// exposes these to authenticated clients via the /runtime/config
// endpoint after login.
type RuntimeConfig struct {
	MqttUsername string `json:"mqttUsername" bson:"mqttUsername"`
	MqttPassword string `json:"mqttPassword" bson:"mqttPassword"`
	TurnUsername string `json:"turnUsername" bson:"turnUsername"`
	TurnPassword string `json:"turnPassword" bson:"turnPassword"`
}
