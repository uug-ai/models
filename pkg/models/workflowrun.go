package models

import (
	"encoding/json"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

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
//	analysis ──WorkflowRun{operation:"event"}──▶ engine
//	engine   ──WorkflowRun{operation:<stage>, storage}──▶ stage worker
//	worker   ──WorkflowRun{operation:<stage>, results}──▶ engine
//
// so each consumer reads and writes the one object instead of reconstructing
// state from a generic bag.
//
// Tag discipline keeps the two representations from bleeding into each other:
//   - `bson:"-"` marks WIRE-ONLY fields (transport role, curated projections,
//     the credential-bearing Storage) so they NEVER persist to Mongo — most
//     importantly Storage, so credentials can never land in run state.
//   - `json:"-"` marks PERSISTENCE-ONLY fields (the run's stored identity and
//     progress tiers) so they never appear on the queue contract.
//   - Fields tagged for both (Key, TraceId, WorkflowId, WorkflowName) are the
//     genuine overlap between message and document.
type WorkflowRun struct {
	// Operation marks the message's role on the workflows queue (wire-only):
	//   - "event": a fresh run hand-off from analysis. It opens the run and
	//     carries the start context in Inputs (e.g. the classification result).
	//   - any other value (e.g. "anpr"): either the engine dispatching that
	//     stage to its worker (Storage populated), or the worker routing its
	//     result back (Results populated). These never collide because the
	//     workflows queue only ever carries the "event" open and worker results —
	//     a dispatch goes to the worker's own queue — so the engine never has to
	//     disambiguate a dispatch from a result.
	Operation string `json:"operation,omitempty" bson:"-"`

	// RunId is the run's identifier on the wire (the hex of the document Id). It
	// is empty on the analysis hand-off — the run is keyed by Key until the
	// engine opens it — and set on every engine→worker dispatch. The persisted
	// identity is Id, so RunId itself is wire-only.
	RunId string `json:"runId,omitempty" bson:"-"`

	// Id is the run document's Mongo identity. Persistence-only — on the wire the
	// run is referenced by RunId (Id.Hex()).
	Id primitive.ObjectID `json:"-" bson:"_id,omitempty"`

	// WorkflowId is the id of the Workflow definition (models.Workflow) this run
	// executes — the authored graph (nodes/edges/trigger) the run is an execution
	// of, as opposed to the global stage registry. Forward-looking: it is set once
	// the engine executes Workflow graphs; while the engine is driven by the flat
	// stage registry it is empty.
	WorkflowId string `json:"workflowId,omitempty" bson:"workflowid,omitempty"`

	// WorkflowName is the human-readable name of that Workflow, carried so a
	// dispatch/result is legible in worker context and logs without a lookup.
	// Populated alongside WorkflowId.
	WorkflowName string `json:"workflowName,omitempty" bson:"workflowname,omitempty"`

	// Key is the media key the run is about: its natural identity, used to load
	// or open the run document. Copied from the recording at hand-off time.
	Key string `json:"key,omitempty" bson:"key"`

	// RecordingTimestamp is the recording's start time (unix seconds), copied
	// from the recording at hand-off time. It is denormalised onto any platform
	// artifact the engine ingests (see Payload) so cleanup expires the artifact
	// on the recording's retention clock rather than the post time. Persisted at
	// open and also echoed on the wire so a worker round-trips it back.
	RecordingTimestamp int64 `json:"recordingTimestamp,omitempty" bson:"recordingtimestamp,omitempty"`

	// UserId scopes the run to an account (the organisation id). Persistence-only:
	// on the wire the same identity travels in the richer User projection.
	UserId string `json:"-" bson:"userid"`

	// TraceId continues the distributed trace across the workflow tail.
	TraceId string `json:"traceId,omitempty" bson:"traceid"`

	// Start and End stamp the run's lifecycle (unix seconds). Persistence-only.
	Start int64 `json:"-" bson:"start"`
	End   int64 `json:"-" bson:"end,omitempty"`

	// User is the curated, secret-free account context a run needs: the account
	// id and organisation for logging/scoping and the account Storage block used
	// to resolve a per-recording vault override. Copied (and scrubbed) from the
	// analysis monitor stage — credential/secret fields never cross the boundary.
	// Wire-only; the persisted scope is UserId.
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

	// Payload is the raw, typed result body a delegated-ingest stage hands back
	// for the platform to persist — the single operation in Operation, in its
	// kind's contract shape (e.g. a detection-run body). It is the channel the
	// shared ingest core reads from, distinct from Results:
	//
	//   - Results is the multi-operation, decoded routing/state ledger the
	//     condition matcher reads and the run persists.
	//   - Payload is a single stage's raw typed bytes for one ingest hop.
	//
	// Lifecycle mirrors Storage and is one-directional (worker → engine):
	//   - A delegated-ingest stage (its WorkflowStage declares an ingest Kind)
	//     sets Payload on its result; the engine routes it through ingest.Ingest
	//     into the stage's platform-owned collection and mirrors its decoded form
	//     into Results so downstream conditions can branch on it.
	//   - A self-persisting stage (no Kind) writes its own collection and returns
	//     its routing values in Results instead; Payload is empty.
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

	// DispatchedOperations are the custom stages the engine has enqueued for this
	// run — the always-stages seeded at open plus any conditional stages that
	// matched. Persistence-only; written idempotently via $addToSet.
	DispatchedOperations []string `json:"-" bson:"dispatchedoperations,omitempty"`

	// ResolvedOperations are the upstream operations whose results the engine has
	// observed (drives conditional dispatch and idempotency). Persistence-only.
	ResolvedOperations []string `json:"-" bson:"resolvedoperations,omitempty"`
}

// WorkflowUser is the secret-free projection of an account carried on a
// WorkflowRun. It exposes only what the workflow tail consumes: identity for
// logging/scoping and the account-level Storage block used to resolve a vault
// override. Every credential/secret/billing field of the full User is
// deliberately omitted so it can never reach the workflows queue or its logs.
type WorkflowUser struct {
	Id             string  `json:"id,omitempty"`
	OrganisationId string  `json:"organisationId,omitempty"`
	Storage        Storage `json:"storage,omitempty"`
}

// WorkflowDevice is the projection of a recording's device carried on a
// WorkflowRun. It holds the device key/name plus where the media is stored and
// served from — exactly the inputs vault-override resolution and logging need,
// without the rest of the Media/Device documents.
type WorkflowDevice struct {
	DeviceKey       string `json:"deviceKey,omitempty"`
	DeviceName      string `json:"deviceName,omitempty"`
	Provider        string `json:"provider,omitempty"`        // media VideoProvider: where the media is served from
	StorageSolution string `json:"storageSolution,omitempty"` // media StorageSolution: where the media is stored
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
