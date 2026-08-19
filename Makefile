BACKEND_DIR := ./backend
FRONTEND_DIR := ./frontend

.DEFAULT_GOAL := help

.PHONY: help setup run dev build test check fmt clean swagger caddy-run caddy-validate frontend-install frontend-dev frontend-build docker-check docker-up docker-down docker-build docker-logs docker-status

help:
	@printf '%s\n' \
		'Cash Factories News' \
		'' \
		'  make setup             Create .env from .env.example' \
		'  make docker-up         Build and start the complete application' \
		'  make docker-logs       Follow application logs' \
		'  make docker-status     Show container status' \
		'  make docker-down       Stop containers' \
		'' \
		'  make run               Run backend locally' \
		'  make frontend-install  Install frontend dependencies' \
		'  make frontend-dev      Run frontend locally' \
		'  make check             Run backend tests and build frontend'

setup:
	@if [ -f .env ]; then \
		echo '.env already exists; keeping it unchanged'; \
	else \
		cp .env.example .env; \
		echo 'Created .env. Add GUARDIAN_API_KEY and ZAI_API_KEY before starting the app.'; \
	fi

run:
	@$(MAKE) -C $(BACKEND_DIR) run

dev:
	@$(MAKE) -C $(BACKEND_DIR) dev

build:
	@$(MAKE) -C $(BACKEND_DIR) build

test:
	@$(MAKE) -C $(BACKEND_DIR) test

check: test frontend-build

fmt:
	@$(MAKE) -C $(BACKEND_DIR) fmt

clean:
	@$(MAKE) -C $(BACKEND_DIR) clean

swagger:
	@$(MAKE) -C $(BACKEND_DIR) swagger

caddy-run:
	@$(MAKE) -C $(BACKEND_DIR) caddy-run

caddy-validate:
	@$(MAKE) -C $(BACKEND_DIR) caddy-validate

frontend-install:
	@cd $(FRONTEND_DIR) && npm install

frontend-dev:
	@cd $(FRONTEND_DIR) && npm run dev

frontend-build:
	@cd $(FRONTEND_DIR) && npm run build

docker-check:
	@docker info >/dev/null 2>&1 || (echo 'Docker is not running. Start Docker Desktop and try again.'; exit 1)

docker-build: docker-check
	@docker compose build

docker-up: docker-check
	@docker compose up --build -d
	@printf '%s\n' \
		'' \
		'Application started:' \
		'  Website: http://localhost:8080' \
		'  Swagger: http://localhost:8080/docs' \
		'  Health:  http://localhost:8080/health' \
		'' \
		'Logs: make docker-logs'

docker-down:
	@docker compose down

docker-logs:
	@docker compose logs -f

docker-status:
	@docker compose ps
