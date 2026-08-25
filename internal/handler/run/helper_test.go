package run

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/nabhag8848/orchex/internal/db"
	"github.com/nabhag8848/orchex/internal/handler"
	"github.com/nabhag8848/orchex/internal/handler/workflow"
)

type testEnv struct {
	t     *testing.T
	echo  *echo.Echo
	store *db.Store
}

func setup(t *testing.T) *testEnv {
	t.Helper()
	loadTestEnv()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set; skipping integration tests")
	}

	pool, err := db.NewPool(t.Context(), dsn)
	if err != nil {
		t.Skipf("database unavailable: %v", err)
	}
	t.Cleanup(pool.Close)

	store := db.NewStore(pool)
	e := echo.New()
	e.Validator = handler.NewRequestValidator()
	e.HTTPErrorHandler = handler.JSONErrorHandler
	workflow.New(store).Register(e.Group("/v1/workflows"))
	New(store).Register(e.Group("/v1/runs"))

	return &testEnv{t: t, echo: e, store: store}
}

func loadTestEnv() {
	if os.Getenv("DATABASE_URL") != "" {
		return
	}
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../.."))
	data, err := os.ReadFile(filepath.Join(root, ".env"))
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || !strings.Contains(line, "=") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		if os.Getenv(k) == "" {
			_ = os.Setenv(k, v)
		}
	}
}

func (te *testEnv) do(method, path string, body any) (int, []byte) {
	te.t.Helper()
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			te.t.Fatalf("marshal: %v", err)
		}
		r = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, r)
	if body != nil {
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	}
	rec := httptest.NewRecorder()
	te.echo.ServeHTTP(rec, req)
	return rec.Code, rec.Body.Bytes()
}

func (te *testEnv) create(name, desc string) workflow.WorkflowDetail {
	te.t.Helper()
	code, raw := te.do(http.MethodPost, "/v1/workflows", map[string]any{
		"name":        name,
		"description": desc,
	})
	if code != http.StatusCreated {
		te.t.Fatalf("create status=%d body=%s", code, raw)
	}
	return decode[workflow.WorkflowDetail](te.t, raw)
}

func (te *testEnv) publishLinear(name string) workflow.PublishWorkflowResponse {
	te.t.Helper()
	created := te.create(name, "start-run")
	code, raw := te.do(http.MethodPut, "/v1/workflows/"+created.ID.String(), linearGraph(name, "valid"))
	if code != http.StatusOK {
		te.t.Fatalf("update status=%d body=%s", code, raw)
	}
	code, raw = te.do(http.MethodPost, "/v1/workflows/"+created.ID.String()+"/publish", map[string]any{})
	if code != http.StatusOK {
		te.t.Fatalf("publish status=%d body=%s", code, raw)
	}
	return decode[workflow.PublishWorkflowResponse](te.t, raw)
}

func (te *testEnv) publishedStartNodeID(workflowID uuid.UUID) uuid.UUID {
	te.t.Helper()
	code, raw := te.do(http.MethodGet, "/v1/workflows/"+workflowID.String()+"?version=published", nil)
	if code != http.StatusOK {
		te.t.Fatalf("get published status=%d body=%s", code, raw)
	}
	detail := decode[workflow.WorkflowDetail](te.t, raw)
	for _, n := range detail.Graph.Nodes {
		if n.NodeType == "start" {
			return n.ID
		}
	}
	te.t.Fatal("published graph has no start node")
	return uuid.Nil
}

func decode[T any](t *testing.T, raw []byte) T {
	t.Helper()
	var out T
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("decode: %v body=%s", err, raw)
	}
	return out
}

func linearGraph(name, desc string) map[string]any {
	start, api, resp := uuid.New(), uuid.New(), uuid.New()
	e1, e2 := uuid.New(), uuid.New()
	return map[string]any{
		"name":        name,
		"description": desc,
		"nodes": []map[string]any{
			{"id": start, "node_type": "start", "name": "Start", "config": map[string]any{}},
			{"id": api, "node_type": "api", "name": "Call API", "config": map[string]any{
				"method": "GET",
				"url":    "https://example.com/health",
			}},
			{"id": resp, "node_type": "response", "name": "Done", "config": map[string]any{
				"status_code": 200,
			}},
		},
		"edges": []map[string]any{
			{"id": e1, "from_node_id": start, "to_node_id": api, "label": "default"},
			{"id": e2, "from_node_id": api, "to_node_id": resp, "label": "default"},
		},
	}
}
