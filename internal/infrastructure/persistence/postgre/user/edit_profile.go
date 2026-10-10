package user

import (
	"context"
	"errors"
	"fmt"
	"identityservice/internal/identity/dto/request"
	"identityservice/internal/identity/repositories"
	postgre "identityservice/internal/infrastructure/persistence/postgre"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var query = `
		UPDATE users
		SET
	`

// EditProfile implements [repositories.UserRepository].
func (r *dbUser) EditProfile(ctx context.Context, req request.EditProfileRequest, id uuid.UUID) error {
	//builder dynamic query
	//value
	args := []any{}

	separator := ""

	if req.PhoneNumber != nil {
		args = append(args, *req.PhoneNumber)

		query += fmt.Sprintf(
			"%s phone_number = $%d",
			separator,
			len(args),
		)

		separator = ","
	}

	if req.Name != nil {
		args = append(args, *req.Name)

		query += fmt.Sprintf(
			"%s username = $%d",
			separator,
			len(args),
		)

		separator = ","
	}

	if len(args) == 0 {
		return nil
	}

	args = append(args, id)

	query += fmt.Sprintf(`
		WHERE id = $%d
		RETURNING updated_at
	`, len(args))

	err := postgre.DBFromContext(ctx, r.conn).QueryRow(
		ctx,
		query,
		args,
		id,
	).Scan()

	if errors.Is(err, pgx.ErrNoRows) {
		return repositories.ErrUserNotFound
	}
	if err != nil {
		return err
	}
	return nil
}
