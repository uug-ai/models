// Package ingest is the transport-agnostic core that receives a producer
// result — from the API or the pipeline queue — and runs the right ordered,
// idempotent sequence of actions for its kind.
//
// "Service" here means a package, not a process: there is no new microservice,
// deployment, queue or network hop. The package is compiled *into* both the
// hub-api and the analyser binaries; each calls Ingest(...) in-process on its
// own request. The package is deliberately infra-free — it only depends on
// types already in the models module and on the sink interfaces declared here,
// so it stays a fast, testable library. Concrete persistence (the Mongo
// upsert, region promotion) is supplied by each app through the Scope sinks.
//
// Routing happens on two independent axes that must not be conflated:
//
//   - Kind  (the operation) — what is this result? detection vs thumbnail vs
//     sprite. Different shapes, different sinks, different action sequences.
//     Routed here, by the handlers registry.
//   - Task  (within a kind) — which flavour? box / anpr / pose inside
//     detection. All share one contract and one collection. Routed *inside*
//     the kind handler (see detection.go).
package ingest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/uug-ai/models/pkg/api"
)

// ErrUnknownKind is returned when no handler is registered for the requested
// operation. A handler-less kind must never reach Ingest — the adapter routes
// a completion to Ingest only when a handler exists; otherwise it is a
// self-persist / generic data.<op> completion that is recorded, not routed.
var ErrUnknownKind = errors.New("ingest: no handler registered for kind")

// Source is the transport/trust axis a result arrived on. It gates the
// optional side-effects of a handler via Action.RunFor — it is NOT the
// detection provenance (that is models.DetectionSource, carried in the
// payload). An API client can never write the database itself, so it is
// always delegated; pipeline is the trusted in-cluster transport.
type Source string

const (
	// SourceAPI is an authenticated external API push. Always delegated; the
	// core write runs, trusted side-effects (e.g. auto-promote to redaction)
	// do not.
	SourceAPI Source = "api"
	// SourcePipeline is an in-cluster pipeline stage handing a result back.
	// Trusted: the full action sequence runs.
	SourcePipeline Source = "pipeline"
)

// Target identifies the recording a result attaches to. The adapter resolves
// the producer's reference (mediaKey / analysisId over the API, or the event's
// fileName over the queue) to this already-resolved identity before calling
// Ingest, so the core never performs a lookup.
type Target struct {
	// Key is the recording's stable key (media.videoFile / analysis.key).
	Key string
	// OrganisationId scopes every write to the owning organisation.
	OrganisationId string
	// DeviceId is denormalised onto stored runs for filtering / cascade cleanup.
	DeviceId string
	// RecordingTimestamp (epoch seconds) is denormalised onto the run so cleanup
	// expires it on the recording's retention clock, not the post time.
	RecordingTimestamp int64
}

// Scope carries the per-request context shared by every action: the transport
// it arrived on and the persistence sinks the app injected. Keeping the sinks
// as interfaces here is what keeps this package infra-free — the concrete
// Mongo implementations live in each app's repository layer.
type Scope struct {
	// Source is the transport/trust axis; gates Action.RunFor.
	Source Source
	// Detections is the sink for the mandatory detection-run upsert.
	Detections DetectionStore
	// Regions is the sink for the optional promote-tracks-to-redaction effect.
	Regions RegionPromoter
}

// Report summarises what a handler stored, mirroring api.PostDetectionsResponse
// so an HTTP adapter can return it directly and a queue adapter can log it.
type Report struct {
	RunId        string
	TracksStored int
	BoxesStored  int
	Rejected     []api.DetectionRejection
	Warnings     []string
}

// Action is one idempotent effect in a handler's ordered sequence. Each action
// owns its own sink — essential because sinks genuinely differ (own collection
// vs enrich-in-place). The dispatcher only orders and gates them.
type Action interface {
	// Name identifies the action for logging / reporting.
	Name() string
	// Apply runs the effect against the already-decoded, typed run. It must be
	// idempotent (keyed upsert / keyed replace) because delivery is at-least-once.
	Apply(ctx context.Context, scope Scope, target Target, run any) error
	// RunFor gates the action by transport/trust. The mandatory persistence
	// action returns true for every source; trusted-only side-effects return
	// true only for SourcePipeline.
	RunFor(source Source) bool
}

// Handler is a kind's pipeline: a Decode step that validates + task-routes +
// normalises the payload once into a typed run, followed by the ordered
// actions that consume that run. The first action is always the mandatory
// persistence; any further actions are the optional, RunFor-gated side-effects.
type Handler struct {
	Kind    string
	Decode  func(scope Scope, target Target, payload json.RawMessage) (run any, report Report, err error)
	Actions []Action
}

// handlers is the kind registry — the one switch the dispatcher routes on.
// Adding a kind is a new entry here plus its Decode/Actions; the dispatcher
// never grows a case.
var handlers = map[string]Handler{
	detectionHandler.Kind: detectionHandler,
	// "thumbnail": thumbnailHandler, "sprite": spriteHandler — migrated from
	// the analyser's hardcoded switch later.
}

// Ingest is the single entry point both transports call. It looks up the
// handler for kind, decodes the payload once, then runs each action in order,
// skipping those the current Source is not trusted for. It owns only this
// shared plumbing — the moment it starts switching on kind to do real work it
// has become the hardcoded switch it replaced.
func Ingest(ctx context.Context, scope Scope, target Target, kind string, payload json.RawMessage) (Report, error) {
	h, ok := handlers[kind]
	if !ok {
		return Report{}, fmt.Errorf("%w: %s", ErrUnknownKind, kind)
	}

	run, report, err := h.Decode(scope, target, payload)
	if err != nil {
		return report, err
	}

	for _, a := range h.Actions {
		if !a.RunFor(scope.Source) {
			continue
		}
		if err := a.Apply(ctx, scope, target, run); err != nil {
			return report, fmt.Errorf("ingest: action %q failed: %w", a.Name(), err)
		}
	}

	return report, nil
}
