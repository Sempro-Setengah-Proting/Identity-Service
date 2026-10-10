package repositories

import (
	"context"
	"errors"
	"identityservice/internal/identity/domain"

	"github.com/google/uuid"
)

var (
	ErrSessionNotFound = errors.New("session not found")
)

type SessionRepository interface {
	CreateSession(ctx context.Context, userID uuid.UUID, deviceID string, refreshToken string) error
	VerifySession(ctx context.Context, refreshToken string, deviceID string) (domain.Session, error)
	UpdateSession(ctx context.Context, sessionID uuid.UUID, refreshToken string) error
	FindSessionByUserID(ctx context.Context, userId uuid.UUID) (domain.Session, error)
	DeleteSession(ctx context.Context, id uuid.UUID) error
}
