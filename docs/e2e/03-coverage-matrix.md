# 03. Матрица тестового покрытия

Основа: `docs/14-traceability.md` (US/UC → AD → SD → ER → CRUD → CMP → API).
Колонки: требование → UC → US → endpoint → тест → тип проверки.
Типы: **P** позитивный, **N** негативный, **C** контрактный (JSON Schema),
**B** бизнес-правило, **E** сквозной E2E, **S** безопасность/роли.

Статус реализации: `OK` — операция смонтирована, `MISS` — отсутствует в
`cmd/api/main.go`. Тесты, требующие `MISS`, помечены `blocked`.

## 1. Аутентификация и доступ (UC-16, UC-18)

| Требование                                | UC    | US   | Endpoint             | Тест                                  | Тип | Статус        |
| ----------------------------------------- | ----- | ---- | -------------------- | ------------------------------------- | --- | ------------- |
| Вход по email+паролю, access 15 мин       | UC-16 | —    | `POST /auth/login`   | `auth.login.success`                  | P,C | OK            |
| Одинаковый 401 на неверный email и пароль | UC-16 | —    | `POST /auth/login`   | `auth.login.wrong-credentials`        | N   | OK            |
| 400 при пустом теле                       | UC-16 | —    | `POST /auth/login`   | `auth.login.empty-body`               | N,C | OK            |
| 429 при переборе                          | UC-16 | —    | `POST /auth/login`   | `auth.login.rate-limit`               | N   | blocked (G9)  |
| Refresh по httpOnly cookie                | UC-16 | —    | `POST /auth/refresh` | `auth.refresh.rotates-access`         | P,C | MISS          |
| Отзыв refresh                             | UC-16 | —    | `POST /auth/logout`  | `auth.logout.revokes`                 | P   | MISS          |
| Профиль, роль, `tenantId`                 | UC-16 | US-8 | `GET /auth/me`       | `auth.me.returns-role`                | P,C | OK (частично) |
| 401 без токена                            | UC-16 | —    | любой приватный      | `auth.no-token.401`                   | N,S | OK            |
| 401 с истёкшим/битым токеном              | UC-16 | —    | любой приватный      | `auth.bad-token.401`                  | N,S | OK            |
| R-2 создаёт пользователя                  | UC-18 | —    | `POST /users`        | `users.create.by-manager`             | P,C | MISS          |
| R-3 требует `tenantId`                    | UC-18 | —    | `POST /users`        | `users.create.tenant-requires-tenant` | N,B | MISS          |
| Дубль email → 409                         | UC-18 | —    | `POST /users`        | `users.create.duplicate-email`        | N   | MISS          |
| Последний R-2 не деактивируется           | UC-18 | —    | `DELETE /users/{id}` | `users.delete.last-manager`           | N,B | MISS          |
| R-1/R-3/R-4 не управляют пользователями   | UC-18 | —    | `/users*`            | `users.forbidden-roles`               | N,S | MISS          |

## 2. Справочники и источники (UC-17)

| Требование                               | UC    | US   | Endpoint                                | Тест                                   | Тип | Статус |
| ---------------------------------------- | ----- | ---- | --------------------------------------- | -------------------------------------- | --- | ------ |
| Создание арендатора                      | UC-17 | US-1 | `POST /tenants`                         | `tenants.create`                       | P,C | MISS   |
| Уникальность `code` → 409                | UC-17 | —    | `POST /tenants`                         | `tenants.create.duplicate-code`        | N   | MISS   |
| R-1 меняет только `priority`             | UC-17 | —    | `PATCH /tenants/{id}`                   | `tenants.patch.marketer-priority-only` | N,B | MISS   |
| Удаление арендатора с событиями → 409    | UC-17 | —    | `DELETE /tenants/{id}`                  | `tenants.delete.with-events`           | N,B | MISS   |
| R-3 видит только свой арендатор          | UC-8  | US-8 | `GET /tenants`, `GET /tenants/{id}`     | `tenants.r3.own-only`                  | N,S | MISS   |
| Смежность/конкуренция категорий          | UC-17 | US-4 | `PUT /tenant-categories/{id}/relations` | `categories.relations.replace`         | P,B | MISS   |
| Типы событий: `top_up`/`usage` не визит  | UC-17 | —    | `POST /event-types`                     | `event-types.counts-as-presence`       | P,B | MISS   |
| Источник с маппингом полей               | UC-17 | US-1 | `POST /data-sources`                    | `data-sources.create.mapping`          | P,C | MISS   |
| Деактивация источника с загрузками → 409 | UC-17 | —    | `DELETE /data-sources/{id}`             | `data-sources.delete.with-imports`     | N,B | MISS   |

## 3. Импорт (UC-1, UC-14)

| Требование                          | UC    | US   | Endpoint                                     | Тест                                   | Тип | Статус |
| ----------------------------------- | ----- | ---- | -------------------------------------------- | -------------------------------------- | --- | ------ |
| Файл принят → 202 `queued`          | UC-1  | US-1 | `POST /imports`                              | `imports.upload.accepted`              | P,C | MISS   |
| Статус и счётчики                   | UC-1  | US-1 | `GET /imports/{id}`                          | `imports.status.counters`              | P,C | MISS   |
| `accepted+rejected+duplicate=total` | UC-1  | US-1 | `GET /imports/{id}`                          | `imports.counters.invariant`           | B   | MISS   |
| Повтор файла → 200 `duplicate=true` | UC-1  | US-1 | `POST /imports`                              | `imports.upload.idempotent`            | B   | MISS   |
| Дубль `external_id` в другом файле  | UC-1  | US-1 | `POST /imports`                              | `imports.upload.duplicate-external-id` | B   | MISS   |
| Строка без `external_id` → rejected | UC-1  | US-1 | `GET /imports/{id}/rejected-rows`            | `imports.rejected.no-external-id`      | N,B | MISS   |
| Плохая дата → rejected              | UC-1  | US-1 | `GET /imports/{id}/rejected-rows`            | `imports.rejected.bad-date`            | N,B | MISS   |
| CSV отклонённых строк               | UC-1  | US-1 | `GET /imports/{id}/rejected-rows?format=csv` | `imports.rejected.csv`                 | P,C | MISS   |
| > 100 000 строк → 413               | UC-1  | US-1 | `POST /imports`                              | `imports.upload.too-large`             | N   | MISS   |
| Неверный формат → 422               | UC-1  | US-1 | `POST /imports`                              | `imports.upload.bad-format`            | N   | MISS   |
| Несуществующий источник → 404       | UC-1  | US-1 | `POST /imports`                              | `imports.upload.unknown-source`        | N   | MISS   |
| R-1/R-3 → 403                       | UC-1  | US-1 | `POST /imports`                              | `imports.upload.forbidden`             | N,S | MISS   |
| Ручной pull → 202                   | UC-14 | US-1 | `POST /data-sources/{id}/sync`               | `sync.manual.accepted`                 | P   | MISS   |
| Повторный pull → 409                | UC-14 | —    | `POST /data-sources/{id}/sync`               | `sync.manual.in-progress`              | N   | MISS   |

## 4. Профиль гостя (UC-2)

| Требование                                | UC   | US   | Endpoint                    | Тест                             | Тип | Статус        |
| ----------------------------------------- | ---- | ---- | --------------------------- | -------------------------------- | --- | ------------- |
| Поиск по телефону в разных форматах       | UC-2 | US-2 | `GET /guests?q=`            | `guests.search.phone-normalized` | P,B | OK (частично) |
| `q` короче 2 → 400                        | UC-2 | US-2 | `GET /guests?q=`            | `guests.search.q-too-short`      | N,C | OK (дефект)   |
| Поглощённый профиль отдаёт `mergedIntoId` | UC-2 | US-2 | `GET /guests/{id}`          | `guests.get.merged-into`         | P,B | OK            |
| Несуществующий гость → 404                | UC-2 | US-2 | `GET /guests/{id}`          | `guests.get.not-found`           | N   | OK            |
| Лента событий по убыванию даты            | UC-2 | US-2 | `GET /guests/{id}/events`   | `guests.events.sorted-desc`      | P,B | OK            |
| Фильтры `from`/`to`/`tenant_id`           | UC-2 | US-2 | `GET /guests/{id}/events`   | `guests.events.filters`          | P   | OK (дефект)   |
| `tenantCount` и список арендаторов        | UC-2 | US-2 | `GET /guests/{id}`          | `guests.get.tenant-count`        | P,B | OK            |
| История контактов                         | UC-2 | US-2 | `GET /guests/{id}/contacts` | `guests.contacts.list`           | P,C | MISS          |
| Soft-delete гостя                         | UC-2 | US-2 | `DELETE /guests/{id}`       | `guests.delete.soft`             | P,B | MISS          |
| R-3 → 403                                 | UC-2 | US-2 | `/guests*`                  | `guests.forbidden-tenant-role`   | N,S | OK (дефект)   |

## 5. Сегменты (UC-3, UC-6, UC-13)

| Требование                                  | UC    | US   | Endpoint                        | Тест                                  | Тип | Статус        |
| ------------------------------------------- | ----- | ---- | ------------------------------- | ------------------------------------- | --- | ------------- |
| Preview: размер и примеры, без записи       | UC-3  | US-3 | `POST /segments/preview`        | `segments.preview.no-write`           | P,B | OK            |
| Preview: арендатор в `was` и `wasNot` → 400 | UC-3  | US-3 | `POST /segments/preview`        | `segments.preview.conflicting-tenant` | N,B | OK            |
| Preview: пустой `was` → 400                 | UC-3  | US-3 | `POST /segments/preview`        | `segments.preview.empty-was`          | N   | OK            |
| Preview: таймаут → 504                      | UC-3  | US-3 | `POST /segments/preview`        | `segments.preview.timeout`            | N   | blocked (G10) |
| Создание сегмента, версия 1                 | UC-3  | US-3 | `POST /segments`                | `segments.create.version-1`           | P,C | OK            |
| Создание без `name` → 400                   | UC-3  | US-3 | `POST /segments`                | `segments.create.no-name`             | N   | OK            |
| Правка правила → версия 2                   | UC-3  | US-3 | `PUT /segments/{id}`            | `segments.update.new-version`         | P,B | MISS          |
| История версий                              | UC-3  | US-3 | `GET /segments/{id}/versions`   | `segments.versions.list`              | P,C | MISS          |
| Состав сегмента постранично                 | UC-3  | US-3 | `GET /segments/{id}/members`    | `segments.members.paged`              | P,C | MISS          |
| 5 шаблонов                                  | UC-3  | US-3 | `GET /segment-templates`        | `segments.templates.five`             | P,C | MISS          |
| Архив сегмента                              | UC-3  | US-3 | `DELETE /segments/{id}`         | `segments.delete.archive`             | P   | MISS          |
| Архив при активной кампании → 409           | UC-3  | —    | `DELETE /segments/{id}`         | `segments.delete.active-campaign`     | N,B | MISS          |
| Пересчёт → 202                              | UC-13 | US-6 | `POST /segments/{id}/recompute` | `segments.recompute.accepted`         | P   | MISS          |
| Повторный пересчёт → 409                    | UC-13 | —    | `POST /segments/{id}/recompute` | `segments.recompute.in-progress`      | N   | MISS          |
| `computedAt` обновляется после события      | UC-6  | US-6 | `GET /segments/{id}`            | `segments.computed-at.after-event`    | E,B | blocked (G3)  |
| R-3 → 403                                   | UC-3  | US-3 | `/segments*`                    | `segments.forbidden-tenant-role`      | N,S | OK (дефект)   |

## 6. Рекомендации (UC-4, UC-15)

| Требование                          | UC    | US   | Endpoint                                    | Тест                             | Тип | Статус        |
| ----------------------------------- | ----- | ---- | ------------------------------------------- | -------------------------------- | --- | ------------- |
| ≤ 3 рекомендаций с обоснованием     | UC-4  | US-4 | `GET /guests/{id}/recommendations`          | `reco.list.max-three`            | P,B | OK            |
| Посещённые арендаторы отсутствуют   | UC-4  | US-4 | `GET /guests/{id}/recommendations`          | `reco.list.excludes-visited`     | B   | OK            |
| Конкурент присутствует с понижением | UC-4  | US-4 | `GET /guests/{id}/recommendations`          | `reco.list.competitor-penalized` | B   | blocked (D1)  |
| Fallback для гостя без истории      | UC-4  | US-4 | `GET /guests/{id}/recommendations`          | `reco.list.fallback-no-history`  | B   | OK (дефект)   |
| Пустой список — валидное состояние  | UC-4  | US-4 | `GET /guests/{id}/recommendations`          | `reco.list.empty-ok`             | P   | OK            |
| Refresh пересчитывает набор         | UC-15 | US-4 | `POST /guests/{id}/recommendations/refresh` | `reco.refresh.replaces-set`      | P,B | OK            |
| Refresh: таймаут → 504              | UC-15 | —    | `POST .../refresh`                          | `reco.refresh.timeout`           | N   | blocked (G10) |
| Агрегат по сегменту                 | UC-4  | US-4 | `GET /segments/{id}/recommendations`        | `reco.segment.aggregate`         | P,C | MISS          |
| R-3 → 403                           | UC-4  | US-4 | `/guests/{id}/recommendations`              | `reco.forbidden-tenant-role`     | N,S | OK (дефект)   |

## 7. Выгрузка и лимиты (UC-5, UC-10, UC-20)

| Требование                            | UC    | US    | Endpoint                      | Тест                                | Тип | Статус |
| ------------------------------------- | ----- | ----- | ----------------------------- | ----------------------------------- | --- | ------ |
| Preview: 4 счётчика                   | UC-5  | US-5  | `POST /exports/preview`       | `exports.preview.counters`          | P,C | MISS   |
| Фильтр по контакту канала             | UC-5  | US-5  | `POST /exports/preview`       | `exports.preview.channel-contact`   | B   | MISS   |
| Создание выгрузки → 201               | UC-5  | US-5  | `POST /exports`               | `exports.create.ready`              | P,C | MISS   |
| Без `Idempotency-Key` → 400           | UC-5  | US-5  | `POST /exports`               | `exports.create.no-idempotency-key` | N   | MISS   |
| Повтор ключа → та же выгрузка         | UC-5  | US-5  | `POST /exports`               | `exports.create.idempotent`         | B   | MISS   |
| Сегмент не `ready` → 409              | UC-5  | US-5  | `POST /exports`               | `exports.create.segment-not-ready`  | N,B | MISS   |
| Все исключены лимитом → 409           | UC-10 | US-10 | `POST /exports`               | `exports.create.all-blocked`        | N,B | MISS   |
| Контрольная группа не в CSV           | UC-5  | US-5  | `GET /exports/{id}/download`  | `exports.download.excludes-control` | B   | MISS   |
| `is_control` фильтр в members         | UC-5  | US-5  | `GET /exports/{id}/members`   | `exports.members.control-filter`    | P   | MISS   |
| Инвариант счётчиков                   | UC-5  | US-5  | `GET /exports/{id}`           | `exports.counters.invariant`        | B   | MISS   |
| Не готова → 409 на download           | UC-5  | —     | `GET /exports/{id}/download`  | `exports.download.not-ready`        | N   | MISS   |
| Правило лимита: `maxContacts<1` → 400 | UC-10 | US-10 | `POST /contact-limits`        | `limits.create.invalid`             | N,B | MISS   |
| Самое строгое правило                 | UC-10 | US-10 | `POST /exports/preview`       | `limits.strictest-wins`             | B   | MISS   |
| Без правил лимит не применяется       | UC-10 | US-10 | `POST /exports/preview`       | `limits.none-configured`            | B   | MISS   |
| R-4 читает, но не меняет лимиты       | UC-10 | US-10 | `/contact-limits*`            | `limits.r4.read-only`               | N,S | MISS   |
| Кампания: конверсия и инкремент       | UC-20 | —     | `GET /campaigns/{id}/results` | `campaigns.results.conversion`      | P,C | MISS   |
| Удаление кампании с выгрузками → 409  | UC-20 | —     | `DELETE /campaigns/{id}`      | `campaigns.delete.with-exports`     | N,B | MISS   |

## 8. Идентичность (UC-11, UC-12)

| Требование                                 | UC    | US    | Endpoint                            | Тест                             | Тип | Статус      |
| ------------------------------------------ | ----- | ----- | ----------------------------------- | -------------------------------- | --- | ----------- |
| Очередь с причиной и score                 | UC-11 | US-11 | `GET /merge-candidates`             | `merge.queue.pending`            | P,C | OK          |
| Фильтр по `resolution`                     | UC-11 | US-11 | `GET /merge-candidates?resolution=` | `merge.queue.filter`             | P   | OK          |
| Детали пары: два профиля                   | UC-11 | US-11 | `GET /merge-candidates/{id}`        | `merge.detail.two-profiles`      | P,C | OK (дефект) |
| Confirm переносит события и идентификаторы | UC-11 | US-11 | `POST .../confirm`                  | `merge.confirm.moves-data`       | P,B | OK          |
| Confirm: поглощённый профиль `merged`      | UC-11 | US-11 | `GET /guests/{id}`                  | `merge.confirm.loser-merged`     | B   | OK          |
| Повторный confirm → 409                    | UC-11 | US-11 | `POST .../confirm`                  | `merge.confirm.already-resolved` | N,B | OK (код)    |
| Reject создаёт запрет                      | UC-11 | US-11 | `POST .../reject`                   | `merge.reject.creates-block`     | P,B | OK          |
| Revert возвращает данные                   | UC-11 | US-11 | `POST .../revert`                   | `merge.revert.restores`          | P,B | OK          |
| Revert нерешённой пары → 409               | UC-11 | —     | `POST .../revert`                   | `merge.revert.not-merged`        | N   | OK          |
| Пара под запретом → 409                    | UC-11 | US-11 | `POST /guests/{id}/merge`           | `merge.manual.blocked`           | N,B | MISS        |
| Ручное объединение с собой → 400           | UC-11 | —     | `POST /guests/{id}/merge`           | `merge.manual.self`              | N   | MISS        |
| Список запретов и снятие                   | UC-11 | US-11 | `/unmerge-blocks*`                  | `merge.blocks.list-delete`       | P   | MISS        |
| Аудит решения                              | UC-11 | US-11 | `GET /audit-logs`                   | `merge.confirm.audit`            | B   | MISS (G4)   |
| R-1 только читает очередь                  | UC-11 | US-11 | `POST .../confirm`                  | `merge.r1.read-only`             | N,S | OK (дефект) |

## 9. Аналитика (UC-7, UC-8, UC-9)

| Требование                        | UC   | US   | Endpoint                         | Тест                                 | Тип | Статус |
| --------------------------------- | ---- | ---- | -------------------------------- | ------------------------------------ | --- | ------ |
| Распределение 1/2/3+              | UC-7 | US-7 | `GET /dashboard/penetration`     | `dashboard.penetration.buckets`      | P,C | MISS   |
| `from > to` → 400                 | UC-7 | US-7 | `GET /dashboard/penetration`     | `dashboard.penetration.bad-period`   | N,B | MISS   |
| Период > 24 мес → 400             | UC-7 | US-7 | `GET /dashboard/penetration`     | `dashboard.penetration.period-limit` | N,B | MISS   |
| Метрики: match rate и др.         | UC-7 | US-7 | `GET /dashboard/metrics`         | `dashboard.metrics.fields`           | P,C | MISS   |
| R-3 → 403                         | UC-7 | US-7 | `/dashboard/*`                   | `dashboard.forbidden-tenant-role`    | N,S | MISS   |
| Агрегаты своего арендатора        | UC-8 | US-8 | `GET /tenants/{id}/aggregates`   | `aggregates.r3.own`                  | P,C | MISS   |
| Чужой арендатор → 403             | UC-8 | US-8 | `GET /tenants/{id}/aggregates`   | `aggregates.r3.foreign`              | N,S | MISS   |
| k = 5: малые ячейки скрыты        | UC-8 | US-8 | `GET /tenants/{id}/aggregates`   | `aggregates.k-threshold`             | B   | MISS   |
| Нет `guest_id` в ответе           | UC-8 | US-8 | `GET /tenants/{id}/aggregates`   | `aggregates.no-guest-ids`            | S   | MISS   |
| R-2 читает любого арендатора      | UC-8 | US-8 | `GET /tenants/{id}/aggregates`   | `aggregates.r2.any`                  | P   | MISS   |
| Качество по источникам            | UC-9 | US-9 | `GET /data-quality/sources`      | `dq.sources.list`                    | P,C | MISS   |
| Источник без загрузок — не ошибка | UC-9 | US-9 | `GET /data-quality/sources/{id}` | `dq.source.no-data`                  | P   | MISS   |
| R-1/R-3 → 403                     | UC-9 | US-9 | `/data-quality/*`                | `dq.forbidden-roles`                 | N,S | MISS   |

## 10. Аудит и служебные (UC-19)

| Требование                        | UC    | US  | Endpoint          | Тест                      | Тип | Статус      |
| --------------------------------- | ----- | --- | ----------------- | ------------------------- | --- | ----------- |
| Журнал с фильтрами                | UC-19 | —   | `GET /audit-logs` | `audit.list.filters`      | P,C | MISS        |
| `requestId` в записи              | UC-19 | —   | `GET /audit-logs` | `audit.record.request-id` | B   | MISS        |
| R-1/R-3 → 403                     | UC-19 | —   | `GET /audit-logs` | `audit.forbidden-roles`   | N,S | MISS        |
| `/healthz` без токена             | —     | —   | `GET /healthz`    | `health.public`           | P,C | OK          |
| `/healthz` проверяет БД и очередь | —     | —   | `GET /healthz`    | `health.checks`           | C   | OK (дефект) |

## 11. Сводка покрытия

| Раздел                        | Требований | Тестов  | OK     | MISS   | blocked |
| ----------------------------- | ---------- | ------- | ------ | ------ | ------- |
| Аутентификация и пользователи | 14         | 14      | 5      | 9      | 1       |
| Справочники и источники       | 9          | 9       | 0      | 9      | 0       |
| Импорт                        | 14         | 14      | 0      | 14     | 0       |
| Профиль гостя                 | 10         | 10      | 7      | 3      | 0       |
| Сегменты                      | 16         | 16      | 6      | 10     | 1       |
| Рекомендации                  | 9          | 9       | 6      | 1      | 2       |
| Выгрузка и лимиты             | 17         | 17      | 0      | 17     | 0       |
| Идентичность                  | 13         | 13      | 9      | 4      | 0       |
| Аналитика                     | 13         | 13      | 0      | 13     | 0       |
| Аудит и служебные             | 5          | 5       | 2      | 3      | 0       |
| **Итого**                     | **120**    | **120** | **35** | **83** | **4**   |

Из 120 запланированных проверок выполнимы сейчас 35 (29 %), из них 12 проходят
только при ручной подготовке данных в БД и 8 фиксируют известные дефекты
(помечены «дефект» в колонке «Статус»).

## 12. Требования без endpoint (проверить нельзя)

| Требование                             | Причина                                 |
| -------------------------------------- | --------------------------------------- |
| Автосклейка по надёжным ключам (UC-12) | актор R-6, REST отсутствует             |
| Пересчёт сегментов по событию (UC-13)  | только очередь                          |
| Расчёт рекомендаций по событию (UC-15) | только очередь                          |
| Схлопывание заданий по `dedup_key`     | нет REST к `job_queue`                  |
| Инкремент против контрольной группы    | требует накопления событий              |
| Match rate, доля оспоренных склеек     | агрегаты по объёму данных               |
| Время «сегмент → выгрузка» ≤ 5 мин     | метрика по `segment_version.created_at` |
| Rate limit логина                      | не реализован, порог не задан           |
| Таймауты 504                           | нет управляемой задержки                |
| Физическое удаление запрещено          | проверяется только косвенно             |
