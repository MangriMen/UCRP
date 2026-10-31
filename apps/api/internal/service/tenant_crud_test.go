package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/MangriMen/UCRP/internal/domain"
)

type stubTenantRepo struct {
	tenant   domain.Tenant
	inserted *domain.Tenant
	entries  []domain.TenantListEntry
}

func (r *stubTenantRepo) Search(context.Context, domain.TenantFilter, int, int) ([]domain.TenantListEntry, int, error) {
	return r.entries, len(r.entries), nil
}

func (r *stubTenantRepo) GetByID(context.Context, string) (*domain.Tenant, error) {
	t := r.tenant
	return &t, nil
}

func (r *stubTenantRepo) Insert(_ context.Context, t *domain.Tenant) error {
	r.inserted = t
	return nil
}

func (r *stubTenantRepo) Update(context.Context, *domain.Tenant) error { return nil }

func (r *stubTenantRepo) SoftDelete(context.Context, string, time.Time) error { return nil }

type stubCategoryRepo struct {
	domain.TenantCategoryRepo
	categories map[string]domain.TenantCategory
}

func (r stubCategoryRepo) GetByID(_ context.Context, id string) (*domain.TenantCategory, error) {
	c, ok := r.categories[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return &c, nil
}

type stubMallRepo struct{ domain.MallRepo }

func (stubMallRepo) GetByID(context.Context, string) (*domain.Mall, error) {
	return &domain.Mall{FloorMin: 1, FloorMax: 5}, nil
}

func newTenantService(repo *stubTenantRepo) *ReferenceService {
	categories := stubCategoryRepo{categories: map[string]domain.TenantCategory{
		"active":   {ID: "active", IsActive: true},
		"inactive": {ID: "inactive", IsActive: false},
	}}
	return NewReferenceService(repo, stubMallRepo{}, categories, nil, nil)
}

func TestCreateTenant_RejectsInvalidInput(t *testing.T) {
	priority := 101
	tests := []struct {
		name string
		req  TenantCreateRequest
	}{
		{"unknown category", TenantCreateRequest{Code: "a", Name: "A", MallID: "m", CategoryID: "missing"}},
		{"inactive category", TenantCreateRequest{Code: "a", Name: "A", MallID: "m", CategoryID: "inactive"}},
		{"unknown kind", TenantCreateRequest{Code: "a", Name: "A", MallID: "m", CategoryID: "active", Kind: "kiosk"}},
		{"priority above range", TenantCreateRequest{Code: "a", Name: "A", MallID: "m", CategoryID: "active", Priority: &priority}},
		{"code too long", TenantCreateRequest{Code: strings.Repeat("x", 65), Name: "A", MallID: "m", CategoryID: "active"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &stubTenantRepo{}
			_, err := newTenantService(repo).CreateTenant(context.Background(), tt.req)
			if !errors.Is(err, domain.ErrValidation) {
				t.Fatalf("err = %v, want validation", err)
			}
			if repo.inserted != nil {
				t.Fatal("tenant inserted despite invalid input")
			}
		})
	}
}

func TestCreateTenant_KeepsZeroPriority(t *testing.T) {
	zero := 0
	repo := &stubTenantRepo{}
	_, err := newTenantService(repo).CreateTenant(context.Background(), TenantCreateRequest{
		Code: "a", Name: "A", MallID: "m", CategoryID: "active", Priority: &zero,
	})
	if err != nil {
		t.Fatal(err)
	}
	if repo.inserted.Priority != 0 {
		t.Fatalf("priority = %d, want 0", repo.inserted.Priority)
	}
}

func TestCreateTenant_DefaultsPriority_WhenOmitted(t *testing.T) {
	repo := &stubTenantRepo{}
	_, err := newTenantService(repo).CreateTenant(context.Background(), TenantCreateRequest{
		Code: "a", Name: "A", MallID: "m", CategoryID: "active",
	})
	if err != nil {
		t.Fatal(err)
	}
	if repo.inserted.Priority != defaultTenantPriority {
		t.Fatalf("priority = %d, want %d", repo.inserted.Priority, defaultTenantPriority)
	}
}

func TestUpdateTenant_RejectsInactiveCategoryAndBlankName(t *testing.T) {
	repo := &stubTenantRepo{tenant: domain.Tenant{ID: "t", CategoryID: "active", MallID: "m", Name: "A", Kind: "mall"}}
	svc := newTenantService(repo)
	if _, err := svc.UpdateTenant(context.Background(), "t", TenantUpdateRequest{CategoryID: "inactive"}); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("inactive category err = %v", err)
	}
	if _, err := svc.UpdateTenant(context.Background(), "t", TenantUpdateRequest{Name: "   "}); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("blank name err = %v", err)
	}
}

func TestSearchTenants_HidesStatisticsWithoutMetricAccess(t *testing.T) {
	visits, fresh := 12, 7
	repo := &stubTenantRepo{entries: []domain.TenantListEntry{{
		Tenant: domain.Tenant{ID: "t"},
		Stats:  domain.TenantStats{GuestsCount: &visits, VisitsCount: &visits, NewGuestsCount: &fresh},
	}}}
	items, _, err := newTenantService(repo).SearchTenants(context.Background(), domain.TenantFilter{}, TenantStatsAccess{NewGuests: true}, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	item := items[0]
	if item.VisitsCount != nil || item.GuestsCount != nil {
		t.Fatal("visit statistics leaked without metric.visits_count.read")
	}
	if item.NewGuestsCount == nil || item.Suppressed {
		t.Fatalf("new guests = %v suppressed = %v", item.NewGuestsCount, item.Suppressed)
	}
}

func TestSearchTenants_MarksSuppressed_WhenGroupIsSmall(t *testing.T) {
	repo := &stubTenantRepo{entries: []domain.TenantListEntry{{Tenant: domain.Tenant{ID: "t"}}}}
	items, _, err := newTenantService(repo).SearchTenants(context.Background(), domain.TenantFilter{}, TenantStatsAccess{Visits: true, NewGuests: true}, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !items[0].Suppressed || items[0].VisitsCount != nil {
		t.Fatalf("item = %+v", items[0])
	}
}
