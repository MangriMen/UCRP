# 09. Sequence Diagrams

Mermaid `sequenceDiagram`. Участники используют единые имена: `SPA` (React), `Handler` (`internal/handler`), `Service` (`internal/service`), `Repo` (`internal/repository`), `DB`, `Queue` (`job_queue`, `platform/queue`), `Worker` (`cmd/worker`). Во всех запросах: заголовок `Authorization: Bearer <JWT>`; middleware кладёт `Principal{user_id, role, tenant_id}` и `request_id` в `context.Context`; ошибки имеют единый формат `{"error":{code,message,details,request_id}}` (реестр, шапка). Ограничение Mermaid: в тексте сообщений нельзя использовать точку с запятой, поэтому сложные условия записаны через запятую. Бюджет таймаутов (Q-23): запрос по умолчанию 30 с, предпросмотр сегмента 10 с, синхронная выгрузка 25 с, внешний HTTP коннектора 30 с, задание воркера 15 мин; превышение даёт 504 с кодом `timeout`.

## SD-1. Вход в систему (UC-16)

```mermaid
sequenceDiagram
  autonumber
  actor U as Пользователь R-1..R-4
  participant SPA as SPA auth
  participant H as AuthHandler
  participant S as AuthService
  participant R as UserRepo
  participant DB as DB
  U->>SPA: email и пароль
  SPA->>H: POST /auth/login
  H->>H: валидация тела, rate limit
  alt тело неверно
    H-->>SPA: 400 validation
  else слишком много попыток
    H-->>SPA: 429
  else
    H->>S: Login(ctx, email, password)
    S->>R: GetByEmail(ctx, email)
    R->>DB: SELECT app_user
    DB-->>R: строка или пусто
    alt нет пользователя, пароль неверен или is_active=false
      S-->>H: ErrUnauthorized
      H-->>SPA: 401 invalid_credentials
    else успех
      S->>R: TouchLastLogin(ctx, id)
      S-->>H: access JWT 15 мин с role и tenant_id, refresh
      H-->>SPA: 200 access, пользователь с эффективными правами, Set-Cookie refresh httpOnly
    end
  end
  Note over SPA,H: Через 15 мин access истекает, любой запрос даёт 401
  SPA->>H: POST /auth/refresh (cookie)
  H->>S: Refresh(ctx, token)
  alt refresh просрочен или отозван
    H-->>SPA: 401, SPA переходит на экран входа
  else
    H-->>SPA: 200 новый access и актуальные права
  end
```

Пояснение. Логин сверяет хэш пароля и активность учётной записи; одинаковый 401 на неверный email и неверный пароль не раскрывает существование пользователя. В `access` кладутся `role` и, для R-3, `tenant_id` — это единственный источник арендатора для всех последующих проверок. `refresh` хранится в httpOnly-cookie и обновляется эндпоинтом `/auth/refresh`; при 401 SPA один раз пытается обновить токен (single-flight, D-29), затем отправляет на вход. Все обращения к БД принимают `ctx`.

## SD-2. Загрузка файла: async 202, статус и построчные ошибки (UC-1, UC-6)

```mermaid
sequenceDiagram
  autonumber
  actor A as Аналитик R-4
  participant SPA as SPA import
  participant H as ImportHandler
  participant S as ImportService
  participant R as Repo
  participant Q as Queue
  participant W as Worker
  A->>SPA: файл и источник
  SPA->>H: POST /imports multipart
  H->>H: JWT, RequireRole R-4, размер
  alt размер больше лимита
    H-->>SPA: 413
  else
    H->>S: Accept(ctx, sourceID, file)
    S->>R: GetSource(ctx, id)
    alt источник не найден
      H-->>SPA: 404
    else формат неверен
      H-->>SPA: 422 bad_format
    else дубликат sha256
      S->>R: FindBatchByHash(ctx, sourceID, hash)
      H-->>SPA: 200 duplicate=true, batch_id
    else новый файл
      S->>R: Tx: InsertBatch queued, Audit, Enqueue
      R->>Q: job import.process
      H-->>SPA: 202 batch_id, status queued
    end
  end
  loop каждые 2 с до completed
    SPA->>H: GET /imports/id
    H-->>SPA: 200 status, total, accepted, rejected, duplicate
  end
  Q->>W: take import.process
  W->>R: UpdateStatus running
  W->>W: fork чанков, валидация, join
  W->>R: InsertRejectedRows, InsertEvents ignore duplicates
  W->>Q: enqueue identity.resolve по чанкам
  W->>R: UpdateCounters, status completed
  SPA->>H: GET /imports/id/rejected-rows
  H-->>SPA: 200 строки с причинами или CSV
```

Пояснение. Клиент получает 202 сразу и далее только читает статус: тело запроса не ждёт обработки (UC-1, US-1). Повторная загрузка того же файла даёт 200 с `duplicate=true` и существующим `batch_id` (D-9); дублей событий нет также на уровне уникального индекса события, поэтому даже изменённый файл с частично теми же строками не создаёт дублей. Построчные ошибки (`invalid_date`, `no_external_id`, `invalid_amount`) попадают в `rejected_row`, их можно скачать CSV. Задание `import.process` безопасно перезапускается (`dedup_key import:{batch_id}`).

## SD-3. API-коннектор подтягивает события (UC-14)

```mermaid
sequenceDiagram
  autonumber
  participant Sch as Scheduler cmd/worker
  participant Q as Queue
  participant W as Worker ingestion
  participant C as ImportService.processPull
  participant X as Система арендатора
  participant R as Repo
  actor A as Аналитик R-4
  participant H as ImportHandler
  Sch->>Q: enqueue connector.pull pull:source:window
  A->>H: POST /data-sources/id/sync
  alt синхронизация уже идёт, dedup_key занят
    H-->>A: 409 sync_in_progress
  else
    H->>Q: enqueue connector.pull
    H-->>A: 202
  end
  Q->>W: take connector.pull
  W->>R: Insert ImportBatch mode=pull, window_from, window_to
  W->>C: Pull(ctx 30 с, window, cursor)
  C->>X: GET events
  alt таймаут или 5xx
    X-->>C: ошибка
    C-->>W: ErrTemporary
    W->>Q: retry, run_after + backoff, attempts+1
  else 401 от системы арендатора
    C-->>W: ErrAuth
    W->>R: batch failed, error_message
  else страница данных
    X-->>C: события JSON
    C-->>W: Rows (общий формат)
    W->>R: validate, normalize, insert ignore duplicates
    W->>Q: enqueue identity.resolve
    W->>R: batch completed, data_source.last_import_at
  end
```

Пояснение. Pull запускается по расписанию (раз в час, Q-10) или вручную; ручной вызов возвращает 409, если для того же окна уже есть активное задание (уникальный `dedup_key`). `ImportService.processPull` запрашивает `config.url` источника (HTTP GET, ответ — CSV) и приводит его к общему виду строк, после чего конвейер валидации и вставки тот же, что у файлов (AD-1b), поэтому идемпотентность и счётчики идентичны. Сетевые сбои повторяются с экспоненциальной паузой до N попыток, постоянные ошибки (неверные учётные данные источника) переводят загрузку в `failed` и отражаются в качестве источников (UC-9). Внешний вызов имеет собственный `ctx` с таймаутом 30 с, наследующий отмену задания.

## SD-4. Сегмент с предпросмотром размера (UC-3, UC-13)

```mermaid
sequenceDiagram
  autonumber
  actor M as Маркетолог R-1
  participant SPA as SPA segments
  participant H as SegmentHandler
  participant S as SegmentService
  participant K as Cache
  participant E as SegmentEvaluator
  participant DB as DB
  participant Q as Queue
  participant W as Worker segments
  M->>SPA: правило был в A и B, не был в C
  SPA->>H: POST /segments/preview (debounce 400 мс)
  H->>S: Preview(ctx 10 с, rule)
  S->>S: валидация правила
  alt правило неверно или арендатор не найден
    H-->>SPA: 400 с details по полям
  else
    S->>K: Get(hash правила)
    alt есть в кэше
      K-->>S: размер и примеры
    else
      S->>E: Preview(ctx, rule)
      E->>DB: COUNT и LIMIT 20
      alt дедлайн 10 с
        H-->>SPA: 504 timeout
      else
        DB-->>E: данные
        S->>K: Set(hash, ttl 60 с)
      end
    end
    H-->>SPA: 200 size, examples
  end
  M->>SPA: Сохранить
  SPA->>H: POST /segments
  H->>S: Create(ctx, rule)
  S->>DB: Tx: segment, version 1 pending
  S->>Q: enqueue segment.recompute segment:id:1
  H-->>SPA: 201
  Q->>W: take job
  W->>E: Evaluate(ctx, version)
  W->>DB: INSERT segment_member, status ready, computed_at
  loop опрос
    SPA->>H: GET /segments/id
    H-->>SPA: 200 status, member_count
  end
```

Пояснение. Предпросмотр не пишет в БД и защищён кэшем и дедлайном 10 с; ответ 400 содержит поля с ошибками правила, 504 — подсказку сузить период (D-6). Сохранение быстрое (201), пересчёт состава выполняет воркер; пока статус `pending` или `computing`, выгрузка из сегмента вернёт 409 (SD-5). Правка через `PUT /segments/{id}` повторяет ту же схему с `version_no+1`. Новое событие (UC-6) ставит `segment.recompute` для сегментов, затронутых гостем, и обновляет `computed_at`.

## SD-5. Выгрузка сегмента: лимит контактов и 409 (UC-5, UC-10)

```mermaid
sequenceDiagram
  autonumber
  actor M as Маркетолог R-1
  participant SPA as SPA export
  participant H as ExportHandler
  participant S as ExportService
  participant L as ContactLimitService
  participant R as Repo
  participant DB as DB
  M->>SPA: канал, формат, контроль 10 процентов
  SPA->>H: POST /exports/preview
  H->>S: Preview(ctx, segmentID, channel)
  S-->>SPA: 200 четыре числа
  SPA->>H: POST /exports Idempotency-Key
  alt заголовка нет
    H-->>SPA: 400
  else
    H->>S: Create(ctx 25 с, cmd)
    S->>R: FindByKey(ctx, user, key)
    alt ключ уже есть
      S-->>SPA: 201 существующая выгрузка
    else
      S->>R: GetSegmentVersion
      alt сегмента нет
        H-->>SPA: 404
      else версия не ready
        H-->>SPA: 409 segment_not_ready
      else
        S->>R: фильтр гостей с контактом канала
        S->>L: Apply(ctx, guests, channel)
        L->>R: счёт contact_log за период по правилам
        L-->>S: допущенные и заблокированные
        alt правил лимита нет
          Note over S: ограничение не применяется (A-12)
        end
        alt допущенных нет
          H-->>SPA: 409 contact_limit_exceeded
        else
          S->>R: Insert(job, members)
          R->>DB: Tx: export_job, export_member (control по seed, 0 = без группы)
          S->>R: contact_log InsertBatch (отдельный запрос, без контрольной группы)
          H-->>SPA: 201 ready
        end
      end
    end
  end
  SPA->>H: GET /exports/id/download
  alt статус не ready
    H-->>SPA: 409
  else
    H-->>SPA: 200 CSV без контрольной группы
  end
```

Пояснение. Предпросмотр показывает, сколько гостей будет исключено лимитом, до создания записи. Повтор `POST /exports` с тем же `Idempotency-Key` не создаёт вторую выгрузку и не дублирует контакты. Лимит применяется только при наличии правил: заказчик не задаёт жёсткий лимит частоты (FAQ В-9, A-12), поэтому при пустой таблице `contact_limit` выгрузка проходит без ограничения. Если правила есть, лимит проверяется один раз, при отборе получателей (`blockedForChannel`), до записи; повторной проверки внутри транзакции нет, поэтому две параллельные выгрузки с разными ключами теоретически могут обойти лимит (D-26 в коде не реализован). `export_job` и `export_member` пишутся одной транзакцией `ExportSQL.Insert` (уровень изоляции по умолчанию, без `BEGIN IMMEDIATE`), `contact_log` добавляется следующим отдельным запросом, запись в `audit_log` при выгрузке не делается. Повторный запрос с тем же ключом возвращает существующую выгрузку, а `UNIQUE (created_by, idempotency_key)` в БД защищает от дубля при гонке. Если ни один гость с контактом не прошёл лимит, возвращается 409 `contact_limit_exceeded`. Контрольная группа опциональна: при `control_percent = 0` она не выделяется, иначе детерминирована по `control_seed` (D-17) и не пишется в `contact_log` (A-19). CSV содержит только не-контрольных получателей; API-выгрузка для провайдера — `GET /exports/{id}/members`.

## SD-6. Ручное объединение и разъединение профилей (UC-11)

```mermaid
sequenceDiagram
  autonumber
  actor A as Аналитик R-4
  participant SPA as SPA profiles
  participant H as IdentityHandler
  participant S as IdentityService
  participant R as Repo
  participant DB as DB
  A->>SPA: открыть пару в очереди
  SPA->>H: GET /merge-candidates/id
  H-->>SPA: 200 пара рядом, причина
  A->>SPA: Объединить
  SPA->>H: POST /merge-candidates/id/confirm
  H->>S: Confirm(ctx, id, actor)
  S->>R: GetCandidate, HasBlock(a, b)
  alt уже решено или есть запрет
    H-->>SPA: 409
  else
    S->>DB: BEGIN
    S->>DB: UPDATE event, guest_identifier SET guest_id=winner, origin_guest_id сохранён
    S->>DB: loser status=merged, merged_into_id, resolution=confirmed, audit merge.confirm
    S->>DB: COMMIT
    H-->>SPA: 200
  end
  Note over A,SPA: Позже выяснилось, что это разные люди
  A->>SPA: Разъединить
  SPA->>H: POST /merge-candidates/id/revert
  H->>S: Revert(ctx, id, actor)
  alt решение не confirmed и не auto_merged
    H-->>SPA: 409
  else
    S->>DB: BEGIN
    S->>DB: события и идентификаторы с origin_guest_id=loser вернуть loser
    S->>DB: loser status=active, merged_into_id=NULL, resolution=reverted
    S->>DB: INSERT merge_block, audit merge.revert
    S->>DB: COMMIT
    S->>R: Enqueue segment.recompute и recommend.refresh для обоих
    H-->>SPA: 200
  end
  A->>SPA: ручное объединение другой пары
  SPA->>H: POST /guests/id/merge target_guest_id
  alt цель не найдена
    H-->>SPA: 404
  else пара под запретом
    H-->>SPA: 409 merge_blocked
  else сама с собой
    H-->>SPA: 400
  end
```

Пояснение. Объединение и разъединение — одна транзакция, поэтому профиль никогда не бывает наполовину слитым. Обратимость основана на `origin_guest_id` в `event` и `guest_identifier` (D-5), включая автообъединения (`auto_merged`). Revert и reject создают запись в `merge_block`, после чего автообъединение эту пару не повторит; снять запрет может только R-4 (`DELETE /merge-blocks/{id}`). После изменения состава профилей ставятся пересчёты сегментов и рекомендаций обоих гостей. Любое действие пишет `audit_log` с `request_id`. R-1 видит очередь только на чтение, при `confirm` получит 403.

## SD-7. Сквозная цепочка ошибок 400 / 401 / 403 / 404 / 409 и изоляция R-3

```mermaid
sequenceDiagram
  autonumber
  actor U as Пользователь
  participant SPA as SPA api client
  participant MW as Middleware auth
  participant H as Handler
  participant S as Service
  participant R as Repo
  U->>SPA: действие
  SPA->>MW: запрос с Bearer
  alt токена нет, просрочен или подпись неверна
    MW-->>SPA: 401, SPA вызывает refresh или переходит на вход
  else
    MW->>H: ctx с Principal, request_id
    alt роль не в списке реестра для маршрута
      H-->>SPA: 403 forbidden_role
    else тело или параметры неверны
      H-->>SPA: 400 с details
    else
      H->>S: вызов use case(ctx, args)
      alt R-3 и tenant_id в пути не равен claim
        S-->>H: ErrForbiddenTenant
        H-->>SPA: 403 (D-13)
      else нет разрешения или подходящего гранта
        S-->>H: ErrForbidden, access.Decide
        H-->>SPA: 403 forbidden_permission или forbidden_scope
      else
        S->>R: запрос с tenantID из Principal
        alt не найдено
          R-->>S: ErrNotFound
          H-->>SPA: 404
        else конфликт состояния
          S-->>H: ErrConflict
          H-->>SPA: 409
        else
          R-->>S: агрегаты с порогом k=5
          H-->>SPA: 200
        end
      end
    end
  end
```

Пояснение. Проверки идут в фиксированном порядке: аутентификация (401), роль по реестру (403), валидация (400), бизнес-правила (403 для чужого арендатора, 404, 409). Изоляция R-3 двойная: `service` сравнивает `tenant_id` из пути с claim и возвращает 403 (D-13), а `access.Decide` отклоняет запрос с `forbidden_tenant`, если `Resource.TenantID` не совпадает с `Principal.TenantID`. Данные арендатора читают репозитории, принимающие `tenantID` параметром и подставляющие его в SQL-фильтр `tenant_id` (например, `TenantGuestSQL` фильтрует события по `e.tenant_id`); сервис передаёт туда `principal.TenantID`, а не значение из запроса. R-3 допущен к `GET /tenants`, `GET /tenants/{id}`, `GET /tenants/{id}/aggregates`, `GET /auth/me`, а к чтению гостей и событий — только по гранту, выданному R-2 (`access.Decide`); ячейки меньше k=5 скрываются (Q-18). Frontend прячет недоступные пункты меню, но решающая проверка всегда на backend. Коды ошибок и тело — единый формат реестра.

## SD-8. Таймауты и распространение context.Context

```mermaid
sequenceDiagram
  autonumber
  participant SPA as SPA
  participant MW as Middleware timeout
  participant H as Handler
  participant S as Service
  participant R as Repo
  participant DB as DB
  participant W as Worker
  SPA->>MW: POST /guests/id/recommendations/refresh
  MW->>H: ctx WithTimeout 10 с, request_id
  H->>S: Refresh(ctx, guestID)
  S->>R: LoadFeed(ctx)
  R->>DB: QueryContext
  Note over DB: запрос не уложился в срок
  DB-->>R: context deadline exceeded
  R-->>S: ctx.Err
  S-->>H: ErrTimeout, транзакция откатана
  H-->>SPA: 504 timeout, request_id
  Note over SPA: клиент закрыл вкладку
  SPA--xMW: отмена соединения
  MW->>H: ctx.Done
  H->>S: отмена доходит до Repo, запрос прерван
  Note over W: SIGTERM воркера
  W->>W: cancel корневого ctx, grace 30 с
  W->>R: задание не завершено, status queued, run_after
```

Пояснение. Каждый входящий запрос получает собственный `ctx` с дедлайном (30 с по умолчанию, 10 с для предпросмотра и принудительного пересчёта, 25 с для синхронной выгрузки); `ctx` первым аргументом проходит через `handler → service → repository` до `QueryContext`/`ExecContext`, поэтому отмена клиента и истечение срока прерывают запросы и откатывают транзакции. Таймаут превращается в 504 `timeout`, обрыв клиента не порождает ответа. Воркер на SIGTERM отменяет корневой `ctx`, доводит текущую транзакцию в пределах grace-периода и возвращает недоделанное задание в очередь; повтор безопасен благодаря идемпотентности (`dedup_key`, уникальный индекс события). Внешние вызовы коннектора имеют вложенный `ctx` 30 с.

## SD-9. Профиль гостя, рекомендации и видимое обновление (UC-2, UC-4, UC-6)

```mermaid
sequenceDiagram
  autonumber
  actor M as Маркетолог R-1
  participant SPA as SPA profiles
  participant H as GuestHandler
  participant S as ProfileService
  participant RS as RecommendationService
  participant R as Repo
  participant Q as Queue
  M->>SPA: поиск по телефону в любом формате
  SPA->>H: GET /guests?q=
  H->>S: Search(ctx, q)
  S->>R: FindByKeys(ctx, нормализованный ключ)
  H-->>SPA: 200 список
  M->>SPA: открыть профиль
  SPA->>H: GET /guests/id
  alt гостя нет
    H-->>SPA: 404
  else гость слит
    H-->>SPA: 200 с merged_into_id, SPA открывает актуальный профиль
  else
    H->>S: Get(ctx, id)
    S->>R: Guest, Identifiers, tenant_count
    H-->>SPA: 200 профиль
  end
  SPA->>H: GET /guests/id/events
  H-->>SPA: 200 лента, дата убывания
  SPA->>H: GET /guests/id/recommendations
  H->>RS: ForGuest(ctx, id)
  RS->>R: Recommendations не более 3 с обоснованием
  H-->>SPA: 200 список или пустое состояние
  Note over Q: пришло новое событие, задания identity.resolve, segment.recompute, recommend.refresh
  loop опрос TanStack Query каждые 5 с
    SPA->>H: GET /guests/id/recommendations
    H-->>SPA: 200 с новым computed_at, метка обновлено
  end
```

Пояснение. Поиск нормализует введённое значение теми же правилами, что и при импорте, поэтому `+7 900…` и `8900…` находят одного гостя. Для слитого профиля ответ несёт `merged_into_id`, и интерфейс переходит на актуальный профиль; для несуществующего возвращается 404, для R-3 маршруты `/guests*` открыты только по гранту и только по данным собственного арендатора, иначе 403. Профиль, лента и рекомендации читаются отдельными быстрыми запросами; чтение рекомендаций не пересчитывает их (D-12). После нового события фоновые задания обновляют ленту, сегменты и рекомендации, а интерфейс видит изменение при опросе с интервалом 5 с (D-22), без ручного запуска (UC-6).
