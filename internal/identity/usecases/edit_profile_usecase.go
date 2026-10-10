package usecases

import (
	"context"
	"identityservice/internal/identity/dto/request"
	"identityservice/internal/identity/repositories"

	"github.com/google/uuid"
)

type editProfileUseCase struct {
	repo repositories.UserRepository
}

type EditProfileUseCase interface {
	EditProfile(ctx context.Context, user request.EditProfileRequest, id uuid.UUID) error
}

func NewEditProfileUseCase(repo repositories.UserRepository) EditProfileUseCase {
	return &editProfileUseCase{
		repo: repo,
	}
}

func (e *editProfileUseCase) EditProfile(ctx context.Context, req request.EditProfileRequest, id uuid.UUID) error {

	err := e.repo.EditProfile(ctx, req, id)
	if err != nil {
		return err
	}

	return nil
}
