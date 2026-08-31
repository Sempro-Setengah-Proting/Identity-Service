package handler

import (
	"identityservice/internal/httpresponse"
	"identityservice/internal/identity/dto/request"
	"identityservice/internal/identity/dto/response"
	"identityservice/internal/identity/usecases"
	"net/http"

	"github.com/labstack/echo/v5"
)


type GoogleAuthHandler struct {
	usecase usecases.GoogleAuthUseCase
}

func NewGoogleAuthHandler(usecase usecases.GoogleAuthUseCase) *GoogleAuthHandler{
	return &GoogleAuthHandler{
		usecase: usecase,
	}
}

func (h GoogleAuthHandler) SignIn(c *echo.Context) error {
	ctx := c.Request().Context()

	validator := new(request.OAuthSignInRequest)
	if err := c.Bind(validator); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.Error("invalid payload"))
	}

	if err := c.Validate(validator); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.Error("payload validation failed"))
	}

	token, err := h.usecase.SignInWithGoogle(ctx, validator.IDToken)
	if err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.Error(err.Error()))
	}
	
	return c.JSON(http.StatusCreated, httpresponse.Success(response.AuthResponse{
		AccessToken: token,
		TokenType: "Bearer",
		ExpiresIn: 1800,
	},"sign in success"))
}