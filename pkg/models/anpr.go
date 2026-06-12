package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// ANPRRun is one automatic-number-plate-recognition run for a recording. Like
// DetectionRun it is stored in its own dedicated collection ("anpr") keyed by
// the recording (Key) so a recording can accumulate many runs without bloating
// its analysis document, and each run is upserted by (Key, Source.RunId) so a
// re-post or at-least-once redelivery replaces rather than duplicates.
//
// ANPR deliberately does NOT reuse the detection/redaction track shape
// (FaceRedactionTrack): a plate read is recognised *text* with candidate reads
// and a jurisdiction, not a per-frame box trajectory selected for redaction.
// The two share only provenance conventions (a Source carrying the upsert key).
type ANPRRun struct {
	// Id is the MongoDB document id, assigned on first insert.
	Id primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	// Key is the recording/media key the run belongs to — the stable identity
	// the collection is keyed by (survives re-analysis).
	Key string `json:"key,omitempty" bson:"key,omitempty"`
	// OrganisationId scopes the run to the owning organisation. Never serialised
	// out (consistent with DetectionRun).
	OrganisationId string `json:"-" bson:"organisationId,omitempty"`
	// DeviceId is denormalised from the recording for convenient filtering and
	// cascade cleanup; it is not authoritative.
	DeviceId string `json:"deviceId,omitempty" bson:"deviceId,omitempty"`
	// Source identifies the producer and carries RunId, the natural upsert key.
	Source ANPRSource `json:"source" bson:"source"`
	// SchemaVersion is the producer's payload schema (optional).
	SchemaVersion string `json:"schemaVersion,omitempty" bson:"schemaVersion,omitempty"`
	// Media describes the recording the plates were read against; kept so a
	// normalised plate box can be rendered back at pixel scale.
	Media ANPRMedia `json:"media,omitempty" bson:"media,omitempty"`
	// Plates are the recognised plates, stored verbatim after the box (if any)
	// is normalised to the [0, 1] frame.
	Plates []ANPRPlate `json:"plates" bson:"plates"`
	// OriginalCoordinateSpace records the space the producer sent plate boxes in
	// ("pixel" or "normalized") before the server normalised them. Empty when no
	// plate carried a box.
	OriginalCoordinateSpace string `json:"originalCoordinateSpace,omitempty" bson:"originalCoordinateSpace,omitempty"`
	// CreatedAt is set by the server when the run is first stored (epoch millis).
	CreatedAt int64 `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	// UpdatedAt is set by the server every time the run is written (epoch millis).
	UpdatedAt int64 `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
	// RecordingTimestamp is the start time (epoch seconds) of the recording,
	// denormalised on write so cleanup expires the run on the recording's
	// retention clock rather than the (possibly much later) post time.
	RecordingTimestamp int64 `json:"recordingTimestamp,omitempty" bson:"recordingTimestamp,omitempty"`
}

// ANPRSource identifies the producer of an ANPR run. RunId is the natural key
// the upsert matches on for a recording.
type ANPRSource struct {
	Kind    string `json:"kind" bson:"kind"` // pipeline | model | import
	Name    string `json:"name" bson:"name"` // producer identifier
	Version string `json:"version" bson:"version"`
	RunId   string `json:"runId" bson:"runId"` // ULID/UUID; upsert key
}

// ANPRMedia describes the recording an ANPR run was produced against. Width and
// Height let a normalised plate box be projected back to pixels.
type ANPRMedia struct {
	Width  int     `json:"width,omitempty" bson:"width,omitempty"`
	Height int     `json:"height,omitempty" bson:"height,omitempty"`
	Fps    float64 `json:"fps,omitempty" bson:"fps,omitempty"`
}

// ANPRPlate is one recognised number plate. The recognised text (Plate) is the
// searchable, normalised form; Display preserves the producer's human-readable
// rendering. Candidates carry the alternative reads an OCR engine ranked below
// the chosen Plate so a consumer can re-rank or audit the recognition.
type ANPRPlate struct {
	// Plate is the recognised registration in a normalised, searchable form
	// (uppercase, no separators), e.g. "1ABC234".
	Plate string `json:"plate" bson:"plate"`
	// Display is the optional human-readable rendering, e.g. "1-ABC-234".
	Display string `json:"display,omitempty" bson:"display,omitempty"`
	// Confidence is the chosen read's confidence in [0, 1].
	Confidence float64 `json:"confidence,omitempty" bson:"confidence,omitempty"`
	// Country is the ISO 3166-1 alpha-2 jurisdiction, e.g. "BE" (optional).
	Country string `json:"country,omitempty" bson:"country,omitempty"`
	// Region is the sub-jurisdiction (state/province) when the producer reports
	// one, e.g. "CA" (optional).
	Region string `json:"region,omitempty" bson:"region,omitempty"`
	// Frame is the representative frame the chosen read came from.
	Frame int64 `json:"frame,omitempty" bson:"frame,omitempty"`
	// TimestampMs is the representative time (ms into the recording) of the read.
	TimestampMs int64 `json:"timestampMs,omitempty" bson:"timestampMs,omitempty"`
	// Box is the plate's location at the representative frame, normalised to the
	// [0, 1] frame. Nil when the producer reported no geometry.
	Box *ANPRBox `json:"box,omitempty" bson:"box,omitempty"`
	// Candidates are alternative reads ranked below Plate, highest confidence
	// first (optional).
	Candidates []ANPRCandidate `json:"candidates,omitempty" bson:"candidates,omitempty"`
	// Meta carries producer-specific extras (e.g. vehicle make/colour) without
	// expanding the core contract. Capped on ingest.
	Meta map[string]interface{} `json:"meta,omitempty" bson:"meta,omitempty"`
}

// ANPRCandidate is one alternative plate read with its confidence in [0, 1].
type ANPRCandidate struct {
	Plate      string  `json:"plate" bson:"plate"`
	Confidence float64 `json:"confidence,omitempty" bson:"confidence,omitempty"`
}

// ANPRBox is a plate bounding box normalised to the [0, 1] frame. It is
// intentionally separate from the redaction TrackBox: an ANPR box is a single
// location for a recognised plate, not an editable per-frame redaction region.
type ANPRBox struct {
	X1 float64 `json:"x1" bson:"x1"`
	Y1 float64 `json:"y1" bson:"y1"`
	X2 float64 `json:"x2" bson:"x2"`
	Y2 float64 `json:"y2" bson:"y2"`
}
