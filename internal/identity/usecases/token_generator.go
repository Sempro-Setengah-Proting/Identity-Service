package usecases

import "identityservice/internal/identity/domain"

type TokenGenerator interface {
	CreateToken(user domain.User) (string, error)
}