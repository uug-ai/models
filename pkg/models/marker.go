package models

import (
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Marker struct {
	Id primitive.ObjectID `json:"id" bson:"_id" example:"507f1f77bcf86cd799439011" required:"true"` // Unique identifier for the marker, generated automatically§§§

	// RBAC information
	DeviceId       string `json:"deviceId" bson:"deviceId" example:"686a906345c1df594939f9j25f4" required:"true"` // DeviceId is used to identify the device associated with the marker
	SiteId         string `json:"siteId,omitempty" bson:"siteId,omitempty" example:"686a906345c1df594pcsr3r45"`   // SiteId is used to identify the site where the marker is located
	GroupId        string `json:"groupId,omitempty" bson:"groupId,omitempty" example:"686a906345c1df594pmt41w4"`  // GroupId is used to identify the group of markers
	OrganisationId string `json:"organisationId" bson:"organisationId" example:"686a906345c1df594pad69f0"`        // OrganisationId is used to identify the organisation that owns the marker, retrieved from the user's access token

	// MediaKeys optionally pins this marker to specific recordings by their stable
	// key (media.videoFile). When set, the marker writer links the marker to
	// exactly these recordings — scoped to the marker's device and organisation —
	// instead of inferring the link from timestamp overlap. This is authoritative:
	// the producer (e.g. a stage worker) knows precisely which recording a
	// detection came from, so it is immune to timing/fps drift. Leave empty to keep
	// the default timestamp-overlap association.
	MediaKeys []string `json:"mediaKeys,omitempty" bson:"mediaKeys,omitempty" example:"1752482068_...device_1920_1080_10000.mp4"` // Recording keys (media.videoFile) this marker attaches to

	// Timing information (all timestamps are in seconds)
	StartTimestamp int64 `json:"startTimestamp" bson:"startTimestamp" example:"1752482068" required:"true"` // Start timestamp of the marker in seconds since epoch
	EndTimestamp   int64 `json:"endTimestamp" bson:"endTimestamp" example:"1752482079" required:"true"`     // End timestamp of the marker in seconds since epoch
	Duration       int64 `json:"duration" bson:"duration" example:"11" required:"true"`                     // Duration of the marker in seconds

	Name        string           `json:"name" bson:"name" example:"2-HCP-007" required:"true"`                                       // Name or identifier for the marker e.g., "a license plate (2-HCP-007), an unique identifier (transaction_id, point of sale), etc."
	Events      []MarkerEvent    `json:"events,omitempty" bson:"events,omitempty"`                                                   // Events associated with the marker, such as motion detected, sound detected, etc.
	Tags        []MarkerTag      `json:"tags,omitempty" bson:"tags,omitempty"`                                                       // Tags associated with the marker for categorization
	Description string           `json:"description,omitempty" bson:"description,omitempty" example:"Person forcably opened a door"` // Description of the marker
	Categories  []MarkerCategory `json:"categories,omitempty" bson:"categories,omitempty"`                                           // Category of the marker e.g., "security", "object", "alert", etc.

	// Additional metadata
	Metadata *MarkerMetadata `json:"metadata,omitempty" bson:"metadata,omitempty"` // Metadata associated with the marker, such as comments and tags

	// AtRuntimeMetadata contains metadata that is generated at runtime, which can include
	// more verbose information about the device's current state, capabilities, or configuration.
	// for example the linked sites details, etc.
	AtRuntimeMetadata *MarkerAtRuntimeMetadata `json:"atRuntimeMetadata,omitempty" bson:"atRuntimeMetadata,omitempty"`

	// Synchronize
	Synchronize *Synchronize `json:"synchronize,omitempty" bson:"synchronize,omitempty"` // Synchronization status with external systems

	// Audit information
	Audit *Audit `json:"audit,omitempty" bson:"audit,omitempty"` // Audit information for tracking changes to the marker
}

func (m *Marker) Validate() error {
	validate := validator.New()
	return validate.Struct(m)
}

type MarkerMetadata struct {
	Comments *Comment `json:"comments,omitempty" bson:"comments,omitempty"` // Additional comments or description of the marker

	// Confidence is the certainty of the detection that produced this marker, in the
	// range [0,1] (e.g. an ANPR plate-read or object-detection score). A nil pointer
	// means no confidence was reported; this keeps the value optional and lets a
	// genuine 0 be distinguished from "unset".
	Confidence *float64 `json:"confidence,omitempty" bson:"confidence,omitempty" example:"0.92"`

	// Source identifies the pipeline that produced this marker (e.g. "anpr",
	// "loitering", "dominantcolors"). This is provenance, distinct from the
	// user-facing Categories on the marker.
	Source string `json:"source,omitempty" bson:"source,omitempty" example:"anpr"`

	// Engine identifies the concrete backend used within the pipeline
	// (e.g. "fast-onnx", "tesseract", "http").
	Engine string `json:"engine,omitempty" bson:"engine,omitempty" example:"fast-onnx"`

	// ModelVersion records the version of the model that produced the detection,
	// so detections can be correlated with model changes after the fact.
	ModelVersion string `json:"modelVersion,omitempty" bson:"modelVersion,omitempty" example:"v1.4.0"`

	// BoundingBox is the normalized [0,1] region of the detection within the frame,
	// suitable for drawing overlays. Nil when no spatial information is available.
	BoundingBox *MarkerBox `json:"boundingBox,omitempty" bson:"boundingBox,omitempty"`

	// Raw is an escape hatch for pipeline-specific metadata that does not warrant a
	// first-class field. Values here are not indexed or validated; promote anything
	// you need to query into a typed field instead.
	Raw map[string]any `json:"raw,omitempty" bson:"raw,omitempty"`
}

// MarkerBox is a normalized bounding box in the range [0,1], where (X, Y) is the
// top-left corner and Width/Height are relative to the frame dimensions.
type MarkerBox struct {
	X      float64 `json:"x" bson:"x" example:"0.12"`           // Left edge, fraction of frame width
	Y      float64 `json:"y" bson:"y" example:"0.34"`           // Top edge, fraction of frame height
	Width  float64 `json:"width" bson:"width" example:"0.20"`   // Box width, fraction of frame width
	Height float64 `json:"height" bson:"height" example:"0.10"` // Box height, fraction of frame height
}

type MarkerAtRuntimeMetadata struct {
	MarkerRanges []MarkerOptionTimeRange `json:"markerRanges,omitempty" bson:"markerRanges,omitempty"`
	TagRanges    []MarkerTagTimeRange    `json:"tagRanges,omitempty" bson:"tagRanges,omitempty"`
	EventRanges  []MarkerEventTimeRange  `json:"eventRanges,omitempty" bson:"eventRanges,omitempty"`
}

/* Marker options */
type MarkerOption struct {
	Id             primitive.ObjectID `json:"id" bson:"_id" example:"507f1f77bcf86cd799439011" required:"true"`        // Unique identifier for the marker, generated automatically§§§
	OrganisationId string             `json:"organisationId" bson:"organisationId" example:"686a906345c1df594pad69f0"` // OrganisationId is used to identify the organisation that owns the marker, retrieved from the user's access token
	Value          string             `bson:"value" json:"value"`
	Text           string             `bson:"text" json:"text"`
	Categories     []string           `json:"categories,omitempty" bson:"categories,omitempty"`
	CreatedAt      int64              `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	UpdatedAt      int64              `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
}

type MarkerOptionTimeRange struct {
	Id             primitive.ObjectID `json:"id" bson:"_id" example:"507f1f77bcf86cd799439011" required:"true"` // Unique identifier for the marker, generated automatically§§§
	Value          string             `bson:"value" json:"value"`
	Text           string             `bson:"text" json:"text"`
	OrganisationId string             `json:"organisationId" bson:"organisationId" example:"686a906345c1df594pad69f0"` // OrganisationId is used to identify the organisation that owns the marker, retrieved from the user's access token
	Start          int64              `json:"start,omitempty" bson:"start,omitempty"`
	End            int64              `json:"end,omitempty" bson:"end,omitempty"`
	DeviceId       string             `bson:"deviceId" json:"deviceId"`
	CreatedAt      int64              `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	UpdatedAt      int64              `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
}

/* Marker Event */
type MarkerEvent struct { // Unique identifier for the event, generated automatically
	StartTimestamp int64  `json:"startTimestamp" bson:"startTimestamp" example:"1752482068" required:"true"` // Start timestamp of the marker in seconds since epoch
	EndTimestamp   int64  `json:"endTimestamp" bson:"endTimestamp" example:"1752482079" required:"true"`     // End timestamp of the marker in seconds since epoch
	Duration       int64  `json:"duration" bson:"duration" example:"11" required:"true"`                     // Duration of the marker in seconds
	Name           string `json:"name,omitempty" bson:"name,omitempty" example:"Motion Detected"`            // Name or identifier for the event e.g., "Motion Detected", "Sound Detected", etc.

	Description string   `json:"description,omitempty" bson:"description,omitempty" example:"Motion detected in the lobby area"` // Description of the event
	Tags        []string `json:"tags,omitempty" bson:"tags,omitempty" example:"[\"urgent\",\"review-needed\"]"`                  // Tags associated with the event for categorization
}

type MarkerEventOption struct {
	Id             primitive.ObjectID `json:"id" bson:"_id" example:"507f1f77bcf86cd799439011" required:"true"` // Unique identifier for the marker, generated automatically§§§
	Value          string             `bson:"value" json:"value"`
	Text           string             `bson:"text" json:"text"`
	OrganisationId string             `json:"organisationId" bson:"organisationId" example:"686a906345c1df594pad69f0"` // OrganisationId is used to identify the organisation that owns the marker, retrieved from the user's access token
	CreatedAt      int64              `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	UpdatedAt      int64              `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
}

type MarkerEventTimeRange struct {
	Id             primitive.ObjectID `json:"id" bson:"_id" example:"507f1f77bcf86cd799439011" required:"true"` // Unique identifier for the marker, generated automatically§§§
	Value          string             `bson:"value" json:"value"`
	Text           string             `bson:"text" json:"text"`
	OrganisationId string             `json:"organisationId" bson:"organisationId" example:"686a906345c1df594pad69f0"` // OrganisationId is used to identify the organisation that owns the marker, retrieved from the user's access token
	Start          int64              `json:"start,omitempty" bson:"start,omitempty"`
	End            int64              `json:"end,omitempty" bson:"end,omitempty"`
	DeviceId       string             `bson:"deviceId" json:"deviceId"`
	CreatedAt      int64              `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	UpdatedAt      int64              `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
}

/* Marker Tag */
type MarkerTag struct {
	Name string `json:"name,omitempty" bson:"name,omitempty" example:"Motion Detected"`
}

type MarkerTagOption struct {
	Id             primitive.ObjectID `json:"id" bson:"_id" example:"507f1f77bcf86cd799439011" required:"true"` // Unique identifier for the marker, generated automatically§§§
	Value          string             `bson:"value" json:"value"`
	Text           string             `bson:"text" json:"text"`
	OrganisationId string             `json:"organisationId" bson:"organisationId" example:"686a906345c1df594pad69f0"` // OrganisationId is used to identify the organisation that owns the marker, retrieved from the user's access token
	CreatedAt      int64              `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	UpdatedAt      int64              `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
}

type MarkerTagTimeRange struct {
	Id             primitive.ObjectID `json:"id" bson:"_id" example:"507f1f77bcf86cd799439011" required:"true"` // Unique identifier for the marker, generated automatically§§§
	Value          string             `bson:"value" json:"value"`
	Text           string             `bson:"text" json:"text"`
	OrganisationId string             `json:"organisationId" bson:"organisationId" example:"686a906345c1df594pad69f0"` // OrganisationId is used to identify the organisation that owns the marker, retrieved from the user's access token
	Start          int64              `json:"start,omitempty" bson:"start,omitempty"`
	End            int64              `json:"end,omitempty" bson:"end,omitempty"`
	DeviceId       string             `bson:"deviceId" json:"deviceId"`
	CreatedAt      int64              `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	UpdatedAt      int64              `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
}

type MarkerCategory struct {
	Name string `json:"name,omitempty" bson:"name,omitempty" example:"security"`
}

type MarkerCategoryOption struct {
	Id             primitive.ObjectID `json:"id" bson:"_id" example:"507f1f77bcf86cd799439011" required:"true"` // Unique identifier for the marker, generated automatically§§§
	Value          string             `bson:"value" json:"value"`
	Text           string             `bson:"text" json:"text"`
	OrganisationId string             `json:"organisationId" bson:"organisationId" example:"686a906345c1df594pad69f0"` // OrganisationId is used to identify the organisation that owns the marker, retrieved from the user's access token
	CreatedAt      int64              `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	UpdatedAt      int64              `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
}
