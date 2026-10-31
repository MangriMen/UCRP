# 06. Object Diagrams (экземпляры данных)

Ограничение нотации. В Mermaid нет отдельной object diagram, поэтому экземпляры показаны через `classDiagram`: каждый «класс» — один экземпляр (строка таблицы), заголовок вида `таблица ид`, атрибуты — значения полей, стрелки — внешние ключи. Скобки в значениях не используются (в `classDiagram` они означают методы). Деньги в минимальных единицах (35000 = 350,00 RUB), даты в UTC. Общий набор арендаторов для всех сценариев: «Кофейня Bean» (категория Кафе, 1 этаж), «ФитЗона» (Спорт, 2 этаж), «Кинотеатр Синема» (Развлечения, 3 этаж).

## Сценарий (a): один гость, события из кассы и CRM склеены по телефону

```mermaid
classDiagram
  direction LR
  class ds_pos["data_source ds-pos"] {
    tenant: Кофейня Bean
    type: pos
  }
  class ds_crm["data_source ds-crm"] {
    tenant: ФитЗона
    type: crm
  }
  class b1["import_batch b-101"] {
    mode: file
    file_name: kassa_2026-09.csv
    status: completed
    accepted_rows: 2
  }
  class b2["import_batch b-102"] {
    mode: pull
    status: completed
    accepted_rows: 1
  }
  class g1["guest g-1001"] {
    display_name: Анна К.
    status: active
  }
  class i1["guest_identifier i-1"] {
    id_type: phone
    value_normalized: 79001234567
    data_source: ds-pos
  }
  class i2["guest_identifier i-2"] {
    id_type: email
    value_normalized: anna@example.com
    data_source: ds-crm
  }
  class e1["event e-1"] {
    type: purchase
    external_id: CHK-77120
    occurred_at: 2026-09-12 12:40
    amount_minor: 35000
  }
  class e2["event e-2"] {
    type: booking
    external_id: BK-5531
    occurred_at: 2026-09-14 19:00
    amount_minor: 240000
  }
  class e3["event e-3"] {
    type: purchase
    external_id: CHK-77310
    occurred_at: 2026-09-20 09:15
    amount_minor: 28000
  }
  ds_pos --> b1 : import_batch.data_source_id
  ds_crm --> b2 : import_batch.data_source_id
  b1 --> e1 : event.import_batch_id
  b1 --> e3 : event.import_batch_id
  b2 --> e2 : event.import_batch_id
  g1 --> e1 : event.guest_id
  g1 --> e2 : event.guest_id
  g1 --> e3 : event.guest_id
  g1 --> i1 : guest_identifier.guest_id
  g1 --> i2 : guest_identifier.guest_id
```

Пояснение. Первая строка кассы (e-1) принесла телефон 79001234567: резолвер не нашёл такого ключа и создал гостя g-1001 с идентификатором i-1. Событие e-2 из CRM фитнеса содержит тот же телефон в другом написании (`+7 (900) 123-45-67`), после нормализации ключ найден, событие привязано к g-1001, а новый ключ — email — добавлен как i-2; автосклейка здесь не потребовалась, это поиск по ключу. В профиле видны три события у двух арендаторов, число посещённых арендаторов (`tenant_count`) равно 2. Повторная загрузка `kassa_2026-09.csv` вернёт 200 `duplicate=true` по sha256 файла, а любая другая загрузка с `CHK-77120` даст строку `duplicate` благодаря `UNIQUE(data_source_id, event_type_id, external_id)`. Match rate по этим трём событиям равен 100%.

| Элемент                 | Go-пакет / слой                                  | React-модуль        | API-контракт              |
| ----------------------- | ------------------------------------------------ | ------------------- | ------------------------- |
| import_batch, event     | `internal/ingestion`, `repository.EventRepo`     | `features/import`   | `GET /imports/{id}`       |
| guest, guest_identifier | `internal/identity` (`IdentityResolver.Resolve`) | `features/profiles` | `GET /guests/{id}`        |
| события в профиле       | `service.ProfileService`                         | `/guests/:guestId`  | `GET /guests/{id}/events` |

---

## Сценарий (b): спорная склейка в очереди и ручное разъединение

### (b1) Состояние: пара в очереди

```mermaid
classDiagram
  direction LR
  class ds_pos["data_source ds-pos"] {
    type: pos
  }
  class ds_crm["data_source ds-crm"] {
    type: crm
  }
  class g2001["guest g-2001"] {
    display_name: Ольга Р.
    status: active
  }
  class g2002["guest g-2002"] {
    display_name: Игорь Р.
    status: active
  }
  class i1["guest_identifier i-11"] {
    id_type: phone
    value_normalized: 79005550101
  }
  class i2["guest_identifier i-12"] {
    id_type: phone
    value_normalized: 79005550101
  }
  class ev1["event e-21"] {
    type: purchase
    external_id: CHK-90001
    guest_id: g-2001
  }
  class ev2["event e-22"] {
    type: booking
    external_id: BK-7001
    guest_id: g-2002
  }
  class mc1["merge_candidate mc-1"] {
    guest_a_id: g-2001
    guest_b_id: g-2002
    reason_code: shared_phone_names_differ
    reason_text: Общий телефон, имена различаются
    score: 55
    resolution: pending
  }
  ds_pos --> i1 : data_source_id
  ds_crm --> i2 : data_source_id
  g2001 --> i1 : guest_id
  g2002 --> i2 : guest_id
  g2001 --> ev1 : event.guest_id
  g2002 --> ev2 : event.guest_id
  mc1 --> g2001 : guest_a_id
  mc1 --> g2002 : guest_b_id
```

Пояснение. Два профиля созданы из разных источников и делят один телефон, но имена различаются (например, один номер на семью). Резолвер не имеет права сливать такую пару автоматически, поэтому создал `merge_candidate` со статусом `pending`, причиной и оценкой 55; пара видна в очереди `/identity/queue` и не меняет профили. События и идентификаторы пока остаются у своих гостей, так что рассылка по телефону не придёт «чужому» человеку с чужой историей. Индекс очереди — `ix_merge_resolution (resolution, created_at)`.

### (b2) Состояние: объединили, затем разъединили вручную

```mermaid
classDiagram
  direction LR
  class u4["app_user u-4"] {
    role: analyst
  }
  class g2001["guest g-2001"] {
    status: active
  }
  class g2002["guest g-2002"] {
    status: active
    merged_into_id: null
  }
  class ev1["event e-21"] {
    guest_id: g-2001
    origin_guest_id: g-2001
  }
  class ev2["event e-22"] {
    guest_id: g-2002
    origin_guest_id: g-2002
  }
  class mc1["merge_candidate mc-1"] {
    resolution: reverted
    winner_guest_id: g-2001
    merged_guest_id: g-2002
  }
  class ub1["unmerge_block ub-1"] {
    guest_a_id: g-2001
    guest_b_id: g-2002
    reason: manual_unmerge
  }
  class a1["audit_log a-1"] {
    action: merge.confirm
    entity_id: mc-1
  }
  class a2["audit_log a-2"] {
    action: merge.revert
    entity_id: mc-1
  }
  u4 --> mc1 : decided_by
  u4 --> ub1 : created_by
  mc1 --> ub1 : candidate_id
  u4 --> a1 : actor_user_id
  u4 --> a2 : actor_user_id
  g2001 --> ev1 : guest_id
  g2002 --> ev2 : guest_id
```

Пояснение. Хронология: R-4 подтвердил пару (`POST /merge-candidates/mc-1/confirm`) — g-2002 стал `merged` с `merged_into_id = g-2001`, а e-22 и i-12 перешли к g-2001 (`origin_guest_id` остался g-2002); в `audit_log` появилась запись `merge.confirm` (a-1). Затем аналитик понял, что это разные люди, и вызвал `POST /merge-candidates/mc-1/revert`: строки с `origin_guest_id = g-2002` вернулись к g-2002, профиль снова `active`, кандидат получил `resolution = reverted`, создан `unmerge_block ub-1` и запись `merge.revert` (a-2) — всё в одной транзакции. На диаграмме показано конечное состояние. Новая загрузка с тем же общим телефоном пару не склеит и в очередь не вернёт, пока запрет не снят `DELETE /unmerge-blocks/ub-1`.

| Элемент         | Go-пакет / слой                                                        | React-модуль                            | API-контракт                                                              |
| --------------- | ---------------------------------------------------------------------- | --------------------------------------- | ------------------------------------------------------------------------- |
| merge_candidate | `internal/identity` (`MergeExecutor`), `repository.MergeCandidateRepo` | `features/profiles` (`/identity/queue`) | `GET /merge-candidates`, `POST /merge-candidates/{id}/confirm`, `/revert` |
| unmerge_block   | `repository.UnmergeBlockRepo`                                          | то же                                   | `GET /unmerge-blocks`, `DELETE /unmerge-blocks/{id}`                      |
| audit_log       | `service.AuditService`                                                 | `features/admin`                        | `GET /audit-logs?entity_id=mc-1`                                          |

---

## Сценарий (c): сегмент «Спящий сосед», рекомендация и выгрузка с контрольной группой

### (c1) Сегмент и выгрузка

Сегмент: «был в Кофейне Bean и ФитЗоне за 90 дней, но не был в Кинотеатре Синема, есть телефон для SMS». Лимит: не более 2 SMS за 7 дней. Выгрузка: канал `sms`, контрольная группа 33%.

```mermaid
classDiagram
  direction LR
  class s1["segment s-1"] {
    name: Спящий сосед, кино
    template_code: sleeping_neighbor
  }
  class sv1["segment_version sv-1"] {
    version_no: 1
    rule: was Кофейня и ФитЗона, was_not Кино, 90 дней, phone
    member_count: 5
    status: ready
  }
  class camp["campaign c-1"] {
    name: Сеансы в кино
    target_tenant: Кинотеатр Синема
    attribution_days: 14
  }
  class lim["contact_limit cl-1"] {
    channel: sms
    max_contacts: 2
    period_days: 7
  }
  class x1["export_job x-1"] {
    channel: sms
    format: csv
    control_percent: 33
    total_in_segment: 5
    no_contact_count: 1
    blocked_by_limit_count: 1
    control_count: 1
    exported_count: 2
  }
  class g1["guest g-1001 Анна"] {
    phone: есть
  }
  class g3["guest g-1003 Мария"] {
    phone: есть
  }
  class g6["guest g-1006 Павел"] {
    phone: есть
  }
  class g4["guest g-1004 Сергей"] {
    phone: нет, только email
  }
  class g5["guest g-1005 Елена"] {
    phone: есть
    sms_за_7_дней: 2
  }
  class em1["export_member em-1"] {
    guest_id: g-1001
    is_control: false
  }
  class em2["export_member em-2"] {
    guest_id: g-1003
    is_control: true
  }
  class em3["export_member em-3"] {
    guest_id: g-1006
    is_control: false
  }
  s1 --> sv1 : segment_version.segment_id
  camp --> s1 : segment_id
  x1 --> sv1 : segment_version_id
  x1 --> camp : campaign_id
  sv1 --> g1 : segment_member
  sv1 --> g3 : segment_member
  sv1 --> g6 : segment_member
  sv1 --> g4 : segment_member
  sv1 --> g5 : segment_member
  x1 --> em1 : export_id
  x1 --> em2 : export_id
  x1 --> em3 : export_id
```

Пояснение. В версии 1 сегмента пять гостей. При создании выгрузки сервис сначала отсекает гостей без телефона (g-1004, `no_contact_count = 1`), затем применяет лимит `cl-1`: у g-1005 уже два SMS за 7 дней, поэтому она исключена (`blocked_by_limit_count = 1`). Из трёх допустимых один гость (33%) детерминированно по `control_seed` отнесён в контрольную группу (g-1003, `is_control = true`); остальные двое (g-1001, g-1006) попадут в CSV. `contact_log` создаётся только для получателей, поэтому у g-1003 контакта нет и его визиты в Кинотеатр за 14 дней станут базой для расчёта инкремента. Лимит учтён до фиксации счётчиков, `exported_count + control_count + no_contact_count + blocked_by_limit_count = total_in_segment` (2 + 1 + 1 + 1 = 5).

### (c2) Рекомендация гостя g-1001 и контакт

```mermaid
classDiagram
  direction LR
  class g1["guest g-1001 Анна"] {
    visited: Кофейня Bean, ФитЗона
  }
  class r1["recommendation r-1"] {
    rank: 1
    score: 0.7800
    algorithm_version: v1
    explanation: Похожие гости ходят в Кинотеатр Синема, Спорт и Развлечения смежны, недавняя активность, 3 этаж
  }
  class f1["recommendation_factor similar_guests"] {
    raw_value: 0.85
    weight: 0.40
    contribution: 0.3400
  }
  class f2["recommendation_factor category_adjacency"] {
    raw_value: 0.70
    weight: 0.25
    contribution: 0.1750
  }
  class f3["recommendation_factor recency"] {
    raw_value: 0.90
    weight: 0.25
    contribution: 0.2250
  }
  class f4["recommendation_factor distance"] {
    raw_value: 0.40
    weight: 0.10
    contribution: 0.0400
  }
  class tcine["tenant Кинотеатр Синема"] {
    category: Развлечения
    floor: 3
  }
  class tcafe["tenant Кофейня Bean"] {
    category: Кафе
  }
  class cl["contact_log ctl-1"] {
    export_id: x-1
    contacted_at: 2026-10-30 10:15
  }
  g1 --> r1 : recommendation.guest_id
  r1 --> tcine : recommendation.tenant_id
  r1 --> f1 : factor
  r1 --> f2 : factor
  r1 --> f3 : factor
  r1 --> f4 : factor
  g1 --> tcafe : был, не рекомендуется
  g1 --> cl : contact_log.guest_id
```

Пояснение. Для g-1001 кандидатами считаются только арендаторы, где она ещё не была, поэтому Кофейня Bean и ФитЗона в списке отсутствуют. Первая рекомендация — Кинотеатр Синема: оценка 0,78 складывается из вкладов факторов (похожие гости 0,34, смежность категорий 0,175, свежесть 0,225, расстояние и этаж 0,04; приоритет арендатора нейтральный и вклада не даёт); текст обоснования хранится снимком в `explanation`. Прямые конкуренты не исключаются, а понижаются фактором `competitor_penalty` — в примере такого кандидата в топ-3 не оказалось из-за низкой суммы, а не из-за фильтра (A-18, D-34). После выгрузки x-1 создана запись `contact_log` для g-1001, которая учитывается лимитом контактов и конверсией (визит в кино до 2026-11-13 — окно 14 дней по умолчанию, `campaign.attribution_days`). Диаграмма иллюстрирует инвариант: `recommendation.tenant_id` ∉ множества арендаторов событий гостя.

| Элемент                                               | Go-пакет / слой                                  | React-модуль               | API-контракт                                                                               |
| ----------------------------------------------------- | ------------------------------------------------ | -------------------------- | ------------------------------------------------------------------------------------------ |
| segment, segment_version, segment_member              | `internal/segments` (`SegmentEvaluator`)         | `features/segments`        | `POST /segments`, `GET /segments/{id}/members`                                             |
| contact_limit, export_job, export_member, contact_log | `internal/export`, `service.ContactLimitService` | `features/export`          | `POST /exports/preview`, `POST /exports` (`Idempotency-Key`), `GET /exports/{id}/download` |
| campaign                                              | `service.CampaignService`                        | `features/export`          | `GET /campaigns/{id}/results`                                                              |
| recommendation, recommendation_factor                 | `internal/recommend` (`Recommender`, `Scorer`)   | `features/recommendations` | `GET /guests/{id}/recommendations`                                                         |

## Чекпойнт

Создано: 3 сценария (a, b1–b2, c1–c2) с экземплярами E-1, E-2, E-3, E-4, E-5, E-8, E-10, E-11, E-12, E-13, E-15, E-16, E-17, E-18, E-19, E-20, E-21. Использованы эндпоинты из реестра; инварианты: нет посещённых мест в рекомендациях, только гости с телефоном в SMS-выгрузке, сумма счётчиков выгрузки равна размеру сегмента.
