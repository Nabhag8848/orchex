package run

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v5"
	"github.com/nabhag8848/orchex/internal/handler"
)

// Stop cancels a run: pending|running|paused → cancelled; already cancelled is idempotent.
func (h *Handler) Stop(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()
	exists, err := h.store.WorkflowRunExists(ctx, id)
	if err != nil {
		return handler.InternalError("failed to stop run")
	}
	if !exists {
		return handler.RunNotFound(id)
	}

	row, err := h.store.StopWorkflowRun(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return handler.Conflict("run cannot be stopped from its current status")
		}
		return handler.InternalError("failed to stop run")
	}
	return c.JSON(http.StatusOK, runFromRow(row))
}
