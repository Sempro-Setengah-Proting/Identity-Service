package user

import (
	"identityservice/internal/identity/repositories"

	"github.com/jackc/pgx/v5/pgxpool"
)

type dbUser struct {
	conn *pgxpool.Pool
}

func NewRepository(conn *pgxpool.Pool) repositories.UserRepository {
	return &dbUser{conn: conn}
}
