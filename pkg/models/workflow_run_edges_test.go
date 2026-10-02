package models

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestWorkflowRunSafeMatchedStartEdgeIds(t *testing.T) {
	ids := []string{"entry-person", "entry-motion"}
	run := WorkflowRun{MatchedStartEdgeIds: ids}
	data, err := bson.Marshal(run)
	if err != nil {
		t.Fatal(err)
	}
	var fields bson.M
	if err := bson.Unmarshal(data, &fields); err != nil {
		t.Fatal(err)
	}
	if _, ok := fields["matchedstartedgeids"]; !ok {
		t.Fatal("missing canonical BSON matchedstartedgeids")
	}
	var loaded WorkflowRun
	if err := bson.Unmarshal(data, &loaded); err != nil || !reflect.DeepEqual(loaded.MatchedStartEdgeIds, ids) {
		t.Fatalf("BSON round trip %+v: %v", loaded.MatchedStartEdgeIds, err)
	}
	wire, err := json.Marshal(loaded)
	if err != nil || !strings.Contains(string(wire), `"matchedStartEdgeIds":["entry-person","entry-motion"]`) {
		t.Fatalf("JSON projection %s: %v", wire, err)
	}
	if strings.Contains(string(wire), "triggerMatch") || strings.Contains(string(wire), "conditions") {
		t.Fatal("safe IDs require a detailed trigger payload")
	}
}
