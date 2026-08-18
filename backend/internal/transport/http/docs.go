package http

import _ "embed"

var (
	//go:embed static/docs/index.html
	swaggerHTML []byte

	//go:embed static/openapi.yaml
	openAPISpec []byte

	//go:embed static/swagger-ui/swagger-ui.css
	swaggerCSS []byte

	//go:embed static/swagger-ui/swagger-ui-bundle.js
	swaggerJS []byte
)
