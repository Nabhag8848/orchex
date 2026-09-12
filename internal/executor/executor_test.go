package executor

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStartExecutorForwardsInput(t *testing.T) {
	input := json.RawMessage(`{"data":{"payload":{"order_id":"order_1"}}}`)
	result, err := NewStartExecutor().Execute(context.Background(), Node{}, input)
	if err != nil {
		t.Fatalf("execute start: %v", err)
	}
	if string(result.Output) != string(input) || result.NextEdgeLabel != "default" {
		t.Fatalf("unexpected start result: %+v", result)
	}
}

func TestConditionalExecutorChoosesBranch(t *testing.T) {
	node := Node{Config: json.RawMessage(`{"expression":"data.payload.status == 200"}`)}
	input := json.RawMessage(`{"data":{"payload":{"status":200}}}`)

	result, err := NewConditionalExecutor().Execute(context.Background(), node, input)
	if err != nil {
		t.Fatalf("execute conditional: %v", err)
	}
	if result.NextEdgeLabel != "true" {
		t.Fatalf("expected true branch, got %q", result.NextEdgeLabel)
	}
}

func TestAPIExecutorUsesPreviousOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		if string(body) != `{"data":{"payload":"hello"}}` {
			t.Fatalf("unexpected request body: %s", body)
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	node := Node{Config: json.RawMessage(`{"method":"POST","url":"` + server.URL + `"}`)}
	input := json.RawMessage(`{"data":{"payload":"hello"}}`)
	result, err := NewAPIExecutor().Execute(context.Background(), node, input)
	if err != nil {
		t.Fatalf("execute API: %v", err)
	}

	var output map[string]any
	if err := json.Unmarshal(result.Output, &output); err != nil {
		t.Fatalf("decode API output: %v", err)
	}
	if output["data"].(map[string]any)["status"] != float64(http.StatusOK) {
		t.Fatalf("unexpected API output: %s", result.Output)
	}
}

func TestFunctionExecutorInvokesSandbox(t *testing.T) {
	sandboxClient := testFunctionSandbox{}
	node := Node{Config: json.RawMessage(`{"source":"return input.data;","timeout_ms":1000}`)}
	input := json.RawMessage(`{"data":{"value":1}}`)
	result, err := NewFunctionExecutor(sandboxClient).Execute(context.Background(), node, input)
	if err != nil {
		t.Fatalf("execute function: %v", err)
	}
	if string(result.Output) != `{"data":{"value":1}}` {
		t.Fatalf("unexpected function output: %s", result.Output)
	}
}

func TestResponseExecutorCompletesRun(t *testing.T) {
	node := Node{Config: json.RawMessage(`{"status_code":201,"headers":{"X-Request-ID":"request_1"}}`)}
	input := json.RawMessage(`{"data":{"order_id":"order_1"}}`)
	result, err := NewResponseExecutor().Execute(context.Background(), node, input)
	if err != nil {
		t.Fatalf("execute response: %v", err)
	}
	if !result.Completed {
		t.Fatal("response result should complete the run")
	}
}

type testFunctionSandbox struct{}

func (testFunctionSandbox) Invoke(_ context.Context, _ string, input json.RawMessage, _ int) (json.RawMessage, error) {
	return input, nil
}
