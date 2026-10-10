package repositories

import (
	"context"
	"identityservice/internal/identity/repositories"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type dbTransaction struct {
	conn *pgxpool.Pool
}
type transactionKey struct{}

type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func DBFromContext(ctx context.Context, fallback *pgxpool.Pool) DBTX {
	if tx, ok := ctx.Value(transactionKey{}).(pgx.Tx); ok {
		return tx
	}

	return fallback
}

func NewTransactionRepository(conn *pgxpool.Pool) repositories.TransactionManager {
	return &dbTransaction{conn: conn}
}

func (t *dbTransaction) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := t.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	txCtx := context.WithValue(ctx, transactionKey{}, tx)
	if err = fn(txCtx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
