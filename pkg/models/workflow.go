package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// WorkflowNode is a single stage instance placed on the workflow canvas. Every
// node is an instance of a catalog stage: StageRef holds the referenced stage's
// Operation key, and the stage definition itself (image, queue, resources,
// dispatch defaults, …) lives in the WorkflowStage catalog entry and is never
// copied onto the node — so editing a stage updates every instance of it. The
// node carries only what is specific to this placement: identity, canvas
// position/label, and optional per-instance parameters (Data). How and when the
// instance fires is expressed by the edges feeding it (see WorkflowEdge.Condition);
// what activates the workflow as a whole is the Workflow's Trigger.
type WorkflowNode struct {
	// Id is this instance's identity within the workflow. It is the stable
	// handle that edges connect to, and the per-instance runtime key when the
	// same stage is placed more than once.
	Id    string  `json:"id" bson:"id"`
	Label string  `json:"label" bson:"label,omitempty"`
	X     float64 `json:"x" bson:"x"`
	Y     float64 `json:"y" bson:"y"`
	// StageRef is the referenced stage's Operation key (the catalog key shared
	// by platform- and user-defined stages), not its Mongo Id. Always set:
	// every node is an instance of a catalog stage, resolved at compile time.
	StageRef string `json:"stageRef" bson:"stageRef"`
	// Data holds optional per-instance parameter values for this placement, keyed
	// by parameter name. They are validated against and defaulted from the
	// referenced stage's declared Params (see WorkflowStage.Params), layered over
	// the stage's catalog defaults.
	Data map[string]interface{} `json:"data,omitempty" bson:"data,omitempty"`
}

// WorkflowEdge is a directed connection from a source node to a target node,
// and is where a workflow expresses routing. An edge with a nil Condition is an
// unconditional dependency: the target runs after the source. An edge with a
// Condition makes the target conditional on the source — the target's stage
// fires only when the source stage resolves and the predicate matches its
// result. A node with one or more conditional incoming edges compiles to a
// DispatchConditional stage whose Needs is the set of those edges' sources;
// otherwise it compiles to DispatchAlways. Because each incoming edge carries
// its own Condition, per-upstream predicates are expressible — unlike a single
// condition on the target node.
type WorkflowEdge struct {
	Id     string `json:"id" bson:"id"`
	Source string `json:"source" bson:"source"`
	// SourcePort optionally selects which of the source stage's declared Outputs
	// (see WorkflowStage.Outputs) this edge reads; Condition is evaluated against
	// that output's result. Empty means the stage's single implicit default port.
	SourcePort string `json:"sourcePort" bson:"sourcePort"`
	Target     string `json:"target" bson:"target"`
	// TargetPort optionally selects which of the target stage's declared Inputs
	// (see WorkflowStage.Inputs) this edge feeds. Empty means the default port.
	TargetPort string `json:"targetPort" bson:"targetPort"`
	// Condition is the structured predicate evaluated against the source stage's
	// result. Nil means the edge is an unconditional dependency. The edge is the
	// authoring source of truth for routing: this Condition is what compiles into
	// the target stage's Needs[].Condition (see WorkflowStage.Needs), which is the
	// derived runtime projection.
	Condition *StageCondition `json:"condition,omitempty" bson:"condition,omitempty"`
}

// WorkflowTriggerType is how a trigger activates its workflow. Automatic
// triggers fire on their own for every matching recording (pipeline-teed by
// hub-pipeline-analysis); manual triggers are launched on demand by a user from
// a UI surface (see Surfaces) against an explicit selection of media. Both kinds
// converge on the same workflows queue, engine and stages — only the run origin
// differs (see WorkflowRun.Origin).
type WorkflowTriggerType string

const (
	// WorkflowTriggerAutomatic fires for every matching recording without user
	// action. Automatic triggers honour Selection and the time window.
	WorkflowTriggerAutomatic WorkflowTriggerType = "automatic"
	// WorkflowTriggerManual is user-launched from a surface against an explicit
	// media selection. Manual triggers ignore Selection / the time window and
	// instead advertise where they can be launched from via Surfaces.
	WorkflowTriggerManual WorkflowTriggerType = "manual"
)

// WorkflowTriggerSurface is a place in the product a manual trigger can be
// launched from. It lets a workflow opt in to (for example) a case's "Run
// workflow" control purely as data, so surfaces list the workflows that apply to
// them without any hard-coded workflow ids.
type WorkflowTriggerSurface string

const (
	// WorkflowSurfaceCase exposes a manual trigger on a case/task, where it runs
	// against the media the user selected in that case.
	WorkflowSurfaceCase WorkflowTriggerSurface = "case"
	// WorkflowSurfaceMedia exposes a manual trigger on an individual media item.
	WorkflowSurfaceMedia WorkflowTriggerSurface = "media"
)

// WorkflowTrigger defines what activates a workflow. Type selects the activation
// mode; the remaining fields are mode-specific. For automatic triggers,
// Selection/StartAt/EndAt/Weekdays scope which recordings and times the workflow
// is eligible for (an empty automatic trigger leaves it eligible at all times for
// everything routed to it). For manual triggers those scheduling fields are
// ignored and Surfaces lists the UI surfaces the workflow can be launched from.
type WorkflowTrigger struct {
	// Type is the activation mode. An empty Type is treated as
	// WorkflowTriggerAutomatic for backwards compatibility with triggers authored
	// before manual triggers existed.
	Type WorkflowTriggerType `json:"type,omitempty" bson:"type,omitempty"`
	// Selection is the set of devices/streams the workflow applies to (automatic).
	Selection string `json:"selection,omitempty" bson:"selection,omitempty"`
	// StartAt and EndAt bound a daily time window, e.g. "08:00"/"18:00" (automatic).
	StartAt string `json:"startAt,omitempty" bson:"startAt,omitempty"`
	EndAt   string `json:"endAt,omitempty" bson:"endAt,omitempty"`
	// Weekdays restricts the workflow to the listed days of week (automatic).
	Weekdays []int `json:"weekdays,omitempty" bson:"weekdays,omitempty"`
	// Surfaces lists the UI surfaces a manual trigger can be launched from
	// (manual). Ignored for automatic triggers.
	Surfaces []WorkflowTriggerSurface `json:"surfaces,omitempty" bson:"surfaces,omitempty"`
}

// EffectiveType returns the trigger's activation mode, defaulting an empty Type
// to WorkflowTriggerAutomatic so legacy triggers keep their original behaviour.
func (t WorkflowTrigger) EffectiveType() WorkflowTriggerType {
	if t.Type == "" {
		return WorkflowTriggerAutomatic
	}
	return t.Type
}

// HasSurface reports whether this (manual) trigger is launchable from surface.
func (t WorkflowTrigger) HasSurface(surface WorkflowTriggerSurface) bool {
	for _, s := range t.Surfaces {
		if s == surface {
			return true
		}
	}
	return false
}

// Workflow is a user-defined automation graph composed of stage-instance nodes
// and the edges that route between them. Triggers say what activates it (a
// workflow may have several — e.g. an automatic trigger and a manual one); the
// nodes/edges say what runs.
type Workflow struct {
	Id          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name        string             `json:"name" bson:"name,omitempty"`
	Description string             `json:"description" bson:"description,omitempty"`
	Enabled     bool               `json:"enabled" bson:"enabled"`
	// Triggers is the set of activation modes for this workflow. A workflow may
	// carry both an automatic and a manual trigger so it runs on its own for
	// matching recordings and can also be launched on demand from a surface.
	Triggers []WorkflowTrigger `json:"triggers,omitempty" bson:"triggers,omitempty"`
	// Trigger is the legacy single-trigger field kept only so workflows persisted
	// before Triggers existed still decode (and are not silently dropped on
	// re-save). New code should read and write Triggers; call NormalizeTriggers to
	// fold any legacy value into Triggers.
	//
	// Deprecated: use Triggers.
	Trigger        *WorkflowTrigger `json:"trigger,omitempty" bson:"trigger,omitempty"`
	Nodes          []WorkflowNode   `json:"nodes" bson:"nodes"`
	Edges          []WorkflowEdge   `json:"edges" bson:"edges"`
	UserId         string           `json:"user_id" bson:"user_id,omitempty"`
	Username       string           `json:"username" bson:"username,omitempty"`
	OrganisationId string           `json:"organisation_id" bson:"organisation_id,omitempty"`
	CreatedAt      int64            `json:"created_at" bson:"created_at,omitempty"`
	UpdatedAt      int64            `json:"updated_at" bson:"updated_at,omitempty"`
}

// NormalizeTriggers folds a legacy single Trigger into the Triggers list and
// clears the deprecated field, so callers only ever have to reason about
// Triggers. It is idempotent: if Triggers is already populated the legacy field
// is simply dropped, and if neither is set it does nothing.
func (w *Workflow) NormalizeTriggers() {
	if w.Trigger != nil {
		if len(w.Triggers) == 0 {
			w.Triggers = append(w.Triggers, *w.Trigger)
		}
		w.Trigger = nil
	}
}

// ManualTriggersForSurface returns the workflow's manual triggers that are
// launchable from surface. It normalizes legacy triggers first, so a workflow
// authored before manual triggers existed (automatic-only) simply yields none.
func (w *Workflow) ManualTriggersForSurface(surface WorkflowTriggerSurface) []WorkflowTrigger {
	w.NormalizeTriggers()
	var out []WorkflowTrigger
	for _, t := range w.Triggers {
		if t.EffectiveType() == WorkflowTriggerManual && t.HasSurface(surface) {
			out = append(out, t)
		}
	}
	return out
}

// Input / Output types for repository operations

type GetWorkflowsInput struct {
	User User `json:"user"`
}

type GetWorkflowsOutput struct {
	Workflows []Workflow `json:"workflows"`
}

type GetWorkflowInput struct {
	User       User   `json:"user"`
	WorkflowId string `json:"workflow_id"`
}

type GetWorkflowOutput struct {
	Workflow *Workflow `json:"workflow"`
}

type CreateWorkflowInput struct {
	User     User     `json:"user"`
	Workflow Workflow `json:"workflow"`
}

type CreateWorkflowOutput struct {
	Workflow *Workflow `json:"workflow"`
}

type UpdateWorkflowInput struct {
	User       User     `json:"user"`
	WorkflowId string   `json:"workflow_id"`
	Workflow   Workflow `json:"workflow"`
}

type UpdateWorkflowOutput struct {
	Workflow *Workflow `json:"workflow"`
}

type DeleteWorkflowInput struct {
	User       User   `json:"user"`
	WorkflowId string `json:"workflow_id"`
}

type DeleteWorkflowOutput struct{}
