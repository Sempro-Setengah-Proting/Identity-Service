package usecases

import (
	"context"
	"errors"
	"identityservice/internal/identity/domain"
	"identityservice/internal/identity/dto/request"
	"identityservice/internal/identity/repositories"
	"testing"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type signInUserRepository struct {
	repositories.UserRepository
	user *domain.User
	err  error
}

func (r signInUserRepository) FindEmail(context.Context, string) (*domain.User, error) {
	return r.user, r.err
}

type signInSessionRepository struct {
	repositories.SessionRepository
	lookupErr, deleteErr, createErr error
	created                         bool
}

func (r *signInSessionRepository) FindSessionByUserID(context.Context, uuid.UUID) (domain.Session, error) {
	return domain.Session{ID: uuid.New()}, r.lookupErr
}

func (r *signInSessionRepository) DeleteSession(context.Context, uuid.UUID) error {
	return r.deleteErr
}

func (r *signInSessionRepository) CreateSession(context.Context, uuid.UUID, string, string) error {
	r.created = true
	return r.createErr
}

type signInAccessToken struct{ err error }

func (g signInAccessToken) GenerateAccessToken(domain.User) (string, error) {
	return "access-token", g.err
}

type signInRefreshToken struct{ err error }

func (g signInRefreshToken) GenerateRefreshToken() (string, error) {
	return "refresh-token", g.err
}

func (signInRefreshToken) GenerateHashToken(string) string { return "hashed-refresh-token" }

func TestSignInErrorsIdentifyFailedStage(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	cause := errors.New("dependency failure")
	for _, tt := range []struct {
		name                                                            string
		userErr, lookupErr, deleteErr, accessErr, refreshErr, createErr error
		provider, hash, password                                        string
		want                                                            error
		wrapped                                                         bool
	}{
		{name: "missing account", userErr: repositories.ErrUserNotFound, want: ErrUserNotFound},
		{name: "user lookup", userErr: cause, want: ErrSignInUserLookup, wrapped: true},
		{name: "different provider", provider: "GOOGLE", want: ErrDifferentAuthProvider},
		{name: "incorrect password", password: "wrong", want: ErrInvalidCredentials},
		{name: "invalid stored hash", hash: "invalid", want: ErrSignInPasswordCheck},
		{name: "session lookup", lookupErr: cause, want: ErrSignInSessionLookup, wrapped: true},
		{name: "session deletion", deleteErr: cause, want: ErrSignInSessionDelete, wrapped: true},
		{name: "access token", accessErr: cause, want: ErrSignInAccessToken, wrapped: true},
		{name: "refresh token", refreshErr: cause, want: ErrSignInRefreshToken, wrapped: true},
		{name: "session creation", createErr: cause, want: ErrSignInSessionCreate, wrapped: true},
		{name: "success with existing session"},
		{name: "success without session", lookupErr: repositories.ErrSessionNotFound},
	} {
		t.Run(tt.name, func(t *testing.T) {
			user := &domain.User{ID: uuid.New(), Provider: "LOCAL", Password: string(hash)}
			if tt.provider != "" {
				user.Provider = tt.provider
			}
			if tt.hash != "" {
				user.Password = tt.hash
			}
			password := "password"
			if tt.password != "" {
				password = tt.password
			}
			sessions := &signInSessionRepository{lookupErr: tt.lookupErr, deleteErr: tt.deleteErr, createErr: tt.createErr}
			uc := NewLocalAuthUseCase(nil, signInUserRepository{user: user, err: tt.userErr}, signInAccessToken{tt.accessErr}, signInRefreshToken{tt.refreshErr}, sessions, nil)
			access, refresh, err := uc.SignIn(context.Background(), request.SignInRequest{Email: "user@example.com", Password: password, DeviceID: "device-001"})
			if !errors.Is(err, tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, err)
			}
			if tt.wrapped && !errors.Is(err, cause) {
				t.Fatal("original dependency error was not preserved")
			}
			if tt.want != nil {
				if access != "" || refresh != "" {
					t.Fatal("failed sign-in returned tokens")
				}
			} else if access != "access-token" || refresh != "refresh-token" || !sessions.created {
				t.Fatal("successful sign-in did not create a session and return tokens")
			}
		})
	}
}
