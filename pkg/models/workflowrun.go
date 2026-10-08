package models

import (
	"crypto/sha256"
	"encoding/json"
	"reflect"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	PermissionWorkflowRunsCreate Permission = "workflow-runs.create"
	PermissionWorkflowRunsRead   Permission = "workflow-runs.read"
	PermissionWorkflowRunsUpdate Permission = "workflow-runs.update"

	// Workflow runs only contain contemporary timestamps. Values below this
	// boundary are legacy Unix seconds; canonical values are Unix milliseconds.
	workflowRunMillisecondTimestampThreshold int64 = 100_000_000_000
)

var workflowRunPermissions = []Permission{
	PermissionWorkflowRunsCreate,
	PermissionWorkflowRunsRead,
	PermissionWorkflowRunsUpdate,
}

// WorkflowRunPermissions returns the canonical workflow-run permission catalog.
func WorkflowRunPermissions() []Permission {
	return append([]Permission(nil), workflowRunPermissions...)
}

// WorkflowRunOrigin records how a run was opened: automatically (teed off the
// pipeline by analysis for a matching recording) or manually (launched on demand
// by a user from a surface). It is the run-side counterpart of a Workflow
// trigger's Type, and shares its values.
type WorkflowRunOrigin string

const (
	// WorkflowOriginAutomatic is a run opened by the automatic pipeline tee. It is
	// also the default for runs opened before origins were recorded.
	WorkflowOriginAutomatic WorkflowRunOrigin = "automatic"
	// WorkflowOriginManual is a run launched on demand by a user from a surface.
	// The engine skips automatic selection/time gating for manual runs.
	WorkflowOriginManual WorkflowRunOrigin = "manual"
)

// WorkflowRunState is the coarse, client-facing lifecycle of a run, derived
// from the persisted Start/End and the dispatched/resolved operation sets. It
// is NOT stored on the run — the run carries only the raw lifecycle fields —
// but computed on read (see WorkflowRun.LifecycleState) so a surface polling a
// run can render "still working" vs "results are in" without duplicating the
// derivation. A dedicated "failed" state is intentionally absent: the run
// document has no failure marker (the engine either finalises a run or leaves
// it open), so a run that never ends reads as running until a caller-side
// deadline gives up on it.
type WorkflowRunState string

const (
	// WorkflowRunStateRunning is an open run: it has not been finalised (End is
	// unset), so at least one dispatched stage is still outstanding.
	WorkflowRunStateRunning WorkflowRunState = "running"
	// WorkflowRunStateCompleted is a finalised run that produced output: it has
	// ended and at least one stage resolved (or accumulated Results), so there is
	// something for the surface to surface.
	WorkflowRunStateCompleted WorkflowRunState = "completed"
	// WorkflowRunStateNoResult is a finalised run that produced nothing: it ended
	// having dispatched and resolved no stages. This is the visible face of the
	// silent no-op — a conditional workflow whose gates matched nothing, or a
	// workflow the engine could not dispatch (e.g. a user/DB workflow absent from
	// its registry) — so a surface can say "completed, no results" rather than
	// implying success.
	WorkflowRunStateNoResult WorkflowRunState = "noResult"
)

// WorkflowRunOperationState reports whether a dispatched workflow operation is
// still outstanding or has been resolved by the engine.
type WorkflowRunOperationState string

const (
	WorkflowRunOperationStateDispatched WorkflowRunOperationState = "dispatched"
	WorkflowRunOperationStateResolved   WorkflowRunOperationState = "resolved"
)

// WorkflowRunOperationStatus is the render-ready lifecycle of one dispatched
// operation. It is derived from DispatchedOperations and ResolvedOperations and
// is never persisted.
type WorkflowRunOperationStatus struct {
	Operation string                    `json:"operation" bson:"-"`
	Status    WorkflowRunOperationState `json:"status" bson:"-"`
}

// WorkflowRunStageState is the client-facing lifecycle of one planned stage.
// It is derived from the execution timestamps, dispatch attempts, and the
// containing run's lifecycle rather than persisted independently.
type WorkflowRunStageState string

const (
	WorkflowRunStageStateWaiting        WorkflowRunStageState = "waiting"
	WorkflowRunStageStateRetrying       WorkflowRunStageState = "retrying"
	WorkflowRunStageStateDispatched     WorkflowRunStageState = "dispatched"
	WorkflowRunStageStateResolved       WorkflowRunStageState = "resolved"
	WorkflowRunStageStateDispatchFailed WorkflowRunStageState = "dispatchFailed"
	WorkflowRunStageStateTimedOut       WorkflowRunStageState = "timedOut"
	WorkflowRunStageStateSkipped        WorkflowRunStageState = "skipped"
)

// WorkflowRunStageExecution is the bounded, durable execution summary for one
// stage in a run. Operation, Name, and Dependencies snapshot the user-visible
// stage plan without copying deployment configuration or secrets. The engine
// owns the persisted facts; State and DurationMs are derived when a run is read.
//
// All timestamps are Unix milliseconds. A dispatch attempt is recorded before
// queue publication, DispatchedAtMs after successful publication, and
// ResolvedAtMs when the engine accepts the stage result.
//
// Deprecated: new run writers store WorkflowRunStage.Execution. Retained for
// legacy split documents and compatibility API projections.
type WorkflowRunStageExecution struct {
	Operation    string   `json:"operation" bson:"operation"`
	Name         string   `json:"name,omitempty" bson:"name,omitempty"`
	Dependencies []string `json:"dependencies,omitempty" bson:"dependencies,omitempty"`

	DispatchAttempts         int    `json:"dispatchAttempts,omitempty" bson:"dispatchattempts,omitempty"`
	FirstDispatchAttemptAtMs int64  `json:"firstDispatchAttemptAtMs,omitempty" bson:"firstdispatchattemptatms,omitempty"`
	LastDispatchAttemptAtMs  int64  `json:"lastDispatchAttemptAtMs,omitempty" bson:"lastdispatchattemptatms,omitempty"`
	DispatchedAtMs           int64  `json:"dispatchedAtMs,omitempty" bson:"dispatchedatms,omitempty"`
	ResolvedAtMs             int64  `json:"resolvedAtMs,omitempty" bson:"resolvedatms,omitempty"`
	LastDispatchErrorCode    string `json:"lastDispatchErrorCode,omitempty" bson:"lastdispatcherrorcode,omitempty"`

	State      WorkflowRunStageState `json:"state,omitempty" bson:"-"`
	DurationMs int64                 `json:"durationMs,omitempty" bson:"-"`
}

// WorkflowRun is the single type the workflow subsystem uses for a run, in both
// of its representations:
//
//   - the self-contained MESSAGE exchanged on the workflows queue and the
//     per-stage queues (JSON), and
//   - the persisted run DOCUMENT in the workflow_runs collection (BSON).
//
// It is deliberately NOT a PipelineEvent: the workflows tail is a separate
// fan-out, so the message carries only the data a run needs — copied from the
// upstream pipeline at hand-off time — rather than inheriting the whole pipeline
// envelope. This gives an explicit, auditable contract for what a run sees and
// what it persists, with the producer (analysis) in full control of what
// crosses the boundary.
//
// The same shape travels every hop of the workflow tail:
//
//	analysis ──WorkflowRun{operation:"event"}──────────────────▶ engine
//	engine   ──WorkflowRun{operation:<stage>, storage}─────────▶ stage worker
//	worker   ──WorkflowRun{operation:<stage>, payload|results}─▶ engine
//
// so each consumer reads and writes the one object instead of reconstructing
// state from a generic bag.
//
// Integrator contract (engine ⇄ a custom stage worker) — the stable surface a
// third-party stage codes against:
//
//   - A worker RECEIVES the engine→worker dispatch above: its Operation, the
//     run identity (RunId, Key) and trace (TraceId), the curated User/Device
//     context, the immutable start context (Inputs) and accumulated upstream
//     outputs (Results), and the Storage credentials to fetch the media.
//   - A worker RETURNS the same envelope it received — echo RunId, Key, TraceId
//     and User so the engine can locate and scope the run — with Storage
//     cleared and its result in exactly ONE channel. A delegated-ingest worker
//     sets Payload to a self-describing block envelope — one or more typed
//     blocks the shared ingest core routes by each block's own type (e.g. a
//     "detection" block carrying a PostDetectionsRequest, optionally followed by
//     "marker" blocks) — which the engine persists and mirrors, grouped by block
//     type, into Results.
//     A self-persisting worker instead writes its own collection and returns
//     just its routing values under Results[operation]. A worker never populates
//     both.
//
// Tag discipline keeps the two representations from bleeding into each other:
//   - `bson:"-"` marks WIRE-ONLY fields (transport role, curated projections,
//     the credential-bearing Storage and SignedURL) so they NEVER persist to
//     Mongo — most importantly the credential carriers, so secrets can never
//     land in run state.
//   - `json:"-"` marks internal ownership fields that must never leave the
//     service boundary.
//   - Persisted lifecycle fields are JSON-visible so API list/detail surfaces
//     can return projected WorkflowRun values directly. They use omitempty and
//     remain absent from the deliberately sparse queue messages.
//   - Fields tagged for both are genuine overlap between message and document.
type WorkflowRun struct {
	// Operation marks the message's role on the workflows queue (wire-only):
	//   - "event": a fresh run hand-off from analysis. It opens the run and
	//     carries the start context in Inputs (e.g. the classification result).
	//   - any other value (e.g. "anpr"): either the engine dispatching that
	//     stage to its worker (Storage populated), or the worker routing its
	//     result back (Payload or Results populated). These never collide because the
	//     workflows queue only ever carries the "event" open and worker results —
	//     a dispatch goes to the worker's own queue — so the engine never has to
	//     disambiguate a dispatch from a result.
	Operation string `json:"operation,omitempty" bson:"-"`

	// RunId is the run's identifier on the wire (the hex of the document Id). It
	// is empty on an untargeted analysis hand-off; after workflow matching, the
	// engine derives a distinct document identity per recording, organisation,
	// and workflow. It is set on every engine→worker dispatch. The persisted
	// identity is Id, so RunId itself is wire-only.
	//
	// It is never set by hand: MarshalJSON derives it from Id whenever a run that
	// has been opened (its _id is set) is serialized, so a producer only has to
	// stamp Id and the two representations can never drift. A run without an Id
	// yet (the analysis hand-off) keeps whatever RunId it was given (normally
	// empty, so runId is omitted).
	RunId string `json:"runId,omitempty" bson:"-"`

	// Id is the run document's Mongo identity. Persistence-only — on the wire the
	// run is referenced by RunId (Id.Hex()). Setting Id is sufficient for the
	// wire identity: MarshalJSON emits RunId from it, so callers never derive the
	// hex themselves.
	Id primitive.ObjectID `json:"-" bson:"_id,omitempty"`

	// WorkflowId is the id of the Workflow definition (models.Workflow) this run
	// executes — the authored graph (nodes/edges/triggers) the run is an execution
	// of. It is empty only on an untargeted analysis hand-off; the engine stamps it
	// as it fans that recording out to each matching config or database workflow.
	WorkflowId string `json:"workflowId,omitempty" bson:"workflowid,omitempty"`

	// WorkflowName is the human-readable name of that Workflow, carried so a
	// dispatch/result is legible in worker context and logs without a lookup.
	// Populated alongside WorkflowId.
	WorkflowName string `json:"workflowName,omitempty" bson:"workflowname,omitempty"`

	// Stages captures the run's compiled rules and per-stage execution facts.
	// New writers use NewWorkflowRunStages after resolving queues and defaults.
	// Legacy routing-only elements decode unchanged; NormalizeStages joins any
	// split lifecycle summaries. Nil means no captured routing (legacy fallback);
	// an explicit empty slice is an authoritative zero-stage plan.
	Stages []WorkflowRunStage `json:"stages,omitempty" bson:"stages"`

	// Origin records how this run was opened — the run-side counterpart of the
	// Workflow's trigger Type. An automatic run was teed off the pipeline by
	// analysis for a matching recording; a manual run was launched on demand by a
	// user from a surface. The engine reads it to skip automatic selection/time
	// gating for on-demand runs; it is persisted so "all runs for a media key" can
	// be filtered by how they started. Empty is treated as automatic for runs
	// opened before origins existed.
	Origin WorkflowRunOrigin `json:"origin,omitempty" bson:"origin,omitempty"`

	// TriggerMatch records the first selecting automatic trigger and all matching
	// Start edge IDs. It is immutable activation provenance, not routing state.
	// It is immutable engine-owned provenance, not stage execution state. Nil
	// means no captured match (legacy or a launch that bypassed automatic matching).
	// Readers must not reconstruct it from Origin or today's workflow definition.
	TriggerMatch *WorkflowRunTriggerMatch `json:"triggerMatch,omitempty" bson:"triggermatch,omitempty"`
	// MatchedStartEdgeIds is the safe activation provenance projection for run lists.
	// It contains no trigger definitions or predicate values.
	MatchedStartEdgeIds []string `json:"matchedStartEdgeIds,omitempty" bson:"matchedstartedgeids,omitempty"`

	// SourceRef ties a manual run back to the thing it was launched from — e.g.
	// the case id when launched from a case surface — so sibling runs fanned out
	// from one user action (one seed per selected media key) can be grouped above
	// the run. Empty for automatic runs. It generalises to any run-grouping handle
	// (a case id today; a temporal device-series id is a forward-looking twin).
	SourceRef string `json:"sourceRef,omitempty" bson:"sourceref,omitempty"`

	// SourceType disambiguates case launches from media launch-group references.
	// Legacy readers infer case ownership from SourceRef and case input metadata.
	SourceType string `json:"sourceType,omitempty" bson:"sourcetype,omitempty"`

	SourceAccess string `json:"sourceAccess,omitempty" bson:"-"`
	SourceLabel  string `json:"sourceLabel,omitempty" bson:"-"`
	CaseMediaId  string `json:"caseMediaId,omitempty" bson:"-"`

	// Key is the media key the run is about. It is copied from the recording at
	// hand-off time and can group all runs for that recording, but is not unique:
	// run state is correlated by Id/RunId because several workflows and manual
	// re-runs may execute over the same key.
	Key string `json:"key,omitempty" bson:"key"`

	// RecordingTimestamp is the recording's start time (unix seconds), copied
	// from the recording at hand-off time. It is denormalised onto any platform
	// artifact the engine ingests (see Payload) so cleanup expires the artifact
	// on the recording's retention clock rather than the post time. It is set on
	// the analysis hand-off and persisted on the run at open; the engine then
	// reads it from the run document when stamping an ingested artifact. It is
	// therefore engine-internal — NOT sent on the engine→worker dispatch and not
	// something a worker has to echo back, so it is not part of the stage
	// contract.
	RecordingTimestamp int64 `json:"recordingTimestamp,omitempty" bson:"recordingtimestamp,omitempty"`

	// OrganisationId scopes the run to the owning organisation. Persistence-only:
	// on the wire the same identity travels in the richer User projection.
	OrganisationId string `json:"-" bson:"organisationId"`

	// ProjectId is persisted here; its wire value travels in User.
	ProjectId *primitive.ObjectID `json:"-" bson:"projectId,omitempty"`

	// TraceId continues the distributed trace across the workflow tail and lets
	// authorized detail surfaces correlate the durable run with telemetry.
	TraceId string `json:"traceId,omitempty" bson:"traceid"`

	// Start and End stamp the run's lifecycle in Unix milliseconds. Readers
	// normalize legacy runs whose values were persisted in Unix seconds.
	Start int64 `json:"start,omitempty" bson:"start"`
	End   int64 `json:"end,omitempty" bson:"end,omitempty"`

	// StageExecutions reads the legacy split timeline. NormalizeStages moves
	// matched summaries into Stages, retaining unmatched history without
	// reconstructing missing routing. Serialization does not normalize implicitly.
	// Deprecated: new writers use Stages[i].Execution.
	StageExecutions []WorkflowRunStageExecution `json:"stageExecutions,omitempty" bson:"stageexecutions,omitempty"`

	// The fields below are API read projections. They are derived or joined by
	// the service after loading a run and never persist back into workflow state.
	State      WorkflowRunState             `json:"state,omitempty" bson:"-"`
	DurationMs int64                        `json:"durationMs,omitempty" bson:"-"`
	MediaId    string                       `json:"mediaId,omitempty" bson:"-"`
	DeviceKey  string                       `json:"deviceKey,omitempty" bson:"-"`
	DeviceName string                       `json:"deviceName,omitempty" bson:"-"`
	Dispatched int                          `json:"dispatched,omitempty" bson:"-"`
	Resolved   int                          `json:"resolved,omitempty" bson:"-"`
	Operations []WorkflowRunOperationStatus `json:"operations,omitempty" bson:"-"`
	HasResults bool                         `json:"hasResults,omitempty" bson:"-"`

	// User is the curated, secret-free account context a run needs: the
	// organisation that owns the recording (for logging/scoping) and the account
	// Storage block used to resolve a per-recording vault override. Copied (and
	// scrubbed) from the analysis monitor stage — credential/secret fields and the
	// individual user id never cross the boundary. Wire-only; the persisted scope
	// is OrganisationId.
	User WorkflowUser `json:"user,omitempty" bson:"-"`

	// Device identifies the recording the run derives from, with the few fields
	// vault-override resolution and logging need (device key/name and where the
	// media is stored/served from). Copied from the recording at hand-off time.
	// Wire-only.
	Device WorkflowDevice `json:"device,omitempty" bson:"-"`

	// Inputs is the immutable start context the run opens with, keyed by the
	// upstream operation that produced it (e.g. "classify" → the classification
	// result). Conditions and stages read upstream context from here; it is set
	// once by analysis and never mutated by the run. Persisted at open so a run
	// reloaded mid-flight still sees its start context.
	Inputs map[string]interface{} `json:"inputs,omitempty" bson:"inputs,omitempty"`

	// Results is the run's accumulated stage outputs, keyed by operation. Each
	// stage worker writes its result under its operation on the way back, and
	// conditions / downstream stages read upstream outputs from here. It grows
	// as the run progresses; the engine records each result into it. Together
	// with Inputs it is the durable condition bag (Results wins on any overlap).
	Results map[string]interface{} `json:"results,omitempty" bson:"results,omitempty"`

	// Payload is the self-describing block envelope a delegated-ingest worker
	// hands back for the platform to persist: one or more typed blocks (e.g. a
	// "detection" block carrying a PostDetectionsRequest, optionally followed by
	// "marker" blocks). It is the channel the shared ingest core reads from,
	// distinct from Results:
	//
	//   - Results is the multi-operation, decoded routing/state ledger the
	//     condition matcher reads and the run persists.
	//   - Payload is one worker's block envelope for a single ingest hop.
	//
	// Lifecycle mirrors Storage and is one-directional (worker → engine):
	//   - A delegated-ingest worker sets Payload on its result; the engine routes
	//     it through ingest.IngestBlocks, persisting each block by its own type
	//     into the platform-owned collection, and mirrors the envelope's blocks
	//     grouped by type into Results[operation] (one array per block type, e.g.
	//     results.<op>.detections) so a downstream conditional stage can test what
	//     the stage produced. The engine targets the run's own recording
	//     (Key/User/Device), so a payload that also carries its own recording
	//     reference (e.g. a PostDetectionsRequest mediaKey/analysisId) has that
	//     reference ignored on the queue path.
	//   - A self-persisting worker writes its own collection and returns its
	//     routing values in Results instead; Payload is empty.
	//
	// `bson:"-"` is load-bearing: the raw body is ingested into its own
	// collection, never duplicated into the run's persisted state. The engine
	// never sets it on an outbound dispatch, so it never travels engine → worker.
	Payload json.RawMessage `json:"payload,omitempty" bson:"-"`

	// Storage carries the credentials a dispatched stage worker needs to fetch
	// the media (global Kerberos Storage plus any resolved per-recording vault
	// override). It is populated by the engine only on the engine→worker
	// dispatch hop and is empty on the analysis hand-off and the worker→engine
	// result. `bson:"-"` is load-bearing: credentials never sit in the run's
	// persisted state.
	Storage *WorkflowStorage `json:"storage,omitempty" bson:"-"`

	// SignedURL is a vault-signed, short-lived URL (HMAC signature + TTL) a
	// dispatched stage worker can use to fetch the run's media directly, instead
	// of constructing the request from the raw Storage credentials. It is the
	// signed URL the upstream pipeline already holds for the recording (the same
	// one carried on PipelinePayload.SignedURL), copied onto the run at the
	// analysis hand-off and carried on the engine→worker dispatch alongside
	// Storage. `bson:"-"` is load-bearing: a signed URL is credential-equivalent
	// and short-lived, so — like Storage — it is wire-only and never lands in the
	// run's persisted state; a stage dispatched after a reload fetches via Storage.
	SignedURL string `json:"signedUrl,omitempty" bson:"-"`

	// Params carries the per-placement parameter values the workflow editor
	// stored for the dispatched stage (WorkflowNode.Data, declared by the
	// stage's WorkflowStage.Params). The engine resolves them from the workflow
	// definition on every engine→worker dispatch, so a worker such as the
	// forwarder can let a tenant override deployment defaults. Values may include
	// StageParamSecret credentials, so `bson:"-"` is load-bearing — like Storage,
	// they never persist in run state — and a worker must never echo them back
	// or copy them into anything it forwards.
	Params map[string]interface{} `json:"params,omitempty" bson:"-"`

	// DispatchedOperations are the operation ids the engine has enqueued for this
	// run — the always-stages seeded at open plus any conditional stages that
	// matched. Every entry is a deployed stage's operation (only stages are ever
	// dispatched), so here stage and operation coincide; the field is named by
	// operation because the stored value is the operation id and to stay
	// symmetric with ResolvedOperations. Written idempotently via $addToSet.
	DispatchedOperations []string `json:"dispatchedOperations,omitempty" bson:"dispatchedoperations,omitempty"`

	// ResolvedOperations are the operation ids whose stage results the engine has
	// observed (each worker hands its result back under its operation). With
	// DispatchedOperations it drives finalization — the run ends once every
	// dispatched operation is resolved — and idempotency.
	//
	// This is narrower than the set a need's gate checks: gate readiness is
	// evaluated against the run's available operations — the keys of Inputs ∪
	// Results, which also include the trigger analysis hands off (e.g. "classify")
	// that seeds Inputs but never resolves as a stage and so never appears here.
	ResolvedOperations []string `json:"resolvedOperations,omitempty" bson:"resolvedoperations,omitempty"`
}

// MarshalJSON is the single place the persisted identity (Id) is projected onto
// the wire identity (RunId), so the two representations can never drift. A value
// ObjectID cannot itself represent "no id yet" in JSON (its zero marshals to the
// all-zero hex, which encoding/json's omitempty does not drop), so the identity
// is carried on the wire by the companion RunId string instead: whenever a run
// that has been opened (its _id is set) is serialized, RunId is emitted as
// Id.Hex(); a run without an Id keeps whatever RunId it was given (normally
// empty, so runId is omitted). Producers therefore only ever set Id — the hex
// projection is automatic and impossible to forget.
//
// The alias type strips WorkflowRun's own MarshalJSON for the inner encode, so
// this does not recurse.
func (r WorkflowRun) MarshalJSON() ([]byte, error) {
	type wire WorkflowRun
	w := wire(r)
	if !r.Id.IsZero() {
		w.RunId = r.Id.Hex()
	}
	var stages *[]WorkflowRunStage
	if r.Stages != nil {
		stages = &r.Stages
	}
	return json.Marshal(struct {
		wire
		Stages *[]WorkflowRunStage `json:"stages,omitempty"`
	}{wire: w, Stages: stages})
}

// LifecycleState derives the coarse, client-facing run state from the raw
// persisted lifecycle fields (Start/End and the dispatched/resolved sets). It
// is the single source of truth for "is this run still working, done, or done
// with nothing to show" so every surface reading a run agrees on the meaning
// without re-implementing the rule:
//
//   - End == 0                                  -> running   (not finalised)
//   - End > 0 && (resolved>0 || Results present) -> completed (finalised, produced)
//   - End > 0 && dispatched==0 && resolved==0    -> noResult  (finalised, no-op)
//
// The lifecycle fields are persistence-only (json:"-"), so this is meant to run
// server-side against a decoded run document; the derived state is what crosses
// the wire.
func (r WorkflowRun) LifecycleState() WorkflowRunState {
	if r.End == 0 {
		return WorkflowRunStateRunning
	}
	if len(r.ResolvedOperations) > 0 || len(r.Results) > 0 || r.HasResults {
		return WorkflowRunStateCompleted
	}
	if len(r.DispatchedOperations) == 0 {
		return WorkflowRunStateNoResult
	}
	// Finalised with stages dispatched but none recorded as resolved. The engine
	// only ends a run once every dispatched op resolves, so this is not expected
	// in practice; treat it as completed (work was done) rather than no-op.
	return WorkflowRunStateCompleted
}

func workflowRunTimestampMilliseconds(timestamp int64) int64 {
	if timestamp > 0 && timestamp < workflowRunMillisecondTimestampThreshold {
		return timestamp * 1000
	}
	return timestamp
}

// PopulateRuntimeFields derives the API lifecycle projection from persisted run
// facts. The receiver should be a read model: none of the derived fields are
// persisted, while legacy second-precision Start and End values are normalized
// to the canonical millisecond representation.
func (r *WorkflowRun) PopulateRuntimeFields(now time.Time) {
	r.Start = workflowRunTimestampMilliseconds(r.Start)
	r.End = workflowRunTimestampMilliseconds(r.End)

	if r.Origin == "" {
		r.Origin = WorkflowOriginAutomatic
	}
	if r.Results != nil {
		r.HasResults = len(r.Results) > 0
	}
	if r.State == "" {
		r.State = r.LifecycleState()
	}

	r.Dispatched = len(r.DispatchedOperations)
	r.Resolved = len(r.ResolvedOperations)
	r.Operations = workflowRunOperationStatuses(r.DispatchedOperations, r.ResolvedOperations)

	endAtMs := r.End
	if endAtMs == 0 && !now.IsZero() {
		endAtMs = now.UnixMilli()
	}
	if r.Start > 0 && endAtMs >= r.Start {
		r.DurationMs = endAtMs - r.Start
	}

	for i := range r.Stages {
		if execution := r.Stages[i].Execution; execution != nil {
			execution.populateRuntimeFields(r.End > 0, endAtMs)
		}
	}
	for i := range r.StageExecutions {
		execution := &r.StageExecutions[i]
		details := execution.details()
		details.populateRuntimeFields(r.End > 0, endAtMs)
		execution.State, execution.DurationMs = details.State, details.DurationMs
	}
}

func workflowRunOperationStatuses(dispatched, resolved []string) []WorkflowRunOperationStatus {
	if len(dispatched) == 0 {
		return nil
	}
	resolvedSet := make(map[string]struct{}, len(resolved))
	for _, operation := range resolved {
		resolvedSet[operation] = struct{}{}
	}
	statuses := make([]WorkflowRunOperationStatus, 0, len(dispatched))
	for _, operation := range dispatched {
		status := WorkflowRunOperationStateDispatched
		if _, ok := resolvedSet[operation]; ok {
			status = WorkflowRunOperationStateResolved
		}
		statuses = append(statuses, WorkflowRunOperationStatus{
			Operation: operation,
			Status:    status,
		})
	}
	return statuses
}

func (e WorkflowRunStageExecution) lifecycleState(runEnded bool) WorkflowRunStageState {
	return e.details().lifecycleState(runEnded)
}

// AutomaticRunObjectID derives the DETERMINISTIC run identity for an automatic
// run of a given workflow over a given recording, from the natural triple
// (media key, organisation, workflow). It is the single source of truth both the
// producer (analysis, which mints it) and any consumer that re-derives it must
// agree on, so the same recording teed into the same workflow always maps to the
// same run: a redelivered pipeline event upserts the one run document rather than
// opening a duplicate. Because it returns a primitive.ObjectID it becomes the
// run's _id, and the wire RunId falls out of it for free via MarshalJSON
// (RunId == Id.Hex()) — the identity is minted once and projected everywhere.
//
// It is intentionally NOT used for manual runs: those mint a fresh ObjectID per
// launch (primitive.NewObjectID) so re-pressing a button opens a new run, while a
// redelivered copy of that one launch still dedupes on its already-minted id.
//
// The triple is hashed with a NUL separator so the field boundaries are
// unambiguous (no two distinct triples can concatenate to the same input), and
// the first 12 bytes of the digest form the ObjectID.
func AutomaticRunObjectID(mediaKey, organisationId, workflowId string) primitive.ObjectID {
	sum := sha256.Sum256([]byte(mediaKey + "\x00" + organisationId + "\x00" + workflowId))
	var id primitive.ObjectID
	copy(id[:], sum[:])
	return id
}

// WorkflowUser carries secret-free source ownership and storage context from
// analysis to the workflow engine.
type WorkflowUser struct {
	OrganisationId string              `json:"organisationId,omitempty"`
	ProjectId      *primitive.ObjectID `json:"projectId,omitempty"`
	Storage        Storage             `json:"storage,omitempty"`
}

// WorkflowDevice is the projection of a recording's device carried on a
// WorkflowRun. It holds the device key/name plus where the media is stored and
// served from — exactly the inputs vault-override resolution and logging need,
// without the rest of the Media/Device documents.
type WorkflowDevice struct {
	DeviceKey       string   `json:"deviceKey,omitempty"`
	DeviceName      string   `json:"deviceName,omitempty"`
	Provider        string   `json:"provider,omitempty"`        // media VideoProvider: where the media is served from
	StorageSolution string   `json:"storageSolution,omitempty"` // media StorageSolution: where the media is stored
	SiteIds         []string `json:"siteIds,omitempty"`         // site ids the device is linked to (Device.SiteIds); a gate value, matchable with contains/in/exists/matches
	GroupIds        []string `json:"groupIds,omitempty"`        // effective group memberships resolved by the caller, never an authorization grant
}

// WorkflowStorage carries the storage credentials a dispatched stage worker
// uses to fetch the media. It pairs the global Kerberos Storage credentials
// with an optional per-recording vault override (so derived artifacts land on
// the same backend as the recording). It only ever travels on the engine→worker
// dispatch hop.
type WorkflowStorage struct {
	Uri       string `json:"uri,omitempty"`
	AccessKey string `json:"accessKey,omitempty"`
	Secret    string `json:"secret,omitempty"`

	VaultOverrideUri       string `json:"vaultOverrideUri,omitempty"`
	VaultOverrideAccessKey string `json:"vaultOverrideAccessKey,omitempty"`
	VaultOverrideSecret    string `json:"vaultOverrideSecret,omitempty"`
	VaultOverrideProvider  string `json:"vaultOverrideProvider,omitempty"`
}

// WorkflowConditionRootMode selects which parts of a run a condition can read.
type WorkflowConditionRootMode string

const (
	// WorkflowConditionRootTrigger is what automatic Start triggers match
	// against when a recording is handed off, before any run exists: the
	// envelope, the run identity scalars and the hand-off Inputs.
	WorkflowConditionRootTrigger WorkflowConditionRootMode = "trigger"
	// WorkflowConditionRootRun is what conditions between stages read during a
	// run: everything in trigger mode plus runId and accumulated Results.
	WorkflowConditionRootRun WorkflowConditionRootMode = "run"
)

// WorkflowConditionPath describes one path a condition can read in the
// projection built by WorkflowRun.ConditionRoot. Array element paths end in
// ".*". Open namespaces (inputs, results) hold one object per operation; the
// platform-owned inputs.classify is described here, other operations by
// contracts. Type is string, number, integer, boolean, or object for an open
// namespace. AutomaticOnly marks paths present only in runs opened by the
// automatic analysis hand-off, not in manually launched runs.
type WorkflowConditionPath struct {
	Path          string                      `json:"path"`
	Type          string                      `json:"type"`
	Modes         []WorkflowConditionRootMode `json:"modes"`
	Open          bool                        `json:"open,omitempty"`
	AutomaticOnly bool                        `json:"automaticOnly,omitempty"`
}

var bothConditionRootModes = []WorkflowConditionRootMode{WorkflowConditionRootTrigger, WorkflowConditionRootRun}

// WorkflowConditionRootSchema lists every path WorkflowRun.ConditionRoot
// exposes. It is the single description of the condition envelope and the
// reference for every field path in workflow contracts:
//
//   - A Start contract field uses an absolute path in this root, for example
//     device.deviceName, or a path inside an open namespace such as
//     inputs.classify.details.*.classified.
//   - A stage contract field is relative to that stage's own namespace,
//     results.<operation>: detections.*.plate in the anpr contract resolves to
//     results.anpr.detections.*.plate.
//
// Envelope paths (device, user, identity scalars) and the platform-owned
// classification hand-off (inputs.classify, see WorkflowClassifyInput) are
// fixed here; the contents of other inputs.<operation> and results.<operation>
// namespaces come from workers and are described by contracts. A field added to WorkflowRun,
// WorkflowDevice or WorkflowUser becomes matchable only when the projection and
// this schema both add it (see the round-trip test).
func WorkflowConditionRootSchema() []WorkflowConditionPath {
	schema := []WorkflowConditionPath{
		{Path: "device.deviceKey", Type: "string", Modes: bothConditionRootModes},
		{Path: "device.deviceName", Type: "string", Modes: bothConditionRootModes},
		{Path: "device.provider", Type: "string", Modes: bothConditionRootModes},
		{Path: "device.storageSolution", Type: "string", Modes: bothConditionRootModes},
		{Path: "device.siteIds.*", Type: "string", Modes: bothConditionRootModes},
		{Path: "device.groupIds.*", Type: "string", Modes: bothConditionRootModes},
		{Path: "user.organisationId", Type: "string", Modes: bothConditionRootModes},
		{Path: "key", Type: "string", Modes: bothConditionRootModes},
		{Path: "operation", Type: "string", Modes: bothConditionRootModes},
		{Path: "traceId", Type: "string", Modes: bothConditionRootModes},
		{Path: "runId", Type: "string", Modes: []WorkflowConditionRootMode{WorkflowConditionRootRun}},
		{Path: "inputs", Type: "object", Modes: bothConditionRootModes, Open: true},
		{Path: "results", Type: "object", Modes: []WorkflowConditionRootMode{WorkflowConditionRootRun}, Open: true},
	}
	return append(schema, structConditionPaths("inputs."+WorkflowClassifyOperation, reflect.TypeOf(WorkflowClassifyInput{}))...)
}

// WorkflowClassifyOperation is the analysis operation whose result opens
// automatic workflow runs and is available as inputs.classify.
const WorkflowClassifyOperation = "classify"

// WorkflowClassifyInput is the classifier result the analysis hand-off carries
// as inputs.classify, exactly as the classifier returns it (analysis stores
// the same document as data.classify). Fields tagged condition:"-" are not
// matchable (coordinate and colour matrices, unused values); condition:"automatic"
// marks fields manual launches do not carry, because they seed inputs.classify
// from the stored analysis through Classify.
type WorkflowClassifyInput struct {
	ObjectCount int                           `json:"objectCount" condition:"automatic"` // tracked objects that moved
	Properties  []string                      `json:"properties"`                        // classes of the objects that moved
	Details     []WorkflowClassifyInputDetail `json:"details"`                           // one entry per tracked object
}

// WorkflowClassifyInputDetail is one tracked object of a classification.
type WorkflowClassifyInputDetail struct {
	Id               string        `json:"id"`
	Classified       string        `json:"classified"` // class name, e.g. car
	Distance         float64       `json:"distance"`
	StaticDistance   float64       `json:"staticDistance"`
	IsStatic         bool          `json:"isStatic"`
	FrameWidth       int           `json:"frameWidth"`
	FrameHeight      int           `json:"frameHeight"`
	Frame            int           `json:"frame"` // first frame the object appeared in
	Frames           []int         `json:"frames" condition:"-"`
	Occurence        int           `json:"occurence"` // frames the object was seen in
	Traject          [][]float64   `json:"traject" condition:"-"`
	TrajectCentroids [][]float64   `json:"trajectCentroids" condition:"-"`
	ColorsBGR        [][][]float64 `json:"colorsBGR" condition:"-"`
	ColorsHLS        [][][]float64 `json:"colorsHLS" condition:"-"`
	ColorsStr        [][]string    `json:"colorsStr" condition:"-"`
	ColorStr         []string      `json:"colorStr"` // primary colour names
	Valid            bool          `json:"valid"`
	W                float64       `json:"w" condition:"-"` // unused by the classifier
	X                float64       `json:"x"`
	Y                float64       `json:"y"`
}

// structConditionPaths lists the matchable paths of a JSON struct: scalars,
// string lists (as ".*" elements) and nested struct lists, skipping fields
// tagged condition:"-". The result follows the struct, so its schema cannot
// drift from the type.
func structConditionPaths(prefix string, kind reflect.Type) []WorkflowConditionPath {
	var paths []WorkflowConditionPath
	for i := 0; i < kind.NumField(); i++ {
		field := kind.Field(i)
		tag := field.Tag.Get("condition")
		if tag == "-" {
			continue
		}
		path := prefix + "." + strings.Split(field.Tag.Get("json"), ",")[0]
		entry := WorkflowConditionPath{Path: path, Modes: bothConditionRootModes, AutomaticOnly: tag == "automatic"}
		fieldType := field.Type
		if fieldType.Kind() == reflect.Slice {
			if fieldType.Elem().Kind() == reflect.Struct {
				paths = append(paths, structConditionPaths(path+".*", fieldType.Elem())...)
				continue
			}
			entry.Path, fieldType = path+".*", fieldType.Elem()
		}
		switch fieldType.Kind() {
		case reflect.String:
			entry.Type = "string"
		case reflect.Bool:
			entry.Type = "boolean"
		case reflect.Int, reflect.Int32, reflect.Int64:
			entry.Type = "integer"
		case reflect.Float32, reflect.Float64:
			entry.Type = "number"
		default:
			panic("unsupported condition field " + path)
		}
		paths = append(paths, entry)
	}
	return paths
}

// workflowConditionEnvelope is the device and user part of every condition
// root. Credentials (user.storage) and the project scope are never exposed.
func workflowConditionEnvelope(device WorkflowDevice, user WorkflowUser) map[string]any {
	return map[string]any{
		"device": map[string]any{
			"deviceKey":       device.DeviceKey,
			"deviceName":      device.DeviceName,
			"provider":        device.Provider,
			"storageSolution": device.StorageSolution,
			// Array gate values are widened to []any so the shared evaluator's
			// contains/in/matches array handling applies.
			"siteIds":  StringsToAny(device.SiteIds),
			"groupIds": StringsToAny(device.GroupIds),
		},
		"user": map[string]any{
			"organisationId": user.OrganisationId,
		},
	}
}

// ConditionRoot projects the run onto the credential-free object conditions
// read (see WorkflowCondition for path semantics). Its paths are listed by
// WorkflowConditionRootSchema, which contracts reference. It is a whitelist: Storage,
// SignedURL, Payload, Params, user.storage, the project scope and lifecycle
// bookkeeping are never exposed. Inputs and Results are used as given; callers
// normalize persisted (BSON) values first. runId is the persisted Id, falling
// back to RunId when the run has not been stored.
func (r WorkflowRun) ConditionRoot(mode WorkflowConditionRootMode) map[string]any {
	root := workflowConditionEnvelope(r.Device, r.User)
	root["key"], root["operation"], root["traceId"] = r.Key, r.Operation, r.TraceId
	if mode != WorkflowConditionRootRun {
		if len(r.Inputs) > 0 {
			root["inputs"] = r.Inputs
		}
		return root
	}
	root["inputs"], root["results"] = r.Inputs, r.Results
	if root["inputs"] == nil {
		root["inputs"] = map[string]any{}
	}
	if root["results"] == nil {
		root["results"] = map[string]any{}
	}
	root["runId"] = r.RunId
	if !r.Id.IsZero() {
		root["runId"] = r.Id.Hex()
	}
	return root
}
