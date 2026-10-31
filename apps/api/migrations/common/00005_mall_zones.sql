-- +goose Up
CREATE TABLE mall (
  id         TEXT         PRIMARY KEY,
  code       VARCHAR(64)  NOT NULL UNIQUE,
  name       VARCHAR(200) NOT NULL,
  is_active  BOOLEAN      NOT NULL DEFAULT TRUE,
  created_at TIMESTAMP    NOT NULL
);

INSERT INTO mall (id, code, name, is_active, created_at)
VALUES ('00000000-0000-0000-0000-000000000001', 'main', 'Основной ТЦ', TRUE, now() AT TIME ZONE 'utc');

ALTER TABLE tenant ADD COLUMN mall_id TEXT REFERENCES mall(id);
UPDATE tenant SET mall_id = '00000000-0000-0000-0000-000000000001';
ALTER TABLE tenant ALTER COLUMN mall_id SET NOT NULL;
CREATE INDEX ix_tenant_mall_floor ON tenant(mall_id, floor);

-- +goose Down
DROP INDEX IF EXISTS ix_tenant_mall_floor;
ALTER TABLE tenant DROP COLUMN IF EXISTS mall_id;
DROP TABLE IF EXISTS mall;
