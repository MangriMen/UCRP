# 07. CRUD-матрица Роль × Сущность, эндпоинты, проверка прав

Источник ролей эндпоинтов — `appendix/endpoint-registry.md`; матрица и реестр обязаны совпадать (проверка в разделе 6).

Легенда: **C** create, **R** read, **U** update, **D** delete (для большинства сущностей — soft-delete/деактивация), **A** только обезличенные агрегаты без строк и без `guest_id`, **own** только данные собственного арендатора (`tenant_id` из JWT), **—** нет доступа, **\*** операция выполняется системой по запросу пользователя (пользователь инициирует, запись делает R-6). R-5 не имеет интерфейса и API-доступа: он — объект данных (E-1, E-2, E-10, E-17, E-19) и получатель рассылки, поэтому в столбце R-5 стоят «—» (объект, не субъект). R-6 действует через очередь `job_queue` напрямую через service-слой, минуя REST.

## 1. Матрица

| Сущность                        | R-1 Маркетолог           | R-2 Управляющий | R-3 Арендатор | R-4 Аналитик | R-5 Гость  | R-6 Система             |
| ------------------------------- | ------------------------ | --------------- | ------------- | ------------ | ---------- | ----------------------- |
| E-1 Гость                       | R                        | A               | —             | R, D         | — (объект) | C, U                    |
| E-2 Событие                     | R                        | A               | A own         | R            | — (объект) | C, U                    |
| E-3 Сегмент                     | C, R, U, D               | R               | —             | C, R, U, D   | —          | U (пересчёт, состав)    |
| E-4 Арендатор                   | R, U (только `priority`) | R               | R own         | C, R, U, D   | —          | R                       |
| E-5 Источник данных             | R                        | R               | —             | C, R, U, D   | —          | R, U (`last_import_at`) |
| E-6 Категория арендатора        | R                        | R               | —             | C, R, U, D   | —          | R                       |
| E-7 Проект                      | R                        | R               | —             | C, R, U, D   | —          | R                       |
| E-8 Кампания                    | C, R, U, D               | R               | —             | R            | —          | R, U (статус)           |
| E-9 Тип события                 | R                        | R               | —             | C, R, U, D   | —          | R                       |
| E-10 Идентификатор гостя        | R                        | —               | —             | R            | — (объект) | C, U                    |
| E-11 Кандидат на склейку        | R                        | A               | —             | C, R, U      | —          | C, U                    |
| E-12 Запрет склейки             | —                        | —               | —             | C, R, D      | —          | R                       |
| E-13 Загрузка                   | —                        | R               | —             | C, R         | —          | C (pull), U             |
| E-14 Отклонённая строка         | —                        | A               | —             | R            | —          | C                       |
| E-15 Рекомендация               | R, U\*                   | —               | —             | R, U\*       | —          | C, U, D (замена набора) |
| E-16 Выгрузка                   | C, R                     | R               | —             | R            | —          | U (статус, счётчики)    |
| E-17 Член выгрузки              | R                        | A               | —             | A            | — (объект) | C                       |
| E-18 Лимит контактов            | C, R, U, D               | R               | —             | R            | —          | R                       |
| E-19 Контакт                    | R                        | A               | —             | R            | — (объект) | C                       |
| E-20 Пользователь               | R own                    | C, R, U, D      | R own         | R own        | —          | —                       |
| E-21 Аудит                      | —                        | R               | —             | R            | —          | C                       |
| T-1 Задание очереди (служебная) | —                        | —               | —             | —            | —          | C, R, U, D              |

Ячейки A у R-2 для E-11, E-14, E-17, E-19 (и у R-4 для E-17) — доли и счётчики в `GET /dashboard/metrics` и `GET /data-quality/sources*` без строк, пар профилей и `guest_id`; строки этих сущностей R-2 не читает.

Особенности: D для E-1 (R-4) — soft-delete (`guest.status='deleted'`); D для E-3 — архивирование; D для E-4/E-5/E-20 — деактивация; физическое удаление запрещено для E-2, E-11, E-12, E-21 (аудируемые данные). В E-11 «U» у R-4 = решения confirm/reject/revert; «C» у R-4 = ручное объединение `POST /guests/{id}/merge`. E-15 «U\*» у R-1/R-4 = `POST /guests/{id}/recommendations/refresh` (запись делает сервис рекомендаций).

## 2. Эндпоинты по сущностям (сверка с реестром)

| Сущность | Эндпоинты (операция: роли)                                                                                                                                                                                                                                                                       |
| -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| E-1      | R: `GET /guests`, `GET /guests/{id}` (R-1, R-4); D: `DELETE /guests/{id}` (R-4); A: `GET /dashboard/penetration` (R-2, R-1, R-4)                                                                                                                                                                 |
| E-2      | R: `GET /guests/{id}/events` (R-1, R-4); A: `GET /dashboard/penetration`, `GET /dashboard/metrics`, `GET /tenants/{id}/aggregates` (R-3 own)                                                                                                                                                     |
| E-3      | `GET /segment-templates`, `GET /segments`, `GET /segments/{id}`, `/members`, `/versions` (R-1, R-4; R-2 — `GET /segments`, `GET /segments/{id}`); C: `POST /segments`, `POST /segments/preview`; U: `PUT /segments/{id}`, `POST /segments/{id}/recompute`; D: `DELETE /segments/{id}` (R-1, R-4) |
| E-4      | `GET /tenants`, `GET /tenants/{id}` (R-1, R-2, R-4; R-3 own); C/D: `POST`, `DELETE` (R-4); U: `PATCH /tenants/{id}` (R-4; R-1 — `priority`)                                                                                                                                                      |
| E-5      | `GET /data-sources`, `GET /data-sources/{id}`, `GET /data-source-types` (R-1, R-2, R-4); CUD: `POST/PATCH/DELETE /data-sources`, `POST/PATCH /data-source-types` (R-4)                                                                                                                           |
| E-6      | R: `GET /tenant-categories`, `GET .../relations` (R-1, R-2, R-4); CUD: `POST/PATCH/DELETE`, `PUT .../relations` (R-4)                                                                                                                                                                            |
| E-7      | R: `GET /projects` (R-1, R-2, R-4); CUD: `POST/PATCH/DELETE /projects` (R-4)                                                                                                                                                                                                                     |
| E-8      | R: `GET /campaigns`, `GET /campaigns/{id}` (R-1, R-2, R-4), `GET /campaigns/{id}/results` (R-1, R-2); CUD: `POST/PATCH/DELETE` (R-1)                                                                                                                                                             |
| E-9      | R: `GET /event-types` (R-1, R-2, R-4); CUD: `POST/PATCH/DELETE` (R-4)                                                                                                                                                                                                                            |
| E-10     | R: в составе `GET /guests/{id}` (R-1, R-4)                                                                                                                                                                                                                                                       |
| E-11     | R: `GET /merge-candidates`, `GET /merge-candidates/{id}` (R-4, R-1); C: `POST /guests/{id}/merge` (R-4); U: `/confirm`, `/reject`, `/revert` (R-4); A: `GET /dashboard/metrics`, `GET /data-quality/sources*` (R-2)                                                                              |
| E-12     | R: `GET /unmerge-blocks` (R-4); D: `DELETE /unmerge-blocks/{id}` (R-4); C: побочный эффект `/reject`, `/revert`                                                                                                                                                                                  |
| E-13     | R: `GET /imports`, `GET /imports/{id}` (R-4, R-2); C: `POST /imports` (R-4), `POST /data-sources/{id}/sync` (R-4)                                                                                                                                                                                |
| E-14     | R: `GET /imports/{id}/rejected-rows` (R-4); A: `GET /data-quality/sources*` (R-4, R-2)                                                                                                                                                                                                           |
| E-15     | R: `GET /guests/{id}/recommendations`, `GET /segments/{id}/recommendations` (R-1, R-4); U\*: `POST /guests/{id}/recommendations/refresh` (R-1, R-4)                                                                                                                                              |
| E-16     | C: `POST /exports`, `POST /exports/preview` (R-1); R: `GET /exports`, `GET /exports/{id}` (R-1, R-2, R-4)                                                                                                                                                                                        |
| E-17     | R: `GET /exports/{id}/members`, `GET /exports/{id}/download` (R-1); A: `GET /dashboard/metrics` (R-2, R-4), `GET /campaigns/{id}/results` (R-2)                                                                                                                                                  |
| E-18     | R: `GET /contact-limits` (R-1, R-2, R-4); CUD: `POST/PATCH/DELETE` (R-1)                                                                                                                                                                                                                         |
| E-19     | R: `GET /guests/{id}/contacts` (R-1, R-4); A: `GET /dashboard/metrics` (R-2)                                                                                                                                                                                                                     |
| E-20     | R own: `GET /auth/me` (все), CRUD: `/users` (R-2), `/auth/login                                                                                                                                                                                                                                  | refresh | logout` (вход) |
| E-21     | R: `GET /audit-logs` (R-2, R-4); C: во всех изменяющих эндпоинтах (service → `AuditRepo.Insert`)                                                                                                                                                                                                 |
| T-1      | нет REST                                                                                                                                                                                                                                                                                         |

## 3. Проверка прав: слои

Двойная защита: backend — единственный источник истины, frontend — лишь удобство и снижение шума.

| Слой                                           | Механизм                                                                                                                                                    | Что проверяется                                   |
| ---------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------- |
| Backend: middleware (`internal/platform/auth`) | JWT-проверка → контекст `Principal{user_id, role, tenant_id}` в `context.Context`                                                                           | 401 при отсутствии/просрочке токена; `request_id` |
| Backend: handler                               | `RequireRole(roles...)` из реестра для каждого маршрута                                                                                                     | 403, если роль не в списке реестра                |
| Backend: service                               | Политика сущности: операции по полям (R-1 меняет в `tenant` только `priority`), состояние (409), для R-3 сверка `claims.tenant_id == id`                    | Бизнес-права, изоляция R-3                        |
| Backend: repository                            | Для R-3 — отдельные методы `...ForTenant(ctx, tenantID, ...)` с обязательным `WHERE tenant_id = ?`; общие методы R-3 вызвать не может (по типу `Principal`) | Фильтр на данных                                  |
| Frontend: роутинг                              | `<RequireRole roles=[…]>` на каждом маршруте; роль из `GET /auth/me`; редирект на «Нет доступа»                                                             | Скрытие недоступных экранов                       |
| Frontend: меню и действия                      | Меню строится из таблицы маршрутов; кнопки (Выгрузить, Объединить, Удалить) скрываются по роли                                                              | UX; не защита                                     |
| Frontend: API-клиент                           | Перехватчик: 401 → refresh → logout; 403 → экран «Нет доступа»; данные других арендаторов запросить нечем (нет URL в клиенте для R-3)                       | Согласование с backend                            |

| Маршрут React                                 | Модуль               | Роли в `RequireRole`                      | Backend-эндпоинты (роли совпадают)       |
| --------------------------------------------- | -------------------- | ----------------------------------------- | ---------------------------------------- |
| `/login`                                      | `auth/`              | публичный                                 | `POST /auth/login`                       |
| `/import`, `/import/:batchId`                 | `features/import`    | R-4 (R-2 — только чтение списка)          | `/imports*` R-4; `GET` списка — R-4, R-2 |
| `/guests`, `/guests/:guestId`                 | `features/profiles`  | R-1, R-4                                  | `/guests*` R-1, R-4                      |
| `/identity/queue`                             | `features/profiles`  | R-4 (R-1 — чтение)                        | `/merge-candidates*`, `/unmerge-blocks*` |
| `/segments`, `/segments/new`, `/segments/:id` | `features/segments`  | R-1, R-4 (R-2 — чтение списка и карточки) | `/segments*`                             |
| `/exports`, `/exports/new`, `/exports/:id`    | `features/export`    | R-1 (R-2, R-4 — чтение журнала)           | `/exports*`                              |
| `/campaigns`                                  | `features/export`    | R-1 (R-2, R-4 — чтение)                   | `/campaigns*`                            |
| `/settings/contact-limits`                    | `features/export`    | R-1 (R-2, R-4 — чтение)                   | `/contact-limits*`                       |
| `/dashboard`                                  | `features/dashboard` | R-2, R-1, R-4                             | `/dashboard/*`                           |
| `/data-quality`                               | `features/dq`        | R-4, R-2                                  | `/data-quality/*`                        |
| `/tenant`                                     | `features/tenant`    | R-3 (и R-2/R-1/R-4 с выбором арендатора)  | `GET /tenants/{id}/aggregates`           |
| `/admin/users`                                | `features/admin`     | R-2                                       | `/users*`                                |
| `/admin/dictionaries`                         | `features/admin`     | R-4 (R-1/R-2 — чтение)                    | справочники                              |
| `/admin/audit`                                | `features/admin`     | R-2, R-4                                  | `GET /audit-logs`                        |

Правило согласования: роли в `RequireRole` ⊆ ролей эндпоинтов реестра для главной операции экрана; расхождение — дефект (проверяется в фазе E).

## 4. Строка R-3: что видит арендатор

| Что        | Видит                                                                                                                                                                               | Не видит                                                                                                                                                  |
| ---------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Профиль    | Своя запись `tenant` (название, категория, этаж)                                                                                                                                    | Другие арендаторы, их приоритеты, источники                                                                                                               |
| Показатели | Агрегаты своего арендатора: число новых гостей, пришедших через комплекс, динамика по месяцам; значения < k = 5 скрыты                                                              | Любые данные по другим арендаторам, общие распределения по комплексу, списки и идентификаторы гостей, события, сегменты, рекомендации, выгрузки, контакты |
| Эндпоинты  | `POST /auth/login`, `POST /auth/refresh`, `POST /auth/logout`, `GET /auth/me`, `GET /tenants` (только своя запись), `GET /tenants/{id}` (own), `GET /tenants/{id}/aggregates` (own) | Всё остальное из реестра → 403                                                                                                                            |
| UI         | Один экран `/tenant`; меню без прочих разделов; переключателя арендатора нет                                                                                                        | Навигация и ссылки на чужие страницы отсутствуют                                                                                                          |

Механизм изоляции (service + repository):

1. Токен R-3 содержит `tenant_id` (создаётся только вместе с `app_user.tenant_id`; CHECK в БД).
2. Handler передаёт `Principal` в service; `TenantAggregateService.Get(ctx, p, id)` при `p.Role == tenant && p.TenantID != id` возвращает `ErrForbidden` → 403 без обращения к БД.
3. Repository: `TenantAggregateRepo.NewGuestsForTenant(ctx, tenantID, period)` всегда подставляет `tenant_id` параметром и возвращает только числа; методов, возвращающих `guest_id` или данные другого арендатора, в этом интерфейсе нет.
4. Малые числа (< 5) скрываются, чтобы нельзя было вычислить отдельного гостя (Q-18).
5. Тесты безопасности (для фазы E): запрос R-3 ко всем эндпоинтам реестра, кроме разрешённых, → 403; запрос с чужим `id` → 403; ответ не содержит `guest_id`.

## 5. Проверка прав по данным R-6 и R-5

| Субъект | Правило                                                                                                                                                                                   |
| ------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| R-6     | Нет входа в REST; воркеры запускаются процессом `cmd/worker` с системным `Principal{role: system}`; все записи воркера аудируются (`actor_user_id = NULL`) для склейки, импорта, выгрузки |
| R-5     | Прямых прав нет; данные — предмет защиты: выгрузка содержит только контакт нужного канала, лимит контактов защищает от спама; удаление гостя — только R-4 (soft-delete)                   |

## 6. Самопроверка согласованности с реестром

| Проверка                                                                                                           | Результат         |
| ------------------------------------------------------------------------------------------------------------------ | ----------------- |
| Каждая сущность E-1…E-21 имеет строку в матрице                                                                    | OK (21 + T-1)     |
| Каждая операция из матрицы имеет эндпоинт или объяснена (побочный эффект / система)                                | OK (см. раздел 2) |
| R-3 в реестре встречается только в: `/auth/*`, `GET /tenants`, `GET /tenants/{id}`, `GET /tenants/{id}/aggregates` | OK                |
| R-1 не имеет операций импорта, управления склейками, пользователями, справочниками (кроме `priority`)              | OK                |
| R-2 не изменяет данные, кроме пользователей                                                                        | OK                |

## Чекпойнт

Создано: матрица R-1…R-6 × E-1…E-21 + T-1, сверка по эндпоинтам, схема слоёв прав, таблица маршрутов React, строка R-3 с механизмом изоляции. Открытые: Q-16, Q-18; решения: D-10, D-13, D-16.
