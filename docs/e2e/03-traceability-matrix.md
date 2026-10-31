# 03. Матрица трассируемости требований (Этап 3)

Основа сопоставления: `docs/14-traceability.md` (US/UC → AD → SD → ER → CRUD → CMP → API),
уточнённая по `docs/openapi.json`, `docs/endpoint-registry.md`, `docs/04-user-stories.md`,
`docs/03-use-cases.md`, `docs/07-crud-matrix.md`.

Детализация по отдельным тестам (120 проверок) — в `03-coverage-matrix.md`.

## Легенда

**Тип проверки**

| Код | Значение                                                       |
| --- | -------------------------------------------------------------- |
| P   | позитивный сценарий                                            |
| N   | негативный сценарий                                            |
| C   | контрактный (метод, URL, параметры, тело, статус, JSON Schema) |
| B   | проверка бизнес-правила или ограничения                        |
| E   | сквозной E2E-сценарий                                          |
| S   | безопасность и разграничение ролей                             |

**Колонки-определения**

| Колонка                         | Значения                                                                                                                                                |
| ------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **EP** — существует ли endpoint | `✓` — есть в контракте и смонтирован в `cmd/api/main.go`; `контракт` — есть в OpenAPI, не смонтирован; `—` — endpoint отсутствует                       |
| **T** — существует ли тест      | `код` — тест реализован в `apps/api/e2e/`; `план` — тест спроектирован в `04-e2e-plan.md`, код не написан; `—` — тест невозможен (причина в примечании) |
| **P**                           | покрыт позитивный сценарий                                                                                                                              |
| **N**                           | покрыт негативный сценарий                                                                                                                              |
| **B**                           | покрыто бизнес-ограничение                                                                                                                              |
| **E2E**                         | требование входит в сквозной сценарий (указан номер)                                                                                                    |

**Важно.** Колонка **T** отражает состояние на момент Этапа 4: `код` — тест
реализован в `apps/api/e2e/` (44 теста), `план` — спроектирован, но не написан,
`—` — невозможен. Реализованные тесты **не запущены**: в среде разработки нет
Go и не запущен демон Docker (`05-implementation.md` §8).

## 1. US-1 / UC-1. Загрузка данных арендатора

| ID       | Требование                                       | UC    | US   | Endpoint                                     | Тест                                   | Тип | EP       | T    | P   | N   | B   | E2E   |
| -------- | ------------------------------------------------ | ----- | ---- | -------------------------------------------- | -------------------------------------- | --- | -------- | ---- | --- | --- | --- | ----- |
| REQ-1.1  | Файл принят, ответ 202 `queued`                  | UC-1  | US-1 | `POST /imports`                              | `imports.upload.accepted`              | P,C | контракт | план | ✓   | —   | —   | E2E-1 |
| REQ-1.2  | Статус и счётчики загрузки                       | UC-1  | US-1 | `GET /imports/{id}`                          | `imports.status.counters`              | P,C | контракт | план | ✓   | —   | —   | E2E-1 |
| REQ-1.3  | `accepted + rejected + duplicate = total`        | UC-1  | US-1 | `GET /imports/{id}`                          | `imports.counters.invariant`           | B   | контракт | план | ✓   | —   | ✓   | E2E-1 |
| REQ-1.4  | Повтор файла → 200 `duplicate=true`, без дублей  | UC-1  | US-1 | `POST /imports`                              | `imports.upload.idempotent`            | B   | контракт | план | ✓   | —   | ✓   | E2E-1 |
| REQ-1.5  | Дубль `external_id` в другом файле → `duplicate` | UC-1  | US-1 | `POST /imports`                              | `imports.upload.duplicate-external-id` | B   | контракт | план | ✓   | —   | ✓   | —     |
| REQ-1.6  | Строка без `external_id` → отклонена             | UC-1  | US-1 | `GET /imports/{id}/rejected-rows`            | `imports.rejected.no-external-id`      | N,B | контракт | план | —   | ✓   | ✓   | E2E-1 |
| REQ-1.7  | Неверная дата → отклонена                        | UC-1  | US-1 | `GET /imports/{id}/rejected-rows`            | `imports.rejected.bad-date`            | N,B | контракт | план | —   | ✓   | ✓   | E2E-1 |
| REQ-1.8  | CSV отклонённых строк                            | UC-1  | US-1 | `GET /imports/{id}/rejected-rows?format=csv` | `imports.rejected.csv`                 | P,C | контракт | план | ✓   | —   | —   | —     |
| REQ-1.9  | Файл > 100 000 строк → 413                       | UC-1  | US-1 | `POST /imports`                              | `imports.upload.too-large`             | N,B | контракт | план | —   | ✓   | ✓   | —     |
| REQ-1.10 | Неверный формат → 422                            | UC-1  | US-1 | `POST /imports`                              | `imports.upload.bad-format`            | N   | контракт | план | —   | ✓   | —   | —     |
| REQ-1.11 | Несуществующий источник → 404                    | UC-1  | US-1 | `POST /imports`                              | `imports.upload.unknown-source`        | N   | контракт | план | —   | ✓   | —   | —     |
| REQ-1.12 | R-1 и R-3 → 403                                  | UC-1  | US-1 | `POST /imports`                              | `imports.upload.forbidden`             | N,S | контракт | план | —   | ✓   | —   | —     |
| REQ-1.13 | Список загрузок с фильтрами                      | UC-1  | US-1 | `GET /imports`                               | `imports.list.filters`                 | P,C | контракт | план | ✓   | —   | —   | —     |
| REQ-1.14 | Ручной pull → 202                                | UC-14 | US-1 | `POST /data-sources/{id}/sync`               | `sync.manual.accepted`                 | P   | контракт | план | ✓   | —   | —   | —     |
| REQ-1.15 | Повторный pull → 409 `sync_in_progress`          | UC-14 | —    | `POST /data-sources/{id}/sync`               | `sync.manual.in-progress`              | N,B | контракт | план | —   | ✓   | ✓   | —     |

## 2. US-2 / UC-2. Единый профиль гостя

| ID       | Требование                                | UC   | US   | Endpoint                    | Тест                             | Тип | EP       | T    | P   | N   | B   | E2E   |
| -------- | ----------------------------------------- | ---- | ---- | --------------------------- | -------------------------------- | --- | -------- | ---- | --- | --- | --- | ----- |
| REQ-2.1  | Поиск по телефону в разных форматах       | UC-2 | US-2 | `GET /guests?q=`            | `guests.search.phone-normalized` | P,B | ✓        | код  | ✓   | —   | ✓   | E2E-1 |
| REQ-2.2  | `q` короче 2 символов → 400               | UC-2 | US-2 | `GET /guests?q=`            | `guests.search.q-too-short`      | N,C | ✓        | код  | —   | ✓   | —   | —     |
| REQ-2.3  | Поглощённый профиль отдаёт `mergedIntoId` | UC-2 | US-2 | `GET /guests/{id}`          | `guests.get.merged-into`         | P,B | ✓        | код  | ✓   | —   | ✓   | E2E-3 |
| REQ-2.4  | Несуществующий гость → 404                | UC-2 | US-2 | `GET /guests/{id}`          | `guests.get.not-found`           | N   | ✓        | код  | —   | ✓   | —   | —     |
| REQ-2.5  | Лента событий по убыванию даты            | UC-2 | US-2 | `GET /guests/{id}/events`   | `guests.events.sorted-desc`      | P,B | ✓        | код  | ✓   | —   | ✓   | E2E-1 |
| REQ-2.6  | Фильтры `from`, `to`, `tenant_id`         | UC-2 | US-2 | `GET /guests/{id}/events`   | `guests.events.filters`          | P   | ✓        | план | ✓   | —   | —   | —     |
| REQ-2.7  | `tenantCount` и список арендаторов        | UC-2 | US-2 | `GET /guests/{id}`          | `guests.get.tenant-count`        | P,B | ✓        | код  | ✓   | —   | ✓   | E2E-1 |
| REQ-2.8  | История контактов гостя                   | UC-2 | US-2 | `GET /guests/{id}/contacts` | `guests.contacts.list`           | P,C | контракт | план | ✓   | —   | —   | E2E-4 |
| REQ-2.9  | Soft-delete гостя                         | UC-2 | US-2 | `DELETE /guests/{id}`       | `guests.delete.soft`             | P,B | контракт | план | ✓   | —   | ✓   | —     |
| REQ-2.10 | R-3 → 403                                 | UC-2 | US-2 | `/guests*`                  | `guests.forbidden-tenant-role`   | N,S | ✓        | план | —   | ✓   | —   | —     |

## 3. US-3 / UC-3. Сегмент без кода

| ID       | Требование                                 | UC   | US   | Endpoint                      | Тест                                  | Тип | EP       | T    | P   | N   | B   | E2E   |
| -------- | ------------------------------------------ | ---- | ---- | ----------------------------- | ------------------------------------- | --- | -------- | ---- | --- | --- | --- | ----- |
| REQ-3.1  | Preview: размер и примеры, без записи в БД | UC-3 | US-3 | `POST /segments/preview`      | `segments.preview.no-write`           | P,B | ✓        | код  | ✓   | —   | ✓   | E2E-1 |
| REQ-3.2  | Арендатор в `was` и `wasNot` → 400         | UC-3 | US-3 | `POST /segments/preview`      | `segments.preview.conflicting-tenant` | N,B | ✓        | код  | —   | ✓   | ✓   | —     |
| REQ-3.3  | Пустой `was` → 400                         | UC-3 | US-3 | `POST /segments/preview`      | `segments.preview.empty-was`          | N   | ✓        | план | —   | ✓   | —   | —     |
| REQ-3.4  | Preview дольше 10 с → 504                  | UC-3 | US-3 | `POST /segments/preview`      | `segments.preview.timeout`            | N   | ✓        | —    | —   | —   | —   | —     |
| REQ-3.5  | Создание сегмента, версия 1                | UC-3 | US-3 | `POST /segments`              | `segments.create.version-1`           | P,C | ✓        | код  | ✓   | —   | —   | E2E-1 |
| REQ-3.6  | Создание без `name` → 400                  | UC-3 | US-3 | `POST /segments`              | `segments.create.no-name`             | N   | ✓        | код  | —   | ✓   | —   | —     |
| REQ-3.7  | Правка правила создаёт версию 2            | UC-3 | US-3 | `PUT /segments/{id}`          | `segments.update.new-version`         | P,B | контракт | план | ✓   | —   | ✓   | —     |
| REQ-3.8  | История версий доступна                    | UC-3 | US-3 | `GET /segments/{id}/versions` | `segments.versions.list`              | P,C | контракт | план | ✓   | —   | —   | —     |
| REQ-3.9  | Постраничный состав сегмента               | UC-3 | US-3 | `GET /segments/{id}/members`  | `segments.members.paged`              | P,C | контракт | план | ✓   | —   | —   | E2E-1 |
| REQ-3.10 | Пять предустановленных шаблонов            | UC-3 | US-3 | `GET /segment-templates`      | `segments.templates.five`             | P,C | контракт | план | ✓   | —   | —   | —     |
| REQ-3.11 | Архивирование сегмента                     | UC-3 | US-3 | `DELETE /segments/{id}`       | `segments.delete.archive`             | P   | контракт | план | ✓   | —   | —   | —     |
| REQ-3.12 | Архив при активной кампании → 409          | UC-3 | —    | `DELETE /segments/{id}`       | `segments.delete.active-campaign`     | N,B | контракт | план | —   | ✓   | ✓   | —     |
| REQ-3.13 | Список сегментов с размером и `computedAt` | UC-3 | US-3 | `GET /segments`               | `segments.list.size`                  | P,C | ✓        | код  | ✓   | —   | —   | —     |
| REQ-3.14 | R-3 → 403                                  | UC-3 | US-3 | `/segments*`                  | `segments.forbidden-tenant-role`      | N,S | ✓        | план | —   | ✓   | —   | —     |

## 4. US-4 / UC-4. Рекомендация гостю

| ID      | Требование                                    | UC    | US   | Endpoint                                    | Тест                             | Тип | EP       | T    | P   | N   | B   | E2E   |
| ------- | --------------------------------------------- | ----- | ---- | ------------------------------------------- | -------------------------------- | --- | -------- | ---- | --- | --- | --- | ----- |
| REQ-4.1 | Не более трёх рекомендаций с обоснованием     | UC-4  | US-4 | `GET /guests/{id}/recommendations`          | `reco.list.max-three`            | P,B | ✓        | код  | ✓   | —   | ✓   | E2E-1 |
| REQ-4.2 | Посещённые арендаторы отсутствуют             | UC-4  | US-4 | `GET /guests/{id}/recommendations`          | `reco.list.excludes-visited`     | B   | ✓        | код  | ✓   | —   | ✓   | E2E-1 |
| REQ-4.3 | Конкурент остаётся с понижающим коэффициентом | UC-4  | US-4 | `GET /guests/{id}/recommendations`          | `reco.list.competitor-penalized` | B   | ✓        | —    | —   | —   | —   | —     |
| REQ-4.4 | Fallback для гостя без истории                | UC-4  | US-4 | `GET /guests/{id}/recommendations`          | `reco.list.fallback-no-history`  | B   | ✓        | план | ✓   | —   | ✓   | —     |
| REQ-4.5 | Пустой список — валидное состояние            | UC-4  | US-4 | `GET /guests/{id}/recommendations`          | `reco.list.empty-ok`             | P   | ✓        | план | ✓   | —   | —   | —     |
| REQ-4.6 | Refresh заменяет набор рекомендаций           | UC-15 | US-4 | `POST /guests/{id}/recommendations/refresh` | `reco.refresh.replaces-set`      | P,B | ✓        | код  | ✓   | —   | ✓   | E2E-2 |
| REQ-4.7 | Refresh дольше 10 с → 504                     | UC-15 | —    | `POST /guests/{id}/recommendations/refresh` | `reco.refresh.timeout`           | N   | ✓        | —    | —   | —   | —   | —     |
| REQ-4.8 | Агрегат рекомендаций по сегменту              | UC-4  | US-4 | `GET /segments/{id}/recommendations`        | `reco.segment.aggregate`         | P,C | контракт | план | ✓   | —   | —   | —     |
| REQ-4.9 | R-3 → 403                                     | UC-4  | US-4 | `/guests/{id}/recommendations`              | `reco.forbidden-tenant-role`     | N,S | ✓        | план | —   | ✓   | —   | —     |

## 5. US-5 / UC-5. Выгрузка сегмента в канал

| ID       | Требование                                   | UC   | US   | Endpoint                     | Тест                                | Тип | EP       | T    | P   | N   | B   | E2E   |
| -------- | -------------------------------------------- | ---- | ---- | ---------------------------- | ----------------------------------- | --- | -------- | ---- | --- | --- | --- | ----- |
| REQ-5.1  | Preview: четыре счётчика                     | UC-5 | US-5 | `POST /exports/preview`      | `exports.preview.counters`          | P,C | контракт | план | ✓   | —   | —   | E2E-1 |
| REQ-5.2  | Фильтр по контакту канала                    | UC-5 | US-5 | `POST /exports/preview`      | `exports.preview.channel-contact`   | B   | контракт | план | ✓   | —   | ✓   | E2E-1 |
| REQ-5.3  | Создание выгрузки → 201                      | UC-5 | US-5 | `POST /exports`              | `exports.create.ready`              | P,C | контракт | план | ✓   | —   | —   | E2E-1 |
| REQ-5.4  | Без `Idempotency-Key` → 400                  | UC-5 | US-5 | `POST /exports`              | `exports.create.no-idempotency-key` | N   | контракт | план | —   | ✓   | —   | —     |
| REQ-5.5  | Повтор ключа → та же выгрузка                | UC-5 | US-5 | `POST /exports`              | `exports.create.idempotent`         | B   | контракт | план | ✓   | —   | ✓   | E2E-1 |
| REQ-5.6  | Сегмент не `ready` → 409 `segment_not_ready` | UC-5 | US-5 | `POST /exports`              | `exports.create.segment-not-ready`  | N,B | контракт | план | —   | ✓   | ✓   | —     |
| REQ-5.7  | Контрольная группа не попадает в CSV         | UC-5 | US-5 | `GET /exports/{id}/download` | `exports.download.excludes-control` | B   | контракт | план | ✓   | —   | ✓   | E2E-7 |
| REQ-5.8  | Фильтр `is_control` в members                | UC-5 | US-5 | `GET /exports/{id}/members`  | `exports.members.control-filter`    | P   | контракт | план | ✓   | —   | —   | E2E-7 |
| REQ-5.9  | Инвариант счётчиков выгрузки                 | UC-5 | US-5 | `GET /exports/{id}`          | `exports.counters.invariant`        | B   | контракт | план | ✓   | —   | ✓   | E2E-1 |
| REQ-5.10 | Скачивание до готовности → 409               | UC-5 | —    | `GET /exports/{id}/download` | `exports.download.not-ready`        | N   | контракт | план | —   | ✓   | —   | —     |
| REQ-5.11 | Журнал выгрузок                              | UC-5 | US-5 | `GET /exports`               | `exports.list.journal`              | P,C | контракт | план | ✓   | —   | —   | —     |
| REQ-5.12 | R-2, R-3, R-4 → 403 на создание              | UC-5 | US-5 | `POST /exports`              | `exports.create.forbidden`          | N,S | контракт | план | —   | ✓   | —   | —     |

## 6. US-6 / UC-6. Обновление при новом событии

| ID      | Требование                                       | UC    | US   | Endpoint                        | Тест                             | Тип | EP       | T    | P   | N   | B   | E2E   |
| ------- | ------------------------------------------------ | ----- | ---- | ------------------------------- | -------------------------------- | --- | -------- | ---- | --- | --- | --- | ----- |
| REQ-6.1 | Новое событие видно в ленте профиля              | UC-6  | US-6 | `GET /guests/{id}/events`       | `event.feed.after-import`        | E,B | контракт | план | ✓   | —   | ✓   | E2E-2 |
| REQ-6.2 | `tenantCount` обновлён                           | UC-6  | US-6 | `GET /guests/{id}`              | `event.tenant-count.updated`     | E,B | ✓        | план | ✓   | —   | ✓   | E2E-2 |
| REQ-6.3 | Состав и размер сегмента изменились              | UC-13 | US-6 | `GET /segments/{id}`            | `event.segment.recomputed`       | E,B | ✓        | план | ✓   | —   | ✓   | E2E-2 |
| REQ-6.4 | `computedAt` обновлён                            | UC-13 | US-6 | `GET /segments/{id}`            | `event.computed-at.changed`      | E,B | ✓        | план | ✓   | —   | ✓   | E2E-2 |
| REQ-6.5 | Принудительный пересчёт → 202                    | UC-13 | US-6 | `POST /segments/{id}/recompute` | `segments.recompute.accepted`    | P   | контракт | план | ✓   | —   | —   | E2E-2 |
| REQ-6.6 | Повторный пересчёт → 409 `recompute_in_progress` | UC-13 | —    | `POST /segments/{id}/recompute` | `segments.recompute.in-progress` | N,B | контракт | план | —   | ✓   | ✓   | —     |

Требование US-6 AC-5 «выгрузка из сегмента в статусе `computing` блокируется» покрыто
REQ-5.6 — отдельный тест не создаётся, чтобы не дублировать проверку.

## 7. US-7 / UC-7. Сводка управляющего

| ID      | Требование                                | UC   | US   | Endpoint                     | Тест                                 | Тип | EP       | T    | P   | N   | B   | E2E |
| ------- | ----------------------------------------- | ---- | ---- | ---------------------------- | ------------------------------------ | --- | -------- | ---- | --- | --- | --- | --- |
| REQ-7.1 | Распределение гостей 1 / 2 / 3+           | UC-7 | US-7 | `GET /dashboard/penetration` | `dashboard.penetration.buckets`      | P,C | контракт | план | ✓   | —   | —   | —   |
| REQ-7.2 | Динамика по гранулярности                 | UC-7 | US-7 | `GET /dashboard/penetration` | `dashboard.penetration.series`       | P   | контракт | план | ✓   | —   | —   | —   |
| REQ-7.3 | `from > to` → 400                         | UC-7 | US-7 | `GET /dashboard/penetration` | `dashboard.penetration.bad-period`   | N,B | контракт | план | —   | ✓   | ✓   | —   |
| REQ-7.4 | Период больше 24 месяцев → 400            | UC-7 | US-7 | `GET /dashboard/penetration` | `dashboard.penetration.period-limit` | N,B | контракт | план | —   | ✓   | ✓   | —   |
| REQ-7.5 | Метрики: match rate, конверсия, инкремент | UC-7 | US-7 | `GET /dashboard/metrics`     | `dashboard.metrics.fields`           | P,C | контракт | план | ✓   | —   | —   | —   |
| REQ-7.6 | R-3 → 403                                 | UC-7 | US-7 | `/dashboard/*`               | `dashboard.forbidden-tenant-role`    | N,S | контракт | план | —   | ✓   | —   | —   |

## 8. US-8 / UC-8. Агрегаты арендатора

| ID      | Требование                               | UC   | US   | Endpoint                            | Тест                         | Тип | EP       | T    | P   | N   | B   | E2E   |
| ------- | ---------------------------------------- | ---- | ---- | ----------------------------------- | ---------------------------- | --- | -------- | ---- | --- | --- | --- | ----- |
| REQ-8.1 | Агрегаты своего арендатора               | UC-8 | US-8 | `GET /tenants/{id}/aggregates`      | `aggregates.r3.own`          | P,C | контракт | план | ✓   | —   | —   | E2E-5 |
| REQ-8.2 | Чужой арендатор → 403 `forbidden_tenant` | UC-8 | US-8 | `GET /tenants/{id}/aggregates`      | `aggregates.r3.foreign`      | N,S | контракт | план | —   | ✓   | —   | E2E-5 |
| REQ-8.3 | Значения меньше k = 5 скрыты             | UC-8 | US-8 | `GET /tenants/{id}/aggregates`      | `aggregates.k-threshold`     | B   | контракт | план | ✓   | —   | ✓   | E2E-5 |
| REQ-8.4 | В ответе нет `guest_id`                  | UC-8 | US-8 | `GET /tenants/{id}/aggregates`      | `aggregates.no-guest-ids`    | S   | контракт | план | ✓   | —   | —   | E2E-5 |
| REQ-8.5 | R-2 читает агрегаты любого арендатора    | UC-8 | US-8 | `GET /tenants/{id}/aggregates`      | `aggregates.r2.any`          | P   | контракт | план | ✓   | —   | —   | —     |
| REQ-8.6 | R-3 видит только свою запись арендатора  | UC-8 | US-8 | `GET /tenants`, `GET /tenants/{id}` | `tenants.r3.own-only`        | N,S | контракт | план | —   | ✓   | —   | E2E-5 |
| REQ-8.7 | R-3 на любой другой эндпоинт → 403       | UC-8 | US-8 | все прочие из реестра               | `r3.endpoint-matrix`         | N,S | ✓        | план | —   | ✓   | —   | E2E-5 |
| REQ-8.8 | Токен R-3 без `tenant_id` → 403          | UC-8 | US-8 | `GET /tenants/{id}/aggregates`      | `aggregates.r3.no-claim`     | N,S | контракт | план | —   | ✓   | —   | —     |
| REQ-8.9 | Деактивированный арендатор → 404         | UC-8 | US-8 | `GET /tenants/{id}/aggregates`      | `aggregates.inactive-tenant` | N   | контракт | план | —   | ✓   | —   | —     |

## 9. US-9 / UC-9. Качество данных по источникам

| ID      | Требование                                         | UC   | US   | Endpoint                         | Тест                 | Тип | EP       | T    | P   | N   | B   | E2E   |
| ------- | -------------------------------------------------- | ---- | ---- | -------------------------------- | -------------------- | --- | -------- | ---- | --- | --- | --- | ----- |
| REQ-9.1 | Качество по каждому источнику                      | UC-9 | US-9 | `GET /data-quality/sources`      | `dq.sources.list`    | P,C | контракт | план | ✓   | —   | —   | E2E-6 |
| REQ-9.2 | Порог 10 % отклонённых или 7 дней → предупреждение | UC-9 | US-9 | `GET /data-quality/sources`      | `dq.sources.warning` | B   | контракт | план | ✓   | —   | ✓   | E2E-6 |
| REQ-9.3 | Источник без загрузок — не ошибка                  | UC-9 | US-9 | `GET /data-quality/sources/{id}` | `dq.source.no-data`  | P   | контракт | план | ✓   | —   | —   | —     |
| REQ-9.4 | Детали и динамика по источнику                     | UC-9 | US-9 | `GET /data-quality/sources/{id}` | `dq.source.detail`   | P,C | контракт | план | ✓   | —   | —   | —     |
| REQ-9.5 | R-1 и R-3 → 403                                    | UC-9 | US-9 | `/data-quality/*`                | `dq.forbidden-roles` | N,S | контракт | план | —   | ✓   | —   | —     |

## 10. US-10 / UC-10. Лимит контактов

| ID        | Требование                                 | UC    | US    | Endpoint                                                    | Тест                           | Тип | EP       | T    | P   | N   | B   | E2E   |
| --------- | ------------------------------------------ | ----- | ----- | ----------------------------------------------------------- | ------------------------------ | --- | -------- | ---- | --- | --- | --- | ----- |
| REQ-10.1  | Правило действует на следующие выгрузки    | UC-10 | US-10 | `POST /contact-limits`                                      | `limits.create.applies`        | P,B | контракт | план | ✓   | —   | ✓   | E2E-4 |
| REQ-10.2  | Гость исключён при превышении лимита       | UC-10 | US-10 | `POST /exports/preview`                                     | `limits.excluded-count`        | B   | контракт | план | ✓   | —   | ✓   | E2E-4 |
| REQ-10.3  | Применяется самое строгое правило          | UC-10 | US-10 | `POST /exports/preview`                                     | `limits.strictest-wins`        | B   | контракт | план | ✓   | —   | ✓   | —     |
| REQ-10.4  | Контрольная группа не учитывается в лимите | UC-10 | US-10 | `POST /exports`                                             | `limits.control-not-counted`   | B   | контракт | план | ✓   | —   | ✓   | E2E-7 |
| REQ-10.5  | Лимит исключил всех → 409                  | UC-10 | US-10 | `POST /exports`                                             | `exports.create.all-blocked`   | N,B | контракт | план | —   | ✓   | ✓   | E2E-4 |
| REQ-10.6  | Без правил ограничение не применяется      | UC-10 | US-10 | `POST /exports/preview`                                     | `limits.none-configured`       | B   | контракт | план | ✓   | —   | ✓   | E2E-4 |
| REQ-10.7  | `maxContacts < 1` → 400                    | UC-10 | US-10 | `POST /contact-limits`                                      | `limits.create.invalid-max`    | N,B | контракт | план | —   | ✓   | ✓   | —     |
| REQ-10.8  | `periodDays < 1` → 400                     | UC-10 | US-10 | `POST /contact-limits`                                      | `limits.create.invalid-period` | N,B | контракт | план | —   | ✓   | ✓   | —     |
| REQ-10.9  | R-4 читает, но не изменяет правила         | UC-10 | US-10 | `/contact-limits*`                                          | `limits.r4.read-only`          | N,S | контракт | план | —   | ✓   | —   | —     |
| REQ-10.10 | Список правил                              | UC-10 | US-10 | `GET /contact-limits`                                       | `limits.list`                  | P,C | контракт | план | ✓   | —   | —   | —     |
| REQ-10.11 | Изменение и удаление правила               | UC-10 | US-10 | `PATCH /contact-limits/{id}`, `DELETE /contact-limits/{id}` | `limits.update-delete`         | P   | контракт | план | ✓   | —   | —   | —     |

## 11. US-11 / UC-11. Подтверждение и разъединение склеек

| ID        | Требование                                   | UC    | US    | Endpoint                                             | Тест                             | Тип | EP       | T    | P   | N   | B   | E2E   |
| --------- | -------------------------------------------- | ----- | ----- | ---------------------------------------------------- | -------------------------------- | --- | -------- | ---- | --- | --- | --- | ----- |
| REQ-11.1  | Очередь с причиной совпадения и `score`      | UC-11 | US-11 | `GET /merge-candidates`                              | `merge.queue.pending`            | P,C | ✓        | код  | ✓   | —   | —   | E2E-3 |
| REQ-11.2  | Фильтр по `resolution`                       | UC-11 | US-11 | `GET /merge-candidates?resolution=`                  | `merge.queue.filter`             | P   | ✓        | код  | ✓   | —   | —   | —     |
| REQ-11.3  | Детали пары: два профиля рядом               | UC-11 | US-11 | `GET /merge-candidates/{id}`                         | `merge.detail.two-profiles`      | P,C | ✓        | код  | ✓   | —   | —   | E2E-3 |
| REQ-11.4  | Confirm переносит события и идентификаторы   | UC-11 | US-11 | `POST /merge-candidates/{id}/confirm`                | `merge.confirm.moves-data`       | P,B | ✓        | код  | ✓   | —   | ✓   | E2E-3 |
| REQ-11.5  | Поглощённый профиль получает `status=merged` | UC-11 | US-11 | `GET /guests/{id}`                                   | `merge.confirm.loser-merged`     | B   | ✓        | код  | ✓   | —   | ✓   | E2E-3 |
| REQ-11.6  | Повторное решение → 409 `already_resolved`   | UC-11 | US-11 | `POST /merge-candidates/{id}/confirm`                | `merge.confirm.already-resolved` | N,B | ✓        | код  | —   | ✓   | ✓   | E2E-3 |
| REQ-11.7  | Reject создаёт запрет склейки                | UC-11 | US-11 | `POST /merge-candidates/{id}/reject`                 | `merge.reject.creates-block`     | P,B | ✓        | код  | ✓   | —   | ✓   | E2E-3 |
| REQ-11.8  | Revert возвращает данные по `originGuestId`  | UC-11 | US-11 | `POST /merge-candidates/{id}/revert`                 | `merge.revert.restores`          | P,B | ✓        | код  | ✓   | —   | ✓   | E2E-3 |
| REQ-11.9  | Revert нерешённой пары → 409                 | UC-11 | —     | `POST /merge-candidates/{id}/revert`                 | `merge.revert.not-merged`        | N   | ✓        | код  | —   | ✓   | —   | —     |
| REQ-11.10 | Пара под запретом → 409 `merge_blocked`      | UC-11 | US-11 | `POST /guests/{id}/merge`                            | `merge.manual.blocked`           | N,B | контракт | план | —   | ✓   | ✓   | E2E-3 |
| REQ-11.11 | Ручное объединение с собой → 400             | UC-11 | —     | `POST /guests/{id}/merge`                            | `merge.manual.self`              | N   | контракт | план | —   | ✓   | —   | —     |
| REQ-11.12 | Список запретов и снятие запрета             | UC-11 | US-11 | `GET /unmerge-blocks`, `DELETE /unmerge-blocks/{id}` | `merge.blocks.list-delete`       | P   | контракт | план | ✓   | —   | —   | E2E-3 |
| REQ-11.13 | Решение по склейке пишется в аудит           | UC-11 | US-11 | `GET /audit-logs`                                    | `merge.confirm.audit`            | B   | контракт | —    | —   | —   | —   | —     |
| REQ-11.14 | R-1 только читает очередь                    | UC-11 | US-11 | `POST /merge-candidates/{id}/confirm`                | `merge.r1.read-only`             | N,S | ✓        | план | —   | ✓   | —   | —     |

## 12. UC-12. Автосклейка профилей

| ID       | Требование                           | UC    | US  | Endpoint | Тест | Тип | EP  | T   | P   | N   | B   | E2E |
| -------- | ------------------------------------ | ----- | --- | -------- | ---- | --- | --- | --- | --- | --- | --- | --- |
| REQ-12.1 | Автосклейка по надёжным ключам       | UC-12 | —   | —        | —    | —   | —   | —   | —   | —   | —   | —   |
| REQ-12.2 | Спорная пара при конфликте признаков | UC-12 | —   | —        | —    | —   | —   | —   | —   | —   | —   | —   |
| REQ-12.3 | Порог автосклейки `score ≥ 80`       | UC-12 | —   | —        | —    | —   | —   | —   | —   | —   | —   | —   |

Требования UC-12 не проверяемы через API: актор R-6, REST отсутствует
(`14-traceability.md` §2, «Причина»). Косвенно наблюдаемы через `merge_candidate`,
но создание кандидата требует фикстуры в БД.

## 13. UC-16. Вход в систему

| ID       | Требование                                | UC    | US   | Endpoint             | Тест                           | Тип | EP       | T    | P   | N   | B   | E2E   |
| -------- | ----------------------------------------- | ----- | ---- | -------------------- | ------------------------------ | --- | -------- | ---- | --- | --- | --- | ----- |
| REQ-16.1 | Вход по email и паролю, access 15 мин     | UC-16 | —    | `POST /auth/login`   | `auth.login.success`           | P,C | ✓        | код  | ✓   | —   | —   | E2E-1 |
| REQ-16.2 | Одинаковый 401 на неверный email и пароль | UC-16 | —    | `POST /auth/login`   | `auth.login.wrong-credentials` | N   | ✓        | код  | —   | ✓   | —   | —     |
| REQ-16.3 | Пустое тело → 400                         | UC-16 | —    | `POST /auth/login`   | `auth.login.empty-body`        | N,C | ✓        | код  | —   | ✓   | —   | —     |
| REQ-16.4 | Перебор паролей → 429                     | UC-16 | —    | `POST /auth/login`   | `auth.login.rate-limit`        | N   | ✓        | —    | —   | —   | —   | —     |
| REQ-16.5 | Refresh по httpOnly cookie                | UC-16 | —    | `POST /auth/refresh` | `auth.refresh.rotates-access`  | P,C | контракт | план | ✓   | —   | —   | —     |
| REQ-16.6 | Отзыв refresh                             | UC-16 | —    | `POST /auth/logout`  | `auth.logout.revokes`          | P   | контракт | план | ✓   | —   | —   | —     |
| REQ-16.7 | Профиль, роль и `tenantId`                | UC-16 | US-8 | `GET /auth/me`       | `auth.me.returns-role`         | P,C | ✓        | код  | ✓   | —   | —   | E2E-5 |
| REQ-16.8 | Запрос без токена → 401                   | UC-16 | —    | любой приватный      | `auth.no-token.401`            | N,S | ✓        | код  | —   | ✓   | —   | —     |
| REQ-16.9 | Битый или просроченный токен → 401        | UC-16 | —    | любой приватный      | `auth.bad-token.401`           | N,S | ✓        | код  | —   | ✓   | —   | —     |

## 14. UC-17. Справочники и источники

| ID        | Требование                               | UC    | US   | Endpoint                                            | Тест                                   | Тип | EP       | T    | P   | N   | B   | E2E   |
| --------- | ---------------------------------------- | ----- | ---- | --------------------------------------------------- | -------------------------------------- | --- | -------- | ---- | --- | --- | --- | ----- |
| REQ-17.1  | Создание арендатора                      | UC-17 | US-1 | `POST /tenants`                                     | `tenants.create`                       | P,C | контракт | план | ✓   | —   | —   | E2E-1 |
| REQ-17.2  | Дубль `code` → 409                       | UC-17 | —    | `POST /tenants`                                     | `tenants.create.duplicate-code`        | N   | контракт | план | —   | ✓   | —   | —     |
| REQ-17.3  | R-1 меняет только `priority`             | UC-17 | —    | `PATCH /tenants/{id}`                               | `tenants.patch.marketer-priority-only` | N,B | контракт | план | —   | ✓   | ✓   | —     |
| REQ-17.4  | Удаление арендатора с событиями → 409    | UC-17 | —    | `DELETE /tenants/{id}`                              | `tenants.delete.with-events`           | N,B | контракт | план | —   | ✓   | ✓   | —     |
| REQ-17.5  | Смежность и конкуренция категорий        | UC-17 | US-4 | `PUT /tenant-categories/{id}/relations`             | `categories.relations.replace`         | P,B | контракт | план | ✓   | —   | ✓   | —     |
| REQ-17.6  | `top_up` и `usage` не считаются визитом  | UC-17 | —    | `POST /event-types`                                 | `event-types.counts-as-presence`       | P,B | контракт | план | ✓   | —   | ✓   | —     |
| REQ-17.7  | Источник с маппингом полей               | UC-17 | US-1 | `POST /data-sources`                                | `data-sources.create.mapping`          | P,C | контракт | план | ✓   | —   | —   | E2E-1 |
| REQ-17.8  | Деактивация источника с загрузками → 409 | UC-17 | —    | `DELETE /data-sources/{id}`                         | `data-sources.delete.with-imports`     | N,B | контракт | план | —   | ✓   | ✓   | —     |
| REQ-17.9  | Проекты арендатора                       | UC-17 | US-2 | `GET /projects`, `POST /projects`                   | `projects.crud`                        | P,C | контракт | план | ✓   | —   | —   | —     |
| REQ-17.10 | Типы источников данных                   | UC-17 | US-1 | `GET /data-source-types`, `POST /data-source-types` | `data-source-types.crud`               | P,C | контракт | план | ✓   | —   | —   | —     |

## 15. UC-18. Управление пользователями

| ID       | Требование                            | UC    | US  | Endpoint                        | Тест                                  | Тип | EP       | T    | P   | N   | B   | E2E   |
| -------- | ------------------------------------- | ----- | --- | ------------------------------- | ------------------------------------- | --- | -------- | ---- | --- | --- | --- | ----- |
| REQ-18.1 | R-2 создаёт пользователя              | UC-18 | —   | `POST /users`                   | `users.create.by-manager`             | P,C | контракт | план | ✓   | —   | —   | E2E-5 |
| REQ-18.2 | Роль `tenant` требует `tenantId`      | UC-18 | —   | `POST /users`                   | `users.create.tenant-requires-tenant` | N,B | контракт | план | —   | ✓   | ✓   | —     |
| REQ-18.3 | Дубль email → 409                     | UC-18 | —   | `POST /users`                   | `users.create.duplicate-email`        | N   | контракт | план | —   | ✓   | —   | —     |
| REQ-18.4 | Последний R-2 не деактивируется → 409 | UC-18 | —   | `DELETE /users/{id}`            | `users.delete.last-manager`           | N,B | контракт | план | —   | ✓   | ✓   | —     |
| REQ-18.5 | R-1, R-3, R-4 → 403                   | UC-18 | —   | `/users*`                       | `users.forbidden-roles`               | N,S | контракт | план | —   | ✓   | —   | —     |
| REQ-18.6 | Список и карточка пользователя        | UC-18 | —   | `GET /users`, `GET /users/{id}` | `users.list-get`                      | P,C | контракт | план | ✓   | —   | —   | —     |

## 16. UC-19. Журнал аудита

| ID       | Требование                         | UC    | US  | Endpoint          | Тест                      | Тип | EP       | T    | P   | N   | B   | E2E |
| -------- | ---------------------------------- | ----- | --- | ----------------- | ------------------------- | --- | -------- | ---- | --- | --- | --- | --- |
| REQ-19.1 | Журнал с фильтрами                 | UC-19 | —   | `GET /audit-logs` | `audit.list.filters`      | P,C | контракт | план | ✓   | —   | —   | —   |
| REQ-19.2 | Каждая запись содержит `requestId` | UC-19 | —   | `GET /audit-logs` | `audit.record.request-id` | B   | контракт | —    | —   | —   | —   | —   |
| REQ-19.3 | R-1 и R-3 → 403                    | UC-19 | —   | `GET /audit-logs` | `audit.forbidden-roles`   | N,S | контракт | план | —   | ✓   | —   | —   |

## 17. UC-20. Кампании и оценка эффекта

| ID       | Требование                           | UC    | US  | Endpoint                      | Тест                            | Тип | EP       | T    | P   | N   | B   | E2E   |
| -------- | ------------------------------------ | ----- | --- | ----------------------------- | ------------------------------- | --- | -------- | ---- | --- | --- | --- | ----- |
| REQ-20.1 | Создание кампании                    | UC-20 | —   | `POST /campaigns`             | `campaigns.create`              | P,C | контракт | план | ✓   | —   | —   | E2E-7 |
| REQ-20.2 | Конверсия и инкремент по кампании    | UC-20 | —   | `GET /campaigns/{id}/results` | `campaigns.results.conversion`  | P,C | контракт | план | ✓   | —   | —   | E2E-7 |
| REQ-20.3 | Удаление кампании с выгрузками → 409 | UC-20 | —   | `DELETE /campaigns/{id}`      | `campaigns.delete.with-exports` | N,B | контракт | план | —   | ✓   | ✓   | —     |
| REQ-20.4 | Окно атрибуции настраиваемое         | UC-20 | —   | `POST /campaigns`             | `campaigns.attribution-window`  | B   | контракт | план | ✓   | —   | ✓   | E2E-7 |

## 18. Служебные требования

| ID     | Требование                        | UC  | US  | Endpoint       | Тест            | Тип | EP  | T    | P   | N   | B   | E2E |
| ------ | --------------------------------- | --- | --- | -------------- | --------------- | --- | --- | ---- | --- | --- | --- | --- |
| REQ-S1 | `/healthz` доступен без токена    | —   | —   | `GET /healthz` | `health.public` | P,C | ✓   | код  | ✓   | —   | —   | —   |
| REQ-S2 | `/healthz` проверяет БД и очередь | —   | —   | `GET /healthz` | `health.checks` | C   | ✓   | план | ✓   | —   | —   | —   |

## 19. Сводка

| Раздел        | Требований | EP `✓` | EP `контракт` | EP `—` | T `код` | T `план` | T `—` | P      | N      | B      | E2E    |
| ------------- | ---------- | ------ | ------------- | ------ | ------- | -------- | ----- | ------ | ------ | ------ | ------ |
| US-1 / UC-1   | 15         | 0      | 15            | 0      | 0       | 15       | 0     | 8      | 7      | 7      | 6      |
| US-2 / UC-2   | 10         | 8      | 2             | 0      | 6       | 4        | 0     | 7      | 3      | 5      | 5      |
| US-3 / UC-3   | 14         | 8      | 6             | 0      | 5       | 8        | 1     | 8      | 6      | 4      | 2      |
| US-4 / UC-4   | 9          | 8      | 1             | 0      | 3       | 4        | 2     | 6      | 2      | 5      | 3      |
| US-5 / UC-5   | 12         | 0      | 12            | 0      | 0       | 12       | 0     | 8      | 4      | 5      | 7      |
| US-6 / UC-6   | 6          | 4      | 2             | 0      | 0       | 6        | 0     | 5      | 1      | 5      | 5      |
| US-7 / UC-7   | 6          | 0      | 6             | 0      | 0       | 6        | 0     | 3      | 3      | 2      | 0      |
| US-8 / UC-8   | 9          | 1      | 8             | 0      | 0       | 9        | 0     | 4      | 5      | 1      | 6      |
| US-9 / UC-9   | 5          | 0      | 5             | 0      | 0       | 5        | 0     | 4      | 1      | 1      | 2      |
| US-10 / UC-10 | 11         | 0      | 11            | 0      | 0       | 11       | 0     | 7      | 4      | 8      | 5      |
| US-11 / UC-11 | 14         | 10     | 4             | 0      | 9       | 4        | 1     | 8      | 5      | 6      | 9      |
| UC-12         | 3          | 0      | 0             | 3      | 0       | 0        | 3     | 0      | 0      | 0      | 0      |
| UC-16         | 9          | 7      | 2             | 0      | 6       | 2        | 1     | 4      | 5      | 0      | 2      |
| UC-17         | 10         | 0      | 10            | 0      | 0       | 10       | 0     | 6      | 4      | 5      | 2      |
| UC-18         | 6          | 0      | 6             | 0      | 0       | 6        | 0     | 2      | 4      | 2      | 1      |
| UC-19         | 3          | 0      | 3             | 0      | 0       | 2        | 1     | 1      | 1      | 0      | 0      |
| UC-20         | 4          | 0      | 4             | 0      | 0       | 4        | 0     | 3      | 1      | 2      | 3      |
| Служебные     | 2          | 2      | 0             | 0      | 1       | 1        | 0     | 2      | 0      | 0      | 0      |
| **Итого**     | **148**    | **48** | **97**        | **3**  | **30**  | **109**  | **9** | **86** | **56** | **58** | **58** |

Числа в колонках P, N, B, E2E не складываются в «Итого»: одно требование может
покрываться несколькими типами проверок одновременно.

## 20. Требования без endpoint (проверить через API нельзя)

| ID       | Требование                                      | Причина                                     | Что нужно                                        |
| -------- | ----------------------------------------------- | ------------------------------------------- | ------------------------------------------------ |
| REQ-12.1 | Автосклейка по надёжным ключам                  | актор R-6, REST отсутствует                 | наблюдение через `merge_candidate` после импорта |
| REQ-12.2 | Спорная пара при конфликте признаков            | то же                                       | то же                                            |
| REQ-12.3 | Порог `score ≥ 80`                              | то же                                       | то же                                            |
| —        | Пересчёт сегментов по событию (UC-13)           | только очередь `segment.recompute`          | косвенно через `computedAt` (REQ-6.3, REQ-6.4)   |
| —        | Расчёт рекомендаций по событию (UC-15)          | только очередь `recommend.refresh`          | косвенно через `POST .../refresh` (REQ-4.6)      |
| —        | Схлопывание заданий по `dedup_key`              | нет REST к `job_queue`                      | тест уровня репозитория, не E2E                  |
| —        | Инкремент против контрольной группы             | требует накопления событий и окна атрибуции | REQ-20.2 при достаточных данных                  |
| —        | Match rate ≥ 70 %, доля оспоренных склеек < 1 % | агрегаты по объёму данных                   | REQ-7.5 при достаточных данных                   |
| —        | Время «сегмент → выгрузка» ≤ 5 мин              | метрика по `segment_version.created_at`     | REQ-7.5                                          |
| —        | Физическое удаление запрещено                   | проверяется только косвенно                 | REQ-2.9                                          |

## 21. Тесты без требования (обоснованные технические проверки)

Эти проверки не привязаны к US/UC, но обоснованы контрактом или архитектурой.

| Тест                                  | Обоснование                                                                                 | Тип |
| ------------------------------------- | ------------------------------------------------------------------------------------------- | --- |
| `contract.openapi.operations-mounted` | сверка 90 операций контракта с роутером; фиксирует расхождение `02-contradictions.md` §2.11 | C   |
| `contract.error-format`               | единый формат `{error:{code,message,details,request_id}}` — требование шапки реестра        | C   |
| `contract.pagination-bounds`          | `page_size` 1..200 и `meta` — `components/parameters.json`                                  | C   |
| `contract.json-naming`                | `camelCase` в контракте против `snake_case` в DTO — `02-contradictions.md` §2.1             | C   |
| `contract.trailing-slash`             | коллекции смонтированы как `/guests/` — `02-contradictions.md` §2.10                        | C   |
| `contract.request-id-header`          | `X-Request-Id` в ответе и в теле ошибки — `httpx.RequestIDMiddleware`                       | C   |
| `contract.healthz-schema`             | `HealthResponse` требует `version`, `time`, `checks`                                        | C   |

## 22. Ограничения матрицы

1. **Колонка T не содержит `✓`.** Тестовый код не написан: в репозитории нет
   ни одного `*_test.go` и нет E2E-инфраструктуры (`02-contradictions.md` G11).
2. **98 требований имеют endpoint только в контракте.** Они не смонтированы в
   `cmd/api/main.go`; тесты для них спроектированы, но выполнимы только после
   реализации (`01-requirements-map.md` §5).
3. **9 требований помечены `T = —`** по объективным причинам: отсутствие
   управляемой задержки (REQ-3.4, REQ-4.7), отсутствие rate limiter (REQ-16.4),
   отсутствие аудита (REQ-11.13, REQ-19.2), отсутствие REST у автосклейки
   (REQ-12.1…REQ-12.3).
4. **REQ-4.3 заблокировано противоречием D1** (`02-contradictions.md`):
   OpenAPI требует исключать конкурента, US-4 AC-3 и A-18 — понижать. Тест
   нельзя написать, не выбрав сторону.
5. **Ролевые проверки (тип S) сейчас не проходят на любом эндпоинте:**
   `auth.RequireRole` не подключён ни к одному маршруту
   (`02-contradictions.md` §2.9). Это дефект реализации, а не тестов.
