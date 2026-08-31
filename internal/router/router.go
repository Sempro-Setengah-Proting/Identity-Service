package router

import (
handler "identityservice/internal/identity/handlers"

	"github.com/labstack/echo/v5"
)

type Dependencies struct {
	GoogleAuthController *handler.GoogleAuthHandler
}

func Register(e *echo.Echo, deps Dependencies){
	identity := e.Group("/identity")
	api := identity.Group("/auth")
	api.POST("/sign-in-google", deps.GoogleAuthController.SignIn)
}