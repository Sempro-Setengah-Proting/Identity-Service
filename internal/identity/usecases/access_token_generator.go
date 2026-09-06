package usecases

import "identityservice/internal/identity/domain"

type TokenGenerator interface {
	GenerateAccessToken(user domain.User) (string, error)
}