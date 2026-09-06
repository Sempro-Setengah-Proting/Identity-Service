package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

type RefreshToken struct {}

func NewRefreshToken() *RefreshToken {
	return &RefreshToken{}
}

func (j *RefreshToken) GenerateRefreshToken() (string, error) {

	bytes := make([]byte, 32)

	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func (j *RefreshToken) GenerateHashToken(token string) string {
	hasher := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hasher[:])
}