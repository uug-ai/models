package models

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestUsagePeriods(t *testing.T) {
	// 23:30 UTC on 30 November is 00:30 on 1 December in Brussels (UTC+1).
	at := time.Date(2026, 11, 30, 23, 30, 0, 0, time.UTC)
	tests := []struct {
		timezone, day, month string
	}{
		{"", "2026-12-01", "2026-12"},
		{"Europe/Brussels", "2026-12-01", "2026-12"},
		{"UTC", "2026-11-30", "2026-11"},
		{"America/New_York", "2026-11-30", "2026-11"},
	}
	for _, test := range tests {
		day, err := UsageDayPeriod(at, test.timezone)
		if err != nil || day != test.day {
			t.Errorf("UsageDayPeriod(%q) = %q, %v, want %q", test.timezone, day, err, test.day)
		}
		month, err := UsageMonthPeriod(at, test.timezone)
		if err != nil || month != test.month {
			t.Errorf("UsageMonthPeriod(%q) = %q, %v, want %q", test.timezone, month, err, test.month)
		}
	}
	if _, err := UsageDayPeriod(at, "Mars/Olympus"); err == nil {
		t.Error("unknown timezone accepted")
	}
	if _, err := UsageMonthPeriod(at, "Mars/Olympus"); err == nil {
		t.Error("unknown timezone accepted")
	}
}

func TestNextUsagePeriod(t *testing.T) {
	brussels, err := OrganisationLocation("")
	if err != nil {
		t.Fatal(err)
	}
	// The day DST ends in Brussels has 25 hours.
	at := time.Date(2026, 10, 25, 1, 0, 0, 0, brussels)
	next, err := NextUsageDay(at, "Europe/Brussels")
	if err != nil || !next.Equal(time.Date(2026, 10, 26, 0, 0, 0, 0, brussels)) || next.Sub(at) != 24*time.Hour {
		t.Errorf("NextUsageDay = %v, %v", next, err)
	}
	month, err := NextUsageMonth(time.Date(2026, 12, 31, 23, 30, 0, 0, time.UTC), "")
	if err != nil || !month.Equal(time.Date(2027, 2, 1, 0, 0, 0, 0, brussels)) {
		t.Errorf("NextUsageMonth = %v, %v", month, err)
	}
	if _, err := NextUsageDay(at, "Mars/Olympus"); err == nil {
		t.Error("unknown timezone accepted")
	}
	if _, err := NextUsageMonth(at, "Mars/Olympus"); err == nil {
		t.Error("unknown timezone accepted")
	}
}

func TestUsageCounterBSON(t *testing.T) {
	counter := UsageCounter{OrganisationId: primitive.NewObjectID(), Kind: UsageKindUploadBytes, Period: "2026-11", Value: 42}
	data, err := bson.Marshal(counter)
	if err != nil {
		t.Fatal(err)
	}
	raw := bson.Raw(data)
	if _, err := raw.LookupErr("projectId"); err == nil {
		t.Error("organisation total stores a projectId")
	}
	if raw.Lookup("kind").StringValue() != "uploadBytes" || raw.Lookup("period").StringValue() != "2026-11" {
		t.Errorf("stored %v", raw)
	}
}

func TestDeviceActivation(t *testing.T) {
	if !(Device{}).IsActivated() {
		t.Error("missing activation is active")
	}
	if !(Device{Activation: &DeviceActivation{}}).IsActivated() {
		t.Error("empty state is active")
	}
	inactive := Device{Activation: &DeviceActivation{State: DeviceActivationStateInactive, Reason: DeviceActivationReasonNoSubscription, At: 1}}
	if inactive.IsActivated() {
		t.Error("inactive device is active")
	}
	data, _ := json.Marshal(inactive.Activation)
	if string(data) != `{"state":"inactive","reason":"no_subscription","at":1}` {
		t.Errorf("json = %s", data)
	}
	if (Device{Activation: &DeviceActivation{State: "paused"}}).IsActivated() {
		t.Error("unknown state is active")
	}
}

func TestMediaDeleteAfter(t *testing.T) {
	start := time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC).Unix()
	if got := MediaDeleteAfter(start, 7); got != time.Date(2026, 11, 8, 10, 0, 0, 0, time.UTC).Unix() {
		t.Errorf("MediaDeleteAfter = %d", got)
	}
	data, _ := bson.Marshal(Media{RetentionDays: 7, DeleteAfter: 5})
	raw := bson.Raw(data)
	if raw.Lookup("retentionDays").Int32() != 7 || raw.Lookup("deleteAfter").Int64() != 5 {
		t.Errorf("stored %v", raw)
	}
}

func TestOrganisationSubscriptionMode(t *testing.T) {
	tests := []struct {
		mode  SubscriptionMode
		want  SubscriptionMode
		valid bool
	}{
		{"", SubscriptionModeLegacy, true},
		{SubscriptionModeLegacy, SubscriptionModeLegacy, true},
		{SubscriptionModePlans, SubscriptionModePlans, true},
		{"stripe", SubscriptionModeLegacy, false},
	}
	for _, test := range tests {
		organisation := Organisation{SubscriptionMode: test.mode}
		if got := organisation.EffectiveSubscriptionMode(); got != test.want {
			t.Errorf("%q: effective %q, want %q", test.mode, got, test.want)
		}
		if err := organisation.ValidateSubscriptionMode(); (err == nil) != test.valid {
			t.Errorf("%q: validate %v", test.mode, err)
		}
	}
	data, _ := bson.Marshal(Organisation{})
	raw := bson.Raw(data)
	for _, key := range []string{"subscriptionMode", "billing"} {
		if _, err := raw.LookupErr(key); err == nil {
			t.Errorf("unset %s is stored", key)
		}
	}
}

func TestProjectKeysAndStorageNotInJSON(t *testing.T) {
	project := Project{
		HubPublicKey: "hub-public", HubPrivateKey: "hub-private",
		EncryptionPublicKey: "rsa-public", EncryptionPrivateKey: "rsa-private",
		Storage: &Storage{Uri: "https://vault", Secret: "secret"}, ArchiveStorage: &Storage{Uri: "https://archive"},
	}
	data, _ := json.Marshal(project)
	for _, secret := range []string{"hub-p", "rsa-p", "vault", "archive", "secret"} {
		if strings.Contains(string(data), secret) {
			t.Errorf("json contains %q: %s", secret, data)
		}
	}
	stored, _ := bson.Marshal(project)
	raw := bson.Raw(stored)
	if raw.Lookup("hubPublicKey").StringValue() != "hub-public" || raw.Lookup("encryptionPrivateKey").StringValue() != "rsa-private" ||
		raw.Lookup("storage", "uri").StringValue() != "https://vault" || raw.Lookup("archiveStorage", "uri").StringValue() != "https://archive" {
		t.Errorf("stored %v", raw)
	}
}
