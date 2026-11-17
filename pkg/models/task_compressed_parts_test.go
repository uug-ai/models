package models

import (
	"encoding/json"
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestCompressedPartURLIsServedButNeverStored(t *testing.T) {
	task := Task{
		CompressedFile:     "owner/cases/1/zips/part-1.zip",
		CompressedProvider: "Archive",
		CompressedParts: []CompressedPart{
			{File: "owner/cases/1/zips/part-1.zip", Provider: "Archive", Size: 10, Files: 2, Url: "https://signed/1"},
			{File: "owner/cases/1/zips/part-2.zip", Provider: "Archive", Size: 20, Files: 3, Url: "https://signed/2"},
		},
	}

	raw, err := bson.Marshal(task)
	if err != nil {
		t.Fatalf("bson.Marshal() error = %v", err)
	}
	var stored Task
	if err := bson.Unmarshal(raw, &stored); err != nil {
		t.Fatalf("bson.Unmarshal() error = %v", err)
	}
	if len(stored.CompressedParts) != 2 {
		t.Fatalf("stored %d parts, want 2", len(stored.CompressedParts))
	}
	for i, part := range stored.CompressedParts {
		if part.Url != "" {
			t.Errorf("part %d: signed URL was stored: %q", i, part.Url)
		}
		if part.File != task.CompressedParts[i].File || part.Size != task.CompressedParts[i].Size || part.Files != task.CompressedParts[i].Files {
			t.Errorf("part %d did not round-trip: %+v", i, part)
		}
	}

	served, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if !strings.Contains(string(served), `"url":"https://signed/2"`) {
		t.Errorf("JSON misses the signed URL: %s", served)
	}
}

func TestTaskWithoutCompressedPartsOmitsTheField(t *testing.T) {
	served, err := json.Marshal(Task{CompressedFile: "owner/cases/1/zips/one.zip"})
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if strings.Contains(string(served), "compressedParts") {
		t.Errorf("compressedParts present on a single-zip task: %s", served)
	}
}
