package models

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestPipelineEventUserAuditShapes(t *testing.T) {
	shapes := []struct {
		name  string
		field string
	}{
		{"missing", ""},
		{"null", `,"audit":null`},
		{"empty array", `,"audit":[]`},
		{"legacy history", `,"audit":[{"action":"login","timestamp":1706000000}]`},
		{"object", `,"audit":{"createdBy":"owner","updatedAt":"2026-09-21T10:00:00Z"}`},
	}
	for _, root := range shapes {
		for _, master := range shapes {
			t.Run(root.name+"/"+master.name, func(t *testing.T) {
				data := []byte(`{"traceId":"trace","monitorStage":{"name":"monitor","user":{"username":"child"` +
					root.field + `,"master":{"username":"parent"` + master.field +
					`,"master":{"username":"grandparent"` + master.field + `}}}}}`)
				var event PipelineEvent
				if err := json.Unmarshal(data, &event); err != nil {
					t.Fatalf("queue decode: %v", err)
				}
				for i, user := 0, &event.MonitorStage.User; user != nil; i, user = i+1, user.Master {
					if user.Audit != (Audit{}) {
						t.Errorf("user depth %d retained audit: %#v", i, user.Audit)
					}
				}
				if event.MonitorStage.User.Username != "child" ||
					event.MonitorStage.User.Master.Username != "parent" ||
					event.MonitorStage.User.Master.Master.Username != "grandparent" {
					t.Fatal("lost non-audit user fields")
				}
				encoded, err := json.Marshal(event)
				if err != nil {
					t.Fatalf("queue forward: %v", err)
				}
				assertPipelineUserAuditOmitted(t, encoded)
				var forwarded PipelineEvent
				if err := json.Unmarshal(encoded, &forwarded); err != nil {
					t.Fatalf("forward decode: %v", err)
				}
				if !reflect.DeepEqual(event, forwarded) {
					t.Fatal("forward changed non-audit fields")
				}
			})
		}
	}
}

func TestPipelineEventJSONPreservesNonAuditFields(t *testing.T) {
	event := pipelineAuditFixture()
	// This shadow restores the original monitor JSON encoder as a baseline,
	// without redefining any of the User or PipelineEvent model fields.
	type eventJSON PipelineEvent
	type monitorJSON MonitorStage
	baseline, err := json.Marshal(struct {
		*eventJSON
		MonitorStage *monitorJSON `json:"monitorStage,omitempty"`
	}{
		eventJSON:    (*eventJSON)(&event),
		MonitorStage: (*monitorJSON)(event.MonitorStage),
	})
	if err != nil {
		t.Fatal(err)
	}
	wantWire := pipelineJSONMap(t, baseline)
	for user := wantWire["monitorStage"].(map[string]any)["user"].(map[string]any); user != nil; {
		delete(user, "audit")
		user, _ = user["master"].(map[string]any)
	}
	for _, input := range []any{event, &event} {
		encoded, err := json.Marshal(input)
		if err != nil {
			t.Fatalf("marshal %T: %v", input, err)
		}
		assertPipelineUserAuditOmitted(t, encoded)
		if got := pipelineJSONMap(t, encoded); !reflect.DeepEqual(got, wantWire) {
			t.Fatalf("marshal %T changed non-audit fields\n got: %#v\nwant: %#v", input, got, wantWire)
		}
		var decoded PipelineEvent
		if err := json.Unmarshal(encoded, &decoded); err != nil {
			t.Fatal(err)
		}
		want := pipelineAuditFixture()
		for user := &want.MonitorStage.User; user != nil; user = user.Master {
			user.Audit = Audit{}
		}
		if !reflect.DeepEqual(decoded, want) {
			t.Fatalf("roundtrip changed non-audit fields\n got: %#v\nwant: %#v", decoded, want)
		}
	}
	for user := &event.MonitorStage.User; user != nil; user = user.Master {
		if user.Audit.CreatedBy != "creator" {
			t.Fatal("marshal mutated source audit")
		}
	}
}

func TestPipelineEventJSONNullAndReuse(t *testing.T) {
	for _, data := range []string{`{}`, `{"monitorStage":null}`} {
		t.Run(data, func(t *testing.T) {
			var event PipelineEvent
			if err := json.Unmarshal([]byte(data), &event); err != nil {
				t.Fatal(err)
			}
			if event.MonitorStage != nil {
				t.Fatal("nil monitor was allocated")
			}
			encoded, err := json.Marshal(event)
			if err != nil {
				t.Fatal(err)
			}
			if _, ok := pipelineJSONMap(t, encoded)["monitorStage"]; ok {
				t.Fatal("nil monitor was serialized")
			}
		})
	}
	for _, tc := range []struct {
		name string
		data string
	}{
		{"missing user", `{"monitorStage":{}}`},
		{"null user", `{"monitorStage":{"user":null}}`},
		{"missing monitor", `{}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			event, want := pipelineAuditFixture(), pipelineAuditFixture()
			if err := json.Unmarshal([]byte(tc.data), &event); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(event, want) {
				t.Fatal("absent/null value overwrote existing fields")
			}
		})
	}
	t.Run("partial objects retain non-audit fields and clear stale audits", func(t *testing.T) {
		event, want := pipelineAuditFixture(), pipelineAuditFixture()
		want.MonitorStage.Name = "updated"
		want.MonitorStage.User.Audit = Audit{}
		want.MonitorStage.User.Master.Audit = Audit{}
		want.MonitorStage.User.Master.Username = "updated-parent"
		if err := json.Unmarshal([]byte(`{"monitorStage":{"name":"updated","user":{"audit":[],"master":{"username":"updated-parent"}}}}`), &event); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(event, want) {
			t.Fatal("partial decode lost non-audit fields or retained stale audit")
		}
	})
	t.Run("explicit null master", func(t *testing.T) {
		event := pipelineAuditFixture()
		if err := json.Unmarshal([]byte(`{"monitorStage":{"user":{"master":null}}}`), &event); err != nil {
			t.Fatal(err)
		}
		if event.MonitorStage.User.Master != nil || event.MonitorStage.User.Username != "child" {
			t.Fatal("null master did not clear only the master pointer")
		}
	})
	t.Run("null followed by duplicate user object", func(t *testing.T) {
		event := pipelineAuditFixture()
		if err := json.Unmarshal([]byte(`{"monitorStage":{"user":null,"user":{"username":"updated","audit":[]}}}`), &event); err != nil {
			t.Fatal(err)
		}
		if event.MonitorStage.User.Username != "updated" || event.MonitorStage.User.Audit != (Audit{}) {
			t.Fatal("user object after null was lost")
		}
	})
	t.Run("explicit null monitor", func(t *testing.T) {
		event := pipelineAuditFixture()
		if err := json.Unmarshal([]byte(`{"monitorStage":null}`), &event); err != nil {
			t.Fatal(err)
		}
		if event.MonitorStage != nil {
			t.Fatal("null monitor did not clear the pointer")
		}
	})
	t.Run("standalone null monitor is a no-op", func(t *testing.T) {
		stage, want := pipelineAuditFixture().MonitorStage, pipelineAuditFixture().MonitorStage
		if err := json.Unmarshal([]byte(`null`), stage); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(stage, want) {
			t.Fatal("null changed a non-pointer monitor destination")
		}
	})
}

func TestPipelineEventJSONRejectsInvalidData(t *testing.T) {
	for _, data := range []string{
		`{"monitorStage":{"user":{"audit":[}}}`,
		`{"monitorStage":{"user":{"audit":{"createdBy":}}}}`,
		`{"monitorStage":{"user":{"master":{"audit":[1,]}}}}`,
		`{"monitorStage":[]}`,
		`{"monitorStage":{"user":[]}}`,
		`{"monitorStage":{"user":{"master":[]}}}`,
		`{"monitorStage":{"user":{"organisationId":"invalid"}}}`,
		`{"monitorStage":{"user":{"activity":"invalid","audit":[]}}}`,
		`{"monitorStage":{"user":{"audit":[]}}} trailing`,
	} {
		t.Run(data, func(t *testing.T) {
			var event PipelineEvent
			if err := json.Unmarshal([]byte(data), &event); err == nil {
				t.Fatal("accepted malformed JSON or invalid non-audit field")
			}
		})
	}
}

func TestPipelineEventJSONRejectsMasterCycles(t *testing.T) {
	event := pipelineAuditFixture()
	event.MonitorStage.User.Master.Master = &event.MonitorStage.User
	if _, err := json.Marshal(event); err == nil {
		t.Fatal("accepted cyclic master chain")
	}
}

func TestPipelineAuditStandaloneJSONAndBSONUnchanged(t *testing.T) {
	event := pipelineAuditFixture()
	user := event.MonitorStage.User
	encoded, err := json.Marshal(user)
	if err != nil {
		t.Fatal(err)
	}
	wire := pipelineJSONMap(t, encoded)
	for current := wire; current != nil; {
		if current["audit"].(map[string]any)["createdBy"] != "creator" {
			t.Fatal("standalone JSON lost audit")
		}
		current, _ = current["master"].(map[string]any)
	}
	var decodedUser User
	if err := json.Unmarshal(encoded, &decodedUser); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(user, decodedUser) {
		t.Fatal("standalone User JSON changed")
	}
	if err := json.Unmarshal([]byte(`{"audit":[]}`), &decodedUser); err == nil {
		t.Fatal("standalone User unexpectedly accepted legacy audit")
	}
	type userBSON User
	type monitorBSON MonitorStage
	for _, tc := range []struct {
		name     string
		value    any
		baseline any
		target   any
	}{
		{"user", &user, (*userBSON)(&user), &User{}},
		{"monitor", event.MonitorStage, (*monitorBSON)(event.MonitorStage), &MonitorStage{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := bson.Marshal(tc.value)
			if err != nil {
				t.Fatal(err)
			}
			want, err := bson.Marshal(tc.baseline)
			if err != nil {
				t.Fatal(err)
			}
			var gotDocument, wantDocument bson.M
			if err := bson.Unmarshal(got, &gotDocument); err != nil {
				t.Fatal(err)
			}
			if err := bson.Unmarshal(want, &wantDocument); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(gotDocument, wantDocument) {
				t.Fatal("BSON representation changed")
			}
			if err := bson.Unmarshal(got, tc.target); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(tc.value, tc.target) {
				t.Fatal("BSON roundtrip lost data")
			}
		})
	}
}

func pipelineJSONMap(t *testing.T, data []byte) map[string]any {
	t.Helper()
	var wire map[string]any
	if err := json.Unmarshal(data, &wire); err != nil {
		t.Fatal(err)
	}
	return wire
}

func assertPipelineUserAuditOmitted(t *testing.T, data []byte) {
	t.Helper()
	wire := pipelineJSONMap(t, data)
	for user := wire["monitorStage"].(map[string]any)["user"].(map[string]any); user != nil; {
		if _, ok := user["audit"]; ok {
			t.Fatal("pipeline emitted user audit")
		}
		user, _ = user["master"].(map[string]any)
	}
	// The historical reader expects an array at every user level.
	type legacyUser struct {
		Audit  []json.RawMessage `json:"audit"`
		Master *legacyUser       `json:"master"`
	}
	var oldReader struct {
		MonitorStage struct {
			User legacyUser `json:"user"`
		} `json:"monitorStage"`
	}
	if err := json.Unmarshal(data, &oldReader); err != nil {
		t.Fatalf("old array-based reader rejected message: %v", err)
	}
}

func pipelineAuditFixture() PipelineEvent {
	id, _ := primitive.ObjectIDFromHex("111111111111111111111111")
	org, _ := primitive.ObjectIDFromHex("222222222222222222222222")
	project, _ := primitive.ObjectIDFromHex("333333333333333333333333")
	stamp := time.Date(2026, 9, 21, 10, 0, 0, 0, time.UTC)
	audit := Audit{CreatedBy: "creator", CreatedAt: stamp, UpdatedBy: "editor", UpdatedAt: stamp, LastAction: "user.updated"}
	activity := Activity{Day: "2026-09-21", Requests: 7, Videos: 3, Images: 2, Usage: 40, Timestamp: 1706000000, Devices: map[string]int64{"camera": 3}}
	highUpload := HighUpload{Requests: 10, StartTimestamp: 1706000000, Notification: 1}
	subscription := Subscription{Id: id, OrganisationId: org, UserId: "payer", StripeId: "stripe", StripePlan: "business", Quantity: 5, Status: "active", CreatedAt: stamp}
	user := User{
		Id: id, OrganisationId: org, ProjectId: project, Username: "child",
		Password: "password-fixture", Email: "fixture@example.invalid", Domain: "domain",
		Role: "owner", CustomRole: "custom", RoleLevel: 2, Days: []string{"2026-09-21"},
		GoogleMFASecret: "mfa-fixture", GoogleMFAEnabled: true, Mfa: true, ForceMFA: 1,
		Audit: audit, FirstName: "First", LastName: "Last", Timezone: "UTC",
		IsActive: 1, ReachedLimit: true, ReachedLimitTimestamp: 1706000000,
		CreatedAt: stamp, UpdatedAt: stamp, Subscription: subscription,
		Storage:        Storage{Uri: "https://example.invalid/vault", AccessKey: "access-fixture", Secret: "secret-fixture", Provider: "vault"},
		ArchiveStorage: Storage{Uri: "https://example.invalid/archive", AccessKey: "archive-fixture", Secret: "archive-secret-fixture", Provider: "archive"},
		HighUpload:     highUpload, Permissions: Permissions{Pages: []string{"media", "devices"}},
		NotificationSettings: &NotificationSettings{Detections: Detections{Enabled: true, ChannelsAll: true}},
		Channels:             &Channels{Pushbullet: Pushbullet{Enabled: true, Valid: true, Apikey: "push-fixture"}},
		Encryption:           Encryption{Enabled: true, HasPassphrase: true, Fingerprint: "fingerprint-fixture", PublicKey: "encryption-fixture", SymmetricKey: "symmetric-fixture"},
		Activity:             []Activity{activity}, Sites: []string{"site"}, Groups: []string{"group"}, Cameras: []string{"camera"},
		Settings:         map[string]any{"enabled": true, "label": "fixture"},
		CustomUsageLimit: 100, CustomDayLimit: 14, CustomAnalysisLimit: 50,
		Plan: "business", PrivateCloud: true, PrivateCloudPlan: "private",
		PublicKey: "public-fixture", PrivateKey: "private-fixture", Bucket: "bucket", Region: "region",
		MasterAccount: "stable-owner", StripeId: "stripe-user", Coupons: []string{"coupon"},
		HLSCallbackURL: "https://example.invalid/hls", OAuthClientID: "oauth-id", OAuthClientSecret: "oauth-fixture",
		Master: &User{Id: org, Username: "parent", Audit: audit, PrivateKey: "parent-fixture",
			Master: &User{Id: project, Username: "grandparent", Audit: audit}},
	}
	return PipelineEvent{
		Request: "persist", Operation: "analysis", Stages: []string{"monitor", "analysis"},
		EventStage: &EventStage{Name: "event", EventData: "event-data", SourceDevice: &PipelineSourceDevice{
			DeviceId: id, DeviceKey: "camera", CloudKey: "cloud", OrganisationId: org.Hex(), ProjectId: &project, OwnerUserId: "stable-owner",
		}},
		MonitorStage: &MonitorStage{Name: "monitor", MonitorData: "monitor-data", OrganisationId: id.Hex(), ProjectId: &org,
			User: user, Subscription: subscription, Plans: map[string]Plan{"business": {Level: 4, UploadLimit: 100, VideoLimit: 200, Usage: 300, AnalysisLimit: 400, DayLimit: 30}}, Activity: activity, HighUpload: highUpload},
		SequenceStage:     &SequenceStage{Name: "sequence", SequenceId: 123},
		AnalysisStage:     &AnalysisStage{Name: "analysis", AnalysisResult: "result"},
		ThrottlerStage:    &ThrottlerStage{Name: "throttler", ThrottleLimit: 12},
		NotificationStage: &NotificationStage{Name: "notification", NotificationType: "email"},
		Storage:           "vault", Provider: "provider", SecondaryProviders: []string{"secondary"},
		TraceId: "trace-fixture", ReceiveCount: 2, Timestamp: 1706000000, FileName: "child/video.mp4",
		Payload: PipelinePayload{Timestamp: 1706000000, FileName: "child/video.mp4", FileSize: 1234, Duration: "3000",
			SourceVaultId: "vault", SourceMediaId: &id, ForwardedAt: 1706000001, Encrypted: true,
			SignedURL: "https://example.invalid/video", OrganisationId: project.Hex(), DeviceId: "camera", DeviceName: "Camera",
			IsFragmented: true, BytesRanges: "0-100", Result: json.RawMessage(`{"result":"fixture"}`),
			Metadata: PipelineMetadata{Timestamp: "1706000000", Duration: "3000", DeviceId: "camera", DeviceName: "Camera", FPS: "25"}},
		Data: map[string]any{"credentials": map[string]any{"key": "fixture"}, "traceparent": "trace-parent"},
	}
}
