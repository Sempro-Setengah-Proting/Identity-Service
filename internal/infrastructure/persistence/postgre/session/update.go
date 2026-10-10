package session

import (
	"context"
	postgre "identityservice/internal/infrastructure/persistence/postgre"

	"github.com/google/uuid"
)

// UpdateSession implements [repositories.SessionRepository].
func (d *dbSession) UpdateSession(ctx context.Context, sessionID uuid.UUID, refreshToken string) error {
	_, err := postgre.DBFromContext(ctx, d.conn).Exec(
		ctx,
		"UPDATE sessions SET refresh_token = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2",
		refreshToken,
		sessionID,
	)
	if err != nil {
		return err
	}
	return nil
}
