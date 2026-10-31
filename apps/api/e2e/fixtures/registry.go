package fixtures

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/MangriMen/UCRP/internal/platform/auth"
	"github.com/google/uuid"
)

type Registry struct {
	db      *DB
	runID   string
	created map[string][]string
}

func NewRegistry(db *DB, runID string) *Registry {
	return &Registry{db: db, runID: runID, created: map[string][]string{}}
}

func (r *Registry) Name(suffix string) string {
	return fmt.Sprintf("e2e-%s-%s", r.runID, suffix)
}

func (r *Registry) Code(suffix string) string {
	prefix := fmt.Sprintf("e2e-%s-", r.runID)
	room := 32 - len(prefix)
	if room < 6 {
		return prefix[:32]
	}
	if len(suffix) <= room {
		return prefix + suffix
	}
	sum := sha256.Sum256([]byte(suffix))
	tail := hex.EncodeToString(sum[:])[:6]
	return prefix + suffix[:room-7] + "-" + tail
}

func (r *Registry) track(table, id string) {
	r.created[table] = append(r.created[table], id)
}

func (r *Registry) NewID() string {
	return uuid.NewString()
}

func (r *Registry) Phone(index int) string {
	sum := sha256.Sum256([]byte(r.runID))
	digits := binary.BigEndian.Uint32(sum[:4]) % 100000000
	return fmt.Sprintf("79%08d%05d", digits, index)
}

func (r *Registry) Cleanup(ctx context.Context) error {
	var firstErr error
	note := func(err error) {
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}

	for _, guestID := range r.created["guest"] {
		note(r.db.Exec(ctx, `DELETE FROM contact_log WHERE guest_id = $1`, guestID))
		note(r.db.Exec(ctx, `DELETE FROM export_member WHERE guest_id = $1`, guestID))
		note(r.db.Exec(ctx, `DELETE FROM segment_member WHERE guest_id = $1`, guestID))
		note(r.db.Exec(ctx, `DELETE FROM recommendation WHERE guest_id = $1`, guestID))
		note(r.db.Exec(ctx, `DELETE FROM merge_block WHERE guest_a_id = $1 OR guest_b_id = $1`, guestID))
		note(r.db.Exec(ctx, `DELETE FROM merge_candidate WHERE guest_a_id = $1 OR guest_b_id = $1`, guestID))
		note(r.db.Exec(ctx, `DELETE FROM event WHERE guest_id = $1 OR origin_guest_id = $1`, guestID))
		note(r.db.Exec(ctx, `DELETE FROM guest_identifier WHERE guest_id = $1 OR origin_guest_id = $1`, guestID))
		note(r.db.Exec(ctx, `UPDATE guest SET status = 'active', merged_into_id = NULL WHERE id = $1`, guestID))
	}

	for _, tenantID := range r.created["tenant"] {
		note(r.db.Exec(ctx, `DELETE FROM rejected_row rr USING import_batch ib, data_source ds WHERE rr.import_batch_id = ib.id AND ib.data_source_id = ds.id AND ds.tenant_id = $1`, tenantID))
		note(r.db.Exec(ctx, `DELETE FROM import_batch ib USING data_source ds WHERE ib.data_source_id = ds.id AND ds.tenant_id = $1`, tenantID))
		note(r.db.Exec(ctx, `DELETE FROM event e USING data_source ds WHERE e.data_source_id = ds.id AND ds.tenant_id = $1`, tenantID))
		note(r.db.Exec(ctx, `DELETE FROM data_source WHERE tenant_id = $1`, tenantID))
	}

	for _, segID := range r.created["segment"] {
		note(r.db.Exec(ctx, `DELETE FROM export_member em USING export_job ej WHERE em.export_id = ej.id AND ej.segment_id = $1`, segID))
		note(r.db.Exec(ctx, `DELETE FROM export_job WHERE segment_id = $1`, segID))
		note(r.db.Exec(ctx, `DELETE FROM segment_member sm USING segment_version sv WHERE sm.segment_version_id = sv.id AND sv.segment_id = $1`, segID))
		note(r.db.Exec(ctx, `DELETE FROM segment_version WHERE segment_id = $1`, segID))
	}

	order := []string{
		"contact_log",
		"export_member",
		"export_job",
		"campaign",
		"contact_limit",
		"segment_version",
		"segment",
		"recommendation",
		"merge_block",
		"merge_candidate",
		"event",
		"guest_identifier",
		"guest",
		"rejected_row",
		"import_batch",
		"data_source",
		"project",
		"app_user",
		"tenant",
		"mall",
		"tenant_category_relation",
		"tenant_category",
		"event_type",
		"data_source_type",
		"job_queue",
	}
	for _, table := range order {
		for _, id := range r.created[table] {
			note(r.db.Exec(ctx, fmt.Sprintf("DELETE FROM %s WHERE id = $1", table), id))
		}
	}
	r.created = map[string][]string{}
	return firstErr
}

type Category struct {
	ID   string
	Code string
	Name string
}

func (r *Registry) Category(ctx context.Context, suffix string) (*Category, error) {
	id := r.NewID()
	code := r.Code("cat-" + suffix)
	err := r.db.Exec(ctx, `
		INSERT INTO tenant_category (id, code, name, is_active, created_at)
		VALUES ($1, $2, $3, TRUE, $4)`, id, code, "E2E category "+suffix, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	r.track("tenant_category", id)
	return &Category{ID: id, Code: code, Name: "E2E category " + suffix}, nil
}

type Tenant struct {
	ID         string
	Code       string
	Name       string
	CategoryID string
	Priority   int
}

func (r *Registry) Tenant(ctx context.Context, suffix, categoryID string, priority int) (*Tenant, error) {
	id := r.NewID()
	code := r.Code("tenant-" + suffix)
	name := "E2E tenant " + suffix
	now := time.Now().UTC()
	err := r.db.Exec(ctx, `
		INSERT INTO tenant (id, code, name, mall_id, category_id, kind, floor, priority, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, (SELECT id FROM mall ORDER BY code LIMIT 1), $4, 'mall', 1, $5, TRUE, $6, $6)`, id, code, name, categoryID, priority, now)
	if err != nil {
		return nil, err
	}
	r.track("tenant", id)
	return &Tenant{ID: id, Code: code, Name: name, CategoryID: categoryID, Priority: priority}, nil
}

type EventType struct {
	ID               string
	Code             string
	CountsAsPresence bool
}

func (r *Registry) EventType(ctx context.Context, suffix string, countsAsPresence bool) (*EventType, error) {
	id := r.NewID()
	code := r.Code("evt-" + suffix)
	err := r.db.Exec(ctx, `
		INSERT INTO event_type (id, code, name, has_amount, counts_as_presence, is_active)
		VALUES ($1, $2, $3, TRUE, $4, TRUE)`, id, code, "E2E event "+suffix, countsAsPresence)
	if err != nil {
		return nil, err
	}
	r.track("event_type", id)
	return &EventType{ID: id, Code: code, CountsAsPresence: countsAsPresence}, nil
}

type DataSource struct {
	ID       string
	TenantID string
	Name     string
}

func (r *Registry) DataSource(ctx context.Context, suffix, tenantID string) (*DataSource, error) {
	typeID := r.NewID()
	typeCode := r.Code("dst-" + suffix)
	if err := r.db.Exec(ctx, `
		INSERT INTO data_source_type (id, code, name, connector_key, supports_file, supports_pull, is_active)
		VALUES ($1, $2, $3, 'csv', TRUE, FALSE, TRUE)`, typeID, typeCode, "E2E source type "+suffix); err != nil {
		return nil, err
	}
	r.track("data_source_type", typeID)

	id := r.NewID()
	name := r.Name("source-" + suffix)
	now := time.Now().UTC()
	var tenantArg any
	if tenantID != "" {
		tenantArg = tenantID
	}
	if err := r.db.Exec(ctx, `
		INSERT INTO data_source (id, tenant_id, type_id, name, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, TRUE, $5, $5)`, id, tenantArg, typeID, name, now); err != nil {
		return nil, err
	}
	r.track("data_source", id)
	return &DataSource{ID: id, TenantID: tenantID, Name: name}, nil
}

type Guest struct {
	ID          string
	DisplayName string
	Status      string
}

func (r *Registry) Guest(ctx context.Context, suffix string) (*Guest, error) {
	id := r.NewID()
	name := r.Name("guest-" + suffix)
	now := time.Now().UTC()
	if err := r.db.Exec(ctx, `
		INSERT INTO guest (id, display_name, status, created_at, updated_at)
		VALUES ($1, $2, 'active', $3, $3)`, id, name, now); err != nil {
		return nil, err
	}
	r.track("guest", id)
	return &Guest{ID: id, DisplayName: name, Status: "active"}, nil
}

type Identifier struct {
	ID     string
	Guest  string
	Type   string
	Value  string
	Source string
}

func (r *Registry) Identifier(ctx context.Context, guestID, idType, value, dataSourceID string) (*Identifier, error) {
	id := r.NewID()
	now := time.Now().UTC()
	if err := r.db.Exec(ctx, `
		INSERT INTO guest_identifier (id, guest_id, origin_guest_id, id_type, value_normalized, data_source_id, first_seen_at)
		VALUES ($1, $2, $2, $3, $4, $5, $6)`, id, guestID, idType, value, dataSourceID, now); err != nil {
		return nil, err
	}
	r.track("guest_identifier", id)
	return &Identifier{ID: id, Guest: guestID, Type: idType, Value: value, Source: dataSourceID}, nil
}

type Event struct {
	ID         string
	GuestID    string
	TenantID   string
	TypeID     string
	SourceID   string
	ExternalID string
	OccurredAt time.Time
	Amount     *int64
}

func (r *Registry) Event(ctx context.Context, guestID, tenantID, eventTypeID, dataSourceID, externalSuffix string, occurredAt time.Time, amount *int64) (*Event, error) {
	id := r.NewID()
	externalID := r.Name("ext-" + externalSuffix)
	if err := r.db.Exec(ctx, `
		INSERT INTO event (id, guest_id, origin_guest_id, tenant_id, event_type_id, data_source_id,
			external_id, occurred_at, amount_minor, currency, created_at)
		VALUES ($1, $2, $2, $3, $4, $5, $6, $7, $8, 'RUB', $9)`,
		id, guestID, tenantID, eventTypeID, dataSourceID, externalID, occurredAt.UTC(), amount, time.Now().UTC()); err != nil {
		return nil, err
	}
	r.track("event", id)
	return &Event{ID: id, GuestID: guestID, TenantID: tenantID, TypeID: eventTypeID, SourceID: dataSourceID, ExternalID: externalID, OccurredAt: occurredAt, Amount: amount}, nil
}

type MergeCandidate struct {
	ID       string
	GuestAID string
	GuestBID string
	Score    int
}

func (r *Registry) MergeCandidate(ctx context.Context, guestAID, guestBID, matchType, reasonCode string, score int) (*MergeCandidate, error) {
	a, b := guestAID, guestBID
	if a > b {
		a, b = b, a
	}
	id := r.NewID()
	if err := r.db.Exec(ctx, `
		INSERT INTO merge_candidate (id, guest_a_id, guest_b_id, match_type, reason_code, reason_text, score, resolution, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'pending', $8)`,
		id, a, b, matchType, reasonCode, "E2E candidate "+reasonCode, score, time.Now().UTC()); err != nil {
		return nil, err
	}
	r.track("merge_candidate", id)
	return &MergeCandidate{ID: id, GuestAID: a, GuestBID: b, Score: score}, nil
}

type Segment struct {
	ID        string
	VersionID string
	Name      string
}

func (r *Registry) Segment(ctx context.Context, suffix, ruleJSON string, memberIDs []string) (*Segment, error) {
	id := r.NewID()
	versionID := r.NewID()
	name := r.Name("segment-" + suffix)
	now := time.Now().UTC()
	if err := r.db.Exec(ctx, `
		INSERT INTO segment (id, name, description, current_version_no, status, created_at, updated_at)
		VALUES ($1, $2, $3, 1, 'active', $4, $4)`, id, name, "E2E segment "+suffix, now); err != nil {
		return nil, err
	}
	r.track("segment", id)
	if err := r.db.Exec(ctx, `
		INSERT INTO segment_version (id, segment_id, version_no, rule, status, member_count, computed_at, created_at)
		VALUES ($1, $2, 1, $3, 'ready', $4, $5, $5)`, versionID, id, ruleJSON, len(memberIDs), now); err != nil {
		return nil, err
	}
	r.track("segment_version", versionID)
	for _, guestID := range memberIDs {
		if err := r.db.Exec(ctx, `
			INSERT INTO segment_member (segment_version_id, guest_id, added_at)
			VALUES ($1, $2, $3)`, versionID, guestID, now); err != nil {
			return nil, err
		}
	}
	return &Segment{ID: id, VersionID: versionID, Name: name}, nil
}

type User struct {
	ID       string
	Email    string
	Role     string
	TenantID string
	Password string
}

func (r *Registry) User(ctx context.Context, suffix, role, tenantID string) (*User, error) {
	password := "e2e-" + r.runID + "-" + suffix
	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("hash fixture password: %w", err)
	}
	id := r.NewID()
	email := r.Name("user-"+suffix) + "@e2e.local"
	var tenantArg any
	if tenantID != "" {
		tenantArg = tenantID
	}
	if err := r.db.Exec(ctx, `
		INSERT INTO app_user (id, email, password_hash, display_name, role, tenant_id, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, TRUE, $7)`,
		id, email, hash, "E2E "+suffix, role, tenantArg, time.Now().UTC()); err != nil {
		return nil, err
	}
	r.track("app_user", id)
	return &User{ID: id, Email: email, Role: role, TenantID: tenantID, Password: password}, nil
}

type ContactLimit struct {
	ID          string
	Channel     string
	MaxContacts int
	PeriodDays  int
}

func (r *Registry) ContactLimit(ctx context.Context, channel string, maxContacts, periodDays int) (*ContactLimit, error) {
	owner, err := r.ownerID(ctx)
	if err != nil {
		return nil, err
	}
	id := r.NewID()
	now := time.Now().UTC()
	var channelArg any
	if channel != "" {
		channelArg = channel
	}
	if err := r.db.Exec(ctx, `
		INSERT INTO contact_limit (id, channel, max_contacts, period_days, is_active, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, TRUE, $5, $6, $6)`,
		id, channelArg, maxContacts, periodDays, owner, now); err != nil {
		return nil, err
	}
	r.track("contact_limit", id)
	return &ContactLimit{ID: id, Channel: channel, MaxContacts: maxContacts, PeriodDays: periodDays}, nil
}

func (r *Registry) ownerID(ctx context.Context) (string, error) {
	var id string
	if err := r.db.QueryRowContext(ctx, `SELECT id FROM app_user ORDER BY created_at LIMIT 1`).Scan(&id); err != nil {
		return "", fmt.Errorf("resolve fixture owner: %w", err)
	}
	return id, nil
}

func (r *Registry) SegmentWithContacts(ctx context.Context, suffix string, guests []*Guest, idType, dataSourceID string) (*Segment, error) {
	memberIDs := make([]string, 0, len(guests))
	for i, g := range guests {
		value := r.Phone(i)
		if idType == "email" {
			value = fmt.Sprintf("%s-guest%d@e2e.local", r.Name("guest"), i)
		}
		if _, err := r.Identifier(ctx, g.ID, idType, value, dataSourceID); err != nil {
			return nil, err
		}
		memberIDs = append(memberIDs, g.ID)
	}
	return r.Segment(ctx, suffix, `{"visitedTenantIds":[],"excludedTenantIds":[],"periodDays":90}`, memberIDs)
}
