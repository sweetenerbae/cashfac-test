# Cash Factories Test Task

Репозиторий организован как монорепа.

## Структура

- `backend` - Go API
- `frontend` - клиентская часть

## Быстрый старт

```bash
make run
```

Команда запускает backend из корня репозитория.

Swagger UI:

- `http://localhost:8080/docs`

Health check:

- `http://localhost:8080/health`

## Команды

- `make run` - запуск backend
- `make dev` - запуск backend в dev-режиме
- `make build` - сборка backend
- `make test` - тесты backend
- `make fmt` - форматирование backend
- `make caddy-run` - запуск Caddy для backend
- `make frontend-install` - установка frontend-зависимостей
- `make frontend-dev` - запуск Vite dev server
- `make frontend-build` - сборка frontend
