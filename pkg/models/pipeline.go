package models

// --------------------------------------------------------------------------------------------------------------------------------
// Pipeline represents a data processing pipeline that can handle various stages of data processing.
// The idea of the pipeline is to process data in a series of steps, where the output of one step becomes the input for the next.
// The initial pipeline object is expanded with each stage of processing.
//
// Pipeline stages:
// 1. event
// 2. monitor
// 3. sequence
// 4. analysis
// 5. throttler
// 6. notification

type PipelineEvent struct {
	Request      string   `json:"request,omitempty"`
	CurrentStage string   `json:"operation,omitempty"`
	Stages       []string `json:"events,omitempty"` // Stages of the pipeline, e.g., event, monitor, sequence, analysis, throttler, notification

	SecondaryProviders []string `json:"secondary_providers,omitempty"`

	Storage      string          `json:"provider,omitempty"`
	Provider     string          `json:"source,omitempty"`
	ReceiveCount int64           `json:"receivecount,omitempty"`
	Timestamp    int64           `json:"date,omitempty"`
	FileName     string          `json:"fileName,omitempty"`
	Payload      PipelinePayload `json:"payload,omitempty"`

	Stage map[string]Stage       `json:"stage,omitempty"`
	Data  map[string]interface{} `json:"data,omitempty"` // We should get rid of this and use the stage map
}

type PipelinePayload struct {
	Timestamp int64  `json:"timestamp,omitempty"`
	FileName  string `json:"key,omitempty"`
	FileSize  int64  `json:"fileSize,omitempty"`
	Duration  string `json:"duration,omitempty"`

	// RBAC
	OrganisationId string `json:"organisationId,omitempty"`
	DeviceId       string `json:"deviceId,omitempty"`
	DeviceName     string `json:"deviceName,omitempty"`

	// MP4 fragmentation
	IsFragmented     bool                         `json:"is_fragmented" bson:"is_fragmented"`
	BytesRanges      string                       `json:"bytes_ranges" bson:"bytes_ranges"`
	BytesRangeOnTime []FragmentedBytesRangeOnTime `json:"bytes_range_on_time" bson:"bytes_range_on_time"`

	// Metadata
	MetaData PipelineMetadata `json:"metadata,omitempty"`
}

type PipelineMetadata struct {

	// Duration of the recording
	Duration          string `json:"duration,omitempty"`
	NumberOfChanges   string `json:"event-numberofchanges,omitempty"`
	UploadTime        string `json:"uploadtime,omitempty"`
	MicroSeconds      string `json:"event-microseconds,omitempty"`
	DeviceId          string `json:"productid,omitempty"`
	DeviceName        string `json:"event-instancename,omitempty"`
	RegionCoordinates string `json:"event-regioncoordinates,omitempty"`
	Timestamp         string `json:"event-timestamp,omitempty"`
	PublicKey         string `json:"publickey,omitempty"`
}

// As defined above we have multiple stages, each with its own set of data and processing logic.
// 1. event, 2. monitor, 3. sequence, 4. analysis, 5. throttler, 6. notification
// We'll define a custom struct for each stage's data, however we should be able to use the stage type
// so we can use it interchangeably.

type Stage interface {
	GetName() string
}

type EventStage struct {
	Name      string
	EventData string // Add fields relevant to event stage
	// Add more fields as needed
}

func (e EventStage) GetName() string { return e.Name }

type MonitorStage struct {
	Name        string
	MonitorData string // Add fields relevant to monitor stage
	// Add more fields as needed
}

func (m MonitorStage) GetName() string { return m.Name }

type SequenceStage struct {
	Name       string
	SequenceId int64 // Add fields relevant to sequence stage
	// Add more fields as needed
}

func (s SequenceStage) GetName() string { return s.Name }

type AnalysisStage struct {
	Name           string
	AnalysisResult string // Add fields relevant to analysis stage
	// Add more fields as needed
}

func (a AnalysisStage) GetName() string { return a.Name }

type ThrottlerStage struct {
	Name          string
	ThrottleLimit int // Add fields relevant to throttler stage
	// Add more fields as needed
}

func (t ThrottlerStage) GetName() string { return t.Name }

type NotificationStage struct {
	Name             string
	NotificationType string // Add fields relevant to notification stage
	// Add more fields as needed
}

func (n NotificationStage) GetName() string { return n.Name }
