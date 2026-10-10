package session

import (
	"context"
	"errors"
	"identityservice/internal/identity/domain"
	"identityservice/internal/identity/repositories"
	postgre "identityservice/internal/infrastructure/persistence/postgre"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// FindSessionByUserID implements [repositories.SessionRepository].
func (d *dbSession) FindSessionByUserID(ctx context.Context, userID uuid.UUID) (domain.Session, error) {
	query := `SELECT ` + selectSessionColumns + ` FROM sessions WHERE user_id = $1 LIMIT 1`
	session, err := scanSession(postgre.DBFromContext(ctx, d.conn).QueryRow(ctx, query, userID))

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Session{}, repositories.ErrSessionNotFound
	}

	if err != nil {
		return domain.Session{}, err
	}
	return session, nil
}
