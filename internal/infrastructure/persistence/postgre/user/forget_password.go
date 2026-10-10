package user

import (
	"context"
	postgre "identityservice/internal/infrastructure/persistence/postgre"
)

// ForgetPassword implements [repositories.UserRepository].
func (r *dbUser) ForgetPassword(ctx context.Context, email string, password string) error {
	_, err := postgre.DBFromContext(ctx, r.conn).Exec(
		ctx,
		"UPDATE users SET password = $1, updated_at = CURRENT_TIMESTAMP WHERE email = $2",
		password,
		email,
	)
	if err != nil {
		return err
	}
	return nil
}
