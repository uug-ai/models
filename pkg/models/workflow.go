package models

import (
	"crypto/sha256"
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	PermissionWorkflowsRead   Permission = "workflows.read"
	PermissionWorkflowsCreate Permission = "workflows.create"
	PermissionWorkflowsUpdate Permission = "workflows.update"
	PermissionWorkflowsDelete Permission = "workflows.delete"
)

var workflowPermissions = []Permission{
	PermissionWorkflowsRead,
	PermissionWorkflowsCreate,
	PermissionWorkflowsUpdate,
	PermissionWorkflowsDelete,
}

// WorkflowPermissions returns the canonical workflow permission catalog.
func WorkflowPermissions() []Permission {
	return append([]Permission(nil), workflowPermissions...)
}

// WorkflowNodeType is what a canvas node represents. An empty Type is a stage,
// so graphs authored before node types existed compile unchanged.
type WorkflowNodeType string

const (
	// WorkflowNodeStage is an instance of a catalog stage, dispatched to its worker.
	WorkflowNodeStage WorkflowNodeType = "stage"
	// WorkflowNodeDevice is the graph's recording source. It selects the devices
	// whose recordings open the workflow and compiles to the automatic trigger
	// (see SyncGraphTriggers), never to a dispatched stage.
	WorkflowNodeDevice WorkflowNodeType = "device"
)

// WorkflowDeviceGateOperation is the operation a condition on a device node's
// outgoing edge gates on: the classifier result every automatic hand-off seeds
// into the run's Inputs.
const WorkflowDeviceGateOperation = "classify"

// WorkflowSeedOperation is the run-opening hand-off operation. It is reserved:
// a stage with this operation would re-enter the engine as a new run.
const WorkflowSeedOperation = "event"

// WorkflowNode is a single stage instance placed on the workflow canvas. Every
// node is an instance of a catalog stage: StageRef holds the referenced stage's
// Operation key, and the stage definition itself (image, queue, resources,
// dispatch defaults, …) lives in the WorkflowStage catalog entry and is never
// copied onto the node — so editing a stage updates every instance of it. The
// node carries only what is specific to this placement: identity, canvas
// position/label, and optional per-instance parameters (Data). How and when the
// instance fires is expressed by the edges feeding it (see WorkflowEdge.Condition);
// what activates the workflow as a whole is the Workflow's Trigger. A device
// node (Type WorkflowNodeDevice) is the exception: it is the graph's source and
// carries Devices instead of a StageRef.
type WorkflowNode struct {
	// Id is this instance's identity within the workflow. It is the stable
	// handle that edges connect to, and the per-instance runtime key when the
	// same stage is placed more than once.
	Id string `json:"id" bson:"id"`
	// Type is what the node represents; empty means WorkflowNodeStage.
	Type  WorkflowNodeType `json:"type,omitempty" bson:"type,omitempty"`
	Label string           `json:"label" bson:"label,omitempty"`
	X     float64          `json:"x" bson:"x"`
	Y     float64          `json:"y" bson:"y"`
	// StageRef is the referenced stage's Operation key (the catalog key shared
	// by platform- and user-defined stages), not its Mongo Id. Set on every
	// stage node and resolved at compile time; empty on a device node.
	StageRef string `json:"stageRef" bson:"stageRef"`
	// Devices scopes a device node to recordings from these devices. Empty means
	// every device the workflow's owner can see. Ignored on stage nodes.
	Devices []DeviceKey `json:"devices,omitempty" bson:"devices,omitempty"`
	// Data holds optional per-instance parameter values for this placement, keyed
	// by parameter name. They are validated against and defaulted from the
	// referenced stage's declared Params (see WorkflowStage.Params), layered over
	// the stage's catalog defaults.
	Data map[string]interface{} `json:"data,omitempty" bson:"data,omitempty"`
}

// EffectiveType returns the node's type, defaulting an empty Type to
// WorkflowNodeStage.
func (n WorkflowNode) EffectiveType() WorkflowNodeType {
	if n.Type == "" {
		return WorkflowNodeStage
	}
	return n.Type
}

// WorkflowEdge connects node IDs, not operations. Its source is a readiness
// gate; its conditions read absolute paths in the run envelope independently of
// that gate. Without predicates, a stage edge still waits for its source.
type WorkflowEdge struct {
	Id     string `json:"id" bson:"id"`
	Source string `json:"source" bson:"source"`
	// SourcePort names a source port for the editor. It does not rebase paths.
	SourcePort string `json:"sourcePort" bson:"sourcePort"`
	Target     string `json:"target" bson:"target"`
	// TargetPort optionally selects which of the target stage's declared Inputs
	// (see WorkflowStage.Inputs) this edge feeds. Empty means the default port.
	TargetPort    string              `json:"targetPort" bson:"targetPort"`
	ConditionMode ConditionMode       `json:"conditionMode,omitempty" bson:"conditionMode,omitempty"`
	Conditions    []WorkflowCondition `json:"conditions,omitempty" bson:"conditions,omitempty"`
	// Condition is the legacy single-predicate form. Do not combine it with
	// Conditions; ConditionSet reads it as a one-item all group.
	Condition *StageCondition `json:"condition,omitempty" bson:"condition,omitempty"`
}

func (e WorkflowEdge) ConditionSet() (WorkflowConditionSet, error) {
	return NormalizeWorkflowConditions(e.ConditionMode, e.Conditions, e.Condition)
}

func (e WorkflowEdge) ValidateConditions() error {
	set, err := e.ConditionSet()
	if err != nil {
		return err
	}
	return ValidateWorkflowConditionSet(set)
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
	// WorkflowSurfaceRedaction exposes a manual trigger on the redaction editor's
	// submit control, launched against the media being redacted. It is a distinct
	// launch point from the generic case/media "Run workflow" surfaces: a workflow
	// opts in to it (typically a detector→redaction graph, e.g. objecttracking →
	// redaction) so the editor lists the applicable workflows purely as data,
	// without a hard-coded workflow id. The reviewed inline-regions submit stays a
	// private one-stage run and does not use this surface.
	WorkflowSurfaceRedaction WorkflowTriggerSurface = "redaction"
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
	//
	// It is a convenience shorthand: a non-empty list compiles (see
	// CompiledConditions) into a single StageCondition — device.deviceKey `in`
	// [keys…] — evaluated by the same operator engine stage conditions use, so
	// device scoping and stage matching stay consistent. Author richer scoping
	// (a device-name pattern, an organisation check, …) with Conditions.
	Devices []DeviceKey `json:"devices,omitempty" bson:"devices,omitempty"`
	// SiteIds and GroupIds are stable membership IDs resolved by the caller, not
	// names or authorization grants. Selectors are OR within each category and
	// AND between populated categories.
	SiteIds  []string `json:"siteIds,omitempty" bson:"siteIds,omitempty"`
	GroupIds []string `json:"groupIds,omitempty" bson:"groupIds,omitempty"`
	// Conditions read the pre-run envelope, including already available,
	// sanitized inputs, never future results. Scope and schedule remain mandatory
	// even when this group's ConditionMode is any.
	ConditionMode ConditionMode       `json:"conditionMode,omitempty" bson:"conditionMode,omitempty"`
	Conditions    []WorkflowCondition `json:"conditions,omitempty" bson:"conditions,omitempty"`
	// WeeklySchedule bounds the automatic trigger to recurring weekly windows,
	// each with its own day, time segments and IANA Timezone. An empty schedule
	// means any time is eligible. Reuses the alert weekly-schedule shape so the
	// same editor and evaluation semantics apply. The Timezone on each entry is
	// the user's timezone captured when the schedule was authored. Time is not
	// path-expressible, so it stays a dedicated field rather than a condition.
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

// MatchesDevice reports whether deviceKey is in this automatic trigger's device
// list: an empty Devices list matches every device, otherwise the key must
// appear in the list (compared against DeviceKey.Key). It is a device-key-only
// convenience used to answer "does this workflow run for this device?" (e.g.
// hub-api's device filter) and deliberately ignores the trigger's richer
// Conditions. The full activation gate the engine evaluates is Matches, which
// runs every compiled condition (see CompiledConditions) against the recording
// envelope.
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

// CompiledConditions returns a flat view for legacy callers inspecting paths.
// Deprecated: this loses ConditionMode. Use MatchesEnvelope for evaluation,
// Validate for authoring validation, and ConditionSet for the explicit group.
func (t WorkflowTrigger) CompiledConditions() []StageCondition {
	return append(t.scopeConditions(), t.Conditions...)
}

func (t WorkflowTrigger) scopeConditions() []WorkflowCondition {
	out := make([]WorkflowCondition, 0, 3)
	if len(t.Devices) > 0 {
		keys := make([]any, 0, len(t.Devices))
		for _, d := range t.Devices {
			keys = append(keys, d.Key)
		}
		out = append(out, StageCondition{Path: "device.deviceKey", Op: ConditionOpIn, Value: keys})
	}
	if len(t.SiteIds) > 0 {
		out = append(out, WorkflowCondition{Path: "device.siteIds.*", Op: ConditionOpIn, Value: StringsToAny(t.SiteIds)})
	}
	if len(t.GroupIds) > 0 {
		out = append(out, WorkflowCondition{Path: "device.groupIds.*", Op: ConditionOpIn, Value: StringsToAny(t.GroupIds)})
	}
	return out
}

func (t WorkflowTrigger) ConditionSet() WorkflowConditionSet {
	return WorkflowConditionSet{ConditionMode: t.ConditionMode, Conditions: t.Conditions}
}

// MatchesEnvelope reports whether root — the credential-free pre-run projection
// of a recording (device.*, user.*, identity scalars; see AutomaticTriggerRoot)
// — satisfies every selector category AND the explicit condition group.
func (t WorkflowTrigger) MatchesEnvelope(root map[string]any) bool {
	return EvaluateConditionSet(WorkflowConditionSet{Conditions: t.scopeConditions()}, root) &&
		EvaluateConditionSet(t.ConditionSet(), root)
}

// AutomaticTriggerRoot builds the pre-run envelope an automatic trigger's
// conditions match against: the device and user scalars known when a recording
// arrives, in the same nested shape stage conditions read (device.<field>,
// user.<field>). It is the trigger-time counterpart of the engine's fuller run
// root, without any operation inputs or results.
func AutomaticTriggerRoot(device WorkflowDevice, user WorkflowUser) map[string]any {
	return AutomaticTriggerRootWithInputs(device, user, nil)
}

// AutomaticTriggerRootWithInputs adds operation inputs already available at
// hand-off. Callers must supply sanitized result payloads, not raw queue messages
// or storage credentials. It never exposes future run results. GroupIds must be
// resolved from trusted membership data by the caller.
func AutomaticTriggerRootWithInputs(device WorkflowDevice, user WorkflowUser, inputs map[string]any) map[string]any {
	root := map[string]any{
		"device": map[string]any{
			"deviceKey":       device.DeviceKey,
			"deviceName":      device.DeviceName,
			"provider":        device.Provider,
			"storageSolution": device.StorageSolution,
			"siteIds":         StringsToAny(device.SiteIds),
			"groupIds":        StringsToAny(device.GroupIds),
		},
		"user": map[string]any{
			"organisationId": user.OrganisationId,
		},
	}
	if len(inputs) > 0 {
		root["inputs"] = inputs
	}
	return root
}

// StringsToAny preserves the []any shape of legacy condition envelopes.
// The evaluator also accepts typed slices. Empty input yields a non-nil slice.
func StringsToAny(ss []string) []any {
	out := make([]any, len(ss))
	for i, s := range ss {
		out[i] = s
	}
	return out
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

// Matches reports whether the recording described by root at instant at satisfies
// this automatic trigger's scope (device/envelope conditions AND weekly
// schedule). It is the single source of truth for the automatic activation gate,
// evaluated by the engine before a WorkflowRun is opened. It is pure: the caller
// builds the envelope (see AutomaticTriggerRoot) and passes the recording
// timestamp. Manual triggers do not use this predicate — they are gated by
// surface, not by device/time.
func (t WorkflowTrigger) Matches(root map[string]any, at time.Time) bool {
	return t.MatchesEnvelope(root) && t.IsScheduledAt(at)
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
	// WorkflowSourceConfig is a workflow loaded from deployment configuration
	// (helm). It is deployment-global — available to every organisation — and
	// ops-managed: created and updated only through deployment configuration, and
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
	// OrganisationId. WorkflowSourceConfig marks a Helm-defined, deployment-global,
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
	Stages   []WorkflowStage `json:"stages,omitempty" bson:"stages,omitempty"`
	UserId   string          `json:"userId" bson:"userId,omitempty"`
	Username string          `json:"username" bson:"username,omitempty"`
	// OrganisationId is the canonical tenant key. It is persisted as the
	// camelCase `organisationId` field to match the platform-wide convention for
	// organisation-owned domain resources. Database-backed workflows are not yet
	// materially used in deployments, so new documents start with this canonical
	// contract while readers may retain a legacy `organisation_id` fallback.
	OrganisationId string `json:"organisationId" bson:"organisationId,omitempty"`
	// ProjectId optionally places the workflow in a project within its organisation.
	// A nil value keeps the workflow organisation-wide.
	ProjectId *primitive.ObjectID `json:"projectId,omitempty" bson:"projectId,omitempty"`
	CreatedAt int64               `json:"createdAt" bson:"createdAt,omitempty"`
	UpdatedAt int64               `json:"updatedAt" bson:"updatedAt,omitempty"`
	// Audit carries bounded actor/timestamp metadata for database-backed user
	// workflows. It is omitted from Helm-defined global workflows unless set.
	Audit *Audit `json:"audit,omitempty" bson:"audit,omitempty"`
}

// UnmarshalJSON accepts the pre-canonical snake_case ownership and timestamp
// fields during the migration window. Marshal output remains canonical because
// the Workflow field tags above are camelCase. A populated canonical field wins
// when both representations are present.
func (w *Workflow) UnmarshalJSON(data []byte) error {
	type workflowAlias Workflow
	decoded := struct {
		*workflowAlias
		LegacyUserId         string `json:"user_id"`
		LegacyOrganisationId string `json:"organisation_id"`
		LegacyCreatedAt      int64  `json:"created_at"`
		LegacyUpdatedAt      int64  `json:"updated_at"`
	}{workflowAlias: (*workflowAlias)(w)}

	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	w.populateLegacyFields(
		decoded.LegacyUserId,
		decoded.LegacyOrganisationId,
		decoded.LegacyCreatedAt,
		decoded.LegacyUpdatedAt,
	)
	return nil
}

// UnmarshalBSON provides the same transitional compatibility for existing
// workflow documents. New and updated documents are always written using the
// canonical camelCase BSON tags on Workflow.
func (w *Workflow) UnmarshalBSON(data []byte) error {
	type workflowAlias Workflow
	var decoded struct {
		Workflow             workflowAlias `bson:",inline"`
		LegacyUserId         string        `bson:"user_id"`
		LegacyOrganisationId string        `bson:"organisation_id"`
		LegacyCreatedAt      int64         `bson:"created_at"`
		LegacyUpdatedAt      int64         `bson:"updated_at"`
	}

	if err := bson.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*w = Workflow(decoded.Workflow)
	w.populateLegacyFields(
		decoded.LegacyUserId,
		decoded.LegacyOrganisationId,
		decoded.LegacyCreatedAt,
		decoded.LegacyUpdatedAt,
	)
	return nil
}

func (w *Workflow) populateLegacyFields(userId, organisationId string, createdAt, updatedAt int64) {
	if w.UserId == "" {
		w.UserId = userId
	}
	if w.OrganisationId == "" {
		w.OrganisationId = organisationId
	}
	if w.CreatedAt == 0 {
		w.CreatedAt = createdAt
	}
	if w.UpdatedAt == 0 {
		w.UpdatedAt = updatedAt
	}
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
// global workflow is one loaded from deployment configuration
// (WorkflowSourceConfig) with no owning organisation, so surfaces list it
// alongside the caller's own workflows without any per-org copy.
func (w *Workflow) IsGlobal() bool {
	return w.EffectiveSource() == WorkflowSourceConfig && w.OrganisationId == ""
}

// EffectiveID resolves the stable, run-facing id of a workflow and is the single
// source of truth for that derivation. A workflow with an explicit Id (a
// persisted user workflow, or a config workflow that carries one) uses it; a
// config workflow loaded from Helm without an id gets a deterministic id derived
// from its Name, so the same definition always yields the same id across
// restarts and every service agrees on its identity without persisting it — a
// run one service seeds resolves to the workflow another service loaded from the
// same definition. The first 12 bytes of a SHA-256 digest form a valid 24-char
// ObjectID hex; callers that need that hex string (a run's WorkflowId) use
// EffectiveID().Hex().
func (w *Workflow) EffectiveID() primitive.ObjectID {
	if !w.Id.IsZero() {
		return w.Id
	}
	sum := sha256.Sum256([]byte(w.Name))
	var id primitive.ObjectID
	copy(id[:], sum[:12])
	return id
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
//
// Device nodes compile to no stage. An edge leaving a device node adds a need
// only when it carries a condition, gated on WorkflowDeviceGateOperation; an
// unconditional device edge adds nothing, so its target starts the run.
func (w *Workflow) CompileStages() []WorkflowStage {
	if len(w.Stages) > 0 {
		return w.Stages
	}
	opByNode := make(map[string]string, len(w.Nodes))
	deviceNodes := make(map[string]bool)
	for _, n := range w.Nodes {
		if n.EffectiveType() == WorkflowNodeDevice {
			deviceNodes[n.Id] = true
			continue
		}
		opByNode[n.Id] = n.StageRef
	}
	incoming := make(map[string][]WorkflowEdge, len(w.Nodes))
	for _, e := range w.Edges {
		incoming[e.Target] = append(incoming[e.Target], e)
	}
	stages := make([]WorkflowStage, 0, len(w.Nodes))
	for _, n := range w.Nodes {
		if deviceNodes[n.Id] {
			continue
		}
		stage := WorkflowStage{Operation: n.StageRef}
		needs := make([]StageDependency, 0, len(incoming[n.Id]))
		for _, e := range incoming[n.Id] {
			need := StageDependency{
				Operation:     opByNode[e.Source],
				Condition:     e.Condition,
				ConditionMode: e.ConditionMode,
				Conditions:    e.Conditions,
			}
			if deviceNodes[e.Source] {
				set, err := e.ConditionSet()
				if err != nil || len(set.Conditions) > 0 {
					need.Operation = WorkflowDeviceGateOperation
					needs = append(needs, need)
				}
				continue
			}
			needs = append(needs, need)
		}
		if len(needs) == 0 {
			stage.Dispatch = DispatchAlways
		} else {
			stage.Dispatch = DispatchConditional
			stage.Needs = needs
		}
		stages = append(stages, stage)
	}
	return stages
}

// SyncGraphTriggers makes the graph's device node the workflow's automatic
// trigger: every automatic trigger is replaced by one scoped to that node's
// Devices, and manual triggers are kept. A graph without a device node leaves
// Triggers untouched. It is idempotent.
// This legacy editor adapter intentionally replaces richer activation settings.
// New Start-node editors must edit Triggers directly instead of calling it.
func (w *Workflow) SyncGraphTriggers() {
	w.NormalizeTriggers()
	var device *WorkflowNode
	for i := range w.Nodes {
		if w.Nodes[i].EffectiveType() == WorkflowNodeDevice {
			device = &w.Nodes[i]
			break
		}
	}
	if device == nil {
		return
	}
	triggers := make([]WorkflowTrigger, 0, len(w.Triggers)+1)
	triggers = append(triggers, WorkflowTrigger{
		Type:    WorkflowTriggerAutomatic,
		Devices: append([]DeviceKey(nil), device.Devices...),
	})
	for _, t := range w.Triggers {
		if t.EffectiveType() == WorkflowTriggerManual {
			triggers = append(triggers, t)
		}
	}
	w.Triggers = triggers
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

// AutomaticMatches reports whether the recording described by root at instant at
// activates this workflow automatically. It normalizes legacy triggers, then is
// true when the workflow is Enabled and any of its automatic triggers matches
// the recording's envelope and time (see WorkflowTrigger.Matches). It is the
// single gate the engine uses to fan an event out to the workflows it should
// open a run for; manual triggers never activate this way.
func (w *Workflow) AutomaticMatches(root map[string]any, at time.Time) bool {
	_, matched := w.firstMatchingAutomaticTrigger(root, at)
	return matched
}

func (w *Workflow) firstMatchingAutomaticTrigger(root map[string]any, at time.Time) (int, bool) {
	if !w.Enabled {
		return -1, false
	}
	w.NormalizeTriggers()
	for i, t := range w.Triggers {
		if t.EffectiveType() == WorkflowTriggerAutomatic && t.Matches(root, at) {
			return i, true
		}
	}
	return -1, false
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
	WorkflowId string `json:"workflowId"`
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
	WorkflowId string   `json:"workflowId"`
	Workflow   Workflow `json:"workflow"`
}

type UpdateWorkflowOutput struct {
	Workflow *Workflow `json:"workflow"`
}

type DeleteWorkflowInput struct {
	User       User   `json:"user"`
	WorkflowId string `json:"workflowId"`
}

type DeleteWorkflowOutput struct{}
