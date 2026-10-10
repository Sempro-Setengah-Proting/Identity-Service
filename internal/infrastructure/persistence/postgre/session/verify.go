package session

import (
	"context"
	"identityservice/internal/identity/domain"
	postgre "identityservice/internal/infrastructure/persistence/postgre"
)

// VerifySession implements [repositories.SessionRepository].
func (d *dbSession) VerifySession(ctx context.Context, refreshToken string, deviceID string) (domain.Session, error) {
	query := `SELECT ` + selectSessionColumns + `
		FROM sessions
		WHERE refresh_token = $1 AND device_id = $2
		LIMIT 1`

	return scanSession(postgre.DBFromContext(ctx, d.conn).QueryRow(ctx, query, refreshToken, deviceID))
}
