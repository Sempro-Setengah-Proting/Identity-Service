package providers

import (
	"context"
	"identityservice/internal/identity/domain"
)

type GoogleVerifier interface {
	VerifyGoogleToken(ctx context.Context, idToken string) (domain.GoogleIdentity, error)
}