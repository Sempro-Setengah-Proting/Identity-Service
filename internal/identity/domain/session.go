package domain

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID           uuid.UUID `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey"`
	UserID       uuid.UUID `gorm:"column:user_id;type:uuid;not null"`
	DeviceID     string    `gorm:"column:device_id;type:varchar(255);not null"`
	RefreshToken string    `gorm:"column:refresh_token;type:varchar(255);not null"`
	CreatedAt    time.Time `gorm:"column:created_at;type:timestamp;not null"`
	UpdatedAt    time.Time `gorm:"column:updated_at;type:timestamp;not null"`
	RevokedAt    *time.Time `gorm:"column:revoked_at;type:timestamp;default:null"`
	ExpiresAt    time.Time `gorm:"column:expires_at;type:timestamp;not null"`
}
