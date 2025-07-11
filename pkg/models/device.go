package models

import "time"

type DevicesWrapper struct {
	Devices []Device `json:"devices" bson:"devices"`
}

type DeviceWrapper struct {
	Device Device `json:"device" bson:"device,omitempty"`
}

type Device struct {
	Key                  string             `json:"key" bson:"key,omitempty"`
	UserId               string             `json:"user_id" bson:"user_id,omitempty"`
	Enterprise           bool               `json:"enterprise" bson:"enterprise,omitempty"`
	Status               string             `json:"status" bson:"status,omitempty"`
	Color                string             `json:"color" bson:"color,omitempty"`
	Brand                string             `json:"brand" bson:"brand,omitempty"`
	Model                string             `json:"model" bson:"model,omitempty"`
	Description          string             `json:"description" bson:"description,omitempty"`
	LastMaintenance      time.Time          `json:"lastMaintenance" bson:"lastMaintenance,omitempty"`
	InstallationDate     time.Time          `json:"installationDate" bson:"installationDate,omitempty"`
	Mute                 int64              `json:"mute" bson:"mute,omitempty"`
	LatestMediaTimestamp int64              `json:"latestMediaTimestamp" bson:"latestMediaTimestamp,omitempty"`
	LatestMedia          Media              `json:"latestMedia" bson:"latestMedia,omitempty"`
	Analytics            []Heartbeat        `json:"analytics" bson:"analytics,omitempty"`
	Sites                []SiteShort        `json:"sites" bson:"sites,omitempty"`
	Presets              []Preset           `json:"presets" bson:"presets,omitempty"`
	Tours                []Tour             `json:"tours" bson:"tours,omitempty"`
	FeaturePermissions   FeaturePermissions `json:"featurePermissions" bson:"featurePermissions"`
	IsActive             bool               `json:"isActive" bson:"isActive"`
	OnFloorPlans         []string           `json:"onFloorPlans" bson:"onFloorPlans,omitempty"`
	FieldOfView          float64            `json:"fieldOfView" bson:"fieldOfView,omitempty"`
	Name                 string             `json:"name" bson:"name,omitempty"`
	//Location             LocationShort `json:"location" bson:"location,omitempty"` // Has no location anymore, this is done on site level.
}

type Preset struct {
	Name  string  `json:"name" bson:"name"`
	Token string  `json:"token" bson:"token"`
	X     float64 `json:"x" bson:"x"`
	Y     float64 `json:"y" bson:"y"`
	Z     float64 `json:"z" bson:"z"`
}

type ONVIFEvents struct {
	Key       string `json:"key" bson:"key,omitempty"`
	Type      string `json:"type" bson:"type,omitempty"`
	Value     string `json:"value" bson:"value,omitempty"`
	Timestamp int64  `json:"timestamp" bson:"timestamp,omitempty"`
}

type Tour struct {
	Name    string   `json:"name" bson:"name,omitempty"`
	Presets []Preset `json:"presets" bson:"presets,omitempty"`
	Current int      `json:"current" bson:"current,omitempty"`
	Running bool     `json:"running" bson:"running,omitempty"`
	Loop    bool     `json:"loop" bson:"loop,omitempty"`
	Speed   float64  `json:"speed" bson:"speed,omitempty"`
}

type DeviceShort struct {
	Key                  string             `json:"key" bson:"key,omitempty"`
	Enterprise           bool               `json:"enterprise" bson:"enterprise,omitempty"`
	Mute                 int64              `json:"mute" bson:"mute,omitempty"`
	Analytics            []HeartbeatShort   `json:"analytics" bson:"analytics,omitempty"`
	Sites                []SiteShort        `json:"sites" bson:"sites,omitempty"`
	LatestMediaTimestamp int64              `json:"latestMediaTimestamp" bson:"latestMediaTimestamp,omitempty"`
	LatestMedia          *MediaShort        `json:"latestMedia" bson:"latestMedia,omitempty"`
	Presets              []Preset           `json:"presets" bson:"presets,omitempty"`
	Tours                []Tour             `json:"tours" bson:"tours,omitempty"`
	FeaturePermissions   FeaturePermissions `json:"featurePermissions" bson:"featurePermissions"`
	Active               bool               `json:"active" bson:"active"`
	//Location             LocationShort `json:"location" bson:"location,omitempty"` // Has no location anymore, this is done on site level.
}

type DeviceMedia struct {
	Key            string      `json:"key" bson:"key,omitempty"`
	CameraName     string      `json:"camera_name" bson:"camera_name,omitempty"`
	MediaType      string      `json:"media_type" bson:"media_type,omitempty"`
	MediaTime      string      `json:"media_time" bson:"media_time,omitempty"`
	MediaTimestamp int64       `json:"media_timestamp" bson:"media_timestamp,omitempty"`
	MediaUrl       string      `json:"media_url" bson:"media_url,omitempty"`
	SpriteUrl      string      `json:"sprite_url" bson:"sprite_url,omitempty"`
	Sites          []SiteShort `json:"sites" bson:"sites,omitempty"`
}

type Heartbeat struct {
	CloudPublicKey string `json:"cloudpublicKey,omitempty" bson:"cloudpublickey,omitempty"`
	Encrypted      bool   `json:"encrypted,omitempty"`
	EncryptedData  []byte `json:"encryptedData,omitempty"`
	// -----------
	Key              string        `json:"key,omitempty"`
	HubEncryption    string        `json:"hub_encryption,omitempty" bson:"hub_encryption,omitempty"`
	E2EEncryption    string        `json:"e2e_encryption,omitempty" bson:"e2e_encryption,omitempty"`
	Enterprise       bool          `json:"enterprise,omitempty"`
	Hash             string        `json:"hash,omitempty"`
	Version          string        `json:"version,omitempty"`
	Release          string        `json:"release,omitempty"`
	MACs             []string      `json:"mac_list,omitempty" bson:"mac_list,omitempty"`
	IPs              []string      `json:"ip_list,omitempty" bson:"ip_list,omitempty"`
	CpuID            string        `json:"cpuid,omitempty" bson:"cpuid,omitempty"`
	CloudUser        string        `json:"clouduser,omitempty" bson:"clouduser,omitempty"`
	CameraName       string        `json:"cameraname,omitempty" bson:"cameraname,omitempty"`
	CameraType       string        `json:"cameratype,omitempty" bson:"cameratype,omitempty"`
	Architecture     string        `json:"architecture,omitempty"`
	Hostname         string        `json:"hostname,omitempty"`
	FreeMemory       string        `json:"freeMemory,omitempty" bson:"freeMemory,omitempty"`
	TotalMemory      string        `json:"totalMemory,omitempty" bson:"totalMemory,omitempty"`
	UsedMemory       string        `json:"usedMemory,omitempty" bson:"usedMemory,omitempty"`
	ProcessMemory    string        `json:"processMemory,omitempty" bson:"processMemory,omitempty"`
	Kubernetes       bool          `json:"kubernetes,omitempty"`
	Docker           bool          `json:"docker,omitempty"`
	Kios             bool          `json:"kios,omitempty"`
	Raspberrypi      bool          `json:"raspberrypi,omitempty"`
	Uptime           string        `json:"uptime,omitempty" bson:"uptime,omitempty"`
	BootTime         string        `json:"boot_time,omitempty" bson:"boot_time,omitempty"`
	Timestamp        int64         `json:"timestamp" bson:"timestamp"`
	SiteID           string        `json:"siteID" bson:"siteID,omitempty"`
	ONVIF            string        `json:"onvif" bson:"onvif,omitempty"`
	ONVIFZoom        string        `json:"onvif_zoom" bson:"onvif_zoom,omitempty"`
	ONVIFPanTilt     string        `json:"onvif_pantilt" bson:"onvif_pantilt,omitempty"`
	ONVIFPresets     string        `json:"onvif_presets" bson:"onvif_presets,omitempty"`
	ONVIFPresetsList []Preset      `json:"onvif_presets_list" bson:"onvif_presets_list,omitempty"`
	ONVIFEventsList  []ONVIFEvents `json:"onvif_events_list" bson:"onvif_events_list,omitempty"`
	CameraConnected  string        `json:"cameraConnected" bson:"cameraConnected,omitempty"`
	HasBackChannel   string        `json:"hasBackChannel" bson:"hasBackChannel,omitempty"`
	//Board           string   `json:"board,omitempty"`
	//Disk1Size       string   `json:"disk1size,omitempty" bson:"disk1size,omitempty"`
	//Disk3Size       string   `json:"disk3size,omitempty" bson:"disk3size,omitempty"`
	//DiskVDASize     string   `json:"diskvdasize,omitempty" bson:"diskvdasize,omitempty"`
	//NumberOfFiles   string   `json:"numberofiles,omitempty" bson:"numberoffiles,omitempty"`
	//Temperature string `json:"temperature,omitempty"`
	//WifiSSID        string   `json:"wifissid,omitempty" bson:"wifissid,omitempty"`
	//WifiStrength    string   `json:"wifistrength,omitempty" bson:"wifisstrength,omitempty"`
}

type HeartbeatShort struct {
	Key              string        `json:"key,omitempty"`
	Timestamp        int64         `json:"timestamp" bson:"timestamp"`
	Uptime           string        `json:"uptime,omitempty" bson:"uptime,omitempty"`
	Enterprise       bool          `json:"enterprise,omitempty"`
	Version          string        `json:"version,omitempty"`
	Release          string        `json:"release,omitempty"`
	CameraName       string        `json:"cameraname,omitempty" bson:"cameraname,omitempty"`
	ONVIF            string        `json:"onvif" bson:"onvif,omitempty"`
	ONVIFZoom        string        `json:"onvif_zoom" bson:"onvif_zoom,omitempty"`
	ONVIFPanTilt     string        `json:"onvif_pantilt" bson:"onvif_pantilt,omitempty"`
	ONVIFPresets     string        `json:"onvif_presets" bson:"onvif_presets,omitempty"`
	ONVIFPresetsList []Preset      `json:"onvif_presets_list" bson:"onvif_presets_list,omitempty"`
	ONVIFEventsList  []ONVIFEvents `json:"onvif_events_list" bson:"onvif_events_list,omitempty"`
	HubEncryption    string        `json:"hub_encryption,omitempty" bson:"hub_encryption,omitempty"`
	E2EEncryption    string        `json:"e2e_encryption,omitempty" bson:"e2e_encryption,omitempty"`
	CameraConnected  string        `json:"cameraConnected" bson:"cameraConnected,omitempty"`
	HasBackChannel   string        `json:"hasBackChannel" bson:"hasBackChannel,omitempty"`
	FreeMemory       string        `json:"freeMemory,omitempty" bson:"freeMemory,omitempty"`
	TotalMemory      string        `json:"totalMemory,omitempty" bson:"totalMemory,omitempty"`
	UsedMemory       string        `json:"usedMemory,omitempty" bson:"usedMemory,omitempty"`
	ProcessMemory    string        `json:"processMemory,omitempty" bson:"processMemory,omitempty"`
}

type HeartbeatOld struct {
	Key            string `json:"key,omitempty"`
	Enterprise     bool   `json:"enterprise,omitempty"`
	Hash           string `json:"hash,omitempty"`
	Version        string `json:"version,omitempty"`
	CpuID          string `json:"cpuId,omitempty" bson:"cpuid,omitempty"`
	CloudUser      string `json:"cloudUser,omitempty" bson:"clouduser,omitempty"`
	CloudPublicKey string `json:"cloudPublicKey,omitempty" bson:"cloudpublickey,omitempty"`
	CameraName     string `json:"cameraName,omitempty" bson:"cameraname,omitempty"`
	CameraType     string `json:"cameraType,omitempty" bson:"cameratype,omitempty"`
	Kubernetes     bool   `json:"kubernetes,omitempty"`
	Docker         bool   `json:"docker,omitempty"`
	Kios           bool   `json:"kios,omitempty"`
	Raspberrypi    bool   `json:"raspberrypi,omitempty"`
	//Board          string `json:"board,omitempty"`
	//Disk1Size      string `json:"disk1Size,omitempty" bson:"disk1size,omitempty"`
	//Disk3Size      string `json:"disk3Size,omitempty" bson:"disk3size,omitempty"`
	//DiskVDASize    string `json:"diskVDASize,omitempty" bson:"diskvdasize,omitempty"`
	//NumberOfFiles  string `json:"numberOfFiles,omitempty" bson:"numberoffiles,omitempty"`
	//Temperature    string `json:"temperature,omitempty"`
	//WifiSSID       string `json:"wifiSSID,omitempty" bson:"wifissid,omitempty"`
	//WifiStrength   string `json:"wifiStrength,omitempty" bson:"wifisstrength,omitempty"`
	Uptime    string `json:"uptime,omitempty"`
	Timestamp int64  `json:"timestamp" bson:"timestamp"`
	SiteID    string `json:"siteID" bson:"siteID,omitempty"`
	ONVIF     string `json:"onvif" bson:"onvif,omitempty"`
}

type Mute struct {
	Mute int64 `json:"mute" bson:"mute,omitempty"`
}
