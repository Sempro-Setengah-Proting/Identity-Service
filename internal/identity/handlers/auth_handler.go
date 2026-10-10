package handler

import (
	"errors"
	"identityservice/internal/httpresponse"
	"identityservice/internal/identity/dto/request"
	"identityservice/internal/identity/dto/response"
	"identityservice/internal/identity/usecases"
	"log"
	"net/http"
	"strings"

	playgroundValidator "github.com/go-playground/validator/v10"
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
	log.Println("[IDENTITY] SignIn handler called")

	payload := new(request.SignInRequest)
	if err := c.Bind(payload); err != nil {
		log.Printf("[IDENTITY] SignIn bind error: %v", err)
		return c.JSON(http.StatusBadRequest, httpresponse.Error("invalid sign-in payload: expected email, password and device_id as strings"))
	}
	if err := c.Validate(payload); err != nil {
		log.Printf("[IDENTITY] SignIn validation error: %v", err)
		return writeSignInValidationError(c, err)
	}

	accToken, refreshToken, err := h.usecase.SignIn(ctx, *payload)
	if err != nil {
		log.Printf("[IDENTITY] SignIn error: %v", err)
		return writeSignInError(c, err)
	}
	return c.JSON(http.StatusOK, httpresponse.Success(response.AuthResponse{
		AccessToken:  accToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    1800,
	}, "sign in success"))
}

func writeSignInValidationError(c *echo.Context, err error) error {
	var validationErrors playgroundValidator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return c.JSON(http.StatusInternalServerError, httpresponse.Error("failed to validate sign-in payload"))
	}
	fieldNames := map[string]string{
		"Email": "email", "Password": "password", "DeviceID": "device_id",
	}
	messages := make([]string, 0, len(validationErrors))
	for _, fieldError := range validationErrors {
		field := fieldNames[fieldError.StructField()]
		if field == "" {
			field = fieldError.Field()
		}
		switch fieldError.Tag() {
		case "required":
			messages = append(messages, field+" is required")
		case "email":
			messages = append(messages, field+" must be a valid email address")
		default:
			messages = append(messages, field+" is invalid")
		}
	}
	return c.JSON(http.StatusBadRequest, httpresponse.Error(strings.Join(messages, "; ")))
}

func writeSignInError(c *echo.Context, err error) error {
	switch {
	case errors.Is(err, usecases.ErrUserNotFound):
		return c.JSON(http.StatusBadRequest, httpresponse.Error("no account found for this email"))
	case errors.Is(err, usecases.ErrDifferentAuthProvider):
		return c.JSON(http.StatusBadRequest, httpresponse.Error("email registered with another provider; use the matching sign-in method"))
	case errors.Is(err, usecases.ErrInvalidCredentials):
		return c.JSON(http.StatusBadRequest, httpresponse.Error("incorrect password"))
	case errors.Is(err, usecases.ErrSignInUserLookup):
		return c.JSON(http.StatusInternalServerError, httpresponse.Error(usecases.ErrSignInUserLookup.Error()))
	case errors.Is(err, usecases.ErrSignInPasswordCheck):
		return c.JSON(http.StatusInternalServerError, httpresponse.Error(usecases.ErrSignInPasswordCheck.Error()))
	case errors.Is(err, usecases.ErrSignInSessionLookup):
		return c.JSON(http.StatusInternalServerError, httpresponse.Error(usecases.ErrSignInSessionLookup.Error()))
	case errors.Is(err, usecases.ErrSignInSessionDelete):
		return c.JSON(http.StatusInternalServerError, httpresponse.Error(usecases.ErrSignInSessionDelete.Error()))
	case errors.Is(err, usecases.ErrSignInAccessToken):
		return c.JSON(http.StatusInternalServerError, httpresponse.Error(usecases.ErrSignInAccessToken.Error()))
	case errors.Is(err, usecases.ErrSignInRefreshToken):
		return c.JSON(http.StatusInternalServerError, httpresponse.Error(usecases.ErrSignInRefreshToken.Error()))
	case errors.Is(err, usecases.ErrSignInSessionCreate):
		return c.JSON(http.StatusInternalServerError, httpresponse.Error(usecases.ErrSignInSessionCreate.Error()))
	default:
		return c.JSON(http.StatusInternalServerError, httpresponse.Error("unexpected error during sign in"))
	}
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

func (h AuthHandler) ForgotPassword(c *echo.Context) error {
	ctx := c.Request().Context()
	payload := new(request.ForgotPasswordRequest)
	if err := c.Bind(payload); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.Error("invalid payload"))
	}
	if err := c.Validate(payload); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.Error("payload validation failed"))
	}
	if payload.NewPassword != payload.ConfirmPassword {
		return c.JSON(http.StatusBadRequest, httpresponse.Error("new password and confirm password do not match"))
	}

	err := h.usecase.ForgotPassword(ctx, *payload)
	if err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.Error(err.Error()))
	}
	return c.JSON(http.StatusOK, httpresponse.Success(nil, "password reset success"))
}

func (h AuthHandler) FindEmail(c *echo.Context) error {
	ctx := c.Request().Context()

	email := c.Param("email")
	user, err := h.usecase.FindEmail(ctx, email)
	if err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.Error(err.Error()))
	}

	return c.JSON(http.StatusOK, httpresponse.Success(response.UserResponse{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
		Provider:    user.Provider,
		ProviderID:  user.ProviderId,
		AvatarURL:   user.AvatarUrl,
	}, "email found"))
}
