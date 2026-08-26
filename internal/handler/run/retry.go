package run

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v5"
	sqlcdb "github.com/nabhag8848/orchex/internal/db/sqlc"
	"github.com/nabhag8848/orchex/internal/handler"
)

// Retry re-executes the checkpointed node on a failed run (happy path: clear failure + outbox insert).
func (h *Handler) Retry(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()
	exists, err := h.store.WorkflowRunExists(ctx, id)
	if err != nil {
		return handler.InternalError("failed to retry run")
	}
	if !exists {
		return handler.RunNotFound(id)
	}

	var retried sqlcdb.WorkflowRun
	err = h.store.InTx(ctx, func(q *sqlcdb.Queries) error {
		var err error
		retried, err = q.RetryWorkflowRun(ctx, id)
		if err != nil {
			return err
		}
		return q.InsertRunNodeJobOutbox(ctx, sqlcdb.InsertRunNodeJobOutboxParams{
			RunID:             retried.ID,
			WorkflowVersionID: retried.WorkflowVersionID,
			NodeID:            retried.CurrentNodeID,
		})
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return handler.Conflict("run cannot be retried from its current status")
		}
		return handler.InternalError("failed to retry run")
	}
	return c.JSON(http.StatusOK, runFromRow(retried))
}
