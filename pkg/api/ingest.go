package api

import "encoding/json"

// Wire format for POST /ingest — the general ingest door. It is the transport
// of the {operation, payload} contract the shared ingest core consumes: a kind
// selector, a recording reference, and the typed result body. The door maps
// this onto ingest.Ingest, which routes by operation and runs that kind's
// ordered actions.
//
// This is the HTTP-transport sibling of the queue's PipelinePayload.Result: the
// Payload here carries the same kind-specific result body (e.g. a
// PostDetectionsRequest for the "detection" kind). The recording reference is
// carried at the envelope level (MediaKey or AnalysisId) so the door never has
// to peek inside the kind-specific payload to resolve the target.
//
// @Router /ingest [post]
type IngestRequest struct {
	// Operation is the kind selector — the registry key the ingest dispatcher
	// routes on (e.g. "detection").
	Operation string `json:"operation"`
	// MediaKey is the recording KEY the result belongs to — the stable string
	// stored as media.videoFile / analysis.key (NOT the media document _id).
	// Provide this or AnalysisId (MediaKey wins when both are present).
	MediaKey string `json:"mediaKey,omitempty"`
	// AnalysisId targets the recording via its analysis document _id (an
	// ObjectID hex), as an alternative to MediaKey.
	AnalysisId string `json:"analysisId,omitempty"`
	// Payload is the kind's typed result body, validated and normalised by the
	// kind's handler (for "detection" this is a PostDetectionsRequest).
	Payload json.RawMessage `json:"payload"`
}

// IngestResponse echoes the result of running a kind's action sequence. Today
// the only kind is "detection", so the report mirrors PostDetectionsResponse;
// the shape is kept here (not aliased) so a future kind with a different report
// does not have to break the detection contract.
type IngestResponse struct {
	RunId        string               `json:"runId"`
	TracksStored int                  `json:"tracksStored"`
	BoxesStored  int                  `json:"boxesStored"`
	Rejected     []DetectionRejection `json:"rejected"`
	Warnings     []string             `json:"warnings"`
}

type IngestSuccessResponse struct {
	SuccessResponse
	Data IngestResponse `json:"data"`
}

type IngestErrorResponse struct {
	ErrorResponse
}
