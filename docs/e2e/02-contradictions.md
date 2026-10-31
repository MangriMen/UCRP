# 02. Противоречия, расхождения и пробелы

Классификация: **D↔D** — документ против документа, **C↔K** — OpenAPI против кода,
**C↔DDL** — OpenAPI против схемы БД, **GAP** — пробел (нет данных/нет реализации).
Для каждого пункта: суть, источники, влияние на тестирование, нужное уточнение.

## 1. Документ против документа (D↔D)

### D1. Конкуренты в рекомендациях: исключать или понижать

- **Что:** `docs/openapi/paths/recommendations.json` (GET `/guests/{id}/recommendations`)
  в `description` утверждает: «direct competitors (`tenant_category_relation.relationType = competitor`)
  are **excluded**». `docs/04-user-stories.md` US-4 AC-3, `docs/01-summary-assumptions.md`
  A-18, `docs/00-open-questions.md` D-34 и `docs/03-use-cases.md` UC-4 требуют
  обратного: конкурент **остаётся** в списке с понижающим коэффициентом.
- **Влияние:** тест «конкурент присутствует в выдаче» и тест «конкурент отсутствует»
  взаимоисключающие. Нельзя написать ни один, не выбрав сторону.
- **Уточнение:** подтвердить, что действует A-18/D-34 (понижение), и исправить
  `description` в OpenAPI. Дополнительно: `RecommendationFactor.factorCode` не
  содержит `competitor_penalty`, хотя глоссарий (`docs/02`) его определяет.

### D2. Статус сегмента: `active/archived` или `pending/computing/ready/failed/archived`

- **Что:** `docs/openapi/components/schemas/segments.json` → `SegmentStatus` =
  `pending|computing|ready|failed|archived`, и `Segment.status` использует этот enum.
  `docs/appendix/ddl.sql` / `apps/api/migrations/common/00001_init.sql`:
  `segment.status CHECK (status IN ('active','archived'))`, а
  `pending|computing|ready|failed` — это `segment_version.status`.
- **Влияние:** контрактный тест схемы `GET /segments/{id}` не может пройти ни при
  какой реализации: `active` не входит в enum контракта, `ready` не входит в CHECK БД.
- **Уточнение:** разделить `SegmentStatus` (active/archived) и
  `SegmentVersionStatus` (pending/computing/ready/failed) в OpenAPI.

### D3. Причина запрета склейки: `manual_reject`/`auto_conflict` или `rejected`

- **Что:** `docs/openapi/components/schemas/identity.json` → `UnmergeBlock.reason`
  enum `manual_unmerge|manual_reject|auto_conflict`. DDL:
  `CHECK (reason IN ('manual_unmerge','rejected'))`. Код
  (`internal/service/identity_service.go`) пишет `"rejected"` и `"manual_unmerge"`.
- **Влияние:** значение `manual_reject` из контракта невозможно сохранить (нарушит
  CHECK), `auto_conflict` не предусмотрен ни в DDL, ни в коде.
- **Уточнение:** привести enum контракта к DDL (`manual_unmerge`, `rejected`) либо
  расширить CHECK и добавить `auto_conflict` в реализацию.

### D4. Объёмы пилота: 80–150 арендаторов или 10–30

- **Что:** `docs/01-summary-assumptions.md` A-7: «80–150 арендаторов, ~40 000
  посетителей/мес». `docs/00-open-questions.md` Q-11: «10–30 арендаторов,
  6–12 месяцев истории, до 1–3 млн событий».
- **Влияние:** определяет объём генерации тестовых данных и необходимость
  нагрузочных проверок. На функциональные E2E не влияет, на performance — влияет.
- **Уточнение:** какое значение считать целевым для пилота.

### D5. Формат токена: `accessToken` или `access_token` — **устарело**

- **Что было:** расхождение OpenAPI (`accessToken`) и ранней реализации (`access_token`).
- **Статус:** закрыто. `internal/handler/auth_handler.go` и клиент
  `apps/web/src/shared/api/client.ts` используют camelCase `accessToken`,
  `tokenType`, `expiresIn`. OpenAPI и реализация согласованы.
- **Влияние:** тесты логина/refresh по контракту применимы без выбора стороны.

### D6. Устаревшие утверждения о состоянии проекта

- **Что:** корневой `README.md` — «`apps/api` ещё нет»; `docs/README.md` — «Код
  приложения на этом этапе не пишется»; `apps/api/README.md` — «Реализация слоёв
  [ ]». Фактически слои `domain/service/repository/handler` реализованы частично,
  миграция `00001_init.sql` есть.
- **Влияние:** нельзя опираться на README при выборе стратегии тестов.
- **Уточнение:** обновить статусы (не блокирует тесты, но вводит в заблуждение).

### D7. Дублирующаяся нумерация разделов в `15-consistency.md`

- **Что:** два раздела «## 12» (критерии приёмки и сверка с ответами заказчика);
  в таблицах §12 и в `00-open-questions.md` D-38 символ `|` внутри code-span ломает
  разметку строки.
- **Влияние:** только читаемость; на тесты не влияет.

## 2. Контракт против кода (C↔K)

Смонтировано 17 операций из 90 (см. `01-requirements-map.md` §5). Ниже — расхождения
по тем операциям, которые **есть**, и перечень отсутствующих.

### 2.1. Именование полей

| Объект           | Контракт                                                                                                                                                          | Реализация                                                                                                                                         |
| ---------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------- |
| `GuestProfile`   | `displayName`, `tenantCount`, `mergedIntoId`, `identifiers[].idType`, `identifiers[].valueNormalized`, `tenants[]` как объекты `GuestTenantSummary`, `segments[]` | `display_name`, `tenant_count`, `merged_into_id`, `identifiers[].id_type`, `identifiers[].value`, `tenants` как `[]string`, `segments` отсутствует |
| `Event`          | `guestId`, `tenantId`, `eventTypeId`, `eventTypeCode`, `tenantName`, `projectId`, `occurredAt`, `amountMinor`, `createdAt`                                        | `event_type_id`, `tenant_id`, `project_id`, `occurred_at`, `amount_minor`, `currency`; нет `guestId`, `tenantName`, `eventTypeCode`, `createdAt`   |
| `MergeCandidate` | `guestAId`, `guestBId`, `reasonCode`, `score`, `resolution`, `createdAt`                                                                                          | `guest_a_id`, `guest_b_id`, `match_type`, `reason_code`, `reason_text`, `score`, `resolution`; нет `createdAt`                                     |
| `Recommendation` | `id`, `guestId`, `tenantId`, `rank`, `score`, `explanation`, `computedAt`                                                                                         | `rank`, `tenant_id`, `score`, `explanation`, `algorithm_version`, `computed_at`; нет `id`, `guestId`                                               |
| `Segment`        | `isTemplate`, `currentVersionNo`, `status`, `memberCount`, `createdAt`, `updatedAt`                                                                               | `current_version_no`, `version_status`, `member_count`, `computed_at`, `rule`; нет `isTemplate`, `createdAt`, `updatedAt`                          |
| `MeResponse`     | `userId`, `email`, `displayName`, `role`, `tenantId`, `permissions`                                                                                               | `id`, `role`, `tenant_id`                                                                                                                          |
| `HealthResponse` | `status`, `version`, `time`, `checks`                                                                                                                             | `status`                                                                                                                                           |

**Влияние:** любой контрактный тест по JSON Schema падает на всех перечисленных
операциях. Это не «мелкие» расхождения: они делают невозможной проверку
«фактическое поведение соответствует OpenAPI» без решения, что считать истиной.

### 2.2. Пагинация и оболочка списков

- Контракт: `page` (с 1), `page_size` (1..200), ответ `{items, meta{page,pageSize,total,totalPages}}`.
- Реализация: `limit`, `offset`, ответ `{items}` без `meta`; `limit` ограничен 100.
- Затронуто: `GET /guests`, `GET /guests/{id}/events`, `GET /segments`,
  `GET /merge-candidates`.
- **Влияние:** проверки `meta`, границ `page_size` и `totalPages` невыполнимы.

### 2.3. Формат ошибок и коды

- Контракт (`common.json#/ErrorResponse`): `code` из фиксированного enum
  (`validation_error`, `invalid_credentials`, `forbidden_role`, `forbidden_tenant`,
  `not_found`, `conflict`, `already_resolved`, `merge_blocked`, `timeout`,
  `internal_error`, …), `details[]`, `request_id`.
- Реализация (`internal/platform/httpx/httpx.go`): `validation_failed` (нет в enum),
  `internal` (нет в enum), `unauthorized`, `not_found`, `conflict`, `forbidden_role`,
  `timeout`; `details` никогда не заполняется; `request_id` заполняется.
- **Влияние:** проверки конкретных кодов (`already_resolved`, `merge_blocked`,
  `segment_not_ready`, `contact_limit_exceeded`) невозможны — реализация отдаёт
  общий `conflict`. Проверка `details[].field` невозможна.

### 2.4. Сегменты: структура запроса и ответа

- Контракт `SegmentCreateRequest`: `{name, description?, templateCode?, rule: SegmentRule}`,
  где `SegmentRule` = `{was[], wasNot[], categoriesWas[], categoriesNot[], periodDays (required), requireContact?, minVisits?, minAmountMinor?}`.
- Реализация `service.SegmentCreateRequest`: `{name, description, template_code, was[], was_not[], period_days}` —
  поля `rule` нет вообще; `was`/`was_not` лежат в корне.
- **Следствие:** тело, валидное по контракту, даёт 400 (поле `was` пустое →
  `validateRule` → `ErrValidation`). Тело, валидное для реализации, не проходит
  схему контракта.
- `SegmentPreviewResponse`: контракт `{size, examples[{guestId, displayName, tenantCount, lastSeenAt}], cached, computedInMs}`;
  реализация `{size, examples: []string}`.
- `SegmentRule.periodDays` в контракте обязателен и ограничен 730; реализация
  подставляет 90 при `<= 0` и не ограничивает сверху.
- `requireContact`, `categoriesWas`, `categoriesWasNot`, `minVisits`, `minAmountMinor`
  реализацией игнорируются.
- `POST /segments` по контракту создаёт версию со статусом `pending` и ставит
  `segment.recompute` в очередь; реализация считает состав синхронно и сразу
  ставит версии `ready` (`segment_repo.go:129`).
- `GET /segments/{id}` по контракту отдаёт `ruleText` (правило на бизнес-языке);
  реализация отдаёт сырой JSON `rule`.

### 2.5. Рекомендации: шкала оценки и fallback

- Контракт: `Recommendation.score` — пример `0.78`, факторы `rawValue`/`weight`
  в диапазоне 0..1. Реализация (`internal/recommend/recommender.go`):
  `score = 0.2*priority + 40` → значения до 60.
- Контракт `RecommendationList` требует `guestId` и `computedAt` на верхнем уровне;
  реализация отдаёт только `items`.
- US-4 AC-4 / A-21 требуют fallback с пометкой «мало данных»; реализация выдаёт
  кандидатов по приоритету без пометки и без признака fallback.
- A-14 задаёт веса 0.4/0.25/0.1 и множитель приоритета; реализация использует
  другую формулу. Контракт числовых весов не фиксирует, поэтому это расхождение
  документации и кода, а не контракта.

### 2.6. Склейки: транзакционность, аудит, коды

- Контракт и US-11 AC-2: подтверждение — **одна транзакция** + запись в аудит.
  Реализация (`identity_service.go`) выполняет `MarkMerged` → `MoveToGuest(events)` →
  `MoveToGuest(identifiers)` → `UpdateDecision` **без транзакции**; аудит не пишется.
- Контракт: `POST /merge-candidates/{id}/confirm` при уже решённой паре → 409
  `already_resolved`; реализация → 409 `conflict`.
- Контракт: `GET /merge-candidates` сортировка по `createdAt` **ascending**;
  реализация — `ORDER BY created_at DESC`.
- Контракт: фильтр `reason_code`; реализация его игнорирует.
- Контракт: `GET /merge-candidates/{id}` возвращает `MergeCandidateDetail`
  (`guestA`, `guestB`, `sharedKeys`, `differingFields`); реализация возвращает
  плоский `MergeCandidateDTO` без профилей.
- Контракт: после confirm ставятся пересчёты сегментов и рекомендаций; реализация
  этого не делает.

### 2.7. Гости: поиск и фильтры

- Контракт: `q` обязателен, `minLength: 2`, иначе 400. Реализация: пустой `q`
  возвращает весь список (200).
- Контракт: `GET /guests` исключает `status=merged`. Реализация исключает только
  `deleted` — поглощённые профили попадают в выдачу.
- Контракт: `GET /guests/{id}/events` поддерживает `from`, `to`, `tenant_id`;
  реализация их игнорирует.
- Контракт: `DELETE /guests/{id}` (R-4) — не реализован.
- Контракт: `GET /guests/{id}/contacts` — не реализован.

### 2.8. Аутентификация

- `POST /auth/refresh` и `POST /auth/logout` не реализованы; `Set-Cookie` с refresh
  при логине не выставляется. Контракт требует httpOnly-cookie.
- `POST /auth/login` 429 (rate limit) не реализован; порог нигде не задан.
- `POST /auth/login` контракт: `password` `minLength: 8`; реализация длину не проверяет.
- `GET /auth/me` не отдаёт `email`, `displayName`, `permissions`.

### 2.9. Роли

- `auth.RequireRole` реализован, но **не подключён ни к одному маршруту**
  (`grep RequireRole` → только определение и README). Любой аутентифицированный
  пользователь получает доступ ко всем смонтированным операциям.
- `x-roles` в контракте содержит нечитаемые машиной значения `"R-1 (read-only)"`,
  `"R-3 (own)"`, `"R-4,R-1 (read-only)"` — автоматическая проверка прав по
  контракту невозможна без ручного маппинга.
- **Влияние:** все негативные проверки 403 (BR-22, BR-24, US-8 AC2/AC3, US-10 AC5,
  US-11 AC6) сейчас не проходят. Это дефект реализации, а не тестов.

### 2.10. Риск: завершающий слэш в маршрутах коллекций

`cmd/api/main.go` регистрирует коллекции как `g.Get("/", …)` внутри
`priv.Route("/guests", …)`, то есть фактический путь — `/api/v1/guests/`.
Контракт описывает `/api/v1/guests` без слэша. `chi` не включает
`RedirectSlashes` в этой сборке, поэтому запрос без слэша может дать 404.

Затронуто: `GET /guests`, `GET /segments`, `POST /segments`,
`GET /merge-candidates`.

**Статус:** требует проверки на запущенном сервисе (Go в среде не установлен,
см. G12). Если подтвердится — это блокирующий дефект: все коллекционные
эндпоинты недоступны по контрактному URL.

### 2.11. Отсутствующие операции (полный перечень)

Импорт: `POST/GET /imports`, `GET /imports/{id}`, `/rejected-rows`,
`POST /data-sources/{id}/sync`.
Справочники: все `/tenants*`, `/tenant-categories*`, `/projects*`, `/event-types*`,
`/data-source-types*`, `/data-sources*`.
Пользователи: `/users*`.
Выгрузки: `/exports*`, `/contact-limits*`.
Кампании: `/campaigns*`.
Аналитика: `/dashboard/*`, `/tenants/{id}/aggregates`, `/data-quality/*`.
Аудит: `/audit-logs`.
Сегменты: `GET /segment-templates`, `PUT/DELETE /segments/{id}`,
`/segments/{id}/members`, `/versions`, `/recompute`, `/segments/{id}/recommendations`.
Гости: `DELETE /guests/{id}`, `/guests/{id}/contacts`.
Идентичность: `POST /guests/{id}/merge`, `/unmerge-blocks*`.
Аутентификация: `/auth/refresh`, `/auth/logout`.

## 3. Контракт против схемы БД (C↔DDL)

| #   | Объект                                  | Контракт                                           | DDL                                                                | Влияние                                                                                         |
| --- | --------------------------------------- | -------------------------------------------------- | ------------------------------------------------------------------ | ----------------------------------------------------------------------------------------------- |
| 1   | `UserRole`                              | enum включает `system`                             | `app_user.role CHECK IN ('marketer','manager','tenant','analyst')` | пользователя с ролью `system` нельзя создать; тест «создать системного пользователя» невыполним |
| 2   | `ExportCreateRequest` `Idempotency-Key` | `maxLength: 128`                                   | `export_job.idempotency_key VARCHAR(64)`                           | ключ длиной 65..128 проходит валидацию контракта и падает на БД                                 |
| 3   | `EventType.code`                        | `maxLength: 64`                                    | `event_type.code VARCHAR(32)`                                      | код длиной 33..64 валиден по контракту, невалиден в БД                                          |
| 4   | `UnmergeBlock.reason`                   | `manual_unmerge`, `manual_reject`, `auto_conflict` | `manual_unmerge`, `rejected`                                       | см. D3                                                                                          |
| 5   | `UnmergeBlock.createdBy`                | nullable                                           | `NOT NULL`                                                         | контракт допускает отсутствие автора, БД — нет                                                  |
| 6   | `AuditLog.actorRole`                    | nullable                                           | `NOT NULL`                                                         | системные записи (R-6) требуют значения, контракт допускает `null`                              |
| 7   | `ContactLimit`                          | нет поля `isActive`                                | `contact_limit.is_active`                                          | контракт не позволяет деактивировать правило без удаления                                       |
| 8   | `SegmentStatus`                         | `pending/computing/ready/failed/archived`          | `segment.status` = `active/archived`                               | см. D2                                                                                          |
| 9   | `PageMeta.pageSize`                     | max 200                                            | реализация ограничивает 100                                        | граница 101..200 не проверяема                                                                  |
| 10  | `Recommendation.score`                  | пример `0.78`, факторы 0..1                        | `NUMERIC(7,4)`                                                     | численно совместимо, семантически — другая шкала (см. 2.5)                                      |

## 4. Пробелы (GAP)

### G1. Нет справочных данных

Миграция `00001_init.sql` создаёт 27 таблиц и **не содержит ни одного `INSERT`**.
Пусты: `tenant_category`, `tenant`, `event_type`, `data_source_type`, `data_source`,
`project`. Без них невозможно: создать событие (FK), посчитать сегмент, получить
рекомендацию, сформировать выгрузку.

### G2. Нет способа создать данные через API

Единственный пользователь создаётся сидом (`cmd/api/seed.go`) с ролью `analyst`
из `UCRP_ADMIN_PASSWORD`. Нет `POST /users`, `POST /tenants`, `POST /event-types`,
`POST /data-sources`, `POST /imports`. Значит, подготовка данных для E2E возможна
только прямыми `INSERT` в БД.

### G3. Воркер не обрабатывает задания

`cmd/worker/main.go`: `jobTypes := []string{}`, задание берётся и сразу помечается
`done`. Конвейер `import.process → identity.resolve → segment.recompute →
recommend.refresh` не работает. Асинхронные проверки (US-6, `computed_at`,
`status=computing`) невыполнимы.

### G4. Нет аудита

`AuditRepo` отсутствует в `internal/repository`; ни один сервис не пишет в
`audit_log`. BR-30 и US-11 AC-2 (запись в аудит) не проверяемы.

### G5. Нет автосклейки

`internal/identity` содержит только `doc.go` и `normalize/doc.go`. `identity.resolve`
не реализован, `merge_candidate` никто не создаёт. Значит, E2E-сценарий US-11
требует ручного создания кандидата в БД.

### G6. Нет импорта

`internal/ingestion` — только `doc.go`. `POST /imports` отсутствует. US-1, US-6,
UC-14 не проверяемы.

### G7. Нет выгрузки

`internal/export` — только `doc.go`. `/exports*`, `/contact-limits*` отсутствуют.
US-5, US-10, UC-20 не проверяемы.

### G8. Нет аналитики

`/dashboard/*`, `/tenants/{id}/aggregates`, `/data-quality/*` отсутствуют.
US-7, US-8, US-9 не проверяемы.

### G9. Нет rate limiting

429 на логине описан в контракте, но не реализован и порог не задан ни в одном
документе. Тест 429 невозможен.

### G10. Не заданы значения для проверки таймаутов

Контракт обещает 504 при 10 с (preview, refresh рекомендаций), 25 с (выгрузка),
30 с (запрос). Управляемого способа вызвать задержку нет; тесты 504 невозможны
без тестового «медленного» режима.

### G11. Нет тестовой инфраструктуры

В репозитории нет ни одного `*_test.go`, нет `playwright.config.*`, нет
`docker-compose` для тестового окружения (есть только `deploy/docker-compose.yml` и `deploy/docker-compose.dev.yml`
с `api` + `worker` + `db`). `bun run test` запускает только `vitest` в `apps/web`.

### G12. Go-тулчейн не установлен в среде

`go version` → команда не найдена; `docker` 29.9.2 и `bun` 1.4.4 доступны.
Запуск Go-тестов и `go run ./cmd/api` локально сейчас невозможен без установки Go
или использования Docker.

## 5. Влияние на тестирование (сводка)

| Категория      | Что можно тестировать сейчас                                                        | Что нельзя                                                        |
| -------------- | ----------------------------------------------------------------------------------- | ----------------------------------------------------------------- |
| Контракт       | `GET /healthz` (частично), `POST /auth/login` (частично), `GET /auth/me` (частично) | все остальные 87 операций; схемы ответов почти везде не совпадают |
| Бизнес-правила | BR-7, BR-8, BR-10 (частично), BR-19, BR-20, BR-21 (при ручной подготовке БД)        | BR-1…BR-4, BR-12…BR-17, BR-22…BR-30                               |
| E2E            | ни один сквозной сценарий US-1…US-11 целиком                                        | все                                                               |
| Роли           | ничего (RequireRole не подключён)                                                   | все проверки 403                                                  |
| Асинхронность  | ничего (воркер не работает)                                                         | US-6, `computed_at`, `status=computing`                           |

## 6. Вопросы, требующие решения до реализации тестов

1. **Что считать источником истины при расхождении контракта и кода** — OpenAPI
   (тогда тесты фиксируют дефекты реализации) или фактическое поведение (тогда
   тесты фиксируют дефекты контракта)? Предлагается: OpenAPI — истина, расхождения
   оформляются как ожидаемо падающие тесты с меткой `contract-drift`.
2. **D1** — конкуренты: понижение (A-18) или исключение (OpenAPI)?
3. **D2** — разделять ли `SegmentStatus` и `SegmentVersionStatus` в контракте?
4. **D3** — какой enum `UnmergeBlock.reason` считать верным?
5. **D5** — `camelCase` или `snake_case` в JSON?
6. **G2** — разрешено ли E2E-тестам готовить данные прямыми `INSERT` в БД
   (фикстуры), или требуется сначала реализовать справочные `POST`-эндпоинты?
7. **G3** — будет ли воркер обрабатывать задания к моменту запуска тестов?
8. **G11/G12** — какой стек и где размещать тесты (см. `04-e2e-plan.md` §5)?
