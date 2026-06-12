package ingest

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/uug-ai/models/pkg/models"
)

// --- fakes -----------------------------------------------------------------

type fakeANPRStore struct {
	runs []models.ANPRRun
	err  error
}

func (f *fakeANPRStore) UpsertANPRRun(_ context.Context, run models.ANPRRun) error {
	if f.err != nil {
		return f.err
	}
	f.runs = append(f.runs, run)
	return nil
}

type fakeMarkerStore struct {
	markers []models.Marker
	err     error
}

func (f *fakeMarkerStore) UpsertMarkers(_ context.Context, markers []models.Marker) error {
	if f.err != nil {
		return f.err
	}
	f.markers = append(f.markers, markers...)
	return nil
}

func newANPRScope(src Source) (Scope, *fakeANPRStore) {
	store := &fakeANPRStore{}
	return Scope{Source: src, ANPR: store}, store
}

func anprPayload(t *testing.T, body map[string]any) json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	return raw
}

// --- tests -----------------------------------------------------------------

func TestIngestANPR_PixelNormalizesAndStores(t *testing.T) {
	scope, store := newANPRScope(SourcePipeline)

	payload := anprPayload(t, map[string]any{
		"schemaVersion":   "1.0",
		"source":          map[string]any{"kind": "model", "name": "acme-anpr", "version": "2", "runId": "RUN-A"},
		"coordinateSpace": "pixel",
		"media":           map[string]any{"width": 1000, "height": 500},
		"plates": []any{
			map[string]any{
				"plate": "1ABC234", "display": "1-ABC-234", "confidence": 0.92,
				"country": "BE", "frame": 12, "timestampMs": 480,
				"x1": 100, "y1": 50, "x2": 300, "y2": 250,
				"candidates": []any{
					map[string]any{"plate": "1ABC234", "confidence": 0.92},
					map[string]any{"plate": "1A8C234", "confidence": 0.41},
				},
			},
		},
	})

	report, err := Ingest(context.Background(), scope, target(), "anpr", payload)
	if err != nil {
		t.Fatalf("Ingest: %v", err)
	}
	detail := report.Detail.(ANPRDetail)
	if detail.PlatesStored != 1 {
		t.Fatalf("want 1 plate stored, got %d", detail.PlatesStored)
	}
	if report.RunId != "RUN-A" {
		t.Fatalf("want RunId RUN-A, got %q", report.RunId)
	}
	if len(store.runs) != 1 {
		t.Fatalf("want 1 stored run, got %d", len(store.runs))
	}

	run := store.runs[0]
	if run.Key != "cam-1_1700000000_rec" || run.OrganisationId != "org-9" || run.DeviceId != "dev-3" {
		t.Fatalf("target identity not stamped: %+v", run)
	}
	if run.RecordingTimestamp != 1700000000 {
		t.Fatalf("want recordingTimestamp stamped, got %d", run.RecordingTimestamp)
	}
	if run.OriginalCoordinateSpace != "pixel" {
		t.Fatalf("want originalCoordinateSpace pixel, got %q", run.OriginalCoordinateSpace)
	}

	p := run.Plates[0]
	if p.Plate != "1ABC234" || p.Display != "1-ABC-234" || p.Country != "BE" {
		t.Fatalf("plate fields not preserved: %+v", p)
	}
	if p.Box == nil {
		t.Fatalf("want a normalised box, got nil")
	}
	if !approx(p.Box.X1, 0.1) || !approx(p.Box.Y1, 0.1) || !approx(p.Box.X2, 0.3) || !approx(p.Box.Y2, 0.5) {
		t.Fatalf("box not normalised to [0,1]: %+v", p.Box)
	}
	if len(p.Candidates) != 2 || p.Candidates[1].Plate != "1A8C234" {
		t.Fatalf("candidates not preserved: %+v", p.Candidates)
	}
}

func TestIngestANPR_CreatesMarkerPerPlate(t *testing.T) {
	anprStore := &fakeANPRStore{}
	markerStore := &fakeMarkerStore{}
	scope := Scope{Source: SourcePipeline, ANPR: anprStore, Markers: markerStore}

	payload := anprPayload(t, map[string]any{
		"source": map[string]any{"kind": "model", "name": "a", "version": "1", "runId": "RUN-M"},
		"plates": []any{
			map[string]any{"plate": "1ABC234", "display": "1-ABC-234", "country": "BE", "confidence": 0.9, "timestampMs": 2000},
			map[string]any{"plate": "ZZ999"},
		},
	})

	if _, err := Ingest(context.Background(), scope, target(), "anpr", payload); err != nil {
		t.Fatalf("Ingest: %v", err)
	}
	if len(markerStore.markers) != 2 {
		t.Fatalf("want one marker per plate (2), got %d", len(markerStore.markers))
	}

	m := markerStore.markers[0]
	if m.Name != "1-ABC-234" {
		t.Errorf("marker name = %q, want the display plate \"1-ABC-234\"", m.Name)
	}
	// target() stamps recordingTimestamp 1700000000; +2000ms => +2s.
	if m.StartTimestamp != 1700000002 || m.EndTimestamp != 1700000002 {
		t.Errorf("marker timing = [%d,%d], want point-in-time 1700000002", m.StartTimestamp, m.EndTimestamp)
	}
	if m.DeviceId != "dev-3" || m.OrganisationId != "org-9" {
		t.Errorf("marker RBAC not stamped: device=%q org=%q", m.DeviceId, m.OrganisationId)
	}
	if len(m.Tags) != 1 || m.Tags[0].Name != "anpr" {
		t.Errorf("marker tags = %+v, want a single \"anpr\" tag", m.Tags)
	}
	// The second plate has no display, so the name falls back to the plate text.
	if markerStore.markers[1].Name != "ZZ999" {
		t.Errorf("marker[1] name = %q, want fallback \"ZZ999\"", markerStore.markers[1].Name)
	}
}

func TestIngestANPR_APISourceSkipsMarkers(t *testing.T) {
	anprStore := &fakeANPRStore{}
	markerStore := &fakeMarkerStore{}
	scope := Scope{Source: SourceAPI, ANPR: anprStore, Markers: markerStore}

	payload := anprPayload(t, map[string]any{
		"source": map[string]any{"kind": "model", "name": "a", "version": "1", "runId": "RUN-N"},
		"plates": []any{map[string]any{"plate": "1ABC234"}},
	})

	if _, err := Ingest(context.Background(), scope, target(), "anpr", payload); err != nil {
		t.Fatalf("Ingest: %v", err)
	}
	if len(anprStore.runs) != 1 {
		t.Fatalf("want the run stored for an API push, got %d", len(anprStore.runs))
	}
	if len(markerStore.markers) != 0 {
		t.Fatalf("want no markers for an API push (trusted-pipeline-only), got %d", len(markerStore.markers))
	}
}

func TestIngestANPR_XYWHForm(t *testing.T) {
	scope, store := newANPRScope(SourcePipeline)
	payload := anprPayload(t, map[string]any{
		"source":          map[string]any{"kind": "model", "name": "a", "version": "1", "runId": "RUN-B"},
		"coordinateSpace": "pixel",
		"media":           map[string]any{"width": 200, "height": 100},
		"plates": []any{
			map[string]any{"plate": "XY99", "x": 20, "y": 10, "w": 40, "h": 20},
		},
	})
	if _, err := Ingest(context.Background(), scope, target(), "anpr", payload); err != nil {
		t.Fatalf("Ingest: %v", err)
	}
	b := store.runs[0].Plates[0].Box
	if b == nil || !approx(b.X1, 0.1) || !approx(b.Y1, 0.1) || !approx(b.X2, 0.3) || !approx(b.Y2, 0.3) {
		t.Fatalf("xywh box not normalised: %+v", b)
	}
}

func TestIngestANPR_NormalizedPassthrough(t *testing.T) {
	scope, store := newANPRScope(SourcePipeline)
	payload := anprPayload(t, map[string]any{
		"source":          map[string]any{"kind": "model", "name": "a", "version": "1", "runId": "RUN-C"},
		"coordinateSpace": "normalized",
		"plates": []any{
			map[string]any{"plate": "ABC", "x1": 0.2, "y1": 0.2, "x2": 0.4, "y2": 0.4},
		},
	})
	if _, err := Ingest(context.Background(), scope, target(), "anpr", payload); err != nil {
		t.Fatalf("Ingest: %v", err)
	}
	b := store.runs[0].Plates[0].Box
	if b == nil || !approx(b.X1, 0.2) || !approx(b.X2, 0.4) {
		t.Fatalf("normalized box not passed through: %+v", b)
	}
}

func TestIngestANPR_NoBoxIsValid(t *testing.T) {
	scope, store := newANPRScope(SourcePipeline)
	payload := anprPayload(t, map[string]any{
		"source": map[string]any{"kind": "import", "name": "a", "version": "1", "runId": "RUN-D"},
		"plates": []any{
			map[string]any{"plate": "NOBOX1", "confidence": 0.5},
		},
	})
	report, err := Ingest(context.Background(), scope, target(), "anpr", payload)
	if err != nil {
		t.Fatalf("Ingest: %v", err)
	}
	if d := report.Detail.(ANPRDetail); d.PlatesStored != 1 {
		t.Fatalf("want 1 plate stored, got %d", d.PlatesStored)
	}
	run := store.runs[0]
	if run.Plates[0].Box != nil {
		t.Fatalf("want nil box, got %+v", run.Plates[0].Box)
	}
	if run.OriginalCoordinateSpace != "" {
		t.Fatalf("want empty originalCoordinateSpace when no box, got %q", run.OriginalCoordinateSpace)
	}
}

func TestIngestANPR_EmptyPlatesStillStoresRun(t *testing.T) {
	scope, store := newANPRScope(SourcePipeline)
	payload := anprPayload(t, map[string]any{
		"source": map[string]any{"kind": "pipeline", "name": "a", "version": "1", "runId": "RUN-E"},
		"plates": []any{},
	})
	report, err := Ingest(context.Background(), scope, target(), "anpr", payload)
	if err != nil {
		t.Fatalf("Ingest: empty plates must not error, got %v", err)
	}
	if d := report.Detail.(ANPRDetail); d.PlatesStored != 0 {
		t.Fatalf("want 0 plates stored, got %d", d.PlatesStored)
	}
	if len(store.runs) != 1 {
		t.Fatalf("want the empty run still stored for provenance, got %d runs", len(store.runs))
	}
}

func TestIngestANPR_RejectsMissingRunId(t *testing.T) {
	scope, _ := newANPRScope(SourcePipeline)
	payload := anprPayload(t, map[string]any{
		"source": map[string]any{"kind": "model", "name": "a", "version": "1"},
		"plates": []any{map[string]any{"plate": "ABC"}},
	})
	_, err := Ingest(context.Background(), scope, target(), "anpr", payload)
	if !errors.Is(err, ErrANPRValidation) {
		t.Fatalf("want ErrANPRValidation, got %v", err)
	}
}

func TestIngestANPR_DropsBadPlates(t *testing.T) {
	scope, store := newANPRScope(SourcePipeline)
	payload := anprPayload(t, map[string]any{
		"source":          map[string]any{"kind": "model", "name": "a", "version": "1", "runId": "RUN-F"},
		"coordinateSpace": "normalized",
		"plates": []any{
			map[string]any{"plate": ""}, // missing text -> dropped
			map[string]any{"plate": "OOF", "x1": 0.5, "y1": 0.5, "x2": 2.0, "y2": 2.0}, // out of frame -> dropped
			map[string]any{"plate": "GOOD", "x1": 0.1, "y1": 0.1, "x2": 0.2, "y2": 0.2},
		},
	})
	report, err := Ingest(context.Background(), scope, target(), "anpr", payload)
	if err != nil {
		t.Fatalf("Ingest: %v", err)
	}
	detail := report.Detail.(ANPRDetail)
	if detail.PlatesStored != 1 {
		t.Fatalf("want 1 surviving plate, got %d", detail.PlatesStored)
	}
	if len(detail.Rejected) != 2 {
		t.Fatalf("want 2 rejected plates, got %+v", detail.Rejected)
	}
	if store.runs[0].Plates[0].Plate != "GOOD" {
		t.Fatalf("want only GOOD stored, got %+v", store.runs[0].Plates)
	}
}

func TestIngestANPR_ClampsConfidence(t *testing.T) {
	scope, store := newANPRScope(SourcePipeline)
	payload := anprPayload(t, map[string]any{
		"source": map[string]any{"kind": "model", "name": "a", "version": "1", "runId": "RUN-G"},
		"plates": []any{
			map[string]any{"plate": "HI", "confidence": 1.7},
			map[string]any{"plate": "LO", "confidence": -0.3},
		},
	})
	report, err := Ingest(context.Background(), scope, target(), "anpr", payload)
	if err != nil {
		t.Fatalf("Ingest: %v", err)
	}
	if !approx(store.runs[0].Plates[0].Confidence, 1.0) || !approx(store.runs[0].Plates[1].Confidence, 0.0) {
		t.Fatalf("confidence not clamped: %+v", store.runs[0].Plates)
	}
	foundWarning := false
	for _, w := range report.Warnings {
		if w == "CONFIDENCE_CLAMPED: 2 plate(s)" {
			foundWarning = true
		}
	}
	if !foundWarning {
		t.Fatalf("want CONFIDENCE_CLAMPED warning, got %v", report.Warnings)
	}
}

func TestIngestANPR_PixelBoxRequiresMedia(t *testing.T) {
	scope, _ := newANPRScope(SourcePipeline)
	payload := anprPayload(t, map[string]any{
		"source":          map[string]any{"kind": "model", "name": "a", "version": "1", "runId": "RUN-H"},
		"coordinateSpace": "pixel",
		"plates": []any{
			map[string]any{"plate": "ABC", "x1": 10, "y1": 10, "x2": 20, "y2": 20},
		},
	})
	_, err := Ingest(context.Background(), scope, target(), "anpr", payload)
	if !errors.Is(err, ErrANPRValidation) {
		t.Fatalf("want ErrANPRValidation for pixel box without media, got %v", err)
	}
}

func TestIngestANPR_NoStoreConfigured(t *testing.T) {
	payload := anprPayload(t, map[string]any{
		"source": map[string]any{"kind": "model", "name": "a", "version": "1", "runId": "RUN-I"},
		"plates": []any{map[string]any{"plate": "ABC"}},
	})
	scope := Scope{Source: SourcePipeline} // no ANPR sink
	_, err := Ingest(context.Background(), scope, target(), "anpr", payload)
	if err == nil {
		t.Fatalf("want error when no ANPRStore configured")
	}
}

func TestIngestANPR_UnknownKindUnaffected(t *testing.T) {
	scope, _ := newANPRScope(SourcePipeline)
	_, err := Ingest(context.Background(), scope, target(), "nope", json.RawMessage(`{}`))
	if !errors.Is(err, ErrUnknownKind) {
		t.Fatalf("want ErrUnknownKind, got %v", err)
	}
}
