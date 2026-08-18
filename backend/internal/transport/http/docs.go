package http

import _ "embed"

var (
	//go:embed static/docs/index.html
	swaggerHTML []byte

	//go:embed static/openapi.yaml
	openAPISpec []byte
)
