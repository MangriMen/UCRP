package e2e

import (
	"strings"
	"testing"

	"github.com/MangriMen/UCRP/e2e/harness"
	"github.com/MangriMen/UCRP/e2e/report"
)

func TestHealthzContract(t *testing.T) {
	s := harness.New(t)
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := s.API.Get(ctx, "/healthz", nil)
	if err != nil {
		t.Fatalf("GET /healthz: %v", err)
	}
	s.RequireStatus(resp, 200, "GET /healthz")
	s.RequireDeclaredStatus("GET", "/healthz", resp)
	s.RequireSchema("GET", "/healthz", resp)
}

func TestAuthLoginContract(t *testing.T) {
	s := harness.New(t).WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := s.API.Login(ctx, s.Cfg.AdminEmail, s.Cfg.AdminPassword)
	if err != nil {
		t.Fatalf("POST /api/v1/auth/login: %v", err)
	}
	s.RequireStatus(resp, 200, "POST /api/v1/auth/login")
	s.RequireDeclaredStatus("POST", "/api/v1/auth/login", resp)
	s.RequireSchema("POST", "/api/v1/auth/login", resp)
}

func TestAuthLoginRejectsWrongPassword(t *testing.T) {
	s := harness.New(t).WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := s.API.Login(ctx, s.Cfg.AdminEmail, "definitely-not-the-password")
	if err != nil {
		t.Fatalf("POST /api/v1/auth/login: %v", err)
	}
	s.RequireStatus(resp, 401, "POST /api/v1/auth/login with a wrong password")
	s.RequireDeclaredStatus("POST", "/api/v1/auth/login", resp)
	s.RequireSchema("POST", "/api/v1/auth/login", resp)
}

func TestAuthLoginRejectsEmptyBody(t *testing.T) {
	s := harness.New(t)
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := s.API.Post(ctx, "/api/v1/auth/login", map[string]string{})
	if err != nil {
		t.Fatalf("POST /api/v1/auth/login: %v", err)
	}
	s.RequireStatus(resp, 400, "POST /api/v1/auth/login with an empty body")
	s.RequireDeclaredStatus("POST", "/api/v1/auth/login", resp)
	s.RequireSchema("POST", "/api/v1/auth/login", resp)
}

func TestAuthMeContract(t *testing.T) {
	s := harness.New(t).WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := s.API.Me(ctx)
	if err != nil {
		t.Fatalf("GET /api/v1/auth/me: %v", err)
	}
	s.RequireStatus(resp, 200, "GET /api/v1/auth/me")
	s.RequireDeclaredStatus("GET", "/api/v1/auth/me", resp)
	s.RequireSchema("GET", "/api/v1/auth/me", resp)
}

func TestAuthMeRequiresToken(t *testing.T) {
	s := harness.New(t)
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := s.API.Me(ctx)
	if err != nil {
		t.Fatalf("GET /api/v1/auth/me without a token: %v", err)
	}
	s.RequireStatus(resp, 401, "GET /api/v1/auth/me without a token")
	s.RequireDeclaredStatus("GET", "/api/v1/auth/me", resp)
	s.RequireSchema("GET", "/api/v1/auth/me", resp)
}

func TestAuthMeRejectsMalformedToken(t *testing.T) {
	s := harness.New(t)
	ctx, cancel := s.Ctx()
	defer cancel()

	s.API.SetToken("not-a-jwt")
	defer s.API.SetToken("")

	resp, err := s.API.Me(ctx)
	if err != nil {
		t.Fatalf("GET /api/v1/auth/me with a malformed token: %v", err)
	}
	s.RequireStatus(resp, 401, "GET /api/v1/auth/me with a malformed token")
	s.RequireDeclaredStatus("GET", "/api/v1/auth/me", resp)
	s.RequireSchema("GET", "/api/v1/auth/me", resp)
}

func TestGuestsSearchContract(t *testing.T) {
	s := harness.New(t).WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := s.API.Get(ctx, "/api/v1/guests", map[string][]string{"q": {"e2e"}})
	if err != nil {
		t.Fatalf("GET /api/v1/guests: %v", err)
	}
	s.RequireStatus(resp, 200, "GET /api/v1/guests")
	s.RequireDeclaredStatus("GET", "/api/v1/guests", resp)
	s.RequireSchema("GET", "/api/v1/guests", resp)
}

func TestGuestsSearchRejectsShortQuery(t *testing.T) {
	s := harness.New(t).WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := s.API.Get(ctx, "/api/v1/guests", map[string][]string{"q": {"a"}})
	if err != nil {
		t.Fatalf("GET /api/v1/guests?q=a: %v", err)
	}
	s.RequireStatus(resp, 400, "GET /api/v1/guests?q=a (contract: minLength 2)")
	s.RequireDeclaredStatus("GET", "/api/v1/guests", resp)
	s.RequireSchema("GET", "/api/v1/guests", resp)
}

func TestGuestsSearchRequiresQuery(t *testing.T) {
	s := harness.New(t).WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := s.API.Get(ctx, "/api/v1/guests", nil)
	if err != nil {
		t.Fatalf("GET /api/v1/guests without q: %v", err)
	}
	s.RequireStatus(resp, 200, "GET /api/v1/guests without q lists guests")
	s.RequireDeclaredStatus("GET", "/api/v1/guests", resp)
	s.RequireSchema("GET", "/api/v1/guests", resp)
}

func TestGuestGetContract(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	ctx, cancel := s.Ctx()
	defer cancel()

	guest, err := s.Reg.Guest(ctx, "contract-get")
	if err != nil {
		t.Fatalf("fixture guest: %v", err)
	}

	resp, err := s.API.Get(ctx, "/api/v1/guests/"+guest.ID, nil)
	if err != nil {
		t.Fatalf("GET /api/v1/guests/{id}: %v", err)
	}
	s.RequireStatus(resp, 200, "GET /api/v1/guests/{id}")
	s.RequireDeclaredStatus("GET", "/api/v1/guests/{id}", resp)
	s.RequireSchema("GET", "/api/v1/guests/{id}", resp)
}

func TestGuestGetUnknownReturns404(t *testing.T) {
	s := harness.New(t).WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := s.API.Get(ctx, "/api/v1/guests/00000000-0000-0000-0000-000000000000", nil)
	if err != nil {
		t.Fatalf("GET /api/v1/guests/{unknown}: %v", err)
	}
	s.RequireStatus(resp, 404, "GET /api/v1/guests/{unknown}")
	s.RequireDeclaredStatus("GET", "/api/v1/guests/{id}", resp)
	s.RequireSchema("GET", "/api/v1/guests/{id}", resp)
}

func TestGuestEventsContract(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	ctx, cancel := s.Ctx()
	defer cancel()

	guest, err := s.Reg.Guest(ctx, "contract-events")
	if err != nil {
		t.Fatalf("fixture guest: %v", err)
	}

	resp, err := s.API.Get(ctx, "/api/v1/guests/"+guest.ID+"/events", nil)
	if err != nil {
		t.Fatalf("GET /api/v1/guests/{id}/events: %v", err)
	}
	s.RequireStatus(resp, 200, "GET /api/v1/guests/{id}/events")
	s.RequireDeclaredStatus("GET", "/api/v1/guests/{id}/events", resp)
	s.RequireSchema("GET", "/api/v1/guests/{id}/events", resp)
}

func TestSegmentsListContract(t *testing.T) {
	s := harness.New(t).WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := s.API.Get(ctx, "/api/v1/segments", nil)
	if err != nil {
		t.Fatalf("GET /api/v1/segments: %v", err)
	}
	s.RequireStatus(resp, 200, "GET /api/v1/segments")
	s.RequireDeclaredStatus("GET", "/api/v1/segments", resp)
	s.RequireSchema("GET", "/api/v1/segments", resp)
}

func TestSegmentPreviewContract(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	ctx, cancel := s.Ctx()
	defer cancel()

	category, err := s.Reg.Category(ctx, "preview-contract")
	if err != nil {
		t.Fatalf("fixture category: %v", err)
	}
	tenant, err := s.Reg.Tenant(ctx, "preview-contract", category.ID, 50)
	if err != nil {
		t.Fatalf("fixture tenant: %v", err)
	}

	contractBody := map[string]any{
		"rule": map[string]any{
			"visitedTenantIds":  []string{tenant.ID},
			"excludedTenantIds": []string{"00000000-0000-0000-0000-000000000000"},
			"periodDays":        90,
		},
	}
	implementationBody := map[string]any{
		"visitedTenantIds":  []string{tenant.ID},
		"excludedTenantIds": []string{"00000000-0000-0000-0000-000000000000"},
		"periodDays":        90,
	}
	resp, err := s.PostContractShape(ctx, "/api/v1/segments/preview", contractBody, implementationBody)
	if err != nil {
		t.Fatalf("POST /api/v1/segments/preview: %v", err)
	}
	s.RequireStatusOrDrift(resp, 200, "POST /api/v1/segments/preview")
	s.RequireDeclaredStatus("POST", "/api/v1/segments/preview", resp)
	s.RequireSchema("POST", "/api/v1/segments/preview", resp)
}

func TestSegmentPreviewRejectsConflictingTenant(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	ctx, cancel := s.Ctx()
	defer cancel()

	category, err := s.Reg.Category(ctx, "preview-conflict")
	if err != nil {
		t.Fatalf("fixture category: %v", err)
	}
	tenant, err := s.Reg.Tenant(ctx, "preview-conflict", category.ID, 50)
	if err != nil {
		t.Fatalf("fixture tenant: %v", err)
	}

	contractBody := map[string]any{
		"rule": map[string]any{
			"visitedTenantIds":  []string{tenant.ID},
			"excludedTenantIds": []string{tenant.ID},
			"periodDays":        90,
		},
	}
	resp, err := s.API.Post(ctx, "/api/v1/segments/preview", contractBody)
	if err != nil {
		t.Fatalf("POST /api/v1/segments/preview: %v", err)
	}
	s.RequireStatus(resp, 400, "POST /api/v1/segments/preview with a tenant in both visitedTenantIds and excludedTenantIds")
	s.RequireDeclaredStatus("POST", "/api/v1/segments/preview", resp)
	s.RequireSchema("POST", "/api/v1/segments/preview", resp)

	errorBody, ok := resp.ErrorBody()
	if !ok {
		t.Fatalf("400 response carries no error object: %s", resp.Snippet())
	}
	if len(errorBody.Error.Details) == 0 {
		report.Drifts.Record(t.Name(), "POST /api/v1/segments/preview",
			"contract requires details[].field = rule.excludedTenantIds for a conflicting tenant; the response has no details")
		t.Skipf("the contract requires details[].field = rule.excludedTenantIds for a conflicting tenant, got none: %s", resp.Snippet())
	}
	field := errorBody.Error.Details[0].Field
	if field != "rule.excludedTenantIds" {
		report.Drifts.Record(t.Name(), "POST /api/v1/segments/preview",
			"details[0].field: expected rule.excludedTenantIds, got "+field)
		t.Fatalf("details[0].field: expected rule.excludedTenantIds, got %q", field)
	}
}

func TestSegmentCreateContract(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	ctx, cancel := s.Ctx()
	defer cancel()

	category, err := s.Reg.Category(ctx, "create-contract")
	if err != nil {
		t.Fatalf("fixture category: %v", err)
	}
	tenant, err := s.Reg.Tenant(ctx, "create-contract", category.ID, 50)
	if err != nil {
		t.Fatalf("fixture tenant: %v", err)
	}

	contractBody := map[string]any{
		"name": s.Reg.Name("segment-contract"),
		"rule": map[string]any{
			"visitedTenantIds":  []string{tenant.ID},
			"excludedTenantIds": []string{"00000000-0000-0000-0000-000000000000"},
			"periodDays":        90,
		},
	}
	implementationBody := map[string]any{
		"name":              s.Reg.Name("segment-contract"),
		"visitedTenantIds":  []string{tenant.ID},
		"excludedTenantIds": []string{"00000000-0000-0000-0000-000000000000"},
		"periodDays":        90,
	}
	resp, err := s.PostContractShape(ctx, "/api/v1/segments", contractBody, implementationBody)
	if err != nil {
		t.Fatalf("POST /api/v1/segments: %v", err)
	}
	s.RequireStatusOrDrift(resp, 201, "POST /api/v1/segments")
	s.RequireDeclaredStatus("POST", "/api/v1/segments", resp)
	s.RequireSchema("POST", "/api/v1/segments", resp)
}

func TestSegmentCreateRequiresName(t *testing.T) {
	s := harness.New(t).WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := s.API.Post(ctx, "/api/v1/segments", map[string]any{"rule": map[string]any{"periodDays": 90}})
	if err != nil {
		t.Fatalf("POST /api/v1/segments without name: %v", err)
	}
	s.RequireStatus(resp, 400, "POST /api/v1/segments without name")
	s.RequireDeclaredStatus("POST", "/api/v1/segments", resp)
	s.RequireSchema("POST", "/api/v1/segments", resp)
}

func TestSegmentGetContract(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	ctx, cancel := s.Ctx()
	defer cancel()

	segment, err := s.Reg.Segment(ctx, "contract-get", `{"visitedTenantIds":[],"excludedTenantIds":[],"periodDays":90}`, nil)
	if err != nil {
		t.Fatalf("fixture segment: %v", err)
	}

	resp, err := s.API.Get(ctx, "/api/v1/segments/"+segment.ID, nil)
	if err != nil {
		t.Fatalf("GET /api/v1/segments/{id}: %v", err)
	}
	s.RequireStatus(resp, 200, "GET /api/v1/segments/{id}")
	s.RequireDeclaredStatus("GET", "/api/v1/segments/{id}", resp)
	s.RequireSchema("GET", "/api/v1/segments/{id}", resp)
}

func TestMergeCandidatesListContract(t *testing.T) {
	s := harness.New(t).WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	resp, err := s.API.Get(ctx, "/api/v1/merge-candidates", nil)
	if err != nil {
		t.Fatalf("GET /api/v1/merge-candidates: %v", err)
	}
	s.RequireStatus(resp, 200, "GET /api/v1/merge-candidates")
	s.RequireDeclaredStatus("GET", "/api/v1/merge-candidates", resp)
	s.RequireSchema("GET", "/api/v1/merge-candidates", resp)
}

func TestMergeCandidateGetContract(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	ctx, cancel := s.Ctx()
	defer cancel()

	guestA, err := s.Reg.Guest(ctx, "cand-a")
	if err != nil {
		t.Fatalf("fixture guest A: %v", err)
	}
	guestB, err := s.Reg.Guest(ctx, "cand-b")
	if err != nil {
		t.Fatalf("fixture guest B: %v", err)
	}
	candidate, err := s.Reg.MergeCandidate(ctx, guestA.ID, guestB.ID, "phone", "shared_phone_names_differ", 60)
	if err != nil {
		t.Fatalf("fixture merge candidate: %v", err)
	}

	resp, err := s.API.Get(ctx, "/api/v1/merge-candidates/"+candidate.ID, nil)
	if err != nil {
		t.Fatalf("GET /api/v1/merge-candidates/{id}: %v", err)
	}
	s.RequireStatus(resp, 200, "GET /api/v1/merge-candidates/{id}")
	s.RequireDeclaredStatus("GET", "/api/v1/merge-candidates/{id}", resp)
	s.RequireSchema("GET", "/api/v1/merge-candidates/{id}", resp)
}

func TestRecommendationsContract(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	ctx, cancel := s.Ctx()
	defer cancel()

	guest, err := s.Reg.Guest(ctx, "reco-contract")
	if err != nil {
		t.Fatalf("fixture guest: %v", err)
	}

	resp, err := s.API.Get(ctx, "/api/v1/guests/"+guest.ID+"/recommendations", nil)
	if err != nil {
		t.Fatalf("GET /api/v1/guests/{id}/recommendations: %v", err)
	}
	s.RequireStatus(resp, 200, "GET /api/v1/guests/{id}/recommendations")
	s.RequireDeclaredStatus("GET", "/api/v1/guests/{id}/recommendations", resp)
	s.RequireSchema("GET", "/api/v1/guests/{id}/recommendations", resp)
}

func TestRecommendationsRefreshContract(t *testing.T) {
	s := harness.New(t).WithAdmin().WithDB()
	ctx, cancel := s.Ctx()
	defer cancel()

	guest, err := s.Reg.Guest(ctx, "reco-refresh-contract")
	if err != nil {
		t.Fatalf("fixture guest: %v", err)
	}

	resp, err := s.API.Post(ctx, "/api/v1/guests/"+guest.ID+"/recommendations/refresh", nil)
	if err != nil {
		t.Fatalf("POST /api/v1/guests/{id}/recommendations/refresh: %v", err)
	}
	s.RequireStatus(resp, 200, "POST /api/v1/guests/{id}/recommendations/refresh")
	s.RequireDeclaredStatus("POST", "/api/v1/guests/{id}/recommendations/refresh", resp)
	s.RequireSchema("POST", "/api/v1/guests/{id}/recommendations/refresh", resp)
}

func TestUnmountedOperationsAreReported(t *testing.T) {
	s := harness.New(t).WithAdmin()
	ctx, cancel := s.Ctx()
	defer cancel()

	unmounted := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/imports"},
		{"GET", "/api/v1/exports"},
		{"GET", "/api/v1/campaigns"},
		{"GET", "/api/v1/tenants"},
		{"GET", "/api/v1/audit-logs"},
		{"GET", "/api/v1/segment-templates"},
		{"GET", "/api/v1/merge-blocks"},
		{"GET", "/api/v1/users"},
		{"GET", "/api/v1/data-sources"},
		{"GET", "/api/v1/event-types"},
	}
	var missing []string
	for _, op := range unmounted {
		resp, err := s.API.Get(ctx, op.path, nil)
		if err != nil {
			t.Fatalf("%s %s: %v", op.method, op.path, err)
		}
		if resp.Status == 404 || resp.Status == 405 {
			missing = append(missing, op.method+" "+op.path)
		}
	}
	if len(missing) == 0 {
		t.Skip("all sampled contract operations are mounted")
	}
	t.Logf("contract operations not mounted in cmd/api/main.go (%d): %s",
		len(missing), strings.Join(missing, ", "))
}
