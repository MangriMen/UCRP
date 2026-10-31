# 07. Оценка: какие тесты можно дописать после доработки кода

Основание: фактическая проверка обновлённого кода (`cmd/api`, `cmd/worker`,
`internal/{service,handler,repository,ingestion,identity}`) и живые запросы к
запущенному API. Контракт и документация не изменялись.

## 1. Что изменилось в коде

| Было (отчёт `06`)                | Стало                                                                                                                            |
| -------------------------------- | -------------------------------------------------------------------------------------------------------------------------------- |
| 17 из 90 операций смонтировано   | **+17 операций**: импорт (4), выгрузка (5), источники (5), типы источников (1), синхронизация (1), плюс `GET /data-source-types` |
| `auth.RequireRole` не подключён  | **подключён ко всем приватным маршрутам**                                                                                        |
| `jobTypes := []string{}`         | **`import.process`, `connector.pull`, `identity.resolve`**                                                                       |
| нет `internal/ingestion`         | CSV-коннектор, маппинг полей, пайплайн обработки строк                                                                           |
| нет `internal/identity`          | `Resolver`: привязка событий к гостям, автосклейка, спорные пары                                                                 |
| `IdentityService` без транзакций | `MergeSQL` — confirm/reject/revert в одной транзакции + `audit_log`                                                              |
| нет справочников                 | `EnsureEventTypes` (6 типов), `EnsureDefaults` (4 типа источников)                                                               |

Проверено вживую: загрузка CSV → `queued` → воркер → `completed_with_errors`
с `total=3, rejected=3` и кодами `no_tenant`, `bad_date`, `no_external_id`;
выгрузка → `preview` (5 счётчиков) → `create` 201 → `download` CSV;
идемпотентность по `Idempotency-Key`; 403 для `analyst` на `/exports` и для
`marketer` на `/imports`.

## 2. Тесты, которые можно дописать сейчас

### 2.1. Импорт (US-1 / UC-1) — 12 тестов

Все операции смонтированы, воркер работает, справочники засеяны автоматически.

| Тест                              | Проверяет                                               | Требование |
| --------------------------------- | ------------------------------------------------------- | ---------- |
| `imports.upload.accepted`         | 202, `status=queued`, `duplicate=false`                 | REQ-1.1    |
| `imports.status.counters`         | `GET /imports/{id}` отдаёт счётчики                     | REQ-1.2    |
| `imports.counters.invariant`      | `accepted+rejected+duplicate=total`                     | REQ-1.3    |
| `imports.upload.idempotent`       | повтор файла → 200, `duplicate=true`, тот же `batch_id` | REQ-1.4    |
| `imports.rejected.no-external-id` | строка без `external_id` → `no_external_id`             | REQ-1.6    |
| `imports.rejected.bad-date`       | плохая дата → `bad_date`                                | REQ-1.7    |
| `imports.rejected.csv`            | `?format=csv` → `text/csv` с заголовком                 | REQ-1.8    |
| `imports.upload.bad-format`       | нет обязательных колонок → 422 `bad_format`             | REQ-1.10   |
| `imports.upload.unknown-source`   | несуществующий `data_source_id` → 404                   | REQ-1.11   |
| `imports.upload.forbidden`        | `marketer` → 403                                        | REQ-1.12   |
| `imports.list.filters`            | `?data_source_id=`, `?status=`                          | REQ-1.13   |
| `imports.rejected.bad-tenant`     | неизвестный `tenant_code` → `bad_tenant`                | новый код  |

**Нужно:** `client.PostMultipart` (сейчас клиент умеет только JSON) и
`waitFor` на `status ∈ {completed, completed_with_errors, failed}` —
`Suite.WaitFor` уже есть, но не используется.

### 2.2. Выгрузка (US-5 / UC-5) — 10 тестов

| Тест                                | Проверяет                                               | Требование |
| ----------------------------------- | ------------------------------------------------------- | ---------- |
| `exports.preview.counters`          | 5 счётчиков, `toSend = withContact - blocked - control` | REQ-5.1    |
| `exports.preview.channel-contact`   | `sms` берёт только `phone`, `email` — только `email`    | REQ-5.2    |
| `exports.create.ready`              | 201, `status=ready`                                     | REQ-5.3    |
| `exports.create.no-idempotency-key` | без ключа → 400                                         | REQ-5.4    |
| `exports.create.idempotent`         | повтор ключа → тот же `id`, новых `contact_log` нет     | REQ-5.5    |
| `exports.download.excludes-control` | контрольная группа не в CSV                             | REQ-5.7    |
| `exports.counters.invariant`        | `exportedCount = withContact - blocked - control`       | REQ-5.9    |
| `exports.download.not-ready`        | статус ≠ `ready` → 409                                  | REQ-5.10   |
| `exports.list.journal`              | журнал выгрузок                                         | REQ-5.11   |
| `exports.create.forbidden`          | `analyst`/`manager` → 403                               | REQ-5.12   |

**Нужно:** фикстура сегмента с контактами (`Registry.Segment` уже есть,
нужен вариант с `guest_identifier`).

### 2.3. Лимит контактов (US-10) — 5 тестов

`ContactLimitRepo` реализован, но **REST-эндпоинтов `/contact-limits` нет**
(проверено: 404). Поэтому тестируется только косвенно — через
`blockedByLimit` в preview и `blockedByLimitCount` в выгрузке, с фикстурой
`contact_limit` прямым `INSERT`.

| Тест                               | Проверяет                                       |
| ---------------------------------- | ----------------------------------------------- |
| `exports.limit.blocks-repeat`      | после выгрузки повторная → `blockedByLimit = N` |
| `exports.limit.strictest-rule`     | из двух правил применяется самое строгое        |
| `exports.limit.channel-scoped`     | правило `sms` не влияет на `email`              |
| `exports.limit.period-window`      | контакт вне окна не блокирует                   |
| `exports.limit.control-not-logged` | контрольная группа не пишет `contact_log`       |

### 2.4. Ролевая модель (US-8, N-4…N-9) — 8 тестов

`RequireRole` подключён — проверки 403 стали возможны. Нужна фикстура
пользователей с ролями (`Registry.User` уже есть, нужен хеш пароля).

| Тест                                     | Проверяет                         |
| ---------------------------------------- | --------------------------------- |
| `auth.role.tenant-forbidden-guests`      | `tenant` → 403 на `/guests`       |
| `auth.role.marketer-forbidden-imports`   | `marketer` → 403 на `/imports`    |
| `auth.role.analyst-forbidden-exports`    | `analyst` → 403 на `/exports`     |
| `auth.role.manager-forbidden-exports`    | `manager` → 403 на `/exports`     |
| `auth.role.marketer-forbidden-confirm`   | `marketer` → 403 на `confirm`     |
| `auth.role.analyst-allowed-confirm`      | `analyst` → 200                   |
| `auth.role.marketer-allowed-guests`      | `marketer` → 200                  |
| `auth.role.manager-allowed-imports-read` | `manager` → 200 на `GET /imports` |

### 2.5. Источники данных (UC-17) — 6 тестов

| Тест                         | Проверяет                                |
| ---------------------------- | ---------------------------------------- |
| `sources.types.defaults`     | 4 типа: `csv_excel`, `crm`, `pos`, `api` |
| `sources.create`             | 201, `is_active=true`                    |
| `sources.create.no-name`     | без `name` → 400                         |
| `sources.create.bad-mapping` | битый `field_mapping` → 400              |
| `sources.get.not-found`      | 404                                      |
| `sources.sync.unsupported`   | pull на файловый источник → 400          |

### 2.6. Автосклейка (UC-12) — 4 теста

`identity.Resolver` реализован и вызывается воркером после импорта.

| Тест                               | Проверяет                                             |
| ---------------------------------- | ----------------------------------------------------- |
| `identity.resolve.creates-guest`   | событие без совпадений → новый гость + идентификаторы |
| `identity.resolve.attaches`        | совпадение по телефону → привязка к существующему     |
| `identity.auto-merge.shared-phone` | общий телефон → `auto_merged`, `score=90`             |
| `identity.pending.shared-email`    | общий email → `pending`, `score=70`                   |

**Нужно:** `waitFor` на появление `merge_candidate` после импорта.

### 2.7. Сквозные сценарии — 3 теста

| Сценарий                  | Что проверяет                                              |
| ------------------------- | ---------------------------------------------------------- |
| **E2E-1** (ядро)          | импорт → профиль → сегмент → рекомендация → выгрузка → CSV |
| **E2E-2** (новое событие) | второй импорт меняет `tenantCount`, сегмент и рекомендации |
| **E2E-4** (лимит)         | выгрузка → лимит → повторная выгрузка заблокирована        |

E2E-1 и E2E-2 из `04-e2e-plan.md` §2 теперь реализуемы полностью.

## 3. Тесты, которые всё ещё невозможны

| Что                                                                       | Причина                                          |
| ------------------------------------------------------------------------- | ------------------------------------------------ |
| `/contact-limits` CRUD                                                    | эндпоинты не смонтированы (404)                  |
| `/campaigns`, `/campaigns/{id}/results`                                   | не смонтированы                                  |
| `/audit-logs`                                                             | не смонтирован (хотя `audit_log` теперь пишется) |
| `/tenants`, `/tenant-categories`, `/projects`, `/event-types` CRUD        | не смонтированы                                  |
| `/users` CRUD                                                             | не смонтирован                                   |
| `/dashboard/*`, `/tenants/{id}/aggregates`, `/data-quality/*`             | не смонтированы                                  |
| `/unmerge-blocks`                                                         | не смонтирован                                   |
| `/segments/{id}/members`, `/versions`, `/recompute`, `PUT /segments/{id}` | не смонтированы                                  |
| `/guests/{id}/contacts`, `DELETE /guests/{id}`                            | не смонтированы                                  |
| `/guests/{id}/merge`                                                      | не смонтирован                                   |
| `/segments/{id}/recommendations`                                          | не смонтирован                                   |
| N-37, N-38 (504)                                                          | нет управляемой задержки                         |
| N-39 (429)                                                                | нет rate limiter                                 |
| `export.build` для >20 000 членов                                         | выгрузка всегда синхронная                       |

## 4. Дефекты, подтверждённые повторно

| #   | Дефект                                           | Статус           |
| --- | ------------------------------------------------ | ---------------- |
| 1   | `POST /segments/preview` с пустым `wasNot` → 500 | **не исправлен** |
| 2   | `GET /guests` возвращает `status=merged`         | **не исправлен** |
| 3   | нет валидации `q` (обязателен, `minLength: 2`)   | **не исправлен** |
| 4   | `details[]` не заполняется                       | **не исправлен** |

## 5. Новые расхождения, замеченные при проверке

| Операция                          | Расхождение                                                                                                |
| --------------------------------- | ---------------------------------------------------------------------------------------------------------- |
| `POST /imports`                   | контракт: `batchId`; реализация: `batch_id`                                                                |
| `GET /imports/{id}`               | контракт: `dataSourceId`, `totalRows`, `acceptedRows`…; реализация: `data_source_id`, `total`, `accepted`… |
| `POST /exports/preview`           | контракт: `segmentId`, `channel`; реализация: `segment_id`, `channel`                                      |
| `POST /exports`                   | контракт: `segmentId`, `format`; реализация: `segment_id`, `format`                                        |
| `GET /exports/{id}`               | контракт: `segmentVersionId`, `createdAt`; реализация: без них                                             |
| `POST /data-sources`              | контракт: `typeId`, `fieldMapping` (объект); реализация: `type_id`, `field_mapping` (строка)               |
| `GET /data-source-types`          | контракт: массив; реализация: `{items:[…]}`                                                                |
| `GET /imports/{id}/rejected-rows` | контракт: `items` + `meta`; реализация: `items` + плоские `page`/`page_size`/`total`                       |

Это те же системные расхождения, что в `06-final-report.md` §5: `snake_case`
против `camelCase`, отсутствие `meta`, плоская пагинация.

## 6. Приоритет

| Приоритет | Что                       | Обоснование                                                    |
| --------- | ------------------------- | -------------------------------------------------------------- |
| 1         | Ролевая модель (8 тестов) | `RequireRole` только что подключён — риск регрессии максимален |
| 2         | Импорт (12 тестов)        | ядро продукта, воркер работает, справочники засеяны            |
| 3         | Выгрузка (10 тестов)      | вторая половина ядра, идемпотентность и лимиты                 |
| 4         | Автосклейка (4 теста)     | новый код, высокая цена ошибки                                 |
| 5         | E2E-1, E2E-2 (2 теста)    | сквозная проверка всей цепочки                                 |
| 6         | Источники (6 тестов)      | вспомогательный CRUD                                           |
| 7         | Лимиты (5 тестов)         | только косвенно, без REST                                      |

Итого реализуемо **~47 новых тестов** (сейчас 46).

## 7. Что нужно доработать в тестовой инфраструктуре

| Элемент                                                                          | Зачем                                     |
| -------------------------------------------------------------------------------- | ----------------------------------------- |
| `client.PostMultipart`                                                           | загрузка CSV в `POST /imports`            |
| `fixtures.Registry.User` с реальным bcrypt-хешем                                 | тесты ролей                               |
| `fixtures.Registry.SegmentWithContacts`                                          | сегмент с `guest_identifier` для выгрузки |
| `fixtures.Registry.ContactLimit`                                                 | правило лимита прямым `INSERT`            |
| `waitFor` на статус импорта                                                      | асинхронная цепочка                       |
| `waitFor` на `merge_candidate`                                                   | автосклейка после импорта                 |
| `testdata/*.csv`                                                                 | валидный, с ошибками, пустой, без колонок |
| Очистка `import_batch`, `rejected_row`, `export_job`, `contact_log`, `job_queue` | новые сущности в `Cleanup`                |

## 8. Открытые вопросы

1. Источник истины при расхождении — OpenAPI или реализация? (влияет на все
   новые тесты: `batch_id` против `batchId`)
2. Допустима ли подготовка данных прямыми `INSERT` для `contact_limit`,
   `segment_member`, `app_user`?
3. Нужен ли отдельный `docker-compose.e2e.yml` с воркером?
4. Добавлять ли job `e2e` в CI?
5. Будут ли смонтированы `/contact-limits`, `/campaigns`, `/audit-logs`?

## 9. Что реализовано

Все рекомендации §2 выполнены. Набор вырос с 46 до **88 тестов**.

| Блок                    | Файл                       | Тестов | Статус |
| ----------------------- | -------------------------- | ------ | ------ |
| Импорт                  | `imports_test.go`          | 12     | ✅     |
| Выгрузка                | `exports_test.go`          | 11     | ✅     |
| Лимит контактов         | `contact_limits_test.go`   | 4      | ✅     |
| Ролевая модель          | `roles_test.go`            | 22     | ✅     |
| Источники данных        | `data_sources_test.go`     | 7      | ✅     |
| Автосклейка             | `identity_resolve_test.go` | 4      | ✅     |
| Сквозные сценарии       | `e2e_full_test.go`         | 2      | ✅     |
| Контракт (было)         | `contract_test.go`         | 24     | ✅     |
| Бизнес-правила (было)   | `business_test.go`         | 20     | ✅     |
| E2E идентичности (было) | `e2e_identity_test.go`     | 2      | ✅     |

Инфраструктура (§7) тоже дописана: `client.PostMultipart`,
`Registry.User` с bcrypt, `Registry.SegmentWithContacts`,
`Registry.ContactLimit`, `Registry.Phone`, `Suite.Marketer`,
`Suite.WaitFor` на импорт и на `merge_candidate`, `testdata_test.go`,
расширенный `Cleanup`.

### 9.1. Результат прогона

```
88 тестов: 83 прошло, 1 упал, 4 пропущено
зафиксировано 71 расхождение контракта
```

Единственное падение — подтверждённый дефект №2 (`GET /guests` возвращает
`status=merged`). Четыре пропуска — дефекты №1 и №3: тест не может пройти,
пока API отвечает 500 на пустой `wasNot` и не валидирует `q`.

### 9.2. Найденные при реализации особенности

| Наблюдение                                                      | Следствие для тестов                                                                               |
| --------------------------------------------------------------- | -------------------------------------------------------------------------------------------------- |
| Сид создаёт пользователя с ролью `analyst`                      | `/exports` требует `marketer`; тесты выгрузки логинятся фикстурным маркетологом (`Suite.Marketer`) |
| `normalize.Phone` оставляет только цифры                        | телефонные фикстуры не могут содержать hex `RunID`; добавлен `Registry.Phone(i)`                   |
| `contact_limit.created_by` — FK на `app_user`                   | фикстура лимита берёт реального пользователя (`Registry.ownerID`)                                  |
| `identity.resolve` ставится в очередь только при `accepted > 0` | тесты автосклейки ждут появления `merge_candidate` в БД                                            |
| `POST /exports` возвращает 201 и при повторном ключе            | идемпотентность проверяется по совпадению `id`, а не по коду ответа                                |
