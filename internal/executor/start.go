package executor

import (
	"context"
	"encoding/json"
)

type StartExecutor struct{}

func NewStartExecutor() *StartExecutor {
	return &StartExecutor{}
}

func (e *StartExecutor) Type() NodeType {
	return NodeTypeStart
}

func (e *StartExecutor) Execute(_ context.Context, _ Node, input json.RawMessage) (Result, error) {
	return Result{Output: input, NextEdgeLabel: "default"}, nil
}
