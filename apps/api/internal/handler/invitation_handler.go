package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/MangriMen/UCRP/internal/domain"
	"github.com/MangriMen/UCRP/internal/platform/auth"
	"github.com/MangriMen/UCRP/internal/platform/httpx"
	"github.com/MangriMen/UCRP/internal/service"
)

type InvitationHandler struct{ invitations *service.InvitationService }

func NewInvitationHandler(invitations *service.InvitationService) *InvitationHandler {
	return &InvitationHandler{invitations: invitations}
}

func (h *InvitationHandler) Issue(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpx.WriteError(w, r, http.StatusUnauthorized, "unauthorized", "Нет пользователя")
		return
	}
	var request struct {
		Email       string      `json:"email"`
		DisplayName string      `json:"displayName"`
		Role        domain.Role `json:"role"`
		TenantID    string      `json:"tenantId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "validation_error", "Некорректное тело")
		return
	}
	err := h.invitations.Issue(r.Context(), principal, request.Email, request.DisplayName, request.Role, request.TenantID)
	if err != nil {
		invitationError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (h *InvitationHandler) Accept(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "validation_error", "Некорректное тело")
		return
	}
	if err := h.invitations.Accept(r.Context(), request.Token, request.Password); err != nil {
		invitationError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *InvitationHandler) Recover(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpx.WriteError(w, r, http.StatusUnauthorized, "unauthorized", "Нет пользователя")
		return
	}
	if err := h.invitations.IssueRecovery(r.Context(), principal, chi.URLParam(r, "id")); err != nil {
		invitationError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func invitationError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, service.ErrInvitationDeliveryUnavailable) {
		httpx.WriteError(w, r, http.StatusServiceUnavailable, "invitation_unavailable", "Доставка приглашений не настроена")
		return
	}
	httpx.WriteDomainError(w, r, err)
}
