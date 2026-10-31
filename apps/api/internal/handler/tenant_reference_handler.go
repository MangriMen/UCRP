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

type tenantReferenceReader interface {
	GetTenant(context.Context, string) (*service.TenantDTO, error)
}

type tenantSearcher interface {
	SearchTenants(context.Context, domain.TenantFilter, service.TenantStatsAccess, int, int) ([]service.TenantListItemDTO, int, error)
}

type TenantReferenceHandler struct {
	authorizer policyAuthorizer
	reference  tenantReferenceReader
	search     tenantSearcher
	legacy     *ReferenceHandler
	guests     *GuestHandler
}

type tenantOwnReference struct {
	ID         string `json:"id"`
	Code       string `json:"code"`
	Name       string `json:"name"`
	CategoryID string `json:"categoryId"`
	Kind       string `json:"kind"`
	Floor      *int   `json:"floor"`
	IsActive   bool   `json:"isActive"`
}

func NewTenantReferenceHandler(authorizer policyAuthorizer, reference tenantReferenceReader, search tenantSearcher, legacy *ReferenceHandler, guests *GuestHandler) *TenantReferenceHandler {
	return &TenantReferenceHandler{authorizer: authorizer, reference: reference, search: search, legacy: legacy, guests: guests}
}

func (h *TenantReferenceHandler) List(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpx.WriteError(w, r, http.StatusUnauthorized, "unauthorized", "Нет пользователя")
		return
	}
	if principal.Role != domain.RoleTenant {
		h.listForManagement(w, r, principal)
		return
	}
	item, ok := h.ownTenant(w, r, principal, principal.TenantID)
	if !ok {
		return
	}
	limit, offset := pageParams(r)
	items := make([]tenantOwnReference, 0, 1)
	if offset <= 0 {
		items = append(items, item)
	}
	writePage(w, items, limit, offset, 1)
}

func (h *TenantReferenceHandler) Get(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpx.WriteError(w, r, http.StatusUnauthorized, "unauthorized", "Нет пользователя")
		return
	}
	if principal.Role != domain.RoleTenant {
		h.legacy.GetTenant(w, r)
		return
	}
	item, ok := h.ownTenant(w, r, principal, chi.URLParam(r, "id"))
	if !ok {
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *TenantReferenceHandler) ownTenant(w http.ResponseWriter, r *http.Request, principal domain.Principal, tenantID string) (tenantOwnReference, bool) {
	if tenantID == "" || tenantID != principal.TenantID {
		httpx.WriteError(w, r, http.StatusForbidden, "forbidden_tenant", "Нет доступа")
		return tenantOwnReference{}, false
	}
	version, err := h.authorizer.PolicyVersion(r.Context())
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return tenantOwnReference{}, false
	}
	decision, err := h.authorizer.Authorize(r.Context(), principal, "reference.read", access.Resource{TenantID: tenantID})
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return tenantOwnReference{}, false
	}
	if !decision.Allowed || decision.PeriodDays > 0 || len(decision.MaskFields) > 0 {
		httpx.WriteError(w, r, http.StatusForbidden, "forbidden_permission", "Нет доступа")
		return tenantOwnReference{}, false
	}
	tenant, err := h.reference.GetTenant(r.Context(), tenantID)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return tenantOwnReference{}, false
	}
	if tenant == nil || tenant.ID != tenantID {
		httpx.WriteError(w, r, http.StatusForbidden, "forbidden_tenant", "Нет доступа")
		return tenantOwnReference{}, false
	}
	currentVersion, err := h.authorizer.PolicyVersion(r.Context())
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return tenantOwnReference{}, false
	}
	if version != currentVersion {
		httpx.WriteDomainError(w, r, domain.ErrConflict)
		return tenantOwnReference{}, false
	}
	return tenantOwnReference{
		ID: tenant.ID, Code: tenant.Code, Name: tenant.Name, CategoryID: tenant.CategoryID,
		Kind: tenant.Kind, Floor: tenant.Floor, IsActive: tenant.IsActive,
	}, true
}

func (h *TenantReferenceHandler) listForManagement(w http.ResponseWriter, r *http.Request, principal domain.Principal) {
	filter, problem := parseTenantFilter(r)
	if problem != "" {
		httpx.WriteError(w, r, http.StatusBadRequest, "validation_error", problem)
		return
	}
	visits, err := authorizeMetric(r.Context(), h.authorizer, principal, visitsMetric, "")
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	newGuests, err := authorizeMetric(r.Context(), h.authorizer, principal, newGuestsMetric, "")
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	switch strings.TrimPrefix(filter.Sort, "-") {
	case "guestsCount", "visitsCount":
		if !visits.Allowed {
			httpx.WriteError(w, r, http.StatusForbidden, "forbidden_permission", "Нет доступа к статистике визитов")
			return
		}
	case "newGuestsCount":
		if !newGuests.Allowed {
			httpx.WriteError(w, r, http.StatusForbidden, "forbidden_permission", "Нет доступа к статистике новых гостей")
			return
		}
	}
	periodDays, threshold := combineMetricScopes(visits, newGuests)
	filter.PeriodFrom = clampPeriodStart(filter.PeriodFrom, periodDays, time.Now().UTC())
	filter.KThreshold = threshold
	limit, offset := pageParams(r)
	if limit <= 0 {
		limit = defaultTenantPageSize
	}
	items, total, err := h.search.SearchTenants(r.Context(), filter, service.TenantStatsAccess{Visits: visits.Allowed, NewGuests: newGuests.Allowed}, limit, offset)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	writePage(w, items, limit, offset, total)
}

func (h *TenantReferenceHandler) Guests(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "id")
	if _, err := h.reference.GetTenant(r.Context(), tenantID); err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	h.guests.search(w, r, tenantID)
}
