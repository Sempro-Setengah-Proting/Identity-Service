package usecases

import (
	"context"

	"identityservice/internal/identity/dto/response"
	"identityservice/internal/identity/repositories"

	"github.com/google/uuid"
)

type findByIdUsecase struct {
	repo repositories.UserRepository
}

type FindByIdUsecase interface {
	FindById(ctx context.Context, id uuid.UUID) (error, response.UserResponse)
}

func NewFindByIdUsecase(repo repositories.UserRepository) FindByIdUsecase {
	return &findByIdUsecase{
		repo: repo,
	}
}

func (e *findByIdUsecase) FindById(ctx context.Context, id uuid.UUID) (error, response.UserResponse) {
	user, err := e.repo.FindByID(ctx, id)
	if err != nil {
		return err, response.UserResponse{}
	}
	res := response.UserResponse{
		ID: user.ID,
		Username: user.Username,
		Email: user.Email,
		PhoneNumber: user.PhoneNumber,
		Provider: user.Provider,
		ProviderID: user.ProviderId,
		AvatarURL: user.AvatarUrl,
	}

	return nil, res
}
