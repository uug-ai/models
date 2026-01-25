package models

import (
	"fmt"
	"strconv"
	"strings"
)

// --------------------------------------------------------------------------------------------------------------------------------
// Pipeline represents a data processing pipeline that can handle various stages of data processing.
// The idea of the pipeline is to process data in a series of steps, where the output of one step becomes the input for the next.
// The initial pipeline object is expanded with each stage of processing.
//
// Pipeline stages:
//
//     event
//       ↓
//    monitor
//       ↓
//    sequence
//       ↓
//    analysis
//       ↓
//    throttler
//       ↓
//    notification
//
// Data flows through each stage sequentially, with relevant information persisted at each step.

type PipelineEvent struct {
	Request   string `json:"request,omitempty"` // ondemand, persist
	Operation string `json:"operation,omitempty"`

	// Stages of the pipeline, e.g., event, monitor, sequence, analysis, throttler, notification
	// Idea is that we persist relevant data in each stage, so we have a good understanding what is used
	// or computed at which stage.
	Stages            []string           `json:"events,omitempty"`
	EventStage        *EventStage        `json:"eventStage,omitempty"`
	MonitorStage      *MonitorStage      `json:"monitorStage,omitempty"`
	SequenceStage     *SequenceStage     `json:"sequenceStage,omitempty"`
	AnalysisStage     *AnalysisStage     `json:"analysisStage,omitempty"`
	ThrottlerStage    *ThrottlerStage    `json:"throttlerStage,omitempty"`
	NotificationStage *NotificationStage `json:"notificationStage,omitempty"`

	Storage            string   `json:"provider,omitempty"`
	Provider           string   `json:"source,omitempty"`
	SecondaryProviders []string `json:"secondary_providers,omitempty"`

	// We are using OpenTelemetry, so we can observe the pipeline more easily.
	TraceId string `json:"traceId,omitempty"`

	ReceiveCount int64 `json:"receivecount,omitempty"`

	Timestamp int64           `json:"date,omitempty"`
	FileName  string          `json:"fileName,omitempty"`
	Payload   PipelinePayload `json:"payload,omitempty"`

	Data map[string]interface{} `json:"data,omitempty"` // We should get rid of this and use the stage map
}

func (pe *PipelineEvent) GetMedia() (Media, error) {

	// We need to extract attributes from the event.
	// 1. The old way, we need to parse the filename (all info is stored in the filename).
	// 2. (or) Use the metadata attributes.

	media := Media{}

	pathParts := strings.Split(pe.Payload.FileName, "/")
	if len(pathParts) < 2 {
		return media, fmt.Errorf("invalid file path format, expected at least 2 parts separated by '/', got: %s", pe.Payload.FileName) // Return empty media if path format is invalid
	}

	// If we have a DeviceId in the metadata, we can use the structured metadata way.
	// This is the new and recommended approach.
	if pe.Payload.Metadata.DeviceId != "" {
		// Set video file
		media.VideoFile = pe.Payload.FileName

		// Get device information from metadata
		media.DeviceName = pe.Payload.Metadata.DeviceName
		media.DeviceId = pe.Payload.Metadata.DeviceId

		// Get time information from the recording
		duration, _ := strconv.ParseInt(pe.Payload.Metadata.Duration, 10, 64)
		media.Duration = int(duration)
		media.StartTimestamp, _ = strconv.ParseInt(pe.Payload.Metadata.Timestamp, 10, 64)

		// Information about where the media is stored and provided from
		media.StorageSolution = pe.Storage
		media.VideoProvider = pe.Provider

		return media, nil
	}

	// Legacy way, parse from filename the different fields. We used to have only the filename, and used the filename to store all attributes.
	// As this was not very flexible, we moved to structured metadata, but we still need to support the legacy way if there
	// is a legacy Vault in place.

	username := pathParts[0]
	_ = username // Currently not used, but could be useful in the future.
	videoFileName := pathParts[1]
	videoFileNamePieces := strings.Split(videoFileName, ".")
	if len(videoFileNamePieces) < 2 {
		return media, fmt.Errorf("invalid video file name format, expected at least 2 parts separated by '.', got: %s", videoFileName) // Return empty media if filename format is invalid
	}

	// Extract attributes from the video file name
	videoFileAttribute := videoFileNamePieces[len(videoFileNamePieces)-2]
	attributes := strings.Split(videoFileAttribute, "_")
	if len(attributes) == 6 {

		// Set video file
		media.VideoFile = pe.Payload.FileName

		// Get device information from the filename, we do not have metadata in the legacy way
		media.DeviceName = attributes[2]
		media.DeviceId = attributes[2]

		// Get time information from the filename
		duration, _ := strconv.ParseInt(attributes[5], 10, 64)
		media.Duration = int(duration)
		media.StartTimestamp, _ = strconv.ParseInt(attributes[0], 10, 64)

		// Information about where the media is stored and provided from
		media.StorageSolution = pe.Storage
		media.VideoProvider = pe.Provider

		return media, nil
	}
	return media, fmt.Errorf("invalid attributes format in video file name: %s, expected 6 attributes, got: %d", videoFileName, len(attributes)) // Return empty media if attributes format is invalid

}

type PipelinePayload struct {
	Timestamp int64  `json:"timestamp,omitempty"`
	FileName  string `json:"key,omitempty"`
	FileSize  int64  `json:"fileSize,omitempty"`
	Duration  string `json:"duration,omitempty"`

	// Signed URL
	SignedURL string `json:"signedUrl,omitempty"`

	// RBAC
	OrganisationId string `json:"organisationId,omitempty"`
	DeviceId       string `json:"deviceId,omitempty"`
	DeviceName     string `json:"deviceName,omitempty"`

	// MP4 fragmentation
	IsFragmented     bool                         `json:"is_fragmented" bson:"is_fragmented"`
	BytesRanges      string                       `json:"bytes_ranges" bson:"bytes_ranges"`
	BytesRangeOnTime []FragmentedBytesRangeOnTime `json:"bytes_range_on_time" bson:"bytes_range_on_time"`

	// Metadata
	Metadata PipelineMetadata `json:"metadata,omitempty"`
}

type PipelineMetadata struct {
	Timestamp       string `json:"event-timestamp,omitempty"`
	Duration        string `json:"duration,omitempty"`
	NumberOfChanges string `json:"event-numberofchanges,omitempty"`

	UploadTime   string `json:"uploadtime,omitempty"`
	MicroSeconds string `json:"event-microseconds,omitempty"`

	DeviceId   string `json:"productid,omitempty"`
	DeviceName string `json:"event-instancename,omitempty"`

	RegionCoordinates string `json:"event-regioncoordinates,omitempty"`
}

// HandlerResult represents the result of message handling
type PipelineAction string

const (
	// Forward indicates the message should be forwarded to the next stage
	PipelineForward PipelineAction = "forward"
	// Cancel indicates the message processing should be cancelled
	PipelineCancel PipelineAction = "cancel"
	// Retry indicates the message should be retried
	PipelineRetry PipelineAction = "retry"
	// Error indicates an error occurred during message processing
	PipelineError PipelineAction = "error"
)

// PipelineMetrics represents processing metrics for a pipeline event
type PipelineMetrics struct {
	ProcessingTime float64 `json:"processingTime,omitempty"`
}

// MessageHandler is a function type for handling pipeline events
type MessageHandler func(PipelineEvent, ...any) (PipelineAction, PipelineEvent, int)
type PrometheusHandler func(PipelineMetrics)

// As defined above we have multiple stages, each with its own set of data and processing logic.
// 1. event, 2. monitor, 3. sequence, 4. analysis, 5. throttler, 6. notification
// We'll define a custom struct for each stage's data, however we should be able to use the stage type
// so we can use it interchangeably.

type Stage interface {
	GetName() string
	UnmarshalJSON([]byte) error
}

type EventStage struct {
	Name      string `json:"name,omitempty"`
	EventData string `json:"eventData,omitempty"` // Add fields relevant to event stage
	// Add more fields as needed
}

// Constructor function for EventStage
func NewEventStage() EventStage {
	return EventStage{
		Name: "event",
	}
}
func (e EventStage) GetName() string { return e.Name }

type MonitorStage struct {
	Name        string `json:"name,omitempty"`
	MonitorData string `json:"monitorData,omitempty"` // Add fields relevant to monitor stage

	// Add more fields as needed
	User         User            `json:"user,omitempty"`
	Subscription Subscription    `json:"subscription,omitempty"`
	Plans        map[string]Plan `json:"plans,omitempty"`
	HighUpload   HighUpload      `json:"highUpload,omitempty"`
	Activity     Activity        `json:"activity,omitempty"`
}

// Constructor function for MonitorStage
func NewMonitorStage() MonitorStage {
	return MonitorStage{
		Name: "monitor",
	}
}
func (m MonitorStage) GetName() string           { return m.Name }
func (m MonitorStage) GetUser() User             { return m.User }
func (m MonitorStage) GetHighUpload() HighUpload { return m.HighUpload }
func (m MonitorStage) GetActivity() Activity     { return m.Activity }

type SequenceStage struct {
	Name       string `json:"name,omitempty"`
	SequenceId int64  `json:"sequenceId,omitempty"` // Add fields relevant to sequence stage
	// Add more fields as needed
}

// Constructor function for SequenceStage
func NewSequenceStage() SequenceStage {
	return SequenceStage{
		Name: "sequence",
	}
}
func (s SequenceStage) GetName() string { return s.Name }

type AnalysisStage struct {
	Name           string `json:"name,omitempty"`
	AnalysisResult string `json:"analysisResult,omitempty"` // Add fields relevant to analysis stage
	// Add more fields as needed
}

// Constructor function for AnalysisStage
func NewAnalysisStage() AnalysisStage {
	return AnalysisStage{
		Name: "analysis",
	}
}
func (a AnalysisStage) GetName() string { return a.Name }

type ThrottlerStage struct {
	Name          string `json:"name,omitempty"`
	ThrottleLimit int    `json:"throttleLimit,omitempty"` // Add fields relevant to throttler stage
	// Add more fields as needed
}

// Constructor function for ThrottlerStage
func NewThrottlerStage() ThrottlerStage {
	return ThrottlerStage{
		Name: "throttler",
	}
}
func (t ThrottlerStage) GetName() string { return t.Name }

type NotificationStage struct {
	Name             string `json:"name,omitempty"`
	NotificationType string `json:"notificationType,omitempty"` // Add fields relevant to notification stage
	// Add more fields as needed
}

// Constructor function for NotificationStage
func NewNotificationStage() NotificationStage {
	return NotificationStage{
		Name: "notification",
	}
}
func (n NotificationStage) GetName() string { return n.Name }

// GetMediaFromPipelineEvent extracts a Media instance from the provided PipelineEvent.
// It supports two parsing modes:
//   - Legacy format: when Payload.Metadata.DeviceId is empty, media fields are derived by
//     parsing the Payload.FileName path and filename components.
//   - New format: when Payload.Metadata.DeviceId is set, media fields are populated from
//     the structured payload and metadata present on the event.
//
// The function returns a zero-value Media and a non-nil error if the event data or filename
// do not conform to the expected formats required to populate the Media fields.

func GetMediaFromPipelineEvent(pipelineEvent PipelineEvent) (Media, error) {
	media := Media{}

	// If DeviceId is not set in metadata, try to parse from filename (this is the legacy way)
	if pipelineEvent.Payload.Metadata.DeviceId == "" {
		pathParts := strings.Split(pipelineEvent.Payload.FileName, "/")
		if len(pathParts) < 2 {
			return media, fmt.Errorf("invalid path format: %s", pipelineEvent.Payload.FileName) // Return empty media if path format is invalid
		}
		// @TODO Fix for users with a . in the username.
		// Could be the case that there is a dot in the username.
		fileName := pathParts[1]
		fileNamePieces := strings.Split(fileName, ".")
		if len(fileNamePieces) < 2 {
			return media, fmt.Errorf("invalid filename format: %s", fileName) // Return empty media if filename format is invalid
		}

		media.VideoFile = pipelineEvent.Payload.FileName
		fileName = fileNamePieces[len(fileNamePieces)-2]
		attributes := strings.Split(fileName, "_")
		if len(attributes) == 6 {
			// Set other fields..
			startTimestamp, err := strconv.ParseInt(attributes[0], 10, 64)
			if err != nil {
				return media, fmt.Errorf("invalid timestamp format: %s", attributes[0])
			}
			media.StartTimestamp = startTimestamp
			media.DeviceName = attributes[2]
			media.DeviceId = attributes[2]
			motionPixels, err := strconv.Atoi(attributes[4])
			if err != nil {
				return media, fmt.Errorf("invalid motion pixels format: %s", attributes[4])
			}
			media.Metadata = &MediaMetadata{
				MotionPixels: motionPixels,
			}
			duration, err := strconv.ParseInt(attributes[5], 10, 64)
			if err != nil {
				return media, fmt.Errorf("invalid duration format: %s", attributes[5])
			}
			media.Duration = int(duration)
		} else {
			return media, fmt.Errorf("invalid attributes format: %s", fileName) // Return empty media if attributes format is invalid
		}

		// If DeviceId is set in metadata, we expect the new format, and can extract more data from the event object.
	} else {
		media.VideoFile = pipelineEvent.Payload.FileName
		startTimestamp, err := strconv.ParseInt(pipelineEvent.Payload.Metadata.Timestamp, 10, 64)
		if err != nil {
			return media, fmt.Errorf("invalid timestamp format: %s", pipelineEvent.Payload.Metadata.Timestamp)
		}
		media.StartTimestamp = startTimestamp
		media.DeviceName = pipelineEvent.Payload.Metadata.DeviceName
		duration, err := strconv.ParseInt(pipelineEvent.Payload.Metadata.Duration, 10, 64)
		if err != nil {
			return media, fmt.Errorf("invalid duration format: %s", pipelineEvent.Payload.Metadata.Duration)
		}
		media.Duration = int(duration)
		media.DeviceId = pipelineEvent.Payload.Metadata.DeviceId
	}

	media.StorageSolution = pipelineEvent.Storage
	return media, nil
}
