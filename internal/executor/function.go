package executor

import (
	"context"
	"encoding/json"
)

type FunctionExecutor struct {
	sandbox FunctionSandbox
}

type FunctionSandbox interface {
	Invoke(context.Context, string, json.RawMessage, int) (json.RawMessage, error)
}

type functionConfig struct {
	Source    string `json:"source"`
	TimeoutMS int    `json:"timeout_ms"`
}

func NewFunctionExecutor(sandboxClient FunctionSandbox) *FunctionExecutor {
	return &FunctionExecutor{sandbox: sandboxClient}
}

func (e *FunctionExecutor) Type() NodeType {
	return NodeTypeFunction
}

func (e *FunctionExecutor) Execute(ctx context.Context, node Node, input json.RawMessage) (Result, error) {
	var config functionConfig
	if err := json.Unmarshal(node.Config, &config); err != nil {
		return Result{}, err
	}
	if config.TimeoutMS == 0 {
		config.TimeoutMS = 5000
	}

	output, err := e.sandbox.Invoke(ctx, config.Source, input, config.TimeoutMS)
	if err != nil {
		return Result{}, err
	}
	return Result{Output: output}, nil
}
