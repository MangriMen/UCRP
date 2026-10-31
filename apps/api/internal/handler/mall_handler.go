package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/MangriMen/UCRP/internal/platform/httpx"
	"github.com/MangriMen/UCRP/internal/service"
)

func (h *ReferenceHandler) ListMalls(w http.ResponseWriter, r *http.Request) {
	items, err := h.ref.ListMalls(r.Context())
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, nonNil(items))
}

func (h *ReferenceHandler) CreateMall(w http.ResponseWriter, r *http.Request) {
	req, ok := decode[service.MallRequest](w, r)
	if !ok {
		return
	}
	item, err := h.ref.CreateMall(r.Context(), req)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, item)
}

func (h *ReferenceHandler) UpdateMall(w http.ResponseWriter, r *http.Request) {
	req, ok := decode[service.MallRequest](w, r)
	if !ok {
		return
	}
	item, err := h.ref.UpdateMall(r.Context(), chi.URLParam(r, "id"), req)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *ReferenceHandler) DeleteMall(w http.ResponseWriter, r *http.Request) {
	if err := h.ref.DeleteMall(r.Context(), chi.URLParam(r, "id")); err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
