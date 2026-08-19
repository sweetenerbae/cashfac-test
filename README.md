# Cash Factories News

Сервис собирает реальные новости The Guardian и позволяет читать один материал в разных интонациях: нейтральной, радостной, грустной и ироничной. Исходная публикация, авторские факты и ссылка на источник остаются доступны на странице новости.

## Быстрый запуск без Make

Для основного варианта запуска нужны Git и запущенный Docker Desktop. GNU Make устанавливать не обязательно.

Клонируйте проект:

```text
git clone https://github.com/sweetenerbae/cashfac-test.git
cd cashfac-test
```

### Windows PowerShell

```powershell
if (-not (Test-Path .env)) { Copy-Item .env.example .env }
notepad .env
docker compose up --build -d
```

### macOS и Linux

```bash
[ -f .env ] || cp .env.example .env
${EDITOR:-nano} .env
docker compose up --build -d
```

Перед запуском добавьте в `.env` два ключа:

```dotenv
GUARDIAN_API_KEY=...
ZAI_API_KEY=...
```

- `GUARDIAN_API_KEY` нужен для загрузки реальных публикаций через [The Guardian Open Platform](https://open-platform.theguardian.com/access/).
- `ZAI_API_KEY` нужен для рерайта через [Z.ai](https://z.ai/).

Если контейнеры были запущены до изменения `.env`, перезапустите их:

```text
docker compose down
docker compose up --build -d
```

После запуска доступны:

| Сервис | Адрес |
| --- | --- |
| Сайт | http://localhost:8080 |
| Swagger UI | http://localhost:8080/docs |
| OpenAPI | http://localhost:8080/openapi.yaml |
| Health check | http://localhost:8080/health |

Основные команды Docker одинаковы в PowerShell, macOS и Linux:

| Действие | Команда |
| --- | --- |
| Собрать и запустить | `docker compose up --build -d` |
| Посмотреть состояние | `docker compose ps` |
| Смотреть логи | `docker compose logs -f` |
| Остановить | `docker compose down` |
| Остановить и удалить данные | `docker compose down -v` |

SQLite хранится в Docker volume `news_data`, поэтому кэш рерайтов переживает обычный перезапуск контейнеров. Для полностью чистого запуска можно удалить контейнеры вместе с volume:

```bash
docker compose down -v
```

## Запуск через Make

Makefile содержит короткие обёртки над теми же Docker-командами. Этот способ необязателен.

На macOS и Linux обычно достаточно установленного `make`. На Windows команда `make --version` должна выполняться из PowerShell; если GNU Make не установлен или не добавлен в `PATH`, используйте запуск без Make выше.

```text
make setup
make docker-up
```

`make setup` создаёт `.env` из `.env.example` и не перезаписывает существующий файл. После первого вызова заполните ключи в `.env`, затем запускайте `make docker-up`.

Управление контейнерами через Make:

```text
make docker-status
make docker-logs
make docker-down
```

## Локальный запуск

Для запуска без Docker нужны Go 1.25 и Node.js 20 или новее.

На macOS или Linux загрузите переменные окружения и запустите backend:

```bash
[ -f .env ] || cp .env.example .env
${EDITOR:-nano} .env
set -a
. ./.env
set +a
cd backend
go run ./cmd/api
```

На Windows PowerShell:

```powershell
if (-not (Test-Path .env)) { Copy-Item .env.example .env }
notepad .env
Get-Content .env | ForEach-Object {
    if ($_ -match '^\s*([^#][^=]*)=(.*)$') {
        [Environment]::SetEnvironmentVariable($matches[1].Trim(), $matches[2].Trim(), 'Process')
    }
}
Set-Location backend
go run ./cmd/api
```

Во втором терминале запустите frontend:

```text
cd frontend
npm install
npm run dev
```

Локальные адреса:

- frontend: http://localhost:5173
- backend: http://localhost:8080
- Swagger UI: http://localhost:8080/docs

Vite проксирует `/api`, `/health`, `/docs` и `/openapi.yaml` на Go-сервис, поэтому frontend использует относительные API-адреса без CORS-настроек.

## Как проверить приложение

1. Откройте http://localhost:8080 при Docker-запуске или http://localhost:5173 при локальном запуске.
2. На чистой базе приложение асинхронно загрузит 10 свежих публикаций The Guardian. Повторную загрузку можно запустить кнопкой «Загрузить свежие новости».
3. Откройте карточку новости. Сначала показывается оригинальный текст и ссылка на публикацию.
4. Выберите интонацию. Первый выбор запускает рерайт через Z.ai; результат сохраняется в SQLite.
5. Повторно выберите ту же интонацию. Backend вернёт сохранённую версию без нового LLM-запроса.
6. На странице отображаются источник результата, длительность, фактическое число LLM-запросов и число запросов, сэкономленных кэшем.

## Как работает загрузка новостей

Frontend сначала запрашивает сохранённую нейтральную ленту. Если база пустая, он создаёт фоновую sync-job через `POST /api/v1/news/sync` и опрашивает её состояние через `/api/v1/jobs/{id}`. Интерфейс остаётся доступным и показывает прогресс загрузки.

Backend получает из The Guardian заголовок, полный текст, дату, изображение, URL и внешний идентификатор публикации. Новости сохраняются в SQLite по одной, а progress job обновляется после каждой записи. После успешной синхронизации записи, которых больше нет в свежей подборке, удаляются. В базе остаётся актуальная подборка, а не бесконечный архив старых новостей.

Синхронизация не обращается к Z.ai. LLM вызывается только для конкретной новости после выбора интонации.

## Как экономятся LLM-запросы

- Рерайт создаётся лениво, а не сразу для всей ленты.
- Версия для пары `новость + интонация` сохраняется в SQLite.
- Кэш используется только пока digest исходного текста совпадает с сохранённым digest.
- Одновременные одинаковые запросы объединяются: один запрос идёт в Z.ai, остальные получают его результат.
- При ошибке автоматических повторов нет, поэтому один пользовательский выбор не запускает скрытую серию платных запросов.
- Возвращаемая `Meta` показывает `generated`, `cache` или `shared`, а также `LLMRequests`, `SavedLLMRequests` и `DurationMs`.

## Контроль исходных фактов

В system prompt модели явно запрещено менять или опускать имена, даты, числа, места, организации, цитаты и последовательность событий. Модель должна менять формулировки, ритм и эмоциональную подачу, но не содержание.

Backend дополнительно:

- отклоняет пустой ответ, ответ с технической меткой и версию, слишком близкую к оригиналу;
- связывает рерайт с checksum исходного текста;
- не использует кэш после изменения оригинала;
- всегда хранит исходный текст и URL первоисточника рядом с рерайтом;
- показывает исходник вместо сомнительной версии, если рерайт завершился ошибкой.

Checksum подтверждает, для какой версии оригинала создан рерайт, но не является семантической проверкой каждого утверждения. Автоматической NER/LLM-верификации фактов в проекте нет; основная защита строится на строгом prompt, неизменяемом оригинале и прозрачной ссылке на источник.

## Архитектура

Backend разделён на слои:

- `internal/domain` — сущности, типы ошибок и интерфейсы репозиториев и внешних клиентов;
- `internal/usecase` — сценарии загрузки, получения и рерайта новостей, а также управление фоновыми job;
- `internal/usecase/source` — клиент The Guardian и локальный stub;
- `internal/usecase/rewriter` — интеграция с Z.ai и контроль качества ответа;
- `internal/usecase/storage` — реализации хранилища на SQLite и in-memory;
- `internal/transport/http` — HTTP handlers, middleware, Swagger UI и OpenAPI;
- `internal/platform/logger` — структурированные логи без вывода ключей, prompt и текстов статей;
- `internal/app` — сборка зависимостей и запуск HTTP-сервера.

Frontend также разделён по ответственности:

- `api` — единый HTTP-клиент и API-функции;
- `hooks` — состояние страницы, polling job, кэширование и сценарии загрузки;
- `pages` — страницы ленты и отдельной новости;
- `components` — карточки, detail-view, изображения, skeleton и метрики;
- `constants` и `utils` — общие значения и форматирование.

## API

| Метод | Endpoint | Назначение |
| --- | --- | --- |
| `GET` | `/health` | Проверка доступности backend |
| `GET` | `/api/v1/news?mood=neutral` | Получение сохранённой ленты |
| `GET` | `/api/v1/news/by-external` | Получение новости по ID источника |
| `GET` | `/api/v1/news/{id}` | Получение новости по внутреннему ID |
| `POST` | `/api/v1/news/sync?limit=10` | Запуск асинхронной загрузки новостей |
| `GET` | `/api/v1/jobs/{id}` | Получение прогресса sync-job |
| `POST` | `/api/v1/news/rewrite` | Рерайт одной новости в выбранной интонации |

Полные схемы запросов и ответов находятся в Swagger UI.

## Переменные окружения

| Переменная | Обязательность | Значение |
| --- | --- | --- |
| `GUARDIAN_API_KEY` | Для реальных новостей | Ключ The Guardian Open Platform |
| `ZAI_API_KEY` | Для рерайта | API-ключ Z.ai |
| `ZAI_MODEL` | Нет | По умолчанию `glm-5.2` |
| `ZAI_BASE_URL` | Нет | Endpoint Chat Completions Z.ai |
| `SQLITE_PATH` | Нет | Локально `data/news.db`, в Docker `/data/news.db` |
| `HTTP_PORT` | Нет | По умолчанию `8080` |
| `LOG_COLOR` | Нет | `auto`, `always` или `never` |

Без `GUARDIAN_API_KEY` backend запускается со stub-источником, а без `ZAI_API_KEY` — с отключённым рерайтом. Для полноценной проверки задания нужны оба ключа.

## Проверки проекта

Без Make:

```text
cd backend
go test ./...
cd ../frontend
npm ci
npm run build
```

Через Make из корня проекта:

```text
make check
```

Список остальных Make-команд доступен через `make help`.

## Возможные проблемы при запуске

`Cannot connect to the Docker daemon` означает, что Docker Desktop не запущен. Запустите Docker Desktop, дождитесь готовности engine и повторите `docker compose up --build -d` или `make docker-up`.

Если PowerShell не находит `make`, это не блокирует запуск проекта. Используйте прямые команды `docker compose` из раздела «Быстрый запуск без Make».

Если PowerShell не находит `.env.example`, проверьте текущую папку:

```powershell
Get-Location
Get-ChildItem -Force .env*
```

Команды подготовки окружения нужно выполнять в корне `cashfac-test`, где находятся `docker-compose.yml`, `Makefile` и `.env.example`.

Если порт `8080` занят, найдите процесс на Windows:

```powershell
Get-NetTCPConnection -LocalPort 8080 -State Listen -ErrorAction SilentlyContinue |
    Select-Object LocalAddress, LocalPort, OwningProcess
```

На macOS или Linux:

```bash
lsof -nP -iTCP:8080 -sTCP:LISTEN
```

Остановите ранее запущенную локальную копию проекта или контейнер через `docker compose down` либо `make docker-down`, затем повторите запуск.

Если лента содержит stub-данные, проверьте `GUARDIAN_API_KEY` и перезапустите сервисы. Если исходник открывается, но интонация не создаётся, проверьте `ZAI_API_KEY` и сообщения `ERROR LLM` через `docker compose logs -f backend` или `make docker-logs`.
