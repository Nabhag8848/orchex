package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type APIExecutor struct{}

type apiConfig struct {
	Method       string            `json:"method"`
	URL          string            `json:"url"`
	Headers      map[string]string `json:"headers"`
	Query        map[string]string `json:"query"`
	BodyTemplate string            `json:"body_template"`
	TimeoutMS    int               `json:"timeout_ms"`
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

	endpoint, err := url.Parse(config.URL)
	if err != nil {
		return Result{}, err
	}
	query := endpoint.Query()
	for key, value := range config.Query {
		query.Set(key, value)
	}
	endpoint.RawQuery = query.Encode()

	timeout := 30 * time.Second
	if config.TimeoutMS != 0 {
		timeout = time.Duration(config.TimeoutMS) * time.Millisecond
	}
	requestContext, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	body, err := mergeBody(config.BodyTemplate, input)
	if err != nil {
		return Result{}, err
	}
	requestBody, err := json.Marshal(body)
	if err != nil {
		return Result{}, err
	}
	httpRequest, err := http.NewRequestWithContext(requestContext, config.Method, endpoint.String(), bytes.NewReader(requestBody))
	if err != nil {
		return Result{}, err
	}
	for key, value := range config.Headers {
		httpRequest.Header.Set(key, value)
	}

	response, err := http.DefaultClient.Do(httpRequest)
	if err != nil {
		return Result{}, err
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return Result{}, err
	}

	output, err := json.Marshal(map[string]any{
		"data": map[string]any{
			"status":  response.StatusCode,
			"headers": flattenHeaders(response.Header),
			"body":    decodeResponseBody(responseBody),
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

func mergeBody(bodyTemplate string, input json.RawMessage) (map[string]any, error) {
	body := make(map[string]any)
	if bodyTemplate != "" {
		if err := json.Unmarshal([]byte(bodyTemplate), &body); err != nil {
			return nil, err
		}
	}

	var envelope struct {
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(input, &envelope); err != nil {
		return nil, err
	}
	for key, value := range envelope.Data {
		body[key] = value
	}
	return body, nil
}
