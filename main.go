package main

import (
	handler "identityservice/internal/identity/handlers"
	"identityservice/internal/identity/usecases"
	"identityservice/internal/infrastructure/config"
	"identityservice/internal/infrastructure/databases"
	emailInfrastructure "identityservice/internal/infrastructure/email"
	"identityservice/internal/infrastructure/oauth"
	repositories "identityservice/internal/infrastructure/persistence/postgre"
	sessionRepositories "identityservice/internal/infrastructure/persistence/postgre/session"
	userRepositories "identityservice/internal/infrastructure/persistence/postgre/user"
	redisRepositories "identityservice/internal/infrastructure/persistence/redis"
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
	defer db.Close()
	redisClient, err := databases.ConnectRedis(cfg.RedisAddress, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		log.Fatal("failed to connect redis:", err)
	}
	defer redisClient.Close()

	jwtProvider := token.NewJWTProvider(
		cfg.JWTSecret,
		cfg.JWTExpiration,
	)

	googleVerifier := oauth.NewGoogleVerifier(cfg.ServerClientId)
	refreshToken := token.NewRefreshToken()

	sessionRepo := sessionRepositories.NewRepository(db)

	authRepo := userRepositories.NewRepository(db)
	registerOTPStore := redisRepositories.NewRegisterOTPStore(redisClient)
	emailSender := emailInfrastructure.NewMailer(cfg)
	txManager := repositories.NewTransactionRepository(db)
	GoogleAuthUsecase := usecases.NewGoogleAuthUseCase(
		txManager,
		authRepo,
		*googleVerifier,
		jwtProvider,
		refreshToken,
		sessionRepo,
	)
	GoogleAuthHandler := handler.NewOAuthHandler(GoogleAuthUsecase)
	LocalAuthUsecase := usecases.NewLocalAuthUseCase(
		txManager,
		authRepo,
		jwtProvider,
		refreshToken,
		sessionRepo,
		registerOTPStore,
	)
	LocalAuthHandler := handler.NewAuthHandler(LocalAuthUsecase)
	RequestRegisterOTPUsecase := usecases.NewRequestRegisterOTPUseCase(authRepo, registerOTPStore, emailSender)
	VerifyRegisterOTPUsecase := usecases.NewVerifyRegisterOTPUseCase(registerOTPStore)
	RegisterOTPHandler := handler.NewRegisterOTPHandler(RequestRegisterOTPUsecase, VerifyRegisterOTPUsecase)
	RequestForgotPasswordOTPUsecase := usecases.NewRequestForgotPasswordOTPUseCase(authRepo, registerOTPStore, emailSender)
	VerifyForgotPasswordOTPUsecase := usecases.NewVerifyForgotPasswordOTPUseCase(registerOTPStore)
	ForgotPasswordOTPHandler := handler.NewForgotPasswordOTPHandler(RequestForgotPasswordOTPUsecase, VerifyForgotPasswordOTPUsecase)
	RefreshTokenUsecase := usecases.NewRefreshTokenUseCase(sessionRepo, jwtProvider, refreshToken, authRepo)
	RefreshTokenHandler := handler.NewRefreshTokenHandler(RefreshTokenUsecase)

	FindByIdUsecase := usecases.NewFindByIdUsecase(authRepo)
	FindProfileHandler := handler.NewFindProfileHandler(FindByIdUsecase)
	EditProfileUsecase := usecases.NewEditProfileUseCase(authRepo)
	EditProfileHandler := handler.NewEditProfileHandler(EditProfileUsecase)

	router.Register(e, router.Dependencies{
		OAuthController:             GoogleAuthHandler,
		AuthController:              LocalAuthHandler,
		RegisterOTPController:       RegisterOTPHandler,
		ForgotPasswordOTPController: ForgotPasswordOTPHandler,
		RefreshTokenController:      RefreshTokenHandler,
		FindProfileHandler:          FindProfileHandler,
		EditProfileHandler:          EditProfileHandler,
	})

	// Start server
	if err := e.Start(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}

}
