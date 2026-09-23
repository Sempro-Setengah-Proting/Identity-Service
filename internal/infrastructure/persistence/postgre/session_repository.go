package repositories

import (
	"context"
	"errors"
	"identityservice/internal/identity/domain"
	"identityservice/internal/identity/repositories"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type dbSession struct {
	conn *gorm.DB
}

// DeleteSessionByUserID implements [repositories.SessionRepository].
func (d *dbSession) DeleteSession(ctx context.Context, id uuid.UUID) error {
	err := dbFromContext(ctx, d.conn).Where("id = ?", id).Delete(&domain.Session{}).Error
	if err != nil {
		return err
	}
	return nil
}

// FindSessionByUserID implements [repositories.SessionRepository].
func (d *dbSession) FindSessionByUserID(ctx context.Context, userId uuid.UUID) (domain.Session, error) {
	var session domain.Session
	err := dbFromContext(ctx, d.conn).Where("user_id = ?", userId).First(&session).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.Session{}, repositories.ErrSessionNotFound
	}

	if err != nil {
		return domain.Session{}, err
	}
	return session, nil
}

// VerifySession implements [repositories.SessionRepository].
func (d *dbSession) VerifySession(ctx context.Context, refreshToken string, deviceID string) (domain.Session, error) {
	var session domain.Session
	err := dbFromContext(ctx, d.conn).Where("refresh_token = ? AND device_id = ?", refreshToken, deviceID).First(&session).Error
	if err != nil {
		return domain.Session{}, err
	}
	return session, nil
}

// CreateSession implements [repositories.SessionRepository].
func (d *dbSession) CreateSession(ctx context.Context, userID uuid.UUID, deviceID string, refreshToken string) error {

	err := dbFromContext(ctx, d.conn).Create(&domain.Session{
		UserID:       userID,
		DeviceID:     deviceID,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().AddDate(0, 0, 30),
	}).Error
	if err != nil {
		return err
	}
	return nil
}

func (d *dbSession) UpdateSession(ctx context.Context, sessionID uuid.UUID, refreshToken string) error {
	err := dbFromContext(ctx, d.conn).Where("id = ?", sessionID).Update("refresh_token", refreshToken).Error
	if err != nil {
		return err
	}
	return nil
}

func NewSessionRepository(conn *gorm.DB) repositories.SessionRepository {
	return &dbSession{conn: conn}
}
