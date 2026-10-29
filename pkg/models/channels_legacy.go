package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// The helpers in this file are the single mapping from the legacy per-user
// channel settings (users.channels and the CustomAlert overrides) to Channel.
// The Hub API legacy write-through, the migration backfill and its verification
// all build channels through them, so the three cannot disagree on a field.

var channelDisplayNames = map[ChannelType]string{
	ChannelTypeEmail:      "Email",
	ChannelTypeSlack:      "Slack",
	ChannelTypeWebhook:    "Webhook",
	ChannelTypePushbullet: "Pushbullet",
	ChannelTypeTelegram:   "Telegram",
	ChannelTypeAlexa:      "Alexa",
	ChannelTypeIfttt:      "IFTTT",
	ChannelTypePushover:   "Pushover",
	ChannelTypeSms:        "SMS",
}

// DisplayName returns the provider's human-readable name.
func (t ChannelType) DisplayName() string {
	if name, ok := channelDisplayNames[t]; ok {
		return name
	}
	return string(t)
}

// LegacyChannel converts one provider of the legacy users.channels document of
// an organisation's master user into a ChannelSourceLegacy channel in the
// organisation's default project. configured is false when the provider holds
// no configuration value at all; Enabled and Valid are copied as stored.
func LegacyChannel(organisationId primitive.ObjectID, legacy Channels, channelType ChannelType) (channel Channel, configured bool) {
	channel = Channel{
		Name:           channelType.DisplayName(),
		Type:           channelType,
		OrganisationId: organisationId.Hex(),
		ProjectId:      DefaultProjectId(organisationId),
		Source:         ChannelSourceLegacy,
		LegacyKey:      LegacyChannelKey(channelType),
	}

	switch channelType {
	case ChannelTypeEmail:
		c := legacy.Email
		channel.Enabled, channel.Valid = c.Enabled, c.Valid
		channel.Config.Email = &ChannelEmailConfig{Address: c.Address}
		return channel, anySet(c.Address)
	case ChannelTypeSlack:
		c := legacy.Slack
		channel.Enabled, channel.Valid = c.Enabled, c.Valid
		channel.Config.Slack = &ChannelSlackConfig{DeliveryMode: c.DeliveryMode, Hook: c.Hook, ApiKey: c.Apikey, Channel: c.Channel, Username: c.Username}
		return channel, anySet(c.Hook, c.Apikey, c.Channel, c.Username)
	case ChannelTypeWebhook:
		c := legacy.Webhook
		channel.Enabled, channel.Valid = c.Enabled, c.Valid
		channel.Config.Webhook = &ChannelWebhookConfig{Url: c.Url}
		return channel, anySet(c.Url)
	case ChannelTypePushbullet:
		c := legacy.Pushbullet
		channel.Enabled, channel.Valid = c.Enabled, c.Valid
		channel.Config.Pushbullet = &ChannelPushbulletConfig{ApiKey: c.Apikey}
		return channel, anySet(c.Apikey)
	case ChannelTypeTelegram:
		c := legacy.Telegram
		channel.Enabled, channel.Valid = c.Enabled, c.Valid
		channel.Config.Telegram = &ChannelTelegramConfig{Token: c.Token, Channel: c.Channel}
		return channel, anySet(c.Token, c.Channel)
	case ChannelTypeAlexa:
		c := legacy.Alexa
		channel.Enabled, channel.Valid = c.Enabled, c.Valid
		channel.Config.Alexa = &ChannelAlexaConfig{AccessCode: c.Accesscode}
		return channel, anySet(c.Accesscode)
	case ChannelTypeIfttt:
		c := legacy.Ifttt
		channel.Enabled, channel.Valid = c.Enabled, c.Valid
		channel.Config.Ifttt = &ChannelIftttConfig{Token: c.Token}
		return channel, anySet(c.Token)
	case ChannelTypePushover:
		c := legacy.Pushover
		channel.Enabled, channel.Valid = c.Enabled, c.Valid
		channel.Config.Pushover = &ChannelPushoverConfig{ApiKey: c.Apikey, SendTo: c.Sendto}
		return channel, anySet(c.Apikey, c.Sendto)
	case ChannelTypeSms:
		c := legacy.Sms
		channel.Enabled, channel.Valid = c.Enabled, c.Valid
		channel.Config.Sms = &ChannelSmsConfig{AccountSid: c.Accountsid, AuthToken: c.AuthToken, From: c.From, To: c.To}
		return channel, anySet(c.Accountsid, c.AuthToken, c.From, c.To)
	}
	return Channel{}, false
}

// LegacyAlertChannel converts a custom alert's overrides for one provider into
// a ChannelSourceLegacyAlert channel in the alert's project (the organisation
// default when the alert has none). overridden is false when the alert does
// not override that provider; the alert then maps to the LegacyChannel.
//
// The merge mirrors hub-pipeline-notification's ComputeChannels: the
// users.channels provider is the base, each non-empty override replaces its
// field, the channel is enabled, and the provider's primary override marks it
// valid. An override ComputeChannels ignores (a Slack bot name without a hook)
// does not count as overriding.
func LegacyAlertChannel(organisationId primitive.ObjectID, alert CustomAlert, legacy Channels, channelType ChannelType) (channel Channel, overridden bool) {
	channel, _ = LegacyChannel(organisationId, legacy, channelType)
	if channel.Type == "" {
		return Channel{}, false
	}

	title := alert.Title
	if title == "" {
		title = "Alert"
	}
	channel.Name = title + " – " + channelType.DisplayName()
	if alert.ProjectId != nil && !alert.ProjectId.IsZero() {
		channel.ProjectId = *alert.ProjectId
	}
	channel.Source = ChannelSourceLegacyAlert
	channel.LegacyKey = LegacyAlertChannelKey(alert.Id, channelType)
	channel.Enabled = true

	config := &channel.Config
	switch channelType {
	case ChannelTypeEmail:
		if alert.EmailEmail != "" {
			config.Email.Address = alert.EmailEmail
			channel.Valid, overridden = true, true
		}
	case ChannelTypeSlack:
		if alert.SlackHook != "" {
			if alert.SlackHook != legacy.Slack.Hook {
				config.Slack.DeliveryMode = SlackDeliveryModeWebhook
			}
			config.Slack.Hook = alert.SlackHook
			config.Slack.Username = alert.SlackBotname
			channel.Valid, overridden = true, true
		}
	case ChannelTypeWebhook:
		if alert.WebhookUrl != "" {
			config.Webhook.Url = alert.WebhookUrl
			channel.Valid, overridden = true, true
		}
	case ChannelTypePushbullet:
		if alert.PushbulletApikey != "" {
			config.Pushbullet.ApiKey = alert.PushbulletApikey
			channel.Valid, overridden = true, true
		}
	case ChannelTypeTelegram:
		if alert.TelegramToken != "" {
			config.Telegram.Token = alert.TelegramToken
			channel.Valid, overridden = true, true
		}
		if alert.TelegramChannel != "" {
			config.Telegram.Channel = alert.TelegramChannel
			overridden = true
		}
	case ChannelTypeAlexa:
		if alert.AlexaToken != "" {
			config.Alexa.AccessCode = alert.AlexaToken
			channel.Valid, overridden = true, true
		}
	case ChannelTypeIfttt:
		if alert.IftttToken != "" {
			config.Ifttt.Token = alert.IftttToken
			channel.Valid, overridden = true, true
		}
	case ChannelTypePushover:
		if alert.PushoverApikey != "" {
			config.Pushover.ApiKey = alert.PushoverApikey
			channel.Valid, overridden = true, true
		}
		if alert.PushoverSendto != "" {
			config.Pushover.SendTo = alert.PushoverSendto
			overridden = true
		}
	case ChannelTypeSms:
		if alert.SMSAccountsid != "" {
			config.Sms.AccountSid = alert.SMSAccountsid
			channel.Valid, overridden = true, true
		}
		if alert.SMSAuthtoken != "" {
			config.Sms.AuthToken = alert.SMSAuthtoken
			overridden = true
		}
		if alert.SMSTelfrom != "" {
			config.Sms.From = alert.SMSTelfrom
			overridden = true
		}
		if alert.SMSTelto != "" {
			config.Sms.To = alert.SMSTelto
			overridden = true
		}
	}
	if !overridden {
		return Channel{}, false
	}
	return channel, true
}

// LegacyAlertOverrideFields returns the CustomAlert BSON fields whose non-empty
// value makes LegacyAlertChannel treat channelType as overridden.
func LegacyAlertOverrideFields(channelType ChannelType) []string {
	switch channelType {
	case ChannelTypeEmail:
		return []string{"email_email"}
	case ChannelTypeSlack:
		return []string{"slack_hook"}
	case ChannelTypeWebhook:
		return []string{"webhook_url"}
	case ChannelTypePushbullet:
		return []string{"pushbullet_apikey"}
	case ChannelTypeTelegram:
		return []string{"telegram_token", "telegram_channel"}
	case ChannelTypeAlexa:
		return []string{"alexa_token"}
	case ChannelTypeIfttt:
		return []string{"ifttt_token"}
	case ChannelTypePushover:
		return []string{"pushover_apikey", "pushover_sendto"}
	case ChannelTypeSms:
		return []string{"sms_accountsid", "sms_authtoken", "sms_telfrom", "sms_telto"}
	}
	return nil
}

func anySet(values ...string) bool {
	for _, value := range values {
		if value != "" {
			return true
		}
	}
	return false
}
