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
	ErrUserNotFound             = errors.New("User not found")
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
	verifiedEmail, err := l.otpRepo.FindRegistrationProof(ctx, registrationTokenHash, domain.OTPRegister)
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
	// buat refresh token
	refreshToken, err := l.refreshToken.GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}
	hashRefreshToken := l.refreshToken.GenerateHashToken(refreshToken)
	// create session

	var accessToken string
	err = l.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		if createErr := l.userRepo.CreateAccount(txCtx, user); createErr != nil {
			return createErr
		}

		accessToken, err = l.accessToken.GenerateAccessToken(*user)
		if err != nil {
			return err
		}

		return l.sessionRepo.CreateSession(txCtx, user.ID, newUser.DeviceID, hashRefreshToken)
	})
	if err != nil {
		return "", "", ErrInternalError
	}

	_, err = l.otpRepo.ConsumeRegistrationProof(
		ctx,
		registrationTokenHash,
		domain.OTPRegister,
	)
	if err != nil {
		return "", "", ErrInternalError
	}

	return accessToken, refreshToken, nil
}

func (l *localAuthUseCase) FindEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := l.userRepo.FindEmail(ctx, email)
	if errors.Is(err, repositories.ErrUserNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, ErrInternalError
	}
	if user.Provider != "LOCAL" {
		return nil, ErrUserNotFound
	}

	return user, nil
}

func (l *localAuthUseCase) ForgotPassword(ctx context.Context, payload request.ForgotPasswordRequest) error {
	registrationTokenHash := HashRegistrationToken(payload.RegistrationToken)
	verifiedEmail, err := l.otpRepo.ConsumeRegistrationProof(
		ctx,
		registrationTokenHash,
		domain.OTPForgotPassword,
	)
	if errors.Is(err, repositories.ErrRegistrationTokenNotFound) {
		return ErrInvalidRegistrationToken
	}
	if err != nil {
		return ErrInternalError
	}

	user, err := l.userRepo.FindEmail(ctx, verifiedEmail)
	if errors.Is(err, repositories.ErrUserNotFound) {
		return ErrUserNotFound
	}
	if err != nil {
		return ErrInternalError
	}
	if user.Provider != "LOCAL" {
		return ErrUserNotFound
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return ErrInternalError
	}
	err = l.userRepo.ForgetPassword(ctx, verifiedEmail, string(hashedPassword))
	if err != nil {
		return ErrInternalError
	}

	return nil
}

type LocalAuthUseCase interface {
	SignIn(ctx context.Context, payload request.SignInRequest) (string, string, error)
	SignUp(ctx context.Context, newUser request.SignUpRequest) (string, string, error)
	FindEmail(ctx context.Context, email string) (*domain.User, error)
	ForgotPassword(ctx context.Context, payload request.ForgotPasswordRequest) error
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
