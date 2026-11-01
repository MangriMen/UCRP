package httpx

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/MangriMen/UCRP/internal/domain"
	"github.com/google/uuid"
)

type ctxKey string

const requestIDKey ctxKey = "request_id"

type ErrorBody struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code      string       `json:"code"`
	Message   string       `json:"message"`
	Details   []ErrorField `json:"details,omitempty"`
	RequestID string       `json:"request_id"`
}

type ErrorField struct {
	Field  string `json:"field"`
	Reason string `json:"reason"`
}

func RequestID(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey).(string); ok {
		return v
	}
	return ""
}

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-Id")
		if id == "" {
			id = uuid.NewString()
		}
		w.Header().Set("X-Request-Id", id)
		next.ServeHTTP(w, r.WithContext(WithRequestID(r.Context(), id)))
	})
}

func TimeoutMiddleware(d time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), d)
			defer cancel()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RecoverMiddleware(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Error("panic", "err", rec, "request_id", RequestID(r.Context()))
					WriteError(w, r, http.StatusInternalServerError, "internal_error", "Внутренняя ошибка")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func WriteError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	WriteJSON(w, status, ErrorBody{Error: ErrorDetail{
		Code:      code,
		Message:   message,
		RequestID: RequestID(r.Context()),
	}})
}

func WriteDomainError(w http.ResponseWriter, r *http.Request, err error) {
	var fieldErr *domain.FieldError
	if errors.As(err, &fieldErr) {
		WriteJSON(w, http.StatusBadRequest, ErrorBody{Error: ErrorDetail{
			Code:      "validation_error",
			Message:   "Некорректные данные",
			Details:   []ErrorField{{Field: fieldErr.Field, Reason: fieldErr.Reason}},
			RequestID: RequestID(r.Context()),
		}})
		return
	}
	switch {
	case errors.Is(err, domain.ErrNotFound):
		WriteError(w, r, http.StatusNotFound, "not_found", "Не найдено")
	case errors.Is(err, domain.ErrConflict):
		WriteError(w, r, http.StatusConflict, "conflict", "Конфликт состояния")
	case errors.Is(err, domain.ErrForbidden):
		WriteError(w, r, http.StatusForbidden, "forbidden_role", "Нет прав")
	case errors.Is(err, domain.ErrUnauthorized):
		WriteError(w, r, http.StatusUnauthorized, "unauthorized", "Нет доступа")
	case errors.Is(err, domain.ErrTimeout), errors.Is(err, context.DeadlineExceeded):
		WriteError(w, r, http.StatusGatewayTimeout, "timeout", "Превышен таймаут")
	case errors.Is(err, domain.ErrValidation):
		WriteError(w, r, http.StatusBadRequest, "validation_error", "Некорректные данные")
	default:
		WriteError(w, r, http.StatusInternalServerError, "internal_error", "Внутренняя ошибка")
	}
}
