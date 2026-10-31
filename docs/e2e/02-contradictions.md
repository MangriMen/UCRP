# 02. Противоречия, расхождения и пробелы

Классификация: **D↔D** — документ против документа, **C↔K** — OpenAPI против кода,
**C↔DDL** — OpenAPI против схемы БД, **GAP** — пробел (нет данных/нет реализации).
Для каждого пункта: суть, источники, влияние на тестирование, нужное уточнение.

Сверка выполнена на коммите `945a5a7` (`Data generate (#39)`) по коду `apps/api`,
схеме `apps/api/migrations/common/00001_init.sql` и контракту `docs/openapi.json`

- `docs/openapi/**`; коллекционные эндпоинты дополнительно проверены запросами к
  запущенному сервису (§2.10).

## 0. Что изменилось с предыдущей редакции

| Было                                                   | Стало                                                                                    |
| ------------------------------------------------------ | ---------------------------------------------------------------------------------------- |
| Смонтировано 17 операций из 90                         | 88 из 90 (§2.11)                                                                         |
| `auth.RequireRole` не подключён, ролевых 403 нет       | подключён ко всем приватным маршрутам (§2.9)                                             |
| Нет импорта, выгрузок, справочников, аналитики, аудита | все группы смонтированы (§5 `01-requirements-map.md`)                                    |
| Воркер помечает задания `done` без обработки           | обрабатывает `import.process`, `connector.pull`, `identity.resolve`, `segment.recompute` |
| Автообъединения нет                                    | `internal/identity.Resolver` создаёт и подтверждает пары                                 |
| Аудит не пишется                                       | пишется в транзакциях объединения (`merge_tx.go`)                                        |
| Объединение без транзакции                             | одна транзакция + `audit_log` (BR-18 закрыт)                                             |
| Токен отдаётся как `access_token`                      | `accessToken`/`tokenType`/`expiresIn` (D5 закрыт)                                        |
| `GET /guests` возвращает `status=merged`               | фильтр `status = 'active'`                                                               |
| `q` не проверяется                                     | 400 при `q` короче 2 символов                                                            |
| Тело `POST /segments` по контракту → 400               | принимается `{name, rule}`; ломающее расхождение снято                                   |
| Коллекции доступны только со слэшем                    | доступны и `/guests`, и `/guests/` (проверено, §2.10)                                    |
| Контрактный тест падает на `validation_failed`         | код отдаёт `validation_error`/`internal_error`                                           |
| Нет E2E-инфраструктуры                                 | `apps/api/e2e` — 88 тестовых случаев (G11)                                               |

Открытыми остаются D1–D4, D6, D7, основная часть C↔K §2.1–§2.8 и C↔DDL §3.

## 1. Документ против документа (D↔D)

### D1. Конкуренты в рекомендациях: исключать или понижать — **открыто**

- **Что:** `docs/openapi/paths/recommendations.json` (GET `/guests/{id}/recommendations`)
  в `description` утверждает: «direct competitors (`tenant_category_relation.relationType = competitor`)
  are **excluded**». `docs/04-user-stories.md` US-4 AC-3, `docs/01-summary-assumptions.md`
  A-18, `docs/00-open-questions.md` D-34 и `docs/03-use-cases.md` UC-4 требуют
  обратного: конкурент **остаётся** в списке с понижающим коэффициентом.
- **Код:** `internal/recommend/recommender.go` не читает `tenant_category_relation`
  вообще; формула — `score = 0.2*priority` (+40, если похожие гости ходят в
  арендатора). То есть ни исключения, ни понижения нет — обе стороны неверны.
- **Дополнительно:** `RecommendationFactor.factorCode` (enum `similar_guests`,
  `category_adjacency`, `recency`, `distance`, `tenant_priority`) не содержит
  `competitor_penalty`, хотя глоссарий `docs/02-glossary.md` его определяет.
- **Влияние:** тест «конкурент присутствует в выдаче» и тест «конкурент отсутствует»
  взаимоисключающие. Пока сторона не выбрана, проверять нечего.
- **Уточнение:** подтвердить A-18/D-34 (понижение), исправить `description` в
  OpenAPI, добавить чтение `tenant_category_relation` в скоринг и фактор
  `competitor_penalty` в контракт.

### D2. Статус сегмента: `active/archived` или `pending/computing/ready/failed/archived` — **открыто**

- **Что:** `docs/openapi/components/schemas/segments.json` → `SegmentStatus` =
  `pending|computing|ready|failed|archived`, и `Segment.status` ссылается на этот
  enum. `docs/appendix/ddl.sql` / `apps/api/migrations/common/00001_init.sql`:
  `segment.status CHECK (status IN ('active','archived'))`, а
  `pending|computing|ready|failed` — это `segment_version.status`.
- **Код:** `internal/domain/segment.go` — `SegmentActive = "active"`,
  `SegmentArchived = "archived"`; версии — `pending|computing|ready|failed`.
  `GET /segments/{id}` отдаёт `"status": "active"`, что не входит в enum контракта.
  Сверх схемы DTO содержит `versionStatus` — контрактом `Segment` это поле не описано.
- **Влияние:** контрактный тест схемы `GET /segments/{id}` и `GET /segments`
  не может пройти ни при какой реализации: `active` не входит в enum контракта,
  `ready` не входит в CHECK БД. Это единственное расхождение, ломающее любую
  проверку схемы сегмента.
- **Уточнение:** разделить `SegmentStatus` (active/archived) и
  `SegmentVersionStatus` (pending/computing/ready/failed) в OpenAPI и добавить
  `versionStatus` в `Segment`.

### D3. Причина запрета объединения: `manual_reject`/`auto_conflict` или `rejected` — **открыто**

- **Что:** `docs/openapi/components/schemas/identity.json` → `MergeBlock.reason`
  enum `manual_unmerge|manual_reject|auto_conflict`. DDL:
  `CHECK (reason IN ('manual_unmerge','rejected'))`. Код
  (`internal/repository/merge_tx.go`) пишет `"rejected"` при reject и
  `"manual_unmerge"` при revert.
- **Влияние:** значение `manual_reject` из контракта невозможно сохранить (нарушит
  CHECK), `auto_conflict` не предусмотрен ни в DDL, ни в коде. Тест «создать
  запрет с причиной `manual_reject`» невыполним.
- **Уточнение:** привести enum контракта к DDL (`manual_unmerge`, `rejected`) либо
  расширить CHECK и реализовать `auto_conflict`.

### D4. Объёмы пилота: 80–150 арендаторов или 10–30

- **Что:** `docs/01-summary-assumptions.md` A-7: «80–150 арендаторов, ~40 000
  посетителей/мес». `docs/00-open-questions.md` Q-11: «10–30 арендаторов,
  6–12 месяцев истории, до 1–3 млн событий».
- **Влияние:** определяет объём генерации тестовых данных и необходимость
  нагрузочных проверок. На функциональные E2E не влияет, на performance — влияет.
- **Уточнение:** какое значение считать целевым для пилота.

### D5. Формат токена: `accessToken` или `access_token` — **закрыто**

- **Было:** OpenAPI — `accessToken`/`tokenType`/`expiresIn`, реализация —
  `access_token`/`expires_in`, клиент читал `data.access_token`.
- **Сейчас:** `internal/handler/auth_handler.go` отдаёт `accessToken`, `tokenType`,
  `expiresIn`, `user`; `MeResponse` — `userId`, `email`, `displayName`, `role`,
  `tenantId`, `permissions`. `apps/web/src/shared/api/client.ts` читает
  `data.accessToken`. Все три артефакта согласованы.
- **Осталось:** состав `permissions` для каждой роли нигде не зафиксирован;
  клиент на него не опирается.

### D6. Устаревшие утверждения о состоянии проекта

- **Что:** корневой `README.md` — «`apps/api` ещё нет»; `docs/README.md` — «Код
  приложения на этом этапе не пишется». Фактически реализованы все слои,
  88 операций и демо-генератор `cmd/seed`.
- **Влияние:** нельзя опираться на README при выборе стратегии тестов.
- **Уточнение:** обновить статусы (не блокирует тесты, но вводит в заблуждение).

### D7. Дублирующаяся нумерация разделов в `15-consistency.md`

- **Что:** два раздела «## 12» (критерии приёмки и сверка с ответами заказчика);
  в таблицах §12 и в `00-open-questions.md` D-38 символ `|` внутри code-span ломает
  разметку строки.
- **Влияние:** только читаемость; на тесты не влияет.

## 2. Контракт против кода (C↔K)

Смонтировано 88 операций из 90 (см. §2.11). Ниже — расхождения по тем операциям,
которые есть.

### 2.1. Именование полей

Контракт объявляет `camelCase` для тела и ответов (вводная часть `docs/openapi.json`)
и `snake_case` только для query-параметров.

| Объект              | Контракт                                                                                                                   | Реализация                                                                                                                                               | Соответствие |
| ------------------- | -------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------ |
| `MeResponse`        | `userId`, `email`, `displayName`, `role`, `tenantId`, `permissions`                                                        | совпадает                                                                                                                                                | ✓            |
| `Guest` (список)    | `id`, `displayName`, `status`, `tenantCount`, `createdAt`, `updatedAt`                                                     | совпадает, есть `meta`                                                                                                                                   | ✓            |
| `GuestProfile`      | `identifiers[].{id,guestId,idType,valueNormalized}`, `tenants[]` как `GuestTenantSummary`, `segments[]`                    | `identifiers[].{idType,value}` (нет `id`, `guestId`, `valueNormalized`), `tenants` как `[]string`, `segments` нет                                        | ✗            |
| `Event`             | `guestId`, `tenantId`, `eventTypeId`, `eventTypeCode`, `tenantName`, `projectId`, `occurredAt`, `amountMinor`, `createdAt` | `eventTypeId`, `tenantId`, `projectId`, `externalId`, `occurredAt`, `amountMinor`, `currency`; нет `guestId`, `tenantName`, `eventTypeCode`, `createdAt` | ✗            |
| `ContactLog`        | `id`, `guestId`, `exportId`, `channel`, `contactedAt`                                                                      | `id`, `export_id`, `contacted_at`; нет `guestId`, `channel`                                                                                              | ✗            |
| `MergeCandidate`    | `guestAId`, `guestBId`, `reasonCode`, `score`, `resolution`, `createdAt`                                                   | `guest_a_id`, `guest_b_id`, `match_type`, `reason_code`, `reason_text`, `score`, `resolution`; нет `createdAt`                                           | ✗            |
| `MergeBlock`        | `guestAId`, `guestBId`, `reason`, `createdAt`                                                                              | `guest_a_id`, `guest_b_id`, `reason`, `created_at`                                                                                                       | ✗            |
| `Recommendation`    | `id`, `guestId`, `tenantId`, `rank`, `score`, `explanation`, `computedAt`                                                  | `rank`, `tenant_id`, `score`, `explanation`, `algorithm_version`, `computed_at`; нет `id`, `guestId`                                                     | ✗            |
| `Segment`           | `isTemplate`, `currentVersionNo`, `status`, `memberCount`, `createdAt`, `updatedAt`, `rule`, `ruleText`                    | те же поля, но `status` = `active` и есть лишнее `versionStatus`; `ruleText` нет                                                                         | частично     |
| `SegmentVersion`    | `id`, `segmentId`, `versionNo`, `rule`, `status`, `memberCount`                                                            | `id`, `versionNo`, `rule`, `status`, `memberCount`, `computedAt`, `createdAt`; нет `segmentId`                                                           | частично     |
| `ImportBatch`       | `dataSourceId`, `totalRows`, `acceptedRows`, `rejectedRows`, `duplicateRows`, `createdAt`                                  | `data_source_id`, `total`, `accepted`, `rejected`, `duplicate`, `created_at`                                                                             | ✗            |
| `ImportAccepted`    | `batchId`, `status`, `duplicate`                                                                                           | `batch_id`, `status`, `duplicate`                                                                                                                        | ✗            |
| `RejectedRow`       | `id`, `importBatchId`, `rowNumber`, `errorCode`, `errorMessage`                                                            | `row_number`, `error_code`, `error_message`, `raw_line`; нет `id`, `importBatchId`                                                                       | ✗            |
| `Export`            | `segmentId`, `segmentVersionId`, `totalInSegment`, `exportedCount`, `createdAt`                                            | `segment_id`, `total_in_segment`, `exported_count`; нет `segmentVersionId`, `createdAt`                                                                  | ✗            |
| `ExportPreview`     | `inSegment`, `withContact`, `blockedByLimit`, `control`, `toSend`                                                          | `in_segment`, `with_contact`, `blocked_by_limit`, `control`, `to_send`                                                                                   | ✗            |
| `ExportMember`      | `guestId`, `contact`, `idType`, `isControl`                                                                                | `guest_id`, `contact`, `id_type`, `is_control`                                                                                                           | ✗            |
| `ContactLimit`      | `maxContacts`, `periodDays`                                                                                                | `max_contacts`, `period_days`, `is_active`                                                                                                               | ✗            |
| `Tenant`            | `categoryId`, `isActive`, `createdAt`, `updatedAt`                                                                         | `category_id`, `is_active`; нет `createdAt`, `updatedAt`                                                                                                 | ✗            |
| `TenantCategory`    | `isActive`                                                                                                                 | `is_active`                                                                                                                                              | ✗            |
| `EventType`         | `hasAmount`, `countsAsPresence`, `isActive`                                                                                | `has_amount`, `counts_as_presence`, `is_active`                                                                                                          | ✗            |
| `DataSource`        | `typeId`, `typeCode`, `fieldMapping`, `scheduleCron`, `isActive`, `lastImportAt`                                           | `type_id`, `field_mapping`, `schedule_cron`, `is_active`, `last_import_at`                                                                               | ✗            |
| `User`              | `displayName`, `isActive`, `lastLoginAt`                                                                                   | `display_name`, `is_active`; нет `lastLoginAt`                                                                                                           | ✗            |
| `AuditLog`          | `actorUserId`, `actorRole`, `entityType`, `entityId`, `requestId`, `createdAt`                                             | `actor_user_id`, `actor_role`, `entity_type`, `entity_id`, `created_at`; нет `requestId`, `ipAddress`                                                    | ✗            |
| `Campaign`          | `targetTenantId`, `segmentId`, `attributionDays`                                                                           | `tenant_id`, `segment_id`, `attribution_days`                                                                                                            | ✗            |
| `CrossTenantReport` | `from`, `to`, `granularity`, `distribution`, `crossTenantRate`, `series`                                                   | `one`, `two`, `three_plus`, `share`, `buckets`                                                                                                           | ✗            |
| `HealthResponse`    | `status`, `version`, `time`, `checks`                                                                                      | `status`, `version`, `time`; нет `checks`                                                                                                                | частично     |

**Влияние:** проверка схемы по JSON Schema проходит только для `/auth/login`,
`/auth/me`, `GET /guests`, `GET /data-source-types` и `/segments` (кроме `status`).
Остальные операции фиксируются как `contract-drift` и не могут считаться
соответствующими контракту. Это главный источник падений контрактных тестов.

### 2.2. Пагинация и оболочка списков

- Контракт: query `page` (с 1), `page_size` (1..200, default 50); ответ
  `{items, meta{page,pageSize,total,totalPages}}`. Для `tenant-categories`,
  `event-types`, `data-source-types`, `contact-limits`, `segment-templates`,
  `data-quality/sources` и `/segments/{id}/versions` контракт требует **массив
  без оболочки**.
- Реализация: `pageParams` читает `limit`/`offset`; `page`/`page_size`
  игнорируются. Верхняя граница `limit` — 100 (`GuestSQL.Search`,
  `SegmentSQL.List`) или 50 (`MergeBlockSQL.List`).
- Оболочка `{items, meta}` есть только у `GET /guests` и `GET /segments`.
  `GET /merge-candidates`, `/imports`, `/exports`, `/users`, `/audit-logs`,
  `/tenants`, `/projects`, `/data-sources`, `/guests/{id}/events`,
  `/guests/{id}/contacts`, `/segments/{id}/members`, `/merge-blocks` отдают
  `{items}` без `meta` (проверено запросами, §2.10).
- Обратное расхождение: `GET /tenant-categories`, `/event-types`,
  `/contact-limits`, `/segment-templates` по контракту — массив, реализация
  отдаёт `{items}`. `GET /data-source-types` отдаёт массив и контракту соответствует.
- `GET /imports/{id}/rejected-rows` отдаёт `{items, page, page_size, total}` —
  ни контрактный `PageMeta`, ни просто `items`.
- `GET /segments/{id}/versions` по контракту — массив `SegmentVersion`,
  реализация — `{items}`.
- **Влияние:** проверки `meta`, границ `page_size` и `totalPages` выполнимы только
  на `/guests` и `/segments`; параметры `page`/`page_size` из контракта не работают
  нигде.

### 2.3. Формат ошибок и коды

- Контракт (`common.json#/ErrorResponse`): `code` из фиксированного enum (29 значений),
  `details[]` (`field`, `reason`), `request_id`.
- Реализация (`internal/platform/httpx`): коды `validation_error`, `internal_error`,
  `unauthorized`, `not_found`, `conflict`, `forbidden_role`, `timeout`;
  `request_id` заполняется. **`details` не заполняется никогда** — структура
  `ErrorDetail` содержит поле, но ни один вызов `WriteError` его не передаёт.
- Вне enum контракта: `missing_columns` (422 при импорте, `import_service.go`).
- Бизнес-коды не различаются: повторное решение по паре и запрет объединения дают
  общий `conflict` вместо `already_resolved`/`merge_blocked`; агрегаты R-3 при
  чужом `tenant_id` отдают `forbidden_role` вместо `forbidden_tenant`.
- Реализованы и соответствуют контракту: `segment_not_ready`,
  `contact_limit_exceeded` (`ExportConflictError`).
- **Влияние:** проверки `details[].field` невыполнимы; проверки конкретных
  бизнес-кодов (`already_resolved`, `merge_blocked`, `forbidden_tenant`)
  фиксируют дефект реализации.

### 2.4. Сегменты: структура запроса и ответа

- **Создание.** Контракт `SegmentCreateRequest` = `{name, description?, templateCode?, rule: SegmentRuleNode}`.
  Реализация принимает и `rule` (`SegmentRuleBody`), и плоские `visitedTenantIds`/`excluded_tenant_ids` в
  корне; `rule` обязателен на уровне хендлера. Ломающее расхождение «тело по
  контракту → 400» снято.
- **Вложенные группы.** `SegmentRuleNode` в контракте — `oneOf [SegmentRule, SegmentRuleGroup]`,
  группы вкладываются произвольно. `domain.SegmentRule.UnmarshalJSON` понимает
  только плоскую форму (`visitedTenantIds`, `excludedTenantIds`, `periodDays`, `requireContact`);
  `operator`/`conditions` молча игнорируются. Веб-форма (`apps/web`,
  `build-segment/model/schema.ts`, `MAX_GROUP_DEPTH = 1`) отправляет именно такую
  структуру — группы из UI теряются на бэкенде.
- **Поля правила.** `visitedCategoryIds`, `excludedCategoryIds`, `minVisits`,
  `minAmountMinor` из контракта не поддерживаются: `domain.SegmentRule` их не
  содержит, `segments.SqlEvaluator` не читает. `periodDays` контрактом ограничен
  1..730, реализация подставляет 90 при `<= 0` и не ограничивает сверху.
- **Preview.** Контракт `{size, examples[{guestId, displayName?, tenantCount?, lastSeenAt?}], cached, computedInMs?}` —
  реализация отдаёт `{size, examples[{guestId}], cached:false}`. Схеме соответствует
  (остальные поля необязательны), но примеры бесполезны для UI. `limit` из
  контракта (1..100, default 20) игнорируется — примеров всегда ≤ 20.
- **Статус версии.** Контракт: `POST /segments` создаёт версию `pending` и ставит
  `segment.recompute`; реализация считает состав **синхронно** и сразу пишет
  `ready` (`SegmentService.Create`). Асинхронность наблюдаема только через
  `PUT /segments/{id}` (создаёт `pending` + очередь).
- **`ruleText`.** Контракт обещает правило на бизнес-языке в `Segment` и
  `SegmentVersion`; реализация отдаёт только сырой JSON `rule`.
- **Фильтры списка.** `GET /segments?status=&q=` контрактом поддержан;
  `SegmentSQL.List` фильтрует только `status <> 'archived'` и игнорирует оба
  параметра.
- **Members/Versions.** Контракт: `SegmentMemberPage` (`{items[{guestId, displayName, tenantCount, lastSeenAt}], meta}`)
  и массив `SegmentVersion`. Реализация: `{items: []string}` и
  `{items: []SegmentVersionDTO}` — members без профилей и без `meta`, versions в
  оболочке вместо массива.

### 2.5. Рекомендации: шкала оценки, форма ответа, fallback

- Контракт: `Recommendation.score` — пример `0.78`; `factors[]` с `rawValue`/`weight`
  в диапазоне 0..1. Реализация (`internal/recommend/recommender.go`):
  `score = 0.2*priority + 40` → значения до 60; факторы не отдаются. Формально
  `score` не ограничен `minimum`/`maximum`, поэтому схема проходит, но семантика
  шкалы другая.
- Контракт `RecommendationList` требует `guestId`; `GET /guests/{id}/recommendations`
  и `POST .../refresh` отдают только `{items}`.
- US-4 AC-4 / A-21 требуют fallback с пометкой «мало данных»; реализация выдаёт
  кандидатов по приоритету без пометки и без признака fallback (BR-11 не реализован).
- A-14 задаёт веса 0.4/0.25/0.1 и множитель приоритета; реализация использует
  другую формулу. Контракт числовых весов не фиксирует, поэтому это расхождение
  документации и кода, а не контракта.
- `GET /segments/{id}/recommendations` отдаёт `{items[{tenant_id, score, guests}]}`;
  такого агрегата в контракте нет — операция описана иначе.

### 2.6. Объединения: транзакционность, аудит, коды, сортировка

Реализовано и соответствует контракту:

- подтверждение выполняется одной транзакцией (`internal/repository/merge_tx.go`):
  `MarkMerged` → перенос событий → перенос идентификаторов →
  `UPDATE merge_candidate` → запись в `audit_log` (действия `merge.confirm`,
  `merge.reject`, `merge.revert`, `merge.manual`, `merge.auto`);
- запрет объединения создаётся при reject (`rejected`) и revert (`manual_unmerge`);
- `POST /guests/{id}/merge` реализован.

Расхождения:

- Контракт: при уже решённой паре и при паре под запретом — `already_resolved`
  и `merge_blocked`; реализация возвращает `domain.ErrConflict` → `conflict`.
- Контракт: `GET /merge-candidates` сортирует по `createdAt` **ascending**;
  реализация — `ORDER BY created_at DESC`.
- Контракт: фильтр `reason_code`; `MergeSQL.List` принимает только `resolution`,
  параметр `reason_code` игнорируется.
- Контракт: `GET /merge-candidates/{id}` возвращает `MergeCandidateDetail`
  (`guestA`, `guestB`, `sharedKeys`, `differingFields`); реализация возвращает
  плоский DTO без профилей.
- Контракт: после confirm ставятся пересчёты сегментов и рекомендаций; реализация
  этого не делает (пересчёт сегментов ставится воркером только после
  `identity.resolve`).

### 2.7. Гости: поиск, фильтры, контакты

Закрыто (проверено в коде и на запущенном сервисе):

- `q` обязателен, минимум 2 символа → 400;
- `GET /guests` и `GET /guests?q=` фильтруют `status = 'active'` — объединённые
  профили в выдачу не попадают;
- `DELETE /guests/{id}` реализован;
- `GET /guests/{id}/contacts` реализован.

Осталось:

- Контракт: `GET /guests/{id}/events` поддерживает `from`, `to`, `tenant_id`;
  реализация их игнорирует (читаются только `limit`/`offset`).
- `GET /guests/{id}/events` и `/contacts` не отдают `meta`, поля в `snake_case`.
- `limit` в `GET /guests/{id}/events` не нормализуется: при отсутствии параметра
  в SQL уходит `0` (в `Search` тот же параметр заменяется на 20).

### 2.8. Аутентификация

Закрыто:

- `POST /auth/refresh` и `POST /auth/logout` реализованы;
- при логине выставляется httpOnly cookie `refresh` (`Path=/api/v1/auth`,
  `SameSite=Strict`), при logout очищается;
- `GET /auth/me` отдаёт `email`, `displayName`, `permissions`.

Осталось:

- `POST /auth/login` 429 (rate limit) не реализован, порог нигде не задан (G9).
- Контракт: `password` `minLength: 8`; реализация длину не проверяет.
- Cookie выставляется с `Secure: false` — расхождение с контрактным описанием
  «httpOnly, Secure, SameSite=Strict».
- `POST /auth/logout` объявлен с `x-roles: any` и контрактом 401, но смонтирован
  **вне** группы `priv`: маршрут отвечает 204 без токена.

### 2.9. Роли

- `auth.RequireRole` подключён к каждому приватному маршруту, ролевая матрица
  соблюдается: `/users*` — только `manager`, `/merge-blocks*` и `POST /imports` —
  только `analyst`, `/exports` (кроме чтения) — только `marketer`,
  `/data-quality/sources` — `manager`/`analyst`. Негативные проверки 403 (BR-22,
  BR-24, US-8 AC2/AC3, US-10 AC5, US-11 AC6) стали выполнимыми.
- Два расхождения матрицы: `GET /tenants` и `GET /tenants/{id}` по контракту
  доступны `R-3 (own)`, но маршруты ограничены `marketer`/`manager`/`analyst` —
  арендатор получает 403 вместо собственной карточки; `PATCH /tenants/{id}`
  контрактом разрешён `R-1` (только поле `priority`), реализация требует `analyst`.
- `x-roles` в контракте содержит нечитаемые машиной значения `"R-1 (read-only)"`,
  `"R-3 (own)"`, `"R-4,R-1 (read-only)"`, `"R-1 (priority)"` — автоматическая
  проверка прав по контракту невозможна без ручного маппинга.

### 2.10. Завершающий слэш в маршрутах коллекций — **закрыто**

Коллекции зарегистрированы как `g.Get("/", …)` внутри `priv.Route("/guests", …)`.
Проверка на запущенном сервисе (`go run ./cmd/api`) с валидным токеном: маршруты
`/api/v1/guests`, `/api/v1/guests/`, `/api/v1/segments`, `/api/v1/segments/`,
`/api/v1/merge-candidates`, `/api/v1/merge-candidates/`, `/api/v1/imports`,
`/api/v1/imports/` отвечают 200. `chi` обслуживает поддерево и для `/guests`, и
для `/guests/*`, поэтому 404 не возникает. Дефект не подтвердился и из перечня
блокирующих исключён.

### 2.11. Отсутствующие операции

Из 90 операций контракта не смонтированы две:

```text
GET /api/v1/campaigns/{id}/results      (UC-20, x-roles R-1,R-2)
GET /api/v1/data-quality/sources/{id}   (UC-9,  x-roles R-4,R-2)
```

Все остальные группы смонтированы: аутентификация, пользователи, справочники и
источники, импорт, гости, идентичность, сегменты, выгрузки, лимиты контактов,
кампании (без `results`), аналитика (без детали источника), аудит.

## 3. Контракт против схемы БД (C↔DDL)

| #   | Объект                                  | Контракт                                           | DDL                                                                | Влияние                                                                                         |
| --- | --------------------------------------- | -------------------------------------------------- | ------------------------------------------------------------------ | ----------------------------------------------------------------------------------------------- |
| 1   | `UserRole`                              | enum включает `system`                             | `app_user.role CHECK IN ('marketer','manager','tenant','analyst')` | пользователя с ролью `system` нельзя создать; тест «создать системного пользователя» невыполним |
| 2   | `ExportCreateRequest` `Idempotency-Key` | `maxLength: 128`                                   | `export_job.idempotency_key VARCHAR(64)`                           | ключ длиной 65..128 проходит валидацию контракта и падает на БД                                 |
| 3   | `EventType.code`                        | `maxLength: 64`                                    | `event_type.code VARCHAR(32)`                                      | код длиной 33..64 валиден по контракту, невалиден в БД                                          |
| 4   | `MergeBlock.reason`                     | `manual_unmerge`, `manual_reject`, `auto_conflict` | `manual_unmerge`, `rejected`                                       | см. D3                                                                                          |
| 5   | `MergeBlock.createdBy`                  | nullable                                           | `NOT NULL`                                                         | контракт допускает отсутствие автора, БД — нет                                                  |
| 6   | `AuditLog.actorRole`                    | nullable                                           | `NOT NULL`                                                         | системные записи (R-6) требуют значения, контракт допускает `null`                              |
| 7   | `ContactLimit`                          | нет поля `isActive`                                | `contact_limit.is_active`                                          | контракт не описывает деактивацию правила, код отдаёт поле                                      |
| 8   | `ContactLimit`                          | `maxContacts ≤ 100`, `periodDays ≤ 365`            | только `>= 1` и `SMALLINT`                                         | значения 101+/366+ проходят реализацию и нарушают контракт                                      |
| 9   | `SegmentStatus`                         | `pending/computing/ready/failed/archived`          | `segment.status` = `active/archived`                               | см. D2                                                                                          |
| 10  | `PageMeta.pageSize`                     | max 200                                            | реализация ограничивает 100                                        | граница 101..200 не проверяема                                                                  |
| 11  | `Recommendation.score`                  | пример `0.78`, факторы 0..1                        | `NUMERIC(7,4)`                                                     | численно совместимо, семантически — другая шкала (см. 2.5)                                      |

## 4. Пробелы (GAP)

### G1. Справочные данные: частично закрыто

Миграция `00001_init.sql` по-прежнему **не содержит ни одного `INSERT`**. Стартовые
сиды выполняет код при запуске API:

- администратор — `UCRP_ADMIN_EMAIL` / `UCRP_ADMIN_PASSWORD`, роль `analyst`;
- типы событий — `ImportCatalogSQL.EnsureEventTypes`;
- типы источников — `DataSourceTypeSQL.EnsureDefaults` (`csv_excel`, `crm`, `pos`, `api`);
- 5 шаблонов сегментов — `SegmentSQL.EnsureTemplates`.

Пусты в чистой БД: `tenant_category`, `tenant`, `project`, `data_source`, а также
`guest`/`event`. Без них нечего считать; наполнение — через API или `cmd/seed`.

### G2. Подготовка данных через API: закрыто

`POST` доступны для `tenants`, `tenant-categories` (+`PUT .../relations`),
`projects`, `event-types`, `data-source-types`, `data-sources`, `users`,
`imports`, `contact-limits`, `campaigns`, `segments`, `exports`, а также
`POST /guests/{id}/merge`. E2E может готовить справочники и события через API
(импорт CSV), не прибегая к `INSERT`. Прямые `INSERT` остаются нужны только для
сущностей без `POST`: `guest`, `event`, `merge_candidate`, `merge_block`,
`segment_member`, `recommendation`, `contact_log`, `job_queue` — именно так
работает `apps/api/e2e/fixtures`.

### G3. Воркер: частично закрыто

`cmd/worker/main.go` обрабатывает `import.process`, `connector.pull`,
`identity.resolve`, `segment.recompute` и после `identity.resolve` ставит
`segment.recompute` для активных сегментов. Не реализовано:

- `recommend.refresh` — тип никем не ставится, рекомендации пересчитываются
  только синхронно через `POST /guests/{id}/recommendations/refresh`;
- `export.build` — выгрузка синхронная, `202` не выдаётся;
- планировщик (`platform/scheduler` пуст) — `connector.pull` только вручную.

Асинхронные проверки US-6 выполнимы: `PUT /segments/{id}` → версия `pending` →
воркер → `ready` и `computedAt`.

### G4. Аудит: частично закрыто

`audit_log` пишется только операциями объединения (`merge.confirm`, `merge.reject`,
`merge.revert`, `merge.manual`, `merge.auto`). `AuditService.Insert` не вызывается
ни из одного другого сервиса: создание сегмента, импорт, выгрузка, изменения
справочников и пользователей записей не оставляют. BR-30 проверяем только для
объединения.

### G5. Автообъединение: закрыто

`internal/identity/resolver.go` реализован: находит существующего гостя по
нормализованному идентификатору, привязывает событие, при совпадении телефона
делает автообъединение (`score 90`, `auto_merged`), при совпадении email ставит пару
в очередь (`pending`, score 70/75). Порог из NB-1 («≥ 80») в коде не используется —
вместо него фиксированные значения по типу ключа.

### G6. Импорт: закрыто

`internal/ingestion` содержит конвейер (`pipeline.go`), CSV-коннектор
(`connector_csv.go`), маппинг (`mapping.go`) и HTTP-коннектор для pull-источников.
`POST /imports` принимает multipart, считает sha256, при повторе отдаёт 200 с
`duplicate=true`, пишет `rejected_row` с кодами `invalid_date`, `no_external_id`,
`invalid_amount`, `unknown_event_type`, `invalid_phone`; `POST /data-sources/{id}/sync`
ставит `connector.pull`.

### G7. Выгрузка: закрыто

`/exports/preview`, `POST /exports`, `GET /exports`, `GET /exports/{id}`,
`/members`, `/download` и `/contact-limits*` реализованы: фильтр по каналу,
контрольная группа исключается из CSV и `contact_log`, идемпотентность по
`Idempotency-Key`, лимиты контактов, коды `segment_not_ready` и
`contact_limit_exceeded`. Выгрузка синхронная (202/504 недостижимы).

### G8. Аналитика: частично закрыто

Реализованы `/dashboard/cross-tenant`, `/dashboard/metrics`,
`/tenants/{id}/aggregates` (с изоляцией R-3 и порогом k = 5) и
`/data-quality/sources`. Отсутствуют `/campaigns/{id}/results` и
`/data-quality/sources/{id}`. Период дашборда не валидируется (BR-25), агрегаты
отдают `{newGuests, suppressed}` вместо `TenantAggregates`.

### G9. Нет rate limiting

429 на логине описан в контракте, но не реализован, порог не задан ни в одном
документе. Тест 429 невозможен.

### G10. Таймауты частично проверяемы

`POST /segments/preview` оборачивается в `context.WithTimeout(10s)` и отдаёт 504
`timeout`; глобальный `TimeoutMiddleware` — 30 с. Управляемого способа вызвать
задержку нет. Синхронной выгрузки с бюджетом 25 с нет — проверка 504 для неё
неприменима.

### G11. Тестовая инфраструктура: закрыто

`apps/api/e2e` — Go-пакет на стандартном `testing` без новых зависимостей:
`harness` (конфиг из env, логин, роли, БД), `fixtures` (подготовка и очистка
данных прямыми `INSERT` с уникальным `RunID`), `client` (HTTP + multipart),
`openapi` (валидация JSON Schema по `docs/openapi.json`), `report` (JUnit +
markdown-отчёт о расхождениях). 129 тестов — 130 функций `Test*` с `TestMain`.
`bun run test` запускает только `vitest` в `apps/web`.

Запуск:

```sh
./apps/api/e2e/run.ps1             # строгий режим (UCRP_E2E_STRICT_CONTRACT=1)
./apps/api/e2e/run.ps1 -ReportOnly # расхождения пишутся, но не валят тест
```

Обязательные env: `UCRP_E2E_BASE_URL`, `UCRP_E2E_DATABASE_URL`,
`UCRP_E2E_ADMIN_EMAIL`, `UCRP_E2E_ADMIN_PASSWORD`, `UCRP_E2E_OPENAPI`.
Без `UCRP_E2E_BASE_URL` набор целиком пропускается; `bun run test` E2E не
запускает — в `apps/api/package.json` нет `test`-скрипта.

Последний сохранённый прогон в `docs/e2e/artifacts/` — 88 тестов, 83 прошло,
1 упало, 4 пропущено, 71 расхождение контракта. Артефакты **предшествуют**
коммитам `663f87e` и `945a5a7`, поэтому их цифры устарели: часть зафиксированных
там расхождений уже закрыта (§0).

### G12. Go-тулчейн есть, но не в PATH

`go version` → `go1.27.3 windows/amd64` при запуске по полному пути
`E:\Programs\go\bin\go.exe`; в `PATH` каталога нет, поэтому `go test ./e2e/...`
нужно вызывать после добавления каталога в `PATH` либо по полному пути.
`docker` доступен, контейнер `ucrp-api-db-1` слушает порт 5543.

## 5. Влияние на тестирование (сводка)

| Категория      | Что можно тестировать сейчас                                                                              | Что нельзя                                                                                                                                                         |
| -------------- | --------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Контракт       | `POST /auth/login`, `GET /auth/me`, `GET /guests`, `GET /data-source-types`, `/segments` (кроме `status`) | схемы `GuestProfile`, `Event`, `MergeCandidate`, `Import*`, `Export*`, `ContactLimit`, `User`, `AuditLog`, `CrossTenantReport`; `meta` вне `/guests` и `/segments` |
| Бизнес-правила | BR-1…BR-10, BR-12…BR-14, BR-17, BR-18, BR-19, BR-26, BR-27, BR-28                                         | BR-11 (fallback), BR-15 (частично), BR-20/BR-21 (коды), BR-22/BR-23 (форма), BR-25, BR-29, BR-30 (кроме объединения)                                               |
| E2E            | импорт → гости → сегмент → рекомендации → выгрузка; объединение и разъединение; лимиты контактов          | стадии, зависящие от `recommend.refresh` и `export.build`; `campaigns/{id}/results`; `data-quality/sources/{id}`                                                   |
| Роли           | вся матрица 403, изоляция R-3 на `/tenants/{id}/aggregates`                                               | карточки арендатора для R-3 (`GET /tenants/{id}` возвращает 403) и `PATCH /tenants/{id}` для R-1                                                                   |
| Асинхронность  | `PUT /segments/{id}` → `pending` → `ready` + `computedAt`; `identity.resolve` после импорта               | `recommend.refresh`, `export.build`, проверки 504 по управляемой задержке                                                                                          |

## 6. Вопросы, требующие решения до реализации тестов

1. **Что считать источником истины при расхождении контракта и кода** — OpenAPI
   (тогда тесты фиксируют дефекты реализации) или фактическое поведение (тогда
   тесты фиксируют дефекты контракта)? Предлагается: OpenAPI — истина, расхождения
   оформляются как ожидаемо падающие тесты с меткой `contract-drift` (режим
   `-ReportOnly` уже это делает).
2. **D1** — конкуренты: понижение (A-18) или исключение (OpenAPI)? Сейчас в коде
   нет ни того, ни другого.
3. **D2** — разделять ли `SegmentStatus` и `SegmentVersionStatus` в контракте?
   Без этого ни один контрактный тест сегмента не проходит.
4. **D3** — какой enum `MergeBlock.reason` считать верным?
5. **`camelCase` или `snake_case`** — контракт объявляет `camelCase`, код отдаёт
   `snake_case` почти везде, кроме `/guests` и `/segments`. Массовая правка
   реализации — самый дорогой пункт списка; нужно решение, кто меняется.
6. **BR-29** — нужен ли `409 tenant_has_events` при удалении арендатора с
   событиями, или soft-delete без проверки считается достаточным?
7. **G3** — должны ли `recommend.refresh` и `export.build` появиться до прогона
   E2E, или асинхронные проверки фиксируются как ограничение?
8. **Данные для прогона** — какой демо-набор считать базовым: `cmd/seed`
   (детерминированные UUIDv5) или фикстуры `e2e/fixtures`? Сейчас наборы
   независимы, и отчёт должен явно указывать, на каком прогон шёл.
