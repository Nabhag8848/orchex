package run

import (
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
	g.POST("/:id", h.Start)
}

func (h *Handler) Start(c *echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return handler.BadRequest("invalid workflow id")
	}

	row, err := h.store.GetWorkflowDetail(c.Request().Context(), sqlcdb.GetWorkflowDetailParams{
		ID:        id,
		Published: true,
	})
	if err != nil {
		return handler.MapQueryError(id, err, "failed to start workflow")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"workflow_id":          row.ID,
		"published_version_id": row.LatestPublishedVersionID,
	})
}
