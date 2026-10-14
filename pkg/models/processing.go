package models

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const ProcessingVersion = 1

// ProcessingRegistration announces the delegated operations expected for a recording.
type ProcessingRegistration struct {
	Version    int      `json:"version" bson:"version"`
	Revision   int      `json:"revision" bson:"revision"`
	Operations []string `json:"operations" bson:"operations"`
}

// ProcessingObject identifies a source recording or generated artifact.
type ProcessingObject struct {
	MediaId  primitive.ObjectID `json:"mediaId" bson:"media_id"`
	Provider string             `json:"provider" bson:"provider"`
	FileName string             `json:"filename" bson:"filename"`
}

// ProcessingJob delegates one revision of one operation without storage credentials.
type ProcessingJob struct {
	Version       int                `json:"version" bson:"version"`
	JobId         primitive.ObjectID `json:"jobId" bson:"job_id"`
	Revision      int                `json:"revision" bson:"revision"`
	Operation     string             `json:"operation" bson:"operation"`
	SourceVaultId string             `json:"sourceVaultId,omitempty" bson:"source_vault_id,omitempty"`
	Source        ProcessingObject   `json:"source" bson:"source"`
	Artifact      ProcessingObject   `json:"artifact" bson:"artifact"`
	Device        string             `json:"device" bson:"device"`
	TraceId       string             `json:"traceId,omitempty" bson:"trace_id,omitempty"`
}

// ProcessingCompletionTarget gives a worker scoped Vault endpoints and authorization.
type ProcessingCompletionTarget struct {
	JobURL              string `json:"jobUrl" bson:"job_url"`
	ArtifactURL         string `json:"artifactUrl" bson:"artifact_url"`
	CompletionURL       string `json:"completionUrl" bson:"completion_url"`
	Token               string `json:"token" bson:"token"`
	CallbackIntegration string `json:"callbackIntegration,omitempty" bson:"callback_integration,omitempty"`
	ExpiresAt           int64  `json:"expiresAt" bson:"expires_at"`
}

// ProcessingArtifactMetadata describes a generated JPEG artifact.
type ProcessingArtifactMetadata struct {
	Width      int `json:"width" bson:"width"`
	Height     int `json:"height" bson:"height"`
	TileWidth  int `json:"tileWidth,omitempty" bson:"tile_width,omitempty"`
	TileHeight int `json:"tileHeight,omitempty" bson:"tile_height,omitempty"`
	Columns    int `json:"columns,omitempty" bson:"columns,omitempty"`
	Rows       int `json:"rows,omitempty" bson:"rows,omitempty"`
	Count      int `json:"count,omitempty" bson:"count,omitempty"`
	Interval   int `json:"interval,omitempty" bson:"interval,omitempty"`
}

// ProcessingCompletion reports the terminal outcome of a delegated job.
type ProcessingCompletion struct {
	Version    int                        `json:"version" bson:"version"`
	Revision   int                        `json:"revision" bson:"revision"`
	Status     string                     `json:"status" bson:"status"`
	ArtifactId primitive.ObjectID         `json:"artifactId,omitempty" bson:"artifact_id,omitempty"`
	Metadata   ProcessingArtifactMetadata `json:"metadata" bson:"metadata"`
	Error      string                     `json:"error,omitempty" bson:"error,omitempty"`
}

// ProcessingResult is the typed body carried by PipelinePayload.Result.
type ProcessingResult struct {
	Version       int                        `json:"version" bson:"version"`
	JobId         primitive.ObjectID         `json:"jobId" bson:"job_id"`
	Revision      int                        `json:"revision" bson:"revision"`
	SourceMediaId primitive.ObjectID         `json:"sourceMediaId" bson:"source_media_id"`
	Artifact      ProcessingObject           `json:"artifact" bson:"artifact"`
	FileSize      int64                      `json:"fileSize" bson:"file_size"`
	Metadata      ProcessingArtifactMetadata `json:"metadata" bson:"metadata"`
}

// Validate verifies a terminal completion against its delegated job.
func (completion ProcessingCompletion) Validate(job ProcessingJob) error {
	if completion.Version != ProcessingVersion || completion.Revision != job.Revision {
		return fmt.Errorf("unsupported version or stale processing revision")
	}
	switch completion.Status {
	case "failed":
		if completion.Error == "" || len(completion.Error) > 1024 || !completion.ArtifactId.IsZero() || completion.Metadata != (ProcessingArtifactMetadata{}) {
			return fmt.Errorf("failed completion requires an error of at most 1024 bytes and no artifact")
		}
	case "completed":
		if completion.ArtifactId != job.Artifact.MediaId || completion.ArtifactId.IsZero() || completion.Error != "" {
			return fmt.Errorf("completed operation requires its captured artifact id and no error")
		}
		metadata := completion.Metadata
		if metadata.Width <= 0 || metadata.Height <= 0 || metadata.Width > 32768 || metadata.Height > 32768 {
			return fmt.Errorf("invalid artifact dimensions")
		}
		if job.Operation == "thumbnail" {
			if metadata.TileWidth != 0 || metadata.TileHeight != 0 || metadata.Columns != 0 || metadata.Rows != 0 || metadata.Count != 0 || metadata.Interval != 0 {
				return fmt.Errorf("thumbnail metadata must not contain sprite tiles")
			}
		} else if job.Operation == "sprite" {
			if metadata.Columns <= 0 || metadata.Rows <= 0 || metadata.TileWidth <= 0 || metadata.TileHeight <= 0 ||
				metadata.Columns > 32768 || metadata.Rows > 32768 || metadata.TileWidth > 32768 || metadata.TileHeight > 32768 ||
				metadata.Columns*metadata.TileWidth != metadata.Width || metadata.Rows*metadata.TileHeight != metadata.Height ||
				metadata.Count <= 0 || metadata.Count > metadata.Columns*metadata.Rows || metadata.Interval <= 0 {
				return fmt.Errorf("sprite tiles must match the artifact dimensions")
			}
		} else {
			return fmt.Errorf("unsupported processing operation")
		}
	default:
		return fmt.Errorf("processing completion status must be completed or failed")
	}
	return nil
}
