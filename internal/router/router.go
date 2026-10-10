package router

import (
	"identityservice/docs"
	"identityservice/internal/httpresponse"
	handler "identityservice/internal/identity/handlers"
	"net/http"

	"github.com/labstack/echo/v5"
)

type Dependencies struct {
	OAuthController             *handler.OAuthHandler
	AuthController              *handler.AuthHandler
	RegisterOTPController       *handler.RegisterOTPHandler
	ForgotPasswordOTPController *handler.ForgotPasswordOTPHandler
	RefreshTokenController      *handler.RefreshTokenHandler
	EditProfileHandler          *handler.EditProfileHandler
	FindProfileHandler          *handler.FindProfileHandler
}

func Register(e *echo.Echo, deps Dependencies) {
	docs.Register(e)
	e.GET("/health", health)

	e.POST("/refresh-token", deps.RefreshTokenController.RefreshToken)
	oauth := e.Group("/oauth")
	oauth.POST("/sign-in-google", deps.OAuthController.SignIn)

	auth := e.Group("/auth")
	auth.POST("/sign-in", deps.AuthController.SignIn)
	auth.POST("/sign-up", deps.AuthController.SignUp)
	auth.POST("/register/otp/request", deps.RegisterOTPController.Request)
	auth.POST("/register/otp/verify", deps.RegisterOTPController.Verify)
	auth.GET("/find-email/:email", deps.AuthController.FindEmail)
	auth.POST("/forgot-password/reset", deps.AuthController.ForgotPassword)
	auth.POST("/forgot-password/otp/request", deps.ForgotPasswordOTPController.Request)
	auth.POST("/forgot-password/otp/verify", deps.ForgotPasswordOTPController.Verify)

	user := e.Group("/user")
	user.GET("/me", deps.FindProfileHandler.FindById)
	user.PUT("/edit", deps.EditProfileHandler.EditProfile)
}

func health(c *echo.Context) error {
	return c.JSON(http.StatusOK, httpresponse.Success(
		map[string]string{"service": "identity-service"},
		"service is healthy",
	))
}
