package models_test

import (
	"bytes"
	"encoding/json"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/uug-ai/models/pkg/api"
	"github.com/uug-ai/models/pkg/models"
)

// anprStageResult builds results.anpr the way the engine does from hub-anpr's
// ingest envelope: one detection block (api.PostDetectionsRequest) and one
// marker block (models.Marker) per vehicle, grouped by type with an "s"
// appended. Values follow hub-anpr: a read plate and an unread one.
func anprStageResult(t *testing.T) map[string]any {
	t.Helper()
	confidence := 0.87
	detection := api.PostDetectionsRequest{
		Task:            "detection",
		Source:          models.DetectionSource{Kind: "pipeline", Name: "hub-anpr", Version: "0.1.0", RunId: "anpr-run"},
		CoordinateSpace: "pixel",
		Media:           models.DetectionMedia{Width: 1920, Height: 1080},
		Categories:      []models.DetectionCategory{{Id: 0, Name: "license_plate", Alias: "anpr"}},
		Tracks: []api.DetectionTrackInput{
			{Id: "3", Label: "AB123CD", Confidence: 0.9, Shape: "rect",
				Meta:  map[string]interface{}{"plate": "AB123CD", "classifiedAs": "car", "recognisedBy": "hub-anpr", "recognisedOcr": true, "unread": false},
				Boxes: []api.DetectionBoxInput{{Frame: 12, Confidence: 0.9, Label: "AB123CD"}}},
			{Id: "4", Label: "unread", Confidence: 0.9, Shape: "rect",
				Meta:  map[string]interface{}{"plate": "", "classifiedAs": "car", "recognisedBy": "hub-anpr", "recognisedOcr": false, "unread": true},
				Boxes: []api.DetectionBoxInput{{Frame: 20, Confidence: 0.9, Label: "unread"}}},
		},
	}
	read := models.Marker{
		StartTimestamp: 1752482068, EndTimestamp: 1752482071, Duration: 3, Name: "AB123CD",
		Description: "Read plate AB123CD on car",
		Categories:  []models.MarkerCategory{{Name: "anpr"}, {Name: "object"}},
		Tags:        []models.MarkerTag{{Name: "anpr"}, {Name: "AB123CD"}, {Name: "car"}},
		Metadata:    &models.MarkerMetadata{Source: "anpr", Engine: "fast", Confidence: &confidence, BoundingBox: &models.MarkerBox{X: 0.1, Y: 0.2, Width: 0.05, Height: 0.02}},
		Detections:  []models.DetectionRef{{RunId: "anpr-run", TrackId: "3"}},
		Events:      []models.MarkerEvent{{StartTimestamp: 1752482068, EndTimestamp: 1752482071, Duration: 3, Name: "ANPR"}},
	}
	unread := models.Marker{
		StartTimestamp: 1752482072, EndTimestamp: 1752482074, Duration: 2, Name: "anpr unread",
		Description: "Detected unread plate on car",
		Metadata:    &models.MarkerMetadata{Source: "anpr"},
		Detections:  []models.DetectionRef{{RunId: "anpr-run", TrackId: "4"}},
	}
	grouped := map[string]any{}
	for _, block := range []struct {
		kind string
		data any
	}{{"detection", detection}, {"marker", read}, {"marker", unread}, {"media-patch", map[string]string{"mediaKey": "m", "description": "Number plate(s) detected: AB123CD"}}} {
		raw, err := json.Marshal(block.data)
		if err != nil {
			t.Fatal(err)
		}
		var decoded any
		if err := json.Unmarshal(raw, &decoded); err != nil {
			t.Fatal(err)
		}
		key := block.kind + "s"
		list, _ := grouped[key].([]any)
		grouped[key] = append(list, decoded)
	}
	return grouped
}

// The real wire types decode strictly into WorkflowAnprResult.
func TestAnprResultMatchesWireTypes(t *testing.T) {
	raw, _ := json.Marshal(anprStageResult(t))
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var result models.WorkflowAnprResult
	if err := decoder.Decode(&result); err != nil {
		t.Fatalf("results.anpr does not match WorkflowAnprResult: %v", err)
	}
	if len(result.Detections) != 1 || len(result.Detections[0].Tracks) != 2 || result.Detections[0].Tracks[0].Meta.Plate != "AB123CD" ||
		!result.Detections[0].Tracks[1].Meta.Unread || len(result.Markers) != 2 || *result.Markers[0].Metadata.Confidence != 0.87 {
		t.Fatalf("decoded result = %+v", result)
	}
}

func jsonKeys(kind reflect.Type) []string {
	var keys []string
	for i := 0; i < kind.NumField(); i++ {
		if name := strings.Split(kind.Field(i).Tag.Get("json"), ",")[0]; name != "" && name != "-" {
			keys = append(keys, name)
		}
	}
	sort.Strings(keys)
	return keys
}

// Every field of the wire types is described (or explicitly excluded with
// condition:"-"), so a field added to them fails here until the type follows.
func TestAnprResultCoversWireFields(t *testing.T) {
	for _, pair := range []struct{ ours, wire any }{
		{models.WorkflowAnprDetection{}, api.PostDetectionsRequest{}},
		{models.WorkflowAnprTrack{}, api.DetectionTrackInput{}},
		{models.WorkflowAnprMarker{}, models.Marker{}},
		{models.WorkflowAnprMarkerMetadata{}, models.MarkerMetadata{}},
	} {
		ours, wire := jsonKeys(reflect.TypeOf(pair.ours)), jsonKeys(reflect.TypeOf(pair.wire))
		if !reflect.DeepEqual(ours, wire) {
			t.Errorf("%T keys %v differ from %T keys %v", pair.ours, ours, pair.wire, wire)
		}
	}
}

func resolvedType(value any) string {
	switch number := value.(type) {
	case string:
		return "string"
	case bool:
		return "boolean"
	case float64:
		if number == float64(int64(number)) {
			return "integer"
		}
		return "number"
	}
	return ""
}

// Every results.anpr schema path resolves with its declared type in a run.
func TestAnprSchemaPathsResolve(t *testing.T) {
	root := models.WorkflowRun{Results: map[string]any{models.WorkflowAnprOperation: anprStageResult(t)}}.ConditionRoot(models.WorkflowConditionRootRun)
	var paths []string
	for _, path := range models.WorkflowConditionRootSchema() {
		if !strings.HasPrefix(path.Path, "results.anpr.") {
			continue
		}
		paths = append(paths, path.Path)
		for _, mode := range path.Modes {
			if mode != models.WorkflowConditionRootRun {
				t.Errorf("%s must exist only in run mode", path.Path)
			}
		}
		candidates, found := models.ResolveCandidates(root, path.Path)
		if !found || len(candidates) == 0 {
			t.Errorf("%s does not resolve", path.Path)
			continue
		}
		for _, candidate := range candidates {
			if got := resolvedType(candidate); got != path.Type && !(path.Type == "number" && got == "integer") {
				t.Errorf("%s resolves to %T, schema says %s", path.Path, candidate, path.Type)
			}
		}
	}
	sort.Strings(paths)
	want := []string{
		"results.anpr.detections.*.tracks.*.meta.classifiedAs",
		"results.anpr.detections.*.tracks.*.meta.plate",
		"results.anpr.detections.*.tracks.*.meta.unread",
		"results.anpr.markers.*.duration",
		"results.anpr.markers.*.metadata.confidence",
		"results.anpr.markers.*.metadata.engine",
		"results.anpr.markers.*.name",
	}
	if !reflect.DeepEqual(paths, want) {
		t.Fatalf("anpr schema paths = %v, want %v", paths, want)
	}
}
