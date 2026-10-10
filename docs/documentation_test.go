package docs

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
)

func TestOpenAPIPathsMatchRegisteredRoutes(t *testing.T) {
	routerSource, err := os.ReadFile("../internal/router/router.go")
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		Paths map[string]map[string]json.RawMessage `json:"paths"`
	}
	if err := json.Unmarshal(openAPISpec, &document); err != nil {
		t.Fatal(err)
	}
	groups := map[string]string{"e": ""}
	groupPattern := regexp.MustCompile(`(\w+) := (\w+)\.Group\("([^"]+)"\)`)
	for _, match := range groupPattern.FindAllStringSubmatch(string(routerSource), -1) {
		groups[match[1]] = groups[match[2]] + match[3]
	}
	routePattern := regexp.MustCompile(`(\w+)\.(GET|POST|PUT|PATCH|DELETE)\("([^"]+)"`)
	parameterPattern := regexp.MustCompile(`:(\w+)`)
	routeCount := 0
	for _, match := range routePattern.FindAllStringSubmatch(string(routerSource), -1) {
		prefix, ok := groups[match[1]]
		if !ok {
			t.Fatalf("unknown route group %q", match[1])
		}
		path := parameterPattern.ReplaceAllString(prefix+match[3], "{$1}")
		method := strings.ToLower(match[2])
		if _, ok := document.Paths[path][method]; !ok {
			t.Errorf("OpenAPI is missing registered route %s %s", match[2], path)
		}
		routeCount++
	}
	operationCount := 0
	for _, operations := range document.Paths {
		for method := range operations {
			switch method {
			case "get", "post", "put", "patch", "delete", "head", "options", "trace":
				operationCount++
			}
		}
	}
	if operationCount != routeCount {
		t.Errorf("OpenAPI has %d operations, router has %d routes", operationCount, routeCount)
	}
}

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
