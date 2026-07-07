package models

import (
	"encoding/json"
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestWorkflowRun_MarshalJSON_ProjectsRunIdFromId asserts the single-source-of-truth
// invariant: RunId is the wire projection of the persisted Id, derived automatically
// at marshal time so a producer only ever sets Id and the two representations can
// never drift.
func TestWorkflowRun_MarshalJSON_ProjectsRunIdFromId(t *testing.T) {
	id := primitive.NewObjectID()

	t.Run("set Id emits runId as the hex", func(t *testing.T) {
		b, err := json.Marshal(WorkflowRun{Operation: "anpr", Key: "media-1", Id: id})
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		var wire map[string]any
		if err := json.Unmarshal(b, &wire); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if got := wire["runId"]; got != id.Hex() {
			t.Errorf("runId = %v, want %s", got, id.Hex())
		}
		// The persisted identity itself never appears on the wire.
		if _, ok := wire["_id"]; ok {
			t.Errorf("wire carried _id; Id must stay persistence-only")
		}
		if _, ok := wire["id"]; ok {
			t.Errorf("wire carried id; Id must stay persistence-only")
		}
	})

	t.Run("zero Id omits runId", func(t *testing.T) {
		b, err := json.Marshal(WorkflowRun{Operation: "event", Key: "media-1"})
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		if strings.Contains(string(b), "runId") {
			t.Errorf("runId present on a run with no Id yet: %s", b)
		}
	})

	t.Run("pointer marshal path also projects", func(t *testing.T) {
		b, err := json.Marshal(&WorkflowRun{Id: id})
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		if !strings.Contains(string(b), id.Hex()) {
			t.Errorf("pointer marshal did not project runId: %s", b)
		}
	})

	t.Run("wire round-trip keeps RunId and leaves Id zero", func(t *testing.T) {
		var r WorkflowRun
		if err := json.Unmarshal([]byte(`{"operation":"anpr","runId":"`+id.Hex()+`"}`), &r); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if r.RunId != id.Hex() {
			t.Errorf("RunId = %q, want %s", r.RunId, id.Hex())
		}
		if !r.Id.IsZero() {
			t.Errorf("Id = %s, want zero (Id is json:\"-\", never read off the wire)", r.Id.Hex())
		}
	})
}

// TestAutomaticRunObjectID asserts the deterministic automatic run identity:
// stable for a given (key, org, workflow) triple, distinct across triples, never
// zero, and — via MarshalJSON — projected onto the wire RunId so producer and
// engine agree on the same runId.
func TestAutomaticRunObjectID(t *testing.T) {
	a := AutomaticRunObjectID("media-1", "org-1", "wf-1")

	t.Run("deterministic for the same triple", func(t *testing.T) {
		if b := AutomaticRunObjectID("media-1", "org-1", "wf-1"); a != b {
			t.Errorf("same triple produced different ids: %s vs %s", a.Hex(), b.Hex())
		}
	})

	t.Run("not the zero id", func(t *testing.T) {
		if a.IsZero() {
			t.Error("derived id should never be zero")
		}
	})

	t.Run("distinct per field, with unambiguous boundaries", func(t *testing.T) {
		cases := map[string]primitive.ObjectID{
			"other key":      AutomaticRunObjectID("media-2", "org-1", "wf-1"),
			"other org":      AutomaticRunObjectID("media-1", "org-2", "wf-1"),
			"other workflow": AutomaticRunObjectID("media-1", "org-1", "wf-2"),
			// Without a separator these two would hash the same bytes.
			"boundary shift": AutomaticRunObjectID("media", "1org", "1wf-1"),
		}
		for name, id := range cases {
			if id == a {
				t.Errorf("%s should differ from the base id", name)
			}
		}
	})

	t.Run("projects onto the wire runId via MarshalJSON", func(t *testing.T) {
		b, err := json.Marshal(WorkflowRun{Operation: "event", Key: "media-1", Id: a})
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		if want := `"runId":"` + a.Hex() + `"`; !strings.Contains(string(b), want) {
			t.Errorf("expected %s in %s", want, b)
		}
	})
}
