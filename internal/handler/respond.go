package handler

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v5"
)

func JSONErrorHandler(c *echo.Context, err error) {
	if r, unwrapErr := echo.UnwrapResponse(c.Response()); unwrapErr == nil && r != nil && r.Committed {
		return
	}

	code := http.StatusInternalServerError
	msg := http.StatusText(code)

	var he *echo.HTTPError
	if errors.As(err, &he) {
		code = he.Code
		if he.Message != "" {
			msg = he.Message
		}
	}

	_ = c.JSON(code, map[string]string{"error": msg})
}

func BadRequest(msg string) error {
	return echo.NewHTTPError(http.StatusBadRequest, msg)
}

func NotFound(id uuid.UUID) error {
	return echo.NewHTTPError(http.StatusNotFound, "workflow "+id.String()+" not found")
}

func InternalError(msg string) error {
	return echo.NewHTTPError(http.StatusInternalServerError, msg)
}

func ServiceUnavailable(msg string) error {
	return echo.NewHTTPError(http.StatusServiceUnavailable, msg)
}

func BindJSON(c *echo.Context, dst any) error {
	if err := echo.BindBody(c, dst); err != nil {
		return BadRequest("invalid json")
	}
	if err := c.Validate(dst); err != nil {
		return BadRequest(err.Error())
	}
	return nil
}

func MapQueryError(id uuid.UUID, err error, fallback string) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return NotFound(id)
	}
	return InternalError(fallback)
}
