package models

type PermissionLevel int

const (
	PermissionNone  PermissionLevel = 1
	PermissionRead  PermissionLevel = 2
	PermissionWrite PermissionLevel = 3
	PermissionAdmin PermissionLevel = 4

	// PTZ permission levels
	PTZZoom    PermissionLevel = 2
	PTZPanTilt PermissionLevel = 3
	PTZFull    PermissionLevel = 4
	PTZPresets PermissionLevel = 5

	// Liveview permission levels
	LiveViewSD PermissionLevel = 2
	LiveViewHD PermissionLevel = 3

	// RemoteConfig permission levels
	RemoteConfigAgent PermissionLevel = 2

	// IO permission levels
	IORead  PermissionLevel = 2
	IOWrite PermissionLevel = 3

	// FloorPlans permission levels
	FloorPlansView  PermissionLevel = 2
	FloorPlansEdit  PermissionLevel = 3
	FloorPlansAdmin PermissionLevel = 4
)

type DeviceFeaturePermissions struct {
	PTZ          PermissionLevel `json:"ptz" bson:"ptz"`
	Liveview     PermissionLevel `json:"liveview" bson:"liveview"`
	RemoteConfig PermissionLevel `json:"remoteConfig" bson:"remoteConfig"`
	IO           PermissionLevel `json:"io" bson:"io"`
	FloorPlans   PermissionLevel `json:"floorPlans" bson:"floorPlans"`
	// Talk..
	// ...
}
