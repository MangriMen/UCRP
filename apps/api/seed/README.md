# Demo data generator

Fills the database with synthetic data for demos: tenants, guests, events, segments, recommendations, campaigns and exports.

## Usage

```sh
cd apps/api
go run ./cmd/seed -dsn "postgres://ucrp:ucrp@127.0.0.1:5543/ucrp?sslmode=disable"
go run ./cmd/seed -dry-run      # print the plan without writing
go run ./cmd/seed -guests 300   # background guests, 10..2000, default 92
go run ./cmd/seed -reset        # delete demo data only
```

Without `-dsn` the value of `UCRP_DATABASE_URL` is used. Migrations are applied automatically. From the repository root: `bun run --cwd apps/api seed` and `seed:reset`.

`-guests` sets only background guests; 16 scenario personas, 8 merged and 4 deleted profiles are always added, so the total is `guests + 28`. Events, segments, recommendations and exports scale with it.

## Properties

- **Reproducible.** IDs are UUIDv5 of stable names (`ucrp-demo/tenant/cafe-bean`), so a rerun produces the same rows.
- **Safe to rerun.** Demo rows are deleted in foreign-key order and recreated. Data created by the app or e2e tests is untouched.
- **Consistent.** Segments are computed by `segments.SqlEvaluator`, recommendations by `recommend.WeightedRecommender`, exports by `service.ExportService`, so derived data matches the business logic.
- **No real personal data.** Fictional names, phones in the `79…` range, emails in `demo.ucrp.test`, cards starting with `4000`.

After seeding, the tool prints a verification table (import row balances, segment and export counters, link integrity, no duplicates) and exits with code 1 if any check fails.

## Accounts

All passwords are `demo12345`. Logins: `marketer`, `manager`, `analyst`, `tenant.cafe`, `tenant.cinema`, `tenant.fit`; email is `<login>@demo.ucrp.test`. The `analyst.old` account is disabled.

## Demo scenarios

- **Sleeping neighbor, cinema.** Segment "visited the cinema, not the cafe"; an SMS export with a 33% control group shows blocking by contact limit and missing phones.
- **Duplicates.** `g-olga` and `g-igor` share a phone and have a `pending` merge candidate.
- **Reversal.** `g-reversal` has a purchase and its `reversal`.
- **Newcomer.** `g-new` has no events, so recommendations fall back to tenant priorities.
- **Churn.** `g-churn` has its last event 60 days ago.

Sample CSV files for manual import through the API are in `seed/data/`.
