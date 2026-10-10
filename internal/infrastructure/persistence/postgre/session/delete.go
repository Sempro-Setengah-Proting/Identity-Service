package session

import (
	"context"
	postgre "identityservice/internal/infrastructure/persistence/postgre"

	"github.com/google/uuid"
)

// DeleteSession implements [repositories.SessionRepository].
func (d *dbSession) DeleteSession(ctx context.Context, id uuid.UUID) error {
	_, err := postgre.DBFromContext(ctx, d.conn).Exec(
		ctx,
		"DELETE FROM sessions WHERE id = $1",
		id,
	)
	if err != nil {
		return err
	}
	return nil
}
