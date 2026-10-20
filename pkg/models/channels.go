package models

import (
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	PermissionChannelsRead   Permission = "channels.read"
	PermissionChannelsCreate Permission = "channels.create"
	PermissionChannelsUpdate Permission = "channels.update"
	PermissionChannelsDelete Permission = "channels.delete"
)

var channelPermissions = []Permission{
	PermissionChannelsRead,
	PermissionChannelsCreate,
	PermissionChannelsUpdate,
	PermissionChannelsDelete,
}

// ChannelPermissions returns the canonical channel permission catalog.
// Reading a channel never exposes its Config; testing a channel requires
// PermissionChannelsUpdate because it delivers to the configured recipient.
func ChannelPermissions() []Permission {
	return append([]Permission(nil), channelPermissions...)
}

// ChannelType identifies the provider a Channel delivers notifications through.
type ChannelType string

const (
	ChannelTypeEmail      ChannelType = "email"
	ChannelTypeSlack      ChannelType = "slack"
	ChannelTypeWebhook    ChannelType = "webhook"
	ChannelTypePushbullet ChannelType = "pushbullet"
	ChannelTypeTelegram   ChannelType = "telegram"
	ChannelTypeAlexa      ChannelType = "alexa"
	ChannelTypeIfttt      ChannelType = "ifttt"
	ChannelTypePushover   ChannelType = "pushover"
	ChannelTypeSms        ChannelType = "sms"
)

var channelTypes = []ChannelType{
	ChannelTypeEmail,
	ChannelTypeSlack,
	ChannelTypeWebhook,
	ChannelTypePushbullet,
	ChannelTypeTelegram,
	ChannelTypeAlexa,
	ChannelTypeIfttt,
	ChannelTypePushover,
	ChannelTypeSms,
}

// ChannelTypes returns every supported channel type in stable order.
func ChannelTypes() []ChannelType {
	return append([]ChannelType(nil), channelTypes...)
}

// IsValid reports whether t is a supported channel type.
func (t ChannelType) IsValid() bool {
	for _, channelType := range channelTypes {
		if t == channelType {
			return true
		}
	}
	return false
}

var (
	ErrChannelProjectMissing     = errors.New("channel project is missing")
	ErrChannelTypeUnknown        = errors.New("unknown channel type")
	ErrChannelConfigMissing      = errors.New("channel config is missing")
	ErrChannelConfigAmbiguous    = errors.New("channel config sets more than one provider")
	ErrChannelConfigTypeMismatch = errors.New("channel config does not match channel type")
)

// Channel is a single notification destination (one Slack hook, one email
// address, ...) owned by a project. It replaces the per-user `users.channels`
// document: an organisation can hold any number of channels per type, and
// alerts reference them by Id.
//
// Config holds recipient identifiers (email addresses, phone numbers) and
// provider credentials. It is persisted but never serialised to JSON, so a
// Channel returned by the API only describes the destination. Writers pass
// Config separately (see api.CreateChannelRequest); only server-side delivery
// reads it from the database.
type Channel struct {
	Id primitive.ObjectID `json:"id" bson:"_id,omitempty"`

	Name        string      `json:"name" bson:"name"`
	Description string      `json:"description,omitempty" bson:"description,omitempty"`
	Type        ChannelType `json:"type" bson:"type"`

	// RBAC information
	OrganisationId string `json:"organisationId" bson:"organisationId"`

	// ProjectId is required: a channel belongs to exactly one project, and an
	// alert may only deliver through channels in its own project. There are no
	// organisation-wide channels; channels migrated from users.channels land in
	// the organisation's default project (see DefaultProjectId), and channels
	// migrated from alert overrides land in that alert's project.
	ProjectId primitive.ObjectID `json:"projectId" bson:"projectId"`

	// Enabled is the owner's switch; Valid records whether the last test
	// delivery succeeded. Delivery requires both.
	Enabled bool `json:"enabled" bson:"enabled"`
	Valid   bool `json:"valid" bson:"valid"`

	// Config is write-only: see the type comment.
	Config ChannelConfig `json:"-" bson:"config"`

	// Source records where a migrated channel came from; empty for channels
	// created through the channel API. LegacyKey identifies the legacy record
	// within the organisation so the migration backfill and the legacy
	// write-through upsert the same document: the ChannelType for
	// ChannelSourceLegacy, "<alertId>/<ChannelType>" for
	// ChannelSourceLegacyAlert. Unique per {organisationId, source, legacyKey}.
	Source    ChannelSource `json:"source,omitempty" bson:"source,omitempty"`
	LegacyKey string        `json:"-" bson:"legacyKey,omitempty"`

	Audit *Audit `json:"audit,omitempty" bson:"audit,omitempty"`
}

// ChannelSource identifies the legacy record a migrated Channel was built from.
type ChannelSource string

const (
	// ChannelSourceLegacy is a provider from the master user's users.channels.
	ChannelSourceLegacy ChannelSource = "legacy"
	// ChannelSourceLegacyAlert is a provider a custom alert overrode with its
	// own credentials, merged over the users.channels config.
	ChannelSourceLegacyAlert ChannelSource = "legacy-alert"
)

// LegacyChannelKey returns the LegacyKey for a channel migrated from
// users.channels.
func LegacyChannelKey(channelType ChannelType) string {
	return string(channelType)
}

// LegacyAlertChannelKey returns the LegacyKey for a channel migrated from a
// custom alert's overrides.
func LegacyAlertChannelKey(alertId primitive.ObjectID, channelType ChannelType) string {
	return alertId.Hex() + "/" + string(channelType)
}

// Validate checks that the channel belongs to a project, has a supported type,
// and that Config sets exactly the provider matching it.
func (c Channel) Validate() error {
	if c.ProjectId.IsZero() {
		return ErrChannelProjectMissing
	}
	if !c.Type.IsValid() {
		return fmt.Errorf("%w %q", ErrChannelTypeUnknown, c.Type)
	}
	configType, err := c.Config.Type()
	if err != nil {
		return err
	}
	if configType != c.Type {
		return fmt.Errorf("%w: type %q, config %q", ErrChannelConfigTypeMismatch, c.Type, configType)
	}
	return nil
}

// ChannelConfig carries the provider settings of a Channel. Exactly one member
// is set, matching Channel.Type.
type ChannelConfig struct {
	Email      *ChannelEmailConfig      `json:"email,omitempty" bson:"email,omitempty"`
	Slack      *ChannelSlackConfig      `json:"slack,omitempty" bson:"slack,omitempty"`
	Webhook    *ChannelWebhookConfig    `json:"webhook,omitempty" bson:"webhook,omitempty"`
	Pushbullet *ChannelPushbulletConfig `json:"pushbullet,omitempty" bson:"pushbullet,omitempty"`
	Telegram   *ChannelTelegramConfig   `json:"telegram,omitempty" bson:"telegram,omitempty"`
	Alexa      *ChannelAlexaConfig      `json:"alexa,omitempty" bson:"alexa,omitempty"`
	Ifttt      *ChannelIftttConfig      `json:"ifttt,omitempty" bson:"ifttt,omitempty"`
	Pushover   *ChannelPushoverConfig   `json:"pushover,omitempty" bson:"pushover,omitempty"`
	Sms        *ChannelSmsConfig        `json:"sms,omitempty" bson:"sms,omitempty"`
}

// Type returns the provider this config is for. It fails when no provider or
// more than one provider is set.
func (c ChannelConfig) Type() (ChannelType, error) {
	set := make([]ChannelType, 0, 1)
	if c.Email != nil {
		set = append(set, ChannelTypeEmail)
	}
	if c.Slack != nil {
		set = append(set, ChannelTypeSlack)
	}
	if c.Webhook != nil {
		set = append(set, ChannelTypeWebhook)
	}
	if c.Pushbullet != nil {
		set = append(set, ChannelTypePushbullet)
	}
	if c.Telegram != nil {
		set = append(set, ChannelTypeTelegram)
	}
	if c.Alexa != nil {
		set = append(set, ChannelTypeAlexa)
	}
	if c.Ifttt != nil {
		set = append(set, ChannelTypeIfttt)
	}
	if c.Pushover != nil {
		set = append(set, ChannelTypePushover)
	}
	if c.Sms != nil {
		set = append(set, ChannelTypeSms)
	}
	switch len(set) {
	case 0:
		return "", ErrChannelConfigMissing
	case 1:
		return set[0], nil
	default:
		return "", fmt.Errorf("%w: %v", ErrChannelConfigAmbiguous, set)
	}
}

type ChannelEmailConfig struct {
	Address string `json:"address" bson:"address"`
}

type ChannelSlackConfig struct {
	// Hook is an incoming-webhook URL. ApiKey and Channel are the bot-token
	// alternative.
	Hook     string `json:"hook,omitempty" bson:"hook,omitempty"`
	ApiKey   string `json:"apiKey,omitempty" bson:"apiKey,omitempty"`
	Channel  string `json:"channel,omitempty" bson:"channel,omitempty"`
	Username string `json:"username,omitempty" bson:"username,omitempty"`
}

type ChannelWebhookConfig struct {
	Url string `json:"url" bson:"url"`
}

type ChannelPushbulletConfig struct {
	ApiKey string `json:"apiKey" bson:"apiKey"`
}

type ChannelTelegramConfig struct {
	Token   string `json:"token" bson:"token"`
	Channel string `json:"channel" bson:"channel"`
}

type ChannelAlexaConfig struct {
	AccessCode string `json:"accessCode" bson:"accessCode"`
}

type ChannelIftttConfig struct {
	Token string `json:"token" bson:"token"`
}

type ChannelPushoverConfig struct {
	ApiKey string `json:"apiKey" bson:"apiKey"`
	SendTo string `json:"sendTo" bson:"sendTo"`
}

type ChannelSmsConfig struct {
	AccountSid string `json:"accountSid" bson:"accountSid"`
	AuthToken  string `json:"authToken" bson:"authToken"`
	From       string `json:"from" bson:"from"`
	To         string `json:"to" bson:"to"`
}

// OrganisationChannels was a planned organisation-wide copy of the legacy
// per-user Channels document. It was never persisted.
//
// Deprecated: use Channel.
type OrganisationChannels struct {
	Id             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	OrganisationId string             `json:"organisationId" bson:"organisationId,omitempty"`
	Channels       Channels           `json:"channels" bson:"channels"`
	Audit          *Audit             `json:"audit,omitempty" bson:"audit,omitempty"`
}

// Channels is the legacy per-user channel settings document stored on
// users.channels. New code uses Channel.
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
