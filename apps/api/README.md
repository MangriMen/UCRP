# apps/api

Go backend for UCRP: HTTP API (`cmd/api`), background workers (`cmd/worker`) and a demo data generator (`cmd/seed`). Module: `github.com/MangriMen/UCRP`.

Layering and dependency rules are in [ARCHITECTURE.md](../../ARCHITECTURE.md). The API contract is `packages/openapi/`; the data model is explained in `docs/design/05-er-data-model.md`.

## Commands

```sh
cd apps/api
go build ./...
go test ./...
go run ./cmd/api
go run ./cmd/worker
```

End-to-end tests against a running API live in `e2e/` (`e2e/run.sh` or `e2e/run.ps1`). Reports are written to `apps/api/e2e-report.xml` and `apps/api/e2e-contract-drift.md` and are git-ignored.

Docker (see `deploy/`), from the repository root:

```sh
cp deploy/.env.example deploy/.env
docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.dev.yml up --build
```

The API then answers on `http://localhost:8080/healthz`.

## Migrations

The backend currently runs on PostgreSQL only (`pgx`, goose dialect `postgres`). Goose files live in `migrations/common/` (written as portable SQL) and `migrations/postgres/` (PostgreSQL-only steps). Files are named `NNNNN_name.sql` with `-- +goose Up` and `-- +goose Down` blocks. An applied migration is never edited; add a new one.

## Configuration

| Variable                                  | Purpose                                                                 |
| ----------------------------------------- | ----------------------------------------------------------------------- |
| `UCRP_HTTP_ADDR`                          | Listen address, for example `:8080`                                     |
| `UCRP_DATABASE_URL`                       | `sqlite://...` or `postgres://...`                                      |
| `UCRP_JWT_SECRET`                         | JWT signing key                                                         |
| `UCRP_UPLOAD_DIR`                         | Directory for uploaded import files                                     |
| `UCRP_LOG_LEVEL`                          | `debug` to `error`                                                      |
| `UCRP_BOOTSTRAP_MANAGER_EMAIL`            | Email of the first manager, used only when no manager exists            |
| `UCRP_BOOTSTRAP_MANAGER_PASSWORD`         | Password of the first manager; must be set together with the email      |
| `UCRP_ADMIN_EMAIL`, `UCRP_ADMIN_PASSWORD` | Initial administrator used by `cmd/api` and `deploy/docker-compose.yml` |
| `UCRP_RECOMMEND_*`                        | Recommender weights and windows, see `internal/platform/config`         |

Bootstrap runs once: on a database without managers it creates the account and a personal data grant. Any existing manager, even an inactive one, disables it, and a restart never restores a revoked grant. Keep the password in environment secrets. A separate PostgreSQL bootstrap test can be pointed at a database with `UCRP_BOOTSTRAP_TEST_DATABASE_URL`; it creates and drops an isolated schema.

## Demo data

`cmd/seed` fills the database with synthetic tenants, guests, events, segments, recommendations, campaigns and exports. See [seed/README.md](seed/README.md).
