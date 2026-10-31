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

CREATE TABLE tenant (
  id          TEXT         PRIMARY KEY,
  code        VARCHAR(64)  NOT NULL UNIQUE,
  name        VARCHAR(200) NOT NULL,
  category_id TEXT         NOT NULL REFERENCES tenant_category(id),
  kind        VARCHAR(16)  NOT NULL DEFAULT 'mall' CHECK (kind IN ('mall','external')),
  floor       SMALLINT     NOT NULL,
  x_m         INTEGER,
  y_m         INTEGER,
  priority    SMALLINT     NOT NULL DEFAULT 50 CHECK (priority BETWEEN 0 AND 100),
  is_active   BOOLEAN      NOT NULL DEFAULT TRUE,
  created_at  TIMESTAMP    NOT NULL,
  updated_at  TIMESTAMP    NOT NULL,
  deleted_at  TIMESTAMP
);
CREATE INDEX ix_tenant_category ON tenant(category_id);

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
