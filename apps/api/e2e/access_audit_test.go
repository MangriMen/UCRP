package e2e

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/MangriMen/UCRP/e2e/harness"
)

func TestAccessAuditFiltersAndPagination(t *testing.T) {
	s := harness.New(t).WithDB()
	ctx, cancel := s.Ctx()
	defer cancel()
	manager, err := s.Reg.User(ctx, "access-audit-manager", "manager", "")
	if err != nil {
		t.Fatalf("fixture manager: %v", err)
	}
	other, err := s.Reg.User(ctx, "access-audit-other", "manager", "")
	if err != nil {
		t.Fatalf("fixture other actor: %v", err)
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()
		if err := s.DB.Exec(cleanupCtx, `DELETE FROM audit_log WHERE actor_user_id IN ($1,$2)`, manager.ID, other.ID); err != nil {
			t.Errorf("cleanup audit fixtures: %v", err)
		}
	})
	now := time.Now().UTC().Truncate(time.Second)
	grantID := s.Reg.NewID()
	for _, row := range []struct {
		actor, action, kind, entity string
		when                        time.Time
	}{
		{manager.ID, "access.grant.create", "access_grant", grantID, now.Add(-2 * time.Minute)},
		{manager.ID, "access.denied", "access", grantID, now.Add(-time.Minute)},
		{manager.ID, "guest.delete", "guest", grantID, now},
		{other.ID, "access.denied", "access", grantID, now},
	} {
		if err := s.DB.Exec(ctx, `
			INSERT INTO audit_log (id, actor_user_id, actor_role, action, entity_type, entity_id, details, created_at)
			VALUES ($1,$2,'manager',$3,$4,$5,'{}',$6)`, s.Reg.NewID(), row.actor, row.action, row.kind, row.entity, row.when); err != nil {
			t.Fatalf("insert audit fixture: %v", err)
		}
	}
	api := s.As(manager.Email, manager.Password)
	get := func(query url.Values) []struct {
		Action string `json:"action"`
	} { t.Helper(); response, err := api.Get(ctx, "/api/v1/access/audit", query); if err != nil {
		t.Fatalf("GET access/audit: %v", err)
	}; s.RequireStatus(response, 200, "GET access/audit"); var body struct {
		Items []struct {
			Action string `json:"action"`
		} `json:"items"`
	}; if err := response.Decode(&body); err != nil {
		t.Fatalf("decode audit: %v", err)
	}; return body.Items }
	query := url.Values{"actor": {manager.ID}, "from": {now.Add(-3 * time.Minute).Format(time.RFC3339)},
		"to": {now.Format(time.RFC3339)}, "limit": {"1"}, "offset": {"0"}}
	items := get(query)
	if len(items) != 1 || items[0].Action != "access.denied" {
		t.Fatalf("first page should contain the newest access denial: %+v", items)
	}
	query.Set("offset", "1")
	items = get(query)
	if len(items) != 1 || items[0].Action != "access.grant.create" {
		t.Fatalf("second page should contain the grant event: %+v", items)
	}
	query.Set("offset", "0")
	query.Set("grant_id", grantID)
	items = get(query)
	if len(items) != 1 || items[0].Action != "access.grant.create" {
		t.Fatalf("grant filter included a denial or unrelated audit entry: %+v", items)
	}
}
