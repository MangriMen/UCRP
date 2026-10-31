# apps/api — Go backend (UCRP)

Модуль: `github.com/MangriMen/UCRP`.  
Слои: `handler → service → domain ← repository`; capability-пакеты зависят только от `domain` / `platform/*`.  
Правила — `docs/ARCHITECTURE.md`, дерево — `docs/appendix/go-package-tree.md`, API — `docs/appendix/endpoint-registry.md`.

## Структура

```text
apps/api/
├── cmd/api/            HTTP-сервер
├── cmd/worker/         воркеры + планировщик
├── internal/
│   ├── domain/         сущности, ошибки, интерфейсы репозиториев и портов
│   ├── handler/        HTTP + dto/
│   ├── service/        сценарии
│   ├── repository/     SQL + sqlstore/
│   ├── ingestion/      конвейер импорта + connector/
│   ├── identity/       резолвер склеек + normalize/
│   ├── segments/       DSL правил, evaluator
│   ├── recommend/      скоринг + factors/
│   ├── export/         каналы и форматы выгрузки
│   └── platform/       auth, httpx, queue, cache, storage, config, logging, scheduler
├── migrations/         goose (common/ + postgres/)
└── docker/             Dockerfile + docker-compose.yml
```

## Команды

```sh
cd apps/api
go build ./...
go run ./cmd/api      # scaffold
go run ./cmd/worker
go test ./...
```

Docker:

```sh
cd apps/api
docker compose -f docker/docker-compose.yml up --build
# API: http://localhost:8080/healthz (когда реализован)
```

## Базовые зависимости

Выбор продиктован архитектурой (SQLite → PG, очередь на таблице, JWT, без CGO в Docker).  
В `go.mod` подключаются **после первого импорта** в коде; здесь — согласованный список.

| Пакет                                                      | Зачем                                                                                     | Где в системе                          |
| ---------------------------------------------------------- | ----------------------------------------------------------------------------------------- | -------------------------------------- |
| `github.com/go-chi/chi/v5`                                 | Лёгкий роутер + middleware (RequireRole, timeouts)                                        | `handler`, `platform/httpx`            |
| `github.com/mattn/go-sqlite3` **или** `modernc.org/sqlite` | SQLite (старт). Предпочтительнее **modernc** — pure Go, без CGO, проще Docker multi-stage | `repository/sqlstore`                  |
| `github.com/jackc/pgx/v5`                                  | PostgreSQL (цель), пул соединений                                                         | `repository/sqlstore`                  |
| `github.com/golang-jwt/jwt/v5`                             | Access/refresh JWT (claims: sub, role, tenant_id)                                         | `platform/auth`                        |
| `golang.org/x/crypto`                                      | bcrypt хэши паролей                                                                       | `platform/auth`, `service.UserService` |
| `github.com/google/uuid`                                   | UUIDv7 (NewV7) — идентификаторы из приложения (D-1)                                       | `domain` / `platform` (IDGen)          |
| `github.com/pressly/goose/v3`                              | Миграции (13 файлов, common + postgres)                                                   | `migrations`, `cmd/*`                  |
| `golang.org/x/sync`                                        | `errgroup` — параллельные чанки импорта (≤4 горутины)                                     | `ingestion`                            |
| `github.com/xuri/excelize/v2`                              | Чтение XLSX наравне с CSV (US-1)                                                          | `ingestion/connector`                  |
| `github.com/stretchr/testify`                              | assert/require/mock — тесты service и repository                                          | `*_test.go`                            |

**Из stdlib (без внешних пакетов):** `net/http` (или chi поверх), `database/sql`, `encoding/csv`, `encoding/json`, `log/slog`, `context`, `time`, `os`.

**Не берём на старте:** ORMs, GraphQL, Redis, Kafka/NATS (очередь — таблица `job_queue`), собственный mailer/SMS (отдаём CSV/API провайдеру).

### Пороги и константы (из docs, не из внешних библиотек)

| Константа                   | Значение                               | Источник                  |
| --------------------------- | -------------------------------------- | ------------------------- |
| Лимит строк импорта         | 100_000 → 413                          | US-1, реестр              |
| Порог автосклейки           | score ≥ 80                             | AD-2, Q-12                |
| Порог скрытия ячеек R-3     | k = 5                                  | US-8, DFD                 |
| Окно атрибуции по умолчанию | 14 дней                                | campaign.attribution_days |
| Таймаут preview сегмента    | 10 с → 504                             | US-3, SD-4                |
| Синхронная выгрузка         | 25 с; >20k членов → 202 `export.build` | AD-5, Q-23                |
| Опрос UI                    | import 2 с; segments/reco 5 с          | US-1, US-6                |

## Env (см. `.env.example` в корне)

| Переменная          | Назначение                                      |
| ------------------- | ----------------------------------------------- |
| `UCRP_HTTP_ADDR`    | напр. `:8080`                                   |
| `UCRP_DATABASE_URL` | `sqlite:///data/ucrp.db` → позже `postgres://…` |
| `UCRP_JWT_SECRET`   | подпись JWT                                     |
| `UCRP_UPLOAD_DIR`   | BlobStore (D-28)                                |
| `UCRP_LOG_LEVEL`    | debug…error                                     |
| `UCRP_BOOTSTRAP_MANAGER_EMAIL` | Email первого управляющего УК, если в базе ещё нет менеджеров |
| `UCRP_BOOTSTRAP_MANAGER_PASSWORD` | Пароль первого управляющего УК; задаётся вместе с email |

Обе переменные bootstrap необязательны, но должны быть заданы вместе. При первом запуске на базе без менеджеров создаются учётная запись и персональный грант на данные; операция выполняется один раз. Наличие любого менеджера, включая неактивного, отключает bootstrap. Повторный запуск не восстанавливает отозванный грант. Пароль храните в секретах окружения, не добавляйте его в Git. Для отдельного PostgreSQL-теста bootstrap можно задать `UCRP_BOOTSTRAP_TEST_DATABASE_URL`; тест создаёт и удаляет изолированную схему.

## Docker

| Файл                        | Назначение                                                               |
| --------------------------- | ------------------------------------------------------------------------ |
| `docker/Dockerfile`         | multi-stage: `golang:1.24-alpine` → `alpine:3.21`; `ARG APP=api\|worker` |
| `docker/docker-compose.yml` | сервисы `api` и `worker`, shared volume `api-data` (SQLite + uploads)    |
| `docker/.dockerignore`      | контекст = `apps/api`                                                    |

Образ без root (uid 10001), CGO off, healthcheck на `/healthz` (появится с реализацией cmd/api).

## Статус

- [x] Дерево пакетов по `go-package-tree.md`
- [x] `go.mod` (`github.com/MangriMen/UCRP`), заглушки `cmd/*`, package docs
- [x] Dockerfile + compose в `docker/` (api + worker, volume `api-data`)
- [x] `migrations/ddl.sql` — копия `docs/appendix/ddl.sql` (source of truth в docs)
- [ ] Реализация слоёв (по этапам / базовому промпту)
- [ ] goose-файлы `migrations/common/00001_…` по `migrations-plan.md`
