# 03. Матрица тестового покрытия

Основа: `docs/14-traceability.md` (US/UC → AD → SD → ER → CRUD → CMP → API).
Колонки: требование → UC → US → endpoint → тест → тип проверки.
Типы: **P** позитивный, **N** негативный, **C** контрактный (JSON Schema),
**B** бизнес-правило, **E** сквозной E2E, **S** безопасность/роли.

Состояние на коммите `945a5a7`: в `apps/api/e2e/` реализовано 88 тестовых
функций, смонтировано 88 из 90 операций контракта (`02-contradictions.md` §2.11).
Набор запускается против живого API и БД: `./apps/api/e2e/run.ps1`.

**Статус**

| Значение         | Смысл                                                       |
| ---------------- | ----------------------------------------------------------- |
| `OK`             | операция смонтирована, тест реализован и проходит           |
| `OK (без теста)` | операция смонтирована, тест не написан                      |
| `OK (дефект)`    | операция смонтирована, поведение расходится с требованием   |
| `blocked (…)`    | операция смонтирована, тест невозможен по указанной причине |
| `MISS`           | операция отсутствует в `cmd/api/main.go`                    |

## 1. Аутентификация и доступ (UC-16, UC-18)

| Требование                                | UC    | US   | Endpoint             | Тест                                                        | Тип | Статус         |
| ----------------------------------------- | ----- | ---- | -------------------- | ----------------------------------------------------------- | --- | -------------- |
| Вход по email+паролю                      | UC-16 | —    | `POST /auth/login`   | `TestAuthLoginContract`                                     | P,C | OK             |
| Одинаковый 401 на неверный email и пароль | UC-16 | —    | `POST /auth/login`   | `TestAuthLoginRejectsWrongPassword`                         | N   | OK             |
| 400 при пустом теле                       | UC-16 | —    | `POST /auth/login`   | `TestAuthLoginRejectsEmptyBody`                             | N,C | OK             |
| 429 при переборе                          | UC-16 | —    | `POST /auth/login`   | `auth.login.rate-limit`                                     | N   | blocked (G9)   |
| Refresh по httpOnly cookie                | UC-16 | —    | `POST /auth/refresh` | `auth.refresh.rotates-access`                               | P,C | OK (без теста) |
| Отзыв refresh                             | UC-16 | —    | `POST /auth/logout`  | `auth.logout.revokes`                                       | P   | OK (без теста) |
| Профиль, роль, `tenantId`                 | UC-16 | US-8 | `GET /auth/me`       | `TestAuthMeContract`, `TestRoleTenantCannotReachPrivateAPI` | P,C | OK             |
| 401 без токена                            | UC-16 | —    | любой приватный      | `TestAuthMeRequiresToken`                                   | N,S | OK             |
| 401 с истёкшим/битым токеном              | UC-16 | —    | любой приватный      | `TestAuthMeRejectsMalformedToken`                           | N,S | OK             |
| R-2 создаёт пользователя                  | UC-18 | —    | `POST /users`        | `TestUserCreateByManager`                                   | P,C | OK             |
| R-3 требует `tenantId`                    | UC-18 | —    | `POST /users`        | `TestUserCreateTenantRequiresTenantID`                      | N,B | OK             |
| Дубль email → 409                         | UC-18 | —    | `POST /users`        | `TestUserCreateDuplicateEmail`                              | N   | OK (дефект)    |
| Последний R-2 не деактивируется           | UC-18 | —    | `DELETE /users/{id}` | `TestUserDeleteLastManagerReturns409`                       | N,B | OK             |
| R-1/R-3/R-4 не управляют пользователями   | UC-18 | —    | `/users*`            | `TestUsersForbiddenForOtherRoles`                           | N,S | OK             |
| Список и карточка пользователя            | UC-18 | —    | `GET /users`         | `TestUserListAndGetContract`                                | P,C | OK             |

## 2. Справочники и источники (UC-17)

| Требование                               | UC    | US   | Endpoint                                | Тест                                                            | Тип | Статус         |
| ---------------------------------------- | ----- | ---- | --------------------------------------- | --------------------------------------------------------------- | --- | -------------- |
| Создание арендатора                      | UC-17 | US-1 | `POST /tenants`                         | `TestTenantCreateAndPatch`                                      | P,C | OK             |
| Уникальность `code` → 409                | UC-17 | —    | `POST /tenants`                         | `TestTenantCreateDuplicateCode`                                 | N   | OK             |
| R-1 меняет только `priority`             | UC-17 | —    | `PATCH /tenants/{id}`                   | `TestTenantDeleteMarketerForbidden`                             | N,B | OK (дефект)    |
| Удаление арендатора с событиями → 409    | UC-17 | —    | `DELETE /tenants/{id}`                  | `TestTenantDeleteWithEventsReturns409`                          | N,B | OK             |
| R-3 видит только свой арендатор          | UC-8  | US-8 | `GET /tenants`, `GET /tenants/{id}`     | `tenants.r3.own-only`                                           | N,S | OK (дефект)    |
| Смежность/конкуренция категорий          | UC-17 | US-4 | `PUT /tenant-categories/{id}/relations` | `TestCategoryRelationsReplace`                                  | P,B | OK             |
| Типы событий: `top_up`/`spend` не визит  | UC-17 | —    | `POST /event-types`                     | `TestEventTypeCountsAsPresence`                                 | P,B | OK             |
| Источник с маппингом полей               | UC-17 | US-1 | `POST /data-sources`                    | `TestDataSourceCreate`, `TestDataSourceCreateRejectsBadMapping` | P,C | OK             |
| Деактивация источника с загрузками → 409 | UC-17 | —    | `DELETE /data-sources/{id}`             | `data-sources.delete.with-imports`                              | N,B | OK (без теста) |
| Типы источников данных                   | UC-17 | US-1 | `GET /data-source-types`                | `TestDataSourceTypesDefaults`                                   | P,C | OK             |

R-1 меняет только `priority` не выполняется: `PATCH /tenants/{id}` требует роль
`analyst`, маркетологу доступ закрыт целиком. R-3 не видит даже свою карточку
арендатора: `GET /tenants` и `GET /tenants/{id}` открыты только
`marketer`/`manager`/`analyst` (`02-contradictions.md` §2.9).

## 3. Импорт (UC-1, UC-14)

| Требование                           | UC    | US   | Endpoint                                     | Тест                                                               | Тип | Статус         |
| ------------------------------------ | ----- | ---- | -------------------------------------------- | ------------------------------------------------------------------ | --- | -------------- |
| Файл принят → 202 `queued`           | UC-1  | US-1 | `POST /imports`                              | `TestImportUploadAccepted`                                         | P,C | OK             |
| Статус и счётчики                    | UC-1  | US-1 | `GET /imports/{id}`                          | `TestImportCountersInvariant`                                      | P,C | OK             |
| `accepted+rejected+duplicate=total`  | UC-1  | US-1 | `GET /imports/{id}`                          | `TestImportCountersInvariant`                                      | B   | OK             |
| Повтор файла → 200 `duplicate=true`  | UC-1  | US-1 | `POST /imports`                              | `TestImportUploadIsIdempotent`                                     | B   | OK             |
| Дубль `external_id` в другом файле   | UC-1  | US-1 | `POST /imports`                              | `TestImportDuplicateExternalIDCountedAsDuplicate`                  | B   | OK             |
| Строка без `external_id` → rejected  | UC-1  | US-1 | `GET /imports/{id}/rejected-rows`            | `TestImportRejectedRowCodes`                                       | N,B | OK             |
| Плохая дата → rejected               | UC-1  | US-1 | `GET /imports/{id}/rejected-rows`            | `TestImportRejectedRowCodes`                                       | N,B | OK             |
| CSV отклонённых строк                | UC-1  | US-1 | `GET /imports/{id}/rejected-rows?format=csv` | `TestImportRejectedRowsCSV`                                        | P,C | OK             |
| > 100 000 строк → 413                | UC-1  | US-1 | `POST /imports`                              | `imports.upload.too-large`                                         | N   | OK (без теста) |
| Неверный формат → 422                | UC-1  | US-1 | `POST /imports`                              | `imports.upload.bad-format`                                        | N   | OK (без теста) |
| Несуществующий источник → 404        | UC-1  | US-1 | `POST /imports`                              | `TestImportUploadUnknownSource`                                    | N   | OK             |
| Без файла → 400                      | UC-1  | US-1 | `POST /imports`                              | `TestImportUploadRequiresFile`, `TestImportUploadRejectsEmptyFile` | N   | OK             |
| Пропущены обязательные колонки       | UC-1  | US-1 | `POST /imports`                              | `TestImportUploadRejectsMissingColumns`                            | N   | OK             |
| R-1/R-3 → 403                        | UC-1  | US-1 | `POST /imports`                              | `TestRoleMatrix`                                                   | N,S | OK             |
| Список и фильтры загрузок            | UC-1  | US-1 | `GET /imports`                               | `TestImportListFilters`                                            | P,C | OK             |
| Неизвестная загрузка → 404           | UC-1  | —    | `GET /imports/{id}`                          | `TestImportGetUnknownReturns404`                                   | N   | OK             |
| Ручной pull → 202                    | UC-14 | US-1 | `POST /data-sources/{id}/sync`               | `sync.manual.accepted`                                             | P   | OK (без теста) |
| Pull для файлового источника → 400   | UC-14 | —    | `POST /data-sources/{id}/sync`               | `TestDataSourceSyncRejectsFileOnlySource`                          | N   | OK             |
| Повторный pull → 409                 | UC-14 | —    | `POST /data-sources/{id}/sync`               | `sync.manual.in-progress`                                          | N   | OK (без теста) |
| Сквозной импорт → сегмент → выгрузка | UC-1  | US-1 | весь цикл                                    | `TestE2EImportToExport`                                            | E   | OK             |

## 4. Профиль гостя (UC-2)

| Требование                                | UC   | US   | Endpoint                    | Тест                                                 | Тип | Статус      |
| ----------------------------------------- | ---- | ---- | --------------------------- | ---------------------------------------------------- | --- | ----------- |
| Поиск по телефону в разных форматах       | UC-2 | US-2 | `GET /guests?q=`            | `TestGuestSearchFindsByIdentifier`                   | P,B | OK          |
| `q` короче 2 → 400                        | UC-2 | US-2 | `GET /guests?q=`            | `TestGuestsSearchRejectsShortQuery`                  | N,C | OK          |
| `q` обязателен                            | UC-2 | US-2 | `GET /guests?q=`            | `TestGuestsSearchRequiresQuery`                      | N,C | OK          |
| Поглощённый профиль отдаёт `mergedIntoId` | UC-2 | US-2 | `GET /guests/{id}`          | `TestMergeCandidateConfirmMovesEventsAndIdentifiers` | P,B | OK          |
| Несуществующий гость → 404                | UC-2 | US-2 | `GET /guests/{id}`          | `TestGuestGetUnknownReturns404`                      | N   | OK          |
| Лента событий по убыванию даты            | UC-2 | US-2 | `GET /guests/{id}/events`   | `TestGuestEventsSortedDescending`                    | P,B | OK          |
| Фильтры `from`/`to`/`tenant_id`           | UC-2 | US-2 | `GET /guests/{id}/events`   | `TestGuestEventsContract`                            | P   | OK (дефект) |
| `tenantCount` и список арендаторов        | UC-2 | US-2 | `GET /guests/{id}`          | `TestGuestProfileReflectsTenantCount`                | P,B | OK          |
| История контактов                         | UC-2 | US-2 | `GET /guests/{id}/contacts` | `TestGuestContactsList`                              | P,C | OK          |
| Soft-delete гостя                         | UC-2 | US-2 | `DELETE /guests/{id}`       | `TestGuestSoftDelete`                                | P,B | OK          |
| R-3 → 403                                 | UC-2 | US-2 | `/guests*`                  | `TestRoleTenantCannotReachPrivateAPI`                | N,S | OK          |

## 5. Сегменты (UC-3, UC-6, UC-13)

| Требование                                                          | UC    | US   | Endpoint                        | Тест                                                                         | Тип | Статус         |
| ------------------------------------------------------------------- | ----- | ---- | ------------------------------- | ---------------------------------------------------------------------------- | --- | -------------- |
| Preview: размер и примеры, без записи                               | UC-3  | US-3 | `POST /segments/preview`        | `TestSegmentPreviewCountsMatchingGuests`, `TestSegmentPreviewDoesNotPersist` | P,B | OK             |
| Preview: арендатор в `visitedTenantIds` и `excludedTenantIds` → 400 | UC-3  | US-3 | `POST /segments/preview`        | `TestSegmentPreviewRejectsConflictingTenant`                                 | N,B | OK             |
| Preview: пустой `visitedTenantIds` → 400                            | UC-3  | US-3 | `POST /segments/preview`        | `TestSegmentPreviewRejectsEmptyWas`                                          | N   | OK             |
| Preview: пустой `excludedTenantIds` допустим                        | UC-3  | US-3 | `POST /segments/preview`        | `TestSegmentPreviewAcceptsEmptyExcludedTenantIds`                            | P   | OK             |
| Preview: таймаут → 504                                              | UC-3  | US-3 | `POST /segments/preview`        | `segments.preview.timeout`                                                   | N   | blocked (G10)  |
| Создание сегмента, версия 1                                         | UC-3  | US-3 | `POST /segments`                | `TestSegmentCreatePersistsMembership`                                        | P,C | OK             |
| Создание без `name` → 400                                           | UC-3  | US-3 | `POST /segments`                | `TestSegmentCreateRequiresName`                                              | N   | OK             |
| Список сегментов                                                    | UC-3  | US-3 | `GET /segments`                 | `TestSegmentsListContract`                                                   | P,C | OK             |
| Карточка сегмента с готовой версией                                 | UC-3  | US-3 | `GET /segments/{id}`            | `TestSegmentGetReportsReadyVersion`                                          | P,C | OK             |
| Правка правила → версия 2                                           | UC-3  | US-3 | `PUT /segments/{id}`            | `TestSegmentUpdateCreatesVersion`                                            | P,B | OK             |
| История версий                                                      | UC-3  | US-3 | `GET /segments/{id}/versions`   | `TestSegmentUpdateCreatesVersion`                                            | P,C | OK             |
| Состав сегмента постранично                                         | UC-3  | US-3 | `GET /segments/{id}/members`    | `TestSegmentMembersPaged`                                                    | P,C | OK             |
| 5 шаблонов                                                          | UC-3  | US-3 | `GET /segment-templates`        | `TestSegmentTemplatesFive`                                                   | P,C | OK             |
| Архив сегмента                                                      | UC-3  | US-3 | `DELETE /segments/{id}`         | `TestSegmentArchiveRemovesItFromActiveList`                                  | P   | OK             |
| Архив при активной кампании → 409                                   | UC-3  | —    | `DELETE /segments/{id}`         | `segments.delete.active-campaign`                                            | N,B | OK (без теста) |
| Пересчёт → 202                                                      | UC-13 | US-6 | `POST /segments/{id}/recompute` | `TestSegmentRecomputeAcceptsAndQueues`                                       | P   | OK             |
| Повторный пересчёт → 409                                            | UC-13 | —    | `POST /segments/{id}/recompute` | `segments.recompute.in-progress`                                             | N   | OK (без теста) |
| `computedAt` обновляется после события                              | UC-6  | US-6 | `GET /segments/{id}`            | `TestE2ENewEventChangesSegmentAndRecommendations`                            | E,B | OK             |
| R-3 → 403                                                           | UC-3  | US-3 | `/segments*`                    | `TestRoleTenantCannotReachPrivateAPI`                                        | N,S | OK             |

## 6. Рекомендации (UC-4, UC-15)

| Требование                          | UC    | US   | Endpoint                                    | Тест                                                            | Тип | Статус         |
| ----------------------------------- | ----- | ---- | ------------------------------------------- | --------------------------------------------------------------- | --- | -------------- |
| ≤ 3 рекомендаций с обоснованием     | UC-4  | US-4 | `GET /guests/{id}/recommendations`          | `TestRecommendationsContract`, `TestE2ESegmentToRecommendation` | P,B | OK             |
| Посещённые арендаторы отсутствуют   | UC-4  | US-4 | `GET /guests/{id}/recommendations`          | `TestRecommendationsExcludeVisitedTenants`                      | B   | OK             |
| Конкурент присутствует с понижением | UC-4  | US-4 | `GET /guests/{id}/recommendations`          | `reco.list.competitor-penalized`                                | B   | blocked (D1)   |
| Fallback для гостя без истории      | UC-4  | US-4 | `GET /guests/{id}/recommendations`          | `reco.list.fallback-no-history`                                 | B   | OK (без теста) |
| Пустой список — валидное состояние  | UC-4  | US-4 | `GET /guests/{id}/recommendations`          | `reco.list.empty-ok`                                            | P   | OK (без теста) |
| Неизвестный гость → 404             | UC-4  | —    | `GET /guests/{id}/recommendations`          | `TestRecommendationsUnknownGuestReturns404`                     | N   | OK             |
| Refresh пересчитывает набор         | UC-15 | US-4 | `POST /guests/{id}/recommendations/refresh` | `TestRecommendationsRefreshReplacesSet`                         | P,B | OK             |
| Refresh: таймаут → 504              | UC-15 | —    | `POST .../refresh`                          | `reco.refresh.timeout`                                          | N   | blocked (G10)  |
| Агрегат по сегменту                 | UC-4  | US-4 | `GET /segments/{id}/recommendations`        | `TestSegmentRecommendationsAggregate`                           | P,C | OK             |
| R-3 → 403                           | UC-4  | US-4 | `/guests/{id}/recommendations`              | `TestRoleTenantCannotReachPrivateAPI`                           | N,S | OK             |

## 7. Выгрузка, лимиты и кампании (UC-5, UC-10, UC-20)

| Требование                             | UC    | US    | Endpoint                      | Тест                                          | Тип | Статус         |
| -------------------------------------- | ----- | ----- | ----------------------------- | --------------------------------------------- | --- | -------------- |
| Preview: 4 счётчика                    | UC-5  | US-5  | `POST /exports/preview`       | `TestExportPreviewCounters`                   | P,C | OK             |
| Фильтр по контакту канала              | UC-5  | US-5  | `POST /exports/preview`       | `TestExportPreviewChannelContactFilter`       | B   | OK             |
| Лимит действует в разрезе канала       | UC-10 | US-10 | `POST /exports/preview`       | `TestContactLimitIsChannelScoped`             | B   | OK             |
| Создание выгрузки → 201                | UC-5  | US-5  | `POST /exports`               | `TestExportCreateReady`                       | P,C | OK             |
| Без `Idempotency-Key` → 400            | UC-5  | US-5  | `POST /exports`               | `TestExportCreateRequiresIdempotencyKey`      | N   | OK             |
| Повтор ключа → та же выгрузка          | UC-5  | US-5  | `POST /exports`               | `TestExportCreateIsIdempotent`                | B   | OK             |
| Сегмент не `ready` → 409               | UC-5  | US-5  | `POST /exports`               | `exports.create.segment-not-ready`            | N,B | OK (без теста) |
| Все исключены лимитом → 409            | UC-10 | US-10 | `POST /exports`               | `exports.create.all-blocked`                  | N,B | OK (без теста) |
| Контрольная группа не в CSV            | UC-5  | US-5  | `GET /exports/{id}/download`  | `TestExportDownloadExcludesControlGroup`      | B   | OK             |
| `is_control` фильтр в members          | UC-5  | US-5  | `GET /exports/{id}/members`   | `exports.members.control-filter`              | P   | OK (без теста) |
| Инвариант счётчиков                    | UC-5  | US-5  | `GET /exports/{id}`           | `TestExportCountersInvariant`                 | B   | OK             |
| Не готова → 409 на download            | UC-5  | —     | `GET /exports/{id}/download`  | `exports.download.not-ready`                  | N   | OK (без теста) |
| Журнал выгрузок                        | UC-5  | US-5  | `GET /exports`                | `TestExportListJournal`                       | P,C | OK             |
| Неизвестная выгрузка → 404             | UC-5  | —     | `GET /exports/{id}`           | `TestExportGetUnknownReturns404`              | N   | OK             |
| Неизвестный канал → 400                | UC-5  | —     | `POST /exports/preview`       | `TestExportPreviewRejectsUnknownChannel`      | N   | OK             |
| Неизвестный сегмент → 404              | UC-5  | —     | `POST /exports/preview`       | `TestExportPreviewRejectsUnknownSegment`      | N   | OK             |
| Правило лимита: `maxContacts<1` → 400  | UC-10 | US-10 | `POST /contact-limits`        | `limits.create.invalid`                       | N,B | OK (без теста) |
| Самое строгое правило                  | UC-10 | US-10 | `POST /exports/preview`       | `TestContactLimitStrictestRuleWins`           | B   | OK             |
| Без правил лимит не применяется        | UC-10 | US-10 | `POST /exports/preview`       | `limits.none-configured`                      | B   | OK (без теста) |
| Контрольная группа не в лимите         | UC-10 | US-10 | `POST /exports`               | `TestContactLimitControlGroupNotLogged`       | B   | OK             |
| R-4 читает, но не меняет лимиты        | UC-10 | US-10 | `/contact-limits*`            | `limits.r4.read-only`                         | N,S | OK (дефект)    |
| Кампания: конверсия и инкремент        | UC-20 | —     | `GET /campaigns/{id}/results` | `campaigns.results.conversion`                | P,C | MISS           |
| Удаление кампании с выгрузками → 409   | UC-20 | —     | `DELETE /campaigns/{id}`      | `TestCampaignDeleteWithExportsReturns409`     | N,B | OK (дефект)    |
| Кампания: окно атрибуции настраиваемое | UC-20 | —     | `POST /campaigns`             | `TestCampaignCreateDefaultsAttributionWindow` | B   | OK (дефект)    |
| Создание кампании                      | UC-20 | —     | `POST /campaigns`             | `TestCampaignCreateRequiresTenantAndSegment`  | P,C | OK (дефект)    |

R-4 не читает правила лимитов: `GET /contact-limits` открыт только
`marketer`/`manager`/`analyst`, tenant-роль получает 403 на всём маршруте
(`02-contradictions.md` §2.9). `GET /campaigns/{id}/results` не смонтирован
(`02-contradictions.md` §2.11).

## 8. Идентичность (UC-11, UC-12)

| Требование                                      | UC    | US    | Endpoint                                 | Тест                                                 | Тип | Статус         |
| ----------------------------------------------- | ----- | ----- | ---------------------------------------- | ---------------------------------------------------- | --- | -------------- |
| Очередь с причиной и score                      | UC-11 | US-11 | `GET /merge-candidates`                  | `TestMergeCandidatesListContract`                    | P,C | OK             |
| Фильтр по `resolution`                          | UC-11 | US-11 | `GET /merge-candidates?resolution=`      | `TestMergeCandidateListFiltersByResolution`          | P   | OK             |
| Детали пары: два профиля                        | UC-11 | US-11 | `GET /merge-candidates/{id}`             | `TestMergeCandidateGetContract`                      | P,C | OK             |
| Неизвестная пара → 404                          | UC-11 | —     | `GET /merge-candidates/{id}`             | `TestMergeCandidateUnknownReturns404`                | N   | OK             |
| Confirm переносит события и идентификаторы      | UC-11 | US-11 | `POST .../confirm`                       | `TestMergeCandidateConfirmMovesEventsAndIdentifiers` | P,B | OK             |
| Confirm: поглощённый профиль `merged`           | UC-11 | US-11 | `GET /guests/{id}`                       | `TestMergeCandidateConfirmMovesEventsAndIdentifiers` | B   | OK             |
| Повторный confirm → 409                         | UC-11 | US-11 | `POST .../confirm`                       | `TestMergeCandidateConfirmTwiceReturns409`           | N,B | OK             |
| Reject создаёт запрет                           | UC-11 | US-11 | `POST .../reject`                        | `TestMergeCandidateRejectCreatesBlock`               | P,B | OK             |
| Revert возвращает данные                        | UC-11 | US-11 | `POST .../revert`                        | `TestMergeCandidateRevertRestoresOrigin`             | P,B | OK             |
| Revert нерешённой пары → 409                    | UC-11 | —     | `POST .../revert`                        | `TestMergeCandidateRevertPendingReturns409`          | N   | OK             |
| Пара под запретом → 409                         | UC-11 | US-11 | `POST /guests/{id}/merge`                | `merge.manual.blocked`                               | N,B | OK (без теста) |
| Ручное объединение с собой → 400                | UC-11 | —     | `POST /guests/{id}/merge`                | `merge.manual.self`                                  | N   | OK (без теста) |
| Список запретов и снятие                        | UC-11 | US-11 | `/merge-blocks*`                         | `merge.blocks.list-delete`                           | P   | OK (без теста) |
| Аудит решения                                   | UC-11 | US-11 | `GET /audit-logs`                        | `merge.confirm.audit`                                | B   | OK (дефект)    |
| R-1 только читает очередь                       | UC-11 | US-11 | `POST .../confirm`                       | `TestRoleMatrix`                                     | N,S | OK             |
| Автообъединение: гость создаётся из импорта     | UC-12 | US-2  | наблюдение через `GET /guests`           | `TestIdentityResolveCreatesGuestFromImport`          | B   | OK             |
| Автообъединение: привязка к существующему гостю | UC-12 | US-2  | наблюдение через `GET /guests`           | `TestIdentityResolveAttachesToExistingGuest`         | B   | OK             |
| Автообъединение по общему телефону              | UC-12 | US-11 | наблюдение через `GET /merge-candidates` | `TestIdentityAutoMergeOnSharedPhone`                 | B   | OK             |
| Спорная пара по общему email                    | UC-12 | US-11 | наблюдение через `GET /merge-candidates` | `TestIdentityPendingCandidateOnSharedEmail`          | B   | OK             |
| Сквозной цикл объединения и разъединения        | UC-11 | US-11 | `/merge-candidates*`                     | `TestE2EIdentityMergeLifecycle`                      | E   | OK             |

Аудит объединения пишется, но `request_id` в записи остаётся `NULL`
(`02-contradictions.md` G4), поэтому требование «каждая запись содержит
`requestId`» не выполняется.

## 9. Аналитика (UC-7, UC-8, UC-9)

| Требование                        | UC   | US   | Endpoint                         | Тест                                      | Тип | Статус         |
| --------------------------------- | ---- | ---- | -------------------------------- | ----------------------------------------- | --- | -------------- |
| Распределение 1/2/3+              | UC-7 | US-7 | `GET /dashboard/cross-tenant`    | `TestCrossTenantDistributionsContract`    | P,C | OK (дефект)    |
| `from > to` → 400                 | UC-7 | US-7 | `GET /dashboard/cross-tenant`    | `TestCrossTenantRejectsInvertedPeriod`    | N,B | OK             |
| Период > 24 мес → 400             | UC-7 | US-7 | `GET /dashboard/cross-tenant`    | `TestCrossTenantRejectsOversizedPeriod`   | N,B | OK             |
| Метрики: match rate и др.         | UC-7 | US-7 | `GET /dashboard/metrics`         | `TestDashboardMetricsFields`              | P,C | OK (дефект)    |
| R-3 → 403                         | UC-7 | US-7 | `/dashboard/*`                   | `dashboard.forbidden-tenant-role`         | N,S | OK (без теста) |
| Агрегаты своего арендатора        | UC-8 | US-8 | `GET /tenants/{id}/aggregates`   | `TestTenantAggregatesTenantRoleIsolation` | P,C | OK             |
| Чужой арендатор → 403             | UC-8 | US-8 | `GET /tenants/{id}/aggregates`   | `TestTenantAggregatesTenantRoleIsolation` | N,S | OK             |
| k = 5: малые ячейки скрыты        | UC-8 | US-8 | `GET /tenants/{id}/aggregates`   | `TestTenantAggregatesRespectsKThreshold`  | B   | OK             |
| Нет `guest_id` в ответе           | UC-8 | US-8 | `GET /tenants/{id}/aggregates`   | `TestTenantAggregatesWithoutGuestIDs`     | S   | OK             |
| R-2 читает любого арендатора      | UC-8 | US-8 | `GET /tenants/{id}/aggregates`   | `aggregates.r2.any`                       | P   | OK (без теста) |
| Качество по источникам            | UC-9 | US-9 | `GET /data-quality/sources`      | `TestDataQualitySourcesList`              | P,C | OK             |
| Источник без загрузок — не ошибка | UC-9 | US-9 | `GET /data-quality/sources/{id}` | `dq.source.no-data`                       | P   | MISS           |
| Детали и динамика по источнику    | UC-9 | US-9 | `GET /data-quality/sources/{id}` | `dq.source.detail`                        | P,C | MISS           |
| R-1/R-3 → 403                     | UC-9 | US-9 | `/data-quality/*`                | `TestDataQualityForbiddenForMarketer`     | N,S | OK             |

`GET /dashboard/cross-tenant` и `GET /dashboard/metrics` отвечают 500, пока таблицы
`event`/`merge_candidate` пусты: `share` и `match_rate` приходят как `NULL`.
Тесты фиксируют это записью в отчёте и пропускаются до исправления.

## 10. Аудит и служебные (UC-19)

| Требование                        | UC    | US  | Endpoint          | Тест                      | Тип | Статус         |
| --------------------------------- | ----- | --- | ----------------- | ------------------------- | --- | -------------- |
| Журнал с фильтрами                | UC-19 | —   | `GET /audit-logs` | `audit.list.filters`      | P,C | OK (без теста) |
| `requestId` в записи              | UC-19 | —   | `GET /audit-logs` | `audit.record.request-id` | B   | OK (дефект)    |
| R-1/R-3 → 403                     | UC-19 | —   | `GET /audit-logs` | `audit.forbidden-roles`   | N,S | OK (без теста) |
| `/healthz` без токена             | —     | —   | `GET /healthz`    | `TestHealthzContract`     | P,C | OK             |
| `/healthz` проверяет БД и очередь | —     | —   | `GET /healthz`    | `health.checks`           | C   | OK (дефект)    |

## 11. Сводка покрытия

| Раздел                        | Требований | OK      | OK (без теста) | OK (дефект) | blocked | MISS  |
| ----------------------------- | ---------- | ------- | -------------- | ----------- | ------- | ----- |
| Аутентификация и пользователи | 15         | 11      | 2              | 1           | 1       | 0     |
| Справочники и источники       | 10         | 7       | 1              | 2           | 0       | 0     |
| Импорт                        | 20         | 16      | 4              | 0           | 0       | 0     |
| Профиль гостя                 | 11         | 10      | 0              | 1           | 0       | 0     |
| Сегменты                      | 19         | 16      | 2              | 0           | 1       | 0     |
| Рекомендации                  | 10         | 6       | 2              | 0           | 2       | 0     |
| Выгрузка, лимиты и кампании   | 25         | 14      | 6              | 4           | 0       | 1     |
| Идентичность                  | 20         | 16      | 3              | 1           | 0       | 0     |
| Аналитика                     | 14         | 8       | 2              | 2           | 0       | 2     |
| Аудит и служебные             | 5          | 1       | 2              | 2           | 0       | 0     |
| **Итого**                     | **149**    | **105** | **24**         | **13**      | **4**   | **3** |

Из 149 требований 105 покрыто реализованными тестами (`OK`), ещё 24 выполнимы,
но тест не написан. Четыре проверки заблокированы внешними причинами:
отсутствие rate limiter (G9), управляемой задержки (G10) и неразрешённое
противоречие D1. Три требования не проверяемы: `GET /campaigns/{id}/results` и
`GET /data-quality/sources/{id}` не смонтированы (`02-contradictions.md` §2.11).
Тринадцать требований помечены `OK (дефект)`: поведение расходится с документом,
расхождение зафиксировано отчётом `UCRP_E2E_DRIFT_REPORT`.

## 12. Требования без endpoint (проверить нельзя)

| Требование                              | Причина                                         |
| --------------------------------------- | ----------------------------------------------- |
| Конверсия и инкремент по кампании       | `GET /campaigns/{id}/results` не смонтирован    |
| Источник без загрузок — не ошибка       | `GET /data-quality/sources/{id}` не смонтирован |
| Детали и динамика по источнику          | то же                                           |
| Пересчёт сегментов по событию (UC-13)   | только очередь `segment.recompute`              |
| Расчёт рекомендаций по событию (UC-15)  | только очередь `recommend.refresh`              |
| Схлопывание заданий по `dedup_key`      | нет REST к `job_queue`                          |
| Инкремент против контрольной группы     | требует накопления событий и окна атрибуции     |
| Match rate, доля оспоренных объединений | агрегаты по объёму данных                       |
| Время «сегмент → выгрузка» ≤ 5 мин      | метрика по `segment_version.created_at`         |
| Rate limit логина                       | примитив не реализован, порог не задан          |
| Таймауты 504                            | нет управляемой задержки                        |
| Физическое удаление запрещено           | проверяется только косвенно                     |

Автообъединение профилей (UC-12) больше не входит в этот список: воркер
`identity.resolve` работает, и её поведение наблюдается через
`GET /merge-candidates` и `GET /guests` (`TestIdentity*`).
