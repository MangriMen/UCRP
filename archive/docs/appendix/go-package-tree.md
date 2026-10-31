# Дерево Go-пакетов (проектное, без кода)

Модуль: `mallhub` (условное имя). Зависимости направлены внутрь: `handler → service → domain ← repository`; пакеты возможностей (`ingestion`, `identity`, `segments`, `recommend`, `export`) зависят только от `domain` и `platform`, на `service` они не ссылаются (сервис вызывает их через интерфейсы портов). Интерфейсы репозиториев объявлены в `domain`, реализации — в `repository`.

```text
cmd/
  api/            точка входа HTTP: конфиг, сборка зависимостей, graceful shutdown
  worker/         точка входа воркеров: цикл Take из job_queue, планировщик connector.pull
migrations/       goose, 13 файлов (см. 05-er-data-model.md)
internal/
  domain/         сущности E-1…E-21 и T-1 (Job), value objects (Principal, Window, Counters),
                  ошибки (ErrNotFound, ErrConflict, ErrForbidden, ErrTimeout), интерфейсы
                  репозиториев и портов (Queue, BlobStore, Cache, Clock, IDGen)
  handler/        HTTP-слой: роутер /api/v1, RequireRole по реестру, разбор запросов
    dto/          запросы/ответы JSON, TenantAggregates, GuestProfileDTO, ErrorResponse
    auth_handler, import_handler, guest_handler, identity_handler, segment_handler,
    recommendation_handler, campaign_handler, export_handler, contact_limit_handler,
    dashboard_handler, tenant_handler,
    dq_handler, reference_handler, user_handler, audit_handler
  service/        сценарии: AuthService, ImportService, ProfileService, IdentityService,
                  SegmentService, RecommendationService, ExportService, ContactLimitService,
                  CampaignService, DashboardService, TenantAggregateService,
                  DataQualityService, ReferenceService, UserService, AuditService
  repository/     SQL-реализации (диалект SQLite/PostgreSQL), по одному на таблицу/агрегат
    sqlstore/     подключение, транзакции (UnitOfWork), хелперы диалекта
    guest_repo, guest_identifier_repo, event_repo, merge_candidate_repo, unmerge_block_repo,
    import_batch_repo, rejected_row_repo, segment_repo, segment_version_repo,
    segment_member_repo, recommendation_repo, export_repo, export_member_repo,
    contact_limit_repo, contact_log_repo, campaign_repo, tenant_repo,
    tenant_category_repo, tenant_category_relation_repo, project_repo, event_type_repo,
    data_source_repo, data_source_type_repo, user_repo, audit_repo, job_queue_repo,
    analytics_repo, tenant_aggregate_repo (только методы ...ForTenant)
  ingestion/      конвейер импорта
    connector/    csv_excel, crm, pos, api (реализации SourceConnector)
    registry, pipeline, validator, mapping
  identity/       normalize (чистые функции, без репозиториев), DefaultResolver (IdentityResolver), Policy, MergeExecutor
  segments/       RuleCompiler, SqlEvaluator (SegmentEvaluator), шаблоны правил
  recommend/      Recommender, WeightedScorer, factors/ (similar_guests, category_adjacency,
                  distance, recency, tenant_priority), explainer
  export/         ExportChannel (sms, push, email), ExportFormat (csv, api),
                  ControlGroupSelector
  platform/       сквозная инфраструктура
    auth/         JWT, refresh, пароли, middleware Principal
    httpx/        request_id, таймауты, recover, формат ошибок, rate limit
    queue/        реализация Queue поверх job_queue, ретраи, dedup
    cache/        in-memory TTL-кэш размеров сегментов
    storage/      BlobStore (локальный каталог uploads/)
    config/       чтение конфигурации (веса скоринга, пороги, таймауты)
    logging/      структурные логи с request_id
    scheduler/    расписание pull и очистки
```

## Правила зависимостей

| Пакет                                                      | Может импортировать                                    | Не может                                                 |
| ---------------------------------------------------------- | ------------------------------------------------------ | -------------------------------------------------------- |
| `domain`                                                   | стандартную библиотеку                                 | всё остальное                                            |
| `handler`                                                  | `service`, `domain`, `platform/auth`, `platform/httpx` | `repository`, `ingestion`…                               |
| `service`                                                  | `domain`, пакеты возможностей через интерфейсы         | `handler`, `repository` (только интерфейсы из `domain`)  |
| `repository`                                               | `domain`, `repository/sqlstore`                        | `service`, `handler`                                     |
| `ingestion`, `identity`, `segments`, `recommend`, `export` | `domain`, `platform/*`                                 | `handler`, друг друга напрямую (через `service`/очередь) |
| `cmd/*`                                                    | всё (композиция)                                       | —                                                        |

Исключения: пакет `identity/normalize` (чистые функции нормализации ключей, без доступа к БД) могут импортировать `ingestion` и `identity`; во всех остальных случаях `ingestion` не вызывает `identity` напрямую. Исключение для Identity: `ingestion` не вызывает `identity` напрямую — он ставит задание `identity.resolve` в очередь. Изоляция R-3: `TenantAggregateService` получает только `tenant_aggregate_repo` с методами `...ForTenant`, прямого доступа к `event_repo`/`guest_repo` у него нет.

## Соответствие компонентам (см. 11-component-diagram.md)

| Компонент              | Пакеты                                                                                                                             |
| ---------------------- | ---------------------------------------------------------------------------------------------------------------------------------- |
| CMP-1 Ingestion        | `internal/ingestion`, `service.ImportService`                                                                                      |
| CMP-2 Identity         | `internal/identity`, `service.IdentityService`                                                                                     |
| CMP-3 Profile & Events | `service.ProfileService`, `repository.guest_repo/event_repo`                                                                       |
| CMP-4 Segments         | `internal/segments`, `service.SegmentService`                                                                                      |
| CMP-5 Recommendations  | `internal/recommend`, `service.RecommendationService`                                                                              |
| CMP-6 Export           | `internal/export`, `service.ExportService`, `service.ContactLimitService`, `service.CampaignService`                               |
| CMP-7 Analytics        | `service.DashboardService`, `service.TenantAggregateService`, `service.DataQualityService`, `service.CampaignService` (результаты) |
| CMP-8 Access & Auth    | `internal/platform/auth`, `service.AuthService`, `service.UserService`                                                             |
| CMP-9 Audit            | `service.AuditService`, `repository.audit_repo`                                                                                    |
| CMP-10 Reference       | `service.ReferenceService`                                                                                                         |

## Чекпойнт

Создано: дерево Go-пакетов `cmd/{api,worker}`, `internal/{domain,handler,service,repository,ingestion,identity,segments,recommend,export,platform}`; соответствие CMP-1…CMP-10. Решения: D-28.
