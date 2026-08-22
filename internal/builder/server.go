package builder

import (
	"github.com/labstack/echo/v5"
	"github.com/nabhag8848/orchex/internal/handler"
	"github.com/nabhag8848/orchex/internal/handler/workflow"
	httpserver "github.com/nabhag8848/orchex/internal/http-server"
)

type Deps struct {
	Workflows *workflow.Handler
}

func NewServer(deps Deps) *echo.Echo {
	e := httpserver.New()

	health := e.Group("/health")
	health.GET("/builder", handler.Health)

	v1 := e.Group("/v1")
	deps.Workflows.Register(v1.Group("/workflows"))

	return e
}
