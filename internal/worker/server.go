package worker

import (
	"github.com/labstack/echo/v5"
	"github.com/nabhag8848/orchex/internal/handler"
	httpserver "github.com/nabhag8848/orchex/internal/http-server"
)

func NewServer() *echo.Echo {
	e := httpserver.New()

	health := e.Group("/health")
	health.GET("/worker", handler.Health)

	return e
}
