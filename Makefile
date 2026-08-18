BACKEND_DIR := ./backend
FRONTEND_DIR := ./frontend

.PHONY: run dev build test fmt clean swagger caddy-run caddy-validate frontend-install frontend-dev frontend-build docker-up docker-down docker-build docker-logs

run:
	@$(MAKE) -C $(BACKEND_DIR) run

dev:
	@$(MAKE) -C $(BACKEND_DIR) dev

build:
	@$(MAKE) -C $(BACKEND_DIR) build

test:
	@$(MAKE) -C $(BACKEND_DIR) test

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

docker-build:
	@docker compose build

docker-up:
	@docker compose up --build -d

docker-down:
	@docker compose down

docker-logs:
	@docker compose logs -f
