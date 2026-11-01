// Package docs serves the OpenAPI specification and a Swagger UI page.
// The spec is embedded in the binary, so docs always ship with the code.
package docs

import (
	_ "embed"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed openapi.yaml
var spec []byte

// Swagger UI is loaded from a CDN (pinned version) to keep the API binary small.
// "openapi.yaml" is relative, so the page works behind any proxy prefix.
const swaggerUI = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Calculator API · Swagger UI</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui-bundle.js" crossorigin></script>
  <script>
    window.ui = SwaggerUIBundle({ url: "openapi.yaml", dom_id: "#swagger-ui", tryItOutEnabled: true });
  </script>
</body>
</html>`

// Spec handles GET /api/openapi.yaml.
func Spec(c *gin.Context) {
	c.Data(http.StatusOK, "application/yaml; charset=utf-8", spec)
}

// UI handles GET /api/docs.
func UI(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(swaggerUI))
}
