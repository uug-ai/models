package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

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
	// action. Automatic triggers honour Devices and the weekly schedule.
	WorkflowTriggerAutomatic WorkflowTriggerType = "automatic"
	// WorkflowTriggerManual is user-launched from a surface against an explicit
	// media selection. Manual triggers ignore Devices / the schedule and instead
	// advertise where they can be launched from via Surfaces.
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
// mode; the remaining fields are mode-specific. For automatic triggers, Devices
// and WeeklySchedule scope which recordings and times the workflow is eligible
// for (an empty automatic trigger leaves it eligible at all times for everything
// routed to it). For manual triggers those scoping fields are ignored and
// Surfaces lists the UI surfaces the workflow can be launched from.
//
// Devices and WeeklySchedule deliberately reuse the same shapes the alert/
// videowall schedules use (DeviceKey, WeeklySchedule/DayTimeRange), so the same
// device pickers and weekly-schedule editors — and the same weekday convention
// (time.Weekday: 0=Sunday) and per-schedule IANA Timezone — apply here.
type WorkflowTrigger struct {
	// Type is the activation mode. An empty Type is treated as
	// WorkflowTriggerAutomatic for backwards compatibility with triggers authored
	// before manual triggers existed.
	Type WorkflowTriggerType `json:"type,omitempty" bson:"type,omitempty"`
	// Devices restricts the automatic trigger to recordings from the listed
	// devices, matched by DeviceKey.Key. An empty list means every device is
	// eligible. Mirrors the alert device selection (see CustomAlert.DevicesList).
	Devices []DeviceKey `json:"devices,omitempty" bson:"devices,omitempty"`
	// WeeklySchedule bounds the automatic trigger to recurring weekly windows,
	// each with its own day, time segments and IANA Timezone. An empty schedule
	// means any time is eligible. Reuses the alert weekly-schedule shape so the
	// same editor and evaluation semantics apply. The Timezone on each entry is
	// the user's timezone captured when the schedule was authored.
	WeeklySchedule []*WeeklySchedule `json:"weeklySchedule,omitempty" bson:"weeklySchedule,omitempty"`
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

// MatchesDevice reports whether a recording from deviceKey is in scope for this
// automatic trigger. An empty Devices list matches every device; otherwise the
// key must appear in the list (compared against DeviceKey.Key, like a stage's
// device.deviceKey need).
func (t WorkflowTrigger) MatchesDevice(deviceKey string) bool {
	if len(t.Devices) == 0 {
		return true
	}
	for _, d := range t.Devices {
		if d.Key == deviceKey {
			return true
		}
	}
	return false
}

// IsScheduledAt reports whether at falls within this automatic trigger's weekly
// schedule. An empty schedule matches any time; otherwise at must fall within an
// enabled weekly segment. Each schedule entry is evaluated in its own IANA
// Timezone (the user's timezone captured at authoring time), so at may be any
// absolute instant — typically the recording timestamp.
func (t WorkflowTrigger) IsScheduledAt(at time.Time) bool {
	if len(t.WeeklySchedule) == 0 {
		return true
	}
	for _, ws := range t.WeeklySchedule {
		if ws.IsActiveAt(at) {
			return true
		}
	}
	return false
}

// Matches reports whether a recording from deviceKey at instant at satisfies this
// automatic trigger's scope (device selection AND weekly schedule). It is the
// single source of truth for the automatic activation gate, evaluated by the
// producer before a WorkflowRun is minted. It is pure: the caller resolves the
// recording's device key and timestamp and passes them in. Manual triggers do
// not use this predicate — they are gated by surface, not by device/time.
func (t WorkflowTrigger) Matches(deviceKey string, at time.Time) bool {
	return t.MatchesDevice(deviceKey) && t.IsScheduledAt(at)
}

// WorkflowSource is the provenance of a workflow document, which also decides
// its availability. It is a named string (not a boolean) so further provenances
// can be added without a schema change. An empty Source is treated as
// WorkflowSourceUser.
type WorkflowSource string

const (
	// WorkflowSourceUser is a workflow authored by a user through the API and
	// scoped to that user's organisation. It is the default (an empty Source is
	// treated as this) and is editable in the UI.
	WorkflowSourceUser WorkflowSource = "user"
	// WorkflowSourceConfig is a workflow seeded from deployment configuration
	// (helm). It is deployment-global — available to every organisation — and
	// ops-managed: created and updated only by the seeding reconcile, and
	// read-only in the UI. A config workflow carries an empty OrganisationId to
	// signal its global scope (see IsGlobal).
	WorkflowSourceConfig WorkflowSource = "config"
)

// Workflow is a user-defined automation graph composed of stage-instance nodes
// and the edges that route between them. Triggers say what activates it (a
// workflow may have several — e.g. an automatic trigger and a manual one); the
// nodes/edges say what runs.
type Workflow struct {
	Id          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name        string             `json:"name" bson:"name,omitempty"`
	Description string             `json:"description" bson:"description,omitempty"`
	Enabled     bool               `json:"enabled" bson:"enabled"`
	// Source is the workflow's provenance and availability (see WorkflowSource).
	// Empty means WorkflowSourceUser: an ordinary user workflow scoped to its
	// OrganisationId. WorkflowSourceConfig marks a helm-seeded, deployment-global,
	// read-only workflow. Every workflow persisted before this field existed
	// decodes as user.
	Source WorkflowSource `json:"source,omitempty" bson:"source,omitempty"`
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
	Trigger *WorkflowTrigger `json:"trigger,omitempty" bson:"trigger,omitempty"`
	Nodes   []WorkflowNode   `json:"nodes" bson:"nodes"`
	Edges   []WorkflowEdge   `json:"edges" bson:"edges"`
	// Stages is the workflow's executable stage set: the runtime-authoritative
	// projection the engine dispatches against. When set it is used as-is (config
	// workflows author it directly in the helm registry form: operation, dispatch,
	// needs, needsMode); when empty it is derived from Nodes+Edges on demand (UI
	// workflows author the graph and CompileStages projects it). Read it through
	// CompileStages, never directly, so both authoring styles resolve uniformly.
	//
	// Only the routing fields (Operation, Dispatch, Needs, NeedsMode) are
	// meaningful here; a stage's deployment fields (image, queue, replicas, …) are
	// resolved by Operation against the shared deployed catalog, not per workflow.
	// Operations need only be unique within a single workflow, not globally.
	Stages         []WorkflowStage `json:"stages,omitempty" bson:"stages,omitempty"`
	UserId         string          `json:"user_id" bson:"user_id,omitempty"`
	Username       string          `json:"username" bson:"username,omitempty"`
	OrganisationId string          `json:"organisation_id" bson:"organisation_id,omitempty"`
	CreatedAt      int64           `json:"created_at" bson:"created_at,omitempty"`
	UpdatedAt      int64           `json:"updated_at" bson:"updated_at,omitempty"`
}

// EffectiveSource returns the workflow's provenance, defaulting an empty Source
// to WorkflowSourceUser so workflows persisted before Source existed keep their
// original (user) behaviour.
func (w *Workflow) EffectiveSource() WorkflowSource {
	if w.Source == "" {
		return WorkflowSourceUser
	}
	return w.Source
}

// IsGlobal reports whether this workflow is available to every organisation. A
// global workflow is one seeded from deployment configuration
// (WorkflowSourceConfig) with no owning organisation, so surfaces list it
// alongside the caller's own workflows without any per-org copy.
func (w *Workflow) IsGlobal() bool {
	return w.EffectiveSource() == WorkflowSourceConfig && w.OrganisationId == ""
}

// CompileStages returns the workflow's executable stage set — the routing the
// engine dispatches against. It is the single entry point for both authoring
// styles: if Stages is populated (config workflows, authored directly in the
// helm registry form) it is returned as-is; otherwise it is derived from the
// graph, projecting each node into a stage and each incoming edge into a need.
//
// The projection follows the graph's routing contract (see WorkflowEdge and
// WorkflowStage.Needs): a node with no incoming edges dispatches always (a start
// stage); a node with one or more incoming edges dispatches conditionally, with
// one need per incoming edge — the need's Operation is the edge's source stage
// (its readiness gate) and the need's Condition is the edge's predicate (nil for
// an unconditional dependency). NeedsMode is left at its default (any). Only
// routing fields are populated; deployment is resolved elsewhere by Operation.
func (w *Workflow) CompileStages() []WorkflowStage {
	if len(w.Stages) > 0 {
		return w.Stages
	}
	opByNode := make(map[string]string, len(w.Nodes))
	for _, n := range w.Nodes {
		opByNode[n.Id] = n.StageRef
	}
	incoming := make(map[string][]WorkflowEdge, len(w.Nodes))
	for _, e := range w.Edges {
		incoming[e.Target] = append(incoming[e.Target], e)
	}
	stages := make([]WorkflowStage, 0, len(w.Nodes))
	for _, n := range w.Nodes {
		stage := WorkflowStage{Operation: n.StageRef}
		edges := incoming[n.Id]
		if len(edges) == 0 {
			stage.Dispatch = DispatchAlways
		} else {
			stage.Dispatch = DispatchConditional
			needs := make([]StageDependency, 0, len(edges))
			for _, e := range edges {
				needs = append(needs, StageDependency{
					Operation: opByNode[e.Source],
					Condition: e.Condition,
				})
			}
			stage.Needs = needs
		}
		stages = append(stages, stage)
	}
	return stages
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

// AutomaticMatches reports whether a recording from deviceKey at instant at
// activates this workflow automatically. It normalizes legacy triggers, then is
// true when the workflow is Enabled and any of its automatic triggers matches
// the recording's device and time (see WorkflowTrigger.Matches). It is the
// single gate the engine uses to fan an event out to the workflows it should
// open a run for; manual triggers never activate this way.
func (w *Workflow) AutomaticMatches(deviceKey string, at time.Time) bool {
	if !w.Enabled {
		return false
	}
	w.NormalizeTriggers()
	for _, t := range w.Triggers {
		if t.EffectiveType() == WorkflowTriggerAutomatic && t.Matches(deviceKey, at) {
			return true
		}
	}
	return false
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
