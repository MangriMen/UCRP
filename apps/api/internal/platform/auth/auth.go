package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/MangriMen/UCRP/internal/domain"
	"github.com/MangriMen/UCRP/internal/platform/httpx"
)

type ctxKey string

const principalKey ctxKey = "principal"

type TokenPair struct {
	AccessToken string
	ExpiresIn   int64
}

type Manager struct {
	secret    []byte
	accessTTL time.Duration
}

func NewManager(secret string, accessTTL time.Duration) *Manager {
	return &Manager{secret: []byte(secret), accessTTL: accessTTL}
}

func (m *Manager) Issue(p domain.Principal) (TokenPair, error) {
	now := time.Now().UTC()
	exp := now.Add(m.accessTTL)
	claims := jwt.MapClaims{
		"sub":  p.UserID,
		"role": string(p.Role),
		"exp":  exp.Unix(),
		"iat":  now.Unix(),
	}
	if p.TenantID != "" {
		claims["tenant_id"] = p.TenantID
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := t.SignedString(m.secret)
	if err != nil {
		return TokenPair{}, err
	}
	return TokenPair{AccessToken: signed, ExpiresIn: int64(m.accessTTL.Seconds())}, nil
}

func (m *Manager) Parse(token string) (domain.Principal, error) {
	claims := jwt.MapClaims{}
	parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return m.secret, nil
	})
	if err != nil || !parsed.Valid {
		return domain.Principal{}, domain.ErrUnauthorized
	}
	sub, _ := claims["sub"].(string)
	role, _ := claims["role"].(string)
	tenantID, _ := claims["tenant_id"].(string)
	if sub == "" || role == "" {
		return domain.Principal{}, domain.ErrUnauthorized
	}
	return domain.Principal{
		UserID:   sub,
		Role:     domain.Role(role),
		TenantID: tenantID,
	}, nil
}

func HashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func PrincipalFrom(ctx context.Context) (domain.Principal, bool) {
	p, ok := ctx.Value(principalKey).(domain.Principal)
	return p, ok
}

func Middleware(m *Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				httpx.WriteError(w, r, http.StatusUnauthorized, "unauthorized", "Нет токена")
				return
			}
			token := strings.TrimPrefix(header, "Bearer ")
			p, err := m.Parse(token)
			if err != nil {
				httpx.WriteError(w, r, http.StatusUnauthorized, "invalid_credentials", "Неверный токен")
				return
			}
			ctx := context.WithValue(r.Context(), principalKey, p)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRole(roles ...domain.Role) func(http.Handler) http.Handler {
	allowed := make(map[domain.Role]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p, ok := PrincipalFrom(r.Context())
			if !ok {
				httpx.WriteError(w, r, http.StatusUnauthorized, "unauthorized", "Нет пользователя")
				return
			}
			if _, ok := allowed[p.Role]; !ok {
				httpx.WriteError(w, r, http.StatusForbidden, "forbidden_role", "Роль не допущена")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func BearerToken(r *http.Request) (string, error) {
	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return "", errors.New("missing bearer token")
	}
	return strings.TrimPrefix(header, "Bearer "), nil
}
