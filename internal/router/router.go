package router

import (
	"identityservice/docs"
	"identityservice/internal/httpresponse"
	handler "identityservice/internal/identity/handlers"
	"net/http"

	"github.com/labstack/echo/v5"
)

type Dependencies struct {
	OAuthController        *handler.OAuthHandler
	AuthController         *handler.AuthHandler
	RegisterOTPController  *handler.RegisterOTPHandler
	RefreshTokenController *handler.RefreshTokenHandler
}

func Register(e *echo.Echo, deps Dependencies) {
	docs.Register(e)
	e.GET("/health", health)

	identity := e.Group("/identity")
	identity.POST("/refresh-token", deps.RefreshTokenController.RefreshToken)
	oauth := identity.Group("/oauth")
	oauth.POST("/sign-in-google", deps.OAuthController.SignIn)

	auth := identity.Group("/auth")
	auth.POST("/sign-in", deps.AuthController.SignIn)
	auth.POST("/sign-up", deps.AuthController.SignUp)
	auth.POST("/register/otp/request", deps.RegisterOTPController.Request)
	auth.POST("/register/otp/verify", deps.RegisterOTPController.Verify)
	auth.GET("/find-email/:email", deps.AuthController.FindEmail)
	auth.POST("/forgot-password/reset", deps.AuthController.ForgotPassword)

}

func health(c *echo.Context) error {
	return c.JSON(http.StatusOK, httpresponse.Success(
		map[string]string{"service": "identity-service"},
		"service is healthy",
	))
}
