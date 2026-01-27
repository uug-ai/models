package models

type Channels struct {
	Pushbullet Pushbullet `json:"pushbullet" bson:"pushbullet,omitempty"`
	Email      Email      `json:"email" bson:"email,omitempty"`
	Slack      Slack      `json:"slack" bson:"slack,omitempty"`
	Telegram   Telegram   `json:"telegram" bson:"telegram,omitempty"`
	Alexa      Alexa      `json:"alexa" bson:"alexa,omitempty"`
	Webhook    Webhook    `json:"webhook" bson:"webhook,omitempty"`
	Ifttt      Ifttt      `json:"ifttt" bson:"ifttt,omitempty"`
	Pushover   Pushover   `json:"pushover" bson:"pushover,omitempty"`
	Sms        Sms        `json:"sms" bson:"sms,omitempty"`
}

type ChannelUpdate struct {
	Channels ChannelUpdatePayload `json:"channels" bson:"channels,omitempty"`
}

type ChannelUpdatePayload struct {
	Channel string      `json:"channel" bson:"channel,omitempty"`
	Payload interface{} `json:"payload" bson:"payload,omitempty"`
}

type Pushbullet struct {
	Enabled bool   `json:"enabled" bson:"enabled,omitempty"`
	Valid   bool   `json:"valid" bson:"valid,omitempty"`
	Apikey  string `json:"apikey" bson:"apikey,omitempty"`
}

type Email struct {
	Enabled bool   `json:"enabled" bson:"enabled,omitempty"`
	Valid   bool   `json:"valid" bson:"valid,omitempty"`
	Address string `json:"address" bson:"address,omitempty"`
}

type Slack struct {
	Enabled  bool   `json:"enabled" bson:"enabled,omitempty"`
	Valid    bool   `json:"valid" bson:"valid,omitempty"`
	Apikey   string `json:"apikey" bson:"apikey,omitempty"`
	Channel  string `json:"channel" bson:"channel,omitempty"`
	Hook     string `json:"hook" bson:"hook,omitempty"`
	Username string `json:"username" bson:"username,omitempty"`
}

type Telegram struct {
	Enabled bool   `json:"enabled" bson:"enabled,omitempty"`
	Valid   bool   `json:"valid" bson:"valid,omitempty"`
	Token   string `json:"token" bson:"token,omitempty"`
	Channel string `json:"channel" bson:"channel,omitempty"`
}

type Alexa struct {
	Enabled    bool   `json:"enabled" bson:"enabled,omitempty"`
	Valid      bool   `json:"valid" bson:"valid,omitempty"`
	Accesscode string `json:"accesscode" bson:"accesscode,omitempty"`
}

type Webhook struct {
	Enabled bool   `json:"enabled" bson:"enabled,omitempty"`
	Valid   bool   `json:"valid" bson:"valid,omitempty"`
	Url     string `json:"url" bson:"url,omitempty"`
}

type Ifttt struct {
	Enabled bool   `json:"enabled" bson:"enabled,omitempty"`
	Valid   bool   `json:"valid" bson:"valid,omitempty"`
	Token   string `json:"token" bson:"token,omitempty"`
}

type Pushover struct {
	Enabled bool   `json:"enabled" bson:"enabled,omitempty"`
	Valid   bool   `json:"valid" bson:"valid,omitempty"`
	Apikey  string `json:"apikey" bson:"apikey,omitempty"`
	Sendto  string `json:"sendto" bson:"sendto,omitempty"`
}

type Sms struct {
	Enabled    bool   `json:"enabled" bson:"enabled,omitempty"`
	Valid      bool   `json:"valid" bson:"valid,omitempty"`
	Accountsid string `json:"accountsid" bson:"accountsid,omitempty"`
	AuthToken  string `json:"authtoken" bson:"authtoken,omitempty"`
	From       string `json:"from" bson:"from,omitempty"`
	To         string `json:"to" bson:"to,omitempty"`
}

type ChannelTest struct {
	Events []string        `json:"events" bson:"events,omitempty"`
	Data   ChannelTestData `json:"data" bson:"data,omitempty"`
}

type ChannelTestData struct {
	UserId  string `json:"userId" bson:"userId,omitempty"`
	Channel string `json:"channel" bson:"channel,omitempty"`
}
