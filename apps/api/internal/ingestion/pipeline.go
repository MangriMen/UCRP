package ingestion

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/MangriMen/UCRP/internal/domain"
	"github.com/MangriMen/UCRP/internal/identity/normalize"
)

type Counters struct {
	Total     int
	Accepted  int
	Rejected  int
	Duplicate int
}

type Pipeline struct {
	batches  domain.ImportBatchRepo
	rejected domain.RejectedRowRepo
	events   domain.EventRepo
	sources  domain.DataSourceRepo
	catalog  domain.ImportCatalog
	clock    domain.Clock
	idGen    domain.IDGen
}

func NewPipeline(
	batches domain.ImportBatchRepo,
	rejected domain.RejectedRowRepo,
	events domain.EventRepo,
	sources domain.DataSourceRepo,
	catalog domain.ImportCatalog,
	clock domain.Clock,
	idGen domain.IDGen,
) *Pipeline {
	return &Pipeline{
		batches:  batches,
		rejected: rejected,
		events:   events,
		sources:  sources,
		catalog:  catalog,
		clock:    clock,
		idGen:    idGen,
	}
}

func (p *Pipeline) ProcessRows(ctx context.Context, batch *domain.ImportBatch, rows []Row, mapping FieldMapping) (Counters, error) {
	if err := p.batches.MarkRunning(ctx, batch.ID, p.clock.Now()); err != nil {
		return Counters{}, err
	}

	var (
		counters Counters
		rejects  []domain.RejectedRow
	)
	lookups := newLookups()

	for i, row := range rows {
		rowNumber := i + 2
		counters.Total++
		rawLine := rawLineOf(row)
		parsed, rowErr := p.parseRow(ctx, row, mapping, lookups, batch)
		if rowErr != nil {
			rejects = append(rejects, domain.RejectedRow{
				ID:            p.idGen.NewID(ctx),
				ImportBatchID: batch.ID,
				RowNumber:     rowNumber,
				ErrorCode:     rowErr.code,
				ErrorMessage:  rowErr.msg,
				RawLine:       rawLine,
				CreatedAt:     p.clock.Now(),
			})
			counters.Rejected++
			if len(rejects) >= 200 {
				if err := p.rejected.InsertBatch(ctx, rejects); err != nil {
					return counters, err
				}
				rejects = rejects[:0]
			}
			continue
		}
		parsed.ID = p.idGen.NewID(ctx)
		parsed.ImportBatchID = batch.ID
		parsed.CreatedAt = p.clock.Now()
		inserted, err := p.events.InsertIgnoreDuplicate(ctx, parsed)
		if err != nil {
			return counters, err
		}
		if inserted {
			counters.Accepted++
		} else {
			counters.Duplicate++
		}
	}
	if err := p.rejected.InsertBatch(ctx, rejects); err != nil {
		return counters, err
	}

	status := domain.ImportCompleted
	if counters.Rejected > 0 {
		status = domain.ImportCompletedWithErrors
	}
	if err := p.batches.MarkFinished(ctx, batch.ID, status, counters.Total, counters.Accepted, counters.Rejected, counters.Duplicate, "", p.clock.Now()); err != nil {
		return counters, err
	}
	if err := p.sources.TouchLastImport(ctx, batch.DataSourceID, p.clock.Now()); err != nil {
		return counters, err
	}
	return counters, nil
}

func (p *Pipeline) Fail(ctx context.Context, batchID, msg string) error {
	return p.batches.MarkFinished(ctx, batchID, domain.ImportFailed, 0, 0, 0, 0, msg, p.clock.Now())
}

type rowError struct {
	code string
	msg  string
}

func (p *Pipeline) parseRow(ctx context.Context, row Row, mapping FieldMapping, lookups *lookups, batch *domain.ImportBatch) (*domain.Event, *rowError) {
	externalID := strings.TrimSpace(get(row, mapping.ExternalID))
	if externalID == "" {
		return nil, &rowError{"no_external_id", "Нет номера операции"}
	}
	if len(externalID) > 128 {
		return nil, &rowError{"bad_external_id", "Слишком длинный номер операции"}
	}

	occurredAt, err := parseDate(get(row, mapping.OccurredAt))
	if err != nil {
		return nil, &rowError{"invalid_date", "Не распознана дата"}
	}

	eventTypeCode := strings.ToLower(strings.TrimSpace(get(row, mapping.EventType)))
	eventTypeID, ok, err := lookups.eventType(ctx, p.catalog, eventTypeCode)
	if err != nil {
		return nil, &rowError{"internal", "Ошибка справочника типов событий"}
	}
	if !ok {
		return nil, &rowError{"bad_event_type", "Неизвестный тип события: " + eventTypeCode}
	}

	tenantID, re := p.resolveTenant(ctx, row, mapping, lookups, batch)
	if re != nil {
		return nil, re
	}

	projectID := ""
	if mapping.ProjectCode != "" {
		if code := strings.TrimSpace(get(row, mapping.ProjectCode)); code != "" {
			pid, ok, err := lookups.project(ctx, p.catalog, tenantID, code)
			if err != nil {
				return nil, &rowError{"internal", "Ошибка справочника проектов"}
			}
			if ok {
				projectID = pid
			}
		}
	}

	amountMinor, err := parseAmount(get(row, mapping.Amount))
	if err != nil {
		return nil, &rowError{"invalid_amount", "Не распознана сумма"}
	}
	currency := strings.ToUpper(strings.TrimSpace(get(row, mapping.Currency)))
	if currency == "" {
		currency = "RUB"
	}

	phone := normalize.Phone(get(row, mapping.Phone))
	email := ""
	if rawEmail := strings.TrimSpace(get(row, mapping.Email)); rawEmail != "" {
		email = normalize.Email(rawEmail)
		if !normalize.ValidEmail(email) {
			return nil, &rowError{"bad_email", "Не распознан email"}
		}
	}
	card := normalize.Card(get(row, mapping.Card))

	matchKeys := map[string]string{}
	if phone != "" {
		matchKeys["phone"] = phone
	}
	if email != "" {
		matchKeys["email"] = email
	}
	if card != "" {
		matchKeys["card"] = card
	}
	matchJSON, err := json.Marshal(matchKeys)
	if err != nil {
		return nil, &rowError{"internal", "Ошибка нормализации ключей"}
	}
	rawJSON, err := json.Marshal(map[string]string(row))
	if err != nil {
		return nil, &rowError{"internal", "Ошибка payload"}
	}

	return &domain.Event{
		TenantID:     tenantID,
		ProjectID:    projectID,
		EventTypeID:  eventTypeID,
		DataSourceID: batch.DataSourceID,
		ExternalID:   externalID,
		OccurredAt:   occurredAt,
		AmountMinor:  amountMinor,
		Currency:     currency,
		MatchKeys:    string(matchJSON),
		RawPayload:   string(rawJSON),
	}, nil
}

func (p *Pipeline) resolveTenant(ctx context.Context, row Row, mapping FieldMapping, lookups *lookups, batch *domain.ImportBatch) (string, *rowError) {
	if mapping.TenantCode != "" {
		if code := strings.TrimSpace(get(row, mapping.TenantCode)); code != "" {
			id, ok, err := lookups.tenant(ctx, p.catalog, code)
			if err != nil {
				return "", &rowError{"internal", "Ошибка справочника арендаторов"}
			}
			if !ok {
				return "", &rowError{"invalid_tenant", "Неизвестный арендатор: " + code}
			}
			return id, nil
		}
	}
	ds, err := lookups.dataSource(ctx, p.sources, batch.DataSourceID)
	if err != nil {
		return "", &rowError{"internal", "Ошибка источника данных"}
	}
	if ds == nil || ds.TenantID == "" {
		return "", &rowError{"no_tenant", "Не задан арендатор события"}
	}
	return ds.TenantID, nil
}

type lookups struct {
	eventTypes map[string]string
	tenants    map[string]string
	projects   map[string]string
	sources    map[string]*domain.DataSource
}

func newLookups() *lookups {
	return &lookups{
		eventTypes: map[string]string{},
		tenants:    map[string]string{},
		projects:   map[string]string{},
		sources:    map[string]*domain.DataSource{},
	}
}

func (l *lookups) eventType(ctx context.Context, catalog domain.ImportCatalog, code string) (string, bool, error) {
	if id, seen := l.eventTypes[code]; seen {
		return id, id != "", nil
	}
	id, ok, err := catalog.EventTypeID(ctx, code)
	if err != nil {
		return "", false, err
	}
	if !ok {
		id = ""
	}
	l.eventTypes[code] = id
	return id, ok, nil
}

func (l *lookups) tenant(ctx context.Context, catalog domain.ImportCatalog, code string) (string, bool, error) {
	if id, seen := l.tenants[code]; seen {
		return id, id != "", nil
	}
	id, ok, err := catalog.TenantIDByCode(ctx, code)
	if err != nil {
		return "", false, err
	}
	if !ok {
		id = ""
	}
	l.tenants[code] = id
	return id, ok, nil
}

func (l *lookups) project(ctx context.Context, catalog domain.ImportCatalog, tenantID, code string) (string, bool, error) {
	key := tenantID + ":" + code
	if id, seen := l.projects[key]; seen {
		return id, id != "", nil
	}
	id, ok, err := catalog.ProjectID(ctx, tenantID, code)
	if err != nil {
		return "", false, err
	}
	if !ok {
		id = ""
	}
	l.projects[key] = id
	return id, ok, nil
}

func (l *lookups) dataSource(ctx context.Context, repo domain.DataSourceRepo, id string) (*domain.DataSource, error) {
	if ds, seen := l.sources[id]; seen {
		return ds, nil
	}
	ds, err := repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	l.sources[id] = ds
	return ds, nil
}

func get(row Row, key string) string {
	if key == "" {
		return ""
	}
	if v, ok := row[strings.ToLower(strings.TrimSpace(key))]; ok {
		return v
	}
	return ""
}

func rawLineOf(row Row) string {
	b, err := json.Marshal(row)
	if err != nil {
		return ""
	}
	return string(b)
}

var dateLayouts = []string{
	time.RFC3339,
	"2006-01-02",
	"2006-01-02 15:04:05",
	"2006-01-02T15:04:05",
	"02.01.2006",
	"02.01.2006 15:04:05",
	"02/01/2006",
	"02/01/2006 15:04:05",
}

func parseDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, fmt.Errorf("empty date")
	}
	for _, layout := range dateLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("bad date")
}

func parseAmount(s string) (*int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, ",", ".")
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil, err
	}
	if f < 0 {
		return nil, fmt.Errorf("negative amount")
	}
	v := int64(math.Round(f * 100))
	return &v, nil
}
