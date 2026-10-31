# Дерево React-модулей (проектное, без кода)

Стек: React + Vite + TypeScript, shadcn/ui (Tailwind), React Router, TanStack Query и Table, React Hook Form. Связь с backend — только REST `/api/v1` через `shared/api`. Структура соответствует `ARCHITECTURE.md`, раздел 3 (`app → features → shared`). Модуль `features/admin` добавлен к списку плана (D-14).

```text
src/
  app/                провайдеры (QueryClient, Router), конфигурация маршрутов, layout,
                      меню по роли, защищённые маршруты, 403/404 страницы, глобальные стили
  features/
    auth/             AuthProvider, LoginPage, useAuth (role, tenantId), RoleGuard
    import/           UploadForm, BatchList, BatchStatus (опрос 2 с), RejectedTable, SyncButton
    profiles/         GuestSearch, GuestProfile, EventFeed, MergeQueue, MergeCompare,
                      UnmergeBlocks, ContactHistory
    segments/         TemplateGallery, RuleBuilder, PreviewPanel, SegmentList, SegmentCard,
                      MemberTable, VersionHistory
    recommendations/  RecommendationList, WhyPopover, SegmentTopTenants
    export/           ExportWizard, CountsCard, ExportJournal, ExportResult, LimitSettings,
                      CampaignList, CampaignResults
    dashboard/        PenetrationChart, MetricsPanel (для R-2, R-1, R-4)
    tenant/           TenantAggregatesPage (единственный раздел R-3), SmallCellNotice
    dq/               SourceQualityTable, SourceDetail
    admin/            Tenants, Categories, Projects, EventTypes, DataSources, Users, AuditLog
  shared/
    api/              клиент fetch, перехват 401 → refresh (single-flight), формат ошибок,
                      QueryClient; хуки TanStack Query по ресурсам живут в features/*/api
    ui/               shadcn-компоненты и базовые: DataTable (виртуализация), FormField,
                      ErrorBanner, StatusBadge, ConfirmDialog, Pagination, EmptyState
    lib/              утилиты (cn и т. п.)
```

## Правила

| Правило                    | Описание                                                                                                                |
| -------------------------- | ----------------------------------------------------------------------------------------------------------------------- |
| Единственный вход к данным | Компоненты не вызывают `fetch` напрямую, только хуки из `shared/api` или `features/*/api`                               |
| Доступ                     | `app/` (маршруты, меню) скрывает пункты по `role`; решающая проверка на backend (403 обрабатывается `ErrorBanner`)      |
| Изоляция R-3               | Для роли `tenant` в меню один пункт `features/tenant`; `tenantId` берётся из `GET /auth/me`, в URL и формах не вводится |
| Состояние                  | Серверное — TanStack Query; локальное — состояние компонента; глобальный стор только для `features/auth`                |
| Обновления                 | Опрос (D-22): статус импорта 2 с, пересчёт сегмента и рекомендации 5 с                                                  |

## Соответствие компонентам

| Компонент              | Модули                                                        |
| ---------------------- | ------------------------------------------------------------- |
| CMP-1 Ingestion        | `features/import`                                             |
| CMP-2 Identity         | `features/profiles` (MergeQueue, MergeCompare, UnmergeBlocks) |
| CMP-3 Profile & Events | `features/profiles` (GuestProfile, EventFeed)                 |
| CMP-4 Segments         | `features/segments`                                           |
| CMP-5 Recommendations  | `features/recommendations`                                    |
| CMP-6 Export           | `features/export`                                             |
| CMP-7 Analytics        | `features/dashboard`, `features/tenant`, `features/dq`        |
| CMP-8 Access & Auth    | `features/auth`, `app/` (маршруты), `features/admin` (Users)  |
| CMP-9 Audit            | `features/admin` (AuditLog)                                   |
| CMP-10 Reference       | `features/admin` (справочники, источники)                     |

## Чекпойнт

Создано: дерево React-модулей `app/`, `shared/{api,ui,lib}`, `features/{auth,import,profiles,segments,recommendations,export,dashboard,tenant,dq,admin}`; соответствие CMP-1…CMP-10. Решения: D-14, D-22, D-29.
