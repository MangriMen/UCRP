# OpenAPI 3.0 — specification structure

Entry point — `../openapi.json`. It is an index: `info`, `servers`, `tags`, `security`, and a `$ref`
map to external files. Every file is ≤ 300 lines.

## Validate and bundle

```powershell
# validate (all external $refs resolve)
npx @redocly/cli lint docs/openapi.json

# bundle into a single file
npx @redocly/cli bundle docs/openapi.json --output docs/openapi.bundled.json

# preview the docs
npx @redocly/cli preview-docs docs/openapi.json
```

## paths/ — operations (104 operations from `endpoint-registry.md`)

| File                                | Registry section    | Endpoints                                                              |
| ----------------------------------- | ------------------- | ---------------------------------------------------------------------- |
| `access.json`, `access-grants.json` | 12. Access          | grants, permissions, presets, simulation, effective rights and audit   |
| `invitations.json`                  | 1–2. Auth and users | create and accept employee invitations; delivery is not configured     |
| `auth.json`                         | 1. Authentication   | login, refresh, logout, me                                             |
| `users.json`                        | 2. Users            | `/users`, `/users/{id}`                                                |
| `reference-tenants.json`            | 3. Reference data   | `/tenants`, `/tenants/{id}`                                            |
| `reference-tenant-categories.json`  | 3. Reference data   | `/tenant-categories*`, `/{id}/relations`                               |
| `reference-projects.json`           | 3. Reference data   | `/projects`, `/projects/{id}`                                          |
| `reference-event-types.json`        | 3. Reference data   | `/event-types`, `/event-types/{id}`                                    |
| `reference-data-sources.json`       | 3. Reference data   | `/data-sources`, `/data-sources/{id}`                                  |
| `reference-data-source-types.json`  | 3. Reference data   | `/data-source-types`, `/{id}`                                          |
| `imports.json`                      | 4. Imports          | `/imports*`, `/data-sources/{id}/sync`                                 |
| `guests.json`                       | 5. Guests           | `/guests`, `/{id}`, `/{id}/events`, `/{id}/contacts`                   |
| `recommendations.json`              | 5. Guests           | `/{id}/recommendations*`, `/segments/{id}/recommendations`             |
| `identity-merge-candidates.json`    | 6. Identity         | `/merge-candidates*` (confirm/reject/revert)                           |
| `identity-unmerge.json`             | 6. Identity         | `/guests/{id}/merge`, `/unmerge-blocks*`                               |
| `segments.json`                     | 7. Segments         | `/segment-templates`, `/segments`, `/segments/preview`                 |
| `segments-detail.json`              | 7. Segments         | `/segments/{id}`, `/{id}/members`, `/{id}/versions`, `/{id}/recompute` |
| `campaigns.json`                    | 8. Campaigns        | `/campaigns*`, `/{id}/results`                                         |
| `exports.json`                      | 9. Exports          | `/exports/preview`, `/exports`, `/exports/{id}`                        |
| `exports-delivery.json`             | 9. Exports          | `/exports/{id}/members`, `/{id}/download`                              |
| `contact-limits.json`               | 9. Exports          | `/contact-limits`, `/contact-limits/{id}`                              |
| `analytics.json`                    | 10. Analytics       | `/dashboard/*`, `/tenants/{id}/aggregates`, `/data-quality/*`          |
| `audit.json`                        | 11. Audit           | `/audit-logs`                                                          |
| `system.json`                       | 11. Service         | `/healthz`                                                             |

## components/schemas/ — models (113 schemas)

| File                                 | ER entities / purpose                                                                                  |
| ------------------------------------ | ------------------------------------------------------------------------------------------------------ |
| `access.json`, `access-results.json` | Access permissions, grants, simulation and audit responses                                             |
| `tenant-guest.json`                  | Tenant-scoped guest, events and aggregate responses                                                    |
| `common.json`                        | `ErrorResponse`, `ErrorDetail`, `PageMeta`, `UserRole`, `Granularity`, `JobAccepted`, `HealthResponse` |
| `users.json`                         | E-20 `User`, authentication (`LoginRequest`, `TokenResponse`, `MeResponse`)                            |
| `reference-tenants.json`             | E-4 `Tenant` + create/update/page                                                                      |
| `reference-tenant-categories.json`   | E-6 `TenantCategory`, `adjacent`/`competitor` relations                                                |
| `reference-projects.json`            | E-7 `Project`                                                                                          |
| `reference-event-types.json`         | E-9 `EventType`                                                                                        |
| `reference-data-sources.json`        | E-5 `DataSource`, `DataSourceType`, `FieldMapping`                                                     |
| `imports.json`                       | E-13 `ImportBatch`, E-14 `RejectedRow`                                                                 |
| `guests.json`                        | E-1 `Guest`, E-10 `GuestIdentifier`, `GuestProfile`                                                    |
| `events.json`                        | E-2 `Event`, E-19 `ContactLog`                                                                         |
| `recommendations.json`               | E-15 `Recommendation`, `RecommendationFactor`                                                          |
| `identity.json`                      | E-11 `MergeCandidate`, E-12 `UnmergeBlock`                                                             |
| `segments.json`                      | E-3 `Segment`, `SegmentVersion`, `SegmentRule`                                                         |
| `segment-requests.json`              | E-3 requests, preview, membership, templates                                                           |
| `campaigns.json`                     | E-8 `Campaign`, `CampaignResults`                                                                      |
| `exports.json`                       | E-16 `Export`, preview/create                                                                          |
| `export-members.json`                | E-17 `ExportMember`, E-18 `ContactLimit`                                                               |
| `analytics.json`                     | Dashboards: penetration, metrics                                                                       |
| `analytics-tenant.json`              | R-3 aggregates (k = 5 threshold), data quality                                                         |
| `audit.json`                         | E-21 `AuditLog`                                                                                        |

## components/ — shared components

| File              | Contents                                                                                                                                                      |
| ----------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `security.json`   | `bearerAuth` (JWT), `cookieAuth` (refresh)                                                                                                                    |
| `parameters.json` | `IdPath`, `Page`, `PageSize`, `From`, `To`, `Granularity`, `IdempotencyKey`                                                                                   |
| `responses.json`  | `BadRequest`, `Unauthorized`, `Forbidden`, `NotFound`, `Conflict`, `PayloadTooLarge`, `UnprocessableEntity`, `TooManyRequests`, `GatewayTimeout`, `NoContent` |

## Editing rules

1. Add an endpoint only if it exists in `endpoint-registry.md`.
2. Put a schema in the matching topic file and add a `$ref` line to `../openapi.json`.
3. Cross-file references are relative, with a JSON Pointer: `./common.json#/PageMeta`.
4. Always run `npx @redocly/cli lint docs/openapi.json` after changes.
