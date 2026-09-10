package run

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v5"
	"github.com/nabhag8848/orchex/internal/handler"
)

// Pause soft-pauses a run: pending|running → paused; already paused is idempotent.
// Does not touch the outbox or in-flight workers.
func (h *Handler) Pause(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()
	exists, err := h.store.WorkflowRunExists(ctx, id)
	if err != nil {
		return handler.InternalError("failed to pause run")
	}
	if !exists {
		return handler.RunNotFound(id)
	}

	row, err := h.store.PauseWorkflowRun(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return handler.Conflict("run cannot be paused from its current status")
		}
		return handler.InternalError("failed to pause run")
	}
	return c.JSON(http.StatusOK, runFromRow(row))
}
