package ingest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/uug-ai/models/pkg/api"
	"github.com/uug-ai/models/pkg/models"
)

// Validation and normalisation limits for the anpr kind. The geometry limits
// (boundsTolerance, clampUnit, withinFrame) and the envelope caps (maxNameLen,
// maxVersionLen, maxRunIdLen, maxMetaBytes) are shared with the detection kind —
// they are generic, not redaction-specific.
const (
	maxPlatesPerRun       = 1000
	maxPlateLen           = 32
	maxCandidatesPerPlate = 16
)

// Per-plate rejection reasons surfaced in the report warnings.
const (
	reasonPlateMissingText    = "plate_missing_text"
	reasonPlateTooLong        = "plate_text_too_long"
	reasonPlateMissingGeom    = "plate_missing_geometry"
	reasonPlateNonPositiveDim = "plate_non_positive_dimensions"
	reasonPlateMissingScale   = "plate_missing_media_scale"
	reasonPlateBoxOutOfFrame  = "plate_box_out_of_frame"
)

// ErrANPRValidation wraps an ANPR request-envelope validation failure (missing
// runId, oversize plate list, missing media scale for pixel boxes). The HTTP
// adapter maps it to a 400; the wrapped message carries the detail.
var ErrANPRValidation = errors.New("ingest: invalid anpr request")

// --- Sink (implemented by each app's repository; kept as an interface so this
// package stays infra-free) -------------------------------------------------

// ANPRStore is the mandatory sink for an ANPR run. The implementation MUST
// upsert by (run.Key, run.Source.RunId) so a redelivered or re-posted run
// replaces rather than duplicates. It is deliberately separate from
// DetectionStore: ANPR runs live in their own collection with their own shape.
type ANPRStore interface {
	UpsertANPRRun(ctx context.Context, run models.ANPRRun) error
}

// MarkerStore is the optional sink for the markers derived from an ANPR run —
// one user-facing marker per recognised plate so plate reads surface on the
// recording timeline and in marker search. The implementation MUST upsert by a
// stable identity — (OrganisationId, DeviceId, Name, StartTimestamp) — so an
// at-least-once redelivery or a re-analysis of the same recording refreshes the
// marker rather than duplicating it.
type MarkerStore interface {
	UpsertMarkers(ctx context.Context, markers []models.Marker) error
}

// --- Kind handler ----------------------------------------------------------

// anprHandler is the automatic-number-plate-recognition kind. Its sequence is
// the mandatory UpsertANPRRun (store the recognised plates as an ANPRRun in the
// dedicated anpr collection) followed by the trusted-only CreateANPRMarkers,
// which surfaces each plate as a user-facing marker — the anpr analogue of the
// detection kind's "store the run AND adjust region-selection" pair.
var anprHandler = Handler{
	Kind:    "anpr",
	Decode:  decodeANPR,
	Actions: []Action{UpsertANPRRun{}, CreateANPRMarkers{}},
}

// ANPRDetail is the anpr kind's ReportDetail: how many plates were stored and
// which were rejected during normalisation.
type ANPRDetail struct {
	PlatesStored int
	Rejected     []api.ANPRRejection
}

// Summary implements ReportDetail.
func (d ANPRDetail) Summary() string {
	return fmt.Sprintf("%d plate(s), %d rejected", d.PlatesStored, len(d.Rejected))
}

// decodeANPR unmarshals the envelope payload into api.PostANPRRequest, validates
// the envelope, normalises each plate (rejecting individual malformed plates
// rather than the whole run), and stamps the target-derived identity onto the
// run. An empty or fully-rejected plate list is NOT an error: ANPR legitimately
// records that it ran and found nothing, so the run is still stored (idempotent
// provenance). It runs once per request; the action consumes its typed output.
func decodeANPR(scope Scope, target Target, payload json.RawMessage) (any, Report, error) {
	var req api.PostANPRRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		return nil, Report{}, fmt.Errorf("ingest: decode anpr payload: %w", err)
	}

	// --- envelope validation ---
	if req.Source.RunId == "" {
		return nil, Report{}, fmt.Errorf("%w: source.runId is required", ErrANPRValidation)
	}
	if len(req.Source.RunId) > maxRunIdLen {
		return nil, Report{}, fmt.Errorf("%w: source.runId exceeds %d characters", ErrANPRValidation, maxRunIdLen)
	}
	if len(req.Source.Name) > maxNameLen {
		return nil, Report{}, fmt.Errorf("%w: source.name exceeds %d characters", ErrANPRValidation, maxNameLen)
	}
	if len(req.Source.Version) > maxVersionLen {
		return nil, Report{}, fmt.Errorf("%w: source.version exceeds %d characters", ErrANPRValidation, maxVersionLen)
	}
	if len(req.Plates) > maxPlatesPerRun {
		return nil, Report{}, fmt.Errorf("%w: too many plates: %d exceeds limit of %d", ErrANPRValidation, len(req.Plates), maxPlatesPerRun)
	}
	if req.SchemaVersion != "" {
		if _, supported := checkSchemaVersion(req.SchemaVersion); !supported {
			return nil, Report{}, fmt.Errorf("%w: %s", ErrUnsupportedSchema, req.SchemaVersion)
		}
	}

	// A box is only present on some plates; resolve the coordinate space once.
	// Validation of the space is only required when at least one plate carries a
	// box — a "no geometry" result needs no media.
	hasAnyBox := false
	for i := range req.Plates {
		if plateHasBox(req.Plates[i]) {
			hasAnyBox = true
			break
		}
	}
	normalized := req.CoordinateSpace != "pixel"
	scaleX, scaleY := 1.0, 1.0
	if hasAnyBox {
		if req.CoordinateSpace != "pixel" && req.CoordinateSpace != "normalized" {
			return nil, Report{}, fmt.Errorf("%w: coordinateSpace must be \"pixel\" or \"normalized\" when a plate carries a box, got %q", ErrANPRValidation, req.CoordinateSpace)
		}
		if !normalized {
			if req.Media.Width <= 0 || req.Media.Height <= 0 {
				return nil, Report{}, fmt.Errorf("%w: media.width and media.height are required when coordinateSpace is \"pixel\"", ErrANPRValidation)
			}
			scaleX = float64(req.Media.Width)
			scaleY = float64(req.Media.Height)
		}
	}

	// --- per-plate normalisation (individual rejects are non-fatal) ---
	detail := ANPRDetail{Rejected: []api.ANPRRejection{}}
	warnings := []string{}
	storedPlates := make([]models.ANPRPlate, 0, len(req.Plates))
	confidenceClamped := 0
	candidatesTruncated := 0
	sawBox := false

	for i := range req.Plates {
		in := req.Plates[i]

		text := strings.TrimSpace(in.Plate)
		if text == "" {
			detail.Rejected = append(detail.Rejected, api.ANPRRejection{Plate: in.Plate, Frame: in.Frame, Reason: reasonPlateMissingText})
			continue
		}
		if len(text) > maxPlateLen {
			detail.Rejected = append(detail.Rejected, api.ANPRRejection{Plate: text, Frame: in.Frame, Reason: reasonPlateTooLong})
			continue
		}
		if msg := validateMetaSize(fmt.Sprintf("plate %q meta", text), in.Meta); msg != "" {
			return nil, Report{}, fmt.Errorf("%w: %s", ErrANPRValidation, msg)
		}

		var box *models.ANPRBox
		if plateHasBox(in) {
			b, reason := normalizeANPRBox(in, normalized, scaleX, scaleY)
			if reason != "" {
				detail.Rejected = append(detail.Rejected, api.ANPRRejection{Plate: text, Frame: in.Frame, Reason: reason})
				continue
			}
			box = b
			sawBox = true
		}

		conf, clamped := clampConfidence(in.Confidence)
		if clamped {
			confidenceClamped++
		}

		candidates := in.Candidates
		if len(candidates) > maxCandidatesPerPlate {
			candidatesTruncated++
			candidates = candidates[:maxCandidatesPerPlate]
		}

		storedPlates = append(storedPlates, models.ANPRPlate{
			Plate:       text,
			Display:     in.Display,
			Confidence:  conf,
			Country:     in.Country,
			Region:      in.Region,
			Frame:       in.Frame,
			TimestampMs: in.TimestampMs,
			Box:         box,
			Candidates:  convertCandidates(candidates),
			Meta:        in.Meta,
		})
		detail.PlatesStored++
	}

	if confidenceClamped > 0 {
		warnings = append(warnings, fmt.Sprintf("CONFIDENCE_CLAMPED: %d plate(s)", confidenceClamped))
	}
	if candidatesTruncated > 0 {
		warnings = append(warnings, fmt.Sprintf("CANDIDATES_TRUNCATED: %d plate(s) over %d", candidatesTruncated, maxCandidatesPerPlate))
	}
	if warning, _ := checkSchemaVersion(req.SchemaVersion); warning != "" {
		warnings = append(warnings, warning)
	}

	originalSpace := ""
	if sawBox {
		originalSpace = req.CoordinateSpace
	}

	run := models.ANPRRun{
		Key:                     target.Key,
		OrganisationId:          target.OrganisationId,
		DeviceId:                target.DeviceId,
		RecordingTimestamp:      target.RecordingTimestamp,
		Source:                  req.Source,
		SchemaVersion:           req.SchemaVersion,
		Media:                   models.ANPRMedia{Width: req.Media.Width, Height: req.Media.Height, Fps: req.Media.Fps},
		Plates:                  storedPlates,
		OriginalCoordinateSpace: originalSpace,
		CreatedAt:               time.Now().UnixMilli(),
	}

	report := Report{RunId: req.Source.RunId, Warnings: warnings, Detail: detail}
	return run, report, nil
}

// plateHasBox reports whether a plate carries any box coordinate (in either the
// {x,y,w,h} or {x1,y1,x2,y2} form). A partially-specified box still counts as
// present so it can be rejected with a missing-geometry reason rather than
// silently dropped.
func plateHasBox(p api.ANPRPlateInput) bool {
	return p.X != nil || p.Y != nil || p.W != nil || p.H != nil ||
		p.X1 != nil || p.Y1 != nil || p.X2 != nil || p.Y2 != nil
}

// normalizeANPRBox converts a wire plate box into a normalised ANPRBox, or
// returns a rejection reason. It mirrors normalizeDetectionBox: pixel boxes are
// divided by the media scale, boxes within boundsTolerance of the frame are
// clamped, those further out are rejected.
func normalizeANPRBox(p api.ANPRPlateInput, alreadyNormalized bool, scaleX, scaleY float64) (*models.ANPRBox, string) {
	var x1, y1, x2, y2 float64

	switch {
	case p.X != nil && p.Y != nil && p.W != nil && p.H != nil:
		if *p.W <= 0 || *p.H <= 0 {
			return nil, reasonPlateNonPositiveDim
		}
		x1, y1 = *p.X, *p.Y
		x2, y2 = *p.X+*p.W, *p.Y+*p.H
	case p.X1 != nil && p.Y1 != nil && p.X2 != nil && p.Y2 != nil:
		x1, y1, x2, y2 = *p.X1, *p.Y1, *p.X2, *p.Y2
		if x2 <= x1 || y2 <= y1 {
			return nil, reasonPlateNonPositiveDim
		}
	default:
		return nil, reasonPlateMissingGeom
	}

	if !alreadyNormalized {
		if scaleX <= 0 || scaleY <= 0 {
			return nil, reasonPlateMissingScale
		}
		x1, x2 = x1/scaleX, x2/scaleX
		y1, y2 = y1/scaleY, y2/scaleY
	}

	if !withinFrame(x1, y1, x2, y2) {
		return nil, reasonPlateBoxOutOfFrame
	}

	return &models.ANPRBox{
		X1: clampUnit(x1),
		Y1: clampUnit(y1),
		X2: clampUnit(x2),
		Y2: clampUnit(y2),
	}, ""
}

// clampConfidence constrains a confidence to [0, 1], reporting whether it was
// out of range so the caller can warn.
func clampConfidence(c float64) (float64, bool) {
	switch {
	case c < 0:
		return 0, true
	case c > 1:
		return 1, true
	default:
		return c, false
	}
}

// convertCandidates maps wire candidates onto stored candidates, clamping each
// confidence to [0, 1]. Candidates with empty text are dropped.
func convertCandidates(in []api.ANPRCandidateInput) []models.ANPRCandidate {
	if len(in) == 0 {
		return nil
	}
	out := make([]models.ANPRCandidate, 0, len(in))
	for _, c := range in {
		text := strings.TrimSpace(c.Plate)
		if text == "" {
			continue
		}
		conf, _ := clampConfidence(c.Confidence)
		out = append(out, models.ANPRCandidate{Plate: text, Confidence: conf})
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// --- Action ----------------------------------------------------------------

// UpsertANPRRun is the mandatory action: it persists the normalised run to the
// anpr collection, keyed by (Key, Source.RunId). It runs for every transport —
// an API push and a pipeline stage store the same run.
type UpsertANPRRun struct{}

func (UpsertANPRRun) Name() string       { return "upsert_anpr_run" }
func (UpsertANPRRun) RunFor(Source) bool { return true }

func (UpsertANPRRun) Apply(ctx context.Context, scope Scope, _ Target, run any) error {
	ar, ok := run.(models.ANPRRun)
	if !ok {
		return fmt.Errorf("ingest: upsert expected models.ANPRRun, got %T", run)
	}
	if scope.ANPR == nil {
		return errors.New("ingest: no ANPRStore configured on scope")
	}
	return scope.ANPR.UpsertANPRRun(ctx, ar)
}

// anprMarkerTag tags every marker created from an ANPR run so ANPR-sourced
// markers are filterable from markers other features create.
const anprMarkerTag = "anpr"

// CreateANPRMarkers is the optional second action: it turns each recognised
// plate into a user-facing marker (one per plate). Like the detection kind's
// PromoteTracksToRegions it is gated to the trusted pipeline transport — an
// external API push stores the run but does not auto-create markers (a policy
// decision, not just routing) — and it is a no-op when no MarkerStore is wired
// or the run recognised no plates.
type CreateANPRMarkers struct{}

func (CreateANPRMarkers) Name() string         { return "create_anpr_markers" }
func (CreateANPRMarkers) RunFor(s Source) bool { return s == SourcePipeline }

func (CreateANPRMarkers) Apply(ctx context.Context, scope Scope, _ Target, run any) error {
	ar, ok := run.(models.ANPRRun)
	if !ok {
		return fmt.Errorf("ingest: create markers expected models.ANPRRun, got %T", run)
	}
	if scope.Markers == nil || len(ar.Plates) == 0 {
		// No marker sink wired (a transport that only stores the run), or nothing
		// recognised — nothing to surface.
		return nil
	}
	markers := make([]models.Marker, 0, len(ar.Plates))
	for i := range ar.Plates {
		markers = append(markers, anprPlateToMarker(ar, ar.Plates[i]))
	}
	return scope.Markers.UpsertMarkers(ctx, markers)
}

// anprPlateToMarker maps one recognised plate onto a marker. The marker is a
// point-in-time event at the plate's representative time (the recording start
// plus the read's offset), named for the human-readable plate, tagged "anpr"
// so ANPR-sourced markers are filterable, with the read summarised in the
// description. Marker timestamps are in seconds (the Marker contract), so the
// plate's millisecond offset is converted. The marker's _id is left to the
// store to assign on first insert so a replay refreshes rather than renames.
func anprPlateToMarker(run models.ANPRRun, plate models.ANPRPlate) models.Marker {
	name := plate.Display
	if name == "" {
		name = plate.Plate
	}

	start := run.RecordingTimestamp + plate.TimestampMs/1000

	desc := fmt.Sprintf("ANPR plate %s", name)
	if plate.Country != "" {
		desc += fmt.Sprintf(" (%s)", plate.Country)
	}
	if plate.Confidence > 0 {
		desc += fmt.Sprintf(", confidence %.2f", plate.Confidence)
	}

	return models.Marker{
		DeviceId:       run.DeviceId,
		OrganisationId: run.OrganisationId,
		StartTimestamp: start,
		EndTimestamp:   start,
		Name:           name,
		Tags:           []models.MarkerTag{{Name: anprMarkerTag}},
		Description:    desc,
	}
}
