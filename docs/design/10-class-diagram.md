# 10. Class Diagrams

Диаграммы показывают контракты (Go-интерфейсы, `<<interface>>`) и их реализации, а не код. Правило слоёв: `handler → service → repository`, зависимости направлены внутрь к `domain`; интерфейсы репозиториев и портов объявляются в `internal/domain` (потребитель владеет контрактом), реализации — в `internal/repository`, `internal/ingestion`, `internal/export` и др. Во всех методах первым параметром идёт `ctx context.Context` (на диаграммах опущен для краткости, где не меняет смысл). Имена совпадают с глоссарием (02) и сигнатурами из 04. Mermaid: в `classDiagram` нет пакетов и приватности Go, поэтому слой указан в пояснении и в таблице. Из-за лимита ~15 узлов модель разбита на шесть диаграмм: CL-1 слои, CL-2 ingestion, CL-3 identity, CL-4 segments и recommend, CL-5 export, CL-6 domain и DTO.

## CL-1. Слои на примере импорта и профиля

```mermaid
classDiagram
  class ImportHandler {
    +Upload(w, r)
    +GetBatch(w, r)
    +ListRejected(w, r)
  }
  class ImportService {
    +Accept(ctx, cmd AcceptImport) ImportReceipt
    +Status(ctx, id) ImportBatchDTO
    +Process(ctx, batchID) error
  }
  class ImportBatchRepo {
    <<interface>>
    +Insert(ctx, b ImportBatch) error
    +FindByHash(ctx, srcID, hash) ImportBatch
    +UpdateCounters(ctx, id, c Counters) error
  }
  class EventRepo {
    <<interface>>
    +InsertIgnoreDuplicates(ctx, evs) Counters
    +ListByGuest(ctx, guestID, page) Events
  }
  class Queue {
    <<interface>>
    +Enqueue(ctx, job Job) error
    +Take(ctx, types, worker) Job
    +Complete(ctx, id) error
  }
  class SQLImportBatchRepo
  class SQLEventRepo
  class DBQueue
  ImportHandler --> ImportService : вызывает
  ImportService --> ImportBatchRepo
  ImportService --> EventRepo
  ImportService --> Queue
  SQLImportBatchRepo ..|> ImportBatchRepo
  SQLEventRepo ..|> EventRepo
  DBQueue ..|> Queue
```

Пояснение. `Handler` только разбирает HTTP, проверяет роль и маппит DTO; бизнес-правила (дубликат, формат, ограничения) живут в `Service`; `Repository` скрывает SQL и диалект (SQLite/PostgreSQL), поэтому смена СУБД не затрагивает верхние слои. `Queue` — порт с реализацией на таблице `job_queue` (D-8) и возможной заменой на брокер. Тот же шаблон применяется ко всем остальным сервисам (`ProfileService`, `SegmentService`, `ExportService`, …).

## CL-2. Ingestion: SourceConnector и его реализации

```mermaid
classDiagram
  class SourceConnector {
    <<interface>>
    +Type() string
    +Open(ctx, src DataSource, in Input) RowReader
    +Pull(ctx, src DataSource, w Window) RowReader
  }
  class RowReader {
    <<interface>>
    +Next(ctx) Row
    +Close() error
  }
  class CsvExcelConnector
  class CrmConnector
  class PosConnector
  class ApiConnector
  class ConnectorRegistry {
    +Register(c SourceConnector)
    +For(typeCode string) SourceConnector
  }
  class RowValidator {
    +Validate(row Row, m FieldMapping) ValidRow
  }
  class Normalizer {
    +Phone(s) string
    +Email(s) string
    +Card(s) string
    +Date(s, layouts) Time
  }
  class IngestionPipeline {
    +Run(ctx, batch ImportBatch) Counters
  }
  CsvExcelConnector ..|> SourceConnector
  CrmConnector ..|> SourceConnector
  PosConnector ..|> SourceConnector
  ApiConnector ..|> SourceConnector
  ConnectorRegistry o-- SourceConnector
  SourceConnector ..> RowReader
  IngestionPipeline --> ConnectorRegistry
  IngestionPipeline --> RowValidator
  RowValidator --> Normalizer
```

Пояснение. Каждый формат источника — одна реализация `SourceConnector`, зарегистрированная в `ConnectorRegistry` по коду типа из справочника `data_source_type`; `IngestionPipeline` не знает о форматах и работает с потоком `Row`. `Open` обслуживает загруженный файл, `Pull` — API-источник с окном времени. Добавление нового формата (например, 1С) сводится к новой реализации и записи в справочнике, а валидатор, нормализатор и конвейер не меняются (Q-7). Нормализатор общий с `identity`, чтобы ключи сравнивались одинаково.

## CL-3. Identity: IdentityResolver и склейка

```mermaid
classDiagram
  class IdentityResolver {
    <<interface>>
    +Resolve(ctx, chunk EventChunk) ResolveResult
    +Rematch(ctx, identifierID) ResolveResult
  }
  class Policy {
    +ReliableKeys() KeyTypes
    +IsConflict(a Guest, b Guest) bool
    +Score(a Guest, b Guest) int
  }
  class MergeExecutor {
    +Merge(ctx, winner, loser, how) MergeResult
    +Revert(ctx, candidateID, actor) error
  }
  class DefaultResolver
  class IdentityService {
    +Confirm(ctx, id, actor) error
    +Reject(ctx, id, actor) error
    +Revert(ctx, id, actor) error
    +ManualMerge(ctx, guestID, targetID, actor) error
  }
  class MergeCandidateRepo {
    <<interface>>
  }
  class MergeBlockRepo {
    <<interface>>
    +Exists(ctx, a, b) bool
  }
  class GuestIdentifierRepo {
    <<interface>>
    +FindByKeys(ctx, keys) Identifiers
  }
  DefaultResolver ..|> IdentityResolver
  DefaultResolver --> Policy
  DefaultResolver --> MergeExecutor
  DefaultResolver --> GuestIdentifierRepo
  DefaultResolver --> MergeBlockRepo
  IdentityService --> MergeExecutor
  IdentityService --> MergeCandidateRepo
  MergeExecutor --> MergeCandidateRepo
```

Пояснение. `IdentityResolver` вызывается воркером для чанка событий; `Policy` инкапсулирует правила надёжных ключей и порог конфликта (Q-12), поэтому смена порога не затрагивает схему. `MergeExecutor` — единственный код, который переписывает `guest_id` и выставляет `origin_guest_id`; им пользуются и автосклейка, и ручные операции `IdentityService`, что гарантирует одинаковую обратимость (D-5, D-25). `MergeBlockRepo.Exists` проверяется перед любой склейкой. Все изменения идут в одной транзакции вместе с записью аудита.

## CL-4. Segments и Recommend

```mermaid
classDiagram
  class SegmentEvaluator {
    <<interface>>
    +Preview(ctx, rule Rule) PreviewResult
    +Evaluate(ctx, v SegmentVersion) MemberSet
    +EvaluateGuest(ctx, v SegmentVersion, guestID) bool
  }
  class RuleCompiler {
    +Compile(rule Rule) Query
    +Validate(rule Rule) Errors
  }
  class SqlEvaluator
  class SegmentService {
    +Preview(ctx, rule) PreviewResult
    +Create(ctx, cmd) Segment
    +NewVersion(ctx, id, rule) SegmentVersion
  }
  class Recommender {
    <<interface>>
    +Refresh(ctx, guestID) Recommendations
    +ForSegment(ctx, segmentID) TenantRanking
    +Fallback(ctx, guestID) Recommendations
  }
  class Scorer {
    <<interface>>
    +Score(ctx, g GuestFeed, t Tenant) ScoreBreakdown
  }
  class WeightedScorer
  class Factor {
    <<interface>>
    +Code() string
    +Value(g GuestFeed, t Tenant) float
  }
  SqlEvaluator ..|> SegmentEvaluator
  SqlEvaluator --> RuleCompiler
  SegmentService --> SegmentEvaluator
  WeightedScorer ..|> Scorer
  WeightedScorer o-- Factor
  RecommendationService --> Recommender
  Recommender --> Scorer
```

Пояснение. `SegmentEvaluator` один контракт для трёх режимов: предпросмотр без записи, полный пересчёт версии и проверка одного гостя при инкрементальном обновлении (UC-6). `RuleCompiler` переводит JSON-правило в параметризованный SQL, чтобы правило не превращалось в произвольный запрос. `Recommender` использует `Scorer`, составленный из независимых `Factor` (похожие гости, смежность категорий, этаж/расстояние, свежесть, приоритет арендатора, понижение конкурента), веса берутся из конфигурации (D-15); новый фактор — новая реализация. Уже посещённые места отфильтровываются до скоринга; прямые конкуренты не отсекаются, а понижаются фактором `competitor_penalty` (A-18, D-34). Для гостя без событий `Recommender` вызывает `Fallback` (A-21, D-37).

## CL-5. Export и ограничение контактов

```mermaid
classDiagram
  class ExportChannel {
    <<interface>>
    +Code() string
    +ContactType() string
    +Render(ctx, m ExportMembers, f Format) Output
  }
  class SmsChannel
  class PushChannel
  class EmailChannel
  class ExportFormat {
    <<interface>>
    +Write(ctx, rows ContactRows) Output
  }
  class CsvFormat
  class ApiFormat
  class ExportService {
    +Preview(ctx, cmd) ExportCounts
    +Create(ctx, cmd, idemKey) Export
    +Download(ctx, id) Output
  }
  class ContactLimitService {
    +Apply(ctx, guests, channel) LimitResult
  }
  class ControlGroupSelector {
    +Pick(guests, percent, seed) Split
  }
  SmsChannel ..|> ExportChannel
  PushChannel ..|> ExportChannel
  EmailChannel ..|> ExportChannel
  CsvFormat ..|> ExportFormat
  ApiFormat ..|> ExportFormat
  ExportService --> ExportChannel
  ExportService --> ContactLimitService
  ExportService --> ControlGroupSelector
  ExportChannel --> ExportFormat
```

Пояснение. Канал определяет, какой тип идентификатора считается контактом (`ContactType`: телефон для SMS и push, email для email, Q-21), формат — как отдаётся результат (CSV-файл или JSON API для провайдера, Q-20). `ExportService` оркестрирует: фильтр по контакту, `ContactLimitService` (самое строгое правило), `ControlGroupSelector` (детерминированно по `control_seed`, D-17) и транзакционную запись. Новый канал или формат — новая реализация без изменения сервиса; внешние провайдеры не вызываются, система только отдаёт данные.

## CL-6. Domain и DTO

```mermaid
classDiagram
  class Guest {
    +ID
    +DisplayName
    +Status
    +MergedIntoID
  }
  class GuestIdentifier {
    +IDType
    +ValueNormalized
    +OriginGuestID
  }
  class Event {
    +ExternalID
    +OccurredAt
    +AmountMinor
    +RawPayload
  }
  class Segment {
    +CurrentVersionNo
    +Status
  }
  class MergeCandidate {
    +Resolution
    +Score
  }
  class Export {
    +Channel
    +ControlSeed
    +Status
  }
  class Principal {
    +UserID
    +Role
    +TenantID
  }
  class GuestProfileDTO {
    +Guest
    +TenantsVisited
    +Events
    +Recommendations
  }
  Guest "1" o-- "many" GuestIdentifier
  Guest "1" o-- "many" Event
  Segment "1" o-- "many" Export
  Guest "1" <-- "many" MergeCandidate
  GuestProfileDTO ..> Guest
  GuestProfileDTO ..> Event
```

Пояснение. `domain` не зависит от HTTP и SQL: это структуры E-1…E-21 (показаны ключевые) и значения вроде `Principal`, `Counters`, `Window`. DTO (`GuestProfileDTO`, `ImportBatchDTO`, `TenantAggregates`) живут в `internal/handler/dto` и собираются сервисами, чтобы JSON-контракт не диктовал форму таблиц. Для R-3 существует отдельный DTO `TenantAggregates` без идентификаторов гостей — типовая гарантия, что строки гостей не могут попасть в ответ арендатору. `Principal` передаётся только через `context.Context`.
