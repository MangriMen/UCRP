# 07. CRUD-матрица Роль × Сущность, эндпоинты, проверка прав

Источник ролей эндпоинтов — `endpoint-registry.md`; матрица и реестр обязаны совпадать.

Легенда: **C** create, **R** read, **U** update, **D** delete (для большинства сущностей — soft-delete/деактивация), **A** только обезличенные агрегаты без строк и без `guest_id`, **own** только данные собственного арендатора (`tenant_id` из JWT), **—** нет доступа, **\*** операция выполняется системой по запросу пользователя (пользователь инициирует, запись делает R-6). R-5 не имеет интерфейса и API-доступа: он — объект данных (E-1, E-2, E-10, E-17, E-19) и получатель рассылки, поэтому в столбце R-5 стоят «—» (объект, не субъект). R-6 действует через очередь `job_queue` напрямую через service-слой, минуя REST.

## 1. Матрица

| Сущность                        | R-1 Маркетолог           | R-2 Управляющий                              | R-3 Арендатор | R-4 Аналитик      | R-5 Гость  | R-6 Система             |
| ------------------------------- | ------------------------ | -------------------------------------------- | ------------- | ----------------- | ---------- | ----------------------- |
| E-1 Гость                       | R                        | A                                            | —             | R, D              | — (объект) | C, U                    |
| E-2 Событие                     | R                        | A                                            | A own         | R                 | — (объект) | C, U                    |
| E-3 Сегмент                     | C, R, U, D               | R                                            | —             | C, R, U, D        | —          | U (пересчёт, состав)    |
| E-4 Арендатор                   | R, U (только `priority`) | R                                            | R own         | C, R, U, D        | —          | R                       |
| E-22 ТЦ                         | R                        | R                                            | —             | C, R, U, D        | —          | R                       |
| E-5 Источник данных             | R                        | R                                            | —             | C, R, U, D        | —          | R, U (`last_import_at`) |
| E-6 Категория арендатора        | R                        | R                                            | —             | C, R, U, D        | —          | R                       |
| E-7 Проект                      | R                        | R                                            | —             | C, R, U, D        | —          | R                       |
| E-8 Кампания                    | C, R, U, D               | R                                            | —             | R                 | —          | R, U (статус)           |
| E-9 Тип события                 | R                        | R                                            | —             | C, R, U, D        | —          | R                       |
| E-10 Идентификатор гостя        | R                        | —                                            | —             | R                 | — (объект) | C, U                    |
| E-11 Кандидат на объединение    | R                        | A                                            | —             | C, R, U           | —          | C, U                    |
| E-12 Запрет объединения         | —                        | —                                            | —             | C, R, D           | —          | R                       |
| E-13 Загрузка                   | —                        | R                                            | —             | C, R              | —          | C (pull), U             |
| E-14 Отклонённая строка         | —                        | A                                            | —             | R                 | —          | C                       |
| E-15 Рекомендация               | R, U\*                   | —                                            | —             | R, U\*            | —          | C, U, D (замена набора) |
| E-16 Выгрузка                   | C, R                     | R                                            | —             | R                 | —          | U (статус, счётчики)    |
| E-17 Член выгрузки              | R                        | A                                            | —             | A                 | — (объект) | C                       |
| E-18 Лимит контактов            | C, R, U, D               | R                                            | —             | R                 | —          | R                       |
| E-19 Контакт                    | R                        | A                                            | —             | R                 | — (объект) | C                       |
| E-20 Пользователь               | R own                    | C (приглашение), R, U (имя), D (деактивация) | R own         | R own             | —          | —                       |
| E-21 Аудит                      | —                        | R                                            | —             | R                 | —          | C                       |
| T-1 Задание очереди (служебная) | —                        | —                                            | —             | —                 | —          | C, R, U, D              |
| T-2 Политика доступа и гранты   | —                        | C, R, U (гранты арендаторам)                 | —             | R (аудит доступа) | —          | R                       |
| T-3 Приглашение                 | —                        | C, R, D (отзыв)                              | —             | —                 | —          | —                       |

Ячейки A у R-2 для E-11, E-14, E-17, E-19 (и у R-4 для E-17) — доли и счётчики в `GET /dashboard/metrics` и `GET /data-quality/sources*` без строк, пар профилей и `guest_id`; строки этих сущностей R-2 не читает.

Особенности: D для E-1 (R-4) — soft-delete (`guest.status='deleted'`); D для E-3 — архивирование; D для E-4/E-5/E-20 — деактивация; физическое удаление запрещено для E-2, E-11, E-12, E-21 (аудируемые данные). В E-11 «U» у R-4 = решения confirm/reject/revert; «C» у R-4 = ручное объединение `POST /guests/{id}/merge`. E-15 «U\*» у R-1/R-4 = `POST /guests/{id}/recommendations/refresh` (запись делает сервис рекомендаций).

## 2. Эндпоинты по сущностям (сверка с реестром)

| Сущность | Эндпоинты (операция: роли)                                                                                                                                                                                                                                                                       |
| -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| E-1      | R: `GET /guests`, `GET /guests/{id}` (R-1, R-4); D: `DELETE /guests/{id}` (R-4); A: `GET /dashboard/cross-tenant` (R-2, R-1, R-4)                                                                                                                                                                |
| E-2      | R: `GET /guests/{id}/events` (R-1, R-4); A: `GET /dashboard/cross-tenant`, `GET /dashboard/metrics`, `GET /tenants/{id}/aggregates` (R-3 own)                                                                                                                                                    |
| E-3      | `GET /segment-templates`, `GET /segments`, `GET /segments/{id}`, `/members`, `/versions` (R-1, R-4; R-2 — `GET /segments`, `GET /segments/{id}`); C: `POST /segments`, `POST /segments/preview`; U: `PUT /segments/{id}`, `POST /segments/{id}/recompute`; D: `DELETE /segments/{id}` (R-1, R-4) |
| E-22     | `GET /malls` (R-1, R-2, R-4); C/U/D: `POST`, `PATCH /malls/{id}`, `DELETE /malls/{id}` (R-4); карта передвижения гостя — `GET /guests/{id}/movement` (R-1, R-2, R-4)                                                                                                                             |
| E-4      | `GET /tenants`, `GET /tenants/{id}` (R-1, R-2, R-4; R-3 own); C/D: `POST`, `DELETE` (R-4); U: `PATCH /tenants/{id}` (R-4; R-1 — `priority`)                                                                                                                                                      |
| E-5      | `GET /data-sources`, `GET /data-sources/{id}`, `GET /data-source-types` (R-1, R-2, R-4); CUD: `POST/PATCH/DELETE /data-sources`, `POST/PATCH /data-source-types` (R-4)                                                                                                                           |
| E-6      | R: `GET /tenant-categories`, `GET .../relations` (R-1, R-2, R-4); CUD: `POST/PATCH/DELETE`, `PUT .../relations` (R-4)                                                                                                                                                                            |
| E-7      | R: `GET /projects` (R-1, R-2, R-4); CUD: `POST/PATCH/DELETE /projects` (R-4)                                                                                                                                                                                                                     |
| E-8      | R: `GET /campaigns`, `GET /campaigns/{id}` (R-1, R-2, R-4), `GET /campaigns/{id}/results` (R-1, R-2); CUD: `POST/PATCH/DELETE` (R-1)                                                                                                                                                             |
| E-9      | R: `GET /event-types` (R-1, R-2, R-4); CUD: `POST/PATCH/DELETE` (R-4)                                                                                                                                                                                                                            |
| E-10     | R: в составе `GET /guests/{id}` (R-1, R-4)                                                                                                                                                                                                                                                       |
| E-11     | R: `GET /merge-candidates`, `GET /merge-candidates/{id}` (R-4, R-1); C: `POST /guests/{id}/merge` (R-4); U: `/confirm`, `/reject`, `/revert` (R-4); A: `GET /dashboard/metrics`, `GET /data-quality/sources*` (R-2)                                                                              |
| E-12     | R: `GET /merge-blocks` (R-4); D: `DELETE /merge-blocks/{id}` (R-4); C: побочный эффект `/reject`, `/revert`                                                                                                                                                                                      |
| E-13     | R: `GET /imports`, `GET /imports/{id}` (R-4, R-2); C: `POST /imports` (R-4), `POST /data-sources/{id}/sync` (R-4)                                                                                                                                                                                |
| E-14     | R: `GET /imports/{id}/rejected-rows` (R-4); A: `GET /data-quality/sources*` (R-4, R-2)                                                                                                                                                                                                           |
| E-15     | R: `GET /guests/{id}/recommendations`, `GET /segments/{id}/recommendations` (R-1, R-4); U\*: `POST /guests/{id}/recommendations/refresh` (R-1, R-4)                                                                                                                                              |
| E-16     | C: `POST /exports`, `POST /exports/preview` (R-1); R: `GET /exports`, `GET /exports/{id}` (R-1, R-2, R-4)                                                                                                                                                                                        |
| E-17     | R: `GET /exports/{id}/members`, `GET /exports/{id}/download` (R-1); A: `GET /dashboard/metrics` (R-2, R-4), `GET /campaigns/{id}/results` (R-2)                                                                                                                                                  |
| E-18     | R: `GET /contact-limits` (R-1, R-2, R-4); CUD: `POST/PATCH/DELETE` (R-1)                                                                                                                                                                                                                         |
| E-19     | R: `GET /guests/{id}/contacts` (R-1, R-4); A: `GET /dashboard/metrics` (R-2)                                                                                                                                                                                                                     |
| E-20     | R own: `GET /auth/me` (все), `POST /auth/login`, `/auth/refresh`, `/auth/logout`; R-2: `GET /users`, `GET /users/{id}`, `PATCH /users/{id}` (имя), `DELETE /users/{id}` (деактивация), `POST /users/invitations`                                                                                 |
| E-21     | R: `GET /audit-logs` (R-2, R-4); C: во всех изменяющих эндпоинтах (service → `AuditRepo.Insert`)                                                                                                                                                                                                 |
| T-1      | нет REST                                                                                                                                                                                                                                                                                         |
| T-2      | R-2: `/access/permissions`, `/access/presets`, `/access/grants*`, `/access/effective`, `/access/simulate`; R-2, R-4: `GET /access/audit`                                                                                                                                                         |
| T-3      | R-2: `POST /users/invitations`; pub: `POST /auth/invitations/accept`                                                                                                                                                                                                                             |

## 3. Проверка прав: слои

Двойная защита: backend — единственный источник истины, frontend — лишь удобство и снижение шума.

| Слой                                           | Механизм                                                                                                                                                    | Что проверяется                                                   |
| ---------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------- |
| Backend: middleware (`internal/platform/auth`) | JWT-проверка → контекст `Principal{user_id, role, tenant_id}` в `context.Context`                                                                           | 401 при отсутствии/просрочке токена; `request_id`                 |
| Backend: handler                               | `RequireRole(roles...)` из реестра для каждого маршрута                                                                                                     | 403, если роль не в списке реестра                                |
| Backend: access                                | `access.Decide`: разрешение в потолке роли, для R-3 — активный грант с областью своего арендатора, пределы (период, строки, порог k, маска полей)           | 403 `forbidden_permission`, `forbidden_scope`, `forbidden_tenant` |
| Backend: service                               | Политика сущности: операции по полям (R-1 меняет в `tenant` только `priority`), состояние (409), для R-3 сверка `claims.tenant_id == id`                    | Бизнес-права, изоляция R-3                                        |
| Backend: repository                            | Для R-3 — отдельные методы `...ForTenant(ctx, tenantID, ...)` с обязательным `WHERE tenant_id = ?`; общие методы R-3 вызвать не может (по типу `Principal`) | Фильтр на данных                                                  |
| Frontend: роутинг                              | `<RequireRole roles=[…]>` на каждом маршруте; роль из `GET /auth/me`; редирект на «Нет доступа»                                                             | Скрытие недоступных экранов                                       |
| Frontend: меню и действия                      | Меню строится из таблицы маршрутов; кнопки (Выгрузить, Объединить, Удалить) скрываются по роли                                                              | UX; не защита                                                     |
| Frontend: API-клиент                           | Перехватчик: 401 → refresh → logout; 403 → экран «Нет доступа»; данные других арендаторов запросить нечем (нет URL в клиенте для R-3)                       | Согласование с backend                                            |

| Маршрут React                                                                            | Страница (`pages/`)                                   | Роли                                       | Backend-эндпоинты                      |
| ---------------------------------------------------------------------------------------- | ----------------------------------------------------- | ------------------------------------------ | -------------------------------------- |
| `/login`                                                                                 | `login`                                               | публичный                                  | `POST /auth/login`                     |
| `/import`, `/import/:id`                                                                 | `imports`, `import-details`                           | R-4 (R-2 — только чтение списка)           | `/imports*`                            |
| `/guests`, `/guests/:id`                                                                 | `guests`, `guest-profile`                             | R-1, R-4                                   | `/guests*`                             |
| `/identity/merge-candidates`, `/identity/merge-candidates/:id`, `/identity/merge-blocks` | `merge-candidates`, `merge-candidate`, `merge-blocks` | R-4 (R-1 — чтение)                         | `/merge-candidates*`, `/merge-blocks*` |
| `/segments`, `/segments/:id`                                                             | `segments`, `segment-details`                         | R-1, R-4 (R-2 — чтение)                    | `/segments*`                           |
| `/exports`                                                                               | `exports`                                             | R-1 (R-2, R-4 — чтение журнала)            | `/exports*`                            |
| `/campaigns`, `/campaigns/:id`                                                           | `campaigns`, `campaign-details`                       | R-1 (R-2, R-4 — чтение)                    | `/campaigns*`                          |
| `/settings/contact-limits`                                                               | `contact-limits`                                      | R-1 (R-2, R-4 — чтение)                    | `/contact-limits*`                     |
| `/data-sources`                                                                          | `data-sources`                                        | R-4 (R-1, R-2 — чтение)                    | `/data-sources*`                       |
| `/dashboard`                                                                             | `dashboard`                                           | R-2, R-1, R-4                              | `/dashboard/*`                         |
| `/data-quality`                                                                          | `data-quality`                                        | R-4, R-2                                   | `/data-quality/*`                      |
| `/tenant`                                                                                | `tenant-aggregates`                                   | R-3 (и R-2, R-1, R-4 с выбором арендатора) | `GET /tenants/{id}/aggregates`         |
| `/admin/users`                                                                           | `admin-users`                                         | R-2                                        | `/users*`                              |
| `/admin/dictionaries`                                                                    | `admin-dictionaries`                                  | R-4 (R-1, R-2 — чтение)                    | справочники                            |
| `/admin/audit`                                                                           | `audit-log`                                           | R-2, R-4                                   | `GET /audit-logs`                      |

Правило согласования: роли в `RequireRole` ⊆ ролей эндпоинтов реестра для главной операции экрана; расхождение — дефект .

## 4. Строка R-3: что видит арендатор

Арендатор получает только то, что разрешено грантом (раздел 7 `05-er-data-model.md`). Без грантов доступны вход, собственная карточка и агрегаты своего арендатора.

| Что        | Видит                                                                                                                                                   | Не видит                                                                     |
| ---------- | ------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------- |
| Профиль    | Своя запись `tenant` (название, категория, этаж)                                                                                                        | Другие арендаторы, их приоритеты, источники                                  |
| Показатели | Агрегаты своего арендатора; разрешённые метрики (например, число событий присутствия, новые гости через комплекс); значения < k (по умолчанию 5) скрыты | Любые данные по другим арендаторам, общие распределения по комплексу         |
| Гости      | По гранту — `GET /guests/{id}` и `/events` только по данным собственного арендатора, с маской полей и лимитом строк                                     | Данные гостей других арендаторов, сегменты, рекомендации, выгрузки, контакты |
| Эндпоинты  | `/auth/*`, `GET /tenants` (только своя запись), `GET /tenants/{id}`, `GET /tenants/{id}/aggregates` и разрешённые грантом операции чтения гостей        | Всё остальное из реестра → 403                                               |
| UI         | Экран `/tenant`; меню без прочих разделов; переключателя арендатора нет                                                                                 | Навигация и ссылки на чужие страницы отсутствуют                             |

Механизм изоляции:

1. Токен R-3 содержит `tenant_id` (создаётся только вместе с `app_user.tenant_id`; CHECK в БД).
2. `access.Decide` отклоняет запрос с `forbidden_tenant`, если `tenant_id` ресурса не совпадает с токеном, и `forbidden_permission`/`forbidden_scope`, если нет подходящего гранта.
3. Repository: методы `...ForTenant` всегда подставляют `tenant_id` из токена и не возвращают данных другого арендатора.
4. Малые числа (< k) скрываются, чтобы нельзя было вычислить отдельного гостя (Q-18).
5. Тесты безопасности: запрос R-3 ко всем эндпоинтам реестра, кроме разрешённых, → 403; запрос с чужим `id` → 403; ответ не содержит данных другого арендатора.

## 5. Проверка прав по данным R-6 и R-5

| Субъект | Правило                                                                                                                                                                                       |
| ------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| R-6     | Нет входа в REST; воркеры запускаются процессом `cmd/worker` с системным `Principal{role: system}`; все записи воркера аудируются (`actor_user_id = NULL`) для объединения, импорта, выгрузки |
| R-5     | Прямых прав нет; данные — предмет защиты: выгрузка содержит только контакт нужного канала, лимит контактов защищает от спама; удаление гостя — только R-4 (soft-delete)                       |
