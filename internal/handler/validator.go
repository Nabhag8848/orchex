package handler

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
)

type requestValidator struct {
	validate *validator.Validate
}

func NewRequestValidator() echo.Validator {
	return &requestValidator{validate: validator.New()}
}

func (v *requestValidator) Validate(i any) error {
	return v.validate.Struct(i)
}
