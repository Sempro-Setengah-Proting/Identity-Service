package handler

import (
	"context"
	"identityservice/internal/identity/dto/request"
	"identityservice/internal/identity/dto/response"
	appValidator "identityservice/internal/infrastructure/validator"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

type profileIdentityStub struct{ id uuid.UUID }

func (s *profileIdentityStub) FindById(_ context.Context, id uuid.UUID) (error, response.UserResponse) {
	s.id = id
	return nil, response.UserResponse{ID: id}
}

func (s *profileIdentityStub) EditProfile(_ context.Context, _ request.EditProfileRequest, id uuid.UUID) error {
	s.id = id
	return nil
}

func TestProfileHandlersReadGatewayIdentityHeader(t *testing.T) {
	id := uuid.New()
	for _, route := range []struct{ method, path, body string }{
		{http.MethodGet, "/user/me", ""},
		{http.MethodPut, "/user/edit", `{"name":"Hafid"}`},
	} {
		for _, header := range []string{"X-User-ID", "user_id"} {
			t.Run(route.path+" "+header, func(t *testing.T) {
				e := echo.New()
				e.Validator = appValidator.NewCustomValidator()
				stub := &profileIdentityStub{}
				e.GET("/user/me", NewFindProfileHandler(stub).FindById)
				e.PUT("/user/edit", NewEditProfileHandler(stub).EditProfile)
				req := httptest.NewRequest(route.method, route.path, strings.NewReader(route.body))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set(header, id.String())
				rec := httptest.NewRecorder()
				e.ServeHTTP(rec, req)
				if header == "X-User-ID" {
					if rec.Code != http.StatusOK || stub.id != id {
						t.Fatalf("gateway identity was not accepted: %d %s", rec.Code, rec.Body.String())
					}
				} else if rec.Code != http.StatusBadRequest || stub.id != uuid.Nil {
					t.Fatal("legacy client header must not supply gateway identity")
				}
			})
		}
	}
}
