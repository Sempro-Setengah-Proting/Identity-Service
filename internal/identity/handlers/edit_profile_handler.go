package handler

import (
	"identityservice/internal/httpresponse"
	"identityservice/internal/identity/dto/request"
	"identityservice/internal/identity/usecases"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

type EditProfileHandler struct {
	usecase usecases.EditProfileUseCase
}

func NewEditProfileHandler(usecase usecases.EditProfileUseCase) *EditProfileHandler {
	return &EditProfileHandler{usecase: usecase}
}

func (h EditProfileHandler) EditProfile(c *echo.Context) error {
	ctx := c.Request().Context()
	rawUserId := c.Request().Header.Get("X-User-ID")
	userId, err := uuid.Parse(rawUserId)
	if err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.Error("invalid user id"))
	}

	validator := new(request.EditProfileRequest)
	if err := c.Bind(validator); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.Error("invalid payload"))
	}

	if err := c.Validate(validator); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.Error("payload validation failed"))
	}

	err = h.usecase.EditProfile(ctx, *validator, userId)
	if err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.Error(err.Error()))
	}

	return c.JSON(http.StatusOK, httpresponse.Success(nil, "Edit profile succesfuly"))
}
