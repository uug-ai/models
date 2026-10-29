package models

import (
	"encoding/json"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestResolveSlackDeliveryMode(t *testing.T) {
	for _, test := range []struct {
		name string
		mode string
		bot  bool
		want string
	}{
		{"legacy webhook", "", false, SlackDeliveryModeWebhook},
		{"legacy bot", "", true, SlackDeliveryModeBot},
		{"explicit webhook with bot saved", SlackDeliveryModeWebhook, true, SlackDeliveryModeWebhook},
		{"explicit bot", SlackDeliveryModeBot, true, SlackDeliveryModeBot},
		{"explicit bot without credentials", SlackDeliveryModeBot, false, SlackDeliveryModeBot},
		{"unknown mode", "other", true, ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			settings := Slack{DeliveryMode: test.mode, Hook: "saved-webhook"}
			if test.bot {
				settings.Apikey, settings.Channel = "saved-token", "C123"
			}
			before := settings
			mode, err := settings.ResolveDeliveryMode()
			if mode != test.want || (err != nil) != (test.want == "") {
				t.Fatalf("mode = %q, error = %v; want %q", mode, err, test.want)
			}
			if settings != before {
				t.Fatal("resolving a mode changed the saved configuration")
			}
		})
	}
}

func TestSlackDeliveryModesPreserveBothConfigurations(t *testing.T) {
	for _, mode := range []string{SlackDeliveryModeWebhook, SlackDeliveryModeBot} {
		t.Run(mode, func(t *testing.T) {
			want := Slack{
				Enabled: true, Valid: true, DeliveryMode: mode,
				Apikey: "saved-token", Channel: "C123",
				Hook: "https://hooks.slack.com/services/example", Username: "Hub",
			}
			for _, codec := range []struct {
				name      string
				marshal   func(interface{}) ([]byte, error)
				unmarshal func([]byte, interface{}) error
			}{
				{"JSON", json.Marshal, json.Unmarshal},
				{"BSON", bson.Marshal, bson.Unmarshal},
			} {
				data, err := codec.marshal(want)
				if err != nil {
					t.Fatal(err)
				}
				var got Slack
				if err := codec.unmarshal(data, &got); err != nil {
					t.Fatal(err)
				}
				if got != want {
					t.Fatalf("%s round trip lost Slack settings", codec.name)
				}
			}
			channel, configured := LegacyChannel(primitive.NewObjectID(), Channels{Slack: want}, ChannelTypeSlack)
			if !configured || *channel.Config.Slack != (ChannelSlackConfig{
				DeliveryMode: mode, ApiKey: want.Apikey, Channel: want.Channel, Hook: want.Hook, Username: want.Username,
			}) {
				t.Fatal("collection mirror lost the mode or inactive configuration")
			}
		})
	}
}

func TestSlackAlertOverrideSelectsItsOwnWebhook(t *testing.T) {
	legacy := Channels{Slack: Slack{
		DeliveryMode: SlackDeliveryModeBot, Apikey: "saved-token", Channel: "C123", Hook: "default-hook",
	}}
	for _, hook := range []string{"default-hook", "override-hook"} {
		channel, overridden := LegacyAlertChannel(primitive.NewObjectID(), CustomAlert{
			Id: primitive.NewObjectID(), SlackHook: hook,
		}, legacy, ChannelTypeSlack)
		wantMode := SlackDeliveryModeBot
		if hook != legacy.Slack.Hook {
			wantMode = SlackDeliveryModeWebhook
		}
		if !overridden || channel.Config.Slack.DeliveryMode != wantMode || channel.Config.Slack.ApiKey != legacy.Slack.Apikey {
			t.Fatal("alert override selected the wrong delivery mode or discarded saved bot settings")
		}
	}
}
