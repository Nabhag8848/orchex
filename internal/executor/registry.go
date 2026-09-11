package executor

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

type NodeType string

const (
	NodeTypeStart       NodeType = "start"
	NodeTypeConditional NodeType = "conditional"
	NodeTypeAPI         NodeType = "api"
	NodeTypeFunction    NodeType = "function"
	NodeTypeResponse    NodeType = "response"
)

type Node struct {
	WorkflowVersionID uuid.UUID
	ID                uuid.UUID
	Name              string
	Type              NodeType
	Config            json.RawMessage
}

type Result struct {
	Output        json.RawMessage
	NextEdgeLabel string
	Completed     bool
}

type Executor interface {
	Type() NodeType
	Execute(context.Context, Node, json.RawMessage) (Result, error)
}

type Registry struct {
	executors map[NodeType]Executor
}

func NewRegistry(executors ...Executor) *Registry {
	registered := make(map[NodeType]Executor, len(executors))
	for _, executor := range executors {
		registered[executor.Type()] = executor
	}
	return &Registry{executors: registered}
}

func (r *Registry) Execute(ctx context.Context, node Node, input json.RawMessage) (Result, error) {
	return r.executors[node.Type].Execute(ctx, node, input)
}
