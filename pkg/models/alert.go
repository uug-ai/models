package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type CustomAlert struct {
	Id                  primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Enabled             bool               `json:"enabled" bson:"enabled"`
	Title               string             `json:"title" bson:"title,omitempty"`
	Description         string             `json:"description" bson:"description,omitempty"`
	ChannelsAll         bool               `json:"channelsAll" bson:"channelsAll"`
	ChannelsList        []string           `json:"channelsList" bson:"channelsList"`
	DevicesAll          bool               `json:"devicesAll" bson:"devicesAll"`
	DevicesList         []DeviceKey        `json:"devicesList" bson:"devicesList"`
	CountingDevicesAll  bool               `json:"countingDevicesAll" bson:"countingDevicesAll"`
	CountingDevicesList []DeviceKey        `json:"countingDevicesList" bson:"countingDevicesList"`
	ClassificationAll   bool               `json:"classificationAll" bson:"classificationAll"`
	ClassificationList  []string           `json:"classificationList" bson:"classificationList"`
	TimeAdvanced        bool               `json:"timeAdvanced" bson:"timeAdvanced"`
	TimeRange1Max       int32              `json:"timeRange1Max" bson:"timeRange1Max"`
	TimeRange1Min       int32              `json:"timeRange1Min" bson:"timeRange1Min"`
	TimeRange2Max       int32              `json:"timeRange2Max" bson:"timeRange2Max"`
	TimeRange2Min       int32              `json:"timeRange2Min" bson:"timeRange2Min"`
	UserId              string             `json:"user_id" bson:"user_id,omitempty"`
	MasterUserId        string             `json:"master_user_id" bson:"master_user_id,omitempty"`
	EmailEmail          string             `json:"email_email" bson:"email_email,omitempty"`
	SlackHook           string             `json:"slack_hook" bson:"slack_hook,omitempty"`
	SlackBotname        string             `json:"slack_botname" bson:"slack_botname,omitempty"`
	PushbulletApikey    string             `json:"pushbullet_apikey" bson:"pushbullet_apikey,omitempty"`
	TelegramToken       string             `json:"telegram_token" bson:"telegram_token,omitempty"`
	TelegramChannel     string             `json:"telegram_channel" bson:"telegram_channel,omitempty"`
	AlexaToken          string             `json:"alexa_token" bson:"alexa_token,omitempty"`
	WebhookUrl          string             `json:"webhook_url" bson:"webhook_url,omitempty"`
	IftttToken          string             `json:"ifttt_token" bson:"ifttt_token,omitempty"`
	SMSAccountsid       string             `json:"sms_accountsid" bson:"sms_accountsid,omitempty"`
	SMSAuthtoken        string             `json:"sms_authtoken" bson:"sms_authtoken,omitempty"`
	SMSTelfrom          string             `json:"sms_telfrom" bson:"sms_telfrom,omitempty"`
	SMSTelto            string             `json:"sms_telto" bson:"sms_telto,omitempty"`
	PushoverApikey      string             `json:"pushover_apikey" bson:"pushover_apikey,omitempty"`
	PushoverSendto      string             `json:"pushover_sendto" bson:"pushover_sendto,omitempty"`
	MotionRegions       []Region           `json:"motionRegions" bson:"motionRegions,omitempty"`
	CountingLines       []Region           `json:"countingLines" bson:"countingLines,omitempty"`
	InputList           []string           `json:"inputList" bson:"inputList,omitempty"`
	InputsAND           bool               `json:"inputsAND" bson:"inputsAND"`
	OutputList          []string           `json:"outputList" bson:"outputList,omitempty"`
	Features            *AlertFeatures     `json:"features,omitempty" bson:"features,omitempty"`
}

type AlertFeatures struct {
	CreateMarker bool `json:"createMarker,omitempty" bson:"createMarker,omitempty"`
}
