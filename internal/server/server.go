package server

import (
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/nabhag8848/orchex/internal/handler"
)

type Deps struct {
	Workflows *handler.WorkflowHandler
}

func NewServer(deps Deps) *echo.Echo {
	e := echo.New()
	e.Use(middleware.RequestID())
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	e.GET("/health/builder", handler.Health)
	e.GET("/v1/workflows/:id", deps.Workflows.Get)

	return e
}
