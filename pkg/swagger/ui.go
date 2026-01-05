package swagger

import (
	"net/http"
	"text/template"
)

// SwaggerUIHTML is a lightweight HTML template that loads Swagger UI from a CDN
const SwaggerUIHTML = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8">
    <title>gRPC API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui.css" />
    <link rel="icon" type="image/png" href="https://unpkg.com/swagger-ui-dist@5.11.0/favicon-32x32.png" />
    <style>
      html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
      *, *:before, *:after { box-sizing: inherit; }
      body { margin:0; background: #fafafa; }
    </style>
  </head>
  <body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-bundle.js" charset="UTF-8"> </script>
    <script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-standalone-preset.js" charset="UTF-8"> </script>
    <script>
    window.onload = function() {
      // Begin Swagger UI call region
      const ui = SwaggerUIBundle({
        url: "/swagger.json", // Point to our generated JSON
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        plugins: [
          SwaggerUIBundle.plugins.DownloadUrl
        ],
        layout: "StandaloneLayout"
      });
      // End Swagger UI call region
      window.ui = ui;
    };
  </script>
  </body>
</html>
`

// ServeUI renders the Swagger UI page
func ServeUI(w http.ResponseWriter, r *http.Request) {
	t, err := template.New("swagger").Parse(SwaggerUIHTML)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.Execute(w, nil)
}

// ServeJSON serves the generated OpenAPI JSON file
func ServeJSON(w http.ResponseWriter, r *http.Request) {
	// Adjust this path if your buf.gen.yaml output path changes
	http.ServeFile(w, r, "api/openapi/apidocs.swagger.json")
}