package validator

import (
	playgroundValidator "github.com/go-playground/validator/v10"
)

type CustomValidator struct {
	validator *playgroundValidator.Validate
}

func NewCustomValidator() *CustomValidator {
	return &CustomValidator{
		validator: playgroundValidator.New(),
	}
}

func (cv *CustomValidator) Validate(i any) error {
	return cv.validator.Struct(i)
}