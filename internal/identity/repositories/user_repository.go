package repositories

import (
	"context"
	"errors"
	"identityservice/internal/identity/domain"
	"identityservice/internal/identity/dto/request"

	"github.com/google/uuid"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type UserRepository interface {
	CreateAccount(ctx context.Context, newUser *domain.User) error
	FindEmail(ctx context.Context, email string) (*domain.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	ForgetPassword(ctx context.Context, email string, password string) error
	EditProfile(ctx context.Context, req request.EditProfileRequest, id uuid.UUID) error
}
