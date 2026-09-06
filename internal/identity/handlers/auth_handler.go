package handler

import (
	"identityservice/internal/httpresponse"
	"identityservice/internal/identity/dto/request"
	"identityservice/internal/identity/dto/response"
	"identityservice/internal/identity/usecases"
	"net/http"

	"github.com/labstack/echo/v5"
)

type AuthHandler struct {
	usecase usecases.LocalAuthUseCase
}

func NewAuthHandler(usecase usecases.LocalAuthUseCase) *AuthHandler {
	return &AuthHandler{usecase: usecase}
}

func (h AuthHandler) SignIn(c *echo.Context) error {
	ctx := c.Request().Context()

	payload := new(request.SignInRequest)
	if err := c.Bind(payload); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.Error("invalid payload"))
	}
	if err := c.Validate(payload); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.Error("payload validation failed"))
	}

	accToken, refreshToken, err := h.usecase.SignIn(ctx, *payload)
	if err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.Error("sign in failed"))
	}
	return c.JSON(http.StatusOK, httpresponse.Success(response.AuthResponse{
		AccessToken:  accToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    1800,
	}, "sign in success"))
}

func (h AuthHandler) SignUp(c *echo.Context) error {
	ctx := c.Request().Context()

	payload := new(request.SignUpRequest)
	if err := c.Bind(payload); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.Error("invalid payload"))
	}
	if err := c.Validate(payload); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.Error("payload validation failed"))
	}

	accToken, refreshToken, err := h.usecase.SignUp(ctx, *payload)
	if err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.Error(err.Error()))
	}
	return c.JSON(http.StatusOK, httpresponse.Success(response.AuthResponse{
		AccessToken:  accToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    1800,
	}, "sign up success"))
}
