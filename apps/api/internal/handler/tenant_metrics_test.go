package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestParseTenantFilter_RejectsUnknownSort(t *testing.T) {
	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/tenants?sort=bogus", nil)
	if _, problem := parseTenantFilter(r); problem == "" {
		t.Fatal("expected a validation problem")
	}
}

func TestParseTenantFilter_ReadsFilters(t *testing.T) {
	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/tenants?q=%20cafe%20&category_id=c1&floor=-1&is_active=false&period_from=2026-01-01&period_to=2026-01-31&sort=-visitsCount", nil)
	filter, problem := parseTenantFilter(r)
	if problem != "" {
		t.Fatal(problem)
	}
	if filter.Query != "cafe" || filter.CategoryID != "c1" || *filter.Floor != -1 || *filter.IsActive || filter.Sort != "-visitsCount" {
		t.Fatalf("filter = %+v", filter)
	}
	if !filter.PeriodTo.Equal(time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("period_to must be exclusive, got %v", filter.PeriodTo)
	}
}

func TestParseTenantFilter_RejectsInvertedPeriod(t *testing.T) {
	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/tenants?period_from=2026-02-01&period_to=2026-01-01", nil)
	if _, problem := parseTenantFilter(r); problem == "" {
		t.Fatal("expected a validation problem")
	}
}

func TestCombineMetricScopes_UsesStrictestAllowedScope(t *testing.T) {
	days, threshold := combineMetricScopes(
		metricScope{Allowed: true, PeriodDays: 90, Threshold: 5},
		metricScope{Allowed: true, PeriodDays: 30, Threshold: 8},
		metricScope{Allowed: false, PeriodDays: 7, Threshold: 20},
	)
	if days != 30 || threshold != 8 {
		t.Fatalf("days = %d threshold = %d", days, threshold)
	}
}

func TestClampPeriodStart_LimitsToGrantedWindow(t *testing.T) {
	now := time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC)
	early := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	if got := clampPeriodStart(&early, 30, now); !got.Equal(now.AddDate(0, 0, -30)) {
		t.Fatalf("got %v", got)
	}
	if got := clampPeriodStart(nil, 0, now); got != nil {
		t.Fatalf("got %v, want nil", got)
	}
}

func TestParseAggregateQuery_RequiresBothBoundsAndLimitsSpan(t *testing.T) {
	for name, path := range map[string]string{
		"only from": "/x?from=2026-01-01",
		"too long":  "/x?from=2024-01-01&to=2026-01-01",
		"inverted":  "/x?from=2026-02-01&to=2026-01-01",
	} {
		t.Run(name, func(t *testing.T) {
			if _, problem := parseAggregateQuery(httptest.NewRequestWithContext(context.Background(), http.MethodGet, path, nil), "t"); problem == "" {
				t.Fatal("expected a validation problem")
			}
		})
	}
	query, problem := parseAggregateQuery(httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/x", nil), "t")
	if problem != "" || query.From != nil {
		t.Fatalf("query = %+v problem = %q", query, problem)
	}
}
