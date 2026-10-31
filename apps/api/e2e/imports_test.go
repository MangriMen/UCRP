package e2e

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/MangriMen/UCRP/e2e/client"
	"github.com/MangriMen/UCRP/e2e/fixtures"
	"github.com/MangriMen/UCRP/e2e/harness"
)

type importWorld struct {
	world  *world
	source *fixtures.DataSource
}

func buildImportWorld(t *testing.T, s *harness.Suite, suffix string) *importWorld {
	t.Helper()
	w := buildWorld(t, s, suffix)
	return &importWorld{world: w, source: w.source}
}

func uploadImport(ctx context.Context, s *harness.Suite, sourceID, fileName string, content []byte) (*client.Response, error) {
	return s.API.PostMultipart(ctx, "/api/v1/imports", map[string]string{"data_source_id": sourceID}, fileName, content)
}

func TestImportUploadAccepted(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	iw := buildImportWorld(t, s, "import-accepted")
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := uploadImport(ctx, s, iw.source.ID, "valid.csv", csvValid(iw.world.tenantA.Code, s.Reg.Phone(1), s.Reg.Phone(2)))
	if err != nil {
		t.Fatalf("POST /api/v1/imports: %v", err)
	}
	s.RequireStatus(resp, 202, "POST /api/v1/imports")
	s.RequireDeclaredStatus("POST", "/api/v1/imports", resp)
	s.RequireSchema("POST", "/api/v1/imports", resp)

	body, err := resp.JSON()
	if err != nil {
		t.Fatalf("decode import response: %v", err)
	}
	obj, _ := body.(map[string]any)
	batchID := harness.StringField(obj, "batch_id", "batchId")
	if batchID == "" {
		t.Fatalf("import response carries no batch id: %s", resp.Snippet())
	}
	if status := harness.StringField(obj, "status"); status != "queued" {
		t.Fatalf("import status: expected queued, got %q", status)
	}
	if duplicate, ok := obj["duplicate"].(bool); !ok || duplicate {
		t.Fatalf("import duplicate flag: expected false, got %v", obj["duplicate"])
	}
	trackImport(ctx, s, batchID)
}

func TestImportCountersInvariant(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	iw := buildImportWorld(t, s, "import-counters")
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := uploadImport(ctx, s, iw.source.ID, "mixed.csv", csvWithRejectedRows(iw.world.tenantA.Code, s.Reg.Phone(3)))
	if err != nil {
		t.Fatalf("POST /api/v1/imports: %v", err)
	}
	s.RequireStatus(resp, 202, "POST /api/v1/imports")
	body, err := resp.JSON()
	if err != nil {
		t.Fatalf("decode import response: %v", err)
	}
	obj, _ := body.(map[string]any)
	batchID := harness.StringField(obj, "batch_id", "batchId")
	trackImport(ctx, s, batchID)

	s.WaitFor(s.Cfg.ImportTimeout, "import batch to finish", func(ctx context.Context) (bool, error) {
		r, err := s.API.Get(ctx, "/api/v1/imports/"+batchID, nil)
		if err != nil {
			return false, err
		}
		if r.Status != 200 {
			return false, nil
		}
		b, err := r.JSON()
		if err != nil {
			return false, err
		}
		o, _ := b.(map[string]any)
		status := harness.StringField(o, "status")
		return status == "completed" || status == "completed_with_errors" || status == "failed", nil
	})

	get, err := s.API.Get(ctx, "/api/v1/imports/"+batchID, nil)
	if err != nil {
		t.Fatalf("GET /api/v1/imports/{id}: %v", err)
	}
	s.RequireStatus(get, 200, "GET /api/v1/imports/{id}")
	s.RequireSchema("GET", "/api/v1/imports/{id}", get)

	decoded, err := get.JSON()
	if err != nil {
		t.Fatalf("decode batch: %v", err)
	}
	batch, _ := decoded.(map[string]any)
	total, _ := harness.IntField(batch, "total_rows", "totalRows", "total")
	accepted, _ := harness.IntField(batch, "accepted_rows", "acceptedRows", "accepted")
	rejected, _ := harness.IntField(batch, "rejected_rows", "rejectedRows", "rejected")
	duplicate, _ := harness.IntField(batch, "duplicate_rows", "duplicateRows", "duplicate")

	if total != 4 {
		t.Fatalf("total rows: expected 4, got %d", total)
	}
	if accepted+rejected+duplicate != total {
		t.Fatalf("counter invariant violated: accepted(%d) + rejected(%d) + duplicate(%d) != total(%d)",
			accepted, rejected, duplicate, total)
	}
	if rejected != 3 {
		t.Fatalf("rejected rows: expected 3 (bad date, no external id, unknown tenant), got %d", rejected)
	}
	if status := harness.StringField(batch, "status"); status != "completed_with_errors" {
		t.Fatalf("batch status: expected completed_with_errors, got %q", status)
	}
}

func TestImportRejectedRowCodes(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	iw := buildImportWorld(t, s, "import-rejected")
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := uploadImport(ctx, s, iw.source.ID, "mixed.csv", csvWithRejectedRows(iw.world.tenantA.Code, s.Reg.Phone(3)))
	if err != nil {
		t.Fatalf("POST /api/v1/imports: %v", err)
	}
	s.RequireStatus(resp, 202, "POST /api/v1/imports")
	body, err := resp.JSON()
	if err != nil {
		t.Fatalf("decode import response: %v", err)
	}
	obj, _ := body.(map[string]any)
	batchID := harness.StringField(obj, "batch_id", "batchId")
	trackImport(ctx, s, batchID)

	s.WaitFor(s.Cfg.ImportTimeout, "import batch to finish", func(ctx context.Context) (bool, error) {
		r, err := s.API.Get(ctx, "/api/v1/imports/"+batchID, nil)
		if err != nil {
			return false, err
		}
		if r.Status != 200 {
			return false, nil
		}
		b, err := r.JSON()
		if err != nil {
			return false, err
		}
		o, _ := b.(map[string]any)
		status := harness.StringField(o, "status")
		return status == "completed" || status == "completed_with_errors" || status == "failed", nil
	})

	rows, err := s.API.Get(ctx, "/api/v1/imports/"+batchID+"/rejected-rows", nil)
	if err != nil {
		t.Fatalf("GET rejected rows: %v", err)
	}
	s.RequireStatus(rows, 200, "GET /api/v1/imports/{id}/rejected-rows")
	s.RequireSchema("GET", "/api/v1/imports/{id}/rejected-rows", rows)

	decoded, err := rows.JSON()
	if err != nil {
		t.Fatalf("decode rejected rows: %v", err)
	}
	items, ok := harness.Items(decoded)
	if !ok {
		t.Fatalf("rejected rows response has no items array: %s", rows.Snippet())
	}
	codes := map[string]bool{}
	for _, item := range items {
		obj, _ := item.(map[string]any)
		codes[harness.StringField(obj, "error_code", "errorCode")] = true
	}
	for _, want := range []string{"bad_date", "no_external_id", "bad_tenant"} {
		if !codes[want] {
			t.Fatalf("rejected rows must contain error code %q, got %v", want, codes)
		}
	}
}

func TestImportRejectedRowsCSV(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	iw := buildImportWorld(t, s, "import-rejected-csv")
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := uploadImport(ctx, s, iw.source.ID, "mixed.csv", csvWithRejectedRows(iw.world.tenantA.Code, s.Reg.Phone(3)))
	if err != nil {
		t.Fatalf("POST /api/v1/imports: %v", err)
	}
	s.RequireStatus(resp, 202, "POST /api/v1/imports")
	body, err := resp.JSON()
	if err != nil {
		t.Fatalf("decode import response: %v", err)
	}
	obj, _ := body.(map[string]any)
	batchID := harness.StringField(obj, "batch_id", "batchId")
	trackImport(ctx, s, batchID)

	s.WaitFor(s.Cfg.ImportTimeout, "import batch to finish", func(ctx context.Context) (bool, error) {
		r, err := s.API.Get(ctx, "/api/v1/imports/"+batchID, nil)
		if err != nil {
			return false, err
		}
		if r.Status != 200 {
			return false, nil
		}
		b, err := r.JSON()
		if err != nil {
			return false, err
		}
		o, _ := b.(map[string]any)
		status := harness.StringField(o, "status")
		return status == "completed" || status == "completed_with_errors" || status == "failed", nil
	})

	csvResp, err := s.API.Get(ctx, "/api/v1/imports/"+batchID+"/rejected-rows", map[string][]string{"format": {"csv"}})
	if err != nil {
		t.Fatalf("GET rejected rows as CSV: %v", err)
	}
	s.RequireStatus(csvResp, 200, "GET rejected rows as CSV")
	contentType := csvResp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/csv") {
		t.Fatalf("rejected rows CSV: expected Content-Type text/csv, got %q", contentType)
	}
	text := string(csvResp.Body)
	if !strings.Contains(text, "row_number") || !strings.Contains(text, "error_code") {
		t.Fatalf("rejected rows CSV must start with a header row, got %q", text)
	}
}

func TestImportUploadIsIdempotent(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	iw := buildImportWorld(t, s, "import-idempotent")
	ctx, cancel := s.Ctx()
	defer cancel()

	content := csvValid(iw.world.tenantA.Code, s.Reg.Phone(1), s.Reg.Phone(2))
	first, err := uploadImport(ctx, s, iw.source.ID, "same.csv", content)
	if err != nil {
		t.Fatalf("first upload: %v", err)
	}
	s.RequireStatus(first, 202, "first upload")
	firstBody, err := first.JSON()
	if err != nil {
		t.Fatalf("decode first upload: %v", err)
	}
	firstObj, _ := firstBody.(map[string]any)
	firstID := harness.StringField(firstObj, "batch_id", "batchId")
	trackImport(ctx, s, firstID)

	second, err := uploadImport(ctx, s, iw.source.ID, "same.csv", content)
	if err != nil {
		t.Fatalf("second upload: %v", err)
	}
	s.RequireStatus(second, 200, "second upload of the same file")
	secondBody, err := second.JSON()
	if err != nil {
		t.Fatalf("decode second upload: %v", err)
	}
	secondObj, _ := secondBody.(map[string]any)
	secondID := harness.StringField(secondObj, "batch_id", "batchId")
	if secondID != firstID {
		t.Fatalf("repeated upload must return the same batch: first=%s second=%s", firstID, secondID)
	}
	if duplicate, ok := secondObj["duplicate"].(bool); !ok || !duplicate {
		t.Fatalf("repeated upload must report duplicate=true, got %v", secondObj["duplicate"])
	}
}

func TestImportUploadRejectsMissingColumns(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	iw := buildImportWorld(t, s, "import-missing-cols")
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := uploadImport(ctx, s, iw.source.ID, "bad.csv", csvMissingColumns())
	if err != nil {
		t.Fatalf("POST /api/v1/imports: %v", err)
	}
	s.RequireStatusOrDrift(resp, 422, "POST /api/v1/imports without required columns")
	s.RequireDeclaredStatus("POST", "/api/v1/imports", resp)
	s.RequireSchema("POST", "/api/v1/imports", resp)
}

func TestImportUploadRejectsEmptyFile(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	iw := buildImportWorld(t, s, "import-empty")
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := uploadImport(ctx, s, iw.source.ID, "empty.csv", csvEmpty())
	if err != nil {
		t.Fatalf("POST /api/v1/imports: %v", err)
	}
	s.RequireStatusOrDrift(resp, 422, "POST /api/v1/imports with an empty file")
	s.RequireDeclaredStatus("POST", "/api/v1/imports", resp)
	s.RequireSchema("POST", "/api/v1/imports", resp)
}

func TestImportUploadUnknownSource(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	iw := buildImportWorld(t, s, "import-unknown-source")
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := uploadImport(ctx, s, "00000000-0000-0000-0000-000000000000", "valid.csv", csvValid(iw.world.tenantA.Code, s.Reg.Phone(1), s.Reg.Phone(2)))
	if err != nil {
		t.Fatalf("POST /api/v1/imports: %v", err)
	}
	s.RequireStatus(resp, 404, "POST /api/v1/imports with an unknown data source")
	s.RequireDeclaredStatus("POST", "/api/v1/imports", resp)
	s.RequireSchema("POST", "/api/v1/imports", resp)
}

func TestImportUploadRequiresFile(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	iw := buildImportWorld(t, s, "import-no-file")
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := s.API.PostMultipart(ctx, "/api/v1/imports", map[string]string{"data_source_id": iw.source.ID}, "", nil)
	if err != nil {
		t.Fatalf("POST /api/v1/imports without a file: %v", err)
	}
	s.RequireStatus(resp, 400, "POST /api/v1/imports without a file")
	s.RequireDeclaredStatus("POST", "/api/v1/imports", resp)
	s.RequireSchema("POST", "/api/v1/imports", resp)
}

func TestImportListFilters(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	iw := buildImportWorld(t, s, "import-list")
	ctx, cancel := s.Ctx()
	defer cancel()

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
	batchID := harness.StringField(obj, "batch_id", "batchId")
	trackImport(ctx, s, batchID)

	list, err := s.API.Get(ctx, "/api/v1/imports", map[string][]string{"data_source_id": {iw.source.ID}})
	if err != nil {
		t.Fatalf("GET /api/v1/imports?data_source_id=: %v", err)
	}
	s.RequireStatus(list, 200, "GET /api/v1/imports filtered by data source")
	s.RequireSchema("GET", "/api/v1/imports", list)

	decoded, err := list.JSON()
	if err != nil {
		t.Fatalf("decode import list: %v", err)
	}
	items, _ := harness.Items(decoded)
	found := false
	for _, item := range items {
		o, _ := item.(map[string]any)
		if harness.StringField(o, "id") == batchID {
			found = true
		}
	}
	if !found {
		t.Fatalf("import list filtered by data_source_id must contain batch %s", batchID)
	}
}

func TestImportGetUnknownReturns404(t *testing.T) {
	s := harness.New(t).WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := s.API.Get(ctx, "/api/v1/imports/00000000-0000-0000-0000-000000000000", nil)
	if err != nil {
		t.Fatalf("GET unknown import: %v", err)
	}
	s.RequireStatus(resp, 404, "GET unknown import batch")
	s.RequireDeclaredStatus("GET", "/api/v1/imports/{id}", resp)
	s.RequireSchema("GET", "/api/v1/imports/{id}", resp)
}

func TestImportDuplicateExternalIDCountedAsDuplicate(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	iw := buildImportWorld(t, s, "import-dup-external")
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := uploadImport(ctx, s, iw.source.ID, "dup.csv", csvDuplicateExternalID(iw.world.tenantA.Code, s.Reg.Phone(4)))
	if err != nil {
		t.Fatalf("POST /api/v1/imports: %v", err)
	}
	s.RequireStatus(resp, 202, "POST /api/v1/imports")
	body, err := resp.JSON()
	if err != nil {
		t.Fatalf("decode import response: %v", err)
	}
	obj, _ := body.(map[string]any)
	batchID := harness.StringField(obj, "batch_id", "batchId")
	trackImport(ctx, s, batchID)

	s.WaitFor(s.Cfg.ImportTimeout, "import batch to finish", func(ctx context.Context) (bool, error) {
		r, err := s.API.Get(ctx, "/api/v1/imports/"+batchID, nil)
		if err != nil {
			return false, err
		}
		if r.Status != 200 {
			return false, nil
		}
		b, err := r.JSON()
		if err != nil {
			return false, err
		}
		o, _ := b.(map[string]any)
		status := harness.StringField(o, "status")
		return status == "completed" || status == "completed_with_errors" || status == "failed", nil
	})

	get, err := s.API.Get(ctx, "/api/v1/imports/"+batchID, nil)
	if err != nil {
		t.Fatalf("GET batch: %v", err)
	}
	decoded, err := get.JSON()
	if err != nil {
		t.Fatalf("decode batch: %v", err)
	}
	batch, _ := decoded.(map[string]any)
	accepted, _ := harness.IntField(batch, "accepted_rows", "acceptedRows", "accepted")
	duplicate, _ := harness.IntField(batch, "duplicate_rows", "duplicateRows", "duplicate")
	if accepted != 1 {
		t.Fatalf("accepted rows: expected 1, got %d", accepted)
	}
	if duplicate != 1 {
		t.Fatalf("duplicate rows: expected 1 (same external_id), got %d", duplicate)
	}
}

func trackImport(ctx context.Context, s *harness.Suite, batchID string) {
	s.T.Helper()
	if batchID == "" {
		return
	}
	s.T.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = s.DB.Exec(cleanupCtx, `DELETE FROM rejected_row WHERE import_batch_id = $1`, batchID)
		_ = s.DB.Exec(cleanupCtx, `DELETE FROM event WHERE import_batch_id = $1`, batchID)
		_ = s.DB.Exec(cleanupCtx, `DELETE FROM job_queue WHERE payload LIKE '%' || $1 || '%'`, batchID)
		_ = s.DB.Exec(cleanupCtx, `DELETE FROM import_batch WHERE id = $1`, batchID)
	})
}
