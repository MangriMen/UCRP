# OpenAPI 3.0 — структура спецификации

Точка входа — `../openapi.json`. Это индекс: `info`, `servers`, `tags`, `security` и карта
`$ref` на внешние файлы. Каждый файл ≤ 300 строк.

## Проверка и сборка

```powershell
# валидация (все внешние $ref разрешаются)
npx @redocly/cli lint docs/openapi.json

# сборка в один файл
npx @redocly/cli bundle docs/openapi.json --output docs/openapi.bundled.json

# превью документации
npx @redocly/cli preview-docs docs/openapi.json
```

## paths/ — операции (90 эндпоинтов из `endpoint-registry.md`)

| Файл | Раздел реестра | Эндпоинты |
|---|---|---|
| `auth.json` | 1. Аутентификация | login, refresh, logout, me |
| `users.json` | 2. Пользователи | `/users`, `/users/{id}` |
| `reference-tenants.json` | 3. Справочники | `/tenants`, `/tenants/{id}` |
| `reference-tenant-categories.json` | 3. Справочники | `/tenant-categories*`, `/{id}/relations` |
| `reference-projects.json` | 3. Справочники | `/projects`, `/projects/{id}` |
| `reference-event-types.json` | 3. Справочники | `/event-types`, `/event-types/{id}` |
| `reference-data-sources.json` | 3. Справочники | `/data-sources`, `/data-sources/{id}` |
| `reference-data-source-types.json` | 3. Справочники | `/data-source-types`, `/{id}` |
| `imports.json` | 4. Импорт | `/imports*`, `/data-sources/{id}/sync` |
| `guests.json` | 5. Гости | `/guests`, `/{id}`, `/{id}/events`, `/{id}/contacts` |
| `recommendations.json` | 5. Гости | `/{id}/recommendations*`, `/segments/{id}/recommendations` |
| `identity-merge-candidates.json` | 6. Идентичность | `/merge-candidates*` (confirm/reject/revert) |
| `identity-unmerge.json` | 6. Идентичность | `/guests/{id}/merge`, `/unmerge-blocks*` |
| `segments.json` | 7. Сегменты | `/segment-templates`, `/segments`, `/segments/preview` |
| `segments-detail.json` | 7. Сегменты | `/segments/{id}`, `/{id}/members`, `/{id}/versions`, `/{id}/recompute` |
| `campaigns.json` | 8. Кампании | `/campaigns*`, `/{id}/results` |
| `exports.json` | 9. Выгрузка | `/exports/preview`, `/exports`, `/exports/{id}` |
| `exports-delivery.json` | 9. Выгрузка | `/exports/{id}/members`, `/{id}/download` |
| `contact-limits.json` | 9. Выгрузка | `/contact-limits`, `/contact-limits/{id}` |
| `analytics.json` | 10. Аналитика | `/dashboard/*`, `/tenants/{id}/aggregates`, `/data-quality/*` |
| `audit.json` | 11. Аудит | `/audit-logs` |
| `system.json` | 11. Служебные | `/healthz` |

## components/schemas/ — модели (113 схем)

| Файл | Сущности ER / назначение |
|---|---|
| `common.json` | `ErrorResponse`, `ErrorDetail`, `PageMeta`, `UserRole`, `Granularity`, `JobAccepted`, `HealthResponse` |
| `users.json` | E-20 `User`, аутентификация (`LoginRequest`, `TokenResponse`, `MeResponse`) |
| `reference-tenants.json` | E-4 `Tenant` + create/update/page |
| `reference-tenant-categories.json` | E-6 `TenantCategory`, связи `adjacent`/`competitor` |
| `reference-projects.json` | E-7 `Project` |
| `reference-event-types.json` | E-9 `EventType` |
| `reference-data-sources.json` | E-5 `DataSource`, `DataSourceType`, `FieldMapping` |
| `imports.json` | E-13 `ImportBatch`, E-14 `RejectedRow` |
| `guests.json` | E-1 `Guest`, E-10 `GuestIdentifier`, `GuestProfile` |
| `events.json` | E-2 `Event`, E-19 `ContactLog` |
| `recommendations.json` | E-15 `Recommendation`, `RecommendationFactor` |
| `identity.json` | E-11 `MergeCandidate`, E-12 `UnmergeBlock` |
| `segments.json` | E-3 `Segment`, `SegmentVersion`, `SegmentRule` |
| `segment-requests.json` | E-3 запросы, preview, состав, шаблоны |
| `campaigns.json` | E-8 `Campaign`, `CampaignResults` |
| `exports.json` | E-16 `Export`, preview/create |
| `export-members.json` | E-17 `ExportMember`, E-18 `ContactLimit` |
| `analytics.json` | Сводки: penetration, метрики |
| `analytics-tenant.json` | Агрегаты R-3 (порог k=5), качество данных |
| `audit.json` | E-21 `AuditLog` |

## components/ — общие компоненты

| Файл | Содержимое |
|---|---|
| `security.json` | `bearerAuth` (JWT), `cookieAuth` (refresh) |
| `parameters.json` | `IdPath`, `Page`, `PageSize`, `From`, `To`, `Granularity`, `IdempotencyKey` |
| `responses.json` | `BadRequest`, `Unauthorized`, `Forbidden`, `NotFound`, `Conflict`, `PayloadTooLarge`, `UnprocessableEntity`, `TooManyRequests`, `GatewayTimeout`, `NoContent` |

## Правила при правках

1. Новый эндпоинт — только если он есть в `endpoint-registry.md`.
2. Схему кладём в тематический файл и добавляем строку-`$ref` в `../openapi.json`.
3. Ссылки между файлами — относительные, с JSON Pointer: `./common.json#/PageMeta`.
4. После правок обязательно `npx @redocly/cli lint docs/openapi.json`.
