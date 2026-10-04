package docs

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
)

func TestOpenAPISpecIsValidJSON(t *testing.T) {
	var document map[string]any
	if err := json.Unmarshal(openAPISpec, &document); err != nil {
		t.Fatalf("openapi.json is not valid JSON: %v", err)
	}
	if document["openapi"] != "3.1.0" {
		t.Fatalf("unexpected OpenAPI version: %v", document["openapi"])
	}
}

func TestDocumentationRoutes(t *testing.T) {
	e := echo.New()
	Register(e)

	tests := []struct {
		path        string
		contentType string
		body        string
	}{
		{path: "/openapi.json", contentType: "application/json", body: `"openapi": "3.1.0"`},
		{path: "/documentation", contentType: "text/html", body: "SwaggerUIBundle"},
		{path: "/documentation/", contentType: "text/html", body: "SwaggerUIBundle"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, tt.path, nil)
			e.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d", recorder.Code)
			}
			if !strings.Contains(recorder.Header().Get("Content-Type"), tt.contentType) {
				t.Fatalf("unexpected Content-Type: %s", recorder.Header().Get("Content-Type"))
			}
			if !strings.Contains(recorder.Body.String(), tt.body) {
				t.Fatalf("response body does not contain %q", tt.body)
			}
		})
	}
}
