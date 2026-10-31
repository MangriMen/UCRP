-- =====================================================================
-- Переносимый DDL (SQLite 3.35+ / PostgreSQL 14+). Сущности E-1..E-21 + T-1.
-- Правила переносимости:
--   * идентификаторы: TEXT, UUIDv7 генерирует приложение (D-1);
--   * время: TIMESTAMP, всегда UTC, задаёт приложение (в PG на этапе миграции -> TIMESTAMPTZ);
--   * деньги: BIGINT в минимальных единицах (копейки) + код валюты (D-2);
--   * JSON: TEXT (в PG отдельной миграцией -> JSONB) (D-3);
--   * булевы: BOOLEAN (SQLite хранит 0/1), TRUE/FALSE допустимы в обеих СУБД;
--   * без триггеров, процедур, ENUM, массивов, SERIAL; ограничения только CHECK/UNIQUE/FK;
--   * SQLite: PRAGMA foreign_keys = ON; journal_mode = WAL; busy_timeout = 5000.
-- Каждая таблица помечена сущностью E-n. Блоки в файлах goose см. 05-er-data-model.md, раздел 6.
-- =====================================================================

-- ---------- Справочники ----------

-- E-6 Категория арендатора
CREATE TABLE tenant_category (
  id          TEXT         PRIMARY KEY,
  code        VARCHAR(64)  NOT NULL UNIQUE,
  name        VARCHAR(200) NOT NULL,
  is_active   BOOLEAN      NOT NULL DEFAULT TRUE,
  created_at  TIMESTAMP    NOT NULL
);

-- E-6 Связи категорий: смежность и прямая конкуренция (используются скорингом)
CREATE TABLE tenant_category_relation (
  category_id          TEXT        NOT NULL REFERENCES tenant_category(id),
  related_category_id  TEXT        NOT NULL REFERENCES tenant_category(id),
  relation_type        VARCHAR(16) NOT NULL CHECK (relation_type IN ('adjacent','competitor')),
  weight_pct           SMALLINT    NOT NULL DEFAULT 100 CHECK (weight_pct BETWEEN 0 AND 100),
  PRIMARY KEY (category_id, related_category_id, relation_type),
  CHECK (category_id <> related_category_id)
);

-- E-4 Арендатор. kind: mall — арендатор комплекса, external — внешний партнёр (A-22, D-38)
CREATE TABLE tenant (
  id          TEXT         PRIMARY KEY,
  code        VARCHAR(64)  NOT NULL UNIQUE,
  name        VARCHAR(200) NOT NULL,
  category_id TEXT         NOT NULL REFERENCES tenant_category(id),
  kind        VARCHAR(16)  NOT NULL DEFAULT 'mall' CHECK (kind IN ('mall','external')),
  floor       SMALLINT     NOT NULL,
  x_m         INTEGER,                      -- координаты на плане в метрах (для расстояния), необязательно
  y_m         INTEGER,
  priority    SMALLINT     NOT NULL DEFAULT 50 CHECK (priority BETWEEN 0 AND 100),
  is_active   BOOLEAN      NOT NULL DEFAULT TRUE,
  created_at  TIMESTAMP    NOT NULL,
  updated_at  TIMESTAMP    NOT NULL,
  deleted_at  TIMESTAMP
);
CREATE INDEX ix_tenant_category ON tenant(category_id);

-- E-7 Проект (объект данных / направление бизнеса арендатора)
CREATE TABLE project (
  id          TEXT         PRIMARY KEY,
  tenant_id   TEXT         NOT NULL REFERENCES tenant(id),
  code        VARCHAR(64)  NOT NULL,
  name        VARCHAR(200) NOT NULL,
  is_active   BOOLEAN      NOT NULL DEFAULT TRUE,
  UNIQUE (tenant_id, code)
);

-- E-9 Тип события.
--   visit    — подтверждённый визит (counts_as_presence = TRUE)
--   purchase — покупка (counts_as_presence = TRUE)
--   booking  — бронь (counts_as_presence = FALSE до подтверждения)
--   top_up   — пополнение счёта, НЕ равно посещению (A-20, D-36)
--   usage    — списание/использование услуги (A-20, D-36)
--   reversal — отмена или возврат, корректирует исходное событие (A-20, D-36)
CREATE TABLE event_type (
  id                   TEXT         PRIMARY KEY,
  code                 VARCHAR(32)  NOT NULL UNIQUE,   -- visit | purchase | booking | top_up | usage | reversal | ...
  name                 VARCHAR(100) NOT NULL,
  has_amount           BOOLEAN      NOT NULL DEFAULT FALSE,
  counts_as_presence   BOOLEAN      NOT NULL DEFAULT TRUE,  -- считается ли "был у арендатора"
  reverses_event_type_id TEXT       REFERENCES event_type(id),  -- для reversal: какой тип корректируется
  is_active            BOOLEAN      NOT NULL DEFAULT TRUE
);

-- E-5 Тип источника (справочник; связан с коннектором в коде по connector_key)
CREATE TABLE data_source_type (
  id            TEXT         PRIMARY KEY,
  code          VARCHAR(32)  NOT NULL UNIQUE,          -- crm | pos | api | csv_excel | ...
  name          VARCHAR(100) NOT NULL,
  connector_key VARCHAR(64)  NOT NULL,                 -- ключ реализации SourceConnector
  supports_file BOOLEAN      NOT NULL DEFAULT TRUE,
  supports_pull BOOLEAN      NOT NULL DEFAULT FALSE,
  config_schema TEXT,                                  -- JSON Schema для data_source.config
  is_active     BOOLEAN      NOT NULL DEFAULT TRUE
);

-- E-5 Источник данных (tenant_id NULL = общий источник УК, напр. Wi-Fi, парковка)
CREATE TABLE data_source (
  id            TEXT         PRIMARY KEY,
  tenant_id     TEXT         REFERENCES tenant(id),
  type_id       TEXT         NOT NULL REFERENCES data_source_type(id),
  name          VARCHAR(200) NOT NULL,
  config        TEXT,                                  -- JSON: url, токен (ссылка на секрет), формат дат
  field_mapping TEXT,                                  -- JSON: колонка источника -> поле события
  schedule_cron VARCHAR(64),                           -- для pull
  is_active     BOOLEAN      NOT NULL DEFAULT TRUE,
  last_import_at TIMESTAMP,
  created_at    TIMESTAMP    NOT NULL,
  updated_at    TIMESTAMP    NOT NULL,
  UNIQUE (tenant_id, name)
);
CREATE INDEX ix_data_source_type ON data_source(type_id);

-- ---------- Пользователи, аудит ----------

-- E-20 Пользователь. role: marketer(R-1) manager(R-2) tenant(R-3) analyst(R-4)
CREATE TABLE app_user (
  id            TEXT         PRIMARY KEY,
  email         VARCHAR(320) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  display_name  VARCHAR(200) NOT NULL,
  role          VARCHAR(16)  NOT NULL CHECK (role IN ('marketer','manager','tenant','analyst')),
  tenant_id     TEXT         REFERENCES tenant(id),
  is_active     BOOLEAN      NOT NULL DEFAULT TRUE,
  last_login_at TIMESTAMP,
  created_at    TIMESTAMP    NOT NULL,
  CHECK ((role = 'tenant' AND tenant_id IS NOT NULL) OR (role <> 'tenant' AND tenant_id IS NULL))
);

-- E-21 Аудит (только INSERT; изменение и удаление запрещены на уровне repository)
CREATE TABLE audit_log (
  id             TEXT         PRIMARY KEY,
  actor_user_id  TEXT         REFERENCES app_user(id),   -- NULL = система (R-6)
  actor_role     VARCHAR(16)  NOT NULL,
  action         VARCHAR(64)  NOT NULL,                  -- merge.confirm, merge.revert, export.create, ...
  entity_type    VARCHAR(32)  NOT NULL,
  entity_id      VARCHAR(64)  NOT NULL,
  details        TEXT,                                   -- JSON: до/после, причина
  request_id     VARCHAR(64),
  created_at     TIMESTAMP    NOT NULL
);
CREATE INDEX ix_audit_entity ON audit_log(entity_type, entity_id, created_at);
CREATE INDEX ix_audit_actor  ON audit_log(actor_user_id, created_at);

-- ---------- Гости и идентификаторы ----------

-- E-1 Гость (soft-delete и слияние: status + merged_into_id + deleted_at)
CREATE TABLE guest (
  id             TEXT         PRIMARY KEY,
  display_name   VARCHAR(200),
  status         VARCHAR(16)  NOT NULL DEFAULT 'active' CHECK (status IN ('active','merged','deleted')),
  merged_into_id TEXT         REFERENCES guest(id),
  created_at     TIMESTAMP    NOT NULL,
  updated_at     TIMESTAMP    NOT NULL,
  deleted_at     TIMESTAMP,
  CHECK ((status = 'merged' AND merged_into_id IS NOT NULL) OR (status <> 'merged' AND merged_into_id IS NULL))
);
CREATE INDEX ix_guest_status ON guest(status);

-- E-10 Идентификатор гостя. id_type проверяется реестром identity (расширяемо без DDL: wifi_mac, car_plate...)
CREATE TABLE guest_identifier (
  id               TEXT         PRIMARY KEY,
  guest_id         TEXT         NOT NULL REFERENCES guest(id),   -- текущий владелец
  origin_guest_id  TEXT         NOT NULL REFERENCES guest(id),   -- владелец до слияния (для revert)
  id_type          VARCHAR(16)  NOT NULL,                        -- phone | email | card
  value_normalized VARCHAR(320) NOT NULL,
  data_source_id   TEXT         NOT NULL REFERENCES data_source(id),
  first_seen_at    TIMESTAMP    NOT NULL,
  UNIQUE (guest_id, id_type, value_normalized)
);
CREATE INDEX ix_gident_lookup ON guest_identifier(id_type, value_normalized);
CREATE INDEX ix_gident_origin ON guest_identifier(origin_guest_id);

-- ---------- Импорт ----------

-- E-13 Загрузка (файл или pull-окно)
CREATE TABLE import_batch (
  id              TEXT         PRIMARY KEY,
  data_source_id  TEXT         NOT NULL REFERENCES data_source(id),
  mode            VARCHAR(8)   NOT NULL CHECK (mode IN ('file','pull')),
  file_name       VARCHAR(255),
  file_hash       CHAR(64),                                   -- sha256; NULL для pull
  file_size_bytes BIGINT,
  window_from     TIMESTAMP,                                  -- для pull
  window_to       TIMESTAMP,
  status          VARCHAR(24)  NOT NULL CHECK (status IN ('queued','running','completed','completed_with_errors','failed')),
  total_rows      INTEGER      NOT NULL DEFAULT 0,
  accepted_rows   INTEGER      NOT NULL DEFAULT 0,
  rejected_rows   INTEGER      NOT NULL DEFAULT 0,
  duplicate_rows  INTEGER      NOT NULL DEFAULT 0,            -- accepted + rejected + duplicate = total
  error_message   VARCHAR(1000),
  created_by      TEXT         REFERENCES app_user(id),       -- NULL = расписание/система
  created_at      TIMESTAMP    NOT NULL,
  started_at      TIMESTAMP,
  finished_at     TIMESTAMP,
  UNIQUE (data_source_id, file_hash)
);
CREATE INDEX ix_batch_source ON import_batch(data_source_id, created_at);

-- E-14 Отклонённая строка
CREATE TABLE rejected_row (
  id              TEXT         PRIMARY KEY,
  import_batch_id TEXT         NOT NULL REFERENCES import_batch(id) ON DELETE CASCADE,
  row_number      INTEGER      NOT NULL,
  error_code      VARCHAR(64)  NOT NULL,                      -- bad_date, no_external_id, bad_amount, ...
  error_message   VARCHAR(500) NOT NULL,
  raw_line        TEXT,
  created_at      TIMESTAMP    NOT NULL,
  UNIQUE (import_batch_id, row_number)
);

-- ---------- События ----------

-- E-2 Событие. guest_id NULL = не привязано (учитывается в match rate); match_keys для повторной привязки
CREATE TABLE event (
  id              TEXT         PRIMARY KEY,
  guest_id        TEXT         REFERENCES guest(id),
  origin_guest_id TEXT         REFERENCES guest(id),
  tenant_id       TEXT         NOT NULL REFERENCES tenant(id),
  project_id      TEXT         REFERENCES project(id),
  event_type_id   TEXT         NOT NULL REFERENCES event_type(id),
  data_source_id  TEXT         NOT NULL REFERENCES data_source(id),
  import_batch_id TEXT         REFERENCES import_batch(id),
  external_id     VARCHAR(128) NOT NULL,
  occurred_at     TIMESTAMP    NOT NULL,
  amount_minor    BIGINT       CHECK (amount_minor IS NULL OR amount_minor >= 0),
  currency        CHAR(3)      NOT NULL DEFAULT 'RUB',
  reverses_event_id TEXT       REFERENCES event(id),          -- для события типа reversal: какое событие отменяется (A-20)
  match_keys      TEXT,                                       -- JSON: нормализованные ключи из строки
  raw_payload     TEXT,                                       -- JSON: исходные атрибуты источника
  created_at      TIMESTAMP    NOT NULL,
  UNIQUE (data_source_id, event_type_id, external_id)
);
CREATE INDEX ix_event_guest_time   ON event(guest_id, occurred_at);
CREATE INDEX ix_event_tenant_time  ON event(tenant_id, occurred_at);
CREATE INDEX ix_event_guest_tenant ON event(guest_id, tenant_id);
CREATE INDEX ix_event_batch        ON event(import_batch_id);
CREATE INDEX ix_event_unmatched    ON event(data_source_id) WHERE guest_id IS NULL;

-- ---------- Склейка ----------

-- E-11 Кандидат на склейку и журнал решений. guest_a_id < guest_b_id (канонический порядок пары)
CREATE TABLE merge_candidate (
  id               TEXT         PRIMARY KEY,
  guest_a_id       TEXT         NOT NULL REFERENCES guest(id),
  guest_b_id       TEXT         NOT NULL REFERENCES guest(id),
  match_type       VARCHAR(16)  NOT NULL,                     -- phone | email | card | name
  reason_code      VARCHAR(48)  NOT NULL,                     -- shared_phone_names_differ, ...
  reason_text      VARCHAR(500) NOT NULL,
  score            SMALLINT     NOT NULL CHECK (score BETWEEN 0 AND 100),
  resolution       VARCHAR(16)  NOT NULL CHECK (resolution IN ('pending','auto_merged','confirmed','rejected','reverted')),
  winner_guest_id  TEXT         REFERENCES guest(id),
  merged_guest_id  TEXT         REFERENCES guest(id),
  decided_by       TEXT         REFERENCES app_user(id),      -- NULL при auto_merged
  decided_at       TIMESTAMP,
  created_at       TIMESTAMP    NOT NULL,
  UNIQUE (guest_a_id, guest_b_id),
  CHECK (guest_a_id < guest_b_id)
);
CREATE INDEX ix_merge_resolution ON merge_candidate(resolution, created_at);

-- E-12 Запрет склейки
CREATE TABLE unmerge_block (
  id           TEXT         PRIMARY KEY,
  guest_a_id   TEXT         NOT NULL REFERENCES guest(id),
  guest_b_id   TEXT         NOT NULL REFERENCES guest(id),
  reason       VARCHAR(24)  NOT NULL CHECK (reason IN ('manual_unmerge','rejected')),
  candidate_id TEXT         REFERENCES merge_candidate(id),
  created_by   TEXT         NOT NULL REFERENCES app_user(id),
  created_at   TIMESTAMP    NOT NULL,
  UNIQUE (guest_a_id, guest_b_id),
  CHECK (guest_a_id < guest_b_id)
);

-- ---------- Сегменты ----------

-- E-3 Сегмент (шаблоны: is_template = TRUE, template_code из реестра)
CREATE TABLE segment (
  id               TEXT          PRIMARY KEY,
  name             VARCHAR(200)  NOT NULL,
  description      VARCHAR(1000),
  template_code    VARCHAR(64),                                -- sleeping_neighbor | single_store | churn_risk | newcomer | high_value
  is_template      BOOLEAN       NOT NULL DEFAULT FALSE,
  current_version_no INTEGER     NOT NULL DEFAULT 1,
  status           VARCHAR(16)   NOT NULL DEFAULT 'active' CHECK (status IN ('active','archived')),
  created_by       TEXT          REFERENCES app_user(id),
  created_at       TIMESTAMP     NOT NULL,
  updated_at       TIMESTAMP     NOT NULL,
  deleted_at       TIMESTAMP
);
CREATE INDEX ix_segment_status ON segment(status);

-- E-3 Версия правила сегмента (неизменяемая)
CREATE TABLE segment_version (
  id           TEXT        PRIMARY KEY,
  segment_id   TEXT        NOT NULL REFERENCES segment(id),
  version_no   INTEGER     NOT NULL,
  rule         TEXT        NOT NULL,                          -- JSON DSL: was[], was_not[], period_days, require_contact, ...
  status       VARCHAR(16) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','computing','ready','failed')),
  member_count INTEGER,                                       -- кэш размера
  computed_at  TIMESTAMP,
  created_by   TEXT        REFERENCES app_user(id),
  created_at   TIMESTAMP   NOT NULL,
  UNIQUE (segment_id, version_no)
);

-- E-3 Материализованный состав версии
CREATE TABLE segment_member (
  segment_version_id TEXT      NOT NULL REFERENCES segment_version(id) ON DELETE CASCADE,
  guest_id           TEXT      NOT NULL REFERENCES guest(id),
  added_at           TIMESTAMP NOT NULL,
  PRIMARY KEY (segment_version_id, guest_id)
);
CREATE INDEX ix_segmember_guest ON segment_member(guest_id);

-- ---------- Кампании и рекомендации ----------

-- E-8 Кампания
CREATE TABLE campaign (
  id                TEXT          PRIMARY KEY,
  name              VARCHAR(200)  NOT NULL,
  description       VARCHAR(1000),
  target_tenant_id  TEXT          REFERENCES tenant(id),     -- куда приглашаем
  segment_id        TEXT          REFERENCES segment(id),
  attribution_days  SMALLINT      NOT NULL DEFAULT 14 CHECK (attribution_days BETWEEN 1 AND 90),
  starts_at         TIMESTAMP,
  ends_at           TIMESTAMP,
  status            VARCHAR(16)   NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','active','finished')),
  created_by        TEXT          NOT NULL REFERENCES app_user(id),
  created_at        TIMESTAMP     NOT NULL
);

-- E-15 Рекомендация (текущий набор гостя; пересчёт заменяет набор транзакцией)
CREATE TABLE recommendation (
  id                TEXT           PRIMARY KEY,
  guest_id          TEXT           NOT NULL REFERENCES guest(id),
  tenant_id         TEXT           NOT NULL REFERENCES tenant(id),
  rank              SMALLINT       NOT NULL CHECK (rank BETWEEN 1 AND 3),
  score             NUMERIC(7,4)   NOT NULL,
  explanation       VARCHAR(1000)  NOT NULL,                 -- человекочитаемое обоснование (снимок)
  algorithm_version VARCHAR(32)    NOT NULL,
  computed_at       TIMESTAMP      NOT NULL,
  UNIQUE (guest_id, rank),
  UNIQUE (guest_id, tenant_id)
);
CREATE INDEX ix_reco_tenant ON recommendation(tenant_id);

-- E-15 Факторы рекомендации (прозрачность скоринга)
CREATE TABLE recommendation_factor (
  recommendation_id TEXT          NOT NULL REFERENCES recommendation(id) ON DELETE CASCADE,
  factor_code       VARCHAR(32)   NOT NULL,                  -- similar_guests | category_adjacency | distance | recency | tenant_priority
  raw_value         NUMERIC(8,4)  NOT NULL,
  weight            NUMERIC(6,3)  NOT NULL,
  contribution      NUMERIC(8,4)  NOT NULL,
  PRIMARY KEY (recommendation_id, factor_code)
);

-- ---------- Выгрузка и контакты ----------

-- E-16 Выгрузка (журнал)
CREATE TABLE export_job (
  id                 TEXT         PRIMARY KEY,
  segment_id         TEXT         NOT NULL REFERENCES segment(id),
  segment_version_id TEXT         NOT NULL REFERENCES segment_version(id),
  campaign_id        TEXT         REFERENCES campaign(id),
  channel            VARCHAR(16)  NOT NULL CHECK (channel IN ('sms','push','email')),
  format             VARCHAR(8)   NOT NULL CHECK (format IN ('csv','api')),
  control_percent    SMALLINT     NOT NULL DEFAULT 0 CHECK (control_percent BETWEEN 0 AND 50),
  control_seed       BIGINT       NOT NULL,
  idempotency_key    VARCHAR(64)  NOT NULL,
  status             VARCHAR(16)  NOT NULL CHECK (status IN ('queued','building','ready','failed')),
  total_in_segment   INTEGER      NOT NULL DEFAULT 0,
  no_contact_count   INTEGER      NOT NULL DEFAULT 0,
  blocked_by_limit_count INTEGER  NOT NULL DEFAULT 0,
  control_count      INTEGER      NOT NULL DEFAULT 0,
  exported_count     INTEGER      NOT NULL DEFAULT 0,
  created_by         TEXT         NOT NULL REFERENCES app_user(id),
  created_at         TIMESTAMP    NOT NULL,
  finished_at        TIMESTAMP,
  UNIQUE (created_by, idempotency_key)
);
CREATE INDEX ix_export_segment ON export_job(segment_id, created_at);

-- E-17 Член выгрузки (получатель или контрольная группа)
CREATE TABLE export_member (
  export_id     TEXT      NOT NULL REFERENCES export_job(id) ON DELETE CASCADE,
  guest_id      TEXT      NOT NULL REFERENCES guest(id),
  identifier_id TEXT      NOT NULL REFERENCES guest_identifier(id),  -- какой контакт использован
  is_control    BOOLEAN   NOT NULL DEFAULT FALSE,
  PRIMARY KEY (export_id, guest_id)
);
CREATE INDEX ix_exmember_guest ON export_member(guest_id);

-- E-18 Лимит контактов (channel NULL = все каналы)
CREATE TABLE contact_limit (
  id            TEXT        PRIMARY KEY,
  channel       VARCHAR(16) CHECK (channel IS NULL OR channel IN ('sms','push','email')),
  max_contacts  SMALLINT    NOT NULL CHECK (max_contacts >= 1),
  period_days   SMALLINT    NOT NULL CHECK (period_days >= 1),
  is_active     BOOLEAN     NOT NULL DEFAULT TRUE,
  created_by    TEXT        NOT NULL REFERENCES app_user(id),
  created_at    TIMESTAMP   NOT NULL,
  updated_at    TIMESTAMP   NOT NULL
);

-- E-19 Контакт (только получатели, не контрольная группа)
CREATE TABLE contact_log (
  id            TEXT        PRIMARY KEY,
  guest_id      TEXT        NOT NULL REFERENCES guest(id),
  export_id     TEXT        NOT NULL REFERENCES export_job(id),
  contacted_at  TIMESTAMP   NOT NULL,
  UNIQUE (export_id, guest_id)
);
CREATE INDEX ix_contact_guest_time ON contact_log(guest_id, contacted_at);

-- ---------- Служебное (T-1, не E-*) ----------

-- T-1 Персистентная очередь заданий воркеров (D-8)
CREATE TABLE job_queue (
  id          TEXT         PRIMARY KEY,
  job_type    VARCHAR(48)  NOT NULL,
  payload     TEXT         NOT NULL,                              -- JSON
  dedup_key   VARCHAR(128),
  status      VARCHAR(12)  NOT NULL CHECK (status IN ('queued','running','done','failed')),
  attempts    SMALLINT     NOT NULL DEFAULT 0,
  run_after   TIMESTAMP    NOT NULL,
  locked_by   VARCHAR(64),
  locked_at   TIMESTAMP,
  last_error  VARCHAR(1000),
  created_at  TIMESTAMP    NOT NULL
);
-- схлопывание активных дублей: один активный dedup_key
CREATE UNIQUE INDEX ux_job_dedup ON job_queue(dedup_key) WHERE status IN ('queued','running');
CREATE INDEX ix_job_pick ON job_queue(status, run_after);
