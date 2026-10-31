package e2e

import (
	"testing"

	"github.com/MangriMen/UCRP/e2e/harness"
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
	if got := harness.StringField(entry, "export_id"); got != exportID {
		t.Fatalf("contact record export_id: expected %s, got %q", exportID, got)
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
