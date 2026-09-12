package executor

import (
	"context"
	"encoding/json"
)

type ResponseExecutor struct{}

type responseConfig struct {
	StatusCode   int               `json:"status_code"`
	Headers      map[string]string `json:"headers"`
	BodyTemplate string            `json:"body_template"`
}

func NewResponseExecutor() *ResponseExecutor {
	return &ResponseExecutor{}
}

func (e *ResponseExecutor) Type() NodeType {
	return NodeTypeResponse
}

func (e *ResponseExecutor) Execute(_ context.Context, node Node, input json.RawMessage) (Result, error) {
	var config responseConfig
	if err := json.Unmarshal(node.Config, &config); err != nil {
		return Result{}, err
	}
	if config.StatusCode == 0 {
		config.StatusCode = 200
	}

	body, err := mergeBody(config.BodyTemplate, input)
	if err != nil {
		return Result{}, err
	}

	output, err := json.Marshal(map[string]any{
		"data": map[string]any{
			"status_code": config.StatusCode,
			"headers":     config.Headers,
			"body":        body,
		},
	})
	if err != nil {
		return Result{}, err
	}
	return Result{Output: output, Completed: true}, nil
}
