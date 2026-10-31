package auth

import (
	"context"
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
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

type Manager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewManager(secret string, accessTTL time.Duration) *Manager {
	return &Manager{secret: []byte(secret), accessTTL: accessTTL, refreshTTL: 7 * 24 * time.Hour}
}

func (m *Manager) Issue(p domain.Principal) (TokenPair, error) {
	access, err := m.sign(p, m.accessTTL, false)
	if err != nil {
		return TokenPair{}, err
	}
	refresh, err := m.sign(p, m.refreshTTL, true)
	if err != nil {
		return TokenPair{}, err
	}
	return TokenPair{AccessToken: access, RefreshToken: refresh, ExpiresIn: int64(m.accessTTL.Seconds())}, nil
}

func (m *Manager) sign(p domain.Principal, ttl time.Duration, refresh bool) (string, error) {
	now := time.Now().UTC()
	claims := jwt.MapClaims{
		"sub":  p.UserID,
		"role": string(p.Role),
		"exp":  now.Add(ttl).Unix(),
		"iat":  now.Unix(),
	}
	if refresh {
		claims["typ"] = "refresh"
	}
	if p.TenantID != "" {
		claims["tenant_id"] = p.TenantID
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(m.secret)
}

func (m *Manager) Parse(token string) (domain.Principal, error) {
	return m.parse(token, false)
}

func (m *Manager) ParseRefresh(token string) (domain.Principal, error) {
	return m.parse(token, true)
}

func (m *Manager) parse(token string, refresh bool) (domain.Principal, error) {
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
	typ, _ := claims["typ"].(string)
	if refresh && typ != "refresh" {
		return domain.Principal{}, domain.ErrUnauthorized
	}
	if !refresh && typ == "refresh" {
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
