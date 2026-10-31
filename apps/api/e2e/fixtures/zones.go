package fixtures

import (
	"context"
	"time"
)

type Mall struct {
	ID   string
	Code string
	Name string
}

func (r *Registry) Mall(ctx context.Context, suffix string) (*Mall, error) {
	id := r.NewID()
	code := r.Code("mall-" + suffix)
	name := "E2E mall " + suffix
	err := r.db.Exec(ctx, `
		INSERT INTO mall (id, code, name, is_active, created_at)
		VALUES ($1, $2, $3, TRUE, $4)`, id, code, name, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	r.track("mall", id)
	return &Mall{ID: id, Code: code, Name: name}, nil
}

func (r *Registry) TenantInZone(ctx context.Context, suffix, categoryID, mallID string, floor int) (*Tenant, error) {
	id := r.NewID()
	code := r.Code("tenant-" + suffix)
	name := "E2E tenant " + suffix
	now := time.Now().UTC()
	err := r.db.Exec(ctx, `
		INSERT INTO tenant (id, code, name, mall_id, category_id, kind, floor, priority, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, 'mall', $6, 50, TRUE, $7, $7)`, id, code, name, mallID, categoryID, floor, now)
	if err != nil {
		return nil, err
	}
	r.track("tenant", id)
	return &Tenant{ID: id, Code: code, Name: name, CategoryID: categoryID, Priority: 50}, nil
}

func (r *Registry) PresenceEvent(ctx context.Context, eventTypeCode, guestID, tenantID, dataSourceID, externalSuffix string, occurredAt time.Time) error {
	if err := r.db.Exec(ctx, `
		INSERT INTO event_type (id, code, name, has_amount, counts_as_presence, is_active)
		SELECT $1::text, $2::text, $2::text, TRUE, TRUE, TRUE
		WHERE NOT EXISTS (SELECT 1 FROM event_type WHERE code = $2::text)`, r.NewID(), eventTypeCode); err != nil {
		return err
	}
	id := r.NewID()
	if err := r.db.Exec(ctx, `
		INSERT INTO event (id, guest_id, origin_guest_id, tenant_id, event_type_id, data_source_id,
			external_id, occurred_at, amount_minor, currency, created_at)
		VALUES ($1, $2, $2, $3, (SELECT id FROM event_type WHERE code = $4), $5, $6, $7, 1000, 'RUB', $8)`,
		id, guestID, tenantID, eventTypeCode, dataSourceID, r.Name("ext-"+externalSuffix), occurredAt.UTC(), time.Now().UTC()); err != nil {
		return err
	}
	r.track("event", id)
	return nil
}
