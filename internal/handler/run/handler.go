package run

import (
	"errors"
	"net/http"

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
