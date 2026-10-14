package models

import (
	"encoding/json"
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestProcessingPipelineEventRoundTrip(t *testing.T) {
	mediaId := primitive.NewObjectID()
	jobId := primitive.NewObjectID()
	artifactId := primitive.NewObjectID()
	result, err := json.Marshal(ProcessingResult{
		Version: ProcessingVersion, JobId: jobId, Revision: 1, SourceMediaId: mediaId,
		Artifact: ProcessingObject{MediaId: artifactId, Provider: "archive", FileName: "account/thumbnail.jpg"},
		FileSize: 123, Metadata: ProcessingArtifactMetadata{Width: 250, Height: 250},
	})
	if err != nil {
		t.Fatal(err)
	}
	event := PipelineEvent{
		Request: "process", Operation: "event", Stages: []string{"thumbnail"},
		Payload:         PipelinePayload{FileName: "account/recording.mp4", Result: result},
		VaultProcessing: &ProcessingRegistration{Version: ProcessingVersion, Revision: 1, Operations: []string{"thumbnail"}},
		VaultJob: &ProcessingJob{
			Version: ProcessingVersion, JobId: jobId, Revision: 1, Operation: "thumbnail",
			Source:   ProcessingObject{MediaId: mediaId, Provider: "archive", FileName: "account/recording.mp4"},
			Artifact: ProcessingObject{MediaId: artifactId, Provider: "archive", FileName: "account/thumbnail.jpg"},
			Device:   "device-1",
		},
		Completion: &ProcessingCompletionTarget{
			JobURL: "https://vault.example/job", ArtifactURL: "https://vault.example/artifact",
			CompletionURL: "https://vault.example/complete", Token: "capability", CallbackIntegration: "hub", ExpiresAt: 1_800_000_000,
		},
	}
	encodedJSON, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	var wire map[string]json.RawMessage
	if err := json.Unmarshal(encodedJSON, &wire); err != nil {
		t.Fatal(err)
	}
	if wire["vaultJob"] == nil || wire["completion"] == nil || wire["vaultProcessing"] == nil || wire["vaultResult"] != nil {
		t.Fatalf("incorrect processing envelope: %s", encodedJSON)
	}
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(wire["payload"], &payload); err != nil || payload["result"] == nil {
		t.Fatalf("processing result is not carried by payload.result: %s", encodedJSON)
	}

	for _, codec := range []struct {
		marshal   func(interface{}) ([]byte, error)
		unmarshal func([]byte, interface{}) error
	}{{json.Marshal, json.Unmarshal}, {bson.Marshal, bson.Unmarshal}} {
		encoded, err := codec.marshal(event)
		if err != nil {
			t.Fatal(err)
		}
		var restored PipelineEvent
		if err := codec.unmarshal(encoded, &restored); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(event, restored) {
			t.Fatalf("processing event did not round-trip: %+v", restored)
		}
	}
}

func TestProcessingCompletionValidation(t *testing.T) {
	job := ProcessingJob{Revision: 1, Operation: "thumbnail", Artifact: ProcessingObject{MediaId: primitive.NewObjectID()}}
	completion := ProcessingCompletion{Version: ProcessingVersion, Revision: 1, Status: "completed", ArtifactId: job.Artifact.MediaId,
		Metadata: ProcessingArtifactMetadata{Width: 250, Height: 250}}
	if err := completion.Validate(job); err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*ProcessingCompletion){
		func(value *ProcessingCompletion) { value.Version = 2 },
		func(value *ProcessingCompletion) { value.Revision = 2 },
		func(value *ProcessingCompletion) { value.ArtifactId = primitive.NewObjectID() },
		func(value *ProcessingCompletion) { value.Metadata.Width = 0 },
		func(value *ProcessingCompletion) { value.Metadata.Columns = 1 },
		func(value *ProcessingCompletion) { value.Error = "not a success" },
	} {
		changed := completion
		change(&changed)
		if err := changed.Validate(job); err == nil {
			t.Fatalf("invalid completion accepted: %+v", changed)
		}
	}

	job.Operation = "sprite"
	completion.Metadata = ProcessingArtifactMetadata{
		Width: 500, Height: 250, Columns: 2, Rows: 1, TileWidth: 250, TileHeight: 250, Count: 2, Interval: 5,
	}
	if err := completion.Validate(job); err != nil {
		t.Fatal(err)
	}
	completion.Metadata.Columns = 3
	if err := completion.Validate(job); err == nil {
		t.Fatal("sprite geometry mismatch accepted")
	}

	failure := ProcessingCompletion{Version: ProcessingVersion, Revision: 1, Status: "failed", Error: "FFmpeg failed"}
	if err := failure.Validate(job); err != nil {
		t.Fatal(err)
	}
	failure.ArtifactId = job.Artifact.MediaId
	if err := failure.Validate(job); err == nil {
		t.Fatal("failed operation accepted an artifact")
	}
}
