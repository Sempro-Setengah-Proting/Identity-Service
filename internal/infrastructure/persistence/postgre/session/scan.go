package session

import (
	"identityservice/internal/identity/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const selectSessionColumns = `
	id,
	user_id,
	device_id,
	refresh_token,
	created_at,
	updated_at,
	revoked_at,
	expires_at`

func scanSession(row pgx.Row) (domain.Session, error) {
	var session domain.Session
	var revokedAt pgtype.Timestamp
	err := row.Scan(
		&session.ID,
		&session.UserID,
		&session.DeviceID,
		&session.RefreshToken,
		&session.CreatedAt,
		&session.UpdatedAt,
		&revokedAt,
		&session.ExpiresAt,
	)
	if err != nil {
		return domain.Session{}, err
	}
	if revokedAt.Valid {
		session.RevokedAt = &revokedAt.Time
	}

	return session, nil
}
