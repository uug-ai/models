package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type AnalysisWrapper struct {
	Id                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Key                string             `json:"key,omitempty"`
	Provider           string             `json:"provider" bson:"provider"`
	Source             string             `json:"source" bson:"source"`
	InProcess          int64              `json:"in_process,omitempty"`
	Timestamp          int64              `json:"timestamp,omitempty"`
	Start              int64              `json:"start,omitempty"`
	End                int64              `json:"end,omitempty"`
	UserId             string             `json:"user_id,omitempty"`
	DeviceId           string             `json:"device_id,omitempty"`
	AsyncOperations    []string           `json:"asyncOperations,omitempty"`
	RequiredOperations []string           `json:"requiredOperations,omitempty"`
	ResolvedOperations []string           `json:"resolvedOperations,omitempty"`
	Favourite          bool               `json:"favourite,omitempty"`
	Data               Analysis           `json:"data,omitempty"`
}

type AnalysisShort struct {
	Id                primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Key               string             `json:"key,omitempty"`
	Provider          string             `json:"provider" bson:"provider"`
	Source            string             `json:"source" bson:"source"`
	FileUrl           string             `json:"src" bson:"src"`
	Timestamp         int64              `json:"timestamp,omitempty"`
	Start             int64              `json:"start,omitempty"`
	UserId            string             `json:"user_id,omitempty"`
	DeviceId          string             `json:"device_id,omitempty"`
	ThumbnaiFile      string             `json:"thumbnail_file,omitempty"`
	ThumbnailProvider string             `json:"thumbnail_provider,omitempty"`
	ThumbnailUrl      string             `json:"thumbnail_src"`
	Favourite         bool               `json:"favourite,omitempty"`
	Data              *Analysis          `json:"data,omitempty"`
}

type Analysis struct {
	Classify      Classify        `json:"classify" bson:"classify"`
	Counting      Counting        `json:"counting" bson:"counting"`
	DominantColor DominantColor   `json:"dominantcolor" bson:"dominantcolor"`
	Thumby        Thumby          `json:"thumby" bson:"thumby"`
	Sprite        Sprite          `json:"sprite" bson:"sprite"`
	FaceRedaction []FaceRedaction `json:"faceRedaction" bson:"faceRedaction"`
}

type Classify struct {
	Properties []string          `json:"properties" bson:"properties"`
	Details    []ClassifyDetails `json:"details" bson:"details"`
}

type ClassifyDetails struct {
	Id               string      `json:"id" bson:"id"`
	X                float64     `json:"x" bson:"x"`
	Y                float64     `json:"y" bson:"y"`
	W                float64     `json:"w" bson:"w"`
	H                float64     `json:"h" bson:"h"`
	FrameWidth       float64     `json:"frameWidth" bson:"frameWidth"`
	FrameHeight      float64     `json:"frameHeight" bson:"frameHeight"`
	MeanColor        float64     `json:"meanColor" bson:"meanColor"`
	Frame            float64     `json:"frame" bson:"frame"`
	Occurence        float64     `json:"occurence" bson:"occurence"`
	Distance         float64     `json:"distance" bson:"distance"`
	Valid            bool        `json:"valid" bson:"valid"`
	Classified       string      `json:"classified" bson:"classified"`
	Traject          [][]float64 `json:"traject" bson:"traject"`
	ColorString      []string    `json:"colorStr" bson:"colorStr"`
	IsStatic         bool        `json:"isStatic" bson:"isStatic"`
	Frames           []int64     `json:"frames" bson:"frames"`
	StaticDistance   float64     `json:"staticDistance" bson:"staticDistance"`
	TrajectCentroids [][]float64 `json:"trajectCentroids" bson:"trajectCentroids"`
}

type Color struct {
	Name  string `json:"name" bson:"name"`
	Count int    `json:"count" bson:"count"`
}

type Counting struct {
	Details []CountingDetail `json:"detail,omitempty"`
	Regions []Region         `json:"regions,omitempty"`
	Records []CountingRecord `json:"records" bson:"records"`
}

type CountingDetail struct {
	DeviceID    string    `json:"deviceId" bson:"deviceId"`
	ObjectName  string    `json:"objectName" bson:"objectName"`
	Timestamp   string    `json:"timestamp" bson:"timestamp"`
	VideoWidth  float64   `json:"videoWidth" bson:"videoWidth"`
	VideoHeight float64   `json:"videoHeight" bson:"videoHeight"`
	Segment     string    `json:"segment" bson:"segment"`
	ObjectId    string    `json:"objectId" bson:"objectId"`
	Position    []float64 `json:"position" bson:"position"`
}

type CountingRecord struct {
	Type       string  `json:"type" bson:"type"`
	SegmentId  string  `json:"segmentId" bson:"segmentId"`
	Username   string  `json:"username" bson:"username"`
	AlertId    string  `json:"alert_id" bson:"alert_id"`
	AlertName  string  `json:"alert_name" bson:"alert_name"`
	Timestamp  string  `json:"timestamp" bson:"timestamp"`
	DeviceId   string  `json:"deviceId" bson:"deviceId"`
	ObjectName string  `json:"objectName" bson:"objectName"`
	Count      int     `json:"count" bson:"count"`
	Duration   float64 `json:"duration" bson:"duration"`
}

type DominantColor struct {
	Rgbs [][]float64 `json:"rgbs" bson:"rgbs"`
	Hexs []string    `json:"hexs" bson:"hexs"`
}

type Thumby struct {
	ThumbnailFile string `json:"filename" bson:"filename"`
	Provider      string `json:"provider" bson:"provider"`
	Base64        string `json:"base64" bson:"base64"`
	Interval      int    `json:"interval" bson:"interval"` // Only for sprites (Hacky)
	//Quality       string `json:"quality" bson:"quality"` -> Is getting type errors due to new mongodb driver.. need to fix.
	//Width  string `json:"width" bson:"width"` -> Is getting type errors due to new mongodb driver.. need to fix.
	//Height string `json:"height" bson:"height"` -> Is getting type errors due to new mongodb driver.. need to fix.
}

type Sprite struct {
	SpriteFile string `json:"filename" bson:"filename"`
	Provider   string `json:"provider" bson:"provider"`
	Interval   int    `json:"interval" bson:"interval"`
}

type AnalysisFilter struct {
	Start      int64    `json:"start" bson:"start"`
	Limit      int64    `json:"limit" bson:"limit"`
	Sort       string   `json:"sort" bson:"sort"`
	Operations []string `json:"operations" bson:"operations"`
}

type Thumbnail struct {
	Key           string `json:"key" bson:"key"`
	ThumbnailFile string `json:"thumbnailFile" bson:"thumbnailFile"`
}

type FaceRedactionTrack struct {
	Id               string             `json:"id" bson:"id"`
	Classified       string             `json:"classified" bson:"classified"`
	Frames           []int64            `json:"frames" bson:"frames"`
	Traject          [][]float64        `json:"traject" bson:"traject"` // [x1, y1, x2, y2, frame]
	ColorString      []string           `json:"colorStr,omitempty" bson:"colorStr,omitempty"`
	Selected         bool               `json:"selected" bson:"selected"`
	DeletedFrames    []int64            `json:"deletedFrames,omitempty" bson:"deletedFrames,omitempty"`
	FrameCoordinates map[int64]TrackBox `json:"frameCoordinates,omitempty" bson:"frameCoordinates,omitempty"` // frame -> [x1, y1, x2, y2]
	// Confidence, ClassId and Shape preserve track-level provenance from a
	// detection run; empty for tracks created by the editor.
	Confidence float64 `json:"confidence,omitempty" bson:"confidence,omitempty"`
	ClassId    *int    `json:"classId,omitempty" bson:"classId,omitempty"`
	Shape      string  `json:"shape,omitempty" bson:"shape,omitempty"`
}

type TrackBox struct {
	X1       float64 `json:"x1" bson:"x1"`
	Y1       float64 `json:"y1" bson:"y1"`
	X2       float64 `json:"x2" bson:"x2"`
	Y2       float64 `json:"y2" bson:"y2"`
	TrackId  string  `json:"trackId,omitempty" bson:"trackId,omitempty"`
	Smoothed bool    `json:"smoothed" bson:"smoothed"`
	Edited   bool    `json:"edited" bson:"edited"`
	// Confidence, ClassId and Label preserve the producer's per-box model
	// output so a stored detection run can be re-thresholded or audited later.
	Confidence float64 `json:"confidence,omitempty" bson:"confidence,omitempty"`
	ClassId    *int    `json:"classId,omitempty" bson:"classId,omitempty"`
	Label      string  `json:"label,omitempty" bson:"label,omitempty"`
}

type FaceRedaction struct {
	Id     primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	Tracks []FaceRedactionTrack `json:"tracks" bson:"tracks"`
	// CaseMediaId back-references the case_media entry (role=edit,
	// action=redaction) created when the redaction is submitted. The
	// lifecycle (queued/processing/completed/failed) is owned by that
	// case_media document; this struct only carries the tracks/inputs
	// the user has defined.
	CaseMediaId *primitive.ObjectID `json:"caseMediaId,omitempty" bson:"caseMediaId,omitempty"`
}

// DetectionRun is one detection run for a recording. Runs are stored in a
// dedicated "detections" collection keyed by the recording (Key) rather than
// embedded on the analysis, so a recording can accumulate many runs without
// bloating its analysis document. Each run is upserted by (Key, Source.RunId)
// so re-posts are idempotent and multiple producers can coexist. It stores the
// tracks verbatim (after coordinate normalisation) alongside provenance about
// the producer. Tracks reuse FaceRedactionTrack so promoting a run into a
// redaction is a direct copy.
type DetectionRun struct {
	// Id is the MongoDB document id, assigned on first insert.
	Id primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	// Key is the recording/media key the run belongs to. It is the stable
	// identity the collection is keyed by (survives re-analysis).
	Key string `json:"key,omitempty" bson:"key,omitempty"`
	// OrganisationId scopes the run to the owning organisation (matched on the
	// "userid" field, consistent with media/analysis). Never serialised out.
	OrganisationId string `json:"-" bson:"userid,omitempty"`
	// DeviceId is denormalised from the recording for convenient filtering and
	// cascade cleanup; it is not authoritative.
	DeviceId string `json:"deviceId,omitempty" bson:"deviceId,omitempty"`
	// Task is a forward-compatibility discriminator for the kind of run.
	// Defaults to "detection" when omitted.
	Task          string               `json:"task,omitempty" bson:"task,omitempty"`
	Source        DetectionSource      `json:"source" bson:"source"`
	SchemaVersion string               `json:"schemaVersion,omitempty" bson:"schemaVersion,omitempty"`
	Media         DetectionMedia       `json:"media,omitempty" bson:"media,omitempty"`
	Categories    []DetectionCategory  `json:"categories,omitempty" bson:"categories,omitempty"`
	Tracks        []FaceRedactionTrack `json:"tracks" bson:"tracks"`
	// OriginalCoordinateSpace records the coordinate space the producer sent
	// ("pixel" or "normalized") before the server normalised it on write.
	// Tracks are always stored normalised; this preserves the original for audit.
	OriginalCoordinateSpace string `json:"originalCoordinateSpace,omitempty" bson:"originalCoordinateSpace,omitempty"`
	// OriginalBoxForm records the box geometry form the producer sent
	// ("xywh", "xyxy", or "mixed") before conversion to the editor's TrackBox.
	OriginalBoxForm string `json:"originalBoxForm,omitempty" bson:"originalBoxForm,omitempty"`
	// CreatedAt is set by the server when the run is first stored (epoch millis).
	CreatedAt int64 `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	// UpdatedAt is set by the server every time the run is written (epoch millis).
	UpdatedAt int64 `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
	// RecordingTimestamp is the start time (epoch seconds) of the recording the
	// run belongs to, denormalised from the analysis on write. It lets cleanup
	// expire a run on the same retention clock as its recording rather than by
	// the (possibly much later) post time.
	RecordingTimestamp int64 `json:"recordingTimestamp,omitempty" bson:"recordingTimestamp,omitempty"`
}

// DetectionTask is the default value of DetectionRun.Task when a producer does
// not send one.
const DetectionTask = "detection"

// DetectionSource identifies the producer of a detection run. RunId is the
// natural key the upsert matches on within an analysis.
type DetectionSource struct {
	Kind           string  `json:"kind" bson:"kind"` // pipeline | model | import
	Name           string  `json:"name" bson:"name"` // producer identifier / editor layer label
	Version        string  `json:"version" bson:"version"`
	RunId          string  `json:"runId" bson:"runId"` // ULID/UUID; upsert key
	InputWidth     int     `json:"inputWidth,omitempty" bson:"inputWidth,omitempty"`
	InputHeight    int     `json:"inputHeight,omitempty" bson:"inputHeight,omitempty"`
	ScoreThreshold float64 `json:"scoreThreshold,omitempty" bson:"scoreThreshold,omitempty"`
	NmsIou         float64 `json:"nmsIou,omitempty" bson:"nmsIou,omitempty"`
	// RotationApplied indicates whether boxes are against the rotated/oriented
	// frame. Defaults to true on the wire; a nil pointer is treated as true.
	RotationApplied *bool `json:"rotationApplied,omitempty" bson:"rotationApplied,omitempty"`
}

// DetectionMedia describes the media the detections were produced against.
type DetectionMedia struct {
	Width      int     `json:"width,omitempty" bson:"width,omitempty"`
	Height     int     `json:"height,omitempty" bson:"height,omitempty"`
	Fps        float64 `json:"fps,omitempty" bson:"fps,omitempty"`
	FrameCount int64   `json:"frameCount,omitempty" bson:"frameCount,omitempty"`
	Rotation   int     `json:"rotation,omitempty" bson:"rotation,omitempty"`
}

// DetectionCategory is one entry in a producer's class taxonomy. Stored verbatim.
type DetectionCategory struct {
	Id    int    `json:"id" bson:"id"`
	Name  string `json:"name" bson:"name"`
	Alias string `json:"alias,omitempty" bson:"alias,omitempty"`
}

// RedactionMode describes the visual technique applied by the
// hub-pipeline-redaction worker over each track region.
type RedactionMode string

const (
	RedactionModeBlur     RedactionMode = "blur"
	RedactionModePixelate RedactionMode = "pixelate"
	RedactionModeBlack    RedactionMode = "black"
)
