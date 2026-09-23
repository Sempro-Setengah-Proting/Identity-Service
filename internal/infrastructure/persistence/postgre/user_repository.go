package repositories

import (
	"context"
	"errors"
	"identityservice/internal/identity/domain"
	"identityservice/internal/identity/repositories"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type dbUser struct {
	conn *gorm.DB
}

// ForgePassword implements [repositories.UserRepository].
func (r *dbUser) ForgetPassword(ctx context.Context, email string, password string) error {
	err := dbFromContext(ctx, r.conn).Model(&domain.User{}).Where("email", email).Update("password", password).Error
	if err != nil {
		return err
	}
	return nil
}

// FindByID implements [repositories.UserRepository].
func (r *dbUser) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var userData domain.User
	err := dbFromContext(ctx, r.conn).Where("id", id).First(&userData).Error
	if err != nil {
		return nil, err
	}
	return &userData, nil
}

// FindEmail implements [repositories.UserRepository].
func (r *dbUser) FindEmail(ctx context.Context, email string) (*domain.User, error) {
	var userData domain.User
	err := dbFromContext(ctx, r.conn).Where("email", email).First(&userData).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repositories.ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	return &userData, nil
}

// createAccount implements [repositories.UserRepository].
func (r *dbUser) CreateAccount(ctx context.Context, newUser *domain.User) error {
	err := dbFromContext(ctx, r.conn).Create(newUser).Error
	if err != nil {
		return err
	}

	return nil
}

func NewUserRepository(conn *gorm.DB) repositories.UserRepository {
	return &dbUser{conn: conn}
}
