# Cash Factories Backend

Сервис получает новости из внешних источников, сохраняет их и отдает тексты в разных эмоциональных режимах.

## Структура

- `cmd/api` - точка входа
- `internal/app` - сборка зависимостей и запуск HTTP-сервера
- `internal/config` - конфигурация
- `internal/domain` - доменные модели и интерфейсы
- `internal/usecase` - бизнес-логика
- `internal/usecase/storage` - in-memory репозиторий для локального запуска
- `internal/usecase/source` - адаптер источника новостей
- `internal/usecase/rewriter` - адаптер сервиса рерайта
- `internal/transport/http` - HTTP API
- `Caddyfile` - конфиг локального reverse proxy
- `Makefile` - команды для локальной работы

## Эндпоинты

- `GET /health`
- `POST /api/v1/news/sync?mood=neutral`
- `GET /api/v1/news?mood=neutral`
- `GET /api/v1/news/{id}`
- `GET /docs`
- `GET /openapi.yaml`

## Запуск

```bash
make run
```

По умолчанию сервис стартует на `:8080`.

Для загрузки реальных новостей нужен ключ Guardian Open Platform:

```bash
GUARDIAN_API_KEY=your_key make run
```

## Docker

Из корня репозитория:

```bash
cp .env.example .env
make docker-up
```

После запуска:

- `http://localhost:8080/health`
- `http://localhost:8080/docs`
