package workflow

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/nabhag8848/orchex/internal/db"
	sqlcdb "github.com/nabhag8848/orchex/internal/db/sqlc"
	"github.com/nabhag8848/orchex/internal/handler"
)

type Handler struct {
	store *db.Store
}

func New(store *db.Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) Register(g *echo.Group) {
	g.POST("", h.Create)
	g.GET("", h.List)
	g.GET("/:id", h.Get)
	g.PUT("/:id", h.Update)
	g.POST("/:id/publish", h.Publish)
	g.DELETE("/:id", h.Archive)
}

func (h *Handler) List(c *echo.Context) error {
	rows, err := h.store.ListWorkflows(c.Request().Context())
	if err != nil {
		return handler.InternalError("failed to list workflows")
	}

	items := make([]Workflow, 0, len(rows))
	for _, row := range rows {
		items = append(items, workflowFromList(row))
	}
	return c.JSON(http.StatusOK, WorkflowList{Items: items})
}

func (h *Handler) Create(c *echo.Context) error {
	var req CreateWorkflowRequest
	if err := handler.BindJSON(c, &req); err != nil {
		return err
	}

	wf, err := h.store.CreateWorkflow(c.Request().Context(), sqlcdb.CreateWorkflowParams{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		return handler.InternalError("failed to create workflow")
	}

	return c.JSON(http.StatusCreated, createdWorkflow(wf))
}

func (h *Handler) Get(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}

	published, err := publishedRequested(c.QueryParam("version"))
	if err != nil {
		return handler.BadRequest(err.Error())
	}

	detail, err := getDetail(c.Request().Context(), h.store.Queries, id, published)
	if err != nil {
		return handler.MapQueryError(id, err, "failed to get workflow")
	}
	return c.JSON(http.StatusOK, detail)
}

func (h *Handler) Update(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()
	exists, err := h.store.WorkflowExists(ctx, id)
	if err != nil {
		return handler.InternalError("failed to get workflow")
	}
	if !exists {
		return handler.NotFound(id)
	}

	var req UpdateWorkflowRequest
	if err := handler.BindJSON(c, &req); err != nil {
		return err
	}
	req = req.withEmptyGraph()

	types, err := loadNodeTypes(ctx, h.store.Queries)
	if err != nil {
		return handler.InternalError("failed to load node types")
	}
	if err := validateGraph(req.Nodes, req.Edges, types); err != nil {
		return handler.BadRequest(err.Error())
	}

	detail, err := h.save(ctx, id, req, types)
	if err != nil {
		return handler.MapQueryError(id, err, "failed to update workflow")
	}
	return c.JSON(http.StatusOK, detail)
}

func (h *Handler) Publish(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}

	result, err := h.publish(c.Request().Context(), id)
	if err != nil {
		var verr *graphValidationError
		if errors.As(err, &verr) {
			return handler.BadRequest(verr.Error())
		}
		return handler.MapQueryError(id, err, "failed to publish workflow")
	}
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) Archive(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}

	result, err := h.store.ArchiveWorkflow(c.Request().Context(), id)
	if err != nil {
		return handler.InternalError("failed to archive workflow")
	}
	if !result.Found {
		return handler.NotFound(id)
	}
	return c.NoContent(http.StatusNoContent)
}

func parseID(c *echo.Context) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return uuid.Nil, handler.BadRequest("invalid workflow id")
	}
	return id, nil
}
