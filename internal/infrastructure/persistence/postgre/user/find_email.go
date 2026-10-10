package user

import (
	"context"
	"errors"
	"identityservice/internal/identity/domain"
	"identityservice/internal/identity/repositories"
	postgre "identityservice/internal/infrastructure/persistence/postgre"

	"github.com/jackc/pgx/v5"
)

const queryFindByEmail = `
	SELECT id, username, email, COALESCE(password, ''),
	       COALESCE(phone_number, ''), provider,
	       COALESCE(provider_id, ''), COALESCE(avatar_url, '')
	FROM users
	WHERE email = $1`

// FindEmail implements [repositories.UserRepository].
func (r *dbUser) FindEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := postgre.DBFromContext(ctx, r.conn).QueryRow(ctx, queryFindByEmail, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
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
