package main

import (
	handler "identityservice/internal/identity/handlers"
	"identityservice/internal/identity/usecases"
	"identityservice/internal/infrastructure/config"
	"identityservice/internal/infrastructure/databases"
	"identityservice/internal/infrastructure/oauth"
	repositories "identityservice/internal/infrastructure/persistence/postgre"
	"identityservice/internal/infrastructure/token"
	appValidator "identityservice/internal/infrastructure/validator"
	"identityservice/internal/router"
	"log"

	"github.com/labstack/echo/v5"
)

func main() {
	e := echo.New()
	e.Validator = appValidator.NewCustomValidator()
	cfg := config.LoadConfig()

	db, err := databases.ConnectDb(cfg.DBUrl)
	if err != nil {
		log.Fatal("failed to connect database:", err)
	}

	jwtProvider := token.NewJWTProvider(
		cfg.JWTSecret,
		cfg.JWTExpiration,
	)

	googleVerifier := oauth.NewGoogleVerifier(cfg.ServerClientId)

	authRepo := repositories.NewUserRepository(db)
	authUsecase := usecases.NewGoogleAuthUseCase(authRepo, *googleVerifier, jwtProvider)
	authHandler := handler.NewGoogleAuthHandler(authUsecase)

	router.Register(e, router.Dependencies{
		GoogleAuthController: authHandler,
	})

	// Start server
	if err := e.Start(":8080"); err != nil {
		log.Fatal(err)
	}

}
