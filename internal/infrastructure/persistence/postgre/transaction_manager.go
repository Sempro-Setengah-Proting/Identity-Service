package repositories

import (
	"context"
	"identityservice/internal/identity/repositories"

	"gorm.io/gorm"
)

type dbTransaction struct {
	conn *gorm.DB
}
type transactionKey struct{}

func NewTransactionRepository(conn *gorm.DB) repositories.TransactionManager {
	return &dbTransaction{conn: conn}
}

func (t *dbTransaction) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return t.conn.WithContext(ctx).Transaction(
		func(tx *gorm.DB) error {
			txCtx := context.WithValue(
				ctx,
				transactionKey{},
				tx,
			)

			return fn(txCtx)
		},
	)
}