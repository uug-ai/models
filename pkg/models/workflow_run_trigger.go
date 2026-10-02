package models

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

// WorkflowRunTriggerMatch captures the automatic selection that opened a run,
// separate from its stages. It contains only the selected trigger's definition,
// not the input envelope, credentials, or other triggers.
type WorkflowRunTriggerMatch struct {
	// Index is zero-based in the normalized trigger list at selection time.
	// It is historical context, not a stable ID into an edited definition.
	Index int `json:"index" bson:"index"`
	// EvaluatedAtMs is the Unix-millisecond instant used for schedule matching
	// (usually recording time), not when the engine processed or opened the run.
	EvaluatedAtMs int64           `json:"evaluatedAtMs" bson:"evaluatedatms"`
	Trigger       WorkflowTrigger `json:"trigger" bson:"trigger"`
	// MatchedEdgeIds captures all matching automatic Start edges in graph order.
	// Legacy trigger lists omit it; Index/Trigger still describe the first match.
	MatchedEdgeIds []string `json:"matchedEdgeIds,omitempty" bson:"matchededgeids,omitempty"`
}

// MatchAutomaticTrigger returns a detached snapshot of the first matching
// automatic trigger, or nil when none matches. It shares AutomaticMatches'
// short-circuit selection and legacy normalization. Callers must synchronize
// graph-derived triggers before matching, just as for AutomaticMatches.
//
// Only a new run selected automatically should persist this record. Replays
// preserve the stored match; manual and explicitly targeted launches must not
// manufacture one. Errors prevent a selected but unencodable snapshot from
// being mistaken for an unmatched workflow.
func (w *Workflow) MatchAutomaticTrigger(root map[string]any, at time.Time) (*WorkflowRunTriggerMatch, error) {
	index, matched := w.firstMatchingAutomaticTrigger(root, at)
	if !matched {
		return nil, nil
	}
	match := WorkflowRunTriggerMatch{
		Index: index, EvaluatedAtMs: at.UnixMilli(), Trigger: w.Triggers[index],
	}
	match.Trigger.Type = match.Trigger.EffectiveType()
	for i := index; i < len(w.Triggers); i++ {
		trigger := w.Triggers[i]
		if trigger.EdgeId != "" && trigger.EffectiveType() == WorkflowTriggerAutomatic &&
			(i == index || trigger.Matches(root, at)) {
			match.MatchedEdgeIds = append(match.MatchedEdgeIds, trigger.EdgeId)
		}
	}
	data, err := bson.Marshal(match)
	if err != nil {
		return nil, fmt.Errorf("encode workflow run trigger match: %w", err)
	}
	var snapshot WorkflowRunTriggerMatch
	if err := bson.Unmarshal(data, &snapshot); err != nil {
		return nil, fmt.Errorf("decode workflow run trigger match: %w", err)
	}
	return &snapshot, nil
}
