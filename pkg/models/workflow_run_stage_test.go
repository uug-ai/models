package models

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func runStageJSON(t *testing.T, value any) any {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal JSON: %v", err)
	}
	var decoded any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("decode JSON: %v", err)
	}
	return decoded
}

func runStageBSON(t *testing.T, value any) bson.M {
	t.Helper()
	data, err := bson.Marshal(value)
	if err != nil {
		t.Fatalf("marshal BSON: %v", err)
	}
	var document bson.M
	if err := bson.Unmarshal(data, &document); err != nil {
		t.Fatalf("decode BSON: %v", err)
	}
	return document
}

func runStageAssertJSON(t *testing.T, got any, wantJSON string) {
	t.Helper()
	var want any
	if err := json.Unmarshal([]byte(wantJSON), &want); err != nil {
		t.Fatalf("decode expected JSON: %v", err)
	}
	if actual := runStageJSON(t, got); !reflect.DeepEqual(actual, want) {
		t.Fatalf("JSON = %#v, want %#v", actual, want)
	}
}

func runStageLegacyFacts(operation string) WorkflowRunStageExecution {
	return WorkflowRunStageExecution{
		Operation:                operation,
		Name:                     "Legacy " + operation,
		Dependencies:             []string{"not-an-execution-rule"},
		DispatchAttempts:         3,
		FirstDispatchAttemptAtMs: 1_100,
		LastDispatchAttemptAtMs:  1_200,
		DispatchedAtMs:           1_300,
		ResolvedAtMs:             1_900,
		LastDispatchErrorCode:    "queue_unavailable",
	}
}

func runStageDetails() *WorkflowRunStageExecutionDetails {
	return &WorkflowRunStageExecutionDetails{
		DispatchAttempts:         3,
		FirstDispatchAttemptAtMs: 1_100,
		LastDispatchAttemptAtMs:  1_200,
		DispatchedAtMs:           1_300,
		ResolvedAtMs:             1_900,
		LastDispatchErrorCode:    "queue_unavailable",
	}
}

func TestWorkflowRunStageSerialization(t *testing.T) {
	execution := runStageDetails()
	execution.State = WorkflowRunStageStateResolved
	execution.DurationMs = 600
	stage := WorkflowRunStage{
		Operation: "anpr", Name: "Read plates", Queue: "tenant.anpr",
		Dispatch: DispatchConditional, NeedsMode: NeedsModeAll,
		Needs: []StageDependency{{
			Operation: "classify",
			Condition: &StageCondition{Path: "device.deviceName", Op: ConditionOpEq, Value: "Entrance"},
		}},
		Execution: execution,
	}
	runStageAssertJSON(t, stage, `{
		"operation":"anpr","name":"Read plates","queue":"tenant.anpr",
		"dispatch":"conditional","needsMode":"all",
		"needs":[{"operation":"classify","condition":{"path":"device.deviceName","op":"eq","value":"Entrance"}}],
		"execution":{"dispatchAttempts":3,"firstDispatchAttemptAtMs":1100,
			"lastDispatchAttemptAtMs":1200,"dispatchedAtMs":1300,"resolvedAtMs":1900,
			"lastDispatchErrorCode":"queue_unavailable","state":"resolved","durationMs":600}
	}`)
	wantBSON := bson.M{
		"operation": "anpr", "name": "Read plates", "queue": "tenant.anpr",
		"dispatch": "conditional", "needsMode": "all",
		"needs": bson.A{bson.M{
			"operation": "classify",
			"condition": bson.M{"path": "device.deviceName", "op": "eq", "value": "Entrance"},
		}},
		"execution": bson.M{
			"dispatchattempts": int32(3), "firstdispatchattemptatms": int64(1_100),
			"lastdispatchattemptatms": int64(1_200), "dispatchedatms": int64(1_300),
			"resolvedatms": int64(1_900), "lastdispatcherrorcode": "queue_unavailable",
		},
	}
	if got := runStageBSON(t, stage); !reflect.DeepEqual(got, wantBSON) {
		t.Fatalf("BSON = %#v, want %#v", got, wantBSON)
	}

	t.Run("run round trips nested execution", func(t *testing.T) {
		run := WorkflowRun{Stages: []WorkflowRunStage{stage}}
		wire := runStageJSON(t, run).(map[string]any)
		if _, exists := wire["stageExecutions"]; exists {
			t.Fatal("canonical stages unexpectedly emitted legacy stageExecutions")
		}
		if got := wire["stages"]; !reflect.DeepEqual(got, runStageJSON(t, []WorkflowRunStage{stage})) {
			t.Fatalf("run stages = %#v", got)
		}
		data, err := bson.Marshal(run)
		if err != nil {
			t.Fatal(err)
		}
		var decoded WorkflowRun
		if err := bson.Unmarshal(data, &decoded); err != nil {
			t.Fatal(err)
		}
		if len(decoded.Stages) != 1 || !reflect.DeepEqual(decoded.Stages[0].Execution, runStageDetails()) {
			t.Fatalf("persisted execution = %#v", decoded.Stages)
		}
		if _, exists := runStageBSON(t, run)["stageexecutions"]; exists {
			t.Fatal("canonical stages unexpectedly persisted legacy stageexecutions")
		}
		wireData, err := json.Marshal(run)
		if err != nil {
			t.Fatal(err)
		}
		var fromJSON WorkflowRun
		if err := json.Unmarshal(wireData, &fromJSON); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(fromJSON.Stages, run.Stages) {
			t.Fatalf("JSON round trip = %#v, want %#v", fromJSON.Stages, run.Stages)
		}
	})

	t.Run("nil execution omitted but empty execution is authoritative", func(t *testing.T) {
		absent := WorkflowRunStage{Operation: "anpr"}
		empty := WorkflowRunStage{Operation: "anpr", Execution: &WorkflowRunStageExecutionDetails{}}
		runStageAssertJSON(t, absent, `{"operation":"anpr"}`)
		runStageAssertJSON(t, empty, `{"operation":"anpr","execution":{}}`)
		if got := runStageBSON(t, absent); !reflect.DeepEqual(got, bson.M{"operation": "anpr"}) {
			t.Fatalf("nil execution BSON = %#v", got)
		}
		if got := runStageBSON(t, empty); !reflect.DeepEqual(got, bson.M{"operation": "anpr", "execution": bson.M{}}) {
			t.Fatalf("empty execution BSON = %#v", got)
		}
	})
}

func TestWorkflowRunStageSnapshotPresence(t *testing.T) {
	for _, tc := range []struct {
		name   string
		stages []WorkflowRunStage
	}{
		{name: "absent"},
		{name: "explicit empty", stages: []WorkflowRunStage{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			run := WorkflowRun{Stages: tc.stages}
			for _, value := range []any{run, &run} {
				wire := runStageJSON(t, value).(map[string]any)
				got, exists := wire["stages"]
				if tc.stages == nil {
					if exists {
						t.Fatalf("nil snapshot emitted stages = %#v", got)
					}
				} else if !exists || !reflect.DeepEqual(got, []any{}) {
					t.Fatalf("explicit empty snapshot = %#v, present = %t", got, exists)
				}
			}
			data, err := bson.Marshal(run)
			if err != nil {
				t.Fatal(err)
			}
			wantType := bsontype.Null
			if tc.stages != nil {
				wantType = bsontype.Array
			}
			if got := bson.Raw(data).Lookup("stages").Type; got != wantType {
				t.Fatalf("BSON stages type = %v, want %v", got, wantType)
			}
			var decoded WorkflowRun
			if err := bson.Unmarshal(data, &decoded); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(decoded.Stages, tc.stages) {
				t.Fatalf("BSON snapshot = %#v, want %#v", decoded.Stages, tc.stages)
			}
			wireData, err := json.Marshal(run)
			if err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(wireData, &decoded); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(decoded.Stages, tc.stages) {
				t.Fatalf("JSON snapshot = %#v, want %#v", decoded.Stages, tc.stages)
			}
		})
	}

	t.Run("BSON inline tenant wrapper decodes all sibling fields", func(t *testing.T) {
		type tenantRun struct {
			WorkflowRun `bson:",inline"`
			TenantId    primitive.ObjectID `bson:"tenantid"`
			Scope       string             `bson:"scope"`
		}
		tenant := primitive.NewObjectID()
		for _, tc := range []struct {
			name    string
			value   any
			present bool
			wantNil bool
		}{
			{name: "absent", wantNil: true},
			{name: "null", present: true, wantNil: true},
			{name: "empty", value: bson.A{}, present: true},
			{name: "populated", value: bson.A{bson.M{"operation": "anpr", "execution": bson.M{}}}, present: true},
		} {
			t.Run(tc.name, func(t *testing.T) {
				doc := bson.M{"tenantid": tenant, "scope": "project-1", "key": "recording-1"}
				if tc.present {
					doc["stages"] = tc.value
				}
				data, err := bson.Marshal(doc)
				if err != nil {
					t.Fatal(err)
				}
				var decoded tenantRun
				if err := bson.Unmarshal(data, &decoded); err != nil {
					t.Fatal(err)
				}
				if decoded.TenantId != tenant || decoded.Scope != "project-1" || decoded.Key != "recording-1" {
					t.Fatalf("inline wrapper lost ownership or run fields: %#v", decoded)
				}
				if (decoded.Stages == nil) != tc.wantNil {
					t.Fatalf("decoded stages = %#v, want nil = %t", decoded.Stages, tc.wantNil)
				}
				if tc.name == "populated" && (len(decoded.Stages) != 1 || decoded.Stages[0].Execution == nil) {
					t.Fatalf("nested execution was lost: %#v", decoded.Stages)
				}
				roundTrip := runStageBSON(t, decoded)
				if roundTrip["tenantid"] != tenant || roundTrip["scope"] != "project-1" {
					t.Fatalf("inline wrapper round trip lost tenant fields: %#v", roundTrip)
				}
			})
		}
	})
}

func runStageSource() []WorkflowStage {
	return []WorkflowStage{{
		Id: primitive.NewObjectID(), Operation: "notify", Name: "Notify",
		Queue: "tenant.notify", Dispatch: DispatchConditional, NeedsMode: NeedsModeAll,
		Needs: []StageDependency{
			{
				Operation: "classify",
				Condition: &StageCondition{
					Path: "device.deviceName", Op: ConditionOpIn, Value: bson.A{"Entrance", "Exit"},
				},
			},
			{
				Operation: "anpr", ConditionMode: ConditionModeAny,
				Conditions: []WorkflowCondition{
					{Path: "device.siteIds", Op: ConditionOpIn, Value: bson.A{"north", "south"}},
					{
						Path: "inputs.classify.objects", Op: ConditionOpAnyMatch,
						Match: &WorkflowPredicateSet{
							ConditionMode: ConditionModeAll,
							Conditions: []WorkflowPredicate{
								{Path: "label", Op: ConditionOpIn, Value: bson.A{"person", "vehicle"}},
								{Path: "confidence", Op: ConditionOpGte, Value: 0.9},
							},
						},
					},
				},
			},
		},
		Description: "Do not snapshot catalog metadata",
		Repository:  "private/notify", Tag: "latest", PullPolicy: "Always", Replicas: 4,
		LogLevel: "debug", Env: map[string]string{"ACCESS_TOKEN": "never-copy-this-secret"},
		Resources: &StageResources{Limits: &StageResourceList{CPU: "2", Memory: "1Gi"}},
		Params:    []StageParam{{Name: "credential", Type: StageParamString, Default: "never-copy-this-secret"}},
		Inputs:    []StagePort{{Name: "media"}}, Outputs: []StagePort{{Name: "result"}},
	}}
}

func runStageMutateNeeds(needs []StageDependency) {
	needs[0].Operation = "changed-gate"
	needs[0].Condition.Path = "device.deviceKey"
	needs[0].Condition.Value.(bson.A)[0] = "changed-legacy-value"
	needs[1].ConditionMode = ConditionModeAll
	needs[1].Conditions[0].Path = "device.groupIds"
	needs[1].Conditions[0].Value.(bson.A)[0] = "changed-plural-value"
	needs[1].Conditions[1].Match.ConditionMode = ConditionModeAny
	needs[1].Conditions[1].Match.Conditions[0].Path = "category"
	needs[1].Conditions[1].Match.Conditions[0].Value.(bson.A)[0] = "changed-predicate-value"
}

func TestNewWorkflowRunStagesRoutingSnapshot(t *testing.T) {
	source := runStageSource()
	before := runStageJSON(t, source)
	stages, err := NewWorkflowRunStages(source)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(runStageJSON(t, source), before) {
		t.Fatal("constructor mutated the source catalog")
	}
	if len(stages) != 1 || stages[0].Execution == nil || *stages[0].Execution != (WorkflowRunStageExecutionDetails{}) {
		t.Fatalf("constructor must create one empty execution: %#v", stages)
	}
	want := WorkflowRunStage{
		Operation: source[0].Operation, Name: source[0].Name, Queue: source[0].Queue,
		Dispatch: source[0].Dispatch, Needs: source[0].Needs, NeedsMode: source[0].NeedsMode,
		Execution: &WorkflowRunStageExecutionDetails{},
	}
	if !reflect.DeepEqual(runStageJSON(t, stages[0]), runStageJSON(t, want)) {
		t.Fatalf("routing snapshot = %#v, want %#v", stages[0], want)
	}
	for _, encoded := range []any{runStageJSON(t, stages[0]), runStageBSON(t, stages[0])} {
		data, err := json.Marshal(encoded)
		if err != nil {
			t.Fatal(err)
		}
		for _, forbidden := range []string{
			`"id"`, `"_id"`, `"description"`, `"repository"`, `"tag"`, `"pullPolicy"`,
			`"replicas"`, `"logLevel"`, `"env"`, `"resources"`, `"params"`, `"inputs"`, `"outputs"`,
			"never-copy-this-secret",
		} {
			if strings.Contains(string(data), forbidden) {
				t.Errorf("snapshot contains catalog/deployment field %s: %s", forbidden, data)
			}
		}
	}
	snapshotBefore := runStageJSON(t, stages)
	runStageMutateNeeds(source[0].Needs)
	source[0].Name = "changed catalog name"
	if !reflect.DeepEqual(runStageJSON(t, stages), snapshotBefore) {
		t.Fatal("snapshot aliases source conditions, predicate groups, or BSON values")
	}
	sourceBefore := runStageJSON(t, source)
	runStageMutateNeeds(stages[0].Needs)
	stages[0].Execution.DispatchAttempts = 9
	if !reflect.DeepEqual(runStageJSON(t, source), sourceBefore) {
		t.Fatal("mutating snapshot changed the source catalog")
	}
}

func TestWorkflowRunStageRoutingProjection(t *testing.T) {
	source := runStageSource()
	stages, err := NewWorkflowRunStages(source)
	if err != nil {
		t.Fatal(err)
	}
	stages[0].Execution = runStageDetails()
	run := WorkflowRun{Stages: stages, StageExecutions: []WorkflowRunStageExecution{runStageLegacyFacts("notify")}}
	before := runStageJSON(t, run)
	routing, err := run.RoutingStages()
	if err != nil {
		t.Fatal(err)
	}
	want := []WorkflowStage{{
		Operation: source[0].Operation, Name: source[0].Name, Queue: source[0].Queue,
		Dispatch: source[0].Dispatch, Needs: source[0].Needs, NeedsMode: source[0].NeedsMode,
	}}
	if !reflect.DeepEqual(runStageJSON(t, routing), runStageJSON(t, want)) {
		t.Fatalf("routing projection = %#v, want %#v", routing, want)
	}
	if !reflect.DeepEqual(runStageJSON(t, run), before) {
		t.Fatal("routing projection mutated or normalized its receiver")
	}
	runStageMutateNeeds(routing[0].Needs)
	routing[0].Name = "changed projection"
	if !reflect.DeepEqual(runStageJSON(t, run), before) {
		t.Fatal("routing projection aliases run conditions or values")
	}
	routingBefore := runStageJSON(t, routing)
	runStageMutateNeeds(run.Stages[0].Needs)
	if !reflect.DeepEqual(runStageJSON(t, routing), routingBefore) {
		t.Fatal("run mutation changed a detached routing projection")
	}
}

func TestWorkflowRunStageSnapshotHelpersPreservePresence(t *testing.T) {
	for _, tc := range []struct {
		name   string
		source []WorkflowStage
	}{
		{name: "nil"},
		{name: "empty", source: []WorkflowStage{}},
		{name: "routing unchanged", source: []WorkflowStage{{Operation: "anpr"}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stages, err := NewWorkflowRunStages(tc.source)
			if err != nil {
				t.Fatal(err)
			}
			if (stages == nil) != (tc.source == nil) || len(stages) != len(tc.source) {
				t.Fatalf("snapshot = %#v, source = %#v", stages, tc.source)
			}
			run := WorkflowRun{Stages: stages}
			projected, err := run.RoutingStages()
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(projected, tc.source) {
				t.Fatalf("routing = %#v, want %#v; defaults and queues belong to the caller", projected, tc.source)
			}
		})
	}
}

func TestWorkflowRunStageSnapshotBSONCopyErrors(t *testing.T) {
	value := make(chan int)
	needs := []StageDependency{{
		Operation: "classify",
		Condition: &StageCondition{Path: "device.deviceName", Op: ConditionOpEq, Value: value},
	}}
	source := []WorkflowStage{{Operation: "notify", Needs: needs}}
	if _, err := NewWorkflowRunStages(source); err == nil {
		t.Fatal("constructor silently accepted an unencodable condition value")
	}
	if source[0].Needs[0].Condition.Value != value {
		t.Fatal("failed constructor mutated the source")
	}
	run := WorkflowRun{Stages: []WorkflowRunStage{{Operation: "notify", Needs: needs}}}
	if _, err := run.RoutingStages(); err == nil {
		t.Fatal("routing projection silently accepted an unencodable condition value")
	}
	if run.Stages[0].Execution != nil || run.Stages[0].Needs[0].Condition.Value != value {
		t.Fatal("failed routing projection mutated the source")
	}
}

func TestWorkflowRunNormalizeStagesLegacyFormats(t *testing.T) {
	legacyJSON := `{
		"stages":[
			{"operation":"anpr","name":"Canonical ANPR","queue":"tenant.anpr","dispatch":"conditional",
			 "needsMode":"all","needs":[{"operation":"classify",
			 "condition":{"path":"device.deviceName","op":"eq","value":"Entrance"}}]},
			{"operation":"notify","dispatch":"always"}
		],
		"stageExecutions":[
			{"operation":"notify","name":"Legacy notify","dependencies":["invented"],"dispatchAttempts":1},
			{"operation":"orphan","name":"Config-only stage","dependencies":["legacy-gate"],"resolvedAtMs":1900},
			{"operation":"anpr","name":"Legacy ANPR","dependencies":["invented"],"dispatchAttempts":3,
			 "firstDispatchAttemptAtMs":1100,"lastDispatchAttemptAtMs":1200,
			 "dispatchedAtMs":1300,"resolvedAtMs":1900,"lastDispatchErrorCode":"queue_unavailable"}
		]}`
	legacyBSON := bson.M{
		"stages": bson.A{
			bson.M{
				"operation": "anpr", "name": "Canonical ANPR", "queue": "tenant.anpr", "dispatch": "conditional",
				"needsMode": "all", "needs": bson.A{bson.M{
					"operation": "classify",
					"condition": bson.M{"path": "device.deviceName", "op": "eq", "value": "Entrance"},
				}},
			},
			bson.M{"operation": "notify", "dispatch": "always"},
		},
		"stageexecutions": bson.A{
			bson.M{"operation": "notify", "name": "Legacy notify", "dependencies": bson.A{"invented"}, "dispatchattempts": 1},
			bson.M{"operation": "orphan", "name": "Config-only stage", "dependencies": bson.A{"legacy-gate"}, "resolvedatms": int64(1_900)},
			bson.M{
				"operation": "anpr", "name": "Legacy ANPR", "dependencies": bson.A{"invented"},
				"dispatchattempts": 3, "firstdispatchattemptatms": int64(1_100),
				"lastdispatchattemptatms": int64(1_200), "dispatchedatms": int64(1_300),
				"resolvedatms": int64(1_900), "lastdispatcherrorcode": "queue_unavailable",
			},
		},
	}
	for _, format := range []string{"JSON", "BSON"} {
		t.Run(format, func(t *testing.T) {
			var run WorkflowRun
			if format == "JSON" {
				if err := json.Unmarshal([]byte(legacyJSON), &run); err != nil {
					t.Fatal(err)
				}
			} else {
				data, err := bson.Marshal(legacyBSON)
				if err != nil {
					t.Fatal(err)
				}
				if err := bson.Unmarshal(data, &run); err != nil {
					t.Fatal(err)
				}
			}
			if len(run.Stages) != 2 || run.Stages[0].Execution != nil || len(run.StageExecutions) != 3 {
				t.Fatalf("decode must leave split storage unchanged: %#v", run)
			}
			before := runStageJSON(t, run)
			routingBefore := runStageJSON(t, run.Stages[0])
			orphanBefore := run.StageExecutions[1]
			wire := runStageJSON(t, run).(map[string]any)
			if len(wire["stageExecutions"].([]any)) != 3 {
				t.Fatal("JSON marshaling removed legacy summaries")
			}
			if _, exists := wire["stages"].([]any)[0].(map[string]any)["execution"]; exists {
				t.Fatal("JSON marshaling normalized a copy without being asked")
			}
			document := runStageBSON(t, run)
			if len(document["stageexecutions"].(bson.A)) != 3 {
				t.Fatal("BSON marshaling removed legacy summaries")
			}
			if _, exists := document["stages"].(bson.A)[0].(bson.M)["execution"]; exists {
				t.Fatal("BSON marshaling normalized a copy without being asked")
			}
			if !reflect.DeepEqual(runStageJSON(t, run), before) || run.Stages[0].Execution != nil {
				t.Fatal("marshaling automatically normalized legacy stages")
			}
			if err := run.NormalizeStages(); err != nil {
				t.Fatal(err)
			}
			if len(run.Stages) != 2 || run.Stages[0].Operation != "anpr" || run.Stages[1].Operation != "notify" {
				t.Fatalf("normalization reordered or invented stages: %#v", run.Stages)
			}
			if !reflect.DeepEqual(run.Stages[0].Execution, runStageDetails()) {
				t.Fatalf("joined lifecycle = %#v, want %#v", run.Stages[0].Execution, runStageDetails())
			}
			withoutExecution := run.Stages[0]
			withoutExecution.Execution = nil
			if !reflect.DeepEqual(runStageJSON(t, withoutExecution), routingBefore) {
				t.Fatal("normalization overwrote canonical routing or name")
			}
			if run.Stages[1].Name != "Legacy notify" || run.Stages[1].Execution == nil ||
				run.Stages[1].Execution.DispatchAttempts != 1 || len(run.Stages[1].Needs) != 0 {
				t.Fatalf("legacy name/facts join invented execution rules: %#v", run.Stages[1])
			}
			if !reflect.DeepEqual(run.StageExecutions, []WorkflowRunStageExecution{orphanBefore}) {
				t.Fatalf("unmatched summaries were discarded or reordered: %#v", run.StageExecutions)
			}
			need := run.Stages[0].Needs[0]
			root := map[string]any{"device": map[string]any{"deviceName": "Entrance"}}
			if !need.Matches(root, map[string]bool{"classify": true}) || need.Matches(root, nil) {
				t.Fatal("normalization coupled condition source to readiness gate")
			}
			normalized := runStageJSON(t, run)
			if err := run.NormalizeStages(); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(runStageJSON(t, run), normalized) {
				t.Fatal("normalization is not idempotent")
			}
		})
	}
}

func TestWorkflowRunNormalizeStagesNestedExecutionWins(t *testing.T) {
	for _, tc := range []struct {
		name      string
		execution WorkflowRunStageExecutionDetails
	}{
		{name: "empty object"},
		{name: "partial facts", execution: WorkflowRunStageExecutionDetails{DispatchedAtMs: 1_500}},
		{name: "derived values", execution: WorkflowRunStageExecutionDetails{State: WorkflowRunStageStateWaiting}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			run := WorkflowRun{
				Stages: []WorkflowRunStage{{
					Operation: "anpr", Name: "Canonical name", Execution: &tc.execution,
				}},
				StageExecutions: []WorkflowRunStageExecution{runStageLegacyFacts("anpr")},
			}
			want := tc.execution
			if err := run.NormalizeStages(); err != nil {
				t.Fatal(err)
			}
			if run.Stages[0].Execution == nil || *run.Stages[0].Execution != want {
				t.Fatalf("nested facts coalesced with legacy zeros/blanks: %#v, want %#v", run.Stages[0].Execution, want)
			}
			if run.Stages[0].Name != "Canonical name" || len(run.Stages[0].Needs) != 0 || len(run.StageExecutions) != 0 {
				t.Fatalf("canonical identity/routing or matched legacy removal violated: %#v", run)
			}
		})
	}
	t.Run("legacy name fills independently of authoritative empty execution", func(t *testing.T) {
		run := WorkflowRun{
			Stages:          []WorkflowRunStage{{Operation: "anpr", Execution: &WorkflowRunStageExecutionDetails{}}},
			StageExecutions: []WorkflowRunStageExecution{runStageLegacyFacts("anpr")},
		}
		if err := run.NormalizeStages(); err != nil {
			t.Fatal(err)
		}
		if run.Stages[0].Name != "Legacy anpr" || *run.Stages[0].Execution != (WorkflowRunStageExecutionDetails{}) {
			t.Fatalf("name fallback changed authoritative empty execution: %#v", run.Stages[0])
		}
	})
}

func TestWorkflowRunNormalizeStagesDoesNotInventRouting(t *testing.T) {
	for _, stages := range []struct {
		name  string
		value []WorkflowRunStage
	}{{name: "nil stages"}, {name: "empty stages", value: []WorkflowRunStage{}}} {
		for _, legacy := range []struct {
			name  string
			value []WorkflowRunStageExecution
		}{
			{name: "nil legacy"}, {name: "empty legacy", value: []WorkflowRunStageExecution{}},
			{name: "orphan legacy", value: []WorkflowRunStageExecution{runStageLegacyFacts("config-only")}},
		} {
			t.Run(stages.name+"/"+legacy.name, func(t *testing.T) {
				run := WorkflowRun{Stages: stages.value, StageExecutions: legacy.value}
				if err := run.NormalizeStages(); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(run.Stages, stages.value) || !reflect.DeepEqual(run.StageExecutions, legacy.value) {
					t.Fatalf("normalization changed routing/presence: stages = %#v, legacy = %#v", run.Stages, run.StageExecutions)
				}
			})
		}
	}
	run := WorkflowRun{
		Stages:          []WorkflowRunStage{{Operation: "anpr", Name: "Same name"}},
		StageExecutions: []WorkflowRunStageExecution{{Operation: "other", Name: "Same name"}},
	}
	if err := run.NormalizeStages(); err != nil {
		t.Fatal(err)
	}
	if run.Stages[0].Execution != nil || len(run.StageExecutions) != 1 {
		t.Fatal("normalization joined by name rather than operation")
	}
}

func TestWorkflowRunNormalizeStagesErrorsAreAtomic(t *testing.T) {
	for _, tc := range []struct {
		name   string
		stages []WorkflowRunStage
		legacy []WorkflowRunStageExecution
	}{
		{name: "empty routing operation", stages: []WorkflowRunStage{{Operation: "anpr"}, {Name: "invalid"}}},
		{name: "duplicate routing operation", stages: []WorkflowRunStage{{Operation: "anpr"}, {Operation: "anpr"}}},
		{name: "empty legacy operation", legacy: []WorkflowRunStageExecution{runStageLegacyFacts("anpr"), {Name: "invalid"}}},
		{name: "duplicate matched legacy operation", legacy: []WorkflowRunStageExecution{runStageLegacyFacts("anpr"), runStageLegacyFacts("anpr")}},
		{name: "duplicate orphan legacy operation", legacy: []WorkflowRunStageExecution{runStageLegacyFacts("anpr"), runStageLegacyFacts("orphan"), runStageLegacyFacts("orphan")}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			run := WorkflowRun{
				Stages:          []WorkflowRunStage{{Operation: "anpr"}},
				StageExecutions: []WorkflowRunStageExecution{runStageLegacyFacts("anpr")},
			}
			if tc.stages != nil {
				run.Stages = tc.stages
			}
			if tc.legacy != nil {
				run.StageExecutions = tc.legacy
			}
			beforeJSON, beforeBSON := runStageJSON(t, run), runStageBSON(t, run)
			for _, normalize := range []func() error{
				run.NormalizeStages,
				func() error { _, err := run.StageExecutionSummaries(); return err },
			} {
				if err := normalize(); err == nil || strings.TrimSpace(err.Error()) == "" {
					t.Fatal("invalid stage identity must return an explicit error")
				}
				if !reflect.DeepEqual(runStageJSON(t, run), beforeJSON) || !reflect.DeepEqual(runStageBSON(t, run), beforeBSON) {
					t.Fatal("failed normalization/projection partially mutated the receiver")
				}
			}
		})
	}
}

func TestWorkflowRunStageExecutionSummaries(t *testing.T) {
	orphan := runStageLegacyFacts("config-only")
	orphan.State = WorkflowRunStageStateResolved
	orphan.DurationMs = 600
	run := WorkflowRun{
		Stages: []WorkflowRunStage{
			{
				Operation: "notify", Name: "Canonical notify", Execution: &WorkflowRunStageExecutionDetails{},
				Needs: []StageDependency{
					{
						Operation: "classify",
						Condition: &StageCondition{Path: "results.anpr.plate", Op: ConditionOpExists},
					},
					{Condition: &StageCondition{Path: "device.deviceName", Op: ConditionOpEq, Value: "Entrance"}},
				},
			},
			{Operation: "anpr", Needs: []StageDependency{{Operation: "classify"}}},
			{Operation: "routing-only", Name: "No execution yet"},
			{
				Operation: "nested", Name: "Nested facts",
				Execution: &WorkflowRunStageExecutionDetails{
					DispatchedAtMs: 1_300, ResolvedAtMs: 1_900, State: WorkflowRunStageStateResolved, DurationMs: 600,
				},
			},
		},
		StageExecutions: []WorkflowRunStageExecution{
			orphan, runStageLegacyFacts("anpr"), runStageLegacyFacts("notify"),
		},
	}
	beforeJSON, beforeBSON := runStageJSON(t, run), runStageBSON(t, run)
	summaries, err := run.StageExecutionSummaries()
	if err != nil {
		t.Fatal(err)
	}
	anpr := runStageLegacyFacts("anpr")
	anpr.Dependencies = []string{"classify"}
	want := []WorkflowRunStageExecution{
		{Operation: "notify", Name: "Canonical notify", Dependencies: []string{"classify"}},
		anpr,
		{Operation: "routing-only", Name: "No execution yet"},
		{Operation: "nested", Name: "Nested facts", DispatchedAtMs: 1_300, ResolvedAtMs: 1_900, State: WorkflowRunStageStateResolved, DurationMs: 600},
		orphan,
	}
	if !reflect.DeepEqual(summaries, want) {
		t.Fatalf("summaries = %#v, want %#v", summaries, want)
	}
	if !reflect.DeepEqual(runStageJSON(t, run), beforeJSON) || !reflect.DeepEqual(runStageBSON(t, run), beforeBSON) {
		t.Fatal("summary projection normalized or mutated the receiver")
	}
	summaries[0].Dependencies[0] = "changed"
	summaries[1].Name = "changed"
	summaries[len(summaries)-1].Dependencies[0] = "changed orphan dependency"
	if !reflect.DeepEqual(runStageJSON(t, run), beforeJSON) {
		t.Fatal("summary result aliases the receiver")
	}
}

func TestWorkflowRunNestedStageRuntimeFields(t *testing.T) {
	for _, tc := range []struct {
		name     string
		end      int64
		endedMs  int64
		nowMs    int64
		facts    WorkflowRunStageExecutionDetails
		state    WorkflowRunStageState
		duration int64
	}{
		{name: "waiting", nowMs: 4_000, state: WorkflowRunStageStateWaiting},
		{name: "retrying", nowMs: 4_000, facts: WorkflowRunStageExecutionDetails{DispatchAttempts: 2, FirstDispatchAttemptAtMs: 1_100}, state: WorkflowRunStageStateRetrying},
		{name: "dispatched", nowMs: 4_000, facts: WorkflowRunStageExecutionDetails{DispatchAttempts: 2, DispatchedAtMs: 1_300}, state: WorkflowRunStageStateDispatched, duration: 2_700},
		{name: "resolved", nowMs: 4_000, facts: WorkflowRunStageExecutionDetails{DispatchedAtMs: 1_300, ResolvedAtMs: 1_900}, state: WorkflowRunStageStateResolved, duration: 600},
		{name: "resolved after end", end: 3, nowMs: 4_000, facts: WorkflowRunStageExecutionDetails{DispatchedAtMs: 1_300, ResolvedAtMs: 1_900}, state: WorkflowRunStageStateResolved, duration: 600},
		{name: "timed out legacy seconds", end: 3, nowMs: 4_000, facts: WorkflowRunStageExecutionDetails{DispatchedAtMs: 1_300}, state: WorkflowRunStageStateTimedOut, duration: 1_700},
		{name: "timed out milliseconds take precedence", end: 3, endedMs: 3_200, nowMs: 4_000, facts: WorkflowRunStageExecutionDetails{DispatchedAtMs: 1_300}, state: WorkflowRunStageStateTimedOut, duration: 1_900},
		{name: "dispatch failed", endedMs: 3_000, nowMs: 4_000, facts: WorkflowRunStageExecutionDetails{DispatchAttempts: 2, LastDispatchErrorCode: "queue_unavailable"}, state: WorkflowRunStageStateDispatchFailed},
		{name: "skipped", end: 3, nowMs: 4_000, state: WorkflowRunStageStateSkipped},
		{name: "resolved without dispatch timestamp", nowMs: 4_000, facts: WorkflowRunStageExecutionDetails{ResolvedAtMs: 1_900}, state: WorkflowRunStageStateResolved},
		{name: "future dispatch does not go negative", nowMs: 4_000, facts: WorkflowRunStageExecutionDetails{DispatchedAtMs: 5_000}, state: WorkflowRunStageStateDispatched},
		{name: "out of order resolution does not go negative", nowMs: 4_000, facts: WorkflowRunStageExecutionDetails{DispatchedAtMs: 2_000, ResolvedAtMs: 1_900}, state: WorkflowRunStageStateResolved},
		{name: "zero now", facts: WorkflowRunStageExecutionDetails{DispatchedAtMs: 1_300}, state: WorkflowRunStageStateDispatched},
	} {
		t.Run(tc.name, func(t *testing.T) {
			execution := tc.facts
			run := WorkflowRun{
				Start: 1, End: tc.end, EndedAtMs: tc.endedMs,
				Stages: []WorkflowRunStage{
					{Operation: "anpr", Execution: &execution},
					{Operation: "legacy-only"},
				},
				StageExecutions: []WorkflowRunStageExecution{{
					Operation:        "legacy-only",
					DispatchAttempts: tc.facts.DispatchAttempts, DispatchedAtMs: tc.facts.DispatchedAtMs,
					ResolvedAtMs: tc.facts.ResolvedAtMs,
				}},
			}
			var now time.Time
			if tc.nowMs != 0 {
				now = time.UnixMilli(tc.nowMs)
			}
			run.PopulateRuntimeFields(now)
			got := run.Stages[0].Execution
			if got.State != tc.state || got.DurationMs != tc.duration {
				t.Fatalf("nested state/duration = (%s, %d), want (%s, %d)", got.State, got.DurationMs, tc.state, tc.duration)
			}
			persistedFacts := *got
			persistedFacts.State, persistedFacts.DurationMs = "", 0
			if persistedFacts != tc.facts {
				t.Fatalf("runtime derivation changed persisted facts: %#v, want %#v", persistedFacts, tc.facts)
			}
			if len(run.StageExecutions) != 1 || run.StageExecutions[0].State != tc.state || run.StageExecutions[0].DurationMs != tc.duration {
				t.Fatalf("legacy runtime projection changed: %#v", run.StageExecutions)
			}
			if run.Stages[1].Execution != nil {
				t.Fatal("runtime population implicitly normalized missing execution")
			}
			doc := runStageBSON(t, run)
			stages := doc["stages"].(bson.A)
			if got := stages[0].(bson.M)["execution"]; !reflect.DeepEqual(got, runStageBSON(t, tc.facts)) {
				t.Fatalf("derived fields persisted in nested execution: %#v", got)
			}
		})
	}
}
