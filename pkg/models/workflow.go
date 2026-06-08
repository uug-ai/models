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

// WorkflowTrigger defines what activates a workflow: the device/stream
// selection it applies to and an optional recurring time window. An empty
// trigger leaves the workflow eligible at all times for everything routed to it.
type WorkflowTrigger struct {
	// Selection is the set of devices/streams the workflow applies to.
	Selection string `json:"selection,omitempty" bson:"selection,omitempty"`
	// StartAt and EndAt bound a daily time window (e.g. "08:00"/"18:00").
	StartAt string `json:"startAt,omitempty" bson:"startAt,omitempty"`
	EndAt   string `json:"endAt,omitempty" bson:"endAt,omitempty"`
	// Weekdays restricts the workflow to the listed days of week.
	Weekdays []int `json:"weekdays,omitempty" bson:"weekdays,omitempty"`
}

// Workflow is a user-defined automation graph composed of stage-instance nodes
// and the edges that route between them. Trigger says what activates it; the
// nodes/edges say what runs.
type Workflow struct {
	Id             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name           string             `json:"name" bson:"name,omitempty"`
	Description    string             `json:"description" bson:"description,omitempty"`
	Enabled        bool               `json:"enabled" bson:"enabled"`
	Trigger        *WorkflowTrigger   `json:"trigger,omitempty" bson:"trigger,omitempty"`
	Nodes          []WorkflowNode     `json:"nodes" bson:"nodes"`
	Edges          []WorkflowEdge     `json:"edges" bson:"edges"`
	UserId         string             `json:"user_id" bson:"user_id,omitempty"`
	Username       string             `json:"username" bson:"username,omitempty"`
	OrganisationId string             `json:"organisationId" bson:"organisationId,omitempty"`
	CreatedAt      int64              `json:"created_at" bson:"created_at,omitempty"`
	UpdatedAt      int64              `json:"updated_at" bson:"updated_at,omitempty"`
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
