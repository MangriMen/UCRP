package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/MangriMen/UCRP/internal/platform/auth"
	"github.com/MangriMen/UCRP/internal/platform/httpx"
	"github.com/MangriMen/UCRP/internal/service"
)

type ContactLimitHandler struct {
	limits *service.ContactLimitService
}

func NewContactLimitHandler(limits *service.ContactLimitService) *ContactLimitHandler {
	return &ContactLimitHandler{limits: limits}
}

func (h *ContactLimitHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, offset := pageParams(r)
	items, err := h.limits.List(r.Context(), limit, offset)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *ContactLimitHandler) Create(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpx.WriteError(w, r, http.StatusUnauthorized, "unauthorized", "Нет пользователя")
		return
	}
	var req service.ContactLimitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "validation_failed", "Некорректное тело")
		return
	}
	item, err := h.limits.Create(r.Context(), req, p.UserID)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, item)
}

func (h *ContactLimitHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req service.ContactLimitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "validation_failed", "Некорректное тело")
		return
	}
	item, err := h.limits.Update(r.Context(), chi.URLParam(r, "id"), req)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *ContactLimitHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.limits.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type SegmentHandler struct {
	segments *service.SegmentService
}

func NewSegmentHandler(segments *service.SegmentService) *SegmentHandler {
	return &SegmentHandler{segments: segments}
}

func (h *SegmentHandler) Preview(w http.ResponseWriter, r *http.Request) {
	var req service.SegmentPreviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "validation_failed", "Некорректное тело")
		return
	}
	res, err := h.segments.Preview(r.Context(), req)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, res)
}

func (h *SegmentHandler) Create(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpx.WriteError(w, r, http.StatusUnauthorized, "unauthorized", "Нет пользователя")
		return
	}
	var req service.SegmentCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "validation_failed", "Некорректное тело")
		return
	}
	seg, err := h.segments.Create(r.Context(), req, p.UserID)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, seg)
}

func (h *SegmentHandler) Get(w http.ResponseWriter, r *http.Request) {
	seg, err := h.segments.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, seg)
}

func (h *SegmentHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, offset := pageParams(r)
	items, err := h.segments.List(r.Context(), limit, offset)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *SegmentHandler) Update(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpx.WriteError(w, r, http.StatusUnauthorized, "unauthorized", "Нет пользователя")
		return
	}
	var req service.SegmentCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "validation_failed", "Некорректное тело")
		return
	}
	seg, err := h.segments.Update(r.Context(), chi.URLParam(r, "id"), req, p.UserID)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, seg)
}

func (h *SegmentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.segments.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *SegmentHandler) Members(w http.ResponseWriter, r *http.Request) {
	limit, offset := pageParams(r)
	versionNo, _ := strconv.Atoi(r.URL.Query().Get("version"))
	items, err := h.segments.Members(r.Context(), chi.URLParam(r, "id"), versionNo, limit, offset)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *SegmentHandler) Versions(w http.ResponseWriter, r *http.Request) {
	items, err := h.segments.Versions(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *SegmentHandler) Recompute(w http.ResponseWriter, r *http.Request) {
	seg, err := h.segments.Recompute(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusAccepted, seg)
}

func (h *SegmentHandler) Templates(w http.ResponseWriter, r *http.Request) {
	items, err := h.segments.Templates(r.Context())
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *ExportHandler) Members(w http.ResponseWriter, r *http.Request) {
	var isControl *bool
	if v := r.URL.Query().Get("is_control"); v != "" {
		b := v == "true" || v == "1"
		isControl = &b
	}
	items, err := h.exports.Members(r.Context(), chi.URLParam(r, "id"), isControl)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *GuestHandler) Contacts(w http.ResponseWriter, r *http.Request) {
	limit, offset := pageParams(r)
	items, err := h.profiles.Contacts(r.Context(), chi.URLParam(r, "id"), limit, offset)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *GuestHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.profiles.SoftDelete(r.Context(), chi.URLParam(r, "id")); err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type ReferenceHandler struct {
	ref *service.ReferenceService
}

func NewReferenceHandler(ref *service.ReferenceService) *ReferenceHandler {
	return &ReferenceHandler{ref: ref}
}

func decode[T any](w http.ResponseWriter, r *http.Request) (T, bool) {
	var req T
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "validation_failed", "Некорректное тело")
		return req, false
	}
	return req, true
}

func (h *ReferenceHandler) ListTenants(w http.ResponseWriter, r *http.Request) {
	limit, offset := pageParams(r)
	items, err := h.ref.ListTenants(r.Context(), limit, offset)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *ReferenceHandler) GetTenant(w http.ResponseWriter, r *http.Request) {
	item, err := h.ref.GetTenant(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *ReferenceHandler) CreateTenant(w http.ResponseWriter, r *http.Request) {
	req, ok := decode[service.TenantCreateRequest](w, r)
	if !ok {
		return
	}
	item, err := h.ref.CreateTenant(r.Context(), req)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, item)
}

func (h *ReferenceHandler) UpdateTenant(w http.ResponseWriter, r *http.Request) {
	req, ok := decode[service.TenantUpdateRequest](w, r)
	if !ok {
		return
	}
	item, err := h.ref.UpdateTenant(r.Context(), chi.URLParam(r, "id"), req)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *ReferenceHandler) DeleteTenant(w http.ResponseWriter, r *http.Request) {
	if err := h.ref.DeleteTenant(r.Context(), chi.URLParam(r, "id")); err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ReferenceHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	items, err := h.ref.ListCategories(r.Context())
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *ReferenceHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	req, ok := decode[service.CategoryRequest](w, r)
	if !ok {
		return
	}
	item, err := h.ref.CreateCategory(r.Context(), req)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, item)
}

func (h *ReferenceHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	req, ok := decode[service.CategoryRequest](w, r)
	if !ok {
		return
	}
	item, err := h.ref.UpdateCategory(r.Context(), chi.URLParam(r, "id"), req)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *ReferenceHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	if err := h.ref.DeleteCategory(r.Context(), chi.URLParam(r, "id")); err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ReferenceHandler) ListRelations(w http.ResponseWriter, r *http.Request) {
	items, err := h.ref.ListRelations(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *ReferenceHandler) ReplaceRelations(w http.ResponseWriter, r *http.Request) {
	req, ok := decode[struct {
		Items []service.RelationDTO `json:"items"`
	}](w, r)
	if !ok {
		return
	}
	items, err := h.ref.ReplaceRelations(r.Context(), chi.URLParam(r, "id"), req.Items)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *ReferenceHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	limit, offset := pageParams(r)
	items, err := h.ref.ListProjects(r.Context(), r.URL.Query().Get("tenant_id"), limit, offset)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *ReferenceHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	req, ok := decode[service.ProjectRequest](w, r)
	if !ok {
		return
	}
	item, err := h.ref.CreateProject(r.Context(), req)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, item)
}

func (h *ReferenceHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	req, ok := decode[service.ProjectRequest](w, r)
	if !ok {
		return
	}
	item, err := h.ref.UpdateProject(r.Context(), chi.URLParam(r, "id"), req)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *ReferenceHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	if err := h.ref.DeleteProject(r.Context(), chi.URLParam(r, "id")); err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ReferenceHandler) ListEventTypes(w http.ResponseWriter, r *http.Request) {
	items, err := h.ref.ListEventTypes(r.Context())
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *ReferenceHandler) CreateEventType(w http.ResponseWriter, r *http.Request) {
	req, ok := decode[service.EventTypeRequest](w, r)
	if !ok {
		return
	}
	item, err := h.ref.CreateEventType(r.Context(), req)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, item)
}

func (h *ReferenceHandler) UpdateEventType(w http.ResponseWriter, r *http.Request) {
	req, ok := decode[service.EventTypeRequest](w, r)
	if !ok {
		return
	}
	item, err := h.ref.UpdateEventType(r.Context(), chi.URLParam(r, "id"), req)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *ReferenceHandler) DeleteEventType(w http.ResponseWriter, r *http.Request) {
	if err := h.ref.DeleteEventType(r.Context(), chi.URLParam(r, "id")); err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
