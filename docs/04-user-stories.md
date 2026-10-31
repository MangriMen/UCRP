# 04. User Stories US-1…US-11

Формулировки и приёмка взяты из `analysis.md` §6.1; критерии переписаны в формате Given/When/Then с позитивными и негативными сценариями. MVP-ядро: US-1…US-6, US-10, US-11; US-7, US-8, US-9 — вторая очередь (в модель и матрицы включены сразу). Эндпоинты — из `appendix/endpoint-registry.md` (префикс `/api/v1`), Go-слои: `handler` (`internal/handler`) → `service` (`internal/service`) → `repository` (`internal/repository`).

React-маршруты и модули (`src/`): `features/import`, `features/profiles` (включая очередь склеек), `features/segments`, `features/recommendations`, `features/export` (включая лимиты и кампании), `features/dashboard`, `features/tenant`, `features/dq`, `features/admin` (D-14), `auth/`, `routes/`, `components/`, `api/`.

---

## US-1. Загрузка данных арендатора (UC-1) — MVP

Как аналитик УК (R-4), я хочу загрузить данные арендатора в любом из четырёх форматов (CRM, касса, API, Excel/CSV), чтобы собрать историю гостей без участия разработчика.

**Acceptance Criteria**
1. Given выбран активный источник и файл CSV/Excel до 100 тыс. строк, When R-4 нажимает «Загрузить», Then API отвечает 202 с `batch_id`, экран показывает статус `queued → running → completed` с обновлением каждые 2 с.
2. Given загрузка завершена, When R-4 открывает её, Then видны счётчики `total`, `accepted`, `rejected`, `duplicate`, причём `accepted + rejected + duplicate = total`.
3. Given в файле строка с неверной датой или без `external_id`, When идёт валидация, Then строка попадает в список отклонённых с номером строки, кодом и причиной на русском; остальные строки принимаются.
4. Given тот же файл (sha256) уже загружен для этого источника, When R-4 загружает его снова, Then API возвращает 200 с существующим `batch_id` и `duplicate=true`; новых событий нет.
5. Given тот же `external_id` встречается в другом файле того же источника, When файл обрабатывается, Then строка считается `duplicate`, событие не создаётся повторно.
6. Given источник типа API с расписанием, When наступает время или R-4 нажимает «Синхронизировать», Then воркер подтягивает события без файла и создаёт загрузку `mode=pull`.
7. Негатив: Given файл больше 100 тыс. строк, When загрузка, Then 413 с понятной причиной; Given неподдерживаемый формат или нет обязательных колонок (по маппингу источника), Then 422 без создания событий; Given пользователь R-1/R-3, Then 403.

**Экран React:** `/import` (выбор источника, drag-n-drop, список загрузок), `/import/:batchId` (статус, счётчики, таблица отклонённых строк с пагинацией и экспортом CSV) — `features/import`, TanStack Table + Query (polling).
**Эндпоинты Go:** `POST /imports`, `GET /imports`, `GET /imports/{id}`, `GET /imports/{id}/rejected-rows`, `POST /data-sources/{id}/sync`; `handler.ImportHandler` → `service.ImportService` → `repository.ImportBatchRepo`, `RejectedRowRepo`, `EventRepo`; воркер `internal/ingestion`.

---

## US-2. Единый профиль гостя (UC-2) — MVP

Как маркетолог УК (R-1), я хочу видеть единый профиль гостя с историей визитов и покупок, чтобы понимать, где человек уже был.

**Acceptance Criteria**
1. Given у гостя события из кассы и CRM, When R-1 открывает профиль, Then виден список событий (тип, арендатор, проект, дата, сумма), отсортированный по дате убывания.
2. Given записи из разных источников с одним телефоном или email, When профиль открыт, Then они объединены в одного гостя, а не показаны как разные.
3. Given гость, When профиль открыт, Then показано число посещённых арендаторов (`tenant_count`) и список их названий.
4. Given поиск по телефону в любом формате (`+7 900…`, `8900…`), When R-1 вводит запрос, Then найден тот же гость (поиск по нормализованному значению).
5. Негатив: Given гость был поглощён склейкой, When открывают его старый `id`, Then ответ содержит `merged_into_id` и интерфейс перенаправляет на актуальный профиль; Given несуществующий `id`, Then 404; Given роль R-3, Then 403.

**Экран React:** `/guests` (поиск и таблица), `/guests/:guestId` (шапка, идентификаторы, лента событий, вкладка рекомендаций, история контактов) — `features/profiles`.
**Эндпоинты Go:** `GET /guests`, `GET /guests/{id}`, `GET /guests/{id}/events`, `GET /guests/{id}/contacts`; `handler.GuestHandler` → `service.ProfileService` → `repository.GuestRepo`, `EventRepo`, `GuestIdentifierRepo`, `ContactLogRepo`.

---

## US-3. Сегмент без кода (UC-3) — MVP

Как маркетолог УК (R-1), я хочу собрать сегмент по правилу «был в A и B, не был в C», чтобы пригласить гостя в новое для него место.

**Acceptance Criteria**
1. Given конструктор сегмента, When R-1 выбирает арендаторов, условие «был / не был», период и (опционально) наличие контакта канала, Then до сохранения показывается размер сегмента и 20 примеров (`POST /segments/preview`).
2. Given выбран шаблон («Спящий сосед», «Один магазин», «Риск ухода», «Новичок», «Высокая ценность»), When R-1 подтверждает, Then форма заполняется параметрами шаблона и правило читается как фраза на бизнес-языке.
3. Given сегмент сохранён, When R-1 открывает его, Then доступен постраничный состав и дата последнего пересчёта.
4. Given правило изменено и сохранено, When сохранение успешно, Then создаётся новая версия правила, прежние версии доступны для просмотра.
5. Негатив: Given арендатор указан одновременно в «был» и «не был», When предпросмотр, Then 400 с указанием поля; Given предпросмотр дольше 10 с, Then 504 и сообщение «сузьте период»; Given роль R-3, Then 403.

**Экран React:** `/segments` (список с размерами), `/segments/new` и `/segments/:id` (конструктор на React Hook Form: блоки «был в», «не был в», период, контакт; панель размера; вкладки «Состав», «Версии») — `features/segments`.
**Эндпоинты Go:** `GET /segment-templates`, `POST /segments/preview`, `POST /segments`, `GET /segments/{id}`, `PUT /segments/{id}`, `GET /segments/{id}/members`, `GET /segments/{id}/versions`; `handler.SegmentHandler` → `service.SegmentService` → `internal/segments` (`SegmentEvaluator`) → `repository.SegmentRepo`, `SegmentVersionRepo`, `SegmentMemberRepo`.

---

## US-4. Рекомендация гостю (UC-4) — MVP

Как маркетолог УК (R-1), я хочу получить рекомендацию, куда пригласить конкретного гостя, чтобы не придумывать оффер вручную.

**Acceptance Criteria**
1. Given у гостя есть события, When R-1 открывает вкладку «Рекомендации», Then показано не более трёх арендаторов с рангом, оценкой и текстовым обоснованием по факторам (например, «Похожие гости ходят в Кинотеатр; Спорт и Развлечения — смежные категории; 2-й этаж → 3-й»).
2. Given гость уже был в арендаторе X, When считаются рекомендации, Then X в списке отсутствует.
3. Given гость бывает у арендатора из категории Y, When кандидат — прямой конкурент по `tenant_category_relation (competitor)`, Then кандидат остаётся в списке, но получает понижающий коэффициент (A-18); исключение включается настройкой категории, а не по умолчанию.
4. Given у гостя нет событий, When открыта вкладка «Рекомендации», Then показана fallback-выдача по популярности и приоритету арендатора с пометкой «мало данных» (A-21).
5. Given у гостя новое событие, When завершён пересчёт, Then рекомендации обновлены (`computed_at` меняется).
6. Негатив: Given кандидатов не осталось (все посещены), When открыт профиль, Then пустое состояние с пояснением, не ошибка; Given роль R-3, Then 403.

**Экран React:** вкладка «Рекомендации» в `/guests/:guestId`, блок «Куда приглашать» в `/segments/:id` — `features/recommendations`.
**Эндпоинты Go:** `GET /guests/{id}/recommendations`, `POST /guests/{id}/recommendations/refresh`, `GET /segments/{id}/recommendations`; `handler.RecommendationHandler` → `service.RecommendationService` → `internal/recommend` (`Recommender`) → `repository.RecommendationRepo`, `EventRepo`, `TenantRepo`, `TenantCategoryRelationRepo`.

---

## US-5. Выгрузка сегмента в канал (UC-5) — MVP

Как маркетолог УК (R-1), я хочу выгрузить сегмент в канал рассылки, чтобы кампания запускалась из одного окна.

**Acceptance Criteria**
1. Given посчитанный сегмент, When R-1 выбирает канал (SMS/пуш/email) и формат (CSV/API), Then предпросмотр показывает: всего в сегменте, с контактом канала, исключено лимитом, в контрольной группе, к отправке.
2. Given канал SMS или пуш, When формируется выгрузка, Then в неё попадают только гости с телефоном; для email — только с email; прочие не попадают.
3. Given задан процент контрольной группы больше нуля, When выгрузка создаётся, Then это количество гостей помечается `is_control=true`, воспроизводимо по seed, и не попадает в CSV и в контакты. Given процент равен 0 (значение по умолчанию), When выгрузка создаётся, Then контрольная группа не выделяется, а выгрузка остаётся валидной (A-19).
4. Given выгрузка создана, When она готова, Then запись есть в журнале (кто, когда, сегмент и его версия, канал, счётчики) и скачивается CSV или читается по API.
5. Given повторный запрос с тем же `Idempotency-Key`, When сервер получает его, Then возвращается та же выгрузка, новый ContactLog не создаётся.
6. Негатив: Given сегмент ещё считается, When создаётся выгрузка, Then 409; Given все гости заблокированы лимитом, Then 409 с причиной (US-10); Given нет `Idempotency-Key`, Then 400; Given роль R-2/R-3/R-4, Then 403.

**Экран React:** `/exports/new?segmentId=` (мастер: канал → контрольная группа → предпросмотр → подтверждение), `/exports` (журнал), `/exports/:id` (счётчики, скачивание, ссылка на API) — `features/export`.
**Эндпоинты Go:** `POST /exports/preview`, `POST /exports`, `GET /exports`, `GET /exports/{id}`, `GET /exports/{id}/members`, `GET /exports/{id}/download`; `handler.ExportHandler` → `service.ExportService` → `internal/export` (`ExportChannel`: `sms`, `push`, `email`; `ExportFormat`: `csv`, `api`) → `repository.ExportRepo`, `ExportMemberRepo`, `ContactLogRepo`, `ContactLimitRepo`.

---

## US-6. Обновление при новом событии (UC-6) — MVP

Как маркетолог УК (R-1), я хочу, чтобы новое событие сразу попадало в профиль и меняло сегмент, чтобы рекомендация оставалась актуальной.

**Acceptance Criteria**
1. Given принято новое событие гостя, When воркер применил его, Then событие видно в ленте профиля, `tenant_count` обновлён.
2. Given гость начал или перестал удовлетворять правилу сегмента, When пересчёт завершён, Then состав и размер сегмента изменились, `computed_at` обновлён.
3. Given открыт экран профиля или сегмента, When пересчёт завершился, Then интерфейс подтягивает изменение при опросе (интервал 5 с) без ручного запуска и показывает метку «обновлено».
4. Given пересчёт сегмента запущен, When приходят ещё события, Then задания схлопываются по `dedup_key` и не дублируются.
5. Негатив: Given воркер упал при пересчёте, When задание перезапущено, Then результат идемпотентен (повторный пересчёт не создаёт дублей членов); Given пересчёт не завершён, Then сегмент имеет статус `computing`, а выгрузка из него блокируется (409).

**Экран React:** индикатор «обновлено/пересчитывается» в `/guests/:guestId` и `/segments/:id` — `features/profiles`, `features/segments`; без отдельного экрана.
**Эндпоинты Go:** чтение `GET /guests/{id}/events`, `GET /segments/{id}` (поля `status`, `computed_at`, `member_count`); принудительно `POST /segments/{id}/recompute`; фоновые задания `identity.resolve`, `segment.recompute`, `recommend.refresh` (`internal/ingestion` → очередь → `internal/segments`, `internal/recommend`).

---

## US-7. Сводка управляющего (UC-7) — 2-я очередь

Как управляющий ТЦ (R-2), я хочу видеть, сколько гостей посещают 1, 2 и 3+ арендатора, чтобы оценить эффект платформы.

**Acceptance Criteria**
1. Given выбран период, When R-2 открывает сводку, Then показано распределение гостей по числу арендаторов (1 / 2 / 3+) и доля cross-tenant penetration (≥ 2 арендаторов).
2. Given выбран период с гранулярностью месяц, When построен график, Then видна динамика каждой доли за период.
3. Given метрики успеха, When открыта вкладка «Метрики», Then первой показана главная метрика — первая покупка гостя у нового арендатора (A-17), а также match rate, доля оспоренных склеек, время сегмент → выгрузка, конверсия и инкремент (если есть данные).
4. Негатив: Given `from > to` или период > 24 месяцев, Then 400; Given роль R-3, Then 403; Given нет данных, Then пустое состояние.

**Экран React:** `/dashboard` (карточки, столбчатая диаграмма Recharts, график динамики) — `features/dashboard`.
**Эндпоинты Go:** `GET /dashboard/penetration`, `GET /dashboard/metrics`; `handler.DashboardHandler` → `service.DashboardService` → `repository.AnalyticsRepo` (агрегирующие запросы по `event`, `merge_candidate`, `export_job`).

---

## US-8. Агрегаты арендатора (UC-8) — 2-я очередь

Как арендатор (R-3), я хочу видеть, сколько новых гостей привёл мне комплекс, чтобы обосновать своё участие в обмене данными.

> Отдельный полноценный кабинет арендатора заказчику не требуется (FAQ В-6, A-23). US-8 остаётся в объёме как разграничение прав и дополнительная ценность для управляющего (R-2 показывает арендатору его выгоду), но при нехватке времени снимается без влияния на ядро продукта.

**Acceptance Criteria**
1. Given пользователь R-3 привязан к арендатору T, When он открывает «Мои показатели», Then видит число новых гостей, пришедших через комплекс (определение A-13), и динамику по месяцам только для T.
2. Given пользователь R-3 запрашивает `/tenants/{чужой id}/aggregates`, When запрос обработан, Then ответ 403, в ответе нет данных; в интерфейсе нет навигации на чужие страницы.
3. Given R-3 вызывает любой другой эндпоинт (профили, сегменты, выгрузки, справочники кроме собственной карточки арендатора), When запрос обработан, Then 403.
4. Given значение агрегата меньше k = 5, When оно рассчитывается, Then скрывается («менее 5»), чтобы нельзя было вычислить отдельного гостя.
5. Given R-2 открывает агрегаты любого арендатора (например, при продлении договора), When запрос выполнен, Then ответ 200.
6. Негатив: Given токен R-3 без `tenant_id` в claims, Then 403; Given арендатор деактивирован, Then 404.

**Экран React:** `/tenant` (карточки показателей и график; для R-3 единственный экран кроме `/login`; в меню нет остальных разделов) — `features/tenant`. Для R-2 — переключатель арендатора на `/dashboard`.
**Эндпоинты Go:** `GET /tenants/{id}/aggregates`, `GET /tenants/{id}`, `GET /auth/me`; `handler.TenantHandler` → `service.TenantAggregateService` (сверка `claims.tenant_id == id`) → `repository.TenantAggregateRepo` (`WHERE tenant_id = :claimTenant`, без выдачи `guest_id`).

---

## US-9. Качество данных по источникам (UC-9) — 2-я очередь

Как аналитик УК (R-4), я хочу видеть качество данных по каждому источнику, чтобы вовремя замечать сбои в выгрузках.

**Acceptance Criteria**
1. Given список источников, When R-4 открывает «Качество данных», Then по каждому виден: дата последней успешной загрузки, доля отклонённых строк за последние 30 дней, число спорных (pending) профилей, связанных с источником.
2. Given доля отклонённых выше порога 10% или данные старше 7 дней, When таблица отображается, Then строка помечена предупреждением.
3. Given выбран источник, When открыта деталь, Then виден график отклонённых по загрузкам и ссылка на конкретную загрузку.
4. Негатив: Given источник без загрузок, Then «данных нет», не ошибка; Given роль R-1/R-3, Then 403.

**Экран React:** `/data-quality` (таблица TanStack Table, детали источника) — `features/dq`.
**Эндпоинты Go:** `GET /data-quality/sources`, `GET /data-quality/sources/{id}`; `handler.DataQualityHandler` → `service.DataQualityService` → `repository.ImportBatchRepo`, `RejectedRowRepo`, `MergeCandidateRepo`.

---

## US-10. Лимит контактов (UC-10) — MVP

Как маркетолог УК (R-1), я хочу ограничить частоту контактов, чтобы не выжигать базу повторными отправками.

> Заказчик не задаёт жёсткий лимит частоты (FAQ В-9), поэтому функция реализуется как **настраиваемая**: при отсутствии правил ограничение не применяется, а рекомендуемая настройка показывается в интерфейсе как подсказка (A-12).

**Acceptance Criteria**
1. Given R-1 задаёт правило «не более 2 контактов за 7 дней» (общее или для канала), When правило сохранено, Then оно действует на следующие выгрузки. Given правил нет, When создаётся выгрузка, Then ограничение не применяется и в предпросмотре указано «лимит не настроен».
2. Given гость получил 2 контакта за 7 дней, When формируется новая выгрузка, Then гость исключён из получателей, в предпросмотре указано «исключено лимитом: N».
3. Given применимо несколько правил, When выгрузка проверяется, Then применяется самое строгое.
4. Given гость в контрольной группе, When считается лимит, Then он в лимите не учитывается и ContactLog для него не создаётся.
5. Негатив: Given выгрузка, где лимит исключил всех, When создаётся, Then 409 с причиной «все гости превысили лимит контактов (правило: …)»; Given `max_contacts < 1` или `period_days < 1`, Then 400; Given роль R-4, Then 403 на изменение (чтение доступно).

**Экран React:** `/settings/contact-limits` (таблица правил, форма) и счётчик «исключено лимитом» в мастере выгрузки — `features/export`.
**Эндпоинты Go:** `GET/POST /contact-limits`, `PATCH/DELETE /contact-limits/{id}`, `POST /exports/preview`, `POST /exports`; `handler.ContactLimitHandler` → `service.ContactLimitService` → `repository.ContactLimitRepo`, `ContactLogRepo`.

---

## US-11. Подтверждение и разъединение склеек (UC-11) — MVP

Как аналитик УК (R-4), я хочу подтверждать или отклонять спорные склейки и разъединять ошибочно объединённые профили, чтобы в одном профиле не оказалось двух разных людей.

**Acceptance Criteria**
1. Given спорная пара, When R-4 открывает очередь, Then пара показана с причиной совпадения (общий ключ, различающиеся имена, число событий у каждого).
2. Given R-4 подтверждает склейку, When действие выполнено, Then события и идентификаторы объединены в один профиль за один шаг (одна транзакция), запись добавлена в аудит.
3. Given R-4 отклоняет пару, When действие выполнено, Then пара получает запрет склейки (E-12) и больше не появляется в очереди.
4. Given объединённый профиль (в том числе автосклейка), When R-4 нажимает «Разъединить», Then события и идентификаторы возвращаются прежним профилям (`origin_guest_id`), создаётся запрет склейки, запись в аудит.
5. Given пара под запретом, When новая загрузка снова даёт общий телефон, Then автосклейка не происходит и в очередь пара не возвращается.
6. Негатив: Given пара уже решена другим пользователем, When повторно подтверждают, Then 409; Given ручное объединение пары под запретом, Then 409 (запрет снимается отдельным `DELETE /unmerge-blocks/{id}`); Given роль R-3, Then 403.

**Экран React:** `/identity/queue` (очередь с фильтрами, сравнение двух профилей, кнопки «Объединить / Отклонить / Разъединить»), кнопка «Разъединить» в профиле — `features/profiles`.
**Эндпоинты Go:** `GET /merge-candidates`, `GET /merge-candidates/{id}`, `POST /merge-candidates/{id}/confirm`, `/reject`, `/revert`, `POST /guests/{id}/merge`, `GET /unmerge-blocks`, `DELETE /unmerge-blocks/{id}`; `handler.IdentityHandler` → `service.IdentityService` → `internal/identity` (`IdentityResolver`, `MergeExecutor`) → `repository.MergeCandidateRepo`, `UnmergeBlockRepo`, `GuestRepo`, `EventRepo`, `GuestIdentifierRepo`, `AuditRepo`.

---

## Сводка US → экран → эндпоинты

| US | UC | Экран (React) | Основные эндпоинты | Роли |
|---|---|---|---|---|
| US-1 | UC-1 | `/import`, `/import/:batchId` | `POST /imports`, `GET /imports/{id}` | R-4 |
| US-2 | UC-2 | `/guests`, `/guests/:guestId` | `GET /guests/{id}` | R-1, R-4 |
| US-3 | UC-3 | `/segments/new`, `/segments/:id` | `POST /segments/preview`, `POST /segments` | R-1, R-4 |
| US-4 | UC-4 | вкладка в профиле | `GET /guests/{id}/recommendations` | R-1, R-4 |
| US-5 | UC-5 | `/exports/new` | `POST /exports` | R-1 |
| US-6 | UC-6 | индикаторы | `GET /segments/{id}` | R-1, R-4 (чтение) |
| US-7 | UC-7 | `/dashboard` | `GET /dashboard/penetration` | R-2 (R-1, R-4) |
| US-8 | UC-8 | `/tenant` | `GET /tenants/{id}/aggregates` | R-3 (own) |
| US-9 | UC-9 | `/data-quality` | `GET /data-quality/sources` | R-4, R-2 |
| US-10 | UC-10 | `/settings/contact-limits` | `/contact-limits`, `POST /exports` | R-1 |
| US-11 | UC-11 | `/identity/queue` | `POST /merge-candidates/{id}/confirm` | R-4 |

## Чекпойнт
Создано: US-1…US-11 с AC (Given/When/Then, позитив + негатив), экраны React, эндпоинты Go. Ссылки на A-13, D-14; отдельно использованы Q-13, Q-18 и порог 10% / 7 дней (D-21 ниже в open-questions).
