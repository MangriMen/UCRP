-- +goose Up

CREATE TABLE tenant_category (
  id          TEXT         PRIMARY KEY,
  code        VARCHAR(64)  NOT NULL UNIQUE,
  name        VARCHAR(200) NOT NULL,
  is_active   BOOLEAN      NOT NULL DEFAULT TRUE,
  created_at  TIMESTAMP    NOT NULL
);

CREATE TABLE tenant_category_relation (
  category_id          TEXT        NOT NULL REFERENCES tenant_category(id),
  related_category_id  TEXT        NOT NULL REFERENCES tenant_category(id),
  relation_type        VARCHAR(16) NOT NULL CHECK (relation_type IN ('adjacent','competitor')),
  weight_pct           SMALLINT    NOT NULL DEFAULT 100 CHECK (weight_pct BETWEEN 0 AND 100),
  PRIMARY KEY (category_id, related_category_id, relation_type),
  CHECK (category_id <> related_category_id)
);

CREATE TABLE mall (
  id         TEXT         PRIMARY KEY,
  code       VARCHAR(64)  NOT NULL UNIQUE,
  name       VARCHAR(200) NOT NULL,
  is_active  BOOLEAN      NOT NULL DEFAULT TRUE,
  floor_min  SMALLINT     NOT NULL DEFAULT 1,
  floor_max  SMALLINT     NOT NULL DEFAULT 1,
  has_zero_floor BOOLEAN  NOT NULL DEFAULT FALSE,
  created_at TIMESTAMP    NOT NULL,
  CHECK (floor_min <= floor_max)
);

INSERT INTO mall (id, code, name, is_active, created_at)
VALUES ('00000000-0000-0000-0000-000000000001', 'main', 'Основной ТЦ', TRUE, now() AT TIME ZONE 'utc');

CREATE TABLE tenant (
  id          TEXT         PRIMARY KEY,
  code        VARCHAR(64)  NOT NULL UNIQUE,
  name        VARCHAR(200) NOT NULL,
  category_id TEXT         NOT NULL REFERENCES tenant_category(id),
  kind        VARCHAR(16)  NOT NULL DEFAULT 'mall' CHECK (kind IN ('mall','external')),
  floor       SMALLINT,
  x_m         INTEGER,
  y_m         INTEGER,
  priority    SMALLINT     NOT NULL DEFAULT 50 CHECK (priority BETWEEN 0 AND 100),
  is_active   BOOLEAN      NOT NULL DEFAULT TRUE,
  created_at  TIMESTAMP    NOT NULL,
  updated_at  TIMESTAMP    NOT NULL,
  deleted_at  TIMESTAMP,
  mall_id     TEXT         NOT NULL REFERENCES mall(id)
);
CREATE INDEX ix_tenant_category ON tenant(category_id);
CREATE INDEX ix_tenant_mall_floor ON tenant(mall_id, floor);
CREATE UNIQUE INDEX ux_tenant_code_lower ON tenant (lower(code));

CREATE TABLE project (
  id          TEXT         PRIMARY KEY,
  tenant_id   TEXT         NOT NULL REFERENCES tenant(id),
  code        VARCHAR(64)  NOT NULL,
  name        VARCHAR(200) NOT NULL,
  is_active   BOOLEAN      NOT NULL DEFAULT TRUE,
  UNIQUE (tenant_id, code)
);

CREATE TABLE event_type (
  id                   TEXT         PRIMARY KEY,
  code                 VARCHAR(32)  NOT NULL UNIQUE,
  name                 VARCHAR(100) NOT NULL,
  has_amount           BOOLEAN      NOT NULL DEFAULT FALSE,
  counts_as_presence   BOOLEAN      NOT NULL DEFAULT TRUE,
  reverses_event_type_id TEXT       REFERENCES event_type(id),
  is_active            BOOLEAN      NOT NULL DEFAULT TRUE
);

CREATE TABLE data_source_type (
  id            TEXT         PRIMARY KEY,
  code          VARCHAR(32)  NOT NULL UNIQUE,
  name          VARCHAR(100) NOT NULL,
  connector_key VARCHAR(64)  NOT NULL,
  supports_file BOOLEAN      NOT NULL DEFAULT TRUE,
  supports_pull BOOLEAN      NOT NULL DEFAULT FALSE,
  config_schema TEXT,
  is_active     BOOLEAN      NOT NULL DEFAULT TRUE
);

CREATE TABLE data_source (
  id            TEXT         PRIMARY KEY,
  tenant_id     TEXT         REFERENCES tenant(id),
  type_id       TEXT         NOT NULL REFERENCES data_source_type(id),
  name          VARCHAR(200) NOT NULL,
  config        TEXT,
  field_mapping TEXT,
  schedule_cron VARCHAR(64),
  is_active     BOOLEAN      NOT NULL DEFAULT TRUE,
  last_import_at TIMESTAMP,
  created_at    TIMESTAMP    NOT NULL,
  updated_at    TIMESTAMP    NOT NULL,
  UNIQUE (tenant_id, name)
);
CREATE INDEX ix_data_source_type ON data_source(type_id);

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
  auth_version  BIGINT       NOT NULL DEFAULT 0,
  updated_at    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CHECK ((role = 'tenant' AND tenant_id IS NOT NULL) OR (role <> 'tenant' AND tenant_id IS NULL))
);

CREATE TABLE audit_log (
  id             TEXT         PRIMARY KEY,
  actor_user_id  TEXT         REFERENCES app_user(id),
  actor_role     VARCHAR(16)  NOT NULL,
  action         VARCHAR(64)  NOT NULL,
  entity_type    VARCHAR(32)  NOT NULL,
  entity_id      VARCHAR(64)  NOT NULL,
  details        TEXT,
  request_id     VARCHAR(64),
  created_at     TIMESTAMP    NOT NULL
);
CREATE INDEX ix_audit_entity ON audit_log(entity_type, entity_id, created_at);
CREATE INDEX ix_audit_actor  ON audit_log(actor_user_id, created_at);

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

CREATE TABLE guest_identifier (
  id               TEXT         PRIMARY KEY,
  guest_id         TEXT         NOT NULL REFERENCES guest(id),
  origin_guest_id  TEXT         NOT NULL REFERENCES guest(id),
  id_type          VARCHAR(16)  NOT NULL,
  value_normalized VARCHAR(320) NOT NULL,
  data_source_id   TEXT         NOT NULL REFERENCES data_source(id),
  first_seen_at    TIMESTAMP    NOT NULL,
  UNIQUE (guest_id, id_type, value_normalized)
);
CREATE INDEX ix_gident_lookup ON guest_identifier(id_type, value_normalized);
CREATE INDEX ix_gident_origin ON guest_identifier(origin_guest_id);

CREATE TABLE import_batch (
  id              TEXT         PRIMARY KEY,
  data_source_id  TEXT         NOT NULL REFERENCES data_source(id),
  mode            VARCHAR(8)   NOT NULL CHECK (mode IN ('file','pull')),
  file_name       VARCHAR(255),
  file_hash       CHAR(64),
  file_size_bytes BIGINT,
  window_from     TIMESTAMP,
  window_to       TIMESTAMP,
  status          VARCHAR(24)  NOT NULL CHECK (status IN ('queued','running','completed','completed_with_errors','failed')),
  total_rows      INTEGER      NOT NULL DEFAULT 0,
  accepted_rows   INTEGER      NOT NULL DEFAULT 0,
  rejected_rows   INTEGER      NOT NULL DEFAULT 0,
  duplicate_rows  INTEGER      NOT NULL DEFAULT 0,
  error_message   VARCHAR(1000),
  created_by      TEXT         REFERENCES app_user(id),
  created_at      TIMESTAMP    NOT NULL,
  started_at      TIMESTAMP,
  finished_at     TIMESTAMP,
  UNIQUE (data_source_id, file_hash)
);
CREATE INDEX ix_batch_source ON import_batch(data_source_id, created_at);

CREATE TABLE rejected_row (
  id              TEXT         PRIMARY KEY,
  import_batch_id TEXT         NOT NULL REFERENCES import_batch(id) ON DELETE CASCADE,
  row_number      INTEGER      NOT NULL,
  error_code      VARCHAR(64)  NOT NULL,
  error_message   VARCHAR(500) NOT NULL,
  raw_line        TEXT,
  created_at      TIMESTAMP    NOT NULL,
  UNIQUE (import_batch_id, row_number)
);

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
  reverses_event_id TEXT       REFERENCES event(id),
  match_keys      TEXT,
  raw_payload     TEXT,
  created_at      TIMESTAMP    NOT NULL,
  UNIQUE (data_source_id, event_type_id, external_id)
);
CREATE INDEX ix_event_guest_time   ON event(guest_id, occurred_at);
CREATE INDEX ix_event_tenant_time  ON event(tenant_id, occurred_at);
CREATE INDEX ix_event_guest_tenant ON event(guest_id, tenant_id);
CREATE INDEX ix_event_batch        ON event(import_batch_id);
CREATE INDEX ix_event_unmatched    ON event(data_source_id) WHERE guest_id IS NULL;

CREATE TABLE merge_candidate (
  id               TEXT         PRIMARY KEY,
  guest_a_id       TEXT         NOT NULL REFERENCES guest(id),
  guest_b_id       TEXT         NOT NULL REFERENCES guest(id),
  match_type       VARCHAR(16)  NOT NULL,
  reason_code      VARCHAR(48)  NOT NULL,
  reason_text      VARCHAR(500) NOT NULL,
  score            SMALLINT     NOT NULL CHECK (score BETWEEN 0 AND 100),
  resolution       VARCHAR(16)  NOT NULL CHECK (resolution IN ('pending','auto_merged','confirmed','rejected','reverted')),
  winner_guest_id  TEXT         REFERENCES guest(id),
  merged_guest_id  TEXT         REFERENCES guest(id),
  decided_by       TEXT         REFERENCES app_user(id),
  decided_at       TIMESTAMP,
  created_at       TIMESTAMP    NOT NULL,
  UNIQUE (guest_a_id, guest_b_id),
  CHECK (guest_a_id < guest_b_id)
);
CREATE INDEX ix_merge_resolution ON merge_candidate(resolution, created_at);

CREATE TABLE merge_block (
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

CREATE TABLE segment (
  id               TEXT          PRIMARY KEY,
  name             VARCHAR(200)  NOT NULL,
  description      VARCHAR(1000),
  template_code    VARCHAR(64),
  is_template      BOOLEAN       NOT NULL DEFAULT FALSE,
  current_version_no INTEGER     NOT NULL DEFAULT 1,
  status           VARCHAR(16)   NOT NULL DEFAULT 'active' CHECK (status IN ('active','archived')),
  created_by       TEXT          REFERENCES app_user(id),
  created_at       TIMESTAMP     NOT NULL,
  updated_at       TIMESTAMP     NOT NULL,
  deleted_at       TIMESTAMP
);
CREATE INDEX ix_segment_status ON segment(status);

CREATE TABLE segment_version (
  id           TEXT        PRIMARY KEY,
  segment_id   TEXT        NOT NULL REFERENCES segment(id),
  version_no   INTEGER     NOT NULL,
  rule         TEXT        NOT NULL,
  status       VARCHAR(16) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','computing','ready','failed')),
  member_count INTEGER,
  computed_at  TIMESTAMP,
  created_by   TEXT        REFERENCES app_user(id),
  created_at   TIMESTAMP   NOT NULL,
  UNIQUE (segment_id, version_no)
);

CREATE TABLE segment_member (
  segment_version_id TEXT      NOT NULL REFERENCES segment_version(id) ON DELETE CASCADE,
  guest_id           TEXT      NOT NULL REFERENCES guest(id),
  added_at           TIMESTAMP NOT NULL,
  PRIMARY KEY (segment_version_id, guest_id)
);
CREATE INDEX ix_segmember_guest ON segment_member(guest_id);

CREATE TABLE campaign (
  id                TEXT          PRIMARY KEY,
  name              VARCHAR(200)  NOT NULL,
  description       VARCHAR(1000),
  target_tenant_id  TEXT          REFERENCES tenant(id),
  segment_id        TEXT          REFERENCES segment(id),
  attribution_days  SMALLINT      NOT NULL DEFAULT 14 CHECK (attribution_days BETWEEN 1 AND 90),
  starts_at         TIMESTAMP,
  ends_at           TIMESTAMP,
  status            VARCHAR(16)   NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','active','finished')),
  created_by        TEXT          NOT NULL REFERENCES app_user(id),
  created_at        TIMESTAMP     NOT NULL
);

CREATE TABLE recommendation (
  id                TEXT           PRIMARY KEY,
  guest_id          TEXT           NOT NULL REFERENCES guest(id),
  tenant_id         TEXT           NOT NULL REFERENCES tenant(id),
  rank              SMALLINT       NOT NULL CHECK (rank BETWEEN 1 AND 3),
  score             NUMERIC(7,4)   NOT NULL,
  explanation       VARCHAR(1000)  NOT NULL,
  algorithm_version VARCHAR(32)    NOT NULL,
  computed_at       TIMESTAMP      NOT NULL,
  UNIQUE (guest_id, rank),
  UNIQUE (guest_id, tenant_id)
);
CREATE INDEX ix_reco_tenant ON recommendation(tenant_id);

CREATE TABLE recommendation_factor (
  recommendation_id TEXT          NOT NULL REFERENCES recommendation(id) ON DELETE CASCADE,
  factor_code       VARCHAR(32)   NOT NULL,
  raw_value         NUMERIC(8,4)  NOT NULL,
  weight            NUMERIC(6,3)  NOT NULL,
  contribution      NUMERIC(8,4)  NOT NULL,
  PRIMARY KEY (recommendation_id, factor_code)
);

CREATE TABLE recommendation_eval (
  id                        TEXT         PRIMARY KEY,
  guest_id                  TEXT         NOT NULL REFERENCES guest(id),
  probabilities             TEXT         NOT NULL,
  ranking                   TEXT         NOT NULL,
  confidence                NUMERIC(6,5),
  recommendation_computed_at TIMESTAMP   NOT NULL,
  evaluated_at              TIMESTAMP    NOT NULL
);
CREATE INDEX ix_reco_eval_guest ON recommendation_eval(guest_id, evaluated_at);

CREATE TABLE export_job (
  id                 TEXT         PRIMARY KEY,
  segment_id         TEXT         NOT NULL REFERENCES segment(id),
  segment_version_id TEXT         NOT NULL REFERENCES segment_version(id),
  campaign_id        TEXT         REFERENCES campaign(id),
  channel            VARCHAR(16)  NOT NULL CHECK (channel IN ('sms','push','email')),
  format             VARCHAR(8)   NOT NULL CHECK (format IN ('csv','api','xlsx')),
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

CREATE TABLE export_member (
  export_id     TEXT      NOT NULL REFERENCES export_job(id) ON DELETE CASCADE,
  guest_id      TEXT      NOT NULL REFERENCES guest(id),
  identifier_id TEXT      NOT NULL REFERENCES guest_identifier(id),
  is_control    BOOLEAN   NOT NULL DEFAULT FALSE,
  PRIMARY KEY (export_id, guest_id)
);
CREATE INDEX ix_exmember_guest ON export_member(guest_id);

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

CREATE TABLE contact_log (
  id            TEXT        PRIMARY KEY,
  guest_id      TEXT        NOT NULL REFERENCES guest(id),
  export_id     TEXT        NOT NULL REFERENCES export_job(id),
  contacted_at  TIMESTAMP   NOT NULL,
  UNIQUE (export_id, guest_id)
);
CREATE INDEX ix_contact_guest_time ON contact_log(guest_id, contacted_at);

CREATE TABLE job_queue (
  id          TEXT         PRIMARY KEY,
  job_type    VARCHAR(48)  NOT NULL,
  payload     TEXT         NOT NULL,
  dedup_key   VARCHAR(128),
  status      VARCHAR(12)  NOT NULL CHECK (status IN ('queued','running','done','failed')),
  attempts    SMALLINT     NOT NULL DEFAULT 0,
  run_after   TIMESTAMP    NOT NULL,
  locked_by   VARCHAR(64),
  locked_at   TIMESTAMP,
  last_error  VARCHAR(1000),
  created_at  TIMESTAMP    NOT NULL
);
CREATE UNIQUE INDEX ux_job_dedup ON job_queue(dedup_key) WHERE status IN ('queued','running');
CREATE INDEX ix_job_pick ON job_queue(status, run_after);

CREATE TABLE user_invitation (
  id TEXT PRIMARY KEY,
  email VARCHAR(320) NOT NULL,
  display_name VARCHAR(200) NOT NULL,
  role VARCHAR(16) NOT NULL CHECK (role IN ('manager', 'tenant')),
  purpose VARCHAR(16) NOT NULL CHECK (purpose IN ('invite', 'recovery')),
  target_user_id TEXT REFERENCES app_user(id),
  target_auth_version BIGINT,
  tenant_id TEXT REFERENCES tenant(id),
  token_hash BYTEA NOT NULL UNIQUE,
  created_by TEXT NOT NULL REFERENCES app_user(id),
  policy_version BIGINT NOT NULL,
  created_at TIMESTAMP NOT NULL,
  expires_at TIMESTAMP NOT NULL,
  consumed_at TIMESTAMP,
  revoked_at TIMESTAMP,
  CHECK (expires_at > created_at),
  CHECK ((role = 'tenant' AND tenant_id IS NOT NULL) OR (role = 'manager' AND tenant_id IS NULL)),
  CHECK ((purpose = 'invite' AND target_user_id IS NULL AND target_auth_version IS NULL)
      OR (purpose = 'recovery' AND target_user_id IS NOT NULL AND target_auth_version IS NOT NULL))
);
CREATE INDEX ix_user_invitation_email ON user_invitation(email);

CREATE TABLE permission (
  code          VARCHAR(64) PRIMARY KEY,
  domain        VARCHAR(32) NOT NULL,
  kind          VARCHAR(16) NOT NULL CHECK (kind IN ('action', 'metric', 'field')),
  title         VARCHAR(200) NOT NULL,
  description   TEXT,
  is_sensitive  BOOLEAN NOT NULL DEFAULT FALSE,
  min_k         SMALLINT CHECK (min_k IS NULL OR min_k >= 0),
  is_active     BOOLEAN NOT NULL DEFAULT TRUE
);
CREATE INDEX ix_permission_domain ON permission(domain, kind);

INSERT INTO permission (code, domain, kind, title, is_sensitive, min_k) VALUES
  ('guest.profile.read', 'guest', 'action', 'Профиль гостя', TRUE, NULL),
  ('guest.profile.search', 'guest', 'action', 'Поиск гостей', TRUE, NULL),
  ('guest.profile.delete', 'guest', 'action', 'Soft-delete гостя', TRUE, NULL),
  ('guest.event.read', 'guest', 'action', 'Лента событий', TRUE, NULL),
  ('guest.identifier.read', 'guest', 'action', 'Идентификаторы (телефон, email, карта)', TRUE, NULL),
  ('guest.contact_history.read', 'guest', 'action', 'История контактов', TRUE, NULL),
  ('segment.read', 'segment', 'action', 'Список и карточка сегмента', FALSE, NULL),
  ('segment.preview', 'segment', 'action', 'Предпросмотр правила', FALSE, NULL),
  ('segment.write', 'segment', 'action', 'Создать/изменить сегмент', FALSE, NULL),
  ('segment.delete', 'segment', 'action', 'Архивировать сегмент', FALSE, NULL),
  ('segment.members.read', 'segment', 'action', 'Состав сегмента (строки гостей)', TRUE, NULL),
  ('recommendation.read', 'recommendation', 'action', 'Рекомендации гостя', TRUE, NULL),
  ('recommendation.refresh', 'recommendation', 'action', 'Пересчёт рекомендаций', FALSE, NULL),
  ('identity.queue.read', 'identity', 'action', 'Очередь спорных объединений', TRUE, NULL),
  ('identity.decide', 'identity', 'action', 'Confirm / reject / revert', TRUE, NULL),
  ('import.read', 'import', 'action', 'Журнал загрузок', FALSE, NULL),
  ('import.create', 'import', 'action', 'Загрузка файла, ручная синхронизация', FALSE, NULL),
  ('import.rejected_rows.read', 'import', 'action', 'Отклонённые строки', FALSE, NULL),
  ('data_source.read', 'data_source', 'action', 'Источники данных', FALSE, NULL),
  ('data_source.write', 'data_source', 'action', 'Создать/изменить источник', FALSE, NULL),
  ('reference.read', 'reference', 'action', 'Справочники (арендаторы, категории, проекты, типы)', FALSE, NULL),
  ('reference.write', 'reference', 'action', 'Изменение справочников', FALSE, NULL),
  ('export.read', 'export', 'action', 'Журнал выгрузок', FALSE, NULL),
  ('export.create', 'export', 'action', 'Создать выгрузку', TRUE, NULL),
  ('export.download', 'export', 'action', 'Скачать CSV получателей', TRUE, NULL),
  ('export.members.read', 'export', 'action', 'Получатели и контрольная группа', TRUE, NULL),
  ('campaign.read', 'campaign', 'action', 'Кампании', FALSE, NULL),
  ('campaign.write', 'campaign', 'action', 'Создать/изменить кампанию', FALSE, NULL),
  ('campaign.results.read', 'campaign', 'action', 'Результаты кампании', FALSE, NULL),
  ('contact_limit.read', 'contact_limit', 'action', 'Правила лимита контактов', FALSE, NULL),
  ('contact_limit.write', 'contact_limit', 'action', 'Изменить лимит', FALSE, NULL),
  ('user.read', 'user', 'action', 'Пользователи', FALSE, NULL),
  ('user.write', 'user', 'action', 'Создать/изменить/деактивировать', TRUE, NULL),
  ('audit.read', 'audit', 'action', 'Журнал аудита', TRUE, NULL),
  ('access.grant.read', 'access', 'action', 'Просмотр грантов', FALSE, NULL),
  ('access.grant.write', 'access', 'action', 'Выдача и отзыв грантов', TRUE, NULL),
  ('access.simulate', 'access', 'action', '«Что увидит арендатор» без выдачи', FALSE, NULL),
  ('data_quality.read', 'data_quality', 'action', 'Качество источников', FALSE, NULL),
  ('dashboard.read', 'dashboard', 'action', 'Дашборд (каркас, без метрик)', FALSE, NULL),
  ('tenant.aggregates.read', 'tenant', 'action', 'Агрегаты любого арендатора', FALSE, NULL),
  ('tenant.aggregates.own', 'tenant', 'action', 'Агрегаты только своего арендатора', FALSE, NULL),
  ('metric.first_purchase_new_tenant.read', 'metric', 'metric', 'Первая покупка у нового арендатора (главная, A-17)', TRUE, 5),
  ('metric.cross_tenant_distribution.read', 'metric', 'metric', 'Cross-tenant: распределение гостей по числу арендаторов (1 / 2 / 3+)', TRUE, 5),
  ('metric.cross_tenant_rate.read', 'metric', 'metric', 'Доля гостей с ≥ 2 арендаторами', TRUE, 5),
  ('metric.match_rate.read', 'metric', 'metric', 'Match rate (≥ 70%)', FALSE, 5),
  ('metric.disputed_merge_rate.read', 'metric', 'metric', 'Доля оспоренных объединений (< 1%)', FALSE, 5),
  ('metric.segment_to_export_time.read', 'metric', 'metric', 'Время «сегмент → выгрузка» (медиана, p95)', FALSE, 5),
  ('metric.conversion_rate.read', 'metric', 'metric', 'Конверсия рассылки (≥ 3%)', TRUE, 5),
  ('metric.conversion_lift.read', 'metric', 'metric', 'Инкремент против контрольной группы', TRUE, 5),
  ('metric.new_guests_via_mall.read', 'metric', 'metric', 'Новые гости, пришедшие через комплекс', TRUE, 5),
  ('metric.repeat_visits.read', 'metric', 'metric', 'Повторные визиты', TRUE, 5),
  ('metric.revenue_total.read', 'metric', 'metric', 'Суммарная выручка по гостю/периоду', TRUE, 5),
  ('metric.avg_check.read', 'metric', 'metric', 'Средний чек', TRUE, 5),
  ('metric.visits_count.read', 'metric', 'metric', 'Число визитов', FALSE, 5),
  ('metric.segment_size.read', 'metric', 'metric', 'Размер сегмента', FALSE, 5),
  ('metric.export_counts.read', 'metric', 'metric', 'Счётчики выгрузки', FALSE, 5),
  ('metric.contact_frequency.read', 'metric', 'metric', 'Частота контактов', FALSE, 5),
  ('metric.data_quality.read', 'metric', 'metric', 'Доля отклонённых строк, свежесть', FALSE, 5),
  ('metric.merge_queue_size.read', 'metric', 'metric', 'Размер очереди спорных объединений', FALSE, 5),
  ('metric.campaign_results.read', 'metric', 'metric', 'Конверсия и инкремент по кампании', TRUE, 5),
  ('metric.tenant_priority.read', 'metric', 'metric', 'Приоритет продвижения арендатора', TRUE, 5),
  ('field.guest.display_name.read', 'field', 'field', 'displayName', TRUE, NULL),
  ('field.guest.phone.read', 'field', 'field', 'identifiers[].valueNormalized для phone', TRUE, NULL),
  ('field.guest.email.read', 'field', 'field', 'identifiers[].valueNormalized для email', TRUE, NULL),
  ('field.guest.card.read', 'field', 'field', 'identifiers[].valueNormalized для card', TRUE, NULL),
  ('field.guest.total_amount.read', 'field', 'field', 'totalAmountMinor, currency', TRUE, NULL),
  ('field.guest.identifiers.read', 'field', 'field', 'Массив identifiers целиком', TRUE, NULL),
  ('field.guest.segments.read', 'field', 'field', 'Массив segments', FALSE, NULL),
  ('field.guest.tenants_visited.read', 'field', 'field', 'Массив tenants, tenantCount', TRUE, NULL),
  ('field.event.amount.read', 'field', 'field', 'event.amountMinor', TRUE, NULL),
  ('field.event.raw_payload.read', 'field', 'field', 'event.rawPayload', TRUE, NULL),
  ('field.event.external_id.read', 'field', 'field', 'event.externalId', FALSE, NULL),
  ('field.export.contact.read', 'field', 'field', 'Контакт в выгрузке', TRUE, NULL),
  ('field.tenant.priority.read', 'field', 'field', 'tenant.priority', FALSE, NULL);

CREATE TABLE role_permission (
  role            VARCHAR(16) NOT NULL CHECK (role IN ('marketer', 'manager', 'tenant', 'analyst', 'system')),
  permission_code VARCHAR(64) NOT NULL REFERENCES permission(code),
  effect          VARCHAR(8) NOT NULL CHECK (effect IN ('allow', 'deny')),
  PRIMARY KEY (role, permission_code)
);

INSERT INTO role_permission (role, permission_code, effect) VALUES
  ('marketer', 'guest.profile.read', 'allow'),
  ('analyst', 'guest.profile.read', 'allow'),
  ('marketer', 'guest.profile.search', 'allow'),
  ('analyst', 'guest.profile.search', 'allow'),
  ('analyst', 'guest.profile.delete', 'allow'),
  ('marketer', 'guest.event.read', 'allow'),
  ('analyst', 'guest.event.read', 'allow'),
  ('marketer', 'guest.identifier.read', 'allow'),
  ('analyst', 'guest.identifier.read', 'allow'),
  ('marketer', 'guest.contact_history.read', 'allow'),
  ('analyst', 'guest.contact_history.read', 'allow'),
  ('marketer', 'segment.read', 'allow'),
  ('manager', 'segment.read', 'allow'),
  ('analyst', 'segment.read', 'allow'),
  ('marketer', 'segment.preview', 'allow'),
  ('analyst', 'segment.preview', 'allow'),
  ('marketer', 'segment.write', 'allow'),
  ('analyst', 'segment.write', 'allow'),
  ('marketer', 'segment.delete', 'allow'),
  ('analyst', 'segment.delete', 'allow'),
  ('marketer', 'segment.members.read', 'allow'),
  ('analyst', 'segment.members.read', 'allow'),
  ('marketer', 'recommendation.read', 'allow'),
  ('analyst', 'recommendation.read', 'allow'),
  ('marketer', 'recommendation.refresh', 'allow'),
  ('analyst', 'recommendation.refresh', 'allow'),
  ('marketer', 'identity.queue.read', 'allow'),
  ('analyst', 'identity.queue.read', 'allow'),
  ('analyst', 'identity.decide', 'allow'),
  ('manager', 'import.read', 'allow'),
  ('analyst', 'import.read', 'allow'),
  ('analyst', 'import.create', 'allow'),
  ('analyst', 'import.rejected_rows.read', 'allow'),
  ('marketer', 'data_source.read', 'allow'),
  ('manager', 'data_source.read', 'allow'),
  ('analyst', 'data_source.read', 'allow'),
  ('analyst', 'data_source.write', 'allow'),
  ('marketer', 'reference.read', 'allow'),
  ('manager', 'reference.read', 'allow'),
  ('tenant', 'reference.read', 'allow'),
  ('analyst', 'reference.read', 'allow'),
  ('analyst', 'reference.write', 'allow'),
  ('marketer', 'export.read', 'allow'),
  ('manager', 'export.read', 'allow'),
  ('analyst', 'export.read', 'allow'),
  ('marketer', 'export.create', 'allow'),
  ('marketer', 'export.download', 'allow'),
  ('marketer', 'export.members.read', 'allow'),
  ('marketer', 'campaign.read', 'allow'),
  ('manager', 'campaign.read', 'allow'),
  ('analyst', 'campaign.read', 'allow'),
  ('marketer', 'campaign.write', 'allow'),
  ('marketer', 'campaign.results.read', 'allow'),
  ('manager', 'campaign.results.read', 'allow'),
  ('marketer', 'contact_limit.read', 'allow'),
  ('manager', 'contact_limit.read', 'allow'),
  ('analyst', 'contact_limit.read', 'allow'),
  ('marketer', 'contact_limit.write', 'allow'),
  ('manager', 'user.read', 'allow'),
  ('manager', 'user.write', 'allow'),
  ('manager', 'audit.read', 'allow'),
  ('analyst', 'audit.read', 'allow'),
  ('manager', 'access.grant.read', 'allow'),
  ('manager', 'access.grant.write', 'allow'),
  ('manager', 'access.simulate', 'allow'),
  ('manager', 'data_quality.read', 'allow'),
  ('analyst', 'data_quality.read', 'allow'),
  ('marketer', 'dashboard.read', 'allow'),
  ('manager', 'dashboard.read', 'allow'),
  ('analyst', 'dashboard.read', 'allow'),
  ('marketer', 'tenant.aggregates.read', 'allow'),
  ('manager', 'tenant.aggregates.read', 'allow'),
  ('analyst', 'tenant.aggregates.read', 'allow'),
  ('tenant', 'tenant.aggregates.own', 'allow'),
  ('marketer', 'metric.first_purchase_new_tenant.read', 'allow'),
  ('manager', 'metric.first_purchase_new_tenant.read', 'allow'),
  ('analyst', 'metric.first_purchase_new_tenant.read', 'allow'),
  ('marketer', 'metric.cross_tenant_distribution.read', 'allow'),
  ('manager', 'metric.cross_tenant_distribution.read', 'allow'),
  ('analyst', 'metric.cross_tenant_distribution.read', 'allow'),
  ('marketer', 'metric.cross_tenant_rate.read', 'allow'),
  ('manager', 'metric.cross_tenant_rate.read', 'allow'),
  ('analyst', 'metric.cross_tenant_rate.read', 'allow'),
  ('marketer', 'metric.match_rate.read', 'allow'),
  ('manager', 'metric.match_rate.read', 'allow'),
  ('analyst', 'metric.match_rate.read', 'allow'),
  ('marketer', 'metric.disputed_merge_rate.read', 'allow'),
  ('manager', 'metric.disputed_merge_rate.read', 'allow'),
  ('analyst', 'metric.disputed_merge_rate.read', 'allow'),
  ('marketer', 'metric.segment_to_export_time.read', 'allow'),
  ('manager', 'metric.segment_to_export_time.read', 'allow'),
  ('analyst', 'metric.segment_to_export_time.read', 'allow'),
  ('marketer', 'metric.conversion_rate.read', 'allow'),
  ('manager', 'metric.conversion_rate.read', 'allow'),
  ('analyst', 'metric.conversion_rate.read', 'allow'),
  ('marketer', 'metric.conversion_lift.read', 'allow'),
  ('manager', 'metric.conversion_lift.read', 'allow'),
  ('analyst', 'metric.conversion_lift.read', 'allow'),
  ('marketer', 'metric.new_guests_via_mall.read', 'allow'),
  ('manager', 'metric.new_guests_via_mall.read', 'allow'),
  ('tenant', 'metric.new_guests_via_mall.read', 'allow'),
  ('analyst', 'metric.new_guests_via_mall.read', 'allow'),
  ('marketer', 'metric.repeat_visits.read', 'allow'),
  ('manager', 'metric.repeat_visits.read', 'allow'),
  ('tenant', 'metric.repeat_visits.read', 'allow'),
  ('analyst', 'metric.repeat_visits.read', 'allow'),
  ('marketer', 'metric.revenue_total.read', 'allow'),
  ('manager', 'metric.revenue_total.read', 'allow'),
  ('analyst', 'metric.revenue_total.read', 'allow'),
  ('marketer', 'metric.avg_check.read', 'allow'),
  ('manager', 'metric.avg_check.read', 'allow'),
  ('analyst', 'metric.avg_check.read', 'allow'),
  ('marketer', 'metric.visits_count.read', 'allow'),
  ('manager', 'metric.visits_count.read', 'allow'),
  ('tenant', 'metric.visits_count.read', 'allow'),
  ('analyst', 'metric.visits_count.read', 'allow'),
  ('marketer', 'metric.segment_size.read', 'allow'),
  ('manager', 'metric.segment_size.read', 'allow'),
  ('analyst', 'metric.segment_size.read', 'allow'),
  ('marketer', 'metric.export_counts.read', 'allow'),
  ('manager', 'metric.export_counts.read', 'allow'),
  ('analyst', 'metric.export_counts.read', 'allow'),
  ('marketer', 'metric.contact_frequency.read', 'allow'),
  ('manager', 'metric.contact_frequency.read', 'allow'),
  ('analyst', 'metric.contact_frequency.read', 'allow'),
  ('manager', 'metric.data_quality.read', 'allow'),
  ('analyst', 'metric.data_quality.read', 'allow'),
  ('marketer', 'metric.merge_queue_size.read', 'allow'),
  ('manager', 'metric.merge_queue_size.read', 'allow'),
  ('analyst', 'metric.merge_queue_size.read', 'allow'),
  ('marketer', 'metric.campaign_results.read', 'allow'),
  ('manager', 'metric.campaign_results.read', 'allow'),
  ('marketer', 'metric.tenant_priority.read', 'allow'),
  ('manager', 'metric.tenant_priority.read', 'allow'),
  ('analyst', 'metric.tenant_priority.read', 'allow');

INSERT INTO role_permission (role, permission_code, effect)
SELECT roles.role, permission.code, 'allow'
FROM (SELECT 'marketer' AS role UNION ALL SELECT 'manager' UNION ALL SELECT 'analyst') roles
CROSS JOIN permission
WHERE permission.code LIKE 'field.guest.%' OR permission.code LIKE 'field.event.%';

INSERT INTO role_permission (role, permission_code, effect)
SELECT 'tenant', permission.code, 'allow'
FROM permission
WHERE permission.code IN ('guest.profile.read', 'guest.event.read', 'guest.identifier.read')
   OR permission.code LIKE 'field.guest.%' OR permission.code LIKE 'field.event.%';

INSERT INTO role_permission (role, permission_code, effect)
VALUES
  ('manager', 'guest.profile.read', 'allow'),
  ('manager', 'guest.profile.search', 'allow'),
  ('manager', 'guest.event.read', 'allow'),
  ('manager', 'guest.identifier.read', 'allow'),
  ('manager', 'guest.contact_history.read', 'allow'),
  ('marketer', 'field.export.contact.read', 'allow'),
  ('marketer', 'field.tenant.priority.read', 'allow'),
  ('manager', 'field.tenant.priority.read', 'allow'),
  ('analyst', 'field.tenant.priority.read', 'allow');

CREATE TABLE access_policy_state (
  id         INTEGER PRIMARY KEY CHECK (id = 1),
  version    BIGINT NOT NULL CHECK (version >= 0),
  updated_at TIMESTAMP NOT NULL
);
INSERT INTO access_policy_state (id, version, updated_at) VALUES (1, 1, CURRENT_TIMESTAMP);

CREATE TABLE access_grant (
  id             TEXT PRIMARY KEY,
  subject_type   VARCHAR(16) NOT NULL CHECK (subject_type IN ('user', 'tenant', 'role')),
  subject_id     TEXT NOT NULL,
  preset_code    VARCHAR(32),
  status         VARCHAR(16) NOT NULL CHECK (status IN ('draft', 'active', 'suspended', 'revoked', 'expired')),
  valid_from     TIMESTAMP NOT NULL,
  valid_to       TIMESTAMP,
  created_by     TEXT REFERENCES app_user(id),
  approved_by    TEXT REFERENCES app_user(id),
  reason         TEXT,
  policy_version BIGINT NOT NULL,
  created_at     TIMESTAMP NOT NULL,
  updated_at     TIMESTAMP NOT NULL,
  CHECK (valid_to IS NULL OR valid_to > valid_from)
);
CREATE INDEX ix_grant_subject ON access_grant(subject_type, subject_id, status);
CREATE INDEX ix_grant_window ON access_grant(status, valid_from, valid_to);

CREATE TABLE access_grant_item (
  grant_id        TEXT NOT NULL REFERENCES access_grant(id),
  permission_code VARCHAR(64) NOT NULL REFERENCES permission(code),
  effect          VARCHAR(8) NOT NULL CHECK (effect IN ('allow', 'deny')),
  PRIMARY KEY (grant_id, permission_code)
);

CREATE TABLE access_grant_scope (
  grant_id    TEXT NOT NULL REFERENCES access_grant(id),
  scope_type  VARCHAR(24) NOT NULL CHECK (scope_type IN
                ('tenant', 'category', 'period_days', 'channel', 'k_threshold', 'row_limit', 'field_mask', 'metric', 'principal_role')),
  scope_value TEXT NOT NULL,
  PRIMARY KEY (grant_id, scope_type, scope_value)
);

INSERT INTO access_grant (id, subject_type, subject_id, status, valid_from, reason, policy_version, created_at, updated_at)
SELECT 'bootstrap-role-' || role, 'role', role, 'active', CURRENT_TIMESTAMP,
       'Initial role access migration', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM (SELECT 'marketer' AS role UNION ALL SELECT 'manager' UNION ALL SELECT 'analyst') roles;

INSERT INTO access_grant_item (grant_id, permission_code, effect)
SELECT 'bootstrap-role-' || role, permission_code, effect
FROM role_permission
WHERE role IN ('marketer', 'manager', 'analyst')
  AND (role <> 'manager' OR (
    permission_code NOT LIKE 'field.guest.%' AND permission_code NOT LIKE 'field.event.%'
    AND permission_code NOT IN ('guest.profile.read', 'guest.profile.search', 'guest.event.read',
                                'guest.identifier.read', 'guest.contact_history.read')
  ));

INSERT INTO audit_log (id, actor_user_id, actor_role, action, entity_type, entity_id, details, created_at)
SELECT 'bootstrap-access-audit-' || md5(id), NULL, 'system', 'access.grant.create',
       'access_grant', id, '{"origin":"migration","policyVersion":1}', CURRENT_TIMESTAMP
FROM access_grant WHERE id LIKE 'bootstrap-role-%';

-- +goose Down
DROP TABLE IF EXISTS contact_log;
DROP TABLE IF EXISTS contact_limit;
DROP TABLE IF EXISTS export_member;
DROP TABLE IF EXISTS export_job;
DROP TABLE IF EXISTS campaign;
DROP TABLE IF EXISTS segment_member;
DROP TABLE IF EXISTS segment_version;
DROP TABLE IF EXISTS segment;
DROP TABLE IF EXISTS recommendation_eval;
DROP TABLE IF EXISTS recommendation_factor;
DROP TABLE IF EXISTS recommendation;
DROP TABLE IF EXISTS merge_block;
DROP TABLE IF EXISTS merge_candidate;
DROP TABLE IF EXISTS event;
DROP TABLE IF EXISTS rejected_row;
DROP TABLE IF EXISTS import_batch;
DROP TABLE IF EXISTS guest_identifier;
DROP TABLE IF EXISTS guest;
DROP TABLE IF EXISTS access_grant_scope;
DROP TABLE IF EXISTS access_grant_item;
DROP TABLE IF EXISTS access_grant;
DROP TABLE IF EXISTS access_policy_state;
DROP TABLE IF EXISTS role_permission;
DROP TABLE IF EXISTS permission;
DROP TABLE IF EXISTS user_invitation;
DROP TABLE IF EXISTS audit_log;
DROP TABLE IF EXISTS app_user;
DROP TABLE IF EXISTS data_source;
DROP TABLE IF EXISTS data_source_type;
DROP TABLE IF EXISTS event_type;
DROP TABLE IF EXISTS project;
DROP TABLE IF EXISTS tenant;
DROP TABLE IF EXISTS mall;
DROP TABLE IF EXISTS tenant_category_relation;
DROP TABLE IF EXISTS tenant_category;
DROP TABLE IF EXISTS job_queue;
