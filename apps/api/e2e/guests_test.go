package e2e

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/MangriMen/UCRP/e2e/client"
	"github.com/MangriMen/UCRP/e2e/harness"
	"github.com/MangriMen/UCRP/internal/domain"
	"github.com/MangriMen/UCRP/internal/repository"
)

func TestGuestContactsList(t *testing.T) {
	s := harness.New(t).WithDB().WithMarketer()
	ew := buildExportWorld(t, s, "guest-contacts", 1)
	ctx, cancel := s.Ctx()
	defer cancel()

	export, err := exportCreate(ctx, s.API, ew.segment.ID, s.Reg.Name("contacts-key"))
	if err != nil {
		t.Fatalf("POST /api/v1/exports: %v", err)
	}
	s.RequireStatus(export, 201, "POST /api/v1/exports")
	exportID := harness.StringField(jsonObject(t, export, "created export"), "id")
	trackExport(s, exportID)

	resp, err := s.API.Get(ctx, "/api/v1/guests/"+ew.guests[0].ID+"/contacts", nil)
	if err != nil {
		t.Fatalf("GET /api/v1/guests/{id}/contacts: %v", err)
	}
	s.RequireStatus(resp, 200, "GET /api/v1/guests/{id}/contacts")
	s.RequireDeclaredStatus("GET", "/api/v1/guests/{id}/contacts", resp)
	s.RequireSchema("GET", "/api/v1/guests/{id}/contacts", resp)

	items, ok := harness.Items(mustJSON(t, resp, "contact history"))
	if !ok || len(items) != 1 {
		t.Fatalf("the exported guest must have exactly one contact record: %s", resp.Snippet())
	}
	entry, _ := items[0].(map[string]any)
	if got := harness.StringField(entry, "exportId"); got != exportID {
		t.Fatalf("contact record exportId: expected %s, got %q", exportID, got)
	}
}

func TestGuestContactsUnknownReturns404(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := s.API.Get(ctx, "/api/v1/guests/00000000-0000-0000-0000-000000000000/contacts", nil)
	if err != nil {
		t.Fatalf("GET /api/v1/guests/{unknown}/contacts: %v", err)
	}
	s.RequireStatus(resp, 404, "GET /api/v1/guests/{unknown}/contacts")
	s.RequireDeclaredStatus("GET", "/api/v1/guests/{id}/contacts", resp)
	s.RequireSchema("GET", "/api/v1/guests/{id}/contacts", resp)
}

func TestGuestSoftDelete(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	guest, err := s.Reg.Guest(ctx, "guest-delete")
	if err != nil {
		t.Fatalf("fixture guest: %v", err)
	}

	resp, err := s.API.Delete(ctx, "/api/v1/guests/"+guest.ID)
	if err != nil {
		t.Fatalf("DELETE /api/v1/guests/{id}: %v", err)
	}
	s.RequireStatus(resp, 204, "DELETE /api/v1/guests/{id}")
	s.RequireDeclaredStatus("DELETE", "/api/v1/guests/{id}", resp)

	status, err := s.DB.Count(ctx, `SELECT COUNT(*) FROM guest WHERE id = $1 AND status = 'deleted' AND deleted_at IS NOT NULL`, guest.ID)
	if err != nil {
		t.Fatalf("read the deleted guest: %v", err)
	}
	if status != 1 {
		t.Fatalf("the guest must be soft-deleted, not removed; matching rows: %d", status)
	}

	search, err := s.API.Get(ctx, "/api/v1/guests", map[string][]string{"q": {guest.DisplayName}})
	if err != nil {
		t.Fatalf("GET /api/v1/guests after deletion: %v", err)
	}
	s.RequireStatus(search, 200, "GET /api/v1/guests after the soft delete")
	items, _ := harness.Items(mustJSON(t, search, "guest search"))
	if containsID(items, guest.ID) {
		t.Fatalf("a soft-deleted guest %s must not appear in search results", guest.ID)
	}
}

func TestGuestDeleteUnknownReturns404(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := s.API.Delete(ctx, "/api/v1/guests/00000000-0000-0000-0000-000000000000")
	if err != nil {
		t.Fatalf("DELETE /api/v1/guests/{unknown}: %v", err)
	}
	s.RequireStatus(resp, 404, "DELETE /api/v1/guests/{unknown}")
	s.RequireDeclaredStatus("DELETE", "/api/v1/guests/{id}", resp)
}

func TestGuestDeleteForbiddenForMarketer(t *testing.T) {
	s := harness.New(t).WithDB().WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	guest, err := s.Reg.Guest(ctx, "guest-delete-role")
	if err != nil {
		t.Fatalf("fixture guest: %v", err)
	}

	marketer := actor(t, s, "marketer")
	resp, err := marketer.Delete(ctx, "/api/v1/guests/"+guest.ID)
	if err != nil {
		t.Fatalf("DELETE /api/v1/guests/{id} as marketer: %v", err)
	}
	s.RequireStatus(resp, 403, "DELETE /api/v1/guests/{id} as marketer")
}

type identifiedGuest struct {
	id    string
	phone string
	email string
}

func guestWithPhoneAndEmail(t *testing.T, s *harness.Suite, suffix string, phoneIndex int) identifiedGuest {
	t.Helper()
	w := buildWorld(t, s, suffix)
	ctx, cancel := s.Ctx()
	defer cancel()

	guest, err := s.Reg.Guest(ctx, suffix)
	if err != nil {
		t.Fatalf("fixture guest: %v", err)
	}
	identified := identifiedGuest{
		id:    guest.ID,
		phone: s.Reg.Phone(phoneIndex),
		email: strings.ToLower(s.Reg.Name(suffix)) + "@example.test",
	}
	if _, err := s.Reg.Identifier(ctx, guest.ID, "phone", identified.phone, w.source.ID); err != nil {
		t.Fatalf("fixture phone: %v", err)
	}
	if _, err := s.Reg.Identifier(ctx, guest.ID, "email", identified.email, w.source.ID); err != nil {
		t.Fatalf("fixture email: %v", err)
	}
	return identified
}

func searchGuestRow(t *testing.T, s *harness.Suite, c *client.Client, guest identifiedGuest) map[string]any {
	t.Helper()
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := c.Get(ctx, "/api/v1/guests", map[string][]string{"q": {guest.phone}})
	if err != nil {
		t.Fatalf("GET /api/v1/guests: %v", err)
	}
	s.RequireStatus(resp, 200, "GET /api/v1/guests by phone")
	s.RequireSchema("GET", "/api/v1/guests", resp)
	for _, raw := range []string{guest.phone, guest.email} {
		if strings.Contains(string(resp.Body), raw) {
			t.Fatalf("the guest list must never return a raw identifier %q: %s", raw, resp.Snippet())
		}
	}
	items, _ := harness.Items(mustJSON(t, resp, "guest search"))
	for _, item := range items {
		row, _ := item.(map[string]any)
		if harness.StringField(row, "id") == guest.id {
			return row
		}
	}
	t.Fatalf("search by phone did not return guest %s: %s", guest.id, resp.Snippet())
	return nil
}

func TestGuestListMasksIdentifiers(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	guest := guestWithPhoneAndEmail(t, s, "masked-ids", 61)

	row := searchGuestRow(t, s, s.API, guest)

	identifiers, ok := row["maskedIdentifiers"].([]any)
	if !ok || len(identifiers) != 2 {
		t.Fatalf("expected two masked identifiers, got %v", row)
	}
	got := map[string]string{}
	for _, item := range identifiers {
		entry, _ := item.(map[string]any)
		got[harness.StringField(entry, "idType")] = harness.StringField(entry, "valueMasked")
	}
	want := map[string]string{
		"phone": "***" + guest.phone[len(guest.phone)-4:],
		"email": guest.email[:1] + "***@example.test",
	}
	for idType, value := range want {
		if got[idType] != value {
			t.Fatalf("%s: expected masked value %q, got %q", idType, value, got[idType])
		}
	}
}

func TestGuestListOmitsIdentifiersWithoutIdentifiersPermission(t *testing.T) {
	s := harness.New(t).WithDB()
	guest := guestWithPhoneAndEmail(t, s, "masked-ids-denied", 62)
	ctx, cancel := s.Ctx()
	defer cancel()

	manager, err := s.Reg.User(ctx, "masked-ids-manager", "manager", "")
	if err != nil {
		t.Fatalf("fixture manager: %v", err)
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()
		for _, query := range []string{
			`DELETE FROM access_grant_item WHERE grant_id IN (SELECT id FROM access_grant WHERE subject_type = 'user' AND subject_id = $1)`,
			`DELETE FROM access_grant_scope WHERE grant_id IN (SELECT id FROM access_grant WHERE subject_type = 'user' AND subject_id = $1)`,
			`DELETE FROM audit_log WHERE entity_id IN (SELECT id FROM access_grant WHERE subject_type = 'user' AND subject_id = $1)`,
			`DELETE FROM access_grant WHERE subject_type = 'user' AND subject_id = $1`,
			`DELETE FROM audit_log WHERE actor_user_id = $1`,
		} {
			if err := s.DB.Exec(cleanupCtx, query, manager.ID); err != nil {
				t.Errorf("cleanup manager grant: %v", err)
			}
		}
		if err := s.DB.Exec(cleanupCtx, `UPDATE access_policy_state SET version = version + 1, updated_at = CURRENT_TIMESTAMP WHERE id = 1`); err != nil {
			t.Errorf("invalidate access policy: %v", err)
		}
	})

	searchPermissions := []string{
		"guest.profile.search", "guest.identifier.read", "field.guest.display_name.read",
		"field.guest.tenants_visited.read", "field.guest.phone.read", "field.guest.email.read", "field.guest.card.read",
	}
	items := make([]domain.AccessGrantItem, 0, len(searchPermissions))
	for _, permission := range searchPermissions {
		items = append(items, domain.AccessGrantItem{PermissionCode: permission, Effect: "allow"})
	}
	repo := repository.NewAccessSQL(s.DB.DB)
	version, err := repo.PolicyVersion(ctx)
	if err != nil {
		t.Fatalf("read policy version: %v", err)
	}
	now := time.Now().UTC()
	grantID := s.Reg.NewID()
	grant := &domain.AccessGrant{ID: grantID, SubjectType: "user", SubjectID: manager.ID,
		Status: "active", ValidFrom: now.Add(-time.Minute), CreatedBy: manager.ID, Reason: "E2E guest search without identifiers",
		CreatedAt: now, UpdatedAt: now, Items: items,
		Scopes: []domain.AccessGrantScope{{Type: "principal_role", Value: "manager"}}}
	audit := &domain.AuditLog{ID: s.Reg.NewID(), ActorUserID: manager.ID, ActorRole: "manager", Action: "access.grant.create", EntityType: "access_grant", EntityID: grantID, CreatedAt: now}
	if err := repo.CreateGrant(ctx, grant, version, audit); err != nil {
		t.Fatalf("fixture manager grant: %v", err)
	}

	row := searchGuestRow(t, s, s.As(manager.Email, manager.Password), guest)

	if _, present := row["maskedIdentifiers"]; present {
		t.Fatalf("maskedIdentifiers must be omitted without field.guest.identifiers.read: %v", row)
	}
	if _, present := row["totalAmountMinor"]; present {
		t.Fatalf("totalAmountMinor must be omitted without field.guest.total_amount.read: %v", row)
	}
}
