package models

import "encoding/json"

// WorkflowAnprOperation is the ANPR stage operation; its result is available to
// later conditions as results.anpr.
const WorkflowAnprOperation = "anpr"

// WorkflowAnprResult is results.anpr as the engine builds it from the ANPR
// worker's ingest envelope: blocks grouped by type, the type name pluralised by
// appending "s". A detection block is an api.PostDetectionsRequest and a marker
// block a Marker; the track meta keys are written by hub-anpr. Fields tagged
// condition:"-" carry constants, coordinates or bookkeeping and are not offered
// to conditions. Contract fields for the anpr stage are paths relative to this
// type (for example detections.*.tracks.*.meta.plate).
type WorkflowAnprResult struct {
	Detections   []WorkflowAnprDetection `json:"detections"`                 // exactly one detection run per stage result
	Markers      []WorkflowAnprMarker    `json:"markers"`                    // one per vehicle, same vehicles as the tracks
	MediaPatches []json.RawMessage       `json:"media-patchs" condition:"-"` // media description update, when a plate was read
}

// WorkflowAnprDetection is the stage's detection run (api.PostDetectionsRequest).
type WorkflowAnprDetection struct {
	MediaKey        string              `json:"mediaKey,omitempty" condition:"-"`
	AnalysisId      string              `json:"analysisId,omitempty" condition:"-"`
	Name            string              `json:"name,omitempty" condition:"-"`
	Task            string              `json:"task,omitempty" condition:"-"`
	SchemaVersion   string              `json:"schemaVersion,omitempty" condition:"-"`
	Source          json.RawMessage     `json:"source" condition:"-"`
	CoordinateSpace string              `json:"coordinateSpace" condition:"-"`
	Media           json.RawMessage     `json:"media,omitempty" condition:"-"`
	Categories      json.RawMessage     `json:"categories,omitempty" condition:"-"`
	Tracks          []WorkflowAnprTrack `json:"tracks"` // one per vehicle that passed the movement gates
}

// WorkflowAnprTrack is one vehicle (api.DetectionTrackInput).
type WorkflowAnprTrack struct {
	Id            string                `json:"id" condition:"-"`
	Label         string                `json:"label,omitempty" condition:"-"` // the plate, or "unread"
	ClassId       json.RawMessage       `json:"classId,omitempty" condition:"-"`
	Confidence    float64               `json:"confidence,omitempty" condition:"-"` // constant 0.9
	Color         string                `json:"color,omitempty" condition:"-"`
	Shape         string                `json:"shape,omitempty" condition:"-"`
	DeletedFrames json.RawMessage       `json:"deletedFrames,omitempty" condition:"-"`
	Meta          WorkflowAnprTrackMeta `json:"meta,omitempty"`
	Boxes         json.RawMessage       `json:"boxes" condition:"-"`
}

// WorkflowAnprTrackMeta is what hub-anpr records per vehicle.
type WorkflowAnprTrackMeta struct {
	Plate         string `json:"plate"`                       // uppercase A-Z0-9, 4-10 characters, no separators; "" when unread
	Unread        bool   `json:"unread"`                      // no plate was read
	ClassifiedAs  string `json:"classifiedAs"`                // the classify label (ANPR_PLATE_CLASSES), not recognised by ANPR
	RecognisedBy  string `json:"recognisedBy" condition:"-"`  // always hub-anpr
	RecognisedOcr bool   `json:"recognisedOcr" condition:"-"` // always !unread
}

// WorkflowAnprMarker is one vehicle's timeline marker (Marker).
type WorkflowAnprMarker struct {
	Id                string                      `json:"id" condition:"-"`
	DeviceId          string                      `json:"deviceId" condition:"-"`
	SiteId            string                      `json:"siteId,omitempty" condition:"-"`
	GroupId           string                      `json:"groupId,omitempty" condition:"-"`
	OrganisationId    string                      `json:"organisationId" condition:"-"`
	ProjectId         json.RawMessage             `json:"projectId,omitempty" condition:"-"`
	MediaKeys         []string                    `json:"mediaKeys,omitempty" condition:"-"`
	StartTimestamp    int64                       `json:"startTimestamp" condition:"-"`
	EndTimestamp      int64                       `json:"endTimestamp" condition:"-"`
	Duration          int64                       `json:"duration"` // seconds the vehicle was visible
	Name              string                      `json:"name"`     // the plate, or "anpr unread"
	Events            json.RawMessage             `json:"events,omitempty" condition:"-"`
	Tags              json.RawMessage             `json:"tags,omitempty" condition:"-"`
	Description       string                      `json:"description,omitempty" condition:"-"`
	Categories        json.RawMessage             `json:"categories,omitempty" condition:"-"`
	Metadata          *WorkflowAnprMarkerMetadata `json:"metadata,omitempty"`
	Detections        json.RawMessage             `json:"detections,omitempty" condition:"-"`
	AtRuntimeMetadata json.RawMessage             `json:"atRuntimeMetadata,omitempty" condition:"-"`
	Synchronize       json.RawMessage             `json:"synchronize,omitempty" condition:"-"`
	Audit             json.RawMessage             `json:"audit,omitempty" condition:"-"`
}

// WorkflowAnprMarkerMetadata is the OCR outcome (MarkerMetadata). Engine and
// Confidence are present only when a plate was read.
type WorkflowAnprMarkerMetadata struct {
	Confidence   *float64        `json:"confidence,omitempty"`           // OCR confidence: ~0-1 for fast, 0.5 for tesseract, unclamped for http
	Source       string          `json:"source,omitempty" condition:"-"` // always anpr
	Engine       string          `json:"engine,omitempty"`               // fast, tesseract, tesseract+opencv or http
	ModelVersion string          `json:"modelVersion,omitempty" condition:"-"`
	BoundingBox  json.RawMessage `json:"boundingBox,omitempty" condition:"-"`
	Raw          json.RawMessage `json:"raw,omitempty" condition:"-"`
	Comments     json.RawMessage `json:"comments,omitempty" condition:"-"`
}
