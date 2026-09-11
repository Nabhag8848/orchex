package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

type ConditionalExecutor struct{}

type conditionalConfig struct {
	Expression string `json:"expression"`
}

func NewConditionalExecutor() *ConditionalExecutor {
	return &ConditionalExecutor{}
}

func (e *ConditionalExecutor) Type() NodeType {
	return NodeTypeConditional
}

func (e *ConditionalExecutor) Execute(_ context.Context, node Node, input json.RawMessage) (Result, error) {
	var config conditionalConfig
	if err := json.Unmarshal(node.Config, &config); err != nil {
		return Result{}, err
	}

	var output map[string]any
	if err := json.Unmarshal(input, &output); err != nil {
		return Result{}, err
	}

	branch, err := evaluateCondition(config.Expression, output)
	if err != nil {
		return Result{}, err
	}

	data := output["data"].(map[string]any)
	data["branch"] = branch
	result, err := json.Marshal(output)
	if err != nil {
		return Result{}, err
	}
	return Result{Output: result, NextEdgeLabel: branch}, nil
}

func evaluateCondition(expression string, input map[string]any) (string, error) {
	operator := "=="
	parts := strings.Split(expression, operator)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid conditional expression %q", expression)
	}

	value, err := resolvePath(strings.TrimSpace(parts[0]), input)
	if err != nil {
		return "", err
	}

	var expected any
	if err := json.Unmarshal([]byte(strings.TrimSpace(parts[1])), &expected); err != nil {
		return "", err
	}
	if reflect.DeepEqual(value, expected) {
		return "true", nil
	}
	return "false", nil
}

func resolvePath(path string, input map[string]any) (any, error) {
	var value any = input
	for _, segment := range strings.Split(path, ".") {
		object, ok := value.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid conditional path %q", path)
		}
		value = object[segment]
	}
	return value, nil
}
