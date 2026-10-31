package handler

import (
	"context"
	"net/http"
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
	resource := access.Resource{TenantID: tenantID, Metric: "metric.visits_count.read"}
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
	requestedMetrics := r.URL.Query()["metrics"]
	for _, requested := range requestedMetrics {
		for _, name := range strings.Split(requested, ",") {
			name = strings.TrimSpace(name)
			if name != "visits_count" && name != "metric.visits_count.read" {
				httpx.WriteError(w, r, http.StatusForbidden, "metric_not_granted", "Нет доступа к метрике")
				return
			}
		}
	}
	metric, err := h.authorizer.Authorize(r.Context(), principal, "metric.visits_count.read", resource)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	if !metric.Allowed {
		if len(requestedMetrics) > 0 {
			httpx.WriteError(w, r, http.StatusForbidden, "metric_not_granted", "Нет доступа к метрике")
			return
		}
		httpx.WriteJSON(w, http.StatusOK, map[string]any{})
		return
	}
	newGuests, err := authorizeMetric(r.Context(), h.authorizer, principal, newGuestsMetric, tenantID)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	period := action.PeriodDays
	if period == 0 || (metric.PeriodDays > 0 && metric.PeriodDays < period) {
		period = metric.PeriodDays
	}
	k := max(action.KThreshold, metric.KThreshold)
	if newGuests.Allowed {
		scopeDays, scopeK := combineMetricScopes(newGuests)
		k = max(k, scopeK)
		if scopeDays > 0 && (period == 0 || scopeDays < period) {
			period = scopeDays
		}
	}
	query.Threshold = k
	query.ShowNewGuests = newGuests.Allowed
	query.From = clampPeriodStart(query.From, period, time.Now().UTC())
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
