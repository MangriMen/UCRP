-- +goose Up
ALTER TABLE tenant_category ADD COLUMN description VARCHAR(1000);
ALTER TABLE tenant_category ADD COLUMN updated_at TIMESTAMP;

-- +goose Down
ALTER TABLE tenant_category DROP COLUMN updated_at;
ALTER TABLE tenant_category DROP COLUMN description;
