package run

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v5"
	"github.com/nabhag8848/orchex/internal/db"
	"github.com/nabhag8848/orchex/internal/handler"
)

type Handler struct {
	store *db.Store
}

func New(store *db.Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) Register(g *echo.Group) {
	g.POST("", h.Start)
	g.GET("/:id", h.Get)
	g.POST("/:id/pause", h.Pause)
}

func (h *Handler) Start(c *echo.Context) error {
	var req StartRunRequest
	if err := handler.BindJSON(c, &req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	exists, err := h.store.WorkflowExists(ctx, req.WorkflowID)
	if err != nil {
		return handler.InternalError("failed to start workflow")
	}
	if !exists {
		return handler.NotFound(req.WorkflowID)
	}

	run, err := h.start(ctx, req.WorkflowID, req)
	if err != nil {
		if errors.Is(err, errMissingStartNode) {
			return handler.InternalError(errMissingStartNode.Error())
		}
		if errors.Is(err, errPublishedWithoutVersion) {
			return handler.InternalError(errPublishedWithoutVersion.Error())
		}
		return handler.MapQueryError(req.WorkflowID, err, "failed to start workflow")
	}
	return c.JSON(http.StatusCreated, run)
}

func (h *Handler) Get(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}

	row, err := h.store.GetWorkflowRun(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return handler.RunNotFound(id)
		}
		return handler.InternalError("failed to get run")
	}
	return c.JSON(http.StatusOK, runFromRow(row))
}

func parseID(c *echo.Context) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return uuid.Nil, handler.BadRequest("invalid run id")
	}
	return id, nil
}
