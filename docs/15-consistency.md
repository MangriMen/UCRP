# 15. Проверка непротиворечивости

Проверка выполнена по файлам `docs/` и `appendix/` сверкой и поиском (сопоставление всех `METHOD /path` с реестром, всех ссылок `Q-n`/`D-n`/`A-n` со списками, таблиц DDL с ER, CRUD, деревом репозиториев; ручная проверка Mermaid). Найденные расхождения исправлены в исходных файлах (раздел 11). Статус: **OK** — расхождений нет; **Исправлено** — найдено и исправлено.

## 1. Сводка чек-листа (шаг 15)

| № | Проверка | Результат |
|---|---|---|
| 1 | Названия сущностей, ролей, полей, ID совпадают везде | Исправлено (раздел 11, п. 2, 5, 6, 7), остальное OK |
| 2 | Каждый UC покрыт Activity, Sequence, эндпоинтом и CRUD-строкой | OK после добавления SD-9; причины для UC без собственной диаграммы — в `14-traceability.md` |
| 3 | Каждая таблица ER используется репозиторием и есть в CRUD | OK (27 из 27) |
| 4 | Роли эндпоинтов = CRUD-матрица; изоляция R-3 везде | Исправлено (4 ячейки CRUD), изоляция OK |
| 5 | Рекомендации в примерах без посещённых мест | OK |
| 6 | В выгрузках только гости с контактом нужного канала | OK |
| 7 | DFD ↔ Sequence ↔ C4 | Исправлено (поток `segment.recompute` в DFD-1B), остальное OK |
| 8 | Метрики `analysis.md` §8 вычислимы из ER | OK; определение «доли оспоренных склеек» уточнено |

## 2. Проверка 1: названия и идентификаторы

| Что сверялось | Метод | Итог |
|---|---|---|
| Эндпоинты в документах и реестре | Извлечены все `METHOD /path` из 18 файлов (кроме реестра) и сопоставлены со 90 строками реестра | Нет ни одного эндпоинта, отсутствующего в реестре; 26 совпадений «по форме» — сокращения в диаграммах (`/imports/id`, `PATCH /contact-limits` без `{id}` в групповых записях) |
| Ссылки `Q-n`, `D-n`, `A-n` | Все ссылки из файлов сопоставлены со списками `00-open-questions.md` и `01-summary-assumptions.md` | Все ссылки разрешаются: Q-1…Q-27, D-1…D-33, A-1…A-16 |
| ID `R-1…R-6`, `E-1…E-21`, `UC-1…UC-20`, `US-1…US-11` | Поиск несуществующих ID (R-7, E-22+, UC-21+, US-12+) | Не найдены |
| Имена ролей в коде | `marketer`, `manager`, `tenant`, `analyst`, `system` — глоссарий, DDL (`CHECK`), объектные диаграммы, SD | Совпадают |
| Имена сущностей и таблиц | `02-glossary.md` ↔ `05-er-data-model.md` ↔ `ddl.sql` (E-16 `Export`/`export_job`, E-20 `User`/`app_user`) | Совпадают |
| Хэндлеры, сервисы, репозитории | Имена из US, SD, AD, CL ↔ `go-package-tree.md` | Расхождения: `RecommendationHandler`, `ContactLimitHandler` отсутствовали в дереве; `SyncHandler` в SD-3 не существует (исправлено) |
| Методы `SourceConnector` | CL-2 (`Open`, `Pull`) ↔ AD-1b (`Read`) | Расхождение исправлено |
| Количество | 27 таблиц DDL = 27 в ER = 27 в D1…D7; 90 эндпоинтов в реестре, 00, 12; 7 типов заданий | Совпадают |

## 3. Проверка 2: покрытие UC

Полная трасса — `14-traceability.md`. Итог по UC-1…UC-20: эндпоинт или причина его отсутствия есть у каждого; CRUD-строка есть у каждого; Sequence есть у каждого (после добавления SD-9 для UC-2; UC-7, UC-9, UC-17…UC-20 опираются на типовую цепочку SD-7); Activity есть у UC-1, 3, 4, 5, 6, 10, 11, 12, 13, 14, 15, остальные — с указанной причиной (чтение или типовой CRUD без разветвлённой логики; решение D-32).

## 4. Проверка 3: таблицы ER ↔ репозитории ↔ CRUD

| Таблица | Сущность (CRUD-строка) | Репозиторий (`go-package-tree.md`) | Основные UC |
|---|---|---|---|
| `tenant_category` | E-6 | `tenant_category_repo` | UC-17, UC-15 |
| `tenant_category_relation` | E-6 | `tenant_category_relation_repo` | UC-17, UC-15 |
| `tenant` | E-4 | `tenant_repo` | UC-17, UC-8, UC-15 |
| `project` | E-7 | `project_repo` | UC-17, UC-2 |
| `event_type` | E-9 | `event_type_repo` | UC-17, UC-1 |
| `data_source_type` | E-5 | `data_source_type_repo` | UC-17, UC-1 |
| `data_source` | E-5 | `data_source_repo` | UC-1, UC-14, UC-17 |
| `app_user` | E-20 | `user_repo` | UC-16, UC-18 |
| `audit_log` | E-21 | `audit_repo` | UC-19, UC-11, UC-5 |
| `guest` | E-1 | `guest_repo` | UC-2, UC-12 |
| `guest_identifier` | E-10 | `guest_identifier_repo` | UC-12, UC-2, UC-5 |
| `import_batch` | E-13 | `import_batch_repo` | UC-1, UC-9, UC-14 |
| `rejected_row` | E-14 | `rejected_row_repo` | UC-1, UC-9 |
| `event` | E-2 | `event_repo`, `analytics_repo`, `tenant_aggregate_repo` | UC-1, UC-2, UC-3, UC-7, UC-8 |
| `merge_candidate` | E-11 | `merge_candidate_repo` | UC-11, UC-12, UC-9 |
| `unmerge_block` | E-12 | `unmerge_block_repo` | UC-11, UC-12 |
| `segment` | E-3 | `segment_repo` | UC-3 |
| `segment_version` | E-3 | `segment_version_repo` | UC-3, UC-13, UC-5 |
| `segment_member` | E-3 | `segment_member_repo` | UC-13, UC-5 |
| `campaign` | E-8 | `campaign_repo` | UC-20 |
| `recommendation` | E-15 | `recommendation_repo` | UC-4, UC-15 |
| `recommendation_factor` | E-15 | `recommendation_repo` | UC-4, UC-15 |
| `export_job` | E-16 | `export_repo` | UC-5, UC-20 |
| `export_member` | E-17 | `export_member_repo` | UC-5, UC-20 |
| `contact_limit` | E-18 | `contact_limit_repo` | UC-10 |
| `contact_log` | E-19 | `contact_log_repo` | UC-5, UC-10, UC-2 |
| `job_queue` | T-1 (служебная) | `job_queue_repo` | UC-6, UC-1, UC-13 |

Репозитории тоже закрыты: в дереве 28 файлов `*_repo`: 26 по таблицам (`recommendation_factor` обслуживается `recommendation_repo`) плюс `analytics_repo` и `tenant_aggregate_repo`; неиспользуемых репозиториев нет — `analytics_repo` обслуживает UC-7, UC-9, `tenant_aggregate_repo` — UC-8. Обратная связь с DDL: каждый `CREATE TABLE` — в разделе 1 `05-er-data-model.md`, в таблице 3НФ и в одном из хранилищ D1…D7 DFD.

## 5. Проверка 4: роли эндпоинтов = CRUD, изоляция R-3

Метод: каждая строка раздела 2 `07-crud-matrix.md` сверена со столбцом «Роли» реестра (роли и операции).

| Результат | Детали |
|---|---|
| Роли и операции | Совпадают, кроме агрегатного чтения: `GET /dashboard/metrics`, `GET /data-quality/sources*`, `GET /campaigns/{id}/results` читают E-11, E-14, E-17, E-19 для R-2 (и E-17 для R-4), а в матрице стояли «—». Исправлено: ячейки заменены на **A**, добавлено пояснение и перечень эндпоинтов |
| R-2 | Записывает только пользователей (`/users`): подтверждено по реестру |
| R-1 | Нет ни одной операции импорта, склеек, пользователей и справочников (кроме `priority` арендатора) |
| Маршруты React | Роли `RequireRole` ⊆ ролей реестра для главной операции экрана (таблица маршрутов 07 §3) — OK |

Изоляция R-3 присутствует в каждом классе артефактов:

| Артефакт | Где |
|---|---|
| Реестр | Правило изоляции; R-3 только у `/auth/*`, `GET /tenants`, `GET /tenants/{id}`, `GET /tenants/{id}/aggregates` |
| CRUD | Строка R-3 и механизм (07 §4), ячейки `own`, `A own` |
| Activity | Преамбула 08: R-3 не участвует ни в одном потоке, кроме UC-8 |
| Sequence | SD-7 (401 → 403 → 400 → 403 по `tenant_id` → 404/409), `...ForTenant` |
| Class | `TenantAggregates` DTO без `guest_id`, `Principal`; репозиторий `tenant_aggregate_repo` только `...ForTenant` |
| Component | CMP-D2: `RequireRole`, проверка `tenant_id` в service, `ForTenant` |
| C4 | L1, L2, L3a, L3b |
| DFD | DFD-0 (один входящий и один исходящий поток R-3), DFD-1B (порог k = 5) |
| ER | `app_user`: CHECK «роль `tenant` ⇒ `tenant_id` задан» |
| OpenAPI | `GET /tenants/{id}/aggregates` без `guest_id`, 403 `forbidden_tenant` |

## 6. Проверка 5: рекомендации без посещённых мест

| Пример | Файл | Посещённые места гостя | Рекомендация | Итог |
|---|---|---|---|---|
| g-1001 Анна | `06-object-diagrams.md` (c2) | Кофейня Bean, ФитЗона | Кинотеатр Синема | OK |
| Пример обоснования | `04-user-stories.md` US-4 | ФитЗона (2 этаж) | Кинотеатр (3 этаж) | OK |
| Ответ API | `appendix/openapi-fragments.md` §4 | Кофейня Bean, ФитЗона | Кинотеатр Синема | OK |
| Инвариант на уровне БД и алгоритма | AD-4 (кандидаты = активные арендаторы без посещённых), `UNIQUE(guest_id, tenant_id)` | — | — | OK; запрет конкурентов — `tenant_category_relation (competitor)` |

## 7. Проверка 6: выгрузки только с контактом канала

| Место | Подтверждение |
|---|---|
| US-5 AC2 | SMS и пуш — только телефон, email — только email |
| AD-5 S3, SD-5 | Фильтр по контакту выполняется до лимита и контрольной группы |
| ER | `export_member.identifier_id NOT NULL` — каждому члену выгрузки сопоставлен конкретный идентификатор |
| Class | `ExportChannel.ContactType()` (CL-5) |
| Пример (06, c1) | Сергей без телефона исключён (`no_contact_count = 1`); в `export_member` только гости с телефоном; сумма счётчиков 2 + 1 + 1 + 1 = 5 |
| OpenAPI | `GET /exports/{id}/members` возвращает только `contact` канала |

Ограничение: соответствие `identifier_id` каналу контролируется сервисом (CHECK в БД невозможен без дублирования канала); учтено как инвариант service-слоя и тест ExportService.

## 8. Проверка 7: DFD ↔ Sequence ↔ C4

| DFD-поток | Sequence | C4 / Component | Итог |
|---|---|---|---|
| TS → 1.1 → D4, D7 | SD-2, SD-3 | L2: Worker → HTTP pull, API → БД, BlobStore | OK |
| 1.2 → D1, D4; 1.2 → D7 (`identity.resolve`) | SD-2 (воркер) | L3a `ingestion` → `queue` | OK |
| 1.3, 1.4 ↔ D1, D3 | SD-6 | L3a, L4 identity | OK |
| D7 → P2 (`segment.recompute`) | SD-4, SD-6 | CMP-D1 `Q -.-> SEG` | Исправлено: в DFD-1B не было потока пересчёта сегментов, добавлен |
| D7 → P3 (`recommend.refresh`) | SD-8, SD-9 | CMP-D1 `Q -.-> REC` | OK |
| R-1 → P2 → D5 | SD-4 | L3a `segments` | OK |
| P4 ↔ D5, D6; P4 → PRV | SD-5 | L2: PRV → API (members); L1: P → PR | OK |
| R-3 ↔ P5 | SD-7 | L1, L2, L3a, CMP-D2 | OK |
| D7 (аудит, очередь) | SD-1…SD-9 | L2: БД с `job_queue` | OK; название D7 выровнено с таблицей хранилищ |
| Хранилища D1…D7 | — | 27 таблиц | OK (3 + 7 + 2 + 2 + 6 + 4 + 3) |

## 9. Проверка 8: метрики `analysis.md` §8 вычислимы из ER

Все показатели считаются запросами по таблицам DDL и отдаются `GET /dashboard/metrics` (`AnalyticsRepo`); отдельных таблиц метрик не нужно.

| Метрика (цель) | Формула | Таблицы / ключевые поля | Эндпоинт |
|---|---|---|---|
| Cross-tenant penetration (рост за период) | Для активных гостей (`guest.status='active'`) число различных `event.tenant_id` среди событий типов с `event_type.counts_as_presence`; распределение 1 / 2 / 3+; доля гостей с ≥ 2 | `guest`, `event`, `event_type`; индекс `ix_event_guest_tenant` | `GET /dashboard/penetration` |
| Match rate (≥ 70%) | События с `guest_id IS NOT NULL` ÷ все принятые события за период (по `event.created_at`); срез по `data_source_id` | `event` (индекс `ix_event_unmatched`) | `GET /dashboard/metrics`, `GET /data-quality/sources` |
| Доля оспоренных склеек (< 1%) | `merge_candidate` с `resolution='reverted'` ÷ (`auto_merged` + `confirmed` + `reverted`) за период по `decided_at`/`created_at` | `merge_candidate` (индекс `ix_merge_resolution`), `audit_log` для расследования | `GET /dashboard/metrics` |
| Время сегмент → выгрузка (≤ 5 мин) | Для первой выгрузки версии: `export_job.finished_at` − `segment_version.created_at`; медиана и 95-й перцентиль | `segment_version`, `export_job` | `GET /dashboard/metrics` |
| Конверсия рассылки (≥ 3%) | Получатели (`contact_log`) с хотя бы одним событием `counts_as_presence` у целевого арендатора кампании в интервале (`contacted_at`, `contacted_at` + `campaign.attribution_days`] ÷ число получателей | `contact_log`, `export_job.campaign_id`, `campaign` (`target_tenant_id`, `attribution_days`), `event`, `event_type` | `GET /campaigns/{id}/results`, `GET /dashboard/metrics` |
| Инкремент визитов (> 0, значимо) | Конверсия получателей − конверсия контрольной группы (`export_member.is_control = TRUE`, окно от `export_job.created_at`); значимость — двухвыборочный тест долей в `DashboardService` по четырём счётчикам | `export_member`, `export_job`, `contact_log`, `campaign`, `event` | `GET /campaigns/{id}/results` |

Допущения расчёта вынесены в D-30, D-31, Q-26, Q-27. Контрольная группа не попадает в `contact_log` (D-17), поэтому её база строится от `export_job.created_at` — это сознательная асимметрия, сохраняющая лимит контактов.

## 10. Ручная проверка Mermaid

43 диаграммы: 3 use case, 4 ER, 5 object, 6 activity, 9 sequence, 6 class, 2 component, 5 C4, 3 DFD. Проверено: корректные ключевые слова и направления; парность `alt/else/end`, `loop/end`, `subgraph/end` (скриптом по всем блокам: несоответствий нет); чётное число кавычек; во всех узлах flowchart текст с пробелами и знаками заключён в `"…"`; в `sequenceDiagram` нет `;` в тексте сообщений, в `classDiagram` нет скобок и `;` в значениях атрибутов экземпляров (исправлено одно `;` в 06 c2); ER-атрибуты используют допустимые типы без скобок; идентификаторы узлов не совпадают с ключевыми словами (`end`, `graph`); число узлов в диаграмме ≤ 15 (максимум 13 в flowchart, 10 сущностей в ER). Известные ограничения нотации описаны в каждом файле (нет нативных swimlanes, use case, object diagram). Визуальный рендер на mermaid.live рекомендуется выполнить перед публикацией.

## 11. Перечень исправлений

| № | Файл | Что исправлено |
|---|---|---|
| 1 | `07-crud-matrix.md` | Ячейки R-2 для E-11, E-14, E-17, E-19 и R-4 для E-17 заменены с «—» на «A» (агрегаты в `/dashboard/metrics`, `/data-quality/*`, `/campaigns/{id}/results`); добавлено пояснение и перечень эндпоинтов в раздел 2 |
| 2 | `04-user-stories.md` | US-5: `ExportChannel` перечислял форматы `csv`, `api`; исправлено на каналы `sms`, `push`, `email` и отдельно `ExportFormat` (`csv`, `api`) |
| 3 | `02-glossary.md` | «Доля оспоренных склеек»: определение переписано так, чтобы вычисляться из `merge_candidate` без аудита (D-31) |
| 4 | `06-object-diagrams.md` | Убран `;` в значении атрибута (c2) — синтаксическая ошибка `classDiagram` |
| 5 | `appendix/go-package-tree.md` | Добавлены `recommendation_handler`, `contact_limit_handler`; `CampaignService` отнесён к CMP-6 (как в `11-component-diagram.md`), а не к CMP-4; зафиксировано исключение: пакет `identity/normalize` с чистыми функциями доступен `ingestion` (раньше схема запрещала такой импорт, но CL-2 и AD-1b его использовали) |
| 6 | `09-sequence-diagrams.md` | SD-3: участник `SyncHandler` заменён на существующий `ImportHandler`; добавлена SD-9 (профиль, рекомендации, видимое обновление; UC-2, UC-4, UC-6) |
| 7 | `10-class-diagram.md`, `08-activity-diagrams.md` | CL-2: `Normalize` уточнён как `internal/identity/normalize`; AD-1b: метод `SourceConnector.Read` заменён на `Open`/`Pull` (по CL-2) |
| 8 | `13-dfd.md` | DFD-1B: добавлен поток `D7 → P2` (`segment.recompute`); таблица хранилищ: D7 переименовано и дополнены читающие процессы |
| 9 | `00-open-questions.md` | Добавлены Q-26, Q-27, D-30…D-33; устаревшие чекпойнты заменены итоговым |

## 12. Критерии приёмки (раздел 6 плана)

| Критерий | Проверка | Результат |
|---|---|---|
| Непротиворечивость | Трассируемость без пробелов (`14`), единый глоссарий (`02`), раздел 2 выше | OK |
| Масштабируемость | Stateless API (JWT), очередь `job_queue` и несколько воркеров, кэш размеров (порт `Cache`), интерфейсы-порты, таблица точек масштабирования (11) | OK |
| Изменяемость | 05 §4 и матрица влияния (11): новый источник, тип события, сегмент, канал — без изменения существующего кода и схемы | OK |
| 3НФ | Обоснование по каждой из 27 таблиц (05 §3), отступления перечислены с источником истины | OK |
| Соответствие Go/React | Под каждой диаграммой таблица «элемент → пакет → модуль → контракт»; деревья пакетов и модулей, соответствие CMP-1…CMP-10 | OK |
| Полнота CRUD | Матрица R-1…R-6 × E-1…E-21 + T-1 (07) | OK |
| Покрытие кейса | Раздел 13 ниже; US-1…US-11 — `04`, `14` | OK |
| Читаемость | 43 диаграммы с пояснениями (3–7 предложений), ≤ 15 узлов | OK |
| Без кода | В `docs/` 43 блока `mermaid`, 2 блока `text` (деревья), SQL DDL, JSON-схемы и примеры тел в OpenAPI-фрагментах; блоков с Go, JS и конфигами нет; вне `docs/` только `plan.md`, `case.txt`, `analysis.*` | OK |

## 13. Покрытие 5 ожидаемых результатов `case.txt`

| Ожидаемый результат | US / UC | Где проектируется | Эндпоинты |
|---|---|---|---|
| Сбор данных из ≥ 3 форматов (CRM/PMS, касса/POS, API, Excel/CSV) | US-1, UC-1, UC-14 | CL-2 (`SourceConnector`: CSV/Excel, CRM, POS, API), AD-1a/1b, SD-2, SD-3, 05 §4 | `POST /imports`, `POST /data-sources/{id}/sync`, `/data-source-types` |
| Единый профиль и лента событий (визит, покупка/бронь, проект, дата, сумма) | US-2, UC-2, UC-12 | `event`, `event_type`, `project`, SD-9, объект (a), AD-2 | `GET /guests/{id}`, `GET /guests/{id}/events` |
| Обновление: новое событие влияет на рекомендации | US-6, UC-6, UC-13, UC-15 | AD-1b → AD-2 → AD-3/AD-4, SD-9, очереди | `GET /segments/{id}`, `POST /segments/{id}/recompute` |
| Сегменты «был в A и B, не был в C» и рекомендация | US-3, US-4, UC-3, UC-4 | AD-3, AD-4, SD-4, объект (c), скоринг по `analysis.md` §7 | `POST /segments/preview`, `POST /segments`, `GET /guests/{id}/recommendations` |
| Выгрузка в пуш/SMS/email | US-5, US-10, UC-5, UC-10 | AD-5, SD-5, CL-5, DFD-1B | `POST /exports`, `GET /exports/{id}/download`, `GET /exports/{id}/members` |

Юридика, согласия, полноценная лояльность и собственная отправка сообщений вне scope (A-1, A-8; `01-summary-assumptions.md` §3).

## Чекпойнт
Проверены 8 пунктов шага 15, критерии раздела 6 и 5 результатов `case.txt`. Исправлено 9 групп расхождений (раздел 11). Добавлено: Q-26, Q-27, D-30…D-33. Итоговые списки — `00-open-questions.md`.
