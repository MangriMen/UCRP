package seed

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

func (s *Seeder) seedGuests(ctx context.Context) error {
	for i := 0; i < s.opts.Guests; i++ {
		g := backgroundGuest(i)
		if err := s.insertGuest(ctx, g); err != nil {
			return err
		}
	}
	for _, g := range demoPersonas() {
		if err := s.insertGuest(ctx, g); err != nil {
			return err
		}
	}
	for _, g := range demoMergedGuests() {
		if err := s.insertGuest(ctx, g); err != nil {
			return err
		}
	}
	for _, g := range demoDeletedGuests() {
		if err := s.insertGuest(ctx, g); err != nil {
			return err
		}
	}
	return nil
}

func (s *Seeder) insertGuest(ctx context.Context, g guestSpec) error {
	id := demoID("guest", g.Key)
	s.guests[g.Key] = id
	createdAt := s.now.AddDate(0, 0, -g.CreatedDaysAgo)
	var name any
	if g.Name != "" {
		name = g.Name
	}
	var mergedInto, deletedAt any
	if g.Status == "merged" {
		target, ok := s.guests[g.MergedInto]
		if !ok {
			return fmt.Errorf("guest %s: unknown merge target %s", g.Key, g.MergedInto)
		}
		mergedInto = target
	}
	if g.Status == "deleted" {
		deletedAt = createdAt.AddDate(0, 0, 30)
	}
	if err := s.exec(ctx, `
		INSERT INTO guest (id, display_name, status, merged_into_id, created_at, updated_at, deleted_at)
		VALUES ($1,$2,$3,$4,$5,$5,$6)
		ON CONFLICT (id) DO NOTHING`,
		id, name, g.Status, mergedInto, createdAt, deletedAt); err != nil {
		return err
	}
	for _, ident := range g.Identifiers {
		identID := demoID("guest_identifier", g.Key+"/"+ident.Type)
		s.identifiers[g.Key+"/"+ident.Type] = identID
		s.identValues[g.Key+"/"+ident.Type] = ident.Value
		if err := s.exec(ctx, `
			INSERT INTO guest_identifier (id, guest_id, origin_guest_id, id_type, value_normalized, data_source_id, first_seen_at)
			VALUES ($1,$2,$2,$3,$4,$5,$6)
			ON CONFLICT (id) DO NOTHING`,
			identID, id, ident.Type, ident.Value, s.sources[ident.Source], createdAt); err != nil {
			return err
		}
	}
	return nil
}

func (s *Seeder) seedEvents(ctx context.Context) error {
	events := make([]eventSpec, 0, s.opts.Guests*16)
	for i := 0; i < s.opts.Guests; i++ {
		events = append(events, backgroundGuest(i).Events...)
	}
	for _, g := range demoPersonas() {
		events = append(events, g.Events...)
	}
	events = append(events, unmatchedEvents()...)

	byExternal := map[string]string{}
	for _, ev := range events {
		id := demoID("event", ev.Source+"/"+ev.ExternalID)
		byExternal[ev.ExternalID] = id
	}
	for _, ev := range events {
		if err := s.insertEvent(ctx, ev, byExternal); err != nil {
			return err
		}
	}
	return nil
}

func unmatchedEvents() []eventSpec {
	return []eventSpec{
		{Tenant: "cafe-bean", Type: "purchase", Source: "pos-bean", DaysAgo: 4, Hour: 10, Minute: 12, Amount: 31000, HasAmount: true, ExternalID: "CHK-999001", Keys: []string{"phone"}},
		{Tenant: "cinema-sinema", Project: "imax", Type: "purchase", Source: "pos-cinema", DaysAgo: 2, Hour: 21, Minute: 30, Amount: 210000, HasAmount: true, ExternalID: "CHK-999002", Keys: []string{"email"}},
	}
}

func (s *Seeder) insertEvent(ctx context.Context, ev eventSpec, byExternal map[string]string) error {
	id := demoID("event", ev.Source+"/"+ev.ExternalID)
	occurredAt := s.now.AddDate(0, 0, -ev.DaysAgo).
		Truncate(24 * time.Hour).
		Add(time.Duration(ev.Hour)*time.Hour + time.Duration(ev.Minute)*time.Minute)

	var guestID, originGuestID any
	if ev.GuestKey != "" {
		guestID = s.guests[ev.GuestKey]
		originGuestID = guestID
	}
	var projectID any
	if ev.Project != "" {
		projectID = s.projects[ev.Tenant+"/"+ev.Project]
	}
	var amount any
	if ev.HasAmount {
		amount = ev.Amount
	}
	var reverses any
	if ev.Reverses != "" {
		reverses = byExternal[ev.Reverses]
	}
	matchKeys := map[string]string{}
	for _, key := range ev.Keys {
		if value, ok := s.identifierValue(ev.GuestKey, key); ok {
			matchKeys[key] = value
		}
	}
	matchJSON, err := json.Marshal(matchKeys)
	if err != nil {
		return err
	}
	rawJSON, err := json.Marshal(map[string]string{
		"external_id": ev.ExternalID,
		"event_type":  ev.Type,
		"tenant_code": ev.Tenant,
	})
	if err != nil {
		return err
	}
	return s.exec(ctx, `
		INSERT INTO event (id, guest_id, origin_guest_id, tenant_id, project_id, event_type_id, data_source_id,
			external_id, occurred_at, amount_minor, currency, reverses_event_id, match_keys, raw_payload, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,'RUB',$11,$12,$13,$14)
		ON CONFLICT (id) DO NOTHING`,
		id, guestID, originGuestID, s.tenants[ev.Tenant], projectID, s.eventTypes[ev.Type], s.sources[ev.Source],
		ev.ExternalID, occurredAt, amount, reverses, string(matchJSON), string(rawJSON), occurredAt)
}

func (s *Seeder) identifierValue(guestKey, idType string) (string, bool) {
	value, ok := s.identValues[guestKey+"/"+idType]
	return value, ok
}
