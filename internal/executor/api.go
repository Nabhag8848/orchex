package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

type APIExecutor struct{}

type apiConfig struct {
	Method string `json:"method"`
	URL    string `json:"url"`
}

func NewAPIExecutor() *APIExecutor {
	return &APIExecutor{}
}

func (e *APIExecutor) Type() NodeType {
	return NodeTypeAPI
}

func (e *APIExecutor) Execute(ctx context.Context, node Node, input json.RawMessage) (Result, error) {
	var config apiConfig
	if err := json.Unmarshal(node.Config, &config); err != nil {
		return Result{}, err
	}

	httpRequest, err := http.NewRequestWithContext(ctx, config.Method, config.URL, bytes.NewReader(input))
	if err != nil {
		return Result{}, err
	}

	response, err := http.DefaultClient.Do(httpRequest)
	if err != nil {
		return Result{}, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return Result{}, err
	}

	output, err := json.Marshal(map[string]any{
		"data": map[string]any{
			"status":  response.StatusCode,
			"headers": flattenHeaders(response.Header),
			"body":    decodeResponseBody(body),
		},
	})
	if err != nil {
		return Result{}, err
	}
	return Result{Output: output, NextEdgeLabel: "default"}, nil
}

func flattenHeaders(headers http.Header) map[string]string {
	flattened := make(map[string]string, len(headers))
	for key, values := range headers {
		flattened[key] = strings.Join(values, ",")
	}
	return flattened
}

func decodeResponseBody(body []byte) any {
	var decoded any
	if json.Unmarshal(body, &decoded) == nil {
		return decoded
	}
	return string(body)
}
