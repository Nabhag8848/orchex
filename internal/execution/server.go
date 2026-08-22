package execution

import (
	"github.com/labstack/echo/v5"
	"github.com/nabhag8848/orchex/internal/handler"
	"github.com/nabhag8848/orchex/internal/handler/run"
	httpserver "github.com/nabhag8848/orchex/internal/http-server"
)

type Deps struct {
	Runs *run.Handler
}

func NewServer(deps Deps) *echo.Echo {
	e := httpserver.New()

	health := e.Group("/health")
	health.GET("/execution", handler.Health)

	v1 := e.Group("/v1")
	deps.Runs.Register(v1.Group("/runs"))

	return e
}
