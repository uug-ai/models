package models

import (
	"encoding/json"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestDevicesGroupNotificationsRoundTrip(t *testing.T) {
	settings := NotificationSettings{Devices: Devices{Enabled: true, GroupNotifications: true}}

	data, err := bson.Marshal(settings)
	if err != nil {
		t.Fatal(err)
	}
	var document bson.M
	if err := bson.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	if document["devices"].(bson.M)["groupNotifications"] != true {
		t.Fatalf("bson devices = %v", document["devices"])
	}

	// Legacy documents without the field decode to false (ungrouped).
	legacy, _ := bson.Marshal(bson.M{"devices": bson.M{"enabled": true}})
	var decoded NotificationSettings
	if err := bson.Unmarshal(legacy, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Devices.GroupNotifications {
		t.Fatal("missing groupNotifications must decode as false")
	}

	encoded, err := json.Marshal(settings)
	if err != nil {
		t.Fatal(err)
	}
	var roundTrip NotificationSettings
	if err := json.Unmarshal(encoded, &roundTrip); err != nil || !roundTrip.Devices.GroupNotifications {
		t.Fatalf("json round trip = %+v, %v", roundTrip.Devices, err)
	}
}
