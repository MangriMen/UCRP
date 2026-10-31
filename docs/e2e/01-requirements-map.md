# 01. Структура требований, бизнес-сущности и карта API

Источники: `docs/03-use-cases.md`, `docs/04-user-stories.md`, `docs/05-er-data-model.md`,
`docs/07-crud-matrix.md`, `docs/09-sequence-diagrams.md`, `docs/14-traceability.md`,
`docs/endpoint-registry.md`, `docs/openapi.json`.

## 1. Структура требований

### 1.1. Роли

| ID  | Роль        | `role` в JWT | Доступ                                                                  |
| --- | ----------- | ------------ | ----------------------------------------------------------------------- |
| R-1 | Маркетолог  | `marketer`   | сегменты, рекомендации, выгрузки, лимиты, кампании, чтение профилей     |
| R-2 | Управляющий | `manager`    | дашборды, журналы, пользователи, чтение справочников и выгрузок         |
| R-3 | Арендатор   | `tenant`     | только свои агрегаты и своя карточка арендатора (`tenant_id` из claims) |
| R-4 | Аналитик    | `analyst`    | импорт, склейки, справочники, качество данных, аудит                    |
| R-5 | Гость       | —            | доступа нет, объект данных                                              |
| R-6 | Система     | `system`     | только очередь `job_queue`, REST не использует                          |

### 1.2. Пользовательские истории и прецеденты

| US    | UC    | Название                          | Приоритет   | Ключевые эндпоинты                                               |
| ----- | ----- | --------------------------------- | ----------- | ---------------------------------------------------------------- |
| US-1  | UC-1  | Загрузка данных арендатора        | MVP         | `POST /imports`, `GET /imports/{id}`, `/rejected-rows`           |
| US-2  | UC-2  | Единый профиль гостя              | MVP         | `GET /guests`, `GET /guests/{id}`, `/events`, `/contacts`        |
| US-3  | UC-3  | Сегмент без кода                  | MVP         | `POST /segments/preview`, `POST /segments`, `GET /segments/{id}` |
| US-4  | UC-4  | Рекомендация гостю                | MVP         | `GET /guests/{id}/recommendations`, `POST .../refresh`           |
| US-5  | UC-5  | Выгрузка сегмента в канал         | MVP         | `POST /exports/preview`, `POST /exports`, `/download`            |
| US-6  | UC-6  | Обновление при новом событии      | MVP         | `GET /segments/{id}`, `POST /segments/{id}/recompute`            |
| US-7  | UC-7  | Сводка управляющего               | 2-я очередь | `GET /dashboard/penetration`, `/dashboard/metrics`               |
| US-8  | UC-8  | Агрегаты арендатора               | 2-я очередь | `GET /tenants/{id}/aggregates`                                   |
| US-9  | UC-9  | Качество данных по источникам     | 2-я очередь | `GET /data-quality/sources`, `/{id}`                             |
| US-10 | UC-10 | Лимит контактов                   | MVP         | `/contact-limits`, `POST /exports/preview`                       |
| US-11 | UC-11 | Подтверждение/разъединение склеек | MVP         | `/merge-candidates*`, `/unmerge-blocks*`                         |

Служебные: UC-12 автосклейка, UC-13 пересчёт сегментов, UC-14 pull из API,
UC-15 расчёт рекомендаций, UC-16 вход, UC-17 справочники, UC-18 пользователи,
UC-19 аудит, UC-20 кампании.

### 1.3. Бизнес-правила, проверяемые через API

Колонка «Статус» сверена с кодом на коммите `945a5a7`: **готово** — правило
реализовано и наблюдаемо через API; **частично** — реализовано не полностью или
возвращает другой код/форму ответа; **нет** — не реализовано.

| ID    | Правило                                                                                          | Источник            | Статус                                                                          |
| ----- | ------------------------------------------------------------------------------------------------ | ------------------- | ------------------------------------------------------------------------------- |
| BR-1  | Идемпотентность импорта: тот же sha256 → 200 + `duplicate=true`, без дублей                      | US-1 AC4, D-9       | готово                                                                          |
| BR-2  | Ключ идемпотентности события `(data_source_id, event_type_id, external_id)`                      | A-4, D-19           | готово                                                                          |
| BR-3  | `accepted + rejected + duplicate = total`                                                        | US-1 AC2            | готово                                                                          |
| BR-4  | Строка без `external_id` / с плохой датой → `rejected_row`                                       | US-1 AC3            | готово                                                                          |
| BR-5  | Поиск гостя по нормализованному телефону (`+7 900…` = `8900…`)                                   | US-2 AC4            | готово                                                                          |
| BR-6  | Поглощённый профиль отдаёт `mergedIntoId`                                                        | US-2 AC5            | готово                                                                          |
| BR-7  | Сегмент: арендатор в `was` и `wasNot` одновременно → 400                                         | US-3 AC5            | готово                                                                          |
| BR-8  | Preview не пишет в БД, возвращает размер и примеры                                               | US-3 AC1            | готово                                                                          |
| BR-9  | Правка правила создаёт новую версию, старые доступны                                             | US-3 AC4, D-7       | готово                                                                          |
| BR-10 | Рекомендации: ≤ 3, без посещённых арендаторов                                                    | US-4 AC1/AC2        | готово                                                                          |
| BR-11 | Гость без истории → fallback с пометкой «мало данных»                                            | US-4 AC4, A-21      | нет                                                                             |
| BR-12 | Выгрузка: только гости с контактом канала (sms/push → phone, email → email)                      | US-5 AC2            | готово                                                                          |
| BR-13 | Контрольная группа не попадает в CSV и в `contact_log`                                           | US-5 AC3, US-10 AC4 | готово                                                                          |
| BR-14 | Повтор с тем же `Idempotency-Key` → та же выгрузка                                               | US-5 AC5            | готово                                                                          |
| BR-15 | Сегмент не `ready` → выгрузка 409 `segment_not_ready`                                            | US-5 AC6, US-6 AC5  | частично (код есть, но `POST /segments` сразу создаёт версию `ready`)           |
| BR-16 | Лимит контактов: применяется самое строгое правило; все исключены → 409 `contact_limit_exceeded` | US-10 AC2/AC3/AC5   | готово                                                                          |
| BR-17 | Без правил лимита ограничение не применяется                                                     | US-10 AC1, A-12     | готово                                                                          |
| BR-18 | Подтверждение склейки переносит события и идентификаторы за одну транзакцию                      | US-11 AC2           | готово (одна транзакция + запись в `audit_log`)                                 |
| BR-19 | Отклонение/разъединение создаёт запрет склейки (E-12)                                            | US-11 AC3/AC4       | готово                                                                          |
| BR-20 | Пара под запретом → 409 `merge_blocked`                                                          | US-11 AC5/AC6       | частично (`conflict`)                                                           |
| BR-21 | Повторное решение по паре → 409 `already_resolved`                                               | US-11 AC6           | частично (`conflict`)                                                           |
| BR-22 | R-3: чужой `tenant_id` → 403 `forbidden_tenant`                                                  | US-8 AC2, D-13      | частично (`forbidden_role`)                                                     |
| BR-23 | R-3: значения < k = 5 скрываются (`suppressed=true`, `value=null`)                               | US-8 AC4, Q-18      | частично (`hidden=true`, `new_guests=0`)                                        |
| BR-24 | R-3: любой другой эндпоинт → 403                                                                 | US-8 AC3            | готово                                                                          |
| BR-25 | Период дашборда: `from > to` или > 24 мес → 400                                                  | US-7 AC4            | нет                                                                             |
| BR-26 | `max_contacts < 1` или `period_days < 1` → 400                                                   | US-10 AC5           | готово                                                                          |
| BR-27 | R-3 обязателен `tenantId` при создании пользователя                                              | UC-18, DDL CHECK    | готово (сервис валидирует обе стороны: R-3 без `tenantId` и не-R-3 с ним → 400) |
| BR-28 | Последний R-2 нельзя деактивировать → 409                                                        | реестр `/users`     | готово                                                                          |
| BR-29 | Удаление арендатора с событиями → 409 `tenant_has_events`                                        | реестр `/tenants`   | нет                                                                             |
| BR-30 | Аудит: каждое изменяющее действие пишет запись                                                   | UC-19, E-21         | частично (только операции склейки)                                              |

### 1.4. Бизнес-правила, не проверяемые через API

| ID    | Правило                                                | Причина                                                                                                                 |
| ----- | ------------------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------- |
| NB-1  | Автосклейка по надёжным ключам, порог score ≥ 80       | `identity.resolve` есть, но порог в коде другой: phone → score 90 `auto`, email → 70/75 `pending`; REST только косвенно |
| NB-2  | Пересчёт сегментов и рекомендаций по событию (≤ 1 мин) | воркер ставит `segment.recompute` после `identity.resolve`, но `recommend.refresh` не ставится никогда                  |
| NB-3  | Схлопывание заданий по `dedup_key`                     | реализовано в `queue.Enqueue` (`ON CONFLICT` для `queued/running`), нет REST к `job_queue`                              |
| NB-4  | Инкремент против контрольной группы                    | требует накопления событий и окна атрибуции                                                                             |
| NB-5  | Match rate ≥ 70 %, доля оспоренных склеек < 1 %        | `GET /dashboard/metrics` отдаёт `match_rate`, `contested_merge_rate`; значения зависят от объёма данных                 |
| NB-6  | Время «сегмент → выгрузка» ≤ 5 мин                     | `avg_segment_to_export_minutes` в `/dashboard/metrics`                                                                  |
| NB-7  | Понижение конкурента коэффициентом, а не исключение    | ни понижения, ни исключения нет: скоринг не читает `tenant_category_relation` (см. D1)                                  |
| NB-8  | Soft-delete без физического удаления                   | проверяется только косвенно (`status`)                                                                                  |
| NB-9  | Rate limit логина (429)                                | не реализован, порог не задан                                                                                           |
| NB-10 | Таймауты 10/25/30 с и коды 504                         | таймаут preview 10 с есть, синхронной выгрузки нет; управляемой задержки нет                                            |

## 2. Бизнес-сущности

| ID   | Сущность            | Таблицы                                        | Ключевые поля                                                                                     |
| ---- | ------------------- | ---------------------------------------------- | ------------------------------------------------------------------------------------------------- |
| E-1  | Гость               | `guest`                                        | `status` (active/merged/deleted), `merged_into_id`, `deleted_at`                                  |
| E-2  | Событие             | `event`                                        | `guest_id` (nullable), `tenant_id`, `event_type_id`, `external_id`, `occurred_at`, `amount_minor` |
| E-3  | Сегмент             | `segment`, `segment_version`, `segment_member` | `current_version_no`, `rule` (JSON), `status`, `member_count`                                     |
| E-4  | Арендатор           | `tenant`                                       | `category_id`, `kind`, `floor`, `priority`                                                        |
| E-5  | Источник данных     | `data_source`, `data_source_type`              | `tenant_id` (nullable), `field_mapping`, `schedule_cron`                                          |
| E-6  | Категория           | `tenant_category`, `tenant_category_relation`  | `relation_type` (adjacent/competitor), `weight_pct`                                               |
| E-7  | Проект              | `project`                                      | `tenant_id`, `code`                                                                               |
| E-8  | Кампания            | `campaign`                                     | `target_tenant_id`, `segment_id`, `attribution_days`, `status`                                    |
| E-9  | Тип события         | `event_type`                                   | `code`, `counts_as_presence`, `reverses_event_type_id`                                            |
| E-10 | Идентификатор гостя | `guest_identifier`                             | `id_type`, `value_normalized`, `origin_guest_id`                                                  |
| E-11 | Кандидат на склейку | `merge_candidate`                              | `guest_a_id < guest_b_id`, `score`, `resolution`                                                  |
| E-12 | Запрет склейки      | `unmerge_block`                                | `reason` (manual_unmerge/rejected), `candidate_id`                                                |
| E-13 | Загрузка            | `import_batch`                                 | `mode`, `status`, `total/accepted/rejected/duplicate_rows`, `file_hash`                           |
| E-14 | Отклонённая строка  | `rejected_row`                                 | `row_number`, `error_code`, `error_message`                                                       |
| E-15 | Рекомендация        | `recommendation`, `recommendation_factor`      | `rank` 1..3, `score`, `explanation`, `algorithm_version`                                          |
| E-16 | Выгрузка            | `export_job`                                   | `channel`, `format`, `control_percent`, `idempotency_key`, счётчики                               |
| E-17 | Член выгрузки       | `export_member`                                | `identifier_id`, `is_control`                                                                     |
| E-18 | Лимит контактов     | `contact_limit`                                | `channel` (nullable), `max_contacts`, `period_days`                                               |
| E-19 | Контакт             | `contact_log`                                  | `guest_id`, `export_id`, `contacted_at`                                                           |
| E-20 | Пользователь        | `app_user`                                     | `role`, `tenant_id`, `is_active`                                                                  |
| E-21 | Аудит               | `audit_log`                                    | `action`, `entity_type`, `entity_id`, `request_id`                                                |
| T-1  | Очередь заданий     | `job_queue`                                    | `job_type`, `dedup_key`, `status`, `attempts`                                                     |

Связи, критичные для E2E: `guest 1—N event`, `guest 1—N guest_identifier`,
`tenant 1—N event`, `segment 1—N segment_version 1—N segment_member`,
`export_job 1—N export_member`, `merge_candidate N—1 guest` (дважды),
`unmerge_block` ссылается на пару гостей и кандидата.

Справочные данные: миграция `00001_init.sql` по-прежнему не содержит `INSERT`.
Стартовые сиды в коде: администратор (`UCRP_ADMIN_EMAIL` / `UCRP_ADMIN_PASSWORD`,
роль `analyst`), типы событий (`ImportCatalogSQL.EnsureEventTypes`), типы
источников (`DataSourceTypeSQL.EnsureDefaults`: `csv_excel`, `crm`, `pos`, `api`) и
5 шаблонов сегментов (`SegmentSQL.EnsureTemplates`). `tenant`, `tenant_category`,
`project`, `data_source` в пустой БД отсутствуют — их создают через API или
демо-генератор `cmd/seed`.

## 3. Карта API (90 операций)

Обозначения: **C** create, **R** read, **U** update, **D** delete/archive,
**A** агрегаты. «Зависит от» — предусловия для E2E. Роли и коды ответов приведены
по контракту; фактические маршруты и их расхождения с контрактом — в §5 и
`02-contradictions.md` §2.

Две операции контракта не смонтированы и помечены в таблицах знаком ✗:
`GET /campaigns/{id}/results` (§3.8) и `GET /data-quality/sources/{id}` (§3.10).

### 3.1. Аутентификация (UC-16)

| M    | Путь            | Назначение                | Роли | CRUD | Обязательное        | Ответы             | Зависит от |
| ---- | --------------- | ------------------------- | ---- | ---- | ------------------- | ------------------ | ---------- |
| POST | `/auth/login`   | Вход email+пароль         | pub  | R    | `email`, `password` | 200, 400, 401, 429 | `app_user` |
| POST | `/auth/refresh` | Новый access по cookie    | pub  | R    | cookie `refresh`    | 200, 401           | login      |
| POST | `/auth/logout`  | Отзыв refresh             | any  | D    | —                   | 204, 401           | login      |
| GET  | `/auth/me`      | Профиль, роль, `tenantId` | any  | R    | —                   | 200, 401           | login      |

### 3.2. Пользователи (UC-18)

| M      | Путь          | Назначение               | Роли | CRUD | Обязательное                                         | Ответы        |
| ------ | ------------- | ------------------------ | ---- | ---- | ---------------------------------------------------- | ------------- |
| GET    | `/users`      | Список                   | R-2  | R    | —                                                    | 200, 401, 403 |
| POST   | `/users`      | Создать                  | R-2  | C    | `email`, `password`, `role`; `tenantId` для `tenant` | 201, 400, 409 |
| GET    | `/users/{id}` | Карточка                 | R-2  | R    | —                                                    | 200, 404      |
| PATCH  | `/users/{id}` | Роль, активность, пароль | R-2  | U    | —                                                    | 200, 400, 404 |
| DELETE | `/users/{id}` | Деактивация (soft)       | R-2  | D    | —                                                    | 204, 404, 409 |

### 3.3. Справочники и источники (UC-17)

| M      | Путь                                | Назначение                  | Роли                 | CRUD | Ответы             |
| ------ | ----------------------------------- | --------------------------- | -------------------- | ---- | ------------------ |
| GET    | `/tenants`                          | Список арендаторов          | R-1,R-2,R-4; R-3 own | R    | 200                |
| GET    | `/tenants/{id}`                     | Карточка                    | R-1,R-2,R-4; R-3 own | R    | 200, 403, 404      |
| POST   | `/tenants`                          | Создать                     | R-4                  | C    | 201, 400, 409      |
| PATCH  | `/tenants/{id}`                     | Изменить (R-1 — `priority`) | R-4; R-1 (priority)  | U    | 200, 400, 403, 404 |
| DELETE | `/tenants/{id}`                     | Soft-delete                 | R-4                  | D    | 204, 404, 409      |
| GET    | `/tenant-categories`                | Категории                   | R-1,R-2,R-4          | R    | 200                |
| POST   | `/tenant-categories`                | Создать                     | R-4                  | C    | 201, 400, 409      |
| PATCH  | `/tenant-categories/{id}`           | Изменить                    | R-4                  | U    | 200, 404           |
| DELETE | `/tenant-categories/{id}`           | Удалить                     | R-4                  | D    | 204, 404, 409      |
| GET    | `/tenant-categories/{id}/relations` | Смежность/конкуренция       | R-1,R-2,R-4          | R    | 200, 404           |
| PUT    | `/tenant-categories/{id}/relations` | Заменить связи              | R-4                  | U    | 200, 400, 404      |
| GET    | `/projects`                         | Проекты (`?tenant_id=`)     | R-1,R-2,R-4          | R    | 200                |
| POST   | `/projects`                         | Создать                     | R-4                  | C    | 201, 400, 409      |
| PATCH  | `/projects/{id}`                    | Изменить                    | R-4                  | U    | 200, 404           |
| DELETE | `/projects/{id}`                    | Удалить                     | R-4                  | D    | 204, 404, 409      |
| GET    | `/event-types`                      | Типы событий                | R-1,R-2,R-4          | R    | 200                |
| POST   | `/event-types`                      | Создать                     | R-4                  | C    | 201, 400, 409      |
| PATCH  | `/event-types/{id}`                 | Изменить/деактивировать     | R-4                  | U    | 200, 404           |
| DELETE | `/event-types/{id}`                 | Удалить                     | R-4                  | D    | 204, 404, 409      |
| GET    | `/data-source-types`                | Типы источников             | R-1,R-2,R-4          | R    | 200                |
| POST   | `/data-source-types`                | Зарегистрировать тип        | R-4                  | C    | 201, 400, 409      |
| PATCH  | `/data-source-types/{id}`           | Изменить                    | R-4                  | U    | 200, 404           |
| GET    | `/data-sources`                     | Источники (`?tenant_id=`)   | R-1,R-2,R-4          | R    | 200                |
| GET    | `/data-sources/{id}`                | Карточка                    | R-1,R-2,R-4          | R    | 200, 404           |
| POST   | `/data-sources`                     | Создать источник            | R-4                  | C    | 201, 400, 409      |
| PATCH  | `/data-sources/{id}`                | Конфиг, маппинг, расписание | R-4                  | U    | 200, 400, 404      |
| DELETE | `/data-sources/{id}`                | Деактивация                 | R-4                  | D    | 204, 404, 409      |

Роли в таблице — по контракту. Фактически `GET /tenants` и `GET /tenants/{id}`
ограничены `marketer`/`manager`/`analyst`, поэтому R-3 получает 403 вместо своей
карточки; `PATCH /tenants/{id}` требует `analyst`, хотя контракт разрешает R-1
менять `priority` (см. `02-contradictions.md` §2.9).

### 3.4. Импорт (UC-1, UC-14)

| M    | Путь                          | Назначение                          | Роли    | CRUD | Обязательное             | Ответы                             | Зависит от             |
| ---- | ----------------------------- | ----------------------------------- | ------- | ---- | ------------------------ | ---------------------------------- | ---------------------- |
| POST | `/imports`                    | multipart `data_source_id` + `file` | R-4     | C    | `data_source_id`, `file` | 202, 200 (dup), 400, 404, 413, 422 | `data_source`          |
| GET  | `/imports`                    | Список (`?data_source_id=&status=`) | R-4,R-2 | R    | —                        | 200                                | —                      |
| GET  | `/imports/{id}`               | Статус и счётчики                   | R-4,R-2 | R    | —                        | 200, 404                           | POST                   |
| GET  | `/imports/{id}/rejected-rows` | Отклонённые строки (`?format=csv`)  | R-4     | R    | —                        | 200, 404                           | POST                   |
| POST | `/data-sources/{id}/sync`     | Ручной pull                         | R-4     | C    | —                        | 202, 404, 409                      | `data_source` типа api |

### 3.5. Гости и профили (UC-2, UC-4)

| M      | Путь                                   | Назначение             | Роли    | CRUD | Обязательное | Ответы        |
| ------ | -------------------------------------- | ---------------------- | ------- | ---- | ------------ | ------------- |
| GET    | `/guests`                              | Поиск (`q` обязателен) | R-1,R-4 | R    | `q` (2..200) | 200, 400      |
| GET    | `/guests/{id}`                         | Профиль                | R-1,R-4 | R    | —            | 200, 404      |
| DELETE | `/guests/{id}`                         | Soft-delete            | R-4     | D    | —            | 204, 404      |
| GET    | `/guests/{id}/events`                  | Лента событий          | R-1,R-4 | R    | —            | 200, 404      |
| GET    | `/guests/{id}/contacts`                | История контактов      | R-1,R-4 | R    | —            | 200, 404      |
| GET    | `/guests/{id}/recommendations`         | ≤ 3 рекомендации       | R-1,R-4 | R    | —            | 200, 404      |
| POST   | `/guests/{id}/recommendations/refresh` | Пересчёт               | R-1,R-4 | U    | —            | 200, 404, 504 |
| GET    | `/segments/{id}/recommendations`       | Агрегат по сегменту    | R-1,R-4 | R    | —            | 200, 404      |

### 3.6. Идентичность (UC-11, UC-12)

| M      | Путь                             | Назначение                            | Роли    | CRUD | Обязательное    | Ответы             |
| ------ | -------------------------------- | ------------------------------------- | ------- | ---- | --------------- | ------------------ |
| GET    | `/merge-candidates`              | Очередь (`?resolution=&reason_code=`) | R-4,R-1 | R    | —               | 200                |
| GET    | `/merge-candidates/{id}`         | Пара профилей + причина               | R-4,R-1 | R    | —               | 200, 404           |
| POST   | `/merge-candidates/{id}/confirm` | Объединить                            | R-4     | U    | —               | 200, 404, 409      |
| POST   | `/merge-candidates/{id}/reject`  | Отклонить + запрет                    | R-4     | U    | —               | 200, 404, 409      |
| POST   | `/merge-candidates/{id}/revert`  | Разъединить + запрет                  | R-4     | U    | —               | 200, 404, 409      |
| POST   | `/guests/{id}/merge`             | Ручное объединение                    | R-4     | C    | `targetGuestId` | 200, 400, 404, 409 |
| GET    | `/unmerge-blocks`                | Список запретов                       | R-4     | R    | —               | 200                |
| DELETE | `/unmerge-blocks/{id}`           | Снять запрет                          | R-4     | D    | —               | 204, 404           |

### 3.7. Сегменты (UC-3, UC-6, UC-13)

| M      | Путь                       | Назначение                     | Роли        | CRUD | Обязательное   | Ответы        |
| ------ | -------------------------- | ------------------------------ | ----------- | ---- | -------------- | ------------- |
| GET    | `/segment-templates`       | 5 шаблонов                     | R-1,R-4     | R    | —              | 200           |
| GET    | `/segments`                | Список (`?status=&q=`)         | R-1,R-2,R-4 | R    | —              | 200           |
| POST   | `/segments`                | Создать (версия 1)             | R-1,R-4     | C    | `name`, `rule` | 201, 400      |
| POST   | `/segments/preview`        | Размер + примеры, без записи   | R-1,R-4     | R    | `rule`         | 200, 400, 504 |
| GET    | `/segments/{id}`           | Карточка, статус, `computedAt` | R-1,R-2,R-4 | R    | —              | 200, 404      |
| PUT    | `/segments/{id}`           | Новая версия правила           | R-1,R-4     | U    | `rule`         | 200, 400, 404 |
| DELETE | `/segments/{id}`           | Архивировать                   | R-1,R-4     | D    | —              | 204, 404, 409 |
| GET    | `/segments/{id}/members`   | Состав (`?page=&version=`)     | R-1,R-4     | R    | —              | 200, 404      |
| GET    | `/segments/{id}/versions`  | История версий                 | R-1,R-4     | R    | —              | 200, 404      |
| POST   | `/segments/{id}/recompute` | Пересчёт (202)                 | R-1,R-4     | U    | —              | 202, 404, 409 |

### 3.8. Кампании (UC-20)

| M      | Путь                        | Назначение            | Роли        | CRUD | Обязательное                          | Ответы        |
| ------ | --------------------------- | --------------------- | ----------- | ---- | ------------------------------------- | ------------- |
| GET    | `/campaigns`                | Список                | R-1,R-2,R-4 | R    | —                                     | 200           |
| POST   | `/campaigns`                | Создать               | R-1         | C    | `name`, `targetTenantId`, `segmentId` | 201, 400      |
| GET    | `/campaigns/{id}`           | Карточка              | R-1,R-2,R-4 | R    | —                                     | 200, 404      |
| PATCH  | `/campaigns/{id}`           | Изменить              | R-1         | U    | —                                     | 200, 404, 409 |
| DELETE | `/campaigns/{id}`           | Удалить               | R-1         | D    | —                                     | 204, 404, 409 |
| GET    | `/campaigns/{id}/results` ✗ | Конверсия и инкремент | R-1,R-2     | A    | —                                     | 200, 404      |

### 3.9. Выгрузка и контакты (UC-5, UC-10)

| M      | Путь                     | Назначение                   | Роли        | CRUD | Обязательное                                        | Ответы                       |
| ------ | ------------------------ | ---------------------------- | ----------- | ---- | --------------------------------------------------- | ---------------------------- |
| POST   | `/exports/preview`       | Счётчики до создания         | R-1         | R    | `segmentId`, `channel`                              | 200, 400, 404                |
| POST   | `/exports`               | Создать выгрузку             | R-1         | C    | `segmentId`, `channel`, `format`, `Idempotency-Key` | 201, 202, 400, 404, 409, 504 |
| GET    | `/exports`               | Журнал                       | R-1,R-2,R-4 | R    | —                                                   | 200                          |
| GET    | `/exports/{id}`          | Карточка и счётчики          | R-1,R-2,R-4 | R    | —                                                   | 200, 404                     |
| GET    | `/exports/{id}/members`  | Получатели (`?is_control=`)  | R-1         | R    | —                                                   | 200, 404                     |
| GET    | `/exports/{id}/download` | CSV (без контрольной группы) | R-1         | R    | —                                                   | 200, 404, 409                |
| GET    | `/contact-limits`        | Правила                      | R-1,R-2,R-4 | R    | —                                                   | 200                          |
| POST   | `/contact-limits`        | Создать правило              | R-1         | C    | `maxContacts` ≥ 1, `periodDays` ≥ 1                 | 201, 400                     |
| PATCH  | `/contact-limits/{id}`   | Изменить                     | R-1         | U    | —                                                   | 200, 400, 404                |
| DELETE | `/contact-limits/{id}`   | Удалить                      | R-1         | D    | —                                                   | 204, 404                     |

### 3.10. Аналитика (UC-7, UC-8, UC-9)

| M   | Путь                           | Назначение                       | Роли                 | CRUD | Ответы        |
| --- | ------------------------------ | -------------------------------- | -------------------- | ---- | ------------- |
| GET | `/dashboard/penetration`       | Распределение 1/2/3+ и динамика  | R-2,R-1,R-4          | A    | 200, 400      |
| GET | `/dashboard/metrics`           | Match rate, конверсия, инкремент | R-2,R-1,R-4          | A    | 200, 400      |
| GET | `/tenants/{id}/aggregates`     | Агрегаты арендатора (k = 5)      | R-3 own; R-2,R-1,R-4 | A    | 200, 403, 404 |
| GET | `/data-quality/sources`        | Качество по источникам           | R-4,R-2              | A    | 200           |
| GET | `/data-quality/sources/{id}` ✗ | Детали источника                 | R-4,R-2              | A    | 200, 404      |

### 3.11. Аудит и служебные (UC-19)

| M   | Путь          | Назначение                                           | Роли    | CRUD | Ответы        |
| --- | ------------- | ---------------------------------------------------- | ------- | ---- | ------------- |
| GET | `/audit-logs` | Журнал (`?entity_type=&entity_id=&actor=&from=&to=`) | R-2,R-4 | R    | 200, 400      |
| GET | `/healthz`    | Живость (вне `/api/v1`)                              | pub     | R    | 200, 503, 429 |

## 4. Задания очереди (R-6, не REST)

Колонка «Воркер» — обрабатывается ли тип в `cmd/worker/main.go` на коммите `945a5a7`.

| `job_type`          | Постановщик                                                               | Идемпотентность             | Воркер                         | UC    |
| ------------------- | ------------------------------------------------------------------------- | --------------------------- | ------------------------------ | ----- |
| `import.process`    | `POST /imports`                                                           | `import:{batch_id}`         | да                             | UC-1  |
| `connector.pull`    | `POST /data-sources/{id}/sync`                                            | `pull:{source_id}:{window}` | да (обрабатывается как импорт) | UC-14 |
| `identity.resolve`  | `import.process`                                                          | `resolve:{batch_id}`        | да                             | UC-12 |
| `identity.rematch`  | —                                                                         | —                           | нет (никем не ставится)        | UC-12 |
| `segment.recompute` | `identity.resolve`, `PUT /segments/{id}`, `POST /segments/{id}/recompute` | `segment:{id}:{version_no}` | да                             | UC-13 |
| `recommend.refresh` | —                                                                         | —                           | нет (никем не ставится)        | UC-15 |
| `export.build`      | —                                                                         | —                           | нет (выгрузка синхронная)      | UC-5  |

Планировщика (`platform/scheduler`) в запуске нет: `connector.pull` ставится
только вручную через `POST /data-sources/{id}/sync`.

## 5. Фактическое состояние реализации (сверка с контрактом)

Смонтировано в `apps/api/cmd/api/main.go` — **88 операций из 90**. Отсутствуют
всего две:

```text
GET /api/v1/campaigns/{id}/results   (контракт: x-roles R-1,R-2, UC-20)
GET /api/v1/data-quality/sources/{id} (контракт: x-roles R-4,R-2, UC-9)
```

Все группы смонтированы: аутентификация (`login`, `refresh`, `logout`, `me`),
пользователи, все справочники (`tenants`, `tenant-categories` + `relations`,
`projects`, `event-types`, `data-source-types`, `data-sources` + `sync`), импорт,
гости (профиль, события, контакты, рекомендации, ручная склейка, soft-delete),
идентичность (`merge-candidates*`, `unmerge-blocks*`), сегменты (шаблоны, CRUD,
`members`, `versions`, `recompute`, `recommendations`), выгрузки (preview, создание,
журнал, `members`, `download`), лимиты контактов, кампании, аналитика, аудит.

### 5.1. Что действительно работает

- `auth.RequireRole` подключён к каждому приватному маршруту — ролевые проверки
  403 выполняются. R-3 допущен к `/tenants/{id}/aggregates`, но карточки
  арендатора (`GET /tenants/{id}`) для него закрыты, хотя контракт их разрешает.
- Воркер обрабатывает `import.process`, `connector.pull`, `identity.resolve`,
  `segment.recompute`; после `identity.resolve` ставит пересчёт активных сегментов.
- `identity.resolve` реализован: создаёт гостя, привязывает к существующему,
  делает автосклейку по телефону (score 90) и ставит пару в очередь по email.
- Склейка подтверждается одной транзакцией и пишет запись в `audit_log`.
- Импорт: sha256-идемпотентность, счётчики, `rejected_row`, CSV отклонённых
  строк; `connector.pull` для источников типа `api`.
- Выгрузка: preview-счётчики, контрольная группа, идемпотентность по ключу,
  CSV без контрольной группы, лимиты контактов.

### 5.2. Что расходится с контрактом

Полный перечень — в `02-contradictions.md`. Кратко: имена JSON-полей всё ещё
`snake_case` у большинства DTO; пагинация принимает `limit`/`offset` (контракт
`page`/`page_size`), оболочка `{items, meta}` есть только у `/guests` и
`/segments`; коды ошибок бизнес-уровня почти не различаются (общий `conflict`);
`details[]` не заполняется; вложенные группы правил сегмента (`SegmentRuleGroup`)
из контракта не поддерживаются; `ruleText` не формируется.
