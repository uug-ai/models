package models

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestWorkflowMatchAutomaticTriggerSelection(t *testing.T) {
	at := time.Date(2026, time.September, 29, 9, 10, 11, 123_000_000, time.UTC)
	root := AutomaticTriggerRoot(WorkflowDevice{DeviceKey: "cam-1"}, WorkflowUser{})
	automatic := WorkflowTrigger{Type: WorkflowTriggerAutomatic}
	miss := WorkflowTrigger{Devices: []DeviceKey{{Key: "cam-2"}}}
	manual := WorkflowTrigger{Type: WorkflowTriggerManual, Surfaces: []WorkflowTriggerSurface{WorkflowSurfaceCase}}
	tests := []struct {
		name  string
		w     Workflow
		index int
	}{
		{"first of multiple matches", Workflow{Enabled: true, Triggers: []WorkflowTrigger{automatic, automatic}}, 0},
		{"skip manual and nonmatching", Workflow{Enabled: true, Triggers: []WorkflowTrigger{manual, miss, automatic, automatic}}, 2},
		{"legacy automatic type", Workflow{Enabled: true, Triggers: []WorkflowTrigger{{}}}, 0},
		{"legacy single trigger", Workflow{Enabled: true, Trigger: &automatic}, 0},
		{"canonical list wins over legacy", Workflow{Enabled: true, Trigger: &automatic, Triggers: []WorkflowTrigger{miss}}, -1},
		{"disabled", Workflow{Triggers: []WorkflowTrigger{automatic}}, -1},
		{"no triggers", Workflow{Enabled: true}, -1},
		{"manual only", Workflow{Enabled: true, Triggers: []WorkflowTrigger{manual}}, -1},
		{"no match", Workflow{Enabled: true, Triggers: []WorkflowTrigger{miss}}, -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched := tt.w.AutomaticMatches(root, at)
			snapshot, err := tt.w.MatchAutomaticTrigger(root, at)
			if err != nil {
				t.Fatal(err)
			}
			if matched != (tt.index >= 0) || (snapshot != nil) != matched {
				t.Fatalf("AutomaticMatches = %v, snapshot = %+v, want index %d", matched, snapshot, tt.index)
			}
			if snapshot == nil {
				return
			}
			if snapshot.Index != tt.index || snapshot.EvaluatedAtMs != at.UnixMilli() ||
				snapshot.Trigger.Type != WorkflowTriggerAutomatic {
				t.Fatalf("unexpected trigger selection: %+v", snapshot)
			}
		})
	}
}

func TestWorkflowMatchAutomaticTriggerScheduleAndGroups(t *testing.T) {
	at := time.Date(2026, time.September, 29, 9, 30, 0, 0, time.UTC)
	w := Workflow{
		Enabled: true,
		Triggers: []WorkflowTrigger{
			{WeeklySchedule: []*WeeklySchedule{{
				Day: int(time.Tuesday), Enabled: true, Timezone: "Europe/Brussels",
				Segments: []DayTimeRange{{Start: 8 * 3600, End: 9 * 3600}},
			}}},
			{
				Devices: []DeviceKey{{Key: "cam-1"}}, SiteIds: []string{"site-1"}, GroupIds: []string{"group-1"},
				ConditionMode: ConditionModeAny,
				Conditions: []WorkflowCondition{
					{Path: "device.deviceName", Op: ConditionOpEq, Value: "wrong"},
					{Path: "device.deviceName", Op: ConditionOpEq, Value: "Entrance"},
				},
				WeeklySchedule: []*WeeklySchedule{{
					Day: int(time.Tuesday), Enabled: true, Timezone: "Europe/Brussels",
					Segments: []DayTimeRange{{Start: 11 * 3600, End: 12 * 3600}},
				}},
			},
		},
	}
	root := AutomaticTriggerRoot(WorkflowDevice{
		DeviceKey: "cam-1", DeviceName: "Entrance", SiteIds: []string{"site-1"}, GroupIds: []string{"group-1"},
	}, WorkflowUser{})
	match, err := w.MatchAutomaticTrigger(root, at)
	if err != nil || match == nil || match.Index != 1 || match.EvaluatedAtMs != at.UnixMilli() {
		t.Fatalf("scheduled group match = %+v, error = %v", match, err)
	}
	if match.Trigger.ConditionMode != ConditionModeAny || match.Trigger.WeeklySchedule[0].Timezone != "Europe/Brussels" {
		t.Fatalf("snapshot lost trigger rules: %+v", match.Trigger)
	}
	late, err := w.MatchAutomaticTrigger(root, at.Add(2*time.Hour))
	if err != nil || late != nil {
		t.Fatalf("outside schedule: %+v, %v", late, err)
	}
	wrongSite := AutomaticTriggerRoot(WorkflowDevice{
		DeviceKey: "cam-1", DeviceName: "Entrance", SiteIds: []string{"site-2"}, GroupIds: []string{"group-1"},
	}, WorkflowUser{})
	if match, err := w.MatchAutomaticTrigger(wrongSite, at); err != nil || match != nil {
		t.Fatalf("conditionMode any bypassed mandatory scope: %+v, %v", match, err)
	}
}

func TestWorkflowMatchAutomaticTriggerDetachedSnapshot(t *testing.T) {
	at := time.Date(2026, time.September, 29, 9, 0, 0, 0, time.UTC)
	labels := []any{"person", "car"}
	w := Workflow{Enabled: true, Triggers: []WorkflowTrigger{{
		Devices: []DeviceKey{{Key: "cam-1"}}, SiteIds: []string{"site-1"}, GroupIds: []string{"group-1"},
		Conditions: []WorkflowCondition{{
			Path: "inputs.classify.details", Op: ConditionOpAnyMatch,
			Match: &WorkflowPredicateSet{Conditions: []WorkflowPredicate{
				{Path: "label", Op: ConditionOpIn, Value: labels},
				{Path: "confidence", Op: ConditionOpGte, Value: 0.8},
			}},
		}},
		WeeklySchedule: []*WeeklySchedule{{
			Day: int(time.Tuesday), Enabled: true, Timezone: "UTC",
			Segments: []DayTimeRange{{Start: 8 * 3600, End: 10 * 3600}},
		}},
	}}}
	root := AutomaticTriggerRootWithInputs(WorkflowDevice{
		DeviceKey: "cam-1", SiteIds: []string{"site-1"}, GroupIds: []string{"group-1"},
	}, WorkflowUser{}, map[string]any{"classify": map[string]any{
		"details":   []any{map[string]any{"label": "person", "confidence": 0.9}},
		"unrelated": "not-trigger-provenance",
	}})
	match, err := w.MatchAutomaticTrigger(root, at)
	if err != nil || match == nil {
		t.Fatalf("match = %+v, error = %v", match, err)
	}
	before, err := json.Marshal(match)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(before, []byte("not-trigger-provenance")) {
		t.Fatal("snapshot copied the input envelope")
	}
	source := &w.Triggers[0]
	if source.Type != "" {
		t.Fatal("capturing effective type mutated the definition")
	}
	source.Devices[0].Key = "changed"
	source.SiteIds[0], source.GroupIds[0] = "changed", "changed"
	source.Conditions[0].Match.Conditions[0].Path = "changed"
	labels[0] = "changed"
	source.WeeklySchedule[0].Segments[0].Start = 0
	after, err := json.Marshal(match)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatalf("editing the definition changed historical provenance: %s, %v", after, err)
	}
	match.Trigger.Devices[0].Key = "snapshot-only"
	match.Trigger.WeeklySchedule[0].Segments[0].Start = 1
	if source.Devices[0].Key != "changed" || source.WeeklySchedule[0].Segments[0].Start != 0 {
		t.Fatal("editing the snapshot changed the definition")
	}
}

func TestWorkflowMatchAutomaticTriggerSnapshotsOnlyFirstMatch(t *testing.T) {
	root := AutomaticTriggerRoot(WorkflowDevice{DeviceKey: "cam-1"}, WorkflowUser{})
	unencodable := WorkflowTrigger{Conditions: []WorkflowCondition{{
		Path: "device.deviceKey", Op: ConditionOpExists, Value: func() {},
	}}}
	w := Workflow{Enabled: true, Triggers: []WorkflowTrigger{{}, unencodable}}
	match, err := w.MatchAutomaticTrigger(root, time.Now())
	if err != nil || match == nil || match.Index != 0 {
		t.Fatalf("later trigger affected first-match snapshot: %+v, %v", match, err)
	}
	w.Triggers = []WorkflowTrigger{unencodable, {}}
	if !w.AutomaticMatches(root, time.Now()) {
		t.Fatal("fixture must match using legacy boolean evaluation")
	}
	if match, err := w.MatchAutomaticTrigger(root, time.Now()); err == nil || match != nil {
		t.Fatalf("snapshot encoding failure silently fell through to another match: %+v, %v", match, err)
	}
}

func TestWorkflowRunTriggerMatchSerialization(t *testing.T) {
	match := &WorkflowRunTriggerMatch{
		Index: 0, EvaluatedAtMs: 1_790_672_400_123,
		Trigger: WorkflowTrigger{Type: WorkflowTriggerAutomatic, Devices: []DeviceKey{{Key: "cam-1"}}},
	}
	run := WorkflowRun{Origin: WorkflowOriginAutomatic, TriggerMatch: match}
	wire, err := json.Marshal(run)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(wire, &fields); err != nil {
		t.Fatal(err)
	}
	var detail map[string]json.RawMessage
	if err := json.Unmarshal(fields["triggerMatch"], &detail); err != nil {
		t.Fatal(err)
	}
	if string(detail["index"]) != "0" || string(detail["evaluatedAtMs"]) != "1790672400123" {
		t.Fatalf("required zero index or millisecond precision lost: %s", wire)
	}
	var fromJSON WorkflowRun
	if err := json.Unmarshal(wire, &fromJSON); err != nil || !reflect.DeepEqual(fromJSON.TriggerMatch, match) {
		t.Fatalf("JSON round trip = %+v, error = %v", fromJSON.TriggerMatch, err)
	}
	data, err := bson.Marshal(run)
	if err != nil {
		t.Fatal(err)
	}
	raw := bson.Raw(data).Lookup("triggermatch").Document()
	if raw.Lookup("index").Int32() != 0 || raw.Lookup("evaluatedatms").Int64() != match.EvaluatedAtMs {
		t.Fatalf("incorrect BSON provenance: %s", raw)
	}
	if raw.Lookup("evaluatedAtMs").Type != 0 || bson.Raw(data).Lookup("triggerMatch").Type != 0 {
		t.Fatal("JSON field spelling leaked into BSON")
	}
	var fromBSON WorkflowRun
	if err := bson.Unmarshal(data, &fromBSON); err != nil || !reflect.DeepEqual(fromBSON.TriggerMatch, match) {
		t.Fatalf("BSON round trip = %+v, error = %v", fromBSON.TriggerMatch, err)
	}
}

func TestWorkflowRunTriggerMatchAbsentStaysUnknown(t *testing.T) {
	for _, origin := range []WorkflowRunOrigin{"", WorkflowOriginAutomatic, WorkflowOriginManual} {
		t.Run(string(origin), func(t *testing.T) {
			for _, explicitNull := range []bool{false, true} {
				owner := primitive.NewObjectID()
				document := bson.M{"userid": owner.Hex(), "origin": origin, "start": int64(1_790_672_400_123)}
				if explicitNull {
					document["triggermatch"] = nil
				}
				data, err := bson.Marshal(document)
				if err != nil {
					t.Fatal(err)
				}
				var legacy struct {
					WorkflowRun `bson:",inline"`
					UserID      string `bson:"userid"`
				}
				if err := bson.Unmarshal(data, &legacy); err != nil {
					t.Fatal(err)
				}
				if err := legacy.NormalizeStages(); err != nil {
					t.Fatal(err)
				}
				legacy.PopulateRuntimeFields(time.Now())
				if legacy.TriggerMatch != nil || legacy.UserID != owner.Hex() {
					t.Fatalf("legacy provenance invented or enclosing ownership lost: %+v", legacy)
				}
				wire, err := json.Marshal(legacy.WorkflowRun)
				if err != nil || bytes.Contains(wire, []byte(`"triggerMatch"`)) {
					t.Fatalf("uncaptured trigger must remain absent: %s, %v", wire, err)
				}
				data, err = bson.Marshal(legacy.WorkflowRun)
				if err != nil || bson.Raw(data).Lookup("triggermatch").Type != 0 {
					t.Fatalf("uncaptured trigger persisted: %v", err)
				}
			}
		})
	}
}
