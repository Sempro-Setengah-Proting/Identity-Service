package token

import (
	"fmt"
	"identityservice/internal/identity/domain"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTProvider struct {
	secret        []byte
	ttl           time.Duration
	// revokedTokens map[string]int64
}

// NewJWTProvider constructs a JWTProvider with the provided secret and TTL.
func NewJWTProvider(secret string, ttl time.Duration) *JWTProvider {
	return &JWTProvider{
		secret:        []byte(strings.TrimSpace(secret)),
		ttl:           ttl,
		// revokedTokens: make(map[string]int64),
	}
}

func (j *JWTProvider) CreateToken(user domain.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(j.ttl).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(j.secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}
