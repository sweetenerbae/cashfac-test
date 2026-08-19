BACKEND_DIR := ./backend
FRONTEND_DIR := ./frontend

.DEFAULT_GOAL := help

.PHONY: help setup run dev build test check fmt clean swagger caddy-run caddy-validate frontend-install frontend-dev frontend-build docker-check docker-up docker-down docker-build docker-logs docker-status

help:
	@echo Cash Factories News
	@echo   make setup             Create .env from .env.example
	@echo   make docker-up         Build and start the complete application
	@echo   make docker-logs       Follow application logs
	@echo   make docker-status     Show container status
	@echo   make docker-down       Stop containers
	@echo   make run               Run backend locally
	@echo   make frontend-install  Install frontend dependencies
	@echo   make frontend-dev      Run frontend locally
	@echo   make check             Run backend tests and build frontend

ifeq ($(OS),Windows_NT)
setup:
	@powershell.exe -NoProfile -NonInteractive -Command "if (Test-Path '.env') { Write-Output '.env already exists; keeping it unchanged' } else { Copy-Item '.env.example' '.env'; Write-Output 'Created .env. Add GUARDIAN_API_KEY and ZAI_API_KEY before starting the app.' }"
else
setup:
	@if [ -f .env ]; then \
		echo '.env already exists; keeping it unchanged'; \
	else \
		cp .env.example .env; \
		echo 'Created .env. Add GUARDIAN_API_KEY and ZAI_API_KEY before starting the app.'; \
	fi
endif

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

ifeq ($(OS),Windows_NT)
docker-check:
	@powershell.exe -NoProfile -NonInteractive -Command "docker info 2>&1 | Out-Null; if ($$LASTEXITCODE -ne 0) { Write-Error 'Docker is not running. Start Docker Desktop and try again.'; exit 1 }"
else
docker-check:
	@docker info >/dev/null 2>&1 || (echo 'Docker is not running. Start Docker Desktop and try again.'; exit 1)
endif

docker-build: docker-check
	@docker compose build

docker-up: docker-check
	@docker compose up --build -d
	@echo Application started:
	@echo   Website: http://localhost:8080
	@echo   Swagger: http://localhost:8080/docs
	@echo   Health:  http://localhost:8080/health
	@echo Logs: make docker-logs

docker-down:
	@docker compose down

docker-logs:
	@docker compose logs -f

docker-status:
	@docker compose ps
