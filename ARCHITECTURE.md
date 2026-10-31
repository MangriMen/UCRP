# UCRP Architecture Guidelines

Правила структуры кода, направления зависимостей и границ модулей. Обязательны для людей и AI-агентов.

Product requirements, design diagrams and the OpenAPI contract live in `docs/`. This document describes **how the code is organized**, not **what the system does**; it is written in Russian like the rest of the design documents.

---

## 1. Общая схема

```text
apps/web (React SPA)  ──REST /api/v1──▶  apps/api (Go)
                                            ├── cmd/api      HTTP-сервер
                                            ├── cmd/worker   воркеры и планировщик
                                            ├── internal/    домен и сценарии
                                            └── migrations/  goose
```

Монорепозиторий на Bun workspaces + Turborepo. Go-модуль один на репозиторий (`github.com/MangriMen/UCRP`), живёт в `apps/api`.

---

## 2. Backend: слои и направление зависимостей

Зависимости направлены строго внутрь:

$$\text{handler} \longrightarrow \text{service} \longrightarrow \text{domain} \longleftarrow \text{repository}$$

| Слой         | Ответственность                                                                   | Может импортировать                                    | Не может                                                |
| ------------ | --------------------------------------------------------------------------------- | ------------------------------------------------------ | ------------------------------------------------------- |
| `domain`     | Сущности, value objects, инварианты, ошибки, **интерфейсы** репозиториев и портов | только стандартную библиотеку                          | всё остальное                                           |
| `service`    | Сценарии (use cases), транзакционные границы, координация                         | `domain`                                               | `handler`, `repository` (только интерфейсы из `domain`) |
| `repository` | SQL-реализации интерфейсов `domain`                                               | `domain`, `repository/sqlstore`                        | `service`, `handler`                                    |
| `handler`    | HTTP: роутинг, разбор запросов, коды ответов, DTO                                 | `service`, `domain`, `platform/httpx`, `platform/auth` | `repository`                                            |
| `cmd/*`      | Композиция: конфиг, DI, graceful shutdown                                         | всё                                                    | —                                                       |

Пакеты бизнес-способностей (`ingestion`, `identity`, `segments`, `recommend`, `export`) зависят только от `domain` и `platform/*`. Друг друга напрямую не вызывают — взаимодействие через `service` или очередь заданий.

### Целевое дерево `apps/api`

```text
apps/api/
├── cmd/
│   ├── api/            # HTTP-точка входа
│   └── worker/         # воркеры: очередь заданий, планировщик
├── internal/
│   ├── domain/         # сущности, value objects, ошибки, интерфейсы портов
│   ├── access/         # политика прав: роли, гранты, разрешения
│   ├── handler/        # HTTP-слой, роутер /api/v1, dto/
│   ├── service/        # сценарии
│   ├── repository/     # SQL-реализации, sqlstore/ (подключение, UnitOfWork)
│   ├── ingestion/      # конвейер импорта, connector/ (csv, crm, pos, api)
│   ├── identity/       # нормализация ключей, резолвер, политика склейки
│   ├── segments/       # компилятор правил, вычислитель
│   ├── recommend/      # скоринг и обоснования
│   ├── export/         # каналы и форматы выгрузки
│   └── platform/       # auth, httpx, queue, cache, storage, config, logging
├── migrations/         # goose
└── go.mod
```

---

## 3. Frontend: вертикальные срезы

`apps/web/src` организован по FSD (облегчённый, без лишних слоёв):

- **`app/`** — композиция: провайдеры, роутер, layout, меню по роли, глобальные стили.
- **`pages/`** — по папке на маршрут; только композиция фич и сущностей.
- **`widgets/`** — создаётся только когда блок переиспользуется на 2+ страницах.
- **`features/`** — действия пользователя (`login`, `upload-import`, `confirm-merge`, `build-segment`, `create-export`, …).
- **`entities/`** — бизнес-сущности (`guest`, `segment`, `import-batch`, `user`, …): типы, api-хуки, моки, малые UI-блоки.
- **`shared/`** — технические срезы без бизнес-логики: `api/` (fetch-клиент, типы из OpenAPI, моки MSW), `ui/`, `lib/`, `config/`.

Внутри среза сегменты `ui/ api/ model/ mocks/ lib/` создаются по необходимости; публичный API — `index.ts`.

### Правила

| Правило                    | Описание                                                                                       |
| -------------------------- | ---------------------------------------------------------------------------------------------- |
| Единственный вход к данным | Компоненты не вызывают `fetch` напрямую — только хуки из `entities/*/api` или `features/*/api` |
| Направление импортов       | `app → pages → widgets → features → entities → shared`. Обратные запрещены (oxlint)            |
| Изоляция срезов            | Срезы одного слоя не импортируют друг друга; общее выносится ниже по слоям                     |
| Серверное состояние        | TanStack Query; глобальный стор — только для аутентификации                                    |
| Проверка прав              | UI скрывает недоступное, но решающая проверка — на backend (403)                               |

---

## 4. Общие пакеты (`packages/`)

| Пакет               | Назначение                                                     |
| ------------------- | -------------------------------------------------------------- |
| `packages/api-docs` | Swagger UI, собираемый из `docs/openapi.json` (`bun run docs`) |

Новый пакет попадает в `packages/` только если его использует **два и более** приложения. Иначе код живёт внутри приложения.

---

## 5. Тестирование

| Уровень           | Что тестируем              | Инструменты               |
| ----------------- | -------------------------- | ------------------------- |
| `domain` (Go)     | Инварианты, чистые функции | `testing`, без БД и сети  |
| `service` (Go)    | Сценарии с моками портов   | `testing`, `testify/mock` |
| `repository` (Go) | SQL против реальной SQLite | `testing`, временная БД   |
| `handler` (Go)    | HTTP-контракты             | `httptest`                |
| Frontend          | Компоненты и хуки          | Vitest, Testing Library   |

Имена тестов описывают ожидаемое поведение и условие: `TestMerge_RejectsConflictingNames_WhenPhoneMatches`.

---

## 6. Комментарии в коде

1. **Не оставляйте комментарии в коде.** Код должен читаться сам. Удаляйте «что делает строка», narration, TODO-шум, placeholder-фразы.
2. Публичный godoc — только если без него невалидна экспортируемая сигнатура; предпочтительно минимум. `doc.go` в скелете — голый `package name`.
3. Стиль и упрощения — по [`.agents/skills/code-simplification/SKILL.md`](.agents/skills/code-simplification/SKILL.md) (ясность > изящество, поведение не меняем, без нейрослопа).

---

## 7. Правила для AI-агентов

1. **Не раздувайте дерево.** Новый файл — только когда существующий стал нечитаемым (> ~400 строк) или появился новый связный модуль.
2. **Соблюдайте направление зависимостей.** Импорт `repository` из `service` или `handler` — ошибка, даже если компилируется.
3. **Не меняйте видимость ради тестов.** Используйте тесты в том же пакете.
4. **Контракт API — источник истины: `docs/openapi.json` и `docs/openapi/`.** Меняя эндпоинт, обновляйте спеку и перегенерируйте типы фронтенда (`gen:api` в `apps/web`).
5. **Не добавляйте зависимость** без проверки, что её нет в стандартной библиотеке или уже подключённых пакетах.
6. **Код без комментариев и без нейрослопа** — `AGENTS.md` + skill `code-simplification`.
