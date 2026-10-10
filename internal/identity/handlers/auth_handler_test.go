package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"identityservice/internal/httpresponse"
	"identityservice/internal/identity/dto/request"
	"identityservice/internal/identity/usecases"
	appValidator "identityservice/internal/infrastructure/validator"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
)

type signInErrorStub struct {
	usecases.LocalAuthUseCase
	err    error
	called bool
}

func (s *signInErrorStub) SignIn(context.Context, request.SignInRequest) (string, string, error) {
	s.called = true
	return "", "", s.err
}

func TestSignInErrorResponses(t *testing.T) {
	validBody := `{"email":"user@example.com","password":"secret","device_id":"device-001"}`
	for _, tt := range []struct {
		name, body string
		err        error
		status     int
		message    string
	}{
		{"invalid JSON", `{`, nil, 400, "invalid sign-in payload: expected email, password and device_id as strings"},
		{"missing fields", `{}`, nil, 400, "email is required; password is required; device_id is required"},
		{"invalid email", `{"email":"invalid","password":"secret","device_id":"device-001"}`, nil, 400, "email must be a valid email address"},
		{"account not found", validBody, usecases.ErrUserNotFound, 400, "no account found for this email"},
		{"provider mismatch", validBody, usecases.ErrDifferentAuthProvider, 400, "email registered with another provider; use the matching sign-in method"},
		{"wrong password", validBody, usecases.ErrInvalidCredentials, 400, "incorrect password"},
		{"user lookup", validBody, usecases.ErrSignInUserLookup, 500, usecases.ErrSignInUserLookup.Error()},
		{"password check", validBody, usecases.ErrSignInPasswordCheck, 500, usecases.ErrSignInPasswordCheck.Error()},
		{"session lookup", validBody, usecases.ErrSignInSessionLookup, 500, usecases.ErrSignInSessionLookup.Error()},
		{"session deletion", validBody, usecases.ErrSignInSessionDelete, 500, usecases.ErrSignInSessionDelete.Error()},
		{"access token", validBody, usecases.ErrSignInAccessToken, 500, usecases.ErrSignInAccessToken.Error()},
		{"refresh token", validBody, usecases.ErrSignInRefreshToken, 500, usecases.ErrSignInRefreshToken.Error()},
		{"session creation", validBody, usecases.ErrSignInSessionCreate, 500, usecases.ErrSignInSessionCreate.Error()},
		{"unexpected", validBody, errors.New("unknown failure"), 500, "unexpected error during sign in"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			e.Validator = appValidator.NewCustomValidator()
			stub := &signInErrorStub{}
			if tt.err != nil {
				stub.err = fmt.Errorf("%w: private dependency details", tt.err)
			}
			e.POST("/auth/sign-in", NewAuthHandler(stub).SignIn)
			req := httptest.NewRequest(http.MethodPost, "/auth/sign-in", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			if rec.Code != tt.status {
				t.Fatalf("expected %d, got %d", tt.status, rec.Code)
			}
			var response httpresponse.Response
			if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
				t.Fatal(err)
			}
			if response.Status != "error" || response.Message != tt.message {
				t.Fatalf("unexpected error response: %s", rec.Body.String())
			}
			if tt.err == nil && stub.called {
				t.Fatal("invalid payload reached the usecase")
			}
		})
	}
}
