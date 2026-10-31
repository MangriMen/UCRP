-- +goose Up
ALTER TABLE app_user ADD COLUMN auth_version BIGINT NOT NULL DEFAULT 0;
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

-- +goose Down
DROP TABLE IF EXISTS user_invitation;
ALTER TABLE app_user DROP COLUMN IF EXISTS auth_version;
