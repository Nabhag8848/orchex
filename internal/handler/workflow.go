package handler

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/nabhag8848/orchex/internal/db"
	sqlcdb "github.com/nabhag8848/orchex/internal/db/sqlc"
)

type WorkflowHandler struct {
	store *db.Store
}

func NewWorkflowHandler(store *db.Store) *WorkflowHandler {
	return &WorkflowHandler{store: store}
}

func (h *WorkflowHandler) Register(g *echo.Group) {
	g.POST("", h.Create)
	g.GET("", h.List)
	g.GET("/:id", h.Get)
	g.PUT("/:id", h.Update)
}

func (h *WorkflowHandler) List(c *echo.Context) error {
	rows, err := h.store.ListWorkflows(c.Request().Context())
	if err != nil {
		return internalError("failed to list workflows")
	}

	items := make([]Workflow, 0, len(rows))
	for _, row := range rows {
		items = append(items, workflowFromList(row))
	}
	return c.JSON(http.StatusOK, WorkflowList{Items: items})
}

func (h *WorkflowHandler) Create(c *echo.Context) error {
	var req CreateWorkflowRequest
	if err := bindJSON(c, &req); err != nil {
		return err
	}

	wf, err := h.store.CreateWorkflow(c.Request().Context(), sqlcdb.CreateWorkflowParams{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		return internalError("failed to create workflow")
	}

	return c.JSON(http.StatusCreated, createdWorkflow(wf))
}

func (h *WorkflowHandler) Get(c *echo.Context) error {
	id, err := workflowID(c)
	if err != nil {
		return err
	}

	published, err := publishedRequested(c.QueryParam("version"))
	if err != nil {
		return badRequest(err.Error())
	}

	detail, err := getDetail(c.Request().Context(), h.store.Queries, id, published)
	if err != nil {
		return mapQueryError(id, err, "failed to get workflow")
	}
	return c.JSON(http.StatusOK, detail)
}

func (h *WorkflowHandler) Update(c *echo.Context) error {
	id, err := workflowID(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()
	exists, err := h.store.WorkflowExists(ctx, id)
	if err != nil {
		return internalError("failed to get workflow")
	}
	if !exists {
		return notFound(id)
	}

	var req UpdateWorkflowRequest
	if err := bindJSON(c, &req); err != nil {
		return err
	}
	req = req.withEmptyGraph()

	types, err := loadNodeTypes(ctx, h.store.Queries)
	if err != nil {
		return internalError("failed to load node types")
	}
	if err := validateGraph(req.Nodes, req.Edges, types); err != nil {
		return badRequest(err.Error())
	}

	detail, err := h.save(ctx, id, req, types)
	if err != nil {
		return mapQueryError(id, err, "failed to update workflow")
	}
	return c.JSON(http.StatusOK, detail)
}
