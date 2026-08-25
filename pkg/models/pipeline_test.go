package models

import (
	"encoding/json"
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestPipelineEventOwnershipSnapshotRoundTripAndGetMedia(t *testing.T) {
	organisationId := primitive.NewObjectID().Hex()
	projectId := primitive.NewObjectID()
	activeUserOrganisationId := primitive.NewObjectID()
	activeUserProjectId := primitive.NewObjectID()
	event := PipelineEvent{
		MonitorStage: &MonitorStage{
			OrganisationId: organisationId,
			ProjectId:      &projectId,
			User: User{
				OrganisationId: activeUserOrganisationId,
				ProjectId:      activeUserProjectId,
			},
		},
		Payload: PipelinePayload{
			FileName:       "user/video.mp4",
			OrganisationId: primitive.NewObjectID().Hex(),
			Metadata: PipelineMetadata{
				DeviceId:  "device-1",
				Timestamp: "1706000000",
				Duration:  "3000",
			},
		},
	}

	encoded, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal pipeline event: %v", err)
	}
	var decoded PipelineEvent
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal pipeline event: %v", err)
	}
	if decoded.MonitorStage == nil || decoded.MonitorStage.ProjectId == nil || *decoded.MonitorStage.ProjectId != projectId {
		t.Fatalf("monitor projectId = %#v, want %s", decoded.MonitorStage, projectId.Hex())
	}
	if decoded.MonitorStage.OrganisationId != organisationId {
		t.Fatalf("monitor organisationId = %q, want %q", decoded.MonitorStage.OrganisationId, organisationId)
	}

	media, err := decoded.GetMedia()
	if err != nil {
		t.Fatalf("get media: %v", err)
	}
	if media.ProjectId == nil || *media.ProjectId != projectId {
		t.Fatalf("media projectId = %v, want %s", media.ProjectId, projectId.Hex())
	}
	if media.OrganisationId != organisationId {
		t.Fatalf("media organisationId = %q, want %q", media.OrganisationId, organisationId)
	}
	if media.OrganisationId == activeUserOrganisationId.Hex() {
		t.Fatalf("media organisationId followed mutable user selection %s", activeUserOrganisationId.Hex())
	}
	if media.OrganisationId == decoded.Payload.OrganisationId {
		t.Fatalf("media organisationId followed request payload %s", decoded.Payload.OrganisationId)
	}
	if *media.ProjectId == activeUserProjectId {
		t.Fatalf("media projectId followed mutable user selection %s", activeUserProjectId.Hex())
	}
	if media.ProjectId == decoded.MonitorStage.ProjectId {
		t.Fatal("media projectId must be a copy, not alias monitor-stage storage")
	}
}

func TestPipelineEventWithoutOwnershipSnapshotRemainsCompatible(t *testing.T) {
	legacyJSON := []byte(`{"monitorStage":{"name":"monitor","user":{"organisationId":"111111111111111111111111","projectId":"222222222222222222222222"}},"payload":{"key":"user/video.mp4","organisationId":"333333333333333333333333","metadata":{"productid":"device-1","event-timestamp":"1706000000","duration":"3000"}}}`)

	var event PipelineEvent
	if err := json.Unmarshal(legacyJSON, &event); err != nil {
		t.Fatalf("unmarshal legacy pipeline event: %v", err)
	}
	media, err := event.GetMedia()
	if err != nil {
		t.Fatalf("get media from legacy event: %v", err)
	}
	if media.ProjectId != nil {
		t.Fatalf("legacy media projectId = %s, want nil", media.ProjectId.Hex())
	}
	if media.OrganisationId != "" {
		t.Fatalf("legacy media organisationId = %q, want empty", media.OrganisationId)
	}
}

func TestPipelineSourceDeviceRoundTripsCanonicalIdentity(t *testing.T) {
	deviceId := primitive.NewObjectID()
	organisationId := primitive.NewObjectID()
	projectId := primitive.NewObjectID()
	ownerUserId := primitive.NewObjectID()
	event := PipelineEvent{
		EventStage: &EventStage{SourceDevice: &PipelineSourceDevice{
			DeviceId:       deviceId,
			DeviceKey:      "device-1",
			OrganisationId: organisationId.Hex(),
			ProjectId:      &projectId,
			OwnerUserId:    ownerUserId.Hex(),
		}},
	}

	encoded, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal pipeline event: %v", err)
	}
	var decoded PipelineEvent
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal pipeline event: %v", err)
	}
	if decoded.EventStage == nil || decoded.EventStage.SourceDevice == nil {
		t.Fatalf("source device = %#v", decoded.EventStage)
	}
	got := decoded.EventStage.SourceDevice
	if got.DeviceId != deviceId || got.DeviceKey != "device-1" || got.OrganisationId != organisationId.Hex() || got.OwnerUserId != ownerUserId.Hex() {
		t.Fatalf("source device = %#v", got)
	}
	if got.ProjectId == nil || *got.ProjectId != projectId {
		t.Fatalf("source project = %v, want %s", got.ProjectId, projectId.Hex())
	}
}

func TestPipelineEventOwnershipSnapshotAppliesToLegacyMedia(t *testing.T) {
	organisationId := primitive.NewObjectID().Hex()
	projectId := primitive.NewObjectID()
	event := PipelineEvent{
		MonitorStage: &MonitorStage{
			OrganisationId: organisationId,
			ProjectId:      &projectId,
		},
		Payload: PipelinePayload{
			FileName: "username/1640000000_region_devicename_motion_1234_5000.mp4",
		},
	}

	media, err := event.GetMedia()
	if err != nil {
		t.Fatalf("get media from legacy filename: %v", err)
	}
	if media.OrganisationId != organisationId {
		t.Fatalf("media organisationId = %q, want %q", media.OrganisationId, organisationId)
	}
	if media.ProjectId == nil || *media.ProjectId != projectId {
		t.Fatalf("media projectId = %v, want %s", media.ProjectId, projectId.Hex())
	}
}

func TestGetMediaFromPipelineEvent_LegacyFormat(t *testing.T) {
	tests := []struct {
		name          string
		pipelineEvent PipelineEvent
		wantMedia     Media
		wantErr       bool
	}{
		{
			name: "valid legacy format",
			pipelineEvent: PipelineEvent{
				Storage: "s3",
				Payload: PipelinePayload{
					FileName: "username/1640000000_region_devicename_motion_1234_5000.mp4",
					Metadata: PipelineMetadata{
						DeviceId: "", // Empty to trigger legacy parsing
					},
				},
			},
			wantMedia: Media{
				VideoFile:       "username/1640000000_region_devicename_motion_1234_5000.mp4",
				StartTimestamp:  1640000000,
				DeviceName:      "devicename",
				DeviceKey:       "devicename",
				Duration:        5000,
				StorageSolution: "s3",
				Metadata: &MediaMetadata{
					MotionPixels: 0,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid path format - missing slash",
			pipelineEvent: PipelineEvent{
				Payload: PipelinePayload{
					FileName: "invalidpath.mp4",
					Metadata: PipelineMetadata{
						DeviceId: "",
					},
				},
			},
			wantMedia: Media{},
			wantErr:   true,
		},
		{
			name: "invalid filename format - missing dot",
			pipelineEvent: PipelineEvent{
				Payload: PipelinePayload{
					FileName: "username/invalidfilename",
					Metadata: PipelineMetadata{
						DeviceId: "",
					},
				},
			},
			wantMedia: Media{},
			wantErr:   true,
		},
		{
			name: "invalid attributes - wrong number of parts",
			pipelineEvent: PipelineEvent{
				Payload: PipelinePayload{
					FileName: "username/1640000000_region.mp4",
					Metadata: PipelineMetadata{
						DeviceId: "",
					},
				},
			},
			wantMedia: Media{},
			wantErr:   true,
		},
		{
			name: "invalid timestamp",
			pipelineEvent: PipelineEvent{
				Payload: PipelinePayload{
					FileName: "username/invalid_region_devicename_motion_1234_5000.mp4",
					Metadata: PipelineMetadata{
						DeviceId: "",
					},
				},
			},
			wantMedia: Media{
				VideoFile:       "username/invalid_region_devicename_motion_1234_5000.mp4",
				StartTimestamp:  0,
				DeviceName:      "devicename",
				DeviceKey:       "devicename",
				Duration:        5000,
				StorageSolution: "",
				Metadata:        &MediaMetadata{MotionPixels: 0},
			},
			wantErr: false,
		},
		{
			name: "invalid motion pixels",
			pipelineEvent: PipelineEvent{
				Payload: PipelinePayload{
					FileName: "username/1640000000_region_devicename_motion_invalid_5000.mp4",
					Metadata: PipelineMetadata{
						DeviceId: "",
					},
				},
			},
			wantMedia: Media{
				VideoFile:       "username/1640000000_region_devicename_motion_invalid_5000.mp4",
				StartTimestamp:  1640000000,
				DeviceName:      "devicename",
				DeviceKey:       "devicename",
				Duration:        5000,
				StorageSolution: "",
				Metadata:        &MediaMetadata{MotionPixels: 0},
			},
			wantErr: false,
		},
		{
			name: "invalid duration",
			pipelineEvent: PipelineEvent{
				Payload: PipelinePayload{
					FileName: "username/1640000000_region_devicename_motion_1234_invalid.mp4",
					Metadata: PipelineMetadata{
						DeviceId: "",
					},
				},
			},
			wantMedia: Media{
				VideoFile:       "username/1640000000_region_devicename_motion_1234_invalid.mp4",
				StartTimestamp:  1640000000,
				DeviceName:      "devicename",
				DeviceKey:       "devicename",
				Duration:        0,
				StorageSolution: "",
				Metadata:        &MediaMetadata{MotionPixels: 0},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMedia, err := tt.pipelineEvent.GetMedia()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMediaFromPipelineEvent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if gotMedia.VideoFile != tt.wantMedia.VideoFile {
					t.Errorf("VideoFile = %v, want %v", gotMedia.VideoFile, tt.wantMedia.VideoFile)
				}
				if gotMedia.StartTimestamp != tt.wantMedia.StartTimestamp {
					t.Errorf("StartTimestamp = %v, want %v", gotMedia.StartTimestamp, tt.wantMedia.StartTimestamp)
				}
				if gotMedia.DeviceName != tt.wantMedia.DeviceName {
					t.Errorf("DeviceName = %v, want %v", gotMedia.DeviceName, tt.wantMedia.DeviceName)
				}
				if gotMedia.DeviceKey != tt.wantMedia.DeviceKey {
					t.Errorf("DeviceKey = %v, want %v", gotMedia.DeviceKey, tt.wantMedia.DeviceKey)
				}
				if gotMedia.Duration != tt.wantMedia.Duration {
					t.Errorf("Duration = %v, want %v", gotMedia.Duration, tt.wantMedia.Duration)
				}
				if gotMedia.StorageSolution != tt.wantMedia.StorageSolution {
					t.Errorf("StorageSolution = %v, want %v", gotMedia.StorageSolution, tt.wantMedia.StorageSolution)
				}
				if tt.wantMedia.Metadata != nil {
					if gotMedia.Metadata == nil {
						t.Errorf("Metadata is nil, want %+v", tt.wantMedia.Metadata)
					} else if gotMedia.Metadata.MotionPixels != tt.wantMedia.Metadata.MotionPixels {
						t.Errorf("Metadata.MotionPixels = %v, want %v", gotMedia.Metadata.MotionPixels, tt.wantMedia.Metadata.MotionPixels)
					}
				}
			}
		})
	}
}

func TestGetMediaFromPipelineEvent_NewFormat(t *testing.T) {
	tests := []struct {
		name          string
		pipelineEvent PipelineEvent
		wantMedia     Media
		wantErr       bool
	}{
		{
			name: "valid new format",
			pipelineEvent: PipelineEvent{
				Storage: "minio",
				Payload: PipelinePayload{
					FileName: "path/to/video.mp4",
					Metadata: PipelineMetadata{
						DeviceId:   "device123",
						DeviceName: "Front Camera",
						Timestamp:  "1640000000",
						Duration:   "3000",
					},
				},
			},
			wantMedia: Media{
				VideoFile:       "path/to/video.mp4",
				StartTimestamp:  1640000000,
				DeviceName:      "Front Camera",
				DeviceKey:       "device123",
				Duration:        3000,
				StorageSolution: "minio",
			},
			wantErr: false,
		},
		{
			name: "invalid timestamp in new format",
			pipelineEvent: PipelineEvent{
				Payload: PipelinePayload{
					FileName: "path/to/video.mp4",
					Metadata: PipelineMetadata{
						DeviceId:   "device123",
						DeviceName: "Front Camera",
						Timestamp:  "invalid",
						Duration:   "3000",
					},
				},
			},
			wantMedia: Media{
				VideoFile:       "path/to/video.mp4",
				StartTimestamp:  0,
				DeviceName:      "Front Camera",
				DeviceKey:       "device123",
				Duration:        3000,
				StorageSolution: "",
			},
			wantErr: false,
		},
		{
			name: "invalid duration in new format",
			pipelineEvent: PipelineEvent{
				Payload: PipelinePayload{
					FileName: "path/to/video.mp4",
					Metadata: PipelineMetadata{
						DeviceId:   "device123",
						DeviceName: "Front Camera",
						Timestamp:  "1640000000",
						Duration:   "invalid",
					},
				},
			},
			wantMedia: Media{
				VideoFile:       "path/to/video.mp4",
				StartTimestamp:  1640000000,
				DeviceName:      "Front Camera",
				DeviceKey:       "device123",
				Duration:        0,
				StorageSolution: "",
			},
			wantErr: false,
		},
		{
			name: "new format with empty device name",
			pipelineEvent: PipelineEvent{
				Storage: "azure",
				Payload: PipelinePayload{
					FileName: "path/to/video.mp4",
					Metadata: PipelineMetadata{
						DeviceId:   "device456",
						DeviceName: "",
						Timestamp:  "1650000000",
						Duration:   "2500",
					},
				},
			},
			wantMedia: Media{
				VideoFile:       "path/to/video.mp4",
				StartTimestamp:  1650000000,
				DeviceName:      "",
				DeviceKey:       "device456",
				Duration:        2500,
				StorageSolution: "azure",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMedia, err := tt.pipelineEvent.GetMedia()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMediaFromPipelineEvent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if gotMedia.VideoFile != tt.wantMedia.VideoFile {
					t.Errorf("VideoFile = %v, want %v", gotMedia.VideoFile, tt.wantMedia.VideoFile)
				}
				if gotMedia.StartTimestamp != tt.wantMedia.StartTimestamp {
					t.Errorf("StartTimestamp = %v, want %v", gotMedia.StartTimestamp, tt.wantMedia.StartTimestamp)
				}
				if gotMedia.DeviceName != tt.wantMedia.DeviceName {
					t.Errorf("DeviceName = %v, want %v", gotMedia.DeviceName, tt.wantMedia.DeviceName)
				}
				if gotMedia.DeviceKey != tt.wantMedia.DeviceKey {
					t.Errorf("DeviceKey = %v, want %v", gotMedia.DeviceKey, tt.wantMedia.DeviceKey)
				}
				if gotMedia.Duration != tt.wantMedia.Duration {
					t.Errorf("Duration = %v, want %v", gotMedia.Duration, tt.wantMedia.Duration)
				}
				if gotMedia.StorageSolution != tt.wantMedia.StorageSolution {
					t.Errorf("StorageSolution = %v, want %v", gotMedia.StorageSolution, tt.wantMedia.StorageSolution)
				}
			}
		})
	}
}

func TestPipelineEvent_GetMedia(t *testing.T) {
	tests := []struct {
		name          string
		event         PipelineEvent
		expectedMedia Media
		expectError   bool
		errorContains string
	}{
		{
			name: "new format with metadata - valid",
			event: PipelineEvent{
				Storage:  "s3",
				Provider: "aws",
				Payload: PipelinePayload{
					FileName: "user123/video.mp4",
					Metadata: PipelineMetadata{
						DeviceId:   "device-001",
						DeviceName: "Front Camera",
						Duration:   "300",
						Timestamp:  "1706000000",
						FPS:        "17.35",
					},
				},
			},
			expectedMedia: Media{
				VideoFile:       "user123/video.mp4",
				DeviceKey:       "device-001",
				DeviceName:      "Front Camera",
				Duration:        300,
				StartTimestamp:  1706000000,
				StorageSolution: "s3",
				VideoProvider:   "aws",
				Metadata:        &MediaMetadata{FPS: 17.35},
			},
			expectError: false,
		},
		{
			name: "legacy format - valid 6 attributes",
			event: PipelineEvent{
				Storage:  "gcs",
				Provider: "google",
				Payload: PipelinePayload{
					FileName: "username/1706000000_attr1_camera1_attr3_500_60.mp4",
					Metadata: PipelineMetadata{
						DeviceId: "", // Empty triggers legacy parsing
					},
				},
			},
			expectedMedia: Media{
				VideoFile:       "username/1706000000_attr1_camera1_attr3_500_60.mp4",
				DeviceKey:       "camera1",
				DeviceName:      "camera1",
				Duration:        60,
				StartTimestamp:  1706000000,
				StorageSolution: "gcs",
				VideoProvider:   "google",
			},
			expectError: false,
		},
		{
			name: "invalid path - missing slash",
			event: PipelineEvent{
				Payload: PipelinePayload{
					FileName: "invalidpath",
					Metadata: PipelineMetadata{
						DeviceId: "",
					},
				},
			},
			expectError:   true,
			errorContains: "invalid file path format",
		},
		{
			name: "legacy format - invalid filename without extension",
			event: PipelineEvent{
				Payload: PipelinePayload{
					FileName: "username/videofile",
					Metadata: PipelineMetadata{
						DeviceId: "",
					},
				},
			},
			expectError:   true,
			errorContains: "invalid video file name format",
		},
		{
			name: "legacy format - wrong number of attributes",
			event: PipelineEvent{
				Payload: PipelinePayload{
					FileName: "username/attr1_attr2_attr3.mp4",
					Metadata: PipelineMetadata{
						DeviceId: "",
					},
				},
			},
			expectError:   true,
			errorContains: "invalid attributes format",
		},
		{
			name: "new format with metadata - empty duration",
			event: PipelineEvent{
				Storage:  "s3",
				Provider: "aws",
				Payload: PipelinePayload{
					FileName: "user123/video.mp4",
					Metadata: PipelineMetadata{
						DeviceId:   "device-001",
						DeviceName: "Front Camera",
						Duration:   "",
						Timestamp:  "1706000000",
					},
				},
			},
			expectedMedia: Media{
				VideoFile:       "user123/video.mp4",
				DeviceKey:       "device-001",
				DeviceName:      "Front Camera",
				Duration:        0, // Empty string parses to 0
				StartTimestamp:  1706000000,
				StorageSolution: "s3",
				VideoProvider:   "aws",
			},
			expectError: false,
		},
		{
			name: "legacy format - only filename with extension",
			event: PipelineEvent{
				Storage:  "azure",
				Provider: "blob",
				Payload: PipelinePayload{
					FileName: "username/1706000000_x_cam_y_100_30.mp4",
					Metadata: PipelineMetadata{
						DeviceId: "",
					},
				},
			},
			expectedMedia: Media{
				VideoFile:       "username/1706000000_x_cam_y_100_30.mp4",
				DeviceKey:       "cam",
				DeviceName:      "cam",
				Duration:        30,
				StartTimestamp:  1706000000,
				StorageSolution: "azure",
				VideoProvider:   "blob",
			},
			expectError: false,
		},
		{
			name: "new format - zero values parsed correctly",
			event: PipelineEvent{
				Storage:  "local",
				Provider: "disk",
				Payload: PipelinePayload{
					FileName: "user/test.mp4",
					Metadata: PipelineMetadata{
						DeviceId:   "dev-123",
						DeviceName: "Test Device",
						Duration:   "0",
						Timestamp:  "0",
					},
				},
			},
			expectedMedia: Media{
				VideoFile:       "user/test.mp4",
				DeviceKey:       "dev-123",
				DeviceName:      "Test Device",
				Duration:        0,
				StartTimestamp:  0,
				StorageSolution: "local",
				VideoProvider:   "disk",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			media, err := tt.event.GetMedia()

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errorContains)
					return
				}
				if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error containing %q, got %q", tt.errorContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if media.VideoFile != tt.expectedMedia.VideoFile {
				t.Errorf("VideoFile: got %q, want %q", media.VideoFile, tt.expectedMedia.VideoFile)
			}
			if media.DeviceKey != tt.expectedMedia.DeviceKey {
				t.Errorf("DeviceKey: got %q, want %q", media.DeviceKey, tt.expectedMedia.DeviceKey)
			}
			if media.DeviceName != tt.expectedMedia.DeviceName {
				t.Errorf("DeviceName: got %q, want %q", media.DeviceName, tt.expectedMedia.DeviceName)
			}
			if media.Duration != tt.expectedMedia.Duration {
				t.Errorf("Duration: got %d, want %d", media.Duration, tt.expectedMedia.Duration)
			}
			if media.StartTimestamp != tt.expectedMedia.StartTimestamp {
				t.Errorf("StartTimestamp: got %d, want %d", media.StartTimestamp, tt.expectedMedia.StartTimestamp)
			}
			if media.StorageSolution != tt.expectedMedia.StorageSolution {
				t.Errorf("StorageSolution: got %q, want %q", media.StorageSolution, tt.expectedMedia.StorageSolution)
			}
			if media.VideoProvider != tt.expectedMedia.VideoProvider {
				t.Errorf("VideoProvider: got %q, want %q", media.VideoProvider, tt.expectedMedia.VideoProvider)
			}
			if media.Metadata != nil && tt.expectedMedia.Metadata != nil && media.Metadata.FPS != tt.expectedMedia.Metadata.FPS {
				t.Errorf("Metadata.FPS: got %v, want %v", media.Metadata.FPS, tt.expectedMedia.Metadata.FPS)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && searchSubstring(s, substr)))
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
