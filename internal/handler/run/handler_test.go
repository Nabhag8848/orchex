package run

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/nabhag8848/orchex/internal/handler/workflow"
)

func TestMarshalStartOutput(t *testing.T) {
	got, err := marshalStartOutput(json.RawMessage(`{"order_id":"1"}`))
	if err != nil {
		t.Fatal(err)
	}
	want := `{"data":{"payload":{"order_id":"1"}}}`
	if string(got) != want {
		t.Fatalf("got=%s want=%s", got, want)
	}

	got, err = marshalStartOutput(json.RawMessage(`null`))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"data":{"payload":null}}` {
		t.Fatalf("null payload got=%s", got)
	}
}

func TestStartRun(t *testing.T) {
	te := setup(t)
	published := te.publishLinear("Start Run")
	startID := te.publishedStartNodeID(published.ID)

	code, raw := te.do(http.MethodPost, "/v1/runs/"+published.ID.String(), map[string]any{
		"payload": map[string]any{"order_id": "1"},
	})
	if code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", code, raw)
	}
	run := decode[Run](t, raw)
	if run.WorkflowID != published.ID {
		t.Fatalf("workflow_id=%s", run.WorkflowID)
	}
	if run.WorkflowVersionID != published.LatestPublishedVersionID {
		t.Fatalf("version=%s want=%s", run.WorkflowVersionID, published.LatestPublishedVersionID)
	}
	if run.Status != "pending" || run.TriggerType != "manual" {
		t.Fatalf("status=%s trigger=%s", run.Status, run.TriggerType)
	}
	if run.CurrentNodeID != startID || run.CurrentNodeAttempt != 1 {
		t.Fatalf("checkpoint node=%s attempt=%d", run.CurrentNodeID, run.CurrentNodeAttempt)
	}
	if run.StartedAt != nil || run.Error != nil {
		t.Fatalf("pending run should have nil started_at and error")
	}
	var lastOutput map[string]any
	if err := json.Unmarshal(run.LastOutput, &lastOutput); err != nil {
		t.Fatalf("last_output: %v", err)
	}
	data, _ := lastOutput["data"].(map[string]any)
	payload, _ := data["payload"].(map[string]any)
	if payload["order_id"] != "1" {
		t.Fatalf("last_output=%s", run.LastOutput)
	}

	jobs, err := te.store.ListRunNodeJobsOutboxByRun(t.Context(), run.ID)
	if err != nil {
		t.Fatalf("list outbox: %v", err)
	}
	if len(jobs) != 1 {
		t.Fatalf("outbox rows=%d", len(jobs))
	}
	job := jobs[0]
	if job.RunID != run.ID || job.WorkflowVersionID != run.WorkflowVersionID {
		t.Fatalf("outbox identity=%+v", job)
	}
	if job.NodeID != startID || job.Attempt != 1 || job.AvailableAt != nil {
		t.Fatalf("outbox job=%+v", job)
	}
}

func TestStartRunPinsPublishedVersion(t *testing.T) {
	te := setup(t)
	created := te.create("Pin Published", "v1")
	code, raw := te.do(http.MethodPut, "/v1/workflows/"+created.ID.String(), linearGraph("Pin Published", "v1"))
	if code != http.StatusOK {
		t.Fatalf("update=%d %s", code, raw)
	}
	code, raw = te.do(http.MethodPost, "/v1/workflows/"+created.ID.String()+"/publish", map[string]any{})
	if code != http.StatusOK {
		t.Fatalf("publish=%d %s", code, raw)
	}
	v1 := decode[workflow.PublishWorkflowResponse](t, raw)

	code, raw = te.do(http.MethodPut, "/v1/workflows/"+created.ID.String(), linearGraph("Pin Published", "v2"))
	if code != http.StatusOK {
		t.Fatalf("fork update=%d %s", code, raw)
	}

	startID := te.publishedStartNodeID(created.ID)
	code, raw = te.do(http.MethodPost, "/v1/runs/"+created.ID.String(), map[string]any{
		"payload": "hello",
	})
	if code != http.StatusCreated {
		t.Fatalf("start=%d %s", code, raw)
	}
	run := decode[Run](t, raw)
	if run.WorkflowVersionID != v1.LatestPublishedVersionID {
		t.Fatalf("pinned %s want %s", run.WorkflowVersionID, v1.LatestPublishedVersionID)
	}
	if run.CurrentNodeID != startID {
		t.Fatalf("start node %s want %s", run.CurrentNodeID, startID)
	}
	if string(run.LastOutput) != `{"data":{"payload":"hello"}}` {
		t.Fatalf("last_output=%s", run.LastOutput)
	}
}

func TestStartRunUnpublished(t *testing.T) {
	te := setup(t)
	created := te.create("Draft Run", "nope")

	code, raw := te.do(http.MethodPost, "/v1/runs/"+created.ID.String(), map[string]any{})
	if code != http.StatusBadRequest {
		t.Fatalf("existing draft missing payload status=%d body=%s", code, raw)
	}

	code, _ = te.do(http.MethodPost, "/v1/runs/"+created.ID.String(), map[string]any{
		"payload": map[string]any{},
	})
	if code != http.StatusNotFound {
		t.Fatalf("draft start status=%d", code)
	}

	code, raw = te.do(http.MethodPut, "/v1/workflows/"+created.ID.String(), linearGraph("Draft Run", "still draft"))
	if code != http.StatusOK {
		t.Fatalf("update=%d %s", code, raw)
	}
	code, _ = te.do(http.MethodPost, "/v1/runs/"+created.ID.String(), map[string]any{
		"payload": map[string]any{},
	})
	if code != http.StatusNotFound {
		t.Fatalf("unpublished graph start status=%d", code)
	}
}

func TestStartRunArchived(t *testing.T) {
	te := setup(t)
	published := te.publishLinear("Archive Then Start")

	code, raw := te.do(http.MethodDelete, "/v1/workflows/"+published.ID.String(), nil)
	if code != http.StatusNoContent {
		t.Fatalf("archive status=%d body=%s", code, raw)
	}

	code, _ = te.do(http.MethodPost, "/v1/runs/"+published.ID.String(), map[string]any{
		"payload": map[string]any{},
	})
	if code != http.StatusNotFound {
		t.Fatalf("archived start status=%d", code)
	}
}

func TestStartRunNotFound(t *testing.T) {
	te := setup(t)
	missing := uuid.New().String()

	code, _ := te.do(http.MethodPost, "/v1/runs/"+missing, map[string]any{})
	if code != http.StatusNotFound {
		t.Fatalf("missing workflow without payload status=%d", code)
	}

	code, _ = te.do(http.MethodPost, "/v1/runs/"+missing, map[string]any{
		"payload": true,
	})
	if code != http.StatusNotFound {
		t.Fatalf("status=%d", code)
	}
}

func TestStartRunValidation(t *testing.T) {
	te := setup(t)
	published := te.publishLinear("Start Validation")

	code, _ := te.do(http.MethodPost, "/v1/runs/not-a-uuid", map[string]any{
		"payload": map[string]any{},
	})
	if code != http.StatusBadRequest {
		t.Fatalf("invalid id status=%d", code)
	}

	code, raw := te.do(http.MethodPost, "/v1/runs/"+published.ID.String(), map[string]any{})
	if code != http.StatusBadRequest {
		t.Fatalf("missing payload status=%d body=%s", code, raw)
	}

	code, raw = te.do(http.MethodPost, "/v1/runs/"+published.ID.String(), map[string]any{
		"payload": nil,
	})
	if code != http.StatusCreated {
		t.Fatalf("null payload status=%d body=%s", code, raw)
	}
	run := decode[Run](t, raw)
	if string(run.LastOutput) != `{"data":{"payload":null}}` {
		t.Fatalf("last_output=%s", run.LastOutput)
	}
}
