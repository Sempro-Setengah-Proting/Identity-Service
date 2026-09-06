package handler

import (
	"identityservice/internal/httpresponse"
	"identityservice/internal/identity/dto/request"
	"identityservice/internal/identity/dto/response"
	"identityservice/internal/identity/usecases"
	"net/http"

	"github.com/labstack/echo/v5"
)

type RefreshTokenHandler struct {
	usecase usecases.RefreshTokenUseCase
}

func NewRefreshTokenHandler(usecase usecases.RefreshTokenUseCase) *RefreshTokenHandler {
	return &RefreshTokenHandler{
		usecase: usecase,
	}
}

func (h RefreshTokenHandler) RefreshToken(c *echo.Context) error {
	ctx := c.Request().Context()

	validator := new(request.RequestRefreshToken)
	if err := c.Bind(validator); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.Error("invalid payload"))
	}

	if err := c.Validate(validator); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.Error("payload validation failed"))
	}

	accToken, refreshToken, err := h.usecase.RefreshToken(ctx, validator.RefreshToken, validator.DeviceID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.Error(err.Error()))
	}

	return c.JSON(http.StatusOK, httpresponse.Success(response.AuthResponse{
		AccessToken:  accToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    1800,
	}, "refresh token success"))
}