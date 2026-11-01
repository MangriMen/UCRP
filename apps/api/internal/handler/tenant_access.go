package handler

import (
	"context"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/MangriMen/UCRP/internal/access"
	"github.com/MangriMen/UCRP/internal/domain"
	"github.com/MangriMen/UCRP/internal/platform/auth"
	"github.com/MangriMen/UCRP/internal/platform/httpx"
	"github.com/MangriMen/UCRP/internal/service"
)

type accessAuthorizer interface {
	Authorize(context.Context, domain.Principal, string, access.Resource) (access.Decision, error)
}

type TenantAccessHandler struct {
	authorizer policyAuthorizer
	analytics  *service.AnalyticsService
}

func NewTenantAccessHandler(authorizer policyAuthorizer, analytics *service.AnalyticsService) *TenantAccessHandler {
	return &TenantAccessHandler{authorizer: authorizer, analytics: analytics}
}

func (h *TenantAccessHandler) Aggregates(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpx.WriteError(w, r, http.StatusUnauthorized, "unauthorized", "Нет пользователя")
		return
	}
	tenantID := chi.URLParam(r, "id")
	version, err := h.authorizer.PolicyVersion(r.Context())
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	resource := access.Resource{TenantID: tenantID, Metric: visitsMetric}
	permission := "tenant.aggregates.read"
	if principal.Role == domain.RoleTenant {
		permission = "tenant.aggregates.own"
	}
	action, err := h.authorizer.Authorize(r.Context(), principal, permission, resource)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	if !action.Allowed {
		httpx.WriteError(w, r, http.StatusForbidden, "forbidden", "Нет доступа")
		return
	}
	query, problem := parseAggregateQuery(r, tenantID)
	if problem != "" {
		httpx.WriteError(w, r, http.StatusBadRequest, "validation_error", problem)
		return
	}
	requested, known := requestedTenantMetrics(r)
	if !known {
		httpx.WriteError(w, r, http.StatusForbidden, "metric_not_granted", "Нет доступа к метрике")
		return
	}
	visits, err := authorizeMetric(r.Context(), h.authorizer, principal, visitsMetric, tenantID)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	if !visits.Allowed {
		if len(requested) > 0 {
			httpx.WriteError(w, r, http.StatusForbidden, "metric_not_granted", "Нет доступа к метрике")
			return
		}
		httpx.WriteJSON(w, http.StatusOK, map[string]any{})
		return
	}
	scopes := []metricScope{
		{Allowed: true, PeriodDays: action.PeriodDays, Threshold: max(action.KThreshold, minSuppressionK)},
		visits,
	}
	for _, metric := range optionalTenantMetrics {
		scope, err := authorizeMetric(r.Context(), h.authorizer, principal, metric, tenantID)
		if err != nil {
			httpx.WriteDomainError(w, r, err)
			return
		}
		if requested[metric] && !scope.Allowed {
			httpx.WriteError(w, r, http.StatusForbidden, "metric_not_granted", "Нет доступа к метрике")
			return
		}
		grantTenantMetric(&query.Metrics, metric, scope.Allowed)
		scopes = append(scopes, scope)
	}
	periodDays, threshold := combineMetricScopes(scopes...)
	query.Threshold = threshold
	query.From = clampPeriodStart(query.From, periodDays, time.Now().UTC())
	result, err := h.analytics.TenantAggregate(r.Context(), query)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	currentVersion, err := h.authorizer.PolicyVersion(r.Context())
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	if version != currentVersion {
		httpx.WriteDomainError(w, r, domain.ErrConflict)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result)
}

const maxAggregatePeriodDays = 366

var optionalTenantMetrics = []string{newGuestsMetric, repeatVisitsMetric, revenueMetric, avgCheckMetric, firstPurchaseMetric}

func grantTenantMetric(metrics *service.TenantAggregateMetrics, metric string, allowed bool) {
	switch metric {
	case newGuestsMetric:
		metrics.NewGuests = allowed
	case repeatVisitsMetric:
		metrics.RepeatVisits = allowed
	case revenueMetric:
		metrics.Revenue = allowed
	case avgCheckMetric:
		metrics.AvgCheck = allowed
	case firstPurchaseMetric:
		metrics.FirstPurchases = allowed
	}
}

func requestedTenantMetrics(r *http.Request) (map[string]bool, bool) {
	requested := map[string]bool{}
	for _, value := range r.URL.Query()["metrics"] {
		for _, name := range strings.Split(value, ",") {
			permission := strings.TrimSpace(name)
			if !strings.HasPrefix(permission, "metric.") {
				permission = "metric." + permission + ".read"
			}
			if permission != visitsMetric && !slices.Contains(optionalTenantMetrics, permission) {
				return nil, false
			}
			requested[permission] = true
		}
	}
	return requested, true
}

func parseAggregateQuery(r *http.Request, tenantID string) (service.TenantAggregateQuery, string) {
	query := service.TenantAggregateQuery{TenantID: tenantID}
	values := r.URL.Query()
	from, problem := parseDateParam(values, "from", false)
	if problem != "" {
		return query, problem
	}
	to, problem := parseDateParam(values, "to", true)
	if problem != "" {
		return query, problem
	}
	if (from == nil) != (to == nil) {
		return query, "from и to задаются вместе"
	}
	if from != nil {
		if !from.Before(*to) {
			return query, "from не может быть позже to"
		}
		if to.Sub(*from) > maxAggregatePeriodDays*24*time.Hour {
			return query, "период не может быть длиннее 366 дней"
		}
	}
	query.From, query.To = from, to
	return query, ""
}
