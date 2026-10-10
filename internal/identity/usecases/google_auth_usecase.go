package usecases

import (
	"context"
	"errors"
	"identityservice/internal/identity/domain"
	"identityservice/internal/identity/repositories"
	"identityservice/internal/infrastructure/oauth"
	"log"
)

var (
	ErrDifferentAuthProvider = errors.New(
		"email registered with another provider",
	)
	ErrInvalidGoogleToken = errors.New(
		"invalid google token",
	)
	ErrInternal = errors.New(
		"internal error",
	)
)

type googleAuthUseCase struct {
	googleVerifer  oauth.GoogleVerifier
	googleAuthRepo repositories.UserRepository
	accessToken    TokenGenerator
	refreshToken   RefreshTokenGenerator
	sessionRepo    repositories.SessionRepository
	txManager      repositories.TransactionManager
}

// SignInWithGoogle implements [GoogleAuthUseCase].
func (g *googleAuthUseCase) SignInWithGoogle(ctx context.Context, idToken string, deviceID string) (string, string, error) {

	//verify id token
	googleUser, err := g.googleVerifer.VerifyGoogleToken(ctx, idToken)
	if err != nil {
		log.Printf("[GOOGLE AUTH] Validation failed: %v", err)
		return "", "", ErrInvalidGoogleToken
	}
	//cari berdasarkan email
	user, err := g.googleAuthRepo.FindEmail(ctx, googleUser.Email)
	if err != nil && !errors.Is(err, repositories.ErrUserNotFound) {
		return "", "", ErrInternal
	}

	isNewUser := errors.Is(err, repositories.ErrUserNotFound)
	if isNewUser {
		user = &domain.User{
			Username:   googleUser.DisplayName,
			Email:      googleUser.Email,
			ProviderId: googleUser.ProvideId,
			AvatarUrl:  googleUser.AvatarUrl,
			Provider:   "GOOGLE",
		}
	}

	// cek apakah providernya sesuai atau tidak
	if user.Provider != "GOOGLE" {
		return "", "", ErrDifferentAuthProvider
	}
	// buat refresh token
	refreshToken, err := g.refreshToken.GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}
	hashRefreshToken := g.refreshToken.GenerateHashToken(refreshToken)

	var accessToken string
	err = g.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		if isNewUser {
			if createErr := g.googleAuthRepo.CreateAccount(txCtx, user); createErr != nil {
				return createErr
			}
		}

		accessToken, err = g.accessToken.GenerateAccessToken(*user)
		if err != nil {
			return err
		}

		return g.sessionRepo.CreateSession(txCtx, user.ID, deviceID, hashRefreshToken)
	})
	if err != nil {
		return "", "", ErrInternal
	}

	return accessToken, refreshToken, nil
}

type GoogleAuthUseCase interface {
	SignInWithGoogle(ctx context.Context, idToken string, deviceID string) (string, string, error)
}

func NewGoogleAuthUseCase(
	txManager repositories.TransactionManager,
	repo repositories.UserRepository,
	verifier oauth.GoogleVerifier,
	accToken TokenGenerator,
	refresh RefreshTokenGenerator,
	sessionRepo repositories.SessionRepository,
) GoogleAuthUseCase {
	return &googleAuthUseCase{
		googleVerifer:  verifier,
		googleAuthRepo: repo,
		accessToken:    accToken,
		refreshToken:   refresh,
		sessionRepo:    sessionRepo,
		txManager:      txManager,
	}
}
