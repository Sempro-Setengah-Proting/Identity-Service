package domain

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	DeviceID     string
	RefreshToken string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	RevokedAt    *time.Time
	ExpiresAt    time.Time
}
