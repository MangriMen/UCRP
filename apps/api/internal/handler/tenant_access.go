package handler

import (
	"context"
	"net/http"
	"strings"

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
	period := action.PeriodDays
	if period == 0 || (metric.PeriodDays > 0 && metric.PeriodDays < period) {
		period = metric.PeriodDays
	}
	k := action.KThreshold
	if metric.KThreshold > k {
		k = metric.KThreshold
	}
	result, err := h.analytics.TenantVisitAggregate(r.Context(), tenantID, period, k)
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
