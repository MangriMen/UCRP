-- +goose Up
CREATE TABLE user_invitation (
  id TEXT PRIMARY KEY,
  email VARCHAR(320) NOT NULL,
  display_name VARCHAR(200) NOT NULL,
  role VARCHAR(16) NOT NULL CHECK (role IN ('manager', 'tenant')),
  tenant_id TEXT REFERENCES tenant(id),
  token_hash BYTEA NOT NULL UNIQUE,
  created_by TEXT NOT NULL REFERENCES app_user(id),
  policy_version BIGINT NOT NULL,
  created_at TIMESTAMP NOT NULL,
  expires_at TIMESTAMP NOT NULL,
  consumed_at TIMESTAMP,
  revoked_at TIMESTAMP,
  CHECK (expires_at > created_at),
  CHECK ((role = 'tenant' AND tenant_id IS NOT NULL) OR (role = 'manager' AND tenant_id IS NULL))
);
CREATE INDEX ix_user_invitation_email ON user_invitation(email);

-- +goose Down
DROP TABLE IF EXISTS user_invitation;
