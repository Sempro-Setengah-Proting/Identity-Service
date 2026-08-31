package repositories

import (
	"context"
	"errors"
	"identityservice/internal/identity/domain"
	"identityservice/internal/identity/repositories"

	"gorm.io/gorm"
)

type dbUser struct {
	conn *gorm.DB
}

// FindEmail implements [repositories.UserRepository].
func (r *dbUser) FindEmail(ctx context.Context, email string) (*domain.User, error) {
	var userData domain.User
	err := r.conn.WithContext(ctx).Where("email", email).First(&userData).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repositories.ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	return &userData, nil
}

// createAccount implements [repositories.UserRepository].
func (r *dbUser) CreateAccount(ctx context.Context, newUser domain.User) error {
	err := r.conn.WithContext(ctx).Create(&newUser).Error
	if err != nil {
		return err
	}

	return nil
}

func NewUserRepository(conn *gorm.DB) repositories.UserRepository {
	return &dbUser{conn: conn}
}
