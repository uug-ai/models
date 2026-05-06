package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// WorkflowNode is a single block placed on the workflow canvas.
type WorkflowNode struct {
	Id        string                 `json:"id" bson:"id"`
	Type      string                 `json:"type" bson:"type"`
	Label     string                 `json:"label" bson:"label,omitempty"`
	X         float64                `json:"x" bson:"x"`
	Y         float64                `json:"y" bson:"y"`
	Selection string                 `json:"selection,omitempty" bson:"selection,omitempty"`
	StartAt   string                 `json:"startAt,omitempty" bson:"startAt,omitempty"`
	EndAt     string                 `json:"endAt,omitempty" bson:"endAt,omitempty"`
	Weekdays  []int                  `json:"weekdays,omitempty" bson:"weekdays,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty" bson:"data,omitempty"`
}

// WorkflowEdge is a connection between two workflow nodes.
type WorkflowEdge struct {
	Id         string `json:"id" bson:"id"`
	Source     string `json:"source" bson:"source"`
	SourcePort string `json:"sourcePort" bson:"sourcePort"`
	Target     string `json:"target" bson:"target"`
	TargetPort string `json:"targetPort" bson:"targetPort"`
}

// Workflow is a user-defined automation graph composed of nodes and edges.
type Workflow struct {
	Id           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name         string             `json:"name" bson:"name,omitempty"`
	Description  string             `json:"description" bson:"description,omitempty"`
	Enabled      bool               `json:"enabled" bson:"enabled"`
	Nodes        []WorkflowNode     `json:"nodes" bson:"nodes"`
	Edges        []WorkflowEdge     `json:"edges" bson:"edges"`
	UserId       string             `json:"user_id" bson:"user_id,omitempty"`
	Username     string             `json:"username" bson:"username,omitempty"`
	MasterUserId string             `json:"master_user_id" bson:"master_user_id,omitempty"`
	CreatedAt    int64              `json:"created_at" bson:"created_at,omitempty"`
	UpdatedAt    int64              `json:"updated_at" bson:"updated_at,omitempty"`
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
