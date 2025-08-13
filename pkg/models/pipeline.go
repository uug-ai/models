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

type Pipeline struct {
	Events             []string        `json:"events,omitempty"`
	Provider           string          `json:"provider,omitempty"`
	SecondaryProviders []string        `json:"secondary_providers,omitempty"`
	Source             string          `json:"source,omitempty"`
	Request            string          `json:"request,omitempty"`
	ReceiveCount       int64           `json:"receivecount,omitempty"`
	Date               int64           `json:"date,omitempty"`
	FileName           string          `json:"fileName,omitempty"`
	Payload            PipelinePayload `json:"payload,omitempty"`

	// If source is AWS, we got some other info.
	Bucket string `json:"bucket,omitempty"`
	Region string `json:"region,omitempty"`

	// For async queue's (e.g. analysis)
	Operation string                 `json:"operation,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

type PipelinePayload struct {
	FileName         string             `json:"key,omitempty"`
	FileSize         int64              `json:"fileSize,omitempty"`
	MetaData         PipelineMetadata   `json:"metadata,omitempty"`
	IsFragmented     bool               `json:"is_fragmented" bson:"is_fragmented"`
	BytesRanges      string             `json:"bytes_ranges" bson:"bytes_ranges"`
	BytesRangeOnTime []BytesRangeOnTime `json:"bytes_range_on_time" bson:"bytes_range_on_time"`
}

type PipelineMetadata struct {
	Capture           string `json:"capture,omitempty"`
	Duration          string `json:"duration,omitempty"`
	NumberOfChanges   string `json:"event-numberofchanges,omitempty"`
	UploadTime        string `json:"uploadtime,omitempty"`
	MicroSeconds      string `json:"event-microseconds,omitempty"`
	ProductId         string `json:"productid,omitempty"`
	InstanceName      string `json:"event-instancename,omitempty"`
	RegionCoordinates string `json:"event-regioncoordinates,omitempty"`
	Timestamp         string `json:"event-timestamp,omitempty"`
	PublicKey         string `json:"publickey,omitempty"`
}

type BytesRangeOnTime struct {
	Duration string `json:"duration" bson:"duration"`
	Time     string `json:"time" bson:"time"`
	Range    string `json:"range" bson:"range"`
}
