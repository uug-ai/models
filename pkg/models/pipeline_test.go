package models

import (
	"testing"
)

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
				DeviceId:        "devicename",
				Duration:        5000,
				StorageSolution: "s3",
				Metadata: &MediaMetadata{
					MotionPixels: 1234,
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
			wantMedia: Media{},
			wantErr:   true,
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
			wantMedia: Media{},
			wantErr:   true,
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
			wantMedia: Media{},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMedia, err := GetMediaFromPipelineEvent(tt.pipelineEvent)
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
				if gotMedia.DeviceId != tt.wantMedia.DeviceId {
					t.Errorf("DeviceId = %v, want %v", gotMedia.DeviceId, tt.wantMedia.DeviceId)
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
				DeviceId:        "device123",
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
			wantMedia: Media{},
			wantErr:   true,
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
			wantMedia: Media{},
			wantErr:   true,
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
				DeviceId:        "device456",
				Duration:        2500,
				StorageSolution: "azure",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMedia, err := GetMediaFromPipelineEvent(tt.pipelineEvent)
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
				if gotMedia.DeviceId != tt.wantMedia.DeviceId {
					t.Errorf("DeviceId = %v, want %v", gotMedia.DeviceId, tt.wantMedia.DeviceId)
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
