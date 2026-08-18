# Cash Factories Test Backend

Сервис для тестового задания Cash Factories: получение реальных новостей, сохранение и выдача версий текста в разных эмоциональных режимах.

## Архитектура

- `cmd/api` - точка входа.
- `internal/app` - сборка зависимостей и запуск HTTP-сервера.
- `internal/config` - конфигурация.
- `internal/domain` - доменные модели и интерфейсы.
- `internal/usecase` - бизнес-логика.
- `internal/usecase/storage` - in-memory репозиторий для локального запуска.
- `internal/usecase/source` - адаптер источника новостей.
- `internal/usecase/rewriter` - адаптер сервиса рерайта.
- `internal/transport/http` - HTTP API.
- `Caddyfile` - конфиг локального reverse proxy.
- `Makefile` - команды для разработки и запуска.

## Что уже есть

- `GET /health`
- `POST /api/v1/news/sync?mood=neutral`
- `GET /api/v1/news?mood=neutral`
- `GET /api/v1/news/{id}`
- `GET /docs`
- `GET /openapi.yaml`

## Как запустить

```bash
make run
```

По умолчанию сервис стартует на `:8080`.

Swagger UI будет доступен по адресу `http://localhost:8080/docs`.

## Make targets

- `make fmt` - форматирование Go-кода.
- `make build` - сборка бинарника в `./bin`.
- `make run` - запуск API напрямую.
- `make dev` - локальный запуск с `HTTP_PORT=8080`.
- `make test` - запуск тестов.
- `make clean` - очистка артефактов.
- `make swagger` - подсказка по Swagger endpoint'ам.
- `make caddy-validate` - проверка `Caddyfile`.
- `make caddy-run` - запуск Caddy как reverse proxy.

## Локальная схема

1. Поднять API: `make run`
2. В другом терминале поднять Caddy: `make caddy-run`
3. Открыть `http://localhost/health` и `http://localhost/docs`
