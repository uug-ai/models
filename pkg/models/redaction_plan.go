package models

import (
	"fmt"
	"math"
	"sort"
	"unicode/utf8"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RedactionPlanSchemaVersion is the version of the RedactionPlan shape this
// package reads and writes. A reader refuses a plan with any other value.
const RedactionPlanSchemaVersion = 1

// RedactionPlan is the saved state of the redaction editor for one video:
// the regions to hide and, for each region, how it is hidden, on which
// frames, and where it sits on those frames. It holds no pixels. It is the
// instruction the redaction worker follows to produce the redacted file, and
// afterwards the record of what that file hides.
//
// There is one plan per redacted version, stored on that version's
// case_media edit row (CaseMedia.Redaction), and at most one unrendered plan
// per source video, inside the draft on the source row
// (CaseMedia.RedactionDraft). The worker receives it as
// api.SubmitFaceRedactionRequest.Plan.
//
// It is unrelated to subscription plans (Plan, PlanDefinition, PlanCatalog).
//
// RedactionRegion.BoxAt is the only rule that turns a plan into boxes. The
// editor previews with the same rule, so the render equals the preview.
type RedactionPlan struct {
	// SchemaVersion is RedactionPlanSchemaVersion.
	SchemaVersion int `json:"schemaVersion" bson:"schemaVersion"`
	// Media records what the editor measured on the video the plan was made
	// against. It is informational and never used to place a box.
	Media *RedactionPlanMedia `json:"media,omitempty" bson:"media,omitempty"`
	// Regions are the areas to hide. Order has no meaning.
	Regions []RedactionRegion `json:"regions" bson:"regions"`
}

// RedactionPlanMedia describes the video a plan was made against.
type RedactionPlanMedia struct {
	Width      int     `json:"width,omitempty" bson:"width,omitempty"`
	Height     int     `json:"height,omitempty" bson:"height,omitempty"`
	Fps        float64 `json:"fps,omitempty" bson:"fps,omitempty"`
	FrameCount int64   `json:"frameCount,omitempty" bson:"frameCount,omitempty"`
}

// RedactionRegionOrigin says where a region came from.
type RedactionRegionOrigin string

const (
	// RedactionRegionOriginDetected is a region copied from a detector's track.
	RedactionRegionOriginDetected RedactionRegionOrigin = "detected"
	// RedactionRegionOriginManual is a region the user drew.
	RedactionRegionOriginManual RedactionRegionOrigin = "manual"
)

// RedactionShape is the outline of a region inside its box.
type RedactionShape string

const (
	// RedactionShapeRectangle fills the whole box.
	RedactionShapeRectangle RedactionShape = "rectangle"
)

// RedactionRegion is one area to hide. Frames are decode-order indexes
// starting at 0.
type RedactionRegion struct {
	// Id is unique within the plan and set by the editor.
	Id string `json:"id" bson:"id"`
	// Label is the name shown in the editor.
	Label  string                `json:"label,omitempty" bson:"label,omitempty"`
	Origin RedactionRegionOrigin `json:"origin" bson:"origin"`
	// Class is what the detector called the region ("face", "person"). Empty
	// for a drawn region.
	Class string `json:"class,omitempty" bson:"class,omitempty"`
	// Source is the detection run and track the region was copied from.
	Source *RedactionRegionSource `json:"source,omitempty" bson:"source,omitempty"`
	Shape  RedactionShape         `json:"shape" bson:"shape"`
	Style  RedactionStyle         `json:"style" bson:"style"`
	// Enabled is the editor's checkbox. A disabled region is stored and not
	// rendered.
	Enabled bool `json:"enabled" bson:"enabled"`
	// Start and End are the first and last frame the region applies to,
	// inclusive.
	Start int64 `json:"start" bson:"start"`
	End   int64 `json:"end" bson:"end"`
	// Hidden lists frame ranges where the region is not drawn.
	Hidden []RedactionFrameRange `json:"hidden,omitempty" bson:"hidden,omitempty"`
	// Keyframes are the region's positions, sorted by frame, at least one.
	// A keyframe outside Start..End still anchors the interpolation.
	Keyframes []RedactionKeyframe `json:"keyframes" bson:"keyframes"`
}

// RedactionRegionSource points at the detector output a region was copied from.
type RedactionRegionSource struct {
	// RunId is DetectionRun.Source.RunId.
	RunId string `json:"runId,omitempty" bson:"runId,omitempty"`
	// TrackId is FaceRedactionTrack.Id within that run.
	TrackId string `json:"trackId,omitempty" bson:"trackId,omitempty"`
}

// RedactionStyle is how a region is hidden.
type RedactionStyle struct {
	Mode RedactionMode `json:"mode" bson:"mode"`
}

// RedactionFrameRange is a range of frames, both ends inclusive.
type RedactionFrameRange struct {
	From int64 `json:"from" bson:"from"`
	To   int64 `json:"to" bson:"to"`
}

// RedactionKeyframe is the position of a region on one frame. Coordinates
// are fractions of the frame as displayed (rotation applied), 0 to 1, with
// X1 < X2 and Y1 < Y2.
type RedactionKeyframe struct {
	Frame int64   `json:"frame" bson:"frame"`
	X1    float64 `json:"x1" bson:"x1"`
	Y1    float64 `json:"y1" bson:"y1"`
	X2    float64 `json:"x2" bson:"x2"`
	Y2    float64 `json:"y2" bson:"y2"`
	// Edited marks a position the user set. Absent on positions copied from
	// a detector.
	Edited bool `json:"edited,omitempty" bson:"edited,omitempty"`
}

// RedactionBox is where a region is drawn on one frame, in the coordinates
// of RedactionKeyframe.
type RedactionBox struct {
	X1 float64 `json:"x1"`
	Y1 float64 `json:"y1"`
	X2 float64 `json:"x2"`
	Y2 float64 `json:"y2"`
}

// RedactionDraft is a plan saved without rendering it. It lives on the
// source row, one per source, shared by everyone who may edit the case.
type RedactionDraft struct {
	Plan RedactionPlan `json:"plan" bson:"plan"`
	// Revision increases by one on every save. A save sends the revision it
	// loaded and is refused when the stored one differs.
	Revision int64 `json:"revision" bson:"revision"`
	// BaseVersionId is the edit row the draft started from, if any.
	BaseVersionId *primitive.ObjectID `json:"baseVersionId,omitempty" bson:"base_version_id,omitempty"`
	UpdatedAt     int64               `json:"updatedAt,omitempty" bson:"updated_at,omitempty"`
	UpdatedBy     string              `json:"updatedBy,omitempty" bson:"updated_by,omitempty"`
}

// RedactionPlanLimits bounds the size of a plan.
type RedactionPlanLimits struct {
	// MaxRegions is the most regions a plan may hold.
	MaxRegions int `json:"maxRegions"`
	// MaxKeyframes is the most keyframes a plan may hold over all regions.
	MaxKeyframes int `json:"maxKeyframes"`
	// MaxLabelLength is the most characters a region label may have.
	MaxLabelLength int `json:"maxLabelLength"`
}

// DefaultRedactionPlanLimits returns the limits Validate applies. 50,000
// keyframes is about 5 MB of BSON, well under MongoDB's 16 MB document limit.
func DefaultRedactionPlanLimits() RedactionPlanLimits {
	return RedactionPlanLimits{MaxRegions: 200, MaxKeyframes: 50000, MaxLabelLength: 80}
}

// DrawnOn reports whether the region is drawn on the frame: it is enabled,
// the frame is within Start..End, and the frame is in no hidden range.
func (r *RedactionRegion) DrawnOn(frame int64) bool {
	if r == nil || !r.Enabled || frame < r.Start || frame > r.End {
		return false
	}
	for _, hidden := range r.Hidden {
		if frame >= hidden.From && frame <= hidden.To {
			return false
		}
	}
	return true
}

// BoxAt returns the region's box on the frame, and false when the region is
// not drawn there (see DrawnOn) or has no keyframes.
//
// The box is the keyframe on that frame when there is one. Between two
// keyframes each coordinate is interpolated linearly. Before the first
// keyframe it is the first keyframe's box, after the last keyframe the last
// one's. Keyframes must be sorted by frame, which Validate checks.
func (r *RedactionRegion) BoxAt(frame int64) (RedactionBox, bool) {
	if !r.DrawnOn(frame) || len(r.Keyframes) == 0 {
		return RedactionBox{}, false
	}
	keyframes := r.Keyframes
	// next is the first keyframe at or after the frame.
	next := sort.Search(len(keyframes), func(i int) bool { return keyframes[i].Frame >= frame })
	if next == len(keyframes) {
		return keyframes[len(keyframes)-1].box(), true
	}
	if next == 0 || keyframes[next].Frame == frame {
		return keyframes[next].box(), true
	}
	a, b := keyframes[next-1], keyframes[next]
	t := float64(frame-a.Frame) / float64(b.Frame-a.Frame)
	return RedactionBox{
		X1: a.X1 + (b.X1-a.X1)*t,
		Y1: a.Y1 + (b.Y1-a.Y1)*t,
		X2: a.X2 + (b.X2-a.X2)*t,
		Y2: a.Y2 + (b.Y2-a.Y2)*t,
	}, true
}

func (k RedactionKeyframe) box() RedactionBox {
	return RedactionBox{X1: k.X1, Y1: k.Y1, X2: k.X2, Y2: k.Y2}
}

// drawsAnything reports whether the region is drawn on at least one frame.
func (r *RedactionRegion) drawsAnything() bool {
	if !r.Enabled || len(r.Keyframes) == 0 || r.Start > r.End {
		return false
	}
	hidden := append([]RedactionFrameRange(nil), r.Hidden...)
	sort.Slice(hidden, func(i, j int) bool { return hidden[i].From < hidden[j].From })
	// frame is the first frame not yet known to be hidden.
	frame := r.Start
	for _, h := range hidden {
		if h.From > frame {
			break
		}
		if h.To >= frame {
			if h.To >= r.End {
				return false
			}
			frame = h.To + 1
		}
	}
	return frame <= r.End
}

// DrawsAnything reports whether at least one region is drawn on at least
// one frame.
func (p *RedactionPlan) DrawsAnything() bool {
	if p == nil {
		return false
	}
	for i := range p.Regions {
		if p.Regions[i].drawsAnything() {
			return true
		}
	}
	return false
}

// KeyframeCount returns the number of keyframes over all regions.
func (p *RedactionPlan) KeyframeCount() int {
	if p == nil {
		return 0
	}
	count := 0
	for i := range p.Regions {
		count += len(p.Regions[i].Keyframes)
	}
	return count
}

// Validate checks the plan's shape against DefaultRedactionPlanLimits. A
// plan without regions is valid: a draft may be empty.
func (p *RedactionPlan) Validate() error {
	return p.ValidateWithLimits(DefaultRedactionPlanLimits())
}

// ValidateForRender is Validate plus the requirement that the plan draws
// something. A plan that draws nothing would render an unchanged video.
func (p *RedactionPlan) ValidateForRender() error {
	if err := p.Validate(); err != nil {
		return err
	}
	if !p.DrawsAnything() {
		return fmt.Errorf("redaction plan: no enabled region is drawn on any frame")
	}
	return nil
}

// ValidateWithLimits checks the plan's shape against the given limits. It
// does no I/O.
func (p *RedactionPlan) ValidateWithLimits(limits RedactionPlanLimits) error {
	if p == nil {
		return fmt.Errorf("redaction plan: nil")
	}
	if p.SchemaVersion != RedactionPlanSchemaVersion {
		return fmt.Errorf("redaction plan: unknown schemaVersion %d (want %d)", p.SchemaVersion, RedactionPlanSchemaVersion)
	}
	if len(p.Regions) > limits.MaxRegions {
		return fmt.Errorf("redaction plan: %d regions, at most %d allowed", len(p.Regions), limits.MaxRegions)
	}
	if count := p.KeyframeCount(); count > limits.MaxKeyframes {
		return fmt.Errorf("redaction plan: %d keyframes, at most %d allowed", count, limits.MaxKeyframes)
	}
	ids := make(map[string]struct{}, len(p.Regions))
	for i := range p.Regions {
		region := &p.Regions[i]
		if region.Id == "" {
			return fmt.Errorf("redaction plan: region %d has no id", i)
		}
		if _, duplicate := ids[region.Id]; duplicate {
			return fmt.Errorf("redaction plan: region id %q is used twice", region.Id)
		}
		ids[region.Id] = struct{}{}
		if err := region.validate(limits); err != nil {
			return fmt.Errorf("redaction plan: region %q: %w", region.Id, err)
		}
	}
	return nil
}

func (r *RedactionRegion) validate(limits RedactionPlanLimits) error {
	if length := utf8.RuneCountInString(r.Label); length > limits.MaxLabelLength {
		return fmt.Errorf("label has %d characters, at most %d allowed", length, limits.MaxLabelLength)
	}
	switch r.Origin {
	case RedactionRegionOriginDetected, RedactionRegionOriginManual:
	default:
		return fmt.Errorf("unknown origin %q", r.Origin)
	}
	switch r.Shape {
	case RedactionShapeRectangle:
	default:
		return fmt.Errorf("unknown shape %q", r.Shape)
	}
	switch r.Style.Mode {
	case RedactionModeBlur, RedactionModePixelate, RedactionModeBlack:
	default:
		return fmt.Errorf("unknown style mode %q", r.Style.Mode)
	}
	if r.Start < 0 || r.End < r.Start {
		return fmt.Errorf("invalid frame range %d..%d", r.Start, r.End)
	}
	for _, hidden := range r.Hidden {
		if hidden.From < 0 || hidden.To < hidden.From {
			return fmt.Errorf("invalid hidden range %d..%d", hidden.From, hidden.To)
		}
	}
	if len(r.Keyframes) == 0 {
		return fmt.Errorf("no keyframes")
	}
	previous := int64(-1)
	for _, keyframe := range r.Keyframes {
		if keyframe.Frame < 0 {
			return fmt.Errorf("keyframe on negative frame %d", keyframe.Frame)
		}
		if keyframe.Frame <= previous {
			return fmt.Errorf("keyframes are not sorted by frame, or frame %d is used twice", keyframe.Frame)
		}
		previous = keyframe.Frame
		if !validFraction(keyframe.X1) || !validFraction(keyframe.Y1) || !validFraction(keyframe.X2) || !validFraction(keyframe.Y2) {
			return fmt.Errorf("keyframe on frame %d has a coordinate outside 0..1", keyframe.Frame)
		}
		if keyframe.X1 >= keyframe.X2 || keyframe.Y1 >= keyframe.Y2 {
			return fmt.Errorf("keyframe on frame %d needs x1 < x2 and y1 < y2", keyframe.Frame)
		}
	}
	return nil
}

func validFraction(value float64) bool {
	return !math.IsNaN(value) && value >= 0 && value <= 1
}
