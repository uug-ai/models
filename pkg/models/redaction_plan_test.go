package models

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// redactionPlanCases mirrors testdata/redaction_plan_cases.json, the cases
// every implementation of the drawing rule runs.
type redactionPlanCases struct {
	Cases []struct {
		Name   string          `json:"name"`
		Region RedactionRegion `json:"region"`
		Expect []struct {
			Frame int64     `json:"frame"`
			Box   []float64 `json:"box"`
		} `json:"expect"`
	} `json:"cases"`
}

func TestRedactionRegionBoxAtFollowsTheSharedCases(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "redaction_plan_cases.json"))
	if err != nil {
		t.Fatalf("read cases: %v", err)
	}
	var file redactionPlanCases
	if err := json.Unmarshal(raw, &file); err != nil {
		t.Fatalf("decode cases: %v", err)
	}
	if len(file.Cases) == 0 {
		t.Fatal("no cases in testdata/redaction_plan_cases.json")
	}

	for _, c := range file.Cases {
		t.Run(c.Name, func(t *testing.T) {
			plan := RedactionPlan{SchemaVersion: RedactionPlanSchemaVersion, Regions: []RedactionRegion{c.Region}}
			if err := plan.Validate(); err != nil {
				t.Fatalf("case region is not a valid region: %v", err)
			}
			if len(c.Expect) == 0 {
				t.Fatal("case expects nothing")
			}
			for _, want := range c.Expect {
				got, drawn := c.Region.BoxAt(want.Frame)
				if want.Box == nil {
					if drawn {
						t.Errorf("frame %d: drawn at %+v, want not drawn", want.Frame, got)
					}
					continue
				}
				if len(want.Box) != 4 {
					t.Fatalf("frame %d: expected box needs 4 numbers, has %d", want.Frame, len(want.Box))
				}
				if !drawn {
					t.Errorf("frame %d: not drawn, want %v", want.Frame, want.Box)
					continue
				}
				for i, value := range []float64{got.X1, got.Y1, got.X2, got.Y2} {
					if math.Abs(value-want.Box[i]) > 1e-9 {
						t.Errorf("frame %d: box = %+v, want %v", want.Frame, got, want.Box)
						break
					}
				}
			}
		})
	}
}

func validRedactionRegion(id string) RedactionRegion {
	return RedactionRegion{
		Id:      id,
		Label:   "Face 1",
		Origin:  RedactionRegionOriginDetected,
		Class:   "face",
		Source:  &RedactionRegionSource{RunId: "run-1", TrackId: "track-1"},
		Shape:   RedactionShapeRectangle,
		Style:   RedactionStyle{Mode: RedactionModeBlur},
		Enabled: true,
		Start:   0,
		End:     100,
		Hidden:  []RedactionFrameRange{{From: 10, To: 20}},
		Keyframes: []RedactionKeyframe{
			{Frame: 0, X1: 0.1, Y1: 0.1, X2: 0.2, Y2: 0.2},
			{Frame: 50, X1: 0.3, Y1: 0.3, X2: 0.4, Y2: 0.4, Edited: true},
		},
	}
}

func validRedactionPlan() RedactionPlan {
	return RedactionPlan{
		SchemaVersion: RedactionPlanSchemaVersion,
		Media:         &RedactionPlanMedia{Width: 1920, Height: 1080, Fps: 25, FrameCount: 101},
		Regions:       []RedactionRegion{validRedactionRegion("r1")},
	}
}

func TestRedactionPlanValidate(t *testing.T) {
	tests := []struct {
		name    string
		change  func(*RedactionPlan)
		wantErr string // empty: valid
	}{
		{name: "valid plan", change: func(*RedactionPlan) {}},
		{name: "no regions is a valid empty draft", change: func(p *RedactionPlan) { p.Regions = nil }},
		{name: "one keyframe is enough", change: func(p *RedactionPlan) { p.Regions[0].Keyframes = p.Regions[0].Keyframes[:1] }},
		{name: "coordinates may touch the frame edges", change: func(p *RedactionPlan) {
			p.Regions[0].Keyframes[0] = RedactionKeyframe{Frame: 0, X1: 0, Y1: 0, X2: 1, Y2: 1}
		}},
		{name: "unknown schema version", change: func(p *RedactionPlan) { p.SchemaVersion = 2 }, wantErr: "schemaVersion"},
		{name: "missing schema version", change: func(p *RedactionPlan) { p.SchemaVersion = 0 }, wantErr: "schemaVersion"},
		{name: "region without id", change: func(p *RedactionPlan) { p.Regions[0].Id = "" }, wantErr: "no id"},
		{name: "duplicate region id", change: func(p *RedactionPlan) {
			p.Regions = append(p.Regions, validRedactionRegion("r1"))
		}, wantErr: "used twice"},
		{name: "label too long", change: func(p *RedactionPlan) { p.Regions[0].Label = strings.Repeat("é", 81) }, wantErr: "label"},
		{name: "label of 80 characters counts characters, not bytes", change: func(p *RedactionPlan) {
			p.Regions[0].Label = strings.Repeat("é", 80)
		}},
		{name: "unknown origin", change: func(p *RedactionPlan) { p.Regions[0].Origin = "imported" }, wantErr: "origin"},
		{name: "missing origin", change: func(p *RedactionPlan) { p.Regions[0].Origin = "" }, wantErr: "origin"},
		{name: "unknown shape", change: func(p *RedactionPlan) { p.Regions[0].Shape = "ellipse" }, wantErr: "shape"},
		{name: "missing shape", change: func(p *RedactionPlan) { p.Regions[0].Shape = "" }, wantErr: "shape"},
		{name: "unknown mode", change: func(p *RedactionPlan) { p.Regions[0].Style.Mode = "emoji" }, wantErr: "mode"},
		{name: "missing mode", change: func(p *RedactionPlan) { p.Regions[0].Style.Mode = "" }, wantErr: "mode"},
		{name: "end before start", change: func(p *RedactionPlan) { p.Regions[0].Start, p.Regions[0].End = 10, 9 }, wantErr: "frame range"},
		{name: "negative start", change: func(p *RedactionPlan) { p.Regions[0].Start = -1 }, wantErr: "frame range"},
		{name: "hidden range backwards", change: func(p *RedactionPlan) {
			p.Regions[0].Hidden = []RedactionFrameRange{{From: 5, To: 4}}
		}, wantErr: "hidden range"},
		{name: "no keyframes", change: func(p *RedactionPlan) { p.Regions[0].Keyframes = nil }, wantErr: "no keyframes"},
		{name: "keyframes not sorted", change: func(p *RedactionPlan) {
			p.Regions[0].Keyframes[0].Frame, p.Regions[0].Keyframes[1].Frame = 50, 0
		}, wantErr: "not sorted"},
		{name: "two keyframes on one frame", change: func(p *RedactionPlan) { p.Regions[0].Keyframes[1].Frame = 0 }, wantErr: "used twice"},
		{name: "negative keyframe frame", change: func(p *RedactionPlan) { p.Regions[0].Keyframes[0].Frame = -1 }, wantErr: "negative frame"},
		{name: "coordinate above 1", change: func(p *RedactionPlan) { p.Regions[0].Keyframes[0].X2 = 1.01 }, wantErr: "outside 0..1"},
		{name: "coordinate below 0", change: func(p *RedactionPlan) { p.Regions[0].Keyframes[0].Y1 = -0.01 }, wantErr: "outside 0..1"},
		{name: "coordinate not a number", change: func(p *RedactionPlan) { p.Regions[0].Keyframes[0].X1 = math.NaN() }, wantErr: "outside 0..1"},
		{name: "pixel coordinates are refused", change: func(p *RedactionPlan) {
			p.Regions[0].Keyframes[0] = RedactionKeyframe{Frame: 0, X1: 100, Y1: 100, X2: 300, Y2: 300}
		}, wantErr: "outside 0..1"},
		{name: "box without width", change: func(p *RedactionPlan) { p.Regions[0].Keyframes[0].X2 = p.Regions[0].Keyframes[0].X1 }, wantErr: "x1 < x2"},
		{name: "box upside down", change: func(p *RedactionPlan) {
			p.Regions[0].Keyframes[0].Y1, p.Regions[0].Keyframes[0].Y2 = 0.2, 0.1
		}, wantErr: "y1 < y2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan := validRedactionPlan()
			tt.change(&plan)
			err := plan.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate() error = %v, want nil", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Validate() error = %v, want it to mention %q", err, tt.wantErr)
			}
		})
	}

	var missing *RedactionPlan
	if err := missing.Validate(); err == nil {
		t.Fatal("Validate() on a nil plan returned nil")
	}
}

func TestRedactionLimits(t *testing.T) {
	limits := DefaultRedactionLimits()
	if limits.MaxRegions != 200 || limits.MaxKeyframes != 50000 || limits.MaxLabelLength != 80 {
		t.Fatalf("DefaultRedactionLimits() = %+v", limits)
	}

	plan := RedactionPlan{SchemaVersion: RedactionPlanSchemaVersion}
	for i := 0; i < 3; i++ {
		plan.Regions = append(plan.Regions, validRedactionRegion(string(rune('a'+i))))
	}
	if got := plan.KeyframeCount(); got != 6 {
		t.Fatalf("KeyframeCount() = %d, want 6", got)
	}

	if err := plan.ValidateWithLimits(RedactionLimits{MaxRegions: 3, MaxKeyframes: 6, MaxLabelLength: 80}); err != nil {
		t.Fatalf("plan exactly at the limits: %v", err)
	}
	if err := plan.ValidateWithLimits(RedactionLimits{MaxRegions: 2, MaxKeyframes: 6, MaxLabelLength: 80}); err == nil || !strings.Contains(err.Error(), "regions") {
		t.Fatalf("one region over the limit: error = %v", err)
	}
	if err := plan.ValidateWithLimits(RedactionLimits{MaxRegions: 3, MaxKeyframes: 5, MaxLabelLength: 80}); err == nil || !strings.Contains(err.Error(), "keyframes") {
		t.Fatalf("one keyframe over the limit: error = %v", err)
	}
}

func TestRedactionPlanValidateForRenderNeedsSomethingDrawn(t *testing.T) {
	tests := []struct {
		name   string
		change func(*RedactionPlan)
		draws  bool
	}{
		{name: "an enabled region with visible frames", change: func(*RedactionPlan) {}, draws: true},
		{name: "no regions", change: func(p *RedactionPlan) { p.Regions = nil }},
		{name: "every region disabled", change: func(p *RedactionPlan) { p.Regions[0].Enabled = false }},
		{name: "hidden over the whole time range", change: func(p *RedactionPlan) {
			p.Regions[0].Hidden = []RedactionFrameRange{{From: 0, To: 100}}
		}},
		{name: "hidden over the whole time range by overlapping ranges", change: func(p *RedactionPlan) {
			p.Regions[0].Hidden = []RedactionFrameRange{{From: 40, To: 100}, {From: 0, To: 60}}
		}},
		{name: "hidden over the whole time range by adjacent ranges", change: func(p *RedactionPlan) {
			p.Regions[0].Hidden = []RedactionFrameRange{{From: 51, To: 200}, {From: 0, To: 50}}
		}},
		{name: "only the last frame visible", change: func(p *RedactionPlan) {
			p.Regions[0].Hidden = []RedactionFrameRange{{From: 0, To: 99}}
		}, draws: true},
		{name: "only the first frame visible", change: func(p *RedactionPlan) {
			p.Regions[0].Hidden = []RedactionFrameRange{{From: 1, To: 100}}
		}, draws: true},
		{name: "one frame visible between two hidden ranges", change: func(p *RedactionPlan) {
			p.Regions[0].Hidden = []RedactionFrameRange{{From: 51, To: 100}, {From: 0, To: 49}}
		}, draws: true},
		{name: "a disabled region next to a drawn one", change: func(p *RedactionPlan) {
			p.Regions[0].Enabled = false
			p.Regions = append(p.Regions, validRedactionRegion("r2"))
		}, draws: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan := validRedactionPlan()
			tt.change(&plan)
			if err := plan.Validate(); err != nil {
				t.Fatalf("Validate() error = %v; a plan that draws nothing is still a valid draft", err)
			}
			if got := plan.DrawsAnything(); got != tt.draws {
				t.Fatalf("DrawsAnything() = %v, want %v", got, tt.draws)
			}
			// DrawsAnything must agree with the rule applied frame by frame.
			byFrame := false
			for i := range plan.Regions {
				for frame := int64(0); frame <= 100 && !byFrame; frame++ {
					_, byFrame = plan.Regions[i].BoxAt(frame)
				}
			}
			if byFrame != tt.draws {
				t.Fatalf("BoxAt over every frame draws = %v, DrawsAnything() = %v", byFrame, tt.draws)
			}
			err := plan.ValidateForRender()
			if tt.draws && err != nil {
				t.Fatalf("ValidateForRender() error = %v, want nil", err)
			}
			if !tt.draws && err == nil {
				t.Fatal("ValidateForRender() accepted a plan that draws nothing")
			}
		})
	}
}

func TestRedactionPlanRoundTripsThroughJSONAndBSON(t *testing.T) {
	plan := validRedactionPlan()

	rawJSON, err := json.Marshal(plan)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	for _, field := range []string{`"schemaVersion":1`, `"regions"`, `"origin":"detected"`, `"shape":"rectangle"`,
		`"style":{"mode":"blur"}`, `"enabled":true`, `"hidden":[{"from":10,"to":20}]`, `"source":{"runId":"run-1","trackId":"track-1"}`,
		`"edited":true`, `"frameCount":101`} {
		if !strings.Contains(string(rawJSON), field) {
			t.Errorf("JSON misses %s: %s", field, rawJSON)
		}
	}
	var fromJSON RedactionPlan
	if err := json.Unmarshal(rawJSON, &fromJSON); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	rawBSON, err := bson.Marshal(plan)
	if err != nil {
		t.Fatalf("bson.Marshal() error = %v", err)
	}
	var fromBSON RedactionPlan
	if err := bson.Unmarshal(rawBSON, &fromBSON); err != nil {
		t.Fatalf("bson.Unmarshal() error = %v", err)
	}

	for name, got := range map[string]RedactionPlan{"JSON": fromJSON, "BSON": fromBSON} {
		if err := got.ValidateForRender(); err != nil {
			t.Fatalf("%s round trip is no longer valid: %v", name, err)
		}
		again, err := json.Marshal(got)
		if err != nil {
			t.Fatalf("%s: json.Marshal() error = %v", name, err)
		}
		if string(again) != string(rawJSON) {
			t.Fatalf("%s round trip changed the plan:\n got %s\nwant %s", name, again, rawJSON)
		}
	}

	// A disabled region must stay disabled: "enabled": false may not be dropped.
	plan.Regions[0].Enabled = false
	rawJSON, _ = json.Marshal(plan)
	if !strings.Contains(string(rawJSON), `"enabled":false`) {
		t.Fatalf("JSON drops enabled=false: %s", rawJSON)
	}
	var stored bson.M
	rawBSON, _ = bson.Marshal(plan)
	if err := bson.Unmarshal(rawBSON, &stored); err != nil {
		t.Fatalf("bson.Unmarshal() error = %v", err)
	}
	region := stored["regions"].(bson.A)[0].(bson.M)
	if enabled, ok := region["enabled"].(bool); !ok || enabled {
		t.Fatalf("BSON drops enabled=false: %v", region)
	}
}

func redactionSourceRow() CaseMedia {
	sourceMediaID := primitive.NewObjectID()
	return CaseMedia{
		Role:          CaseMediaRoleSource,
		SourceMediaId: &sourceMediaID,
		VideoFile:     "owner/camera/1700000000_recording.mp4",
		Status:        CaseMediaStatusCompleted,
	}
}

func redactionEditRow(editType CaseMediaEditType) CaseMedia {
	parentID := primitive.NewObjectID()
	return CaseMedia{
		Role:     CaseMediaRoleEdit,
		ParentId: &parentID,
		Action:   CaseMediaActionRedaction,
		EditType: editType,
		Version:  1,
		Status:   CaseMediaStatusQueued,
	}
}

func TestCaseMediaValidateRegionsEdit(t *testing.T) {
	plan := validRedactionPlan()

	valid := redactionEditRow(CaseMediaEditTypeRegions)
	valid.Redaction = &plan
	if err := valid.Validate(); err != nil {
		t.Fatalf("regions edit with a plan: %v", err)
	}

	withoutPlan := redactionEditRow(CaseMediaEditTypeRegions)
	if err := withoutPlan.Validate(); err == nil || !strings.Contains(err.Error(), "requires redaction") {
		t.Fatalf("regions edit without a plan: error = %v", err)
	}

	nothingDrawn := validRedactionPlan()
	nothingDrawn.Regions[0].Enabled = false
	empty := redactionEditRow(CaseMediaEditTypeRegions)
	empty.Redaction = &nothingDrawn
	if err := empty.Validate(); err == nil || !strings.Contains(err.Error(), "no enabled region") {
		t.Fatalf("regions edit whose plan draws nothing: error = %v", err)
	}

	malformed := validRedactionPlan()
	malformed.Regions[0].Keyframes = nil
	broken := redactionEditRow(CaseMediaEditTypeRegions)
	broken.Redaction = &malformed
	if err := broken.Validate(); err == nil || !strings.Contains(err.Error(), "no keyframes") {
		t.Fatalf("regions edit with a malformed plan: error = %v", err)
	}

	// The face_* types keep working without a plan, and may not carry one.
	for _, editType := range []CaseMediaEditType{CaseMediaEditTypeFaceBlur, CaseMediaEditTypeFacePixelate, CaseMediaEditTypeFaceMask} {
		legacy := redactionEditRow(editType)
		if err := legacy.Validate(); err != nil {
			t.Fatalf("%s edit without a plan: %v", editType, err)
		}
		legacy.Redaction = &plan
		if err := legacy.Validate(); err == nil || !strings.Contains(err.Error(), "only allowed") {
			t.Fatalf("%s edit with a plan: error = %v", editType, err)
		}
	}

	trim := redactionEditRow("")
	trim.Action = CaseMediaActionTrim
	trim.Redaction = &plan
	if err := trim.Validate(); err == nil || !strings.Contains(err.Error(), "only allowed") {
		t.Fatalf("trim edit with a plan: error = %v", err)
	}

	unknown := redactionEditRow("face_emoji")
	if err := unknown.Validate(); err == nil {
		t.Fatal("unknown redaction editType was accepted")
	}

	draftOnEdit := redactionEditRow(CaseMediaEditTypeRegions)
	draftOnEdit.Redaction = &plan
	draftOnEdit.RedactionDraft = &RedactionDraft{Plan: plan, Revision: 1}
	if err := draftOnEdit.Validate(); err == nil || !strings.Contains(err.Error(), "redactionDraft") {
		t.Fatalf("edit row with a draft: error = %v", err)
	}
}

func TestCaseMediaValidateRedactionDraftOnSource(t *testing.T) {
	plan := validRedactionPlan()

	source := redactionSourceRow()
	if err := source.Validate(); err != nil {
		t.Fatalf("source row without a draft: %v", err)
	}

	source.RedactionDraft = &RedactionDraft{Plan: plan, Revision: 3}
	if err := source.Validate(); err != nil {
		t.Fatalf("source row with a draft: %v", err)
	}

	// A draft may be empty or draw nothing.
	source.RedactionDraft = &RedactionDraft{Plan: RedactionPlan{SchemaVersion: RedactionPlanSchemaVersion}, Revision: 1}
	if err := source.Validate(); err != nil {
		t.Fatalf("source row with an empty draft: %v", err)
	}

	malformed := validRedactionPlan()
	malformed.Regions[0].Style.Mode = "emoji"
	source.RedactionDraft = &RedactionDraft{Plan: malformed, Revision: 1}
	if err := source.Validate(); err == nil || !strings.Contains(err.Error(), "redactionDraft") {
		t.Fatalf("source row with a malformed draft: error = %v", err)
	}

	withPlan := redactionSourceRow()
	withPlan.Redaction = &plan
	if err := withPlan.Validate(); err == nil || !strings.Contains(err.Error(), "must not set redaction") {
		t.Fatalf("source row with a rendered plan: error = %v", err)
	}
}

func TestCaseMediaStoresPlanAndDraftUnderTheirOwnFields(t *testing.T) {
	plan := validRedactionPlan()
	baseID := primitive.NewObjectID()

	edit := redactionEditRow(CaseMediaEditTypeRegions)
	edit.Redaction = &plan
	source := redactionSourceRow()
	source.RedactionDraft = &RedactionDraft{Plan: plan, Revision: 4, BaseVersionId: &baseID, UpdatedAt: 1764064800, UpdatedBy: "user-1"}

	var storedEdit, storedSource bson.M
	for row, into := range map[*CaseMedia]*bson.M{&edit: &storedEdit, &source: &storedSource} {
		raw, err := bson.Marshal(row)
		if err != nil {
			t.Fatalf("bson.Marshal() error = %v", err)
		}
		if err := bson.Unmarshal(raw, into); err != nil {
			t.Fatalf("bson.Unmarshal() error = %v", err)
		}
	}

	if storedEdit["edit_type"] != "regions" {
		t.Fatalf("edit_type = %v, want regions", storedEdit["edit_type"])
	}
	if _, ok := storedEdit["redaction"].(bson.M); !ok {
		t.Fatalf("edit row stores no redaction document: %v", storedEdit["redaction"])
	}
	if _, ok := storedEdit["redaction_draft"]; ok {
		t.Fatal("edit row stores a redaction_draft")
	}
	if _, ok := storedEdit["params"]; ok {
		t.Fatal("edit row with a plan stores params")
	}

	draft, ok := storedSource["redaction_draft"].(bson.M)
	if !ok {
		t.Fatalf("source row stores no redaction_draft document: %v", storedSource["redaction_draft"])
	}
	if draft["revision"] != int64(4) || draft["base_version_id"] != baseID || draft["updated_by"] != "user-1" {
		t.Fatalf("redaction_draft fields did not store as expected: %v", draft)
	}
	if _, ok := storedSource["redaction"]; ok {
		t.Fatal("source row stores a redaction")
	}

	// Rows without a plan or a draft store neither field.
	var plain bson.M
	raw, _ := bson.Marshal(redactionSourceRow())
	if err := bson.Unmarshal(raw, &plain); err != nil {
		t.Fatalf("bson.Unmarshal() error = %v", err)
	}
	if _, ok := plain["redaction"]; ok {
		t.Fatal("row without a plan stores redaction")
	}
	if _, ok := plain["redaction_draft"]; ok {
		t.Fatal("row without a draft stores redaction_draft")
	}
}
