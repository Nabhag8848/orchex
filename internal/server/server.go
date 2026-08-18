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
	e.Validator = handler.NewRequestValidator()
	e.HTTPErrorHandler = handler.JSONErrorHandler
	e.Use(middleware.RequestID())
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	health := e.Group("/health")
	health.GET("/builder", handler.Health)

	v1 := e.Group("/v1")
	deps.Workflows.Register(v1.Group("/workflows"))

	return e
}
