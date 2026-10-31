-- +goose Up
CREATE UNIQUE INDEX ux_tenant_code_lower ON tenant (lower(code));

-- +goose Down
DROP INDEX IF EXISTS ux_tenant_code_lower;
