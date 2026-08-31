package oauth

import (
	"context"
	"errors"
	"identityservice/internal/identity/domain"

	"cloud.google.com/go/auth/credentials/idtoken"
)

//decrypt si id token -> tujuan utama
type GoogleVerifier struct{
	ClientId string
}

func NewGoogleVerifier(clientId string) *GoogleVerifier {
	return &GoogleVerifier{
		ClientId: clientId,
	}
}

func (g *GoogleVerifier) VerifyGoogleToken(ctx context.Context, token string) (*domain.GoogleIdentity, error) {
	payload, err := idtoken.Validate(ctx,token, g.ClientId)

	if err != nil {
		return nil, errors.New("invalid goggle id token")
	}

	return &domain.GoogleIdentity{
		ProvideId: payload.Subject,
		Email: payload.Claims["email"].(string),
		EmailVerified: payload.Claims["email_verified"].(bool),
		DisplayName: payload.Claims["name"].(string),
		AvatarUrl: payload.Claims["picture"].(string),
	}, nil
}
