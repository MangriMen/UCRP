# 04. План E2E-тестов

Состояние на коммите `945a5a7`: план **реализован** в `apps/api/e2e/` (Go
`testing`, 129 тестовых функций), 88 из 90 операций контракта смонтированы,
воркер обрабатывает `import.process`, `connector.pull`, `identity.resolve`,
`segment.recompute`. Этот документ описывает, что именно проверяется, чем
подтверждено и что осталось за пределами автоматизации. Матрицы покрытия и
трассируемости — `03-coverage-matrix.md` и `03-traceability-matrix.md`.

## 1. Принципы

1. **Источник истины — OpenAPI.** Расхождение контракта и кода фиксируется
   записью в отчёте о расхождениях (`report.Drifts`), а не подгонкой ожидаемого
   статуса под реализацию. В строгом режиме (`UCRP_E2E_STRICT_CONTRACT=1`, это
   значение по умолчанию в `run.ps1`) расхождение валит тест.
2. **Проверяется состояние, а не только код ответа.** После каждой мутации
   читается сущность через API или БД и сверяется инвариант: счётчики, статусы,
   `computed_at`, число записей `contact_log`.
3. **Никаких `sleep`.** Асинхронные состояния ожидаются опросом с дедлайном
   (`Suite.WaitFor`), интервал 200 мс, дедлайны 60 с для импорта и 30 с для
   остального.
4. **Независимость.** Каждый тест создаёт свои данные с префиксом
   `e2e-{runId}-`, где `runId` — случайные 8 hex-символов на прогон.
5. **Изоляция.** Тесты работают против отдельного экземпляра API и БД.
   Прямые `INSERT` разрешены только для сущностей без `POST` (см. §5.4).
6. **Негативные тесты — только подтверждённые контрактом или кодом.**
   Ожидаемый статус каждого негативного теста проверен по `docs/openapi.json`
   и по фактическому ответу живого API.
7. **Один дефект — один тест.** Тест, падающий из-за дефекта продукта,
   остаётся в наборе: он документирует дефект и станет зелёным после
   исправления.

## 2. Сквозные сценарии (E2E)

Все семь сценариев реализованы. Ссылка — на тестовую функцию, которая
их выполняет.

### E2E-1. Импорт → профиль → сегмент → рекомендация → выгрузка

Тест: `TestE2EImportToExport`. Трасса: US-1 → US-2 → US-3 → US-4 → US-5.

1. Фикстуры создают категорию, арендаторов A/B/C, тип события и источник.
2. `POST /imports` с CSV из одной строки (телефон, арендатор A) → 202 `queued`.
3. Воркер `import.process` → `identity.resolve` создаёт гостя; ожидание
   появления `guest_identifier` через `WaitFor`.
4. `GET /guests/{id}` → 200 и валидная схема; `GET /guests/{id}/events` → ровно
   одно событие, привязанное к разрешённому профилю.
5. `POST /segments` (реализация принимает `rule`, контракт — плоскую форму;
   `PostContractShape` фиксирует расхождение и повторяет запрос в форме кода).
6. `POST /guests/{id}/recommendations/refresh` → 200.
7. `POST /exports/preview` → `to_send ≥ 1`.
8. `POST /exports` с `idempotency_key` → 201; `GET /exports/{id}/download` → CSV
   содержит и `guest_id`, и телефон.

Идемпотентность выгрузки проверяется отдельно: `TestExportCreateIsIdempotent`.

### E2E-2. Новое событие меняет профиль и рекомендации

Тест: `TestE2ENewEventChangesSegmentAndRecommendations` (US-6).

1. Первый импорт создаёт гостя с событием в арендаторе A.
2. Второй импорт добавляет событие того же гостя в арендаторе B.
3. Ожидание: событие привязано к тому же `guest_id`, `tenantCount` вырос,
   лента событий содержит оба события.
4. `POST /guests/{id}/recommendations/refresh` → среди рекомендаций больше нет
   A и B (посещённые арендаторы исключены).

Пересчёт сегмента по событию (UC-13) в этом тесте не проверяется: воркер
`segment.recompute` работает, но событие не меняет сохранённый сегмент без
явного `PUT /segments/{id}` — см. `02-contradictions.md` G3.

### E2E-3. Объединение профилей: подтверждение и разъединение

Тест: `TestE2EIdentityMergeLifecycle` (US-11).

1. Фикстура создаёт двух гостей с общим телефоном и кандидата `pending`
   (`score = 55`).
2. `GET /merge-candidates?resolution=pending` → кандидат в очереди.
3. `GET /merge-candidates/{id}` → 200, валидная схема.
4. `POST .../confirm` → 200, `winnerGuestId` ≠ `mergedGuestId`.
5. `GET /guests/{loser}` → `status = merged`, `mergedIntoId = winner`.
6. `GET /guests/{winner}/events` → оба события; `tenantCount = 2`.
7. Повторный `confirm` → 409.
8. `POST .../revert` → 200; события вернулись поглощённому профилю.
9. В `merge_block` ровно одна запись на пару.

Отдельно: `TestMergeCandidateRejectCreatesBlock`, `TestMergeCandidateRevertPendingReturns409`,
`TestIdentityAutoMergeOnSharedPhone` (автообъединение при совпадении телефона),
`TestIdentityPendingCandidateOnSharedEmail` (спорная пара при совпадении email).

### E2E-4. Лимит контактов блокирует повторную выгрузку

Тесты: `TestContactLimitBlocksRepeatExport`, `TestContactLimitStrictestRuleWins`,
`TestContactLimitIsChannelScoped`, `TestContactLimitControlGroupNotLogged` (US-10).

1. Сегмент из трёх гостей с телефоном.
2. `POST /exports` (sms) → 201, три записи `contact_log`.
3. `POST /exports/preview` (sms) → `blocked_by_limit = 3`, `to_send = 0`.
4. Два правила (общее `maxContacts = 5` и sms `maxContacts = 1`) → применяется
   строгое: заблокированы все.
5. Лимит sms не влияет на push — проверяется через отдельный канал.
6. При `control_percent = 50` контрольная группа не попадает в `contact_log`.

### E2E-5. Изоляция арендатора

Тесты: `TestRoleTenantCannotReachPrivateAPI`, `TestRoleMatrix`,
`TestTenantAggregatesTenantRoleIsolation` (US-8).

1. Фикстура создаёт арендатора и пользователя с ролью `tenant`.
2. `GET /auth/me` → 200.
3. `GET /tenants/{own}/aggregates` → 200; `GET /tenants/{foreign}/aggregates` → 403.
4. Перебор `/guests`, `/segments`, `/merge-candidates`, `/imports`, `/exports`,
   `/data-sources` от имени tenant → 403 на каждом.
5. Матрица ролей: 21 подслучай по `marketer`/`manager`/`analyst` на реальных
   маршрутах (`TestRoleMatrix`).

Известное расхождение: `GET /tenants` и `GET /tenants/{id}` закрыты для R-3
полностью, хотя контракт разрешает чтение своего арендатора
(`02-contradictions.md` §2.9).

### E2E-6. Справочники и качество данных

Тесты: `TestDataQualitySourcesList`, `TestDataSourceTypesDefaults`,
`TestDataSourceCreate`, `TestDataSourceCreateRejectsBadMapping` (US-9).

1. `GET /data-source-types` → минимальный набор `csv_excel`, `crm`, `pos`, `api`.
2. `POST /data-sources` с маппингом → 201; без обязательных полей маппинга → 400.
3. `GET /data-quality/sources` → строка на каждый источник с полями
   `rejected_rate`, `pending_merges`, `warning`.
4. `GET /data-quality/sources/{id}` — операции нет в роутере, тест отсутствует.

### E2E-7. Кампания и эффект

Тесты: `TestCampaignCreateRequiresTenantAndSegment`,
`TestCampaignCreateDefaultsAttributionWindow`,
`TestCampaignDeleteWithExportsReturns409`, `TestCampaignDeleteWithoutExports`,
`TestCampaignCreateForbiddenForManagerAndAnalyst` (UC-20).

1. `POST /campaigns` с `tenant_id`/`segment_id` → 201, окно атрибуции по
   умолчанию 14 дней.
2. Выгрузка с `campaign_id` → 201.
3. `DELETE /campaigns/{id}` при наличии выгрузок → 409; без выгрузок → 204.
4. R-2 и R-4 → 403 на создание.

**Текущий дефект:** `GET /campaigns`, `POST /campaigns` и остальные операции
кампаний отвечают 500 — репозиторий обращается к несуществующей колонке
`campaign.tenant_id` (в DDL она называется `target_tenant_id`). Сценарии с
созданием кампании помечены `RequireStatusOrDrift` и пропускаются, фиксируя
расхождение. `GET /campaigns/{id}/results` не смонтирован, поэтому расчёт
конверсии и инкремента не проверяется.

## 3. Негативные сценарии (подтверждённые документацией)

| #    | Сценарий                                                                          | Ожидание                                        | Источник               |
| ---- | --------------------------------------------------------------------------------- | ----------------------------------------------- | ---------------------- |
| N-1  | `POST /auth/login` с неверным паролем                                             | 401 `invalid_credentials`                       | SD-1                   |
| N-2  | Приватный запрос без `Authorization`                                              | 401 `unauthorized`                              | реестр                 |
| N-3  | Битый/просроченный JWT                                                            | 401                                             | SD-1                   |
| N-4  | R-3 на `/guests`                                                                  | 403 `forbidden_role`                            | US-8 AC3               |
| N-5  | R-3 на чужой `tenant_id`                                                          | 403 `forbidden_tenant`                          | US-8 AC2               |
| N-6  | R-1 на `POST /imports`                                                            | 403                                             | US-1 AC7               |
| N-7  | R-2/R-3/R-4 на `POST /exports`                                                    | 403                                             | US-5 AC6               |
| N-8  | R-4 на изменение `/contact-limits`                                                | 403                                             | US-10 AC5              |
| N-9  | R-1 на `POST .../confirm`                                                         | 403                                             | US-11 AC6              |
| N-10 | `GET /guests?q=a`                                                                 | 400                                             | OpenAPI `minLength: 2` |
| N-11 | `GET /guests/{unknown}`                                                           | 404                                             | US-2 AC5               |
| N-12 | `POST /segments/preview` с арендатором в `visitedTenantIds` и `excludedTenantIds` | 400, `details[].field = rule.excludedTenantIds` | US-3 AC5               |
| N-13 | `POST /segments` без `name`                                                       | 400                                             | OpenAPI required       |
| N-14 | `POST /imports` файл > 100 000 строк                                              | 413                                             | US-1 AC7               |
| N-15 | `POST /imports` без обязательных колонок                                          | 422                                             | US-1 AC7               |
| N-16 | `POST /imports` несуществующий `data_source_id`                                   | 404                                             | SD-2                   |
| N-17 | `POST /exports` без `Idempotency-Key`                                             | 400                                             | US-5 AC6               |
| N-18 | `POST /exports` из сегмента в статусе `computing`                                 | 409 `segment_not_ready`                         | US-5 AC6               |
| N-19 | `POST /exports` когда лимит исключил всех                                         | 409 `contact_limit_exceeded`                    | US-10 AC5              |
| N-20 | `GET /exports/{id}/download` до готовности                                        | 409 `export_not_ready`                          | OpenAPI                |
| N-21 | `POST /contact-limits` `maxContacts = 0`                                          | 400                                             | US-10 AC5              |
| N-22 | `POST /contact-limits` `periodDays = 0`                                           | 400                                             | US-10 AC5              |
| N-23 | Повторный `confirm` по решённой паре                                              | 409 `already_resolved`                          | US-11 AC6              |
| N-24 | `revert` пары в статусе `pending`                                                 | 409                                             | OpenAPI                |
| N-25 | `POST /guests/{id}/merge` пара под запретом                                       | 409 `merge_blocked`                             | US-11 AC6              |
| N-26 | `POST /guests/{id}/merge` с собой                                                 | 400                                             | OpenAPI                |
| N-27 | `DELETE /tenants/{id}` с событиями                                                | 409 `tenant_has_events`                         | реестр                 |
| N-28 | `DELETE /segments/{id}` при активной кампании                                     | 409 `segment_in_active_campaign`                | OpenAPI                |
| N-29 | `DELETE /campaigns/{id}` с выгрузками                                             | 409 `campaign_has_exports`                      | реестр                 |
| N-30 | `DELETE /users/{id}` последнего R-2                                               | 409                                             | реестр                 |
| N-31 | `POST /users` роль `tenant` без `tenantId`                                        | 400                                             | UC-18, DDL CHECK       |
| N-32 | `POST /users` дубль email                                                         | 409                                             | реестр                 |
| N-33 | `GET /dashboard/cross-tenant` `from > to`                                         | 400                                             | US-7 AC4               |
| N-34 | `GET /dashboard/cross-tenant` период > 24 мес                                     | 400                                             | US-7 AC4               |
| N-35 | `POST /data-sources/{id}/sync` при идущей синхронизации                           | 409 `sync_in_progress`                          | SD-3                   |
| N-36 | `POST /segments/{id}/recompute` при идущем пересчёте                              | 409 `recompute_in_progress`                     | реестр                 |
| N-37 | `POST /segments/preview` дольше 10 с                                              | 504 `timeout`                                   | US-3 AC5               |
| N-38 | `POST /guests/{id}/recommendations/refresh` дольше 10 с                           | 504                                             | OpenAPI                |
| N-39 | `POST /auth/login` при переборе                                                   | 429 `rate_limited`                              | SD-1                   |

N-37, N-38, N-39 нереализуемы без тестового режима задержки и rate limiter
(см. `02-contradictions.md` G9, G10). Остальные проверки покрыты тестами;
где реализация отвечает другим кодом, расхождение попадает в
`UCRP_E2E_DRIFT_REPORT`, а не подменяется ожиданием.

Ожидаемые коды берутся из `docs/openapi.json` (`RequireDeclaredStatus`):
тест, получивший недекларированный статус, фиксирует расхождение.

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

### 5.1. Стек (реализован)

| Слой          | Выбор                                     | Состояние                                                      |
| ------------- | ----------------------------------------- | -------------------------------------------------------------- |
| Раннер        | Go `testing` в `apps/api/e2e/`            | работает, 129 тестовых функций                                 |
| HTTP-клиент   | `net/http` + `e2e/client`                 | `Get`, `Post`, `Put`, `Patch`, `Delete`, `PostMultipart`, `Me` |
| Проверка схем | `santhosh-tekuri/jsonschema/v6`           | `RequireSchema`, `RequireDeclaredStatus`, формат ошибки        |
| Фикстуры      | `database/sql` + `pgx` в `e2e/fixtures`   | 19 фабрик: категории, арендаторы, гости, события, сегменты     |
| Отчёт         | JUnit XML + markdown-отчёт о расхождениях | `UCRP_E2E_JUNIT`, `UCRP_E2E_DRIFT_REPORT`                      |
| Окружение     | живой API + `docker compose` (Postgres)   | `apps/api/docker/docker-compose.yml`                           |

Воркер обязателен: без `cmd/worker` сценарии с импортом ждут `WaitFor` 60 с на
каждый файл и падают по таймауту.

### 5.2. Структура пакета

```text
apps/api/e2e/
├── main_test.go               # JUnit и drift-отчёт после прогона
├── actors_test.go             # клиент под роль, jsonObject/mustJSON
├── testdata_test.go           # CSV-фикстуры
├── world_test.go              # категория + 3 арендатора + тип события + источник
├── contract_test.go           # контрактные проверки и unmounted operations
├── business_test.go           # профиль, сегменты, рекомендации, объединения
├── imports_test.go            # импорт, счётчики, rejected rows, идемпотентность
├── exports_test.go            # preview, create, download, журнал
├── contact_limits_test.go     # лимиты по каналам и контрольная группа
├── identity_resolve_test.go   # автообъединение и очередь кандидатов
├── data_sources_test.go       # типы источников и маппинг
├── roles_test.go              # матрица ролей и tenant-изоляция
├── users_test.go              # пользователи: CRUD, дубль email, последний R-2
├── references_test.go         # арендаторы, категории, проекты, типы событий
├── analytics_test.go          # cross-tenant, metrics, aggregates, data quality
├── segments_test.go           # версии, участники, архив, пересчёт, шаблоны
├── guests_test.go             # контакты, soft-delete, роли
├── campaigns_test.go          # кампании: создание, окно атрибуции, удаление
├── e2e_full_test.go           # E2E-1 и E2E-2
├── e2e_identity_test.go       # E2E-3 и сегмент → рекомендации
├── client/  fixtures/  harness/  openapi/  report/
└── run.ps1  run.sh            # запуск с обязательными env
```

### 5.3. Конфигурация запуска

```text
UCRP_E2E_BASE_URL          http://127.0.0.1:8080
UCRP_E2E_DATABASE_URL      postgres://ucrp:ucrp@127.0.0.1:5543/ucrp?sslmode=disable
UCRP_E2E_ADMIN_EMAIL       учётная запись администратора
UCRP_E2E_ADMIN_PASSWORD
UCRP_E2E_RUN_ID            генерируется, если не задан
UCRP_E2E_OPENAPI           путь к docs/openapi.json
UCRP_E2E_STRICT_CONTRACT   1 — расхождение валит тест, 0 — только отчёт
UCRP_E2E_JUNIT             путь к JUnit XML
UCRP_E2E_DRIFT_REPORT      путь к markdown-отчёту о расхождениях
```

Запуск:

```sh
docker compose -f apps/api/docker/docker-compose.yml up -d db
./apps/api/e2e/run.ps1              # строгий режим
./apps/api/e2e/run.ps1 -ReportOnly  # только отчёт о расхождениях
```

Логин администратора берётся из переменных, а не из константы: в демо-сиде
(`cmd/seed`) пароль `demo12345`, при сиде из окружения — значение
`UCRP_ADMIN_PASSWORD`. Без `UCRP_E2E_BASE_URL` набор целиком пропускается.

### 5.4. Подготовка и очистка данных

- **Подготовка:** `fixtures.Registry` вставляет данные прямыми `INSERT`
  с префиксом `e2e-{runId}-` в `code`/`display_name`/`external_id`.
  Реестр покрывает `tenant_category`, `tenant`, `event_type`,
  `data_source_type`, `data_source`, `guest`, `guest_identifier`, `event`,
  `merge_candidate`, `segment` + `segment_version` + `segment_member`,
  `app_user` и `contact_limit`.
- **Очистка:** `Registry.Cleanup` удаляет по `t.Cleanup` в порядке, обратном
  FK: `contact_log → export_member → export_job → campaign → contact_limit →
  segment_member → segment_version → segment → recommendation → merge_block →
  merge_candidate → event → guest_identifier → guest → rejected_row →
  import_batch → data_source → project → app_user → tenant →
  tenant_category_relation → tenant_category → event_type → data_source_type →
  job_queue`.
- **Изоляция:** уникальный `RunID` на прогон; тесты не зависят от порядка.

### 5.5. Ожидание асинхронных состояний

```text
Suite.WaitFor(deadline, what, fn) — опрос fn до true или дедлайна.
Интервал: 200 мс; ImportTimeout 60 с; остальные дедлайны 30 с.
Применяется к появлению guest_identifier после импорта, привязке события
и готовности версии сегмента.
```

## 6. Ограничения тестирования

| Ограничение                                      | Причина    | Что делаем                                                       |
| ------------------------------------------------ | ---------- | ---------------------------------------------------------------- |
| Нет REST для `guest`, `event`, `merge_candidate` | G2         | фикстуры прямыми `INSERT`                                        |
| `recommend.refresh` не ставится воркером         | G3         | пересчёт только через `POST .../refresh`                         |
| `export.build` не реализован                     | G3         | выгрузка синхронная, 202/504 недостижимы                         |
| Аудит пишется только для объединения             | G4         | `request_id` в `audit_log` остаётся `NULL`, REQ-19.2 не проверен |
| Нет rate limit                                   | G9         | N-39 не реализуем                                                |
| Нет управляемой задержки                         | G10        | N-37, N-38 не реализуемы                                         |
| `GET/POST /campaigns*` отвечают 500              | новый      | `RequireStatusOrDrift`, сценарий пропускается                    |
| `GET /dashboard/*` отвечает 500 на пустых данных | новый      | `RequireStatusOrDrift`, расхождение фиксируется                  |
| Дубль email в `POST /users` даёт 500             | новый      | `RequireStatusOrDrift`, расхождение фиксируется                  |
| `GET /data-quality/sources/{id}` не смонтирован  | `02` §2.11 | тест отсутствует                                                 |
| `GET /campaigns/{id}/results` не смонтирован     | `02` §2.11 | тест отсутствует                                                 |

## 7. Что осталось за пределами автоматизации

1. Расчёт конверсии и инкремента кампании (REQ-20.2) — операции нет.
2. Детали и динамика по источнику (REQ-9.3, REQ-9.4) — операции нет.
3. `requestId` в записи аудита (REQ-19.2) — колонка всегда `NULL`.
4. Rate limit логина (REQ-16.4) и таймауты 504 (REQ-3.4, REQ-4.7) — нет
   управляемого механизма.
5. Планировщик `connector.pull` — задание ставится только вручную.

## 8. Статус прогона

Последний полный прогон (живой API + воркер + Postgres): **129 тестов,
0 падений, 10 пропусков, 12 записей в отчёте о расхождениях**.

| Пропущенный тест                              | Причина                                        |
| --------------------------------------------- | ---------------------------------------------- |
| `TestCampaignCreateRequiresTenantAndSegment`  | `POST /campaigns` → 500 (колонка `tenant_id`)  |
| `TestCampaignCreateDefaultsAttributionWindow` | то же                                          |
| `TestCampaignDeleteWithExportsReturns409`     | то же                                          |
| `TestCampaignDeleteWithoutExports`            | то же                                          |
| `TestCrossTenantDistributionsContract`        | `share = NULL` на пустых данных → 500          |
| `TestDashboardMetricsFields`                  | `match_rate = NULL` на пустых данных → 500     |
| `TestUserCreateDuplicateEmail`                | 500 вместо 409                                 |
| `TestUserDeleteLastManagerReturns409`         | в БД больше одного активного manager           |
| `TestSegmentPreviewRejectsConflictingTenant`  | нет `details[].field = rule.excludedTenantIds` |
| `TestUnmountedOperationsAreReported`          | фиксирует два несмонтированных маршрута        |

## 9. Открытые вопросы к заказчику и команде

1. **Источник истины при расхождении контракта и кода** — OpenAPI или
   реализация? Действующее соглашение: OpenAPI; расхождение фиксируется
   отчётом и валит тест только в строгом режиме.
2. **D1** — конкурент в рекомендациях: понижение (A-18) или исключение (OpenAPI)?
3. **D2** — разделять ли `SegmentStatus` и `SegmentVersionStatus`?
4. **D3** — верный enum `MergeBlock.reason`?
5. **D5** — `camelCase` или `snake_case` в JSON-полях?
6. **Кампании** — исправить `campaign.tenant_id` → `target_tenant_id`
   в репозитории: сейчас все операции отдают 500.
7. **Аналитика** — `NULL` в `share`/`match_rate` на пустых данных: `COALESCE`
   в SQL или nullable-поле в DTO?
8. **Дубль email** — вернуть 409 вместо 500: уникальное ограничение не
   транслируется в `domain.ErrConflict`.
9. **CI** — добавлять ли job `e2e` в `.github/workflows/ci.yml` (сейчас его нет)?
10. **База для прогона** — допустим ли прогон против dev-базы или нужна
    отдельная БД?
