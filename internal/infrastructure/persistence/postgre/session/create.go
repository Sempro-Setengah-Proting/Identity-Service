package session

import (
	"context"
	postgre "identityservice/internal/infrastructure/persistence/postgre"
	"time"

	"github.com/google/uuid"
)

// CreateSession implements [repositories.SessionRepository].
func (d *dbSession) CreateSession(ctx context.Context, userID uuid.UUID, deviceID string, refreshToken string) error {
	const query = `
		INSERT INTO sessions (user_id, device_id, refresh_token, expires_at)
		VALUES ($1, $2, $3, $4)`

	_, err := postgre.DBFromContext(ctx, d.conn).Exec(
		ctx,
		query,
		userID,
		deviceID,
		refreshToken,
		time.Now().AddDate(0, 0, 30),
	)
	if err != nil {
		return err
	}
	return nil
}
