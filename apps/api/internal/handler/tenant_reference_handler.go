package handler

import (
	"context"
	"net/http"

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

type TenantReferenceHandler struct {
	authorizer policyAuthorizer
	reference  tenantReferenceReader
	legacy     *ReferenceHandler
}

type tenantOwnReference struct {
	ID         string `json:"id"`
	Code       string `json:"code"`
	Name       string `json:"name"`
	CategoryID string `json:"category_id"`
	Kind       string `json:"kind"`
	Floor      int    `json:"floor"`
	IsActive   bool   `json:"is_active"`
}

func NewTenantReferenceHandler(authorizer policyAuthorizer, reference tenantReferenceReader, legacy *ReferenceHandler) *TenantReferenceHandler {
	return &TenantReferenceHandler{authorizer: authorizer, reference: reference, legacy: legacy}
}

func (h *TenantReferenceHandler) List(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpx.WriteError(w, r, http.StatusUnauthorized, "unauthorized", "Нет пользователя")
		return
	}
	if principal.Role != domain.RoleTenant {
		h.legacy.ListTenants(w, r)
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
