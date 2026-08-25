package workflow

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestCreateWorkflow(t *testing.T) {
	te := setup(t)

	code, raw := te.do(http.MethodPost, "/v1/workflows", map[string]any{
		"name":        "Create Test",
		"description": "from handler test",
	})
	if code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", code, raw)
	}
	detail := decode[WorkflowDetail](t, raw)
	if detail.Name != "Create Test" {
		t.Fatalf("name=%q", detail.Name)
	}
	if detail.Status != "draft" {
		t.Fatalf("status=%q", detail.Status)
	}
	if detail.LatestPublishedVersionID != nil {
		t.Fatalf("expected nil published pointer")
	}
	if detail.Graph.Version != 1 || detail.Graph.PublishedAt != nil {
		t.Fatalf("graph=%+v", detail.Graph)
	}
	if len(detail.Graph.Nodes) != 0 || len(detail.Graph.Edges) != 0 {
		t.Fatalf("expected empty graph")
	}
}

func TestCreateWorkflowValidation(t *testing.T) {
	te := setup(t)
	code, raw := te.do(http.MethodPost, "/v1/workflows", map[string]any{"name": ""})
	if code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", code, raw)
	}
}

func TestListWorkflows(t *testing.T) {
	te := setup(t)
	created := te.create("List Me", "listed")

	code, raw := te.do(http.MethodGet, "/v1/workflows", nil)
	if code != http.StatusOK {
		t.Fatalf("status=%d body=%s", code, raw)
	}
	list := decode[WorkflowList](t, raw)
	found := false
	for _, item := range list.Items {
		if item.ID == created.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("created workflow %s not in list", created.ID)
	}
}

func TestGetWorkflow(t *testing.T) {
	te := setup(t)
	created := te.create("Get Me", "get")

	code, raw := te.do(http.MethodGet, "/v1/workflows/"+created.ID.String(), nil)
	if code != http.StatusOK {
		t.Fatalf("status=%d body=%s", code, raw)
	}
	detail := decode[WorkflowDetail](t, raw)
	if detail.ID != created.ID || detail.Name != "Get Me" {
		t.Fatalf("detail=%+v", detail)
	}

	code, raw = te.do(http.MethodGet, "/v1/workflows/"+created.ID.String()+"?version=published", nil)
	if code != http.StatusNotFound {
		t.Fatalf("unpublished published-get status=%d body=%s", code, raw)
	}

	code, raw = te.do(http.MethodGet, "/v1/workflows/"+created.ID.String()+"?version=nope", nil)
	if code != http.StatusBadRequest {
		t.Fatalf("bad version status=%d body=%s", code, raw)
	}
}

func TestGetWorkflowNotFound(t *testing.T) {
	te := setup(t)
	code, _ := te.do(http.MethodGet, "/v1/workflows/"+uuid.New().String(), nil)
	if code != http.StatusNotFound {
		t.Fatalf("status=%d", code)
	}
}

func TestUpdateWorkflow(t *testing.T) {
	te := setup(t)
	created := te.create("Update Me", "before")

	payload := linearGraph("Update Me", "after")
	code, raw := te.do(http.MethodPut, "/v1/workflows/"+created.ID.String(), payload)
	if code != http.StatusOK {
		t.Fatalf("status=%d body=%s", code, raw)
	}
	detail := decode[WorkflowDetail](t, raw)
	if detail.Description == nil || *detail.Description != "after" {
		t.Fatalf("description=%v", detail.Description)
	}
	if len(detail.Graph.Nodes) != 3 || len(detail.Graph.Edges) != 2 {
		t.Fatalf("nodes=%d edges=%d", len(detail.Graph.Nodes), len(detail.Graph.Edges))
	}
	if detail.Graph.PublishedAt != nil {
		t.Fatalf("draft should have nil published_at")
	}
}

func TestUpdateWorkflowSoftValidation(t *testing.T) {
	te := setup(t)
	created := te.create("Soft Save", "x")

	code, raw := te.do(http.MethodPut, "/v1/workflows/"+created.ID.String(), incompleteGraph("Soft Save"))
	if code != http.StatusOK {
		t.Fatalf("incomplete draft should save; status=%d body=%s", code, raw)
	}
}

func TestUpdateWorkflowConfigValidation(t *testing.T) {
	te := setup(t)
	created := te.create("Bad Config", "x")

	start, api, resp := uuid.New(), uuid.New(), uuid.New()
	e1, e2 := uuid.New(), uuid.New()
	payload := map[string]any{
		"name": "Bad Config",
		"nodes": []map[string]any{
			{"id": start, "node_type": "start", "name": "Start", "config": map[string]any{}},
			{"id": api, "node_type": "api", "name": "Call API", "config": map[string]any{}},
			{"id": resp, "node_type": "response", "name": "Done", "config": map[string]any{"status_code": 200}},
		},
		"edges": []map[string]any{
			{"id": e1, "from_node_id": start, "to_node_id": api, "label": "default"},
			{"id": e2, "from_node_id": api, "to_node_id": resp, "label": "default"},
		},
	}

	code, raw := te.do(http.MethodPut, "/v1/workflows/"+created.ID.String(), payload)
	if code != http.StatusBadRequest {
		t.Fatalf("invalid api config should 400; status=%d body=%s", code, raw)
	}
	if msg := errMessage(t, raw); !contains(msg, "config_schema") {
		t.Fatalf("error=%q", msg)
	}

	code, raw = te.do(http.MethodPut, "/v1/workflows/"+created.ID.String(), linearGraph("Bad Config", "ok"))
	if code != http.StatusOK {
		t.Fatalf("valid config should save; status=%d body=%s", code, raw)
	}
	detail := decode[WorkflowDetail](t, raw)
	var apiNode *Node
	for i := range detail.Graph.Nodes {
		if detail.Graph.Nodes[i].NodeType == "api" {
			apiNode = &detail.Graph.Nodes[i]
			break
		}
	}
	if apiNode == nil {
		t.Fatal("missing api node")
	}
	var cfg map[string]any
	if err := json.Unmarshal(apiNode.Config, &cfg); err != nil {
		t.Fatalf("decode config: %v", err)
	}
	if cfg["method"] != "GET" || cfg["url"] != "https://example.com/health" {
		t.Fatalf("config not persisted: %v", cfg)
	}
}

func TestPublishWorkflow(t *testing.T) {
	te := setup(t)
	created := te.create("Publish Me", "to publish")

	code, raw := te.do(http.MethodPost, "/v1/workflows/"+created.ID.String()+"/publish", map[string]any{})
	if code != http.StatusBadRequest {
		t.Fatalf("empty publish status=%d body=%s", code, raw)
	}
	if msg := errMessage(t, raw); msg == "" || !contains(msg, "empty") {
		// empty graph error
		if !contains(msg, "empty") {
			t.Fatalf("error=%q", msg)
		}
	}

	code, raw = te.do(http.MethodPut, "/v1/workflows/"+created.ID.String(), incompleteGraph("Publish Me"))
	if code != http.StatusOK {
		t.Fatalf("update status=%d body=%s", code, raw)
	}
	code, raw = te.do(http.MethodPost, "/v1/workflows/"+created.ID.String()+"/publish", map[string]any{})
	if code != http.StatusBadRequest {
		t.Fatalf("bad graph publish status=%d body=%s", code, raw)
	}

	code, raw = te.do(http.MethodPut, "/v1/workflows/"+created.ID.String(), linearGraph("Publish Me", "valid"))
	if code != http.StatusOK {
		t.Fatalf("update status=%d body=%s", code, raw)
	}

	code, raw = te.do(http.MethodPost, "/v1/workflows/"+created.ID.String()+"/publish", map[string]any{})
	if code != http.StatusOK {
		t.Fatalf("publish status=%d body=%s", code, raw)
	}
	published := decode[PublishWorkflowResponse](t, raw)
	if published.Status != "published" {
		t.Fatalf("status=%q", published.Status)
	}
	if published.Description == nil || *published.Description != "valid" {
		t.Fatalf("description=%v", published.Description)
	}
	if published.LatestPublishedVersionID == uuid.Nil || published.PublishedVersion.Version != 1 {
		t.Fatalf("published=%+v", published)
	}
	firstAt := published.LastPublishedAt
	firstVer := published.PublishedVersion.ID

	code, raw = te.do(http.MethodPost, "/v1/workflows/"+created.ID.String()+"/publish", map[string]any{})
	if code != http.StatusOK {
		t.Fatalf("idempotent publish status=%d body=%s", code, raw)
	}
	again := decode[PublishWorkflowResponse](t, raw)
	if !again.LastPublishedAt.Equal(firstAt) || again.PublishedVersion.ID != firstVer {
		t.Fatalf("idempotent changed pointers: %+v vs %+v", published, again)
	}

	code, raw = te.do(http.MethodGet, "/v1/workflows/"+created.ID.String()+"?version=published", nil)
	if code != http.StatusOK {
		t.Fatalf("get published status=%d body=%s", code, raw)
	}
	got := decode[WorkflowDetail](t, raw)
	if got.Graph.PublishedAt == nil {
		t.Fatalf("expected published graph")
	}
}

func TestPublishAfterEditForksVersion(t *testing.T) {
	te := setup(t)
	created := te.create("Fork Publish", "v1")

	code, raw := te.do(http.MethodPut, "/v1/workflows/"+created.ID.String(), linearGraph("Fork Publish", "v1"))
	if code != http.StatusOK {
		t.Fatalf("update=%d %s", code, raw)
	}
	code, raw = te.do(http.MethodPost, "/v1/workflows/"+created.ID.String()+"/publish", map[string]any{})
	if code != http.StatusOK {
		t.Fatalf("publish=%d %s", code, raw)
	}
	v1 := decode[PublishWorkflowResponse](t, raw)

	code, raw = te.do(http.MethodPut, "/v1/workflows/"+created.ID.String(), linearGraph("Fork Publish", "v2"))
	if code != http.StatusOK {
		t.Fatalf("fork update=%d %s", code, raw)
	}
	detail := decode[WorkflowDetail](t, raw)
	if detail.Graph.Version != 2 || detail.Graph.PublishedAt != nil {
		t.Fatalf("expected draft v2, got %+v", detail.Graph)
	}
	if detail.LatestPublishedVersionID == nil || *detail.LatestPublishedVersionID != v1.PublishedVersion.ID {
		t.Fatalf("live pointer should stay on v1")
	}

	code, raw = te.do(http.MethodPost, "/v1/workflows/"+created.ID.String()+"/publish", map[string]any{})
	if code != http.StatusOK {
		t.Fatalf("publish v2=%d %s", code, raw)
	}
	v2 := decode[PublishWorkflowResponse](t, raw)
	if v2.PublishedVersion.Version != 2 {
		t.Fatalf("version=%d", v2.PublishedVersion.Version)
	}
	if v2.LatestPublishedVersionID == v1.PublishedVersion.ID {
		t.Fatalf("live pointer should move to v2")
	}
}

func TestArchiveWorkflow(t *testing.T) {
	te := setup(t)
	created := te.create("Archive Me", "bye")

	code, raw := te.do(http.MethodDelete, "/v1/workflows/"+created.ID.String(), nil)
	if code != http.StatusNoContent {
		t.Fatalf("archive status=%d body=%s", code, raw)
	}

	code, _ = te.do(http.MethodGet, "/v1/workflows/"+created.ID.String(), nil)
	if code != http.StatusNotFound {
		t.Fatalf("get archived status=%d", code)
	}

	code, _ = te.do(http.MethodPost, "/v1/workflows/"+created.ID.String()+"/publish", map[string]any{})
	if code != http.StatusNotFound {
		t.Fatalf("publish archived status=%d", code)
	}

	code, _ = te.do(http.MethodDelete, "/v1/workflows/"+uuid.New().String(), nil)
	if code != http.StatusNotFound {
		t.Fatalf("archive missing status=%d", code)
	}
}

func contains(s, sub string) bool {
	return strings.Contains(s, sub)
}
