package handler

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/MangriMen/UCRP/internal/access"
	"github.com/MangriMen/UCRP/internal/domain"
)

const (
	visitsMetric          = "metric.visits_count.read"
	newGuestsMetric       = "metric.new_guests_via_mall.read"
	repeatVisitsMetric    = "metric.repeat_visits.read"
	revenueMetric         = "metric.revenue_total.read"
	avgCheckMetric        = "metric.avg_check.read"
	firstPurchaseMetric   = "metric.first_purchase_new_tenant.read"
	minSuppressionK       = 5
	defaultTenantPageSize = 50
)

var tenantListSortFields = map[string]bool{
	"name": true, "-name": true,
	"floor": true, "-floor": true,
	"priority": true, "-priority": true,
	"createdAt": true, "-createdAt": true,
	"guestsCount": true, "-guestsCount": true,
	"visitsCount": true, "-visitsCount": true,
	"newGuestsCount": true, "-newGuestsCount": true,
}

type metricScope struct {
	Allowed    bool
	PeriodDays int
	Threshold  int
}

func authorizeMetric(ctx context.Context, authorizer accessAuthorizer, principal domain.Principal, metric, tenantID string) (metricScope, error) {
	decision, err := authorizer.Authorize(ctx, principal, metric, access.Resource{TenantID: tenantID, Metric: metric})
	if err != nil {
		return metricScope{}, err
	}
	return metricScope{
		Allowed:    decision.Allowed && len(decision.MaskFields) == 0,
		PeriodDays: decision.PeriodDays,
		Threshold:  max(decision.KThreshold, minSuppressionK),
	}, nil
}

func combineMetricScopes(scopes ...metricScope) (periodDays, threshold int) {
	threshold = minSuppressionK
	for _, scope := range scopes {
		if !scope.Allowed {
			continue
		}
		threshold = max(threshold, scope.Threshold)
		if scope.PeriodDays > 0 && (periodDays == 0 || scope.PeriodDays < periodDays) {
			periodDays = scope.PeriodDays
		}
	}
	return periodDays, threshold
}

func clampPeriodStart(from *time.Time, periodDays int, now time.Time) *time.Time {
	if periodDays <= 0 {
		return from
	}
	earliest := now.AddDate(0, 0, -periodDays)
	if from == nil || from.Before(earliest) {
		return &earliest
	}
	return from
}

func parseTenantFilter(r *http.Request) (domain.TenantFilter, string) {
	query := r.URL.Query()
	filter := domain.TenantFilter{
		Query:      strings.TrimSpace(query.Get("q")),
		MallID:     strings.TrimSpace(query.Get("mall_id")),
		CategoryID: strings.TrimSpace(query.Get("category_id")),
		Sort:       query.Get("sort"),
	}
	if filter.Sort != "" && !tenantListSortFields[filter.Sort] {
		return filter, "sort: допустимы name, floor, priority, createdAt, guestsCount, visitsCount, newGuestsCount с префиксом - для убывания"
	}
	if raw := query.Get("floor"); raw != "" {
		floor, err := strconv.Atoi(raw)
		if err != nil {
			return filter, "floor: ожидается целое число"
		}
		filter.Floor = &floor
	}
	if raw := query.Get("is_active"); raw != "" {
		active, err := strconv.ParseBool(raw)
		if err != nil {
			return filter, "is_active: ожидается true или false"
		}
		filter.IsActive = &active
	}
	from, problem := parseDateParam(query, "period_from", false)
	if problem != "" {
		return filter, problem
	}
	to, problem := parseDateParam(query, "period_to", true)
	if problem != "" {
		return filter, problem
	}
	if from != nil && to != nil && !from.Before(*to) {
		return filter, "period_from не может быть позже period_to"
	}
	filter.PeriodFrom, filter.PeriodTo = from, to
	return filter, ""
}
