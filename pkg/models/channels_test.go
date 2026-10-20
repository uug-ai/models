package models

import (
	"encoding/json"
	"errors"
	"reflect"
	"slices"
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestChannelJSONNeverCarriesConfig(t *testing.T) {
	projectId := primitive.NewObjectID()
	channel := Channel{
		Name:           "Security desk",
		Type:           ChannelTypeEmail,
		OrganisationId: primitive.NewObjectID().Hex(),
		ProjectId:      projectId,
		Enabled:        true,
		Config:         ChannelConfig{Email: &ChannelEmailConfig{Address: "desk@example.com"}},
	}

	encoded, err := json.Marshal(channel)
	if err != nil {
		t.Fatalf("marshal channel: %v", err)
	}
	if strings.Contains(string(encoded), "desk@example.com") || strings.Contains(string(encoded), `"config"`) {
		t.Fatalf("channel JSON exposes config: %s", encoded)
	}

	var decoded Channel
	if err := json.Unmarshal([]byte(`{"type":"email","config":{"email":{"address":"x@example.com"}}}`), &decoded); err != nil {
		t.Fatalf("unmarshal channel: %v", err)
	}
	if decoded.Config.Email != nil {
		t.Fatal("channel JSON must not bind config; it is passed separately on writes")
	}
}

func TestChannelBSONPersistsConfigAndOwnership(t *testing.T) {
	organisationId := primitive.NewObjectID().Hex()
	projectId := primitive.NewObjectID()
	channel := Channel{
		Type:           ChannelTypeSms,
		OrganisationId: organisationId,
		ProjectId:      projectId,
		Config:         ChannelConfig{Sms: &ChannelSmsConfig{AccountSid: "sid", AuthToken: "token", From: "+1", To: "+2"}},
	}

	encoded, err := bson.Marshal(channel)
	if err != nil {
		t.Fatalf("marshal channel: %v", err)
	}
	var document bson.M
	if err := bson.Unmarshal(encoded, &document); err != nil {
		t.Fatalf("unmarshal channel: %v", err)
	}
	if document["organisationId"] != organisationId {
		t.Fatalf("organisationId = %v, want %s", document["organisationId"], organisationId)
	}
	if document["projectId"] != projectId {
		t.Fatalf("projectId = %v, want %s", document["projectId"], projectId.Hex())
	}

	var decoded Channel
	if err := bson.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("decode channel: %v", err)
	}
	if decoded.Config.Sms == nil || decoded.Config.Sms.AuthToken != "token" {
		t.Fatalf("config did not round-trip through BSON: %+v", decoded.Config)
	}
}

func TestChannelValidate(t *testing.T) {
	email := ChannelConfig{Email: &ChannelEmailConfig{Address: "a@example.com"}}
	projectId := primitive.NewObjectID()
	tests := []struct {
		name    string
		channel Channel
		wantErr error
	}{
		{name: "valid", channel: Channel{ProjectId: projectId, Type: ChannelTypeEmail, Config: email}},
		{name: "missing project", channel: Channel{Type: ChannelTypeEmail, Config: email}, wantErr: ErrChannelProjectMissing},
		{name: "unknown type", channel: Channel{ProjectId: projectId, Type: "fax", Config: email}, wantErr: ErrChannelTypeUnknown},
		{name: "missing config", channel: Channel{ProjectId: projectId, Type: ChannelTypeEmail}, wantErr: ErrChannelConfigMissing},
		{name: "mismatched config", channel: Channel{ProjectId: projectId, Type: ChannelTypeSlack, Config: email}, wantErr: ErrChannelConfigTypeMismatch},
		{
			name: "ambiguous config",
			channel: Channel{ProjectId: projectId, Type: ChannelTypeEmail, Config: ChannelConfig{
				Email:   &ChannelEmailConfig{Address: "a@example.com"},
				Webhook: &ChannelWebhookConfig{Url: "https://example.com"},
			}},
			wantErr: ErrChannelConfigAmbiguous,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.channel.Validate()
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("Validate() error = %v, want %v", err, test.wantErr)
			}
		})
	}
}

// TestChannelConfigMembersMatchChannelTypes walks ChannelConfig by reflection so
// a provider added to the struct but not to ChannelConfig.Type or the
// ChannelType catalog fails here instead of at runtime.
func TestChannelConfigMembersMatchChannelTypes(t *testing.T) {
	configType := reflect.TypeOf(ChannelConfig{})
	seen := make([]ChannelType, 0, configType.NumField())

	for i := 0; i < configType.NumField(); i++ {
		field := configType.Field(i)
		if field.Type.Kind() != reflect.Pointer || field.Type.Elem().Kind() != reflect.Struct {
			t.Fatalf("ChannelConfig.%s must be a pointer to a provider config struct", field.Name)
		}
		jsonName, _, _ := strings.Cut(field.Tag.Get("json"), ",")
		bsonName, _, _ := strings.Cut(field.Tag.Get("bson"), ",")
		if jsonName != bsonName {
			t.Fatalf("ChannelConfig.%s json tag %q != bson tag %q", field.Name, jsonName, bsonName)
		}
		want := ChannelType(jsonName)
		if !want.IsValid() {
			t.Fatalf("ChannelConfig.%s tag %q is not a ChannelType", field.Name, jsonName)
		}

		var config ChannelConfig
		reflect.ValueOf(&config).Elem().Field(i).Set(reflect.New(field.Type.Elem()))
		got, err := config.Type()
		if err != nil || got != want {
			t.Fatalf("ChannelConfig{%s}.Type() = %q, %v; want %q", field.Name, got, err, want)
		}
		seen = append(seen, want)
	}

	if !slices.Equal(seen, ChannelTypes()) {
		t.Fatalf("ChannelConfig members %v, ChannelTypes() %v: keep them in the same order", seen, ChannelTypes())
	}
}

func TestChannelTypesReturnsCopy(t *testing.T) {
	types := ChannelTypes()
	types[0] = "changed"
	if !slices.Equal(ChannelTypes()[:1], []ChannelType{ChannelTypeEmail}) {
		t.Fatal("ChannelTypes returned mutable catalog storage")
	}
	if ChannelType("changed").IsValid() {
		t.Fatal("IsValid accepted an unknown channel type")
	}
}

func TestChannelLegacyKeyIsPersistedButNotSerialised(t *testing.T) {
	alertId := primitive.NewObjectID()
	if got, want := LegacyChannelKey(ChannelTypeSlack), "slack"; got != want {
		t.Fatalf("LegacyChannelKey = %q, want %q", got, want)
	}
	if got, want := LegacyAlertChannelKey(alertId, ChannelTypeSms), alertId.Hex()+"/sms"; got != want {
		t.Fatalf("LegacyAlertChannelKey = %q, want %q", got, want)
	}

	channel := Channel{
		ProjectId: primitive.NewObjectID(),
		Type:      ChannelTypeSms,
		Source:    ChannelSourceLegacyAlert,
		LegacyKey: LegacyAlertChannelKey(alertId, ChannelTypeSms),
	}

	encoded, err := json.Marshal(channel)
	if err != nil {
		t.Fatalf("marshal channel: %v", err)
	}
	if strings.Contains(string(encoded), "legacyKey") || !strings.Contains(string(encoded), `"source":"legacy-alert"`) {
		t.Fatalf("channel JSON = %s, want source without legacyKey", encoded)
	}

	data, err := bson.Marshal(channel)
	if err != nil {
		t.Fatalf("marshal channel: %v", err)
	}
	var document bson.M
	if err := bson.Unmarshal(data, &document); err != nil {
		t.Fatalf("unmarshal channel: %v", err)
	}
	if document["source"] != string(ChannelSourceLegacyAlert) || document["legacyKey"] != channel.LegacyKey {
		t.Fatalf("channel BSON source/legacyKey = %v/%v", document["source"], document["legacyKey"])
	}
}
