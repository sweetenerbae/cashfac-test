APP_NAME := cashfac-news-api
MAIN_PKG := ./cmd/api
BUILD_DIR := ./bin
GO_FILES := $(shell find . -name '*.go' -type f)
ENV_FILE := ../.env

.PHONY: fmt build run dev clean test swagger caddy-run caddy-validate

fmt:
	gofmt -w $(GO_FILES)

build:
	go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PKG)

run:
	@set -a; \
	if [ -f $(ENV_FILE) ]; then . $(ENV_FILE); fi; \
	set +a; \
	go run $(MAIN_PKG)

dev:
	@set -a; \
	if [ -f $(ENV_FILE) ]; then . $(ENV_FILE); fi; \
	set +a; \
	HTTP_PORT=8080 go run $(MAIN_PKG)

test:
	go test ./...

clean:
	rm -rf $(BUILD_DIR)

swagger:
	@echo "Swagger UI: http://localhost:8080/docs"
	@echo "OpenAPI spec: http://localhost:8080/openapi.yaml"

caddy-validate:
	caddy validate --config Caddyfile

caddy-run:
	caddy run --config Caddyfile
