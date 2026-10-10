package session

import (
	"identityservice/internal/identity/repositories"

	"github.com/jackc/pgx/v5/pgxpool"
)

type dbSession struct {
	conn *pgxpool.Pool
}

func NewRepository(conn *pgxpool.Pool) repositories.SessionRepository {
	return &dbSession{conn: conn}
}
