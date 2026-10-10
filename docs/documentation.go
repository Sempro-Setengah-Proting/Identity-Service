package docs

import (
	_ "embed"
	"net/http"

	"github.com/labstack/echo/v5"
)

//go:embed openapi.json
var openAPISpec []byte

// The Swagger UI assets are pinned to a specific major-compatible release.
const swaggerUI = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>TrapVal Identity API</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.33.1/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.33.1/swagger-ui-bundle.js"></script>
  <script>
    window.onload = function () {
      SwaggerUIBundle({
        url: window.location.pathname.replace(/\/documentation\/?$/, "/openapi.json"),
        dom_id: "#swagger-ui",
        deepLinking: true,
        presets: [SwaggerUIBundle.presets.apis],
        layout: "BaseLayout"
      });
    };
  </script>
</body>
</html>`

// Register exposes the raw OpenAPI document and its Swagger UI.
func Register(e *echo.Echo) {
	e.GET("/openapi.json", openAPI)
	e.GET("/documentation", swagger)
	e.GET("/documentation/", swagger)
}

func openAPI(c *echo.Context) error {
	return c.Blob(http.StatusOK, "application/json; charset=utf-8", openAPISpec)
}

func swagger(c *echo.Context) error {
	return c.HTML(http.StatusOK, swaggerUI)
}
