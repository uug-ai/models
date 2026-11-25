package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestCanRequestOnDemandForward(t *testing.T) {
	id := primitive.NewObjectID()
	zero := primitive.NilObjectID
	cases := []struct {
		name  string
		media Media
		want  bool
	}{
		{"metadata-only with source", Media{StorageSolution: "metadata", SourceVaultId: "v1", SourceMediaId: &id}, true},
		{"case and spaces", Media{StorageSolution: " Metadata ", SourceVaultId: "v1", SourceMediaId: &id}, true},
		{"stored", Media{StorageSolution: "kstorage", SourceVaultId: "v1", SourceMediaId: &id}, false},
		{"no vault", Media{StorageSolution: "metadata", SourceVaultId: " ", SourceMediaId: &id}, false},
		{"no source id", Media{StorageSolution: "metadata", SourceVaultId: "v1"}, false},
		{"zero source id", Media{StorageSolution: "metadata", SourceVaultId: "v1", SourceMediaId: &zero}, false},
	}
	for _, c := range cases {
		if got := c.media.CanRequestOnDemandForward(); got != c.want {
			t.Errorf("%s: %v, want %v", c.name, got, c.want)
		}
	}
	if OnDemandForwardTopic("AKIA1") != "kerberos/vault/AKIA1" {
		t.Error("topic")
	}
}

func TestNewOnDemandForwardMessagesMatchesWhatTheVaultReads(t *testing.T) {
	now := time.Date(2026, 11, 25, 12, 0, 0, 0, time.UTC)
	messages, err := NewOnDemandForwardMessages("vault-1", []string{"a", "b", "a", " "}, now)
	if err != nil || len(messages) != 1 {
		t.Fatalf("messages = %+v, err %v", messages, err)
	}
	raw, err := json.Marshal(messages[0])
	if err != nil {
		t.Fatal(err)
	}
	// The fields api/forward in Vault decodes and checks.
	var vault struct {
		Mid       string `json:"mid"`
		VaultId   string `json:"vault_id"`
		Timestamp int64  `json:"timestamp"`
		Payload   struct {
			Version string `json:"version"`
			Action  string `json:"action"`
			VaultId string `json:"vault_id"`
			Value   struct {
				RequestId  string `json:"request_id"`
				Recordings []struct {
					SourceMediaId string `json:"source_media_id"`
				} `json:"recordings"`
			} `json:"value"`
		} `json:"payload"`
	}
	if err := json.Unmarshal(raw, &vault); err != nil {
		t.Fatal(err)
	}
	if vault.Mid == "" || vault.Timestamp != now.Unix() || vault.VaultId != "vault-1" ||
		vault.Payload.Version != "1.0" || vault.Payload.Action != "forward-recordings" || vault.Payload.VaultId != "vault-1" ||
		vault.Payload.Value.RequestId == "" || len(vault.Payload.Value.Recordings) != 2 ||
		vault.Payload.Value.Recordings[0].SourceMediaId != "a" || vault.Payload.Value.Recordings[1].SourceMediaId != "b" {
		t.Fatalf("message = %s", raw)
	}
	if !strings.Contains(string(raw), `"encrypted":false`) || len(messages[0].PayloadHash) != 64 {
		t.Fatalf("envelope = %s", raw)
	}
}

func TestNewOnDemandForwardMessagesSplitsLargeRequests(t *testing.T) {
	ids := make([]string, OnDemandForwardMaxRecordings*2+1)
	for i := range ids {
		ids[i] = fmt.Sprintf("id-%d", i)
	}
	messages, err := NewOnDemandForwardMessages("vault-1", ids, time.Now())
	if err != nil || len(messages) != 3 {
		t.Fatalf("messages = %d, err %v", len(messages), err)
	}
	if len(messages[0].Payload.Value.Recordings) != OnDemandForwardMaxRecordings || len(messages[2].Payload.Value.Recordings) != 1 {
		t.Fatal("messages not cut at the Vault's limit")
	}
	if messages[0].Mid == messages[1].Mid || messages[0].Payload.Value.RequestId == messages[1].Payload.Value.RequestId {
		t.Fatal("messages share an id")
	}
	if none, err := NewOnDemandForwardMessages("vault-1", nil, time.Now()); err != nil || len(none) != 0 {
		t.Fatalf("no ids gave %d messages, err %v", len(none), err)
	}
	if _, err := NewOnDemandForwardMessages(" ", ids, time.Now()); err == nil {
		t.Fatal("an empty vault id was accepted")
	}
}
