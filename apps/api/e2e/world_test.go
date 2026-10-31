package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/MangriMen/UCRP/e2e/fixtures"
	"github.com/MangriMen/UCRP/e2e/harness"
)

type world struct {
	category  *fixtures.Category
	tenantA   *fixtures.Tenant
	tenantB   *fixtures.Tenant
	tenantC   *fixtures.Tenant
	eventType *fixtures.EventType
	source    *fixtures.DataSource
}

func buildWorld(t *testing.T, s *harness.Suite, suffix string) *world {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	category, err := s.Reg.Category(ctx, suffix)
	if err != nil {
		t.Fatalf("fixture category: %v", err)
	}
	tenantA, err := s.Reg.Tenant(ctx, suffix+"-a", category.ID, 80)
	if err != nil {
		t.Fatalf("fixture tenant A: %v", err)
	}
	tenantB, err := s.Reg.Tenant(ctx, suffix+"-b", category.ID, 60)
	if err != nil {
		t.Fatalf("fixture tenant B: %v", err)
	}
	tenantC, err := s.Reg.Tenant(ctx, suffix+"-c", category.ID, 40)
	if err != nil {
		t.Fatalf("fixture tenant C: %v", err)
	}
	eventType, err := s.Reg.EventType(ctx, suffix, true)
	if err != nil {
		t.Fatalf("fixture event type: %v", err)
	}
	source, err := s.Reg.DataSource(ctx, suffix, tenantA.ID)
	if err != nil {
		t.Fatalf("fixture data source: %v", err)
	}
	return &world{
		category:  category,
		tenantA:   tenantA,
		tenantB:   tenantB,
		tenantC:   tenantC,
		eventType: eventType,
		source:    source,
	}
}

func (w *world) visit(t *testing.T, s *harness.Suite, guestID, tenantID, suffix string, daysAgo int) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	occurredAt := time.Now().UTC().AddDate(0, 0, -daysAgo)
	if _, err := s.Reg.Event(ctx, guestID, tenantID, w.eventType.ID, w.source.ID, suffix, occurredAt, nil); err != nil {
		t.Fatalf("fixture event %s: %v", suffix, err)
	}
}
