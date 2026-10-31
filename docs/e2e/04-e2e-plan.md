# 04. План E2E-тестов

## 1. Принципы

1. **Источник истины — OpenAPI.** Расхождение контракта и кода оформляется как
   падающий тест с меткой `contract-drift`, а не как «подгонка» теста под код.
2. **Проверяется состояние, а не только код ответа.** Каждый сценарий читает
   данные после операции и сверяет инварианты.
3. **Никаких `sleep`.** Ожидание асинхронных состояний — опрос с дедлайном
   (`waitFor(() => …, timeout)`), а не фиксированные паузы.
4. **Независимость.** Каждый тест создаёт свои данные с уникальным префиксом
   (`e2e-{runId}-{n}`) и не зависит от порядка запуска.
5. **Изоляция.** Тесты не пишут в dev-базу; отдельная БД на прогон.
6. **Негативные тесты — только подтверждённые документацией** (см. `03-coverage-matrix.md`).

## 2. Сквозные сценарии (E2E)

### E2E-1. Импорт → профиль → сегмент → рекомендация → выгрузка (ядро продукта)

Трасса: US-1 → US-2 → US-3 → US-4 → US-5, UC-1/2/3/4/5, SD-2/SD-4/SD-5.

1. R-4 создаёт категорию, 3 арендатора (A, B, C), тип события `purchase`
   (`countsAsPresence=true`), источник `csv_excel` с маппингом.
2. R-4 загружает CSV: гость `g1` — события в A и B; гость `g2` — только в A;
   гость `g3` — только в B; строка без `external_id`; строка с плохой датой.
3. Ожидание `GET /imports/{id}` → `completed_with_errors`; проверка
   `accepted + rejected + duplicate = total`, `rejected = 2`.
4. `GET /imports/{id}/rejected-rows` → 2 строки с кодами `no_external_id`, `bad_date`.
5. `GET /guests?q=<телефон g1>` → найден один гость; `GET /guests/{id}` →
   `tenantCount = 2`, `tenants` содержит A и B.
6. `GET /guests/{id}/events` → 2 события, отсортированы по `occurredAt` desc.
7. `POST /segments/preview` с `was=[A,B]`, `wasNot=[C]` → `size = 1` (только g1).
8. `POST /segments` → 201, версия 1; ожидание `status = ready` и `computedAt`.
9. `GET /segments/{id}/members` → ровно g1.
10. `GET /guests/{g1}/recommendations` → ≤ 3, среди них C, нет A и B.
11. `POST /exports/preview` (channel=sms) → `inSegment=1`, `withContact=1`.
12. `POST /exports` с `Idempotency-Key` → 201; `GET /exports/{id}` → инвариант
    счётчиков; `GET /exports/{id}/download` → CSV с одной строкой g1.
13. Повтор `POST /exports` с тем же ключом → тот же `id`, новых `contact_log` нет.

**Инварианты:** BR-1…BR-4, BR-7, BR-8, BR-10, BR-12, BR-14, BR-15.

### E2E-2. Новое событие меняет сегмент и рекомендации (US-6)

1. Исходное состояние из E2E-1 (или своё): g2 был только в A.
2. R-4 загружает файл с событием g2 в B.
3. Ожидание: `GET /guests/{g2}/events` содержит новое событие;
   `tenantCount` стал 2.
4. Ожидание: `GET /segments/{id}` → `computedAt` изменился, `memberCount` вырос.
5. `GET /guests/{g2}/recommendations` → набор изменился, B исчез.
6. `POST /segments/{id}/recompute` → 202; повторный вызов до завершения → 409.

**Инварианты:** BR-9, BR-15, US-6 AC1–AC4.

### E2E-3. Склейка профилей: подтверждение, разъединение, запрет (US-11)

1. Подготовка: два гостя с общим телефоном и разными именами → кандидат
   `pending` со `score < 80` (создаётся автосклейкой; при её отсутствии —
   фикстурой в БД, см. §6).
2. `GET /merge-candidates?resolution=pending` → кандидат в списке, есть `score`
   и `reasonCode`.
3. `GET /merge-candidates/{id}` → два профиля, `sharedKeys` содержит телефон.
4. `POST .../confirm` → 200; `GET /guests/{loser}` → `status=merged`,
   `mergedIntoId=winner`; события и идентификаторы перешли победителю.
5. Повторный `POST .../confirm` → 409 `already_resolved`.
6. `POST .../revert` → 200; события вернулись по `originGuestId`;
   `GET /unmerge-blocks` содержит пару.
7. `POST /guests/{loser}/merge` с той же парой → 409 `merge_blocked`.
8. `DELETE /unmerge-blocks/{id}` → 204; повторное объединение проходит.
9. `GET /audit-logs?entity_type=merge_candidate&entity_id={id}` → записи
   `merge.confirm`, `merge.revert`.

**Инварианты:** BR-18…BR-21, BR-30.

### E2E-4. Лимит контактов блокирует повторную выгрузку (US-10)

1. Сегмент с 3 гостями, у всех телефон.
2. `POST /contact-limits` `{maxContacts: 1, periodDays: 7}` → 201.
3. `POST /exports` (sms) → 201, 3 получателя, 3 записи `contact_log`.
4. `POST /exports/preview` (sms) → `blockedByLimit = 3`, `toSend = 0`.
5. `POST /exports` (sms) → 409 `contact_limit_exceeded`.
6. `POST /contact-limits` `{maxContacts: 0}` → 400.
7. `DELETE /contact-limits/{id}`; `POST /exports/preview` → `blockedByLimit = 0`.

**Инварианты:** BR-16, BR-17, BR-26.

### E2E-5. Изоляция арендатора (US-8)

1. R-2 создаёт пользователя `tenant` с `tenantId = T1`; R-2 создаёт `T2`.
2. Логин R-3 → `GET /auth/me` → `tenantId = T1`.
3. `GET /tenants/{T1}/aggregates` → 200, нет `guest_id` в ответе.
4. `GET /tenants/{T2}/aggregates` → 403 `forbidden_tenant`.
5. `GET /tenants` → только T1.
6. Перебор всех прочих эндпоинтов реестра → 403 (табличный тест).
7. Значение агрегата < 5 → `suppressed = true`, `value = null`.

**Инварианты:** BR-22…BR-24.

### E2E-6. Справочники и качество данных (US-9)

1. R-4 создаёт источник, загружает 2 файла: один с 20 % отклонённых строк.
2. `GET /data-quality/sources` → `rejectedShare30d ≈ 0.2`, `hasWarning = true`
   (порог 10 %).
3. Источник без загрузок → `lastSuccessfulImportAt = null`, не ошибка.
4. `GET /data-quality/sources/{id}` → динамика по загрузкам.

**Инварианты:** US-9 AC1–AC4, D-21.

### E2E-7. Кампания и эффект (UC-20)

1. Сегмент + кампания с `targetTenantId = C`, `attributionDays = 14`.
2. Выгрузка с `campaignId` и `controlPercent = 20`.
3. Событие-покупка у C для получателя в пределах окна.
4. `GET /campaigns/{id}/results` → `conversions ≥ 1`, `conversionRate > 0`,
   `controlGroup > 0`, инкремент вычислен.
5. Событие вне окна атрибуции не засчитывается.

**Инварианты:** A-6, A-19, D-30.

## 3. Негативные сценарии (подтверждённые документацией)

| #    | Сценарий                                                  | Ожидание                             | Источник               |
| ---- | --------------------------------------------------------- | ------------------------------------ | ---------------------- |
| N-1  | `POST /auth/login` с неверным паролем                     | 401 `invalid_credentials`            | SD-1                   |
| N-2  | Приватный запрос без `Authorization`                      | 401 `unauthorized`                   | реестр                 |
| N-3  | Битый/просроченный JWT                                    | 401                                  | SD-1                   |
| N-4  | R-3 на `/guests`                                          | 403 `forbidden_role`                 | US-8 AC3               |
| N-5  | R-3 на чужой `tenant_id`                                  | 403 `forbidden_tenant`               | US-8 AC2               |
| N-6  | R-1 на `POST /imports`                                    | 403                                  | US-1 AC7               |
| N-7  | R-2/R-3/R-4 на `POST /exports`                            | 403                                  | US-5 AC6               |
| N-8  | R-4 на изменение `/contact-limits`                        | 403                                  | US-10 AC5              |
| N-9  | R-1 на `POST .../confirm`                                 | 403                                  | US-11 AC6              |
| N-10 | `GET /guests?q=a`                                         | 400                                  | OpenAPI `minLength: 2` |
| N-11 | `GET /guests/{unknown}`                                   | 404                                  | US-2 AC5               |
| N-12 | `POST /segments/preview` с арендатором в `was` и `wasNot` | 400, `details[].field = rule.wasNot` | US-3 AC5               |
| N-13 | `POST /segments` без `name`                               | 400                                  | OpenAPI required       |
| N-14 | `POST /imports` файл > 100 000 строк                      | 413                                  | US-1 AC7               |
| N-15 | `POST /imports` без обязательных колонок                  | 422                                  | US-1 AC7               |
| N-16 | `POST /imports` несуществующий `data_source_id`           | 404                                  | SD-2                   |
| N-17 | `POST /exports` без `Idempotency-Key`                     | 400                                  | US-5 AC6               |
| N-18 | `POST /exports` из сегмента в статусе `computing`         | 409 `segment_not_ready`              | US-5 AC6               |
| N-19 | `POST /exports` когда лимит исключил всех                 | 409 `contact_limit_exceeded`         | US-10 AC5              |
| N-20 | `GET /exports/{id}/download` до готовности                | 409 `export_not_ready`               | OpenAPI                |
| N-21 | `POST /contact-limits` `maxContacts = 0`                  | 400                                  | US-10 AC5              |
| N-22 | `POST /contact-limits` `periodDays = 0`                   | 400                                  | US-10 AC5              |
| N-23 | Повторный `confirm` по решённой паре                      | 409 `already_resolved`               | US-11 AC6              |
| N-24 | `revert` пары в статусе `pending`                         | 409                                  | OpenAPI                |
| N-25 | `POST /guests/{id}/merge` пара под запретом               | 409 `merge_blocked`                  | US-11 AC6              |
| N-26 | `POST /guests/{id}/merge` с собой                         | 400                                  | OpenAPI                |
| N-27 | `DELETE /tenants/{id}` с событиями                        | 409 `tenant_has_events`              | реестр                 |
| N-28 | `DELETE /segments/{id}` при активной кампании             | 409 `segment_in_active_campaign`     | OpenAPI                |
| N-29 | `DELETE /campaigns/{id}` с выгрузками                     | 409 `campaign_has_exports`           | реестр                 |
| N-30 | `DELETE /users/{id}` последнего R-2                       | 409                                  | реестр                 |
| N-31 | `POST /users` роль `tenant` без `tenantId`                | 400                                  | UC-18, DDL CHECK       |
| N-32 | `POST /users` дубль email                                 | 409                                  | реестр                 |
| N-33 | `GET /dashboard/penetration` `from > to`                  | 400                                  | US-7 AC4               |
| N-34 | `GET /dashboard/penetration` период > 24 мес              | 400                                  | US-7 AC4               |
| N-35 | `POST /data-sources/{id}/sync` при идущей синхронизации   | 409 `sync_in_progress`               | SD-3                   |
| N-36 | `POST /segments/{id}/recompute` при идущем пересчёте      | 409 `recompute_in_progress`          | реестр                 |
| N-37 | `POST /segments/preview` дольше 10 с                      | 504 `timeout`                        | US-3 AC5               |
| N-38 | `POST /guests/{id}/recommendations/refresh` дольше 10 с   | 504                                  | OpenAPI                |
| N-39 | `POST /auth/login` при переборе                           | 429 `rate_limited`                   | SD-1                   |

N-37, N-38, N-39 нереализуемы без тестового режима задержки и rate limiter
(см. `02-contradictions.md` G9, G10).

## 4. Проверки бизнес-правил (структура)

Для каждого правила из `01-requirements-map.md` §1.3 фиксируется:

| Поле                  | Содержание                               |
| --------------------- | ---------------------------------------- |
| Начальное состояние   | какие сущности созданы и в каком статусе |
| Участвующие endpoints | последовательность вызовов               |
| Допустимое действие   | запрос, который должен пройти            |
| Ожидаемое состояние   | чтение после операции + инвариант        |
| Запрещённое действие  | запрос, который должен вернуть ошибку    |
| Ожидаемая ошибка      | код HTTP + `error.code` из контракта     |

Пример для BR-16 (лимит контактов):

```text
Начальное: сегмент S с 3 гостями, у всех phone; правило {maxContacts:1, periodDays:7}
Действие:  POST /exports {segmentId:S, channel:sms, format:csv} + Idempotency-Key
Ожидание:  201; GET /exports/{id} → exportedCount=3; contact_log = 3
Запрет:    POST /exports (тот же сегмент, новый ключ)
Ошибка:    409, error.code = contact_limit_exceeded
```

## 5. Стек и структура тестов

### 5.1. Предлагаемый стек

| Слой          | Выбор                                      | Обоснование                                                                                            |
| ------------- | ------------------------------------------ | ------------------------------------------------------------------------------------------------------ |
| Раннер        | **Go `testing`** в `apps/api/e2e/`         | тот же язык и тулчейн, что у API; `go test -race` уже в CI; не тянет новых зависимостей в JS-воркспейс |
| HTTP-клиент   | `net/http` + тонкая обёртка                | без внешних зависимостей                                                                               |
| Проверка схем | `github.com/santhosh-tekuri/jsonschema/v6` | валидирует ответы по `docs/openapi.json`                                                               |
| Фикстуры      | SQL через `database/sql` + `pgx`           | справочники нельзя создать через API (G2)                                                              |
| Окружение     | `docker compose` (Postgres + api + worker) | уже есть `deploy/docker-compose.yml` + `deploy/docker-compose.dev.yml`                                 |
| Отчёт         | `go test -json` → JUnit через `gotestsum`  | пригодно для CI                                                                                        |

Альтернатива (если нужен единый JS-стек): **Vitest + `openapi-fetch`** в
`apps/web/e2e/` — переиспользует существующий `vitest`, но требует Node-клиента
к БД для фикстур и не покрывает Go-часть.

### 5.2. Структура (вариант Go)

```text
apps/api/e2e/
├── main_test.go            # TestMain: поднять окружение, миграции, сид
├── client/
│   ├── client.go           # HTTP-клиент: базовый URL, токен, разбор ошибок
│   ├── auth.go             # login, me, refresh
│   └── schema.go           # валидация ответа по OpenAPI
├── fixtures/
│   ├── db.go               # подключение к тестовой БД
│   ├── reference.go        # категории, арендаторы, типы событий, источники
│   ├── guests.go           # гости, идентификаторы, события
│   └── cleanup.go          # удаление по префиксу runId
├── data/
│   ├── import-valid.csv
│   ├── import-rejected.csv
│   └── import-large.csv    # генерируется в тесте
├── auth_test.go
├── guests_test.go
├── segments_test.go
├── recommendations_test.go
├── identity_test.go
├── imports_test.go
├── exports_test.go
├── analytics_test.go
├── e2e_core_test.go        # E2E-1
├── e2e_event_test.go       # E2E-2
├── e2e_identity_test.go    # E2E-3
├── e2e_limits_test.go      # E2E-4
├── e2e_tenant_test.go      # E2E-5
└── e2e_campaign_test.go    # E2E-7
```

### 5.3. Конфигурация запуска

```text
UCRP_E2E_BASE_URL      http://127.0.0.1:8080
UCRP_E2E_DATABASE_URL  postgres://ucrp:ucrp@localhost:5543/ucrp_e2e?sslmode=disable
UCRP_E2E_ADMIN_EMAIL   e2e-admin@ucrp.local
UCRP_E2E_ADMIN_PASSWORD
UCRP_E2E_RUN_ID        генерируется, если не задан
UCRP_E2E_OPENAPI       путь к docs/openapi.json
```

Запуск:

```sh
docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.dev.yml up -d db
cd apps/api && go test ./e2e/... -race -count=1
```

### 5.4. Подготовка и очистка данных

- **Подготовка:** `fixtures` вставляет справочники и гостей прямыми `INSERT`
  с префиксом `e2e-{runId}-` в `code`/`display_name`/`external_id`.
- **Очистка:** `TestMain` после прогона удаляет строки по префиксу в порядке,
  обратном FK: `contact_log → export_member → export_job → segment_member →
  segment_version → segment → recommendation → merge_candidate → unmerge_block →
  event → guest_identifier → guest → import_batch → rejected_row → data_source →
  project → tenant → tenant_category_relation → tenant_category → app_user`.
- **Изоляция:** отдельная БД `ucrp_e2e`, пересоздаётся перед прогоном.

### 5.5. Ожидание асинхронных состояний

```text
waitFor(deadline, interval, fn) — опрос fn до true или дедлайна.
Используется для: import status, segment status/computedAt, export status.
Интервалы: 200 мс; дедлайны: импорт 60 с, сегмент 30 с, выгрузка 30 с.
```

## 6. Ограничения тестирования

| Ограничение                                   | Причина    | Что делаем                                                         |
| --------------------------------------------- | ---------- | ------------------------------------------------------------------ |
| Нет `POST` для справочников                   | G2         | фикстуры прямыми `INSERT`                                          |
| Нет автосклейки                               | G5         | кандидат создаётся фикстурой                                       |
| Воркер не обрабатывает задания                | G3         | асинхронные проверки помечены `blocked`                            |
| Нет аудита                                    | G4         | проверки аудита помечены `blocked`                                 |
| Нет rate limit                                | G9         | N-39 помечен `blocked`                                             |
| Нет управляемой задержки                      | G10        | N-37, N-38 помечены `blocked`                                      |
| Go не установлен в среде                      | G12        | запуск через Docker                                                |
| 83 из 120 операций не смонтированы            | §5 `01`    | тесты написаны, но помечены `blocked` до реализации                |
| Коллекции смонтированы со слэшем (`/guests/`) | `02` §2.10 | первый прогон подтверждает или опровергает 404 по контрактному URL |

## 7. Порядок реализации (после согласования)

1. Каркас: `TestMain`, клиент, валидатор схем, фикстуры, очистка.
2. Контрактные тесты по 17 смонтированным операциям (фиксируют расхождения
   из `02-contradictions.md` §2).
3. Бизнес-тесты на смонтированных операциях: сегменты, рекомендации, склейки,
   профиль.
4. E2E-3 (склейки) и E2E-5 (изоляция) — единственные сквозные, реализуемые
   без импорта и выгрузки.
5. Остальные E2E — после реализации соответствующих эндпоинтов и воркера.

## 8. Открытые вопросы к заказчику и команде

1. **Источник истины при расхождении контракта и кода** — OpenAPI или реализация?
   (влияет на 83 теста)
2. **D1** — конкурент в рекомендациях: понижение (A-18) или исключение (OpenAPI)?
3. **D2** — разделять ли `SegmentStatus` и `SegmentVersionStatus`?
4. **D3** — верный enum `UnmergeBlock.reason`?
5. **D5** — `camelCase` или `snake_case` в JSON-полях?
6. **G2** — допустима ли подготовка данных прямыми `INSERT` в тестовую БД?
7. **G3** — когда воркер начнёт обрабатывать `import.process` и остальные задания?
8. **Стек** — Go `testing` в `apps/api/e2e/` или Vitest в `apps/web/e2e/`?
9. **Окружение** — поднимать ли отдельный `docker-compose.e2e.yml` в репозитории?
10. **CI** — добавлять ли job `e2e` в `.github/workflows/ci.yml` (сейчас его нет)?
11. **Объёмы** — D4: 80–150 или 10–30 арендаторов для генератора данных?
12. **Приоритет** — писать тесты на 17 существующих операций сейчас или ждать
    реализации остальных 73?
