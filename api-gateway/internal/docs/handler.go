package docs

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Routes() chi.Router {
	r := chi.NewRouter()

	// serves the interactive Swagger UI
	r.Get("/", serveUI)

	// serves the raw OpenAPI spec — Swagger UI fetches this
	r.Get("/swagger.yaml", serveSpec)

	return r
}

func serveUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(swaggerHTML))
}

func serveSpec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/yaml")
	_, _ = w.Write([]byte(openapiSpec))
}

// swaggerHTML is the full Swagger UI page.
// It loads Swagger UI from CDN and points it at our local spec endpoint.
const swaggerHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Cargo Platform API</title>
  <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/5.17.14/swagger-ui.min.css">
  <style>
    * { box-sizing: border-box; margin: 0; padding: 0; }
    body { background: #0f0f0f; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; }

    #header {
      background: #141414;
      border-bottom: 1px solid #222;
      padding: 16px 24px;
      display: flex;
      align-items: center;
      gap: 12px;
    }
    #header .logo {
      font-size: 18px;
      font-weight: 700;
      color: #fff;
      letter-spacing: -0.3px;
    }
    #header .version {
      font-size: 11px;
      color: #555;
      background: #1e1e1e;
      border: 1px solid #2a2a2a;
      border-radius: 4px;
      padding: 2px 8px;
    }
    #header .env {
      margin-left: auto;
      font-size: 11px;
      color: #3dd68c;
      background: #0d2818;
      border: 1px solid #1a4a2e;
      border-radius: 4px;
      padding: 2px 8px;
    }

    #swagger-ui { max-width: 1200px; margin: 0 auto; padding: 24px; }

    /* dark theme overrides */
    .swagger-ui { background: transparent !important; }
    .swagger-ui .scheme-container,
    .swagger-ui .wrapper,
    .swagger-ui section.models { background: transparent !important; box-shadow: none !important; }
    .swagger-ui .info { margin: 0 0 24px; }
    .swagger-ui .info .title { color: #fff !important; font-size: 24px !important; }
    .swagger-ui .info p,
    .swagger-ui .info li,
    .swagger-ui .info table { color: #aaa !important; }
    .swagger-ui .info a { color: #60a5fa !important; }
    .swagger-ui .opblock-tag { color: #fff !important; border-bottom: 1px solid #222 !important; }
    .swagger-ui .opblock-tag:hover { background: #1a1a1a !important; }
    .swagger-ui .opblock { border-radius: 8px !important; margin-bottom: 8px !important; border: 1px solid #222 !important; }
    .swagger-ui .opblock .opblock-summary { border-radius: 7px !important; }
    .swagger-ui .opblock.opblock-post { background: #0d1f2d !important; border-color: #1a4a6b !important; }
    .swagger-ui .opblock.opblock-get  { background: #0d1f14 !important; border-color: #1a4a2e !important; }
    .swagger-ui .opblock.opblock-put  { background: #1f1a0d !important; border-color: #4a3a1a !important; }
    .swagger-ui .opblock.opblock-delete { background: #1f0d0d !important; border-color: #4a1a1a !important; }
    .swagger-ui .opblock-summary-method { border-radius: 4px !important; font-size: 11px !important; min-width: 60px !important; }
    .swagger-ui .opblock-summary-path,
    .swagger-ui .opblock-summary-description { color: #ccc !important; }
    .swagger-ui .opblock-body pre,
    .swagger-ui .highlight-code { background: #0a0a0a !important; border-radius: 6px !important; }
    .swagger-ui textarea,
    .swagger-ui input[type=text],
    .swagger-ui input[type=password],
    .swagger-ui input[type=email] {
      background: #1a1a1a !important;
      color: #fff !important;
      border: 1px solid #333 !important;
      border-radius: 6px !important;
    }
    .swagger-ui select { background: #1a1a1a !important; color: #fff !important; border: 1px solid #333 !important; }
    .swagger-ui .btn { border-radius: 6px !important; }
    .swagger-ui .btn.execute { background: #2563eb !important; border-color: #2563eb !important; }
    .swagger-ui .btn.authorize { background: #059669 !important; border-color: #059669 !important; color: #fff !important; }
    .swagger-ui .model-box,
    .swagger-ui section.models .model-container { background: #141414 !important; border: 1px solid #222 !important; border-radius: 8px !important; }
    .swagger-ui .model { color: #ccc !important; }
    .swagger-ui table thead tr th,
    .swagger-ui table thead tr td { color: #888 !important; border-bottom: 1px solid #222 !important; }
    .swagger-ui .parameter__name,
    .swagger-ui .parameter__type { color: #ccc !important; }
    .swagger-ui .response-col_status { color: #3dd68c !important; }
    .swagger-ui .responses-inner h4,
    .swagger-ui .responses-inner h5 { color: #aaa !important; }
    .swagger-ui .topbar { display: none !important; }
    .swagger-ui .info .base-url { color: #60a5fa !important; }
    /* hide the default title area — we have our own header */
    .swagger-ui .info hgroup.main { display: none !important; }
  </style>
</head>
<body>

<div id="header">
  <span class="logo">📦 Cargo Platform</span>
  <span class="version">v0.1.0</span>
  <span class="env">● local dev</span>
</div>

<div id="swagger-ui"></div>

<script src="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/5.17.14/swagger-ui-bundle.min.js"></script>
<script>
  SwaggerUIBundle({
    url: "/docs/swagger.yaml",
    dom_id: "#swagger-ui",
    presets: [SwaggerUIBundle.presets.apis, SwaggerUIBundle.SwaggerUIStandalonePreset],
    layout: "BaseLayout",
    deepLinking: true,
    displayRequestDuration: true,
    defaultModelsExpandDepth: 1,
    defaultModelExpandDepth: 1,
    docExpansion: "list",
    filter: true,
    persistAuthorization: true,
  });
</script>
</body>
</html>`