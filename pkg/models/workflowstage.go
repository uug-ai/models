package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// Dispatch modes for a workflow stage. Dispatch is a closed string enum: a
// stage's Dispatch field must hold one of these exact values (it is a string
// rather than a boolean so further modes — e.g. scheduled or manual — can be
// added without a schema change). An empty Dispatch defaults to DispatchAlways.
//
//	always      — enqueued at workflow start, unconditionally.
//	conditional — enqueued only when one of its upstream dependencies resolves
//	              and that dependency's Condition matches the upstream result.
const (
	DispatchAlways      = "always"
	DispatchConditional = "conditional"
)

// StageCondition is a structured predicate evaluated against an upstream
// operation's returned result. No free-form expressions are allowed: a
// condition is a single (path, op, value) triple.
type StageCondition struct {
	Path  string `json:"path" bson:"path"`   // dot-path into the upstream op's result
	Op    string `json:"op" bson:"op"`       // eq | ne | contains | exists | gt | lt
	Value any    `json:"value" bson:"value"` // comparison operand (unused for "exists")
}

// StageDependency is one upstream a conditional stage waits on, paired with the
// predicate that must match that upstream's result. It mirrors a single
// incoming workflow edge: Operation is the edge's source stage, and Condition
// is the edge's condition (nil for an unconditional dependency).
type StageDependency struct {
	// Operation is the upstream stage this dependency waits on.
	Operation string `json:"operation" bson:"operation"`
	// Condition is the predicate evaluated against the upstream's result. Nil
	// means the dependency matches as soon as the upstream resolves.
	Condition *StageCondition `json:"condition,omitempty" bson:"condition,omitempty"`
}

// StageResourceList is a compute request/limit pair for a stage's workers,
// mirroring a Kubernetes resource list.
type StageResourceList struct {
	CPU    string `json:"cpu,omitempty" bson:"cpu,omitempty"`
	Memory string `json:"memory,omitempty" bson:"memory,omitempty"`
}

// StageResources describes the compute requests and limits applied to the
// workers that run a stage.
type StageResources struct {
	Requests *StageResourceList `json:"requests,omitempty" bson:"requests,omitempty"`
	Limits   *StageResourceList `json:"limits,omitempty" bson:"limits,omitempty"`
}

// WorkflowStage is a reusable, keyed stage definition — a catalog entry. The
// same stage can be referenced by many workflows: a workflow's canvas nodes
// reference a stage by its Operation key (see WorkflowNode.StageRef) rather
// than embedding a copy, so editing a stage updates every workflow that uses
// it.
//
// A stage is resolved from one of two catalog sources, by Operation key:
//
//   - Platform-defined stages, supplied by the chart via the runtime registry
//     (PIPELINE_STAGE_REGISTRY). These have no Id.
//   - User-defined stages, stored in Mongo and managed through the catalog CRUD
//     below. These have an Id.
//
// It carries two groups of fields:
//
//   - Routing — how the orchestrator dispatches and resolves the stage
//     (operation, dispatch, needs). Routing is compiled from the workflow's
//     edges: each incoming edge becomes a Needs entry (its upstream operation
//     plus that edge's optional condition), and Dispatch is conditional when
//     there is at least one such dependency. Operation is unique and binds the
//     stage's queue and resolution.
//   - Deployment — how the stage's workers are deployed (repository, tag,
//     replicas, queue, resources, …). These describe the running service that
//     consumes the stage's queue.
type WorkflowStage struct {
	// --- Catalog identity ---

	// Id is the catalog entry's Mongo id for user-defined stages. It is empty
	// for platform-defined stages supplied via the runtime registry.
	Id primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	// Name is a human-friendly catalog name shown in the stage library.
	Name string `json:"name,omitempty" bson:"name,omitempty"`
	// Description explains what the stage does, for the catalog UI.
	Description string `json:"description,omitempty" bson:"description,omitempty"`

	// --- Routing ---

	// Operation uniquely identifies the stage and binds its queue, dispatch and
	// resolution. It is the key that workflow nodes reference. Two stages may
	// never share an operation.
	Operation string `json:"operation" bson:"operation"`
	// Dispatch is when the stage runs, as a closed string enum: DispatchAlways
	// or DispatchConditional (see the Dispatch consts). It is a string, not a
	// boolean, so new modes can be added later without a schema change. Empty
	// defaults to DispatchAlways.
	Dispatch string `json:"dispatch,omitempty" bson:"dispatch,omitempty"`
	// Needs lists the upstream dependencies of a conditional stage — its fan-in,
	// compiled one-to-one from the workflow's incoming edges. At least one entry
	// is required when Dispatch is DispatchConditional; ignored otherwise. The
	// runtime re-evaluates the stage whenever any listed upstream resolves, and
	// the stage fires for the first dependency whose Condition matches that
	// upstream's result (a nil Condition matches unconditionally).
	Needs []StageDependency `json:"needs,omitempty" bson:"needs,omitempty"`

	// --- Deployment ---

	// Repository is the container image (without tag) for the stage's workers.
	Repository string `json:"repository,omitempty" bson:"repository,omitempty"`
	// Tag is the image tag deployed for the stage.
	Tag string `json:"tag,omitempty" bson:"tag,omitempty"`
	// PullPolicy is the image pull policy (e.g. "IfNotPresent", "Always").
	PullPolicy string `json:"pullPolicy,omitempty" bson:"pullPolicy,omitempty"`
	// Replicas is the desired number of worker pods for the stage.
	Replicas int `json:"replicas,omitempty" bson:"replicas,omitempty"`
	// Queue is the queue the stage's workers consume from. Defaults to a name
	// derived from Operation when empty.
	Queue string `json:"queue,omitempty" bson:"queue,omitempty"`
	// LogLevel is the worker log verbosity (trace | debug | info | warn | error).
	LogLevel string `json:"logLevel,omitempty" bson:"logLevel,omitempty"`
	// Resources are the compute requests/limits for the stage's workers.
	Resources *StageResources `json:"resources,omitempty" bson:"resources,omitempty"`
	// Env is extra environment passed to the stage's workers.
	Env map[string]string `json:"env,omitempty" bson:"env,omitempty"`
}

// Input / Output types for the user-defined stage catalog

type GetWorkflowStagesInput struct {
	User User `json:"user"`
}

type GetWorkflowStagesOutput struct {
	Stages []WorkflowStage `json:"stages"`
}

type GetWorkflowStageInput struct {
	User    User   `json:"user"`
	StageId string `json:"stage_id"`
}

type GetWorkflowStageOutput struct {
	Stage *WorkflowStage `json:"stage"`
}

type CreateWorkflowStageInput struct {
	User  User          `json:"user"`
	Stage WorkflowStage `json:"stage"`
}

type CreateWorkflowStageOutput struct {
	Stage *WorkflowStage `json:"stage"`
}

type UpdateWorkflowStageInput struct {
	User    User          `json:"user"`
	StageId string        `json:"stage_id"`
	Stage   WorkflowStage `json:"stage"`
}

type UpdateWorkflowStageOutput struct {
	Stage *WorkflowStage `json:"stage"`
}

type DeleteWorkflowStageInput struct {
	User    User   `json:"user"`
	StageId string `json:"stage_id"`
}

type DeleteWorkflowStageOutput struct{}
