package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// Dispatch is the closed enum of dispatch modes for a workflow stage. It is a
// named string — not a boolean — so further modes (e.g. scheduled or manual)
// can be added without a schema change, and named so the permitted values live
// in the type rather than in a comment. An empty Dispatch defaults to
// DispatchAlways.
//
//	always      — enqueued at workflow start, unconditionally.
//	conditional — enqueued only when one of its upstream dependencies resolves
//	              and that dependency's Condition matches the upstream result.
type Dispatch string

const (
	DispatchAlways      Dispatch = "always"
	DispatchConditional Dispatch = "conditional"
)

// ConditionOp is the closed enum of comparison operators a StageCondition may
// use. Named so the allowed operators are part of the type instead of a comment.
type ConditionOp string

const (
	ConditionOpEq       ConditionOp = "eq"       // equal
	ConditionOpNe       ConditionOp = "ne"       // not equal
	ConditionOpContains ConditionOp = "contains" // field (array element / string substring) contains Value
	ConditionOpIn       ConditionOp = "in"       // field value is one of Value (a list); inverse of contains
	ConditionOpExists   ConditionOp = "exists"   // path present (Value ignored)
	ConditionOpGt       ConditionOp = "gt"       // greater than (numeric)
	ConditionOpGte      ConditionOp = "gte"      // greater than or equal (numeric)
	ConditionOpLt       ConditionOp = "lt"       // less than (numeric)
	ConditionOpLte      ConditionOp = "lte"      // less than or equal (numeric)
)

// StageCondition is a structured predicate evaluated against an upstream
// operation's returned result. No free-form expressions are allowed: a
// condition is a single (path, op, value) triple.
type StageCondition struct {
	Path  string      `json:"path" bson:"path"`   // dot-path into the upstream op's result
	Op    ConditionOp `json:"op" bson:"op"`       // see the ConditionOp consts
	Value any         `json:"value" bson:"value"` // comparison operand (unused for ConditionOpExists)
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

// StageParamType is the closed enum of declared parameter types. It tells the
// authoring UI how to render a parameter and lets a node's Data be validated
// instead of being free-form.
type StageParamType string

const (
	StageParamString  StageParamType = "string"
	StageParamNumber  StageParamType = "number"
	StageParamBoolean StageParamType = "boolean"
	StageParamSelect  StageParamType = "select"
)

// StageParam declares one configurable parameter a stage accepts. The catalog
// stage owns the declaration; a WorkflowNode supplies the per-instance value
// under the same Name in its Data map (see WorkflowNode.Data). Declaring params
// here is what gives Data a schema to validate against and default from.
type StageParam struct {
	Name     string         `json:"name" bson:"name"`
	Label    string         `json:"label,omitempty" bson:"label,omitempty"`
	Type     StageParamType `json:"type" bson:"type"`
	Required bool           `json:"required,omitempty" bson:"required,omitempty"`
	// Default is applied when a node supplies no value for this parameter.
	Default any `json:"default,omitempty" bson:"default,omitempty"`
	// Options enumerates the permitted values when Type is StageParamSelect.
	Options []string `json:"options,omitempty" bson:"options,omitempty"`
}

// StagePort is a named connection point on a stage. Outputs are the results a
// downstream edge can read (selected by WorkflowEdge.SourcePort, and the space a
// StageCondition.Path indexes into); inputs are the slots an upstream edge feeds
// (selected by WorkflowEdge.TargetPort). A stage that declares no ports exposes a
// single implicit default port, so simple linear stages need none.
type StagePort struct {
	Name  string `json:"name" bson:"name"`
	Label string `json:"label,omitempty" bson:"label,omitempty"`
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
// It carries three groups of fields:
//
//   - Routing — how the orchestrator dispatches and resolves the stage
//     (operation, dispatch, needs). Routing has a single owner: for a user
//     workflow the graph's edges are authoritative and Dispatch/Needs is the
//     compiled projection of them (each incoming edge becomes a Needs entry —
//     its upstream operation plus that edge's optional condition — and Dispatch
//     is conditional when there is at least one), so these fields are derived
//     and must not be hand-edited on a shared entry; for a platform-defined
//     stage the projection is authored chart-side and is global. Operation is
//     unique and binds the stage's queue and resolution.
//   - Contract — what the stage accepts and exposes (params, inputs, outputs).
//     Params give WorkflowNode.Data a schema; Inputs/Outputs are the named
//     ports workflow edges attach to.
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
	// Dispatch is when the stage runs: DispatchAlways or DispatchConditional (the
	// closed Dispatch enum). Empty defaults to DispatchAlways. For a user
	// workflow it is derived from the graph's edges (conditional when the node has
	// at least one incoming edge); see the Routing note above.
	Dispatch Dispatch `json:"dispatch,omitempty" bson:"dispatch,omitempty"`
	// Needs lists the upstream dependencies of a conditional stage — its fan-in.
	// It is the compiled, runtime-authoritative projection of the workflow's
	// incoming edges (one entry per edge), not an independently editable field. At
	// least one entry is required when Dispatch is DispatchConditional; ignored
	// otherwise. The runtime re-evaluates the stage whenever any listed upstream
	// resolves, and the stage fires for the first dependency whose Condition
	// matches that upstream's result (a nil Condition matches unconditionally).
	Needs []StageDependency `json:"needs,omitempty" bson:"needs,omitempty"`

	// --- Contract ---

	// Params declares the configurable parameters this stage accepts. A node's
	// Data is validated against and defaulted from these (see WorkflowNode.Data);
	// empty means the stage takes no parameters.
	Params []StageParam `json:"params,omitempty" bson:"params,omitempty"`
	// Inputs and Outputs declare the stage's named ports — the connection points
	// workflow edges attach to (WorkflowEdge.TargetPort / SourcePort). Empty means
	// a single implicit default port.
	Inputs  []StagePort `json:"inputs,omitempty" bson:"inputs,omitempty"`
	Outputs []StagePort `json:"outputs,omitempty" bson:"outputs,omitempty"`

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
