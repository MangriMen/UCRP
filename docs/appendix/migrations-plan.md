# План миграций goose

Источник структуры — `ddl.sql` (27 таблиц: E-1…E-21 в 26 таблицах плюс T-1 `job_queue`); обоснование и отличия СУБД — `../05-er-data-model.md`, разделы 5–6. Файлы миграций пока не пишем (п. 0.5 плана): здесь только порядок, содержимое и правила.

## 1. Каталоги и именование

| Каталог | Назначение |
|---|---|
| `migrations/common/` | Переносимый SQL, применяется к SQLite и PostgreSQL; содержимое — строго `ddl.sql` |
| `migrations/postgres/` | Только PostgreSQL, применяется после `common` |

Имя файла `NNNNN_name.sql`; в каждом файле блоки `-- +goose Up` и `-- +goose Down`. `Down` удаляет объекты в порядке, обратном зависимостям. SQLite: перед миграциями на каждое соединение `PRAGMA foreign_keys = ON`, `journal_mode = WAL`, `busy_timeout = 5000`.

## 2. Последовательность

| № | Файл | Таблицы и индексы | Зависит от | Проверка после применения |
|---|---|---|---|---|
| 1 | `00001_reference.sql` | `tenant_category`, `tenant_category_relation`, `tenant`, `project`, `event_type`, `data_source_type`, `data_source` | — | FK на категории, `priority` в 0..100 |
| 2 | `00002_users_audit.sql` | `app_user` (CHECK роль и `tenant_id`), `audit_log` | 1 | CHECK: `tenant` ⇒ `tenant_id` задан |
| 3 | `00003_guests.sql` | `guest`, `guest_identifier` | 1 | CHECK `merged` ⇒ `merged_into_id` |
| 4 | `00004_import.sql` | `import_batch` (`UNIQUE(data_source_id, file_hash)`), `rejected_row` | 1, 2 | повтор файла нарушает UNIQUE |
| 5 | `00005_events.sql` | `event` (`UNIQUE(data_source_id, event_type_id, external_id)`) и 5 индексов, в т.ч. частичный `ix_event_unmatched` | 1, 3, 4 | вставка дубликата игнорируется |
| 6 | `00006_identity.sql` | `merge_candidate`, `unmerge_block` (CHECK `guest_a_id < guest_b_id`) | 2, 3 | зеркальная пара отклоняется |
| 7 | `00007_segments.sql` | `segment`, `segment_version`, `segment_member` | 2, 3 | `UNIQUE(segment_id, version_no)` |
| 8 | `00008_campaign_reco.sql` | `campaign`, `recommendation` (`UNIQUE(guest_id, rank)`, `UNIQUE(guest_id, tenant_id)`), `recommendation_factor` | 1, 2, 3, 7 | ранг только 1..3 |
| 9 | `00009_export_contact.sql` | `export_job` (`UNIQUE(created_by, idempotency_key)`), `export_member`, `contact_limit`, `contact_log` | 3, 7, 8 | повтор ключа нарушает UNIQUE |
| 10 | `00010_job_queue.sql` | `job_queue` и частичный уникальный индекс `ux_job_dedup` | — | второй активный `dedup_key` отклоняется |
| 11 | `00011_seed_reference.sql` | Seed: типы событий (`visit`, `purchase`, `booking`), типы источников (`crm`, `pos`, `api`, `csv_excel`), 5 шаблонов сегментов, базовые категории и связи | 1, 7 | повторное применение не меняет данных |
| 12 | `postgres/00101_pg_types.sql` | `TIMESTAMP → TIMESTAMPTZ`, `TEXT → JSONB` для `rule`, `raw_payload`, `match_keys`, `config`, `field_mapping`, `details`, `payload`, `config_schema`; GIN при необходимости | 1–10 | запросы приложения не меняются |
| 13 | `postgres/00102_pg_perf.sql` | Партиционирование `event` по месяцам (при необходимости), дополнительные индексы | 12 | план запроса ленты гостя |

## 3. Правила

| Правило | Описание |
|---|---|
| Неизменяемость | Применённая миграция не правится; изменение — новая миграция |
| Переносимость | В `common` только `CREATE TABLE`, `ADD COLUMN`, индексы; без ENUM, триггеров, массивов, `SERIAL`; смена типов — только в `postgres/` |
| Seed | Идемпотентен (`INSERT … ON CONFLICT DO NOTHING`); права ролей в БД не хранятся (код, D-10) |
| Новый канал выгрузки | Миграция одной строки: расширение CHECK у `export_job.channel` и `contact_limit.channel` (D-11); в SQLite — пересоздание таблицы |
| Новая роль | Миграция CHECK у `app_user.role` (D-10, Q-16) |
| Новый тип источника или события | Только seed-запись, схема не меняется |
| Тест на обеих СУБД | CI применяет `common` на SQLite и PostgreSQL, затем загружает набор данных из `../06-object-diagrams.md` |
| Переход SQLite → PostgreSQL | Остановка воркеров, выгрузка таблиц в порядке миграций 1–10, загрузка в PG, затем применение `postgres/` и запуск; обратный путь не поддерживается |
| Очистка | `job_queue` со статусом `done` — 7 дней, файлы загрузок — 30 дней (Q-25), планировщик `platform/scheduler` |

## Чекпойнт
Зафиксированы 13 миграций (11 общих, 2 PostgreSQL). Связанные решения: D-1, D-2, D-3, D-8, D-10, D-11; вопросы: Q-11, Q-16, Q-25.
