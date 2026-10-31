package handler

import (
	"encoding/json"
	"net/http"

	"github.com/MangriMen/UCRP/internal/platform/auth"
	"github.com/MangriMen/UCRP/internal/platform/httpx"
	"github.com/MangriMen/UCRP/internal/service"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: authSvc}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	AccessToken string             `json:"access_token"`
	ExpiresIn   int64              `json:"expires_in"`
	User        service.PublicUser `json:"user"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" || req.Password == "" {
		httpx.WriteError(w, r, http.StatusBadRequest, "validation_failed", "Некорректное тело")
		return
	}
	res, err := h.auth.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	setRefreshCookie(w, res.RefreshToken)
	httpx.WriteJSON(w, http.StatusOK, loginResponse{
		AccessToken: res.AccessToken,
		ExpiresIn:   res.ExpiresIn,
		User:        res.User,
	})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpx.WriteError(w, r, http.StatusUnauthorized, "unauthorized", "Нет пользователя")
		return
	}
	me, err := h.auth.Me(r.Context(), p.UserID)
	if err != nil {
		httpx.WriteDomainError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, me)
}

func (h *AuthHandler) Healthz(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
