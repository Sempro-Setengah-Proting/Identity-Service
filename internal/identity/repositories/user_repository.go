package repositories

import (
	"context"
	"errors"
	"identityservice/internal/identity/domain"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type UserRepository interface {
	CreateAccount(ctx context.Context, newUser domain.User) error
	FindEmail(ctx context.Context, email string) (*domain.User, error)
}

