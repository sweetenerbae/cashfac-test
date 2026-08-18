# Cashfac Test Backend

Минимальный каркас backend-сервиса на Go под тестовое задание с новостями и эмоциональным рерайтом.

## Архитектура

- `cmd/api` - точка входа.
- `internal/app` - сборка зависимостей и запуск HTTP-сервера.
- `internal/config` - конфигурация.
- `internal/domain` - доменные модели и интерфейсы.
- `internal/usecase` - бизнес-логика.
- `internal/usecase/storage` - in-memory репозиторий для локального старта.
- `internal/usecase/source` - клиент источника новостей, пока заглушка.
- `internal/usecase/rewriter` - клиент AI-рерайта, пока заглушка.
- `internal/transport/http` - HTTP API.
- `Caddyfile` - локальный reverse proxy и заготовка под серверную схему.
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
3. Открыть:
   - `http://localhost/health`
   - `http://localhost/docs`

В этой конфигурации Go-сервис слушает `127.0.0.1:8080`, а Caddy принимает внешний HTTP на `:80` и проксирует запросы внутрь.

## Серверная схема

Для серверной версии этот же `Caddyfile` можно использовать как основу:

- Caddy принимает трафик снаружи.
- Go API работает как внутренний сервис.
- при необходимости включается `auto_https` и ставится домен.

Примерно это потом можно завернуть в `systemd` или `docker-compose`, не меняя архитектуру приложения.

## Что дальше

1. Подключить реальный источник новостей: RSS/API.
2. Подключить SQLite/Postgres вместо in-memory.
3. Подключить реальный AI rewrite adapter.
4. Добавить валидацию фактов и описать стратегию в README.
5. Поднять frontend-грид новостей и экран сравнения original/rewrite.
