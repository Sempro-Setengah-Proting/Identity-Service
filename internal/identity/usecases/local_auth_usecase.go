package usecases

import (
	"context"
	"errors"
	"identityservice/internal/identity/domain"
	"identityservice/internal/identity/dto/request"
	"identityservice/internal/identity/repositories"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserAlreadyExists        = errors.New("user not found")
	ErrInvalidCredentials       = errors.New("invalid credentials")
	ErrInternalError            = errors.New("Internal error")
	ErrInvalidRegistrationToken = errors.New("invalid registration token")
)

type localAuthUseCase struct {
	txManager    repositories.TransactionManager
	userRepo     repositories.UserRepository
	accessToken  TokenGenerator
	refreshToken RefreshTokenGenerator
	sessionRepo  repositories.SessionRepository
	otpRepo      repositories.RegisterOTPStore
}

// SignIn implements [LocalAuthUseCase].
func (l *localAuthUseCase) SignIn(ctx context.Context, payload request.SignInRequest) (string, string, error) {
	user, err := l.userRepo.FindEmail(ctx, payload.Email)

	if errors.Is(err, repositories.ErrUserNotFound) {
		return "", "", ErrInternalError
	}
	if err != nil {
		return "", "", ErrInternalError
	}
	if user.Provider != "LOCAL" {
		return "", "", ErrDifferentAuthProvider
	}
	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password)); err != nil {
		return "", "", ErrInvalidCredentials
	}
	//cari session
	session, err := l.sessionRepo.FindSessionByUserID(ctx, user.ID)

	if err != nil && !errors.Is(err, repositories.ErrSessionNotFound) {
		return "", "", ErrInvalidCredentials
	}
	//hapus session
	if err == nil {
		err = l.sessionRepo.DeleteSession(ctx, session.ID)
		if err != nil {
			return "", "", ErrInternalError
		}
	}
	//buat token baru
	accessToken, err := l.accessToken.GenerateAccessToken(*user)
	if err != nil {
		return "", "", ErrInternalError
	}
	refreshToken, err := l.refreshToken.GenerateRefreshToken()
	if err != nil {
		return "", "", ErrInternalError
	}
	hashRefreshToken := l.refreshToken.GenerateHashToken(refreshToken)
	// create kembali session
	err = l.sessionRepo.CreateSession(ctx, user.ID, payload.DeviceID, hashRefreshToken)
	if err != nil {
		return "", "", ErrInternalError
	}
	return accessToken, refreshToken, nil
}

// SignUp implements [LocalAuthUseCase].
func (l *localAuthUseCase) SignUp(ctx context.Context, newUser request.SignUpRequest) (string, string, error) {
	//hash regis token
	registrationTokenHash := HashRegistrationToken(newUser.RegistrationToken)
	//cek apakah hash regis token ada di redis
	verifiedEmail, err := l.otpRepo.FindRegistrationProof(ctx, registrationTokenHash)
	if errors.Is(err, repositories.ErrRegistrationTokenNotFound) {
		return "", "", ErrInvalidRegistrationToken
	}
	if err != nil {
		return "", "", ErrInternalError
	}

	if verifiedEmail != newUser.Email {
		return "", "", ErrInvalidRegistrationToken
	}

	user, err := l.userRepo.FindEmail(ctx, newUser.Email)
	if !errors.Is(err, repositories.ErrUserNotFound) {
		return "", "", ErrInternalError
	}
	if err == nil {
		return "", "", ErrUserAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", "", err
	}

	user = &domain.User{
		Email:       newUser.Email,
		Password:    string(hashedPassword),
		PhoneNumber: newUser.PhoneNumber,
		Username:    newUser.Name,
		Provider:    "LOCAL",
	}
	//buat account dan dapatkan user id
	// buat acces token
	accessToken, err := l.accessToken.GenerateAccessToken(*user)
	if err != nil {
		return "", "", err
	}
	// buat refresh token
	refreshToken, err := l.refreshToken.GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}
	hashRefreshToken := l.refreshToken.GenerateHashToken(refreshToken)
	// create session

	l.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		err = l.userRepo.CreateAccount(txCtx, *user)
		if err != nil {
			return err
		}
		err = l.sessionRepo.CreateSession(txCtx, user.ID, newUser.DeviceID, hashRefreshToken)
		if err != nil {
			return err
		}
		return nil
	})

	err = l.otpRepo.DeleteRegisterOTP(
		ctx,
		newUser.Email,
		registrationTokenHash,
	)
	if err != nil {
		return "", "", ErrInternalError
	}

	return accessToken, refreshToken, nil
}

type LocalAuthUseCase interface {
	SignIn(ctx context.Context, payload request.SignInRequest) (string, string, error)
	SignUp(ctx context.Context, newUser request.SignUpRequest) (string, string, error)
}

func NewLocalAuthUseCase(
	txManager repositories.TransactionManager,
	userRepo repositories.UserRepository,
	accessToken TokenGenerator,
	refreshToken RefreshTokenGenerator,
	sessionRepo repositories.SessionRepository,
	otpRepo repositories.RegisterOTPStore,
) LocalAuthUseCase {
	return &localAuthUseCase{
		txManager:    txManager,
		userRepo:     userRepo,
		accessToken:  accessToken,
		refreshToken: refreshToken,
		sessionRepo:  sessionRepo,
		otpRepo:      otpRepo,
	}
}
