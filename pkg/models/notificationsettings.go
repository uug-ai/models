package models

type NotificationSettings struct {
	Detections Detections `json:"detections" bson:"detections,omitempty"`
	Devices    Devices    `json:"devices" bson:"devices,omitempty"`
	Highupload Highupload `json:"highupload" bson:"highupload,omitempty"`
}

type NotificationUpdate struct {
	NotificationSettings NotificationUpdatePayload `json:"notificationSettings" bson:"notificationSettings,omitempty"`
}

type NotificationUpdatePayload struct {
	Type    string      `json:"type" bson:"type,omitempty"`
	Payload interface{} `json:"payload" bson:"payload,omitempty"`
}

type Detections struct {
	Enabled            bool        `json:"enabled" bson:"enabled,omitempty"`
	ChannelsAll        bool        `json:"channelsAll" bson:"channelsAll,omitempty"`
	ChannelsList       []string    `json:"channelsList" bson:"channelsList,omitempty"`
	DevicesAll         bool        `json:"devicesAll" bson:"devicesAll,omitempty"`
	DevicesList        []DeviceKey `json:"devicesList" bson:"devicesList,omitempty"`
	ClassificationAll  bool        `json:"classificationAll" bson:"classificationAll,omitempty"`
	ClassificationList []string    `json:"classificationList" bson:"classificationList,omitempty"`
	TimeAdvanced       bool        `json:"timeAdvanced" bson:"timeAdvanced,omitempty"`
	TimeRange1Max      int32       `json:"timeRange1Max" bson:"timeRange1Max"`
	TimeRange1Min      int32       `json:"timeRange1Min" bson:"timeRange1Min"`
	TimeRange2Max      int32       `json:"timeRange2Max" bson:"timeRange2Max"`
	TimeRange2Min      int32       `json:"timeRange2Min" bson:"timeRange2Min"`
}

type Devices struct {
	Enabled      bool        `json:"enabled" bson:"enabled,omitempty"`
	ChannelsAll  bool        `json:"channelsAll" bson:"channelsAll,omitempty"`
	ChannelsList []string    `json:"channelsList" bson:"channelsList,omitempty"`
	DevicesAll   bool        `json:"devicesAll" bson:"devicesAll,omitempty"`
	DevicesList  []DeviceKey `json:"devicesList" bson:"devicesList,omitempty"`
	Duration     int32       `json:"duration" bson:"duration,omitempty"`
}

type Highupload struct {
	Enabled      bool        `json:"enabled" bson:"enabled,omitempty"`
	ChannelsAll  bool        `json:"channelsAll" bson:"channelsAll,omitempty"`
	ChannelsList []string    `json:"channelsList" bson:"channelsList,omitempty"`
	DevicesAll   bool        `json:"devicesAll" bson:"devicesAll,omitempty"`
	DevicesList  []DeviceKey `json:"devicesList" bson:"devicesList,omitempty"`
	Requests     int32       `json:"requests" bson:"requests,omitempty"`
}

type RegionPoint struct {
	X float64 `json:"x" bson:"x,omitempty"`
	Y float64 `json:"y" bson:"y,omitempty"`
}
