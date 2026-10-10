package handler

import (
	"fmt"
	"identityservice/internal/httpresponse"
	"identityservice/internal/identity/usecases"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

type FindProfileHandler struct {
	usecases usecases.FindByIdUsecase
}

func NewFindProfileHandler(usecases usecases.FindByIdUsecase) *FindProfileHandler {
	return &FindProfileHandler{usecases: usecases}
}

func (h FindProfileHandler) FindById(c *echo.Context) error {
	ctx := c.Request().Context()
	rawUserId := c.Request().Header.Get("X-User-ID")
	fmt.Print("rawUserId: ", rawUserId)
	userId, err := uuid.Parse(rawUserId)
	if err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.Error("invalid user id"))
	}

	err, userResponse := h.usecases.FindById(ctx, userId)
	if err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.Error(err.Error()))
	}

	return c.JSON(http.StatusOK, httpresponse.Success(userResponse, "Find profile success"))
}
