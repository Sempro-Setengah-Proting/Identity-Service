package user

import (
	"context"
	"identityservice/internal/identity/domain"
	postgre "identityservice/internal/infrastructure/persistence/postgre"
)

// CreateAccount implements [repositories.UserRepository].
func (r *dbUser) CreateAccount(ctx context.Context, newUser *domain.User) error {
	const query = `
		INSERT INTO users (
			username,
			email,
			password,
			phone_number,
			provider,
			provider_id,
			avatar_url
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`

	err := postgre.DBFromContext(ctx, r.conn).QueryRow(
		ctx,
		query,
		newUser.Username,
		newUser.Email,
		newUser.Password,
		newUser.PhoneNumber,
		newUser.Provider,
		newUser.ProviderId,
		newUser.AvatarUrl,
	).Scan(&newUser.ID, &newUser.CreatedAt, &newUser.UpdatedAt)
	if err != nil {
		return err
	}

	return nil
}
