package api

import "github.com/uug-ai/models/pkg/models"

// PostANPR
//
// Wire format for an automatic-number-plate-recognition result. It is the typed
// body an ANPR producer returns — either as the `payload` of a delegated-ingest
// workflow stage (operation "anpr", routed through the shared ingest core) or,
// in future, as the body of an HTTP ingest door. The server validates and
// normalises it into a models.ANPRRun stored in the dedicated "anpr"
// collection, keyed by the recording and upserted by (recording key,
// Source.RunId).
//
// It deliberately does not reuse PostDetectionsRequest: a plate read is
// recognised text with candidate reads and a jurisdiction, a different contract
// from a bounding-box detection track.
type PostANPRRequest struct {
	// MediaKey is the recording KEY the run belongs to — the stable string
	// stored as media.videoFile / analysis.key (NOT the media document _id).
	// Provide this or AnalysisId (MediaKey wins when both are present). Only the
	// HTTP door uses these; over the workflows queue the engine resolves the
	// target from the run envelope.
	MediaKey string `json:"mediaKey,omitempty"`
	// AnalysisId targets the recording via its analysis document _id (an
	// ObjectID hex), as an alternative to MediaKey.
	AnalysisId string `json:"analysisId,omitempty"`
	// SchemaVersion is the producer's payload schema (optional).
	SchemaVersion string `json:"schemaVersion,omitempty"`
	// Source identifies the producer; Source.RunId is the upsert key.
	Source models.ANPRSource `json:"source"`
	// CoordinateSpace is the space plate boxes are sent in: "pixel" or
	// "normalized". Required only when a plate carries a box.
	CoordinateSpace string `json:"coordinateSpace,omitempty"`
	// Media gives the recording dimensions used to normalise pixel-space boxes.
	// Required when CoordinateSpace is "pixel" and any plate carries a box.
	Media ANPRMediaInput `json:"media,omitempty"`
	// Plates are the recognised plates. An empty list is valid — it records that
	// ANPR ran and found nothing.
	Plates []ANPRPlateInput `json:"plates"`
}

// ANPRMediaInput is the recording descriptor used to normalise pixel boxes.
type ANPRMediaInput struct {
	Width  int     `json:"width,omitempty"`
	Height int     `json:"height,omitempty"`
	Fps    float64 `json:"fps,omitempty"`
}

// ANPRPlateInput is one recognised plate on the wire. The box is optional and
// accepts both the preferred {x, y, w, h} (top-left + size) and the legacy
// {x1, y1, x2, y2} forms; pointers let the server detect which form was sent
// and distinguish "no geometry" from a zero coordinate.
type ANPRPlateInput struct {
	// Plate is the recognised registration; stored normalised for search.
	Plate string `json:"plate"`
	// Display is the optional human-readable rendering.
	Display string `json:"display,omitempty"`
	// Confidence is the read confidence; clamped to [0, 1] on ingest.
	Confidence float64 `json:"confidence,omitempty"`
	// Country is the ISO 3166-1 alpha-2 jurisdiction (optional).
	Country string `json:"country,omitempty"`
	// Region is the sub-jurisdiction (optional).
	Region string `json:"region,omitempty"`
	// Frame is the representative frame the read came from (optional).
	Frame int64 `json:"frame,omitempty"`
	// TimestampMs is the representative time in ms into the recording (optional).
	TimestampMs int64 `json:"timestampMs,omitempty"`
	// Box geometry — preferred {x, y, w, h}.
	X *float64 `json:"x,omitempty"`
	Y *float64 `json:"y,omitempty"`
	W *float64 `json:"w,omitempty"`
	H *float64 `json:"h,omitempty"`
	// Box geometry — legacy {x1, y1, x2, y2}.
	X1 *float64 `json:"x1,omitempty"`
	Y1 *float64 `json:"y1,omitempty"`
	X2 *float64 `json:"x2,omitempty"`
	Y2 *float64 `json:"y2,omitempty"`
	// Candidates are alternative reads ranked below Plate (optional).
	Candidates []ANPRCandidateInput `json:"candidates,omitempty"`
	// Meta carries producer-specific extras (e.g. vehicle attributes); capped on
	// ingest.
	Meta map[string]interface{} `json:"meta,omitempty"`
}

// ANPRCandidateInput is one alternative plate read on the wire.
type ANPRCandidateInput struct {
	Plate      string  `json:"plate"`
	Confidence float64 `json:"confidence,omitempty"`
}

// PostANPRResponse echoes what was stored plus any plates rejected during
// normalisation.
type PostANPRResponse struct {
	RunId        string          `json:"runId"`
	PlatesStored int             `json:"platesStored"`
	Rejected     []ANPRRejection `json:"rejected"`
	Warnings     []string        `json:"warnings"`
}

// ANPRRejection identifies a single plate that failed validation/normalisation.
type ANPRRejection struct {
	Plate  string `json:"plate"`
	Frame  int64  `json:"frame"`
	Reason string `json:"reason"`
}

type PostANPRSuccessResponse struct {
	SuccessResponse
	Data PostANPRResponse `json:"data"`
}

type PostANPRErrorResponse struct {
	ErrorResponse
}
