-- +goose Up
ALTER TABLE app_user ADD COLUMN updated_at TIMESTAMP;
UPDATE app_user SET updated_at = created_at;
ALTER TABLE app_user ALTER COLUMN updated_at SET NOT NULL;
ALTER TABLE app_user ALTER COLUMN updated_at SET DEFAULT CURRENT_TIMESTAMP;

-- +goose Down
ALTER TABLE app_user DROP COLUMN IF EXISTS updated_at;
