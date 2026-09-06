package usecases

import (
	"context"
	"errors"
	"identityservice/internal/identity/repositories"
	"time"
)

var (
	ErrInvalidSession      = errors.New("invalid session")
	ErrRefreshTokenExpired = errors.New("refresh token expired")
)

type refreshTokenUsecase struct {
	sessionRepo  repositories.SessionRepository
	accToken     TokenGenerator
	refreshToken RefreshTokenGenerator
	userRepo     repositories.UserRepository
}

type RefreshTokenUseCase interface {
	RefreshToken(ctx context.Context, refreshToken string, deviceID string) (string, string, error)
}

func NewRefreshTokenUseCase() RefreshTokenUseCase {
	return &refreshTokenUsecase{}
}

func (r *refreshTokenUsecase) RefreshToken(ctx context.Context, refreshToken string, deviceID string) (string, string, error) {
	//hash resfresh token lama 
	oldHashRefreshToken := r.refreshToken.GenerateHashToken(refreshToken)
	
	//cari apakah refresh token untuk device id ini valid
	session, err := r.sessionRepo.VerifySession(ctx, oldHashRefreshToken, deviceID)
	if err != nil {
		return "", "", err
	}
	// 3. Cek device
	if session.DeviceID != deviceID {
		return "", "", ErrInvalidSession
	}

	// 4. Cek revoked
	if session.RevokedAt != nil {
		return "", "", ErrInvalidSession
	}

	// 5. Cek absolute expiration
	if time.Now().After(session.ExpiresAt) {
		return "", "", ErrRefreshTokenExpired
	}
	user, err := r.userRepo.FindByID(ctx, session.UserID)
	if err != nil {
		return "", "", err
	}
	//generate new access token
	accessToken, err := r.accToken.GenerateAccessToken(*user)
	if err != nil {
		return "", "", ErrInternal
	}
	//generate new refresh token
	newRefreshToken, err := r.refreshToken.GenerateRefreshToken()
	if err != nil {
		return "", "", ErrInternal
	}
	hashRefreshToken := r.refreshToken.GenerateHashToken(newRefreshToken)
	// update session
	err = r.sessionRepo.UpdateSession(ctx, session.ID, hashRefreshToken)
	if err != nil {
		return "", "", ErrInternal
	}

	return accessToken, newRefreshToken, nil
}
