package httpserver

import (
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/nabhag8848/orchex/internal/handler"
)

func New() *echo.Echo {
	e := echo.New()
	e.Validator = handler.NewRequestValidator()
	e.HTTPErrorHandler = handler.JSONErrorHandler
	e.Use(middleware.RequestID())
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())
	return e
}
