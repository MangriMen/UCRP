# Фрагменты OpenAPI (запросы и ответы ключевых эндпоинтов)

Формат: проектные схемы контрактов, не реализация. Полный перечень из 90 эндпоинтов, роли и коды — `endpoint-registry.md` (источник истины). Базовый путь `/api/v1`, JSON UTF-8, время RFC 3339 UTC, идентификаторы — строки UUIDv7 (в примерах сокращены: `g-1001`). Авторизация: `Authorization: Bearer <JWT>`; JWT-claims: `sub`, `role` (`marketer|manager|tenant|analyst`), `tenant_id` (только для `tenant`), `exp`.

## 1. Единый формат ошибки

Все ответы 4xx/5xx имеют одну схему; `code` — машинный код, `message` — текст на русском, `details` — ошибки по полям (для 400/422), `request_id` совпадает с заголовком `X-Request-Id` и записью `audit_log.request_id`.

```json
{
  "error": {
    "code": "contact_limit_exceeded",
    "message": "Все гости превысили лимит контактов (правило: 2 SMS за 7 дней)",
    "details": [{ "field": "segment_id", "reason": "blocked_by_limit=1, no_contact=1" }],
    "request_id": "req-01J9Z6"
  }
}
```

| HTTP | `error.code`                                                                                                               | Когда                                                                         |
| ---- | -------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------- |
| 400  | `validation_failed`, `idempotency_key_required`, `invalid_rule`                                                            | неверное тело или параметры, нет `Idempotency-Key`, тенант в «был» и «не был» |
| 401  | `unauthorized`, `invalid_credentials`, `token_expired`                                                                     | нет или просрочен токен, неверный логин                                       |
| 403  | `forbidden_role`, `forbidden_tenant`                                                                                       | роль не в списке реестра; R-3 запросил чужой `tenant_id`                      |
| 404  | `not_found`                                                                                                                | объект не найден                                                              |
| 409  | `contact_limit_exceeded`, `segment_not_ready`, `merge_blocked`, `already_resolved`, `sync_in_progress`, `export_not_ready` | конфликт состояния                                                            |
| 413  | `file_too_large`                                                                                                           | более 100 тыс. строк или лимит размера                                        |
| 422  | `bad_format`, `missing_columns`                                                                                            | неверный формат файла или нет обязательных колонок                            |
| 429  | `rate_limited`                                                                                                             | слишком много попыток входа                                                   |
| 504  | `timeout`                                                                                                                  | истёк дедлайн `context`                                                       |

## 2. Аутентификация

`POST /auth/login` — публичный.

```json
{ "email": "analyst@mall.example", "password": "string" }
```

Ответ 200: тело `{ "access_token": "jwt", "expires_in": 900, "user": { "id": "u-4", "role": "analyst", "tenant_id": null } }` и `Set-Cookie: refresh=...; HttpOnly; Secure; SameSite=Strict`. 400 — неверное тело, 401 `invalid_credentials` — неверный email, пароль или неактивная учётная запись (одинаковый ответ), 429 — превышен лимит попыток. `POST /auth/refresh` (cookie) → 200 с новым `access_token` или 401.

## 3. Импорт (UC-1)

`POST /imports` — `multipart/form-data`: `data_source_id` (string), `file` (CSV/XLSX, до 100 тыс. строк). Роль: R-4.

Ответ 202 (новый файл поставлен в очередь):

```json
{ "batch_id": "b-101", "status": "queued", "duplicate": false }
```

Ответ 200 (тот же sha256 для того же источника, новых событий нет):

```json
{ "batch_id": "b-101", "status": "completed", "duplicate": true }
```

Прочие ответы: 400 (нет полей формы), 404 (источник), 413, 422 — единый формат ошибки.

`GET /imports/{id}` — статус для опроса каждые 2 с. Роли: R-4, R-2.

```json
{
  "id": "b-101",
  "data_source_id": "ds-pos",
  "mode": "file",
  "status": "completed_with_errors",
  "total": 1000,
  "accepted": 940,
  "rejected": 12,
  "duplicate": 48,
  "started_at": "2026-09-30T10:00:03Z",
  "finished_at": "2026-09-30T10:00:41Z"
}
```

Инвариант: `accepted + rejected + duplicate = total`. Статусы: `queued`, `running`, `completed`, `completed_with_errors`, `failed`.

`GET /imports/{id}/rejected-rows?page=1&format=json|csv` — R-4.

```json
{
  "items": [
    {
      "row_number": 17,
      "error_code": "bad_date",
      "error_message": "Не распознана дата «31.31.2026»",
      "raw_line": "..."
    },
    {
      "row_number": 52,
      "error_code": "no_external_id",
      "error_message": "Нет номера операции",
      "raw_line": "..."
    }
  ],
  "page": 1,
  "page_size": 50,
  "total": 12
}
```

`POST /data-sources/{id}/sync` — R-4; ответ 202 `{ "batch_id": "b-102", "status": "queued" }`, 404, 409 `sync_in_progress`.

## 4. Профиль и рекомендации (UC-2, UC-4)

`GET /guests/{id}` — R-1, R-4.

```json
{
  "id": "g-1001",
  "display_name": "Анна К.",
  "status": "active",
  "merged_into_id": null,
  "tenant_count": 2,
  "tenants": ["Кофейня Bean", "ФитЗона"],
  "identifiers": [
    { "id_type": "phone", "value": "79001234567", "source_id": "ds-pos" },
    { "id_type": "email", "value": "anna@example.com", "source_id": "ds-crm" }
  ],
  "segments": ["s-1"]
}
```

Для слитого профиля возвращается `status: "merged"` и `merged_into_id` актуального гостя.

`GET /guests/{id}/events` — лента, сортировка по `occurred_at` по убыванию.

```json
{
  "items": [
    {
      "id": "e-3",
      "type": "purchase",
      "tenant": "Кофейня Bean",
      "project": null,
      "occurred_at": "2026-09-20T09:15:00Z",
      "amount_minor": 28000,
      "currency": "RUB"
    }
  ],
  "next_page": null
}
```

`GET /guests/{id}/recommendations` — не более 3 элементов; арендаторов, где гость уже был, в списке нет.

```json
{
  "guest_id": "g-1001",
  "computed_at": "2026-10-30T09:00:00Z",
  "algorithm_version": "v1",
  "items": [
    {
      "rank": 1,
      "tenant_id": "t-cine",
      "tenant_name": "Кинотеатр Синема",
      "score": 0.78,
      "explanation": "Похожие гости ходят в Кинотеатр Синема, Спорт и Развлечения смежны, недавняя активность, 3 этаж",
      "factors": [
        { "code": "similar_guests", "raw_value": 0.85, "weight": 0.4, "contribution": 0.34 },
        { "code": "category_adjacency", "raw_value": 0.7, "weight": 0.25, "contribution": 0.175 },
        { "code": "recency", "raw_value": 0.9, "weight": 0.25, "contribution": 0.225 },
        { "code": "distance", "raw_value": 0.4, "weight": 0.1, "contribution": 0.04 }
      ]
    }
  ]
}
```

`POST /guests/{id}/recommendations/refresh` — R-1, R-4; синхронный пересчёт с таймаутом 10 с: 200 (тело как выше), 404, 504 `timeout`.

## 5. Сегменты (UC-3)

`POST /segments/preview` — R-1, R-4; не пишет в БД, таймаут 10 с.

```json
{
  "rule": {
    "was": ["t-bean", "t-fit"],
    "was_not": ["t-cine"],
    "period_days": 90,
    "require_contact": "phone"
  }
}
```

Ответ 200: `{ "size": 5, "examples": [ { "guest_id": "g-1001", "display_name": "Анна К." } ], "rule_text": "Был в Кофейня Bean и ФитЗона, не был в Кинотеатр Синема за 90 дней, есть телефон" }` (примеров до 20). 400 `invalid_rule` с `details` (например, арендатор одновременно в `was` и `was_not`), 504 `timeout`.

`POST /segments` — создаёт сегмент и версию 1; тело: `name`, `rule` или `template_code` (`sleeping_neighbor|single_store|churn_risk|newcomer|high_value`). Ответ 201:

```json
{ "id": "s-1", "version_no": 1, "status": "pending", "member_count": null, "computed_at": null }
```

`GET /segments/{id}` — после пересчёта `status: "ready"`, `member_count`, `computed_at`; статусы версии: `pending`, `computing`, `ready`, `failed`.

## 6. Выгрузка (UC-5, UC-10)

`POST /exports/preview` — R-1.

```json
{ "segment_id": "s-1", "channel": "sms", "control_percent": 33 }
```

Ответ 200 (четыре числа плюс к отправке; сумма равна размеру сегмента):

```json
{ "in_segment": 5, "with_contact": 4, "blocked_by_limit": 1, "control": 1, "to_send": 2 }
```

Здесь `in_segment = no_contact(1) + blocked_by_limit(1) + control(1) + to_send(2)`; поле `with_contact` — гости с контактом канала до применения лимита.

`POST /exports` — R-1; обязателен заголовок `Idempotency-Key`.

```json
{
  "segment_id": "s-1",
  "channel": "sms",
  "format": "csv",
  "control_percent": 33,
  "campaign_id": "c-1"
}
```

| Ответ | Тело / причина                                                                                                                                                                                             |
| ----- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 201   | `{ "id": "x-1", "status": "ready", "total_in_segment": 5, "no_contact_count": 1, "blocked_by_limit_count": 1, "control_count": 1, "exported_count": 2 }`; повтор с тем же ключом возвращает ту же выгрузку |
| 202   | `{ "id": "x-2", "status": "queued" }` — более 20 000 членов, строится заданием `export.build`                                                                                                              |
| 400   | нет `Idempotency-Key` или неверное тело                                                                                                                                                                    |
| 404   | сегмента нет                                                                                                                                                                                               |
| 409   | `segment_not_ready` или `contact_limit_exceeded` (все исключены; `details` содержит разбивку причин)                                                                                                       |
| 504   | `timeout`                                                                                                                                                                                                  |

`GET /exports/{id}/download` — CSV только не-контрольных получателей, колонки `guest_id,channel,contact,segment_name,campaign_id` (Q-20); 409 `export_not_ready`, если статус не `ready`.

`GET /exports/{id}/members?is_control=false` — API-выгрузка для провайдера (токен, Q-24):

```json
{
  "export_id": "x-1",
  "channel": "sms",
  "segment_name": "Спящий сосед, кино",
  "campaign_id": "c-1",
  "items": [{ "guest_id": "g-1001", "contact": "79001234567", "is_control": false }],
  "next_page": null
}
```

В `items` попадают только гости с контактом канала (`sms`, `push` — телефон; `email` — email, Q-21).

## 7. Склейки (UC-11)

`POST /merge-candidates/{id}/confirm` — R-4; тело пустое. Ответ 200: `{ "id": "mc-1", "resolution": "confirmed", "winner_guest_id": "g-2001", "merged_guest_id": "g-2002" }`. 404, 409 `already_resolved` или `merge_blocked`. `POST /merge-candidates/{id}/revert` → 200 `{ "resolution": "reverted", "unmerge_block_id": "ub-1" }`, 409, если решение не `confirmed` и не `auto_merged`. `POST /merge-candidates/{id}/reject` → 200, создаёт запрет склейки. `POST /guests/{id}/merge` с телом `{ "target_guest_id": "g-2002" }` → 200; 400 (сам с собой), 404, 409 `merge_blocked`.

## 8. Аналитика и изоляция R-3 (UC-7, UC-8)

`GET /tenants/{id}/aggregates?from=2026-01-01&to=2026-10-31&granularity=month` — R-3 только для `id` из claims, R-2/R-1/R-4 любой. Для чужого `id` у R-3 — 403 `forbidden_tenant`. В ответе нет `guest_id`; ячейки меньше k = 5 скрыты.

```json
{
  "tenant_id": "t-cine",
  "series": [
    { "period": "2026-09", "new_guests_via_mall": 14, "hidden": false },
    { "period": "2026-10", "new_guests_via_mall": null, "hidden": true }
  ],
  "min_cell": 5
}
```

`GET /dashboard/metrics?from=&to=` — R-2, R-1, R-4:

```json
{
  "period": { "from": "2026-09-01", "to": "2026-09-30" },
  "match_rate": 0.82,
  "contested_merge_share": 0.004,
  "segment_to_export_minutes_p50": 3.2,
  "conversion": 0.041,
  "incremental_visits": {
    "recipients_rate": 0.041,
    "control_rate": 0.026,
    "lift": 0.015,
    "significant": true
  }
}
```

Источники каждого показателя и формулы — `docs/15-consistency.md`, раздел 8. Поле может быть `null`, если данных нет (пустое состояние).

## Чекпойнт

Описаны: единый формат ошибок, 202/200 импорта, профиль и рекомендации, сегменты, выгрузка (201/202/409), склейки, агрегаты R-3, метрики. Новых `Q`/`D` не добавлено.
