# Cash Factories Backend

Go-сервис, который:

- загружает реальные новости из The Guardian
- сохраняет их в SQLite
- отдает тексты в нескольких эмоциональных режимах
- хранит только актуальную подборку без накопления старых новостей

## Структура

- `cmd/api` - точка входа
- `internal/app` - инициализация приложения и HTTP-сервера
- `internal/config` - конфигурация
- `internal/domain` - доменные модели и интерфейсы
- `internal/usecase` - бизнес-логика
- `internal/usecase/source` - клиент внешнего источника новостей
- `internal/usecase/rewriter` - клиент сервиса рерайта
- `internal/usecase/storage` - SQLite и in-memory хранилища
- `internal/transport/http` - HTTP API и swagger

## Запуск

```bash
make run
```

Сервис поднимается на `http://localhost:8080`.

## Основные эндпоинты

- `GET /health`
- `GET /api/v1/news?mood=neutral`
- `GET /api/v1/news/{id}`
- `GET /api/v1/news/by-external?external_id=...`
- `GET /api/v1/news/rewrite?external_id=...&mood=...`
- `POST /api/v1/news/sync?mood=neutral&limit=10`
- `GET /api/v1/jobs/{id}`
- `GET /docs`
- `GET /openapi.yaml`

## Конфигурация

Основные переменные:

- `HTTP_PORT`
- `GUARDIAN_API_KEY`
- `ZAI_API_KEY`
- `ZAI_MODEL`
- `SQLITE_PATH`

`make run` подхватывает корневой `.env`, если он есть.

## Docker

Из корня репозитория:

```bash
make docker-up
```

После запуска:

- `http://localhost:8080/health`
- `http://localhost:8080/docs`
