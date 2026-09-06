package usecases

import (
	"context"
	"errors"
	"identityservice/internal/identity/domain"
	"identityservice/internal/identity/repositories"
	"identityservice/internal/infrastructure/oauth"
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
}

// SignInWithGoogle implements [GoogleAuthUseCase].
func (g *googleAuthUseCase) SignInWithGoogle(ctx context.Context, idToken string, deviceID string) (string, string, error) {

	//verify id token
	googleUser, err := g.googleVerifer.VerifyGoogleToken(ctx, idToken)
	if err != nil {
		return "", "", ErrInvalidGoogleToken
	}
	//cari berdasarkan email
	user, err := g.googleAuthRepo.FindEmail(ctx, googleUser.Email)
	if err != nil && !errors.Is(err, repositories.ErrUserNotFound) {
		return "", "", ErrInternal
	}
	//jika user ditemukan maka akan ke login jika ga ada yg maka harus registrasi

	//user tidak di temukan
	if errors.Is(err, repositories.ErrUserNotFound) {
		//register
		newUser := domain.User{
			Username:   googleUser.DisplayName,
			Email:      googleUser.Email,
			ProviderId: googleUser.ProvideId,
			AvatarUrl:  googleUser.AvatarUrl,
			Provider:   "GOOGLE",
		}

		err = g.googleAuthRepo.CreateAccount(ctx, newUser)
		if err != nil {
			return "", "", ErrInternal
		}
		user = &newUser
	}

	// cek apakah providernya sesuai atau tidak
	if user.Provider != "GOOGLE" {
		return "", "", ErrDifferentAuthProvider
	}
	// buat acces token
	accessToken, err := g.accessToken.GenerateAccessToken(*user)
	if err != nil {
		return "", "", err
	}
	// buat refresh token
	refreshToken, err := g.refreshToken.GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}
	HashRefreshToken := g.refreshToken.GenerateHashToken(refreshToken)
	if err != nil {
		return "", "", err
	}
	// create session
	err = g.sessionRepo.CreateSession(ctx, user.ID, deviceID, HashRefreshToken)
	if err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

type GoogleAuthUseCase interface {
	SignInWithGoogle(ctx context.Context, idToken string, deviceID string) (string, string, error)
}

func NewGoogleAuthUseCase(repo repositories.UserRepository, verifier oauth.GoogleVerifier, accToken TokenGenerator, refresh RefreshTokenGenerator, sessionRepo repositories.SessionRepository) GoogleAuthUseCase {
	return &googleAuthUseCase{
		googleVerifer:  verifier,
		googleAuthRepo: repo,
		accessToken:    accToken,
		refreshToken:   refresh,
		sessionRepo:    sessionRepo,
	}
}
