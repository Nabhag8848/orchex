package run

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v5"
	sqlcdb "github.com/nabhag8848/orchex/internal/db/sqlc"
	"github.com/nabhag8848/orchex/internal/handler"
)

// Resume continues a paused run from current_node_id (happy path: status flip + outbox insert).
func (h *Handler) Resume(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()
	exists, err := h.store.WorkflowRunExists(ctx, id)
	if err != nil {
		return handler.InternalError("failed to resume run")
	}
	if !exists {
		return handler.RunNotFound(id)
	}

	var resumed sqlcdb.WorkflowRun
	err = h.store.InTx(ctx, func(q *sqlcdb.Queries) error {
		var err error
		resumed, err = q.ResumeWorkflowRun(ctx, id)
		if err != nil {
			return err
		}
		return q.InsertRunNodeJobOutbox(ctx, sqlcdb.InsertRunNodeJobOutboxParams{
			RunID:             resumed.ID,
			WorkflowVersionID: resumed.WorkflowVersionID,
			NodeID:            resumed.CurrentNodeID,
		})
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return handler.Conflict("run cannot be resumed from its current status")
		}
		return handler.InternalError("failed to resume run")
	}
	return c.JSON(http.StatusOK, runFromRow(resumed))
}
