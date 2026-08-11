package handler

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v5"
	sqlcdb "github.com/nabhag8848/orchex/internal/db/sqlc"
)

type WorkflowHandler struct {
	q *sqlcdb.Queries
}

func NewWorkflowHandler(q *sqlcdb.Queries) *WorkflowHandler {
	return &WorkflowHandler{q: q}
}

func (h *WorkflowHandler) Get(c *echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid workflow id"})
	}

	wf, err := h.q.GetWorkflow(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "workflow not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get workflow"})
	}

	return c.JSON(http.StatusOK, wf)
}
