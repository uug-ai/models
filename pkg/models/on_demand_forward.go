package models

// On-demand forwarding.
//
// A recording an edge Vault has not forwarded to the cloud yet exists in
// Hub as metadata only (storageSolution "metadata"). It can be requested
// from that Vault: a "forward-recordings" message on the MQTT topic
// kerberos/vault/<cloudKey>, where cloudKey is the Hub account key the
// Vault forwards for (the owner's amazon_access_key_id). The Vault marks
// the recordings for forwarding and uploads them; the media document then
// gets a cloud storage solution.
//
// These are the rule and the message the Hub frontend uses
// (shared/models/on-demand-forward.ts and MqttService.publish) and the
// Vault reads (api/forward), for Go services.

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	// MetadataOnlyStorageSolution marks a recording whose video is still
	// on its edge Vault.
	MetadataOnlyStorageSolution = "metadata"
	// OnDemandForwardVersion and OnDemandForwardAction identify a
	// forward request.
	OnDemandForwardVersion = "1.0"
	OnDemandForwardAction  = "forward-recordings"
	// OnDemandForwardMaxRecordings is the most recordings one message may
	// hold; the Vault refuses larger requests.
	OnDemandForwardMaxRecordings = 200
)

// IsMetadataOnlyStorage reports whether storageSolution marks a recording
// whose video is not in cloud storage yet.
func IsMetadataOnlyStorage(storageSolution string) bool {
	return strings.EqualFold(strings.TrimSpace(storageSolution), MetadataOnlyStorageSolution)
}

// CanRequestOnDemandForward reports whether a recording can be requested
// from its edge Vault: it is metadata-only and carries the source Vault
// and the recording's id there.
func CanRequestOnDemandForward(storageSolution, sourceVaultId string, sourceMediaId *primitive.ObjectID) bool {
	return IsMetadataOnlyStorage(storageSolution) &&
		strings.TrimSpace(sourceVaultId) != "" &&
		sourceMediaId != nil && !sourceMediaId.IsZero()
}

// CanRequestOnDemandForward reports whether m can be requested from its
// edge Vault (see the function of the same name).
func (m Media) CanRequestOnDemandForward() bool {
	return CanRequestOnDemandForward(m.StorageSolution, m.SourceVaultId, m.SourceMediaId)
}

// OnDemandForwardTopic is the MQTT topic the Vaults forwarding for
// cloudKey listen on.
func OnDemandForwardTopic(cloudKey string) string {
	return "kerberos/vault/" + cloudKey
}

// OnDemandForwardRecording is one recording asked for, by its id on the
// edge Vault (the media's sourceMediaId).
type OnDemandForwardRecording struct {
	SourceMediaId string `json:"source_media_id"`
}

// OnDemandForwardValue is the request id and the recordings asked for.
type OnDemandForwardValue struct {
	RequestId  string                     `json:"request_id"`
	Recordings []OnDemandForwardRecording `json:"recordings"`
}

// OnDemandForwardRequest is the payload of a forward request.
type OnDemandForwardRequest struct {
	Version string               `json:"version"`
	Action  string               `json:"action"`
	VaultId string               `json:"vault_id"`
	Value   OnDemandForwardValue `json:"value"`
}

// OnDemandForwardMessage is the MQTT message around a forward request,
// as the Hub frontend sends it unencrypted.
type OnDemandForwardMessage struct {
	Mid         string                 `json:"mid"`
	Timestamp   int64                  `json:"timestamp"`
	Hidden      bool                   `json:"hidden"`
	Encrypted   bool                   `json:"encrypted"`
	Fingerprint string                 `json:"fingerprint"`
	VaultId     string                 `json:"vault_id"`
	Payload     OnDemandForwardRequest `json:"payload"`
	PayloadHash string                 `json:"payload_hash"`
}

// NewOnDemandForwardMessages asks vaultId for the recordings with the
// given ids on that Vault, in as many messages as needed to keep each at
// OnDemandForwardMaxRecordings or fewer. Repeated and empty ids are left
// out; no ids gives no messages.
func NewOnDemandForwardMessages(vaultId string, sourceMediaIds []string, now time.Time) ([]OnDemandForwardMessage, error) {
	vaultId = strings.TrimSpace(vaultId)
	if vaultId == "" {
		return nil, fmt.Errorf("on-demand forward: vault id is empty")
	}
	seen := make(map[string]bool, len(sourceMediaIds))
	recordings := make([]OnDemandForwardRecording, 0, len(sourceMediaIds))
	for _, id := range sourceMediaIds {
		id = strings.TrimSpace(id)
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		recordings = append(recordings, OnDemandForwardRecording{SourceMediaId: id})
	}
	var messages []OnDemandForwardMessage
	for start := 0; start < len(recordings); start += OnDemandForwardMaxRecordings {
		end := start + OnDemandForwardMaxRecordings
		if end > len(recordings) {
			end = len(recordings)
		}
		requestId, err := newUUID()
		if err != nil {
			return nil, err
		}
		mid, err := newUUID()
		if err != nil {
			return nil, err
		}
		payload := OnDemandForwardRequest{
			Version: OnDemandForwardVersion,
			Action:  OnDemandForwardAction,
			VaultId: vaultId,
			Value:   OnDemandForwardValue{RequestId: requestId, Recordings: recordings[start:end]},
		}
		raw, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("on-demand forward: encode payload: %w", err)
		}
		hash := sha256.Sum256(raw)
		messages = append(messages, OnDemandForwardMessage{
			Mid:         mid,
			Timestamp:   now.Unix(),
			VaultId:     vaultId,
			Payload:     payload,
			PayloadHash: hex.EncodeToString(hash[:]),
		})
	}
	return messages, nil
}

// newUUID returns a random (version 4) UUID.
func newUUID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("on-demand forward: random id: %w", err)
	}
	b[6] = b[6]&0x0f | 0x40
	b[8] = b[8]&0x3f | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
