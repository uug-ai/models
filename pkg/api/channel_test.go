package api

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/uug-ai/models/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestCreateSlackChannelRequestPersistsConfig(t *testing.T) {
	tests := []struct {
		name         string
		configJSON   string
		wantConfig   *models.ChannelSlackConfig
		wantDocument bson.M
	}{
		{
			name:         "bot",
			configJSON:   `{"apiKey":"test-bot-token","channel":"C123456789"}`,
			wantConfig:   &models.ChannelSlackConfig{ApiKey: "test-bot-token", Channel: "C123456789"},
			wantDocument: bson.M{"apiKey": "test-bot-token", "channel": "C123456789"},
		},
		{
			name:         "webhook",
			configJSON:   `{"hook":"https://hooks.slack.com/services/test","username":"Hub"}`,
			wantConfig:   &models.ChannelSlackConfig{Hook: "https://hooks.slack.com/services/test", Username: "Hub"},
			wantDocument: bson.M{"hook": "https://hooks.slack.com/services/test", "username": "Hub"},
		},
		{
			name:         "webhook with saved bot",
			configJSON:   `{"deliveryMode":"webhook","apiKey":"test-bot-token","channel":"C123456789","hook":"https://hooks.slack.com/services/test","username":"Hub"}`,
			wantConfig:   &models.ChannelSlackConfig{DeliveryMode: "webhook", ApiKey: "test-bot-token", Channel: "C123456789", Hook: "https://hooks.slack.com/services/test", Username: "Hub"},
			wantDocument: bson.M{"deliveryMode": "webhook", "apiKey": "test-bot-token", "channel": "C123456789", "hook": "https://hooks.slack.com/services/test", "username": "Hub"},
		},
		{
			name:         "bot with saved webhook",
			configJSON:   `{"deliveryMode":"bot","apiKey":"test-bot-token","channel":"C123456789","hook":"https://hooks.slack.com/services/test","username":"Hub"}`,
			wantConfig:   &models.ChannelSlackConfig{DeliveryMode: "bot", ApiKey: "test-bot-token", Channel: "C123456789", Hook: "https://hooks.slack.com/services/test", Username: "Hub"},
			wantDocument: bson.M{"deliveryMode": "bot", "apiKey": "test-bot-token", "channel": "C123456789", "hook": "https://hooks.slack.com/services/test", "username": "Hub"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			organisationId := primitive.NewObjectID().Hex()
			projectId := primitive.NewObjectID()
			payload := fmt.Sprintf(`{
				"channel": {
					"name": "Slack notifications",
					"type": "slack",
					"organisationId": %q,
					"projectId": %q,
					"enabled": true
				},
				"config": {"slack": %s}
			}`, organisationId, projectId.Hex(), test.configJSON)

			var request CreateChannelRequest
			if err := json.Unmarshal([]byte(payload), &request); err != nil {
				t.Fatalf("decode create request: %v", err)
			}
			if !reflect.DeepEqual(request.Config.Slack, test.wantConfig) {
				t.Fatalf("request Slack config = %#v, want %#v", request.Config.Slack, test.wantConfig)
			}

			channel := request.Channel
			channel.Config = request.Config
			if err := channel.Validate(); err != nil {
				t.Fatalf("validate channel: %v", err)
			}
			encoded, err := bson.Marshal(channel)
			if err != nil {
				t.Fatalf("marshal channel: %v", err)
			}
			var document bson.M
			if err := bson.Unmarshal(encoded, &document); err != nil {
				t.Fatalf("decode channel document: %v", err)
			}
			if document["organisationId"] != organisationId || document["projectId"] != projectId {
				t.Fatalf("channel ownership changed: %#v", document)
			}
			if document["type"] != "slack" || document["enabled"] != true || document["valid"] != false {
				t.Fatalf("channel metadata = %#v", document)
			}
			wantDocument := bson.M{"slack": test.wantDocument}
			if !reflect.DeepEqual(document["config"], wantDocument) {
				t.Fatalf("stored config = %#v, want %#v", document["config"], wantDocument)
			}

			var stored models.Channel
			if err := bson.Unmarshal(encoded, &stored); err != nil {
				t.Fatalf("read stored channel: %v", err)
			}
			if !reflect.DeepEqual(stored.Config, request.Config) {
				t.Fatalf("stored config did not round-trip: %#v", stored.Config)
			}
			response, err := json.Marshal(CreateChannelResponse{Channel: stored})
			if err != nil {
				t.Fatalf("marshal create response: %v", err)
			}
			if strings.Contains(string(response), `"config"`) {
				t.Fatalf("create response exposes channel config: %s", response)
			}
		})
	}
}
