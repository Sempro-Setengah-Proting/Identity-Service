package user

import (
	"context"
	"errors"
	"identityservice/internal/identity/domain"
	"identityservice/internal/identity/repositories"
	postgre "identityservice/internal/infrastructure/persistence/postgre"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)
const queryFindByID = `
	SELECT id, username, email,
	       COALESCE(phone_number, ''), provider,
	       COALESCE(provider_id, ''), COALESCE(avatar_url, '')
	FROM users
	WHERE id = $1`

// FindByID implements [repositories.UserRepository].
func (r *dbUser) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var user domain.User
	err := postgre.DBFromContext(ctx, r.conn).QueryRow(ctx, queryFindByID, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PhoneNumber,
		&user.Provider,
		&user.ProviderId,
		&user.AvatarUrl,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repositories.ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	return &user, nil
}
