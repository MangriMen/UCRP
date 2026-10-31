package e2e

import (
	"context"
	"testing"

	"github.com/MangriMen/UCRP/e2e/harness"
)

func TestIdentityResolveCreatesGuestFromImport(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	iw := buildImportWorld(t, s, "resolve-create")
	ctx, cancel := s.Ctx()
	defer cancel()

	before, err := s.DB.Count(ctx, `SELECT COUNT(*) FROM guest`)
	if err != nil {
		t.Fatalf("count guests before import: %v", err)
	}

	resp, err := uploadImport(ctx, s, iw.source.ID, "valid.csv", csvValid(iw.world.tenantA.Code, s.Reg.Phone(1), s.Reg.Phone(2)))
	if err != nil {
		t.Fatalf("POST /api/v1/imports: %v", err)
	}
	s.RequireStatus(resp, 202, "POST /api/v1/imports")
	body, err := resp.JSON()
	if err != nil {
		t.Fatalf("decode import response: %v", err)
	}
	obj, _ := body.(map[string]any)
	batchID := harness.StringField(obj, "batchId")
	trackImport(s, batchID)

	s.WaitFor(s.Cfg.ImportTimeout, "identity resolution to attach events", func(ctx context.Context) (bool, error) {
		attached, err := s.DB.Count(ctx, `SELECT COUNT(*) FROM event WHERE import_batch_id = $1 AND guest_id IS NOT NULL`, batchID)
		if err != nil {
			return false, err
		}
		return attached >= 2, nil
	})

	after, err := s.DB.Count(ctx, `SELECT COUNT(*) FROM guest`)
	if err != nil {
		t.Fatalf("count guests after import: %v", err)
	}
	if after <= before {
		t.Fatalf("identity resolution must create guests for unmatched events: before=%d after=%d", before, after)
	}

	identifiers, err := s.DB.Count(ctx, `
		SELECT COUNT(*) FROM guest_identifier i
		JOIN event e ON e.guest_id = i.guest_id
		WHERE e.import_batch_id = $1 AND i.id_type = 'phone'`, batchID)
	if err != nil {
		t.Fatalf("count identifiers: %v", err)
	}
	if identifiers == 0 {
		t.Fatalf("identity resolution must create phone identifiers for imported events")
	}
}

func TestIdentityResolveAttachesToExistingGuest(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	iw := buildImportWorld(t, s, "resolve-attach")
	ctx, cancel := s.Ctx()
	defer cancel()

	phone := s.Reg.Phone(15)
	existing, err := s.Reg.Guest(ctx, "resolve-attach-existing")
	if err != nil {
		t.Fatalf("fixture guest: %v", err)
	}
	if _, err := s.Reg.Identifier(ctx, existing.ID, "phone", phone, iw.source.ID); err != nil {
		t.Fatalf("fixture identifier: %v", err)
	}

	content := []byte(csvHeader +
		"e2e-attach,2026-10-05T10:00:00Z,purchase,1000," + phone + ",," + iw.world.tenantA.Code + "\n")
	resp, err := uploadImport(ctx, s, iw.source.ID, "attach.csv", content)
	if err != nil {
		t.Fatalf("POST /api/v1/imports: %v", err)
	}
	s.RequireStatus(resp, 202, "POST /api/v1/imports")
	body, err := resp.JSON()
	if err != nil {
		t.Fatalf("decode import response: %v", err)
	}
	obj, _ := body.(map[string]any)
	batchID := harness.StringField(obj, "batchId")
	trackImport(s, batchID)

	s.WaitFor(s.Cfg.ImportTimeout, "event to attach to the existing guest", func(ctx context.Context) (bool, error) {
		n, err := s.DB.Count(ctx, `SELECT COUNT(*) FROM event WHERE import_batch_id = $1 AND guest_id = $2`, batchID, existing.ID)
		if err != nil {
			return false, err
		}
		return n == 1, nil
	})
}

func TestIdentityAutoMergeOnSharedPhone(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	iw := buildImportWorld(t, s, "resolve-automerge")
	ctx, cancel := s.Ctx()
	defer cancel()

	phone := s.Reg.Phone(16)
	first, err := s.Reg.Guest(ctx, "automerge-first")
	if err != nil {
		t.Fatalf("fixture first guest: %v", err)
	}
	second, err := s.Reg.Guest(ctx, "automerge-second")
	if err != nil {
		t.Fatalf("fixture second guest: %v", err)
	}
	if _, err := s.Reg.Identifier(ctx, first.ID, "phone", phone, iw.source.ID); err != nil {
		t.Fatalf("fixture first identifier: %v", err)
	}
	if _, err := s.Reg.Identifier(ctx, second.ID, "phone", phone, iw.source.ID); err != nil {
		t.Fatalf("fixture second identifier: %v", err)
	}

	content := []byte(csvHeader +
		"e2e-merge,2026-10-06T10:00:00Z,purchase,1000," + phone + ",," + iw.world.tenantA.Code + "\n")
	resp, err := uploadImport(ctx, s, iw.source.ID, "merge.csv", content)
	if err != nil {
		t.Fatalf("POST /api/v1/imports: %v", err)
	}
	s.RequireStatus(resp, 202, "POST /api/v1/imports")
	body, err := resp.JSON()
	if err != nil {
		t.Fatalf("decode import response: %v", err)
	}
	obj, _ := body.(map[string]any)
	batchID := harness.StringField(obj, "batchId")
	trackImport(s, batchID)

	s.WaitFor(s.Cfg.ImportTimeout, "auto-merge candidate for the shared phone", func(ctx context.Context) (bool, error) {
		n, err := s.DB.Count(ctx, `
			SELECT COUNT(*) FROM merge_candidate
			WHERE (guest_a_id = $1 AND guest_b_id = $2) OR (guest_a_id = $2 AND guest_b_id = $1)`,
			first.ID, second.ID)
		if err != nil {
			return false, err
		}
		return n == 1, nil
	})

	score, err := s.DB.Count(ctx, `
		SELECT COUNT(*) FROM merge_candidate
		WHERE ((guest_a_id = $1 AND guest_b_id = $2) OR (guest_a_id = $2 AND guest_b_id = $1))
		  AND resolution = 'auto_merged' AND score = 90 AND match_type = 'phone'`,
		first.ID, second.ID)
	if err != nil {
		t.Fatalf("read merge candidate: %v", err)
	}
	if score != 1 {
		t.Fatalf("a shared phone must produce an auto_merged candidate with score 90 and match_type phone")
	}
}

func TestIdentityPendingCandidateOnSharedEmail(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	iw := buildImportWorld(t, s, "resolve-pending")
	ctx, cancel := s.Ctx()
	defer cancel()

	email := s.Reg.Name("shared") + "@e2e.local"
	first, err := s.Reg.Guest(ctx, "pending-first")
	if err != nil {
		t.Fatalf("fixture first guest: %v", err)
	}
	second, err := s.Reg.Guest(ctx, "pending-second")
	if err != nil {
		t.Fatalf("fixture second guest: %v", err)
	}
	if _, err := s.Reg.Identifier(ctx, first.ID, "email", email, iw.source.ID); err != nil {
		t.Fatalf("fixture first identifier: %v", err)
	}
	if _, err := s.Reg.Identifier(ctx, second.ID, "email", email, iw.source.ID); err != nil {
		t.Fatalf("fixture second identifier: %v", err)
	}

	content := []byte(csvHeader +
		"e2e-pending,2026-10-07T10:00:00Z,purchase,1000,," + email + "," + iw.world.tenantA.Code + "\n")
	resp, err := uploadImport(ctx, s, iw.source.ID, "pending.csv", content)
	if err != nil {
		t.Fatalf("POST /api/v1/imports: %v", err)
	}
	s.RequireStatus(resp, 202, "POST /api/v1/imports")
	body, err := resp.JSON()
	if err != nil {
		t.Fatalf("decode import response: %v", err)
	}
	obj, _ := body.(map[string]any)
	batchID := harness.StringField(obj, "batchId")
	trackImport(s, batchID)

	s.WaitFor(s.Cfg.ImportTimeout, "pending candidate for the shared email", func(ctx context.Context) (bool, error) {
		n, err := s.DB.Count(ctx, `
			SELECT COUNT(*) FROM merge_candidate
			WHERE ((guest_a_id = $1 AND guest_b_id = $2) OR (guest_a_id = $2 AND guest_b_id = $1))
			  AND resolution = 'pending' AND score = 70 AND match_type = 'email'`,
			first.ID, second.ID)
		if err != nil {
			return false, err
		}
		return n == 1, nil
	})
}
