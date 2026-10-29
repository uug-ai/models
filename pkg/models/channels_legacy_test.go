package models

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var legacyAlertOverrideBSONFields = []string{
	"email_email", "slack_hook", "slack_botname", "pushbullet_apikey", "telegram_token",
	"telegram_channel", "alexa_token", "webhook_url", "ifttt_token", "sms_accountsid",
	"sms_authtoken", "sms_telfrom", "sms_telto", "pushover_apikey", "pushover_sendto",
}

// legacyProvider returns the users.channels sub-struct for channelType, found
// by its bson tag so a renamed legacy field cannot slip past the tests.
func legacyProvider(t *testing.T, legacy *Channels, channelType ChannelType) reflect.Value {
	t.Helper()
	v := reflect.ValueOf(legacy).Elem()
	for i := 0; i < v.NumField(); i++ {
		name, _, _ := strings.Cut(v.Type().Field(i).Tag.Get("bson"), ",")
		if name == string(channelType) {
			return v.Field(i)
		}
	}
	t.Fatalf("Channels has no %q provider", channelType)
	return reflect.Value{}
}

func stringFields(v reflect.Value) map[string]string {
	values := map[string]string{}
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).Kind() == reflect.String {
			values[v.Type().Field(i).Name] = v.Field(i).String()
		}
	}
	return values
}

// TestLegacyChannelCarriesEveryLegacyValue fills every string field of each
// legacy provider with a distinct value and requires the new config to hold
// exactly those values, so no legacy field is dropped or mixed up.
func TestLegacyChannelCarriesEveryLegacyValue(t *testing.T) {
	organisationId := primitive.NewObjectID()

	for _, channelType := range ChannelTypes() {
		t.Run(string(channelType), func(t *testing.T) {
			var legacy Channels
			provider := legacyProvider(t, &legacy, channelType)
			var want []string
			for i := 0; i < provider.NumField(); i++ {
				field := provider.Field(i)
				switch field.Kind() {
				case reflect.String:
					value := string(channelType) + "-" + provider.Type().Field(i).Name
					field.SetString(value)
					want = append(want, value)
				case reflect.Bool:
					field.SetBool(true)
				}
			}

			channel, configured := LegacyChannel(organisationId, legacy, channelType)
			if !configured {
				t.Fatal("configured = false for a fully populated provider")
			}
			if err := channel.Validate(); err != nil {
				t.Fatalf("Validate() = %v", err)
			}
			if !channel.Enabled || !channel.Valid {
				t.Fatalf("enabled/valid = %v/%v, want copied true/true", channel.Enabled, channel.Valid)
			}
			if channel.OrganisationId != organisationId.Hex() || channel.ProjectId != DefaultProjectId(organisationId) {
				t.Fatalf("ownership = %s/%s", channel.OrganisationId, channel.ProjectId.Hex())
			}
			if channel.Source != ChannelSourceLegacy || channel.LegacyKey != string(channelType) {
				t.Fatalf("source/legacyKey = %q/%q", channel.Source, channel.LegacyKey)
			}

			configValue := reflect.ValueOf(channel.Config)
			var got []string
			for i := 0; i < configValue.NumField(); i++ {
				if !configValue.Field(i).IsNil() {
					for _, value := range stringFields(configValue.Field(i).Elem()) {
						got = append(got, value)
					}
				}
			}
			slices.Sort(got)
			slices.Sort(want)
			if !slices.Equal(got, want) {
				t.Fatalf("config values = %v, want %v", got, want)
			}
		})
	}
}

func TestLegacyChannelUnconfiguredProvider(t *testing.T) {
	legacy := Channels{Email: Email{Enabled: true}}
	for _, channelType := range ChannelTypes() {
		if _, configured := LegacyChannel(primitive.NewObjectID(), legacy, channelType); configured {
			t.Fatalf("%s: configured = true without any value", channelType)
		}
	}
	if channel, configured := LegacyChannel(primitive.NewObjectID(), legacy, "fax"); configured || channel.Type != "" {
		t.Fatal("unknown type must return an empty channel")
	}
}

// TestLegacyAlertOverrideFieldsMatchMerge sets each CustomAlert override field
// on its own and requires LegacyAlertChannel to count it as an override exactly
// when LegacyAlertOverrideFields lists it.
func TestLegacyAlertOverrideFieldsMatchMerge(t *testing.T) {
	alertType := reflect.TypeOf(CustomAlert{})
	fieldIndex := map[string]int{}
	for i := 0; i < alertType.NumField(); i++ {
		name, _, _ := strings.Cut(alertType.Field(i).Tag.Get("bson"), ",")
		fieldIndex[name] = i
	}

	for _, channelType := range ChannelTypes() {
		listed := LegacyAlertOverrideFields(channelType)
		for _, field := range listed {
			if !slices.Contains(legacyAlertOverrideBSONFields, field) {
				t.Fatalf("%s lists %q, which is not an alert override field", channelType, field)
			}
		}
		for _, field := range legacyAlertOverrideBSONFields {
			index, ok := fieldIndex[field]
			if !ok {
				t.Fatalf("CustomAlert has no %q field", field)
			}
			alert := CustomAlert{Id: primitive.NewObjectID()}
			reflect.ValueOf(&alert).Elem().Field(index).SetString("override")

			_, overridden := LegacyAlertChannel(primitive.NewObjectID(), alert, Channels{}, channelType)
			if want := slices.Contains(listed, field); overridden != want {
				t.Fatalf("%s with only %s set: overridden = %v, want %v", channelType, field, overridden, want)
			}
		}
	}
}

func TestLegacyAlertChannelMergesLikeComputeChannels(t *testing.T) {
	organisationId := primitive.NewObjectID()
	alertProject := primitive.NewObjectID()
	legacy := Channels{
		Slack:    Slack{Enabled: false, Valid: false, Hook: "user-hook", Apikey: "user-key", Channel: "#user", Username: "user-bot"},
		Telegram: Telegram{Enabled: false, Valid: false, Token: "user-token", Channel: "user-chat"},
		Sms:      Sms{Valid: false, Accountsid: "user-sid", AuthToken: "user-auth", From: "+1", To: "+2"},
	}
	alert := CustomAlert{
		Id:              primitive.NewObjectID(),
		Title:           "Front door",
		ProjectId:       &alertProject,
		SlackHook:       "alert-hook",
		SlackBotname:    "alert-bot",
		TelegramChannel: "alert-chat",
		SMSTelto:        "+3",
	}

	slack, ok := LegacyAlertChannel(organisationId, alert, legacy, ChannelTypeSlack)
	if !ok {
		t.Fatal("slack hook override not detected")
	}
	if got := *slack.Config.Slack; got != (ChannelSlackConfig{DeliveryMode: SlackDeliveryModeWebhook, Hook: "alert-hook", ApiKey: "user-key", Channel: "#user", Username: "alert-bot"}) {
		t.Fatalf("slack config = %+v", got)
	}
	if !slack.Enabled || !slack.Valid {
		t.Fatalf("slack enabled/valid = %v/%v, want true/true", slack.Enabled, slack.Valid)
	}
	if slack.ProjectId != alertProject || slack.Name != "Front door – Slack" {
		t.Fatalf("slack project/name = %s/%q", slack.ProjectId.Hex(), slack.Name)
	}
	if slack.Source != ChannelSourceLegacyAlert || slack.LegacyKey != LegacyAlertChannelKey(alert.Id, ChannelTypeSlack) {
		t.Fatalf("slack source/legacyKey = %q/%q", slack.Source, slack.LegacyKey)
	}
	if err := slack.Validate(); err != nil {
		t.Fatalf("Validate() = %v", err)
	}

	// A channel-only Telegram override keeps the user's token and validity.
	telegram, ok := LegacyAlertChannel(organisationId, alert, legacy, ChannelTypeTelegram)
	if !ok || *telegram.Config.Telegram != (ChannelTelegramConfig{Token: "user-token", Channel: "alert-chat"}) || telegram.Valid {
		t.Fatalf("telegram = %+v valid=%v ok=%v", telegram.Config.Telegram, telegram.Valid, ok)
	}

	// A recipient-only SMS override keeps the user's credentials and validity.
	sms, ok := LegacyAlertChannel(organisationId, alert, legacy, ChannelTypeSms)
	if !ok || *sms.Config.Sms != (ChannelSmsConfig{AccountSid: "user-sid", AuthToken: "user-auth", From: "+1", To: "+3"}) || sms.Valid {
		t.Fatalf("sms = %+v valid=%v ok=%v", sms.Config.Sms, sms.Valid, ok)
	}

	if _, ok := LegacyAlertChannel(organisationId, alert, legacy, ChannelTypeEmail); ok {
		t.Fatal("email has no override and must map to the legacy channel")
	}

	// An alert without a project lands in the organisation's default project.
	alert.ProjectId = nil
	slack, _ = LegacyAlertChannel(organisationId, alert, legacy, ChannelTypeSlack)
	if slack.ProjectId != DefaultProjectId(organisationId) {
		t.Fatalf("project = %s, want organisation default", slack.ProjectId.Hex())
	}
}

func TestChannelTypeDisplayName(t *testing.T) {
	for _, channelType := range ChannelTypes() {
		if _, ok := channelDisplayNames[channelType]; !ok {
			t.Fatalf("%s has no display name", channelType)
		}
	}
	if got := ChannelType("fax").DisplayName(); got != "fax" {
		t.Fatalf("unknown DisplayName = %q", got)
	}
}
