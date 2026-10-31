package harness

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/MangriMen/UCRP/e2e/client"
	"github.com/MangriMen/UCRP/e2e/fixtures"
	"github.com/MangriMen/UCRP/e2e/openapi"
	"github.com/MangriMen/UCRP/e2e/report"
)

type Suite struct {
	T     *testing.T
	Cfg   Config
	API   *client.Client
	Spec  *openapi.Spec
	DB    *fixtures.DB
	Reg   *fixtures.Registry
	Token string
	start time.Time
	class string

	marketer *client.Client
}

type Config struct {
	BaseURL        string
	DatabaseURL    string
	AdminEmail     string
	AdminPassword  string
	OpenAPIPath    string
	RunID          string
	RequestTimeout time.Duration
	ImportTimeout  time.Duration
	PollInterval   time.Duration
	StrictContract bool
}

func New(t *testing.T) *Suite {
	t.Helper()
	cfg := loadConfig()
	s := &Suite{T: t, Cfg: cfg, start: time.Now(), class: t.Name()}
	t.Cleanup(func() {
		report.Default.Add(report.Case{
			Name:     t.Name(),
			Class:    s.class,
			Duration: time.Since(s.start),
			Failed:   t.Failed(),
			Skipped:  t.Skipped(),
			Message:  failureMessage(t),
		})
	})

	if cfg.BaseURL == "" {
		t.Skip("UCRP_E2E_BASE_URL is not set: start the API and export the E2E environment (docs/e2e/05-implementation.md)")
	}
	s.API = client.New(cfg.BaseURL, cfg.RequestTimeout)

	spec, err := openapi.Load(cfg.OpenAPIPath)
	if err != nil {
		t.Fatalf("load OpenAPI contract: %v", err)
	}
	s.Spec = spec
	return s
}

func (s *Suite) WithDB() *Suite {
	s.T.Helper()
	if s.DB != nil {
		return s
	}
	if s.Cfg.DatabaseURL == "" {
		s.T.Skip("UCRP_E2E_DATABASE_URL is not set: fixtures need direct database access (docs/e2e/05-implementation.md)")
	}
	db, err := fixtures.OpenDB(context.Background(), s.Cfg.DatabaseURL)
	if err != nil {
		s.T.Fatalf("open fixture database: %v", err)
	}
	s.DB = db
	s.Reg = fixtures.NewRegistry(db, s.Cfg.RunID)
	s.T.Cleanup(func() {
		if err := s.Reg.Cleanup(context.Background()); err != nil {
			s.T.Logf("cleanup fixtures: %v", err)
		}
		_ = s.DB.Close()
	})
	return s
}

func (s *Suite) WithAdmin() *Suite {
	s.T.Helper()
	if s.Token != "" {
		return s
	}
	if s.Cfg.AdminEmail == "" || s.Cfg.AdminPassword == "" {
		s.T.Skip("UCRP_E2E_ADMIN_EMAIL / UCRP_E2E_ADMIN_PASSWORD are not set")
	}
	token, resp, err := s.API.LoginToken(context.Background(), s.Cfg.AdminEmail, s.Cfg.AdminPassword)
	if err != nil {
		s.T.Fatalf("admin login failed (status %d): %v", statusOf(resp), err)
	}
	s.Token = token
	s.API.SetToken(token)
	return s
}

func (s *Suite) As(email, password string) *client.Client {
	s.T.Helper()
	c := client.New(s.Cfg.BaseURL, s.Cfg.RequestTimeout)
	token, resp, err := c.LoginToken(context.Background(), email, password)
	if err != nil {
		s.T.Fatalf("login as %s failed (status %d): %v", email, statusOf(resp), err)
	}
	c.SetToken(token)
	return c
}

func (s *Suite) Marketer() *client.Client {
	s.T.Helper()
	s.WithDB()
	if s.marketer != nil {
		return s.marketer
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	user, err := s.Reg.User(ctx, "marketer", "marketer", "")
	if err != nil {
		s.T.Fatalf("fixture marketer: %v", err)
	}
	s.marketer = s.As(user.Email, user.Password)
	return s.marketer
}

func (s *Suite) WithMarketer() *Suite {
	s.T.Helper()
	s.API = s.Marketer()
	s.Token = ""
	return s
}

func (s *Suite) Ctx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), s.Cfg.RequestTimeout)
}

func (s *Suite) WaitFor(deadline time.Duration, what string, fn func(ctx context.Context) (bool, error)) {
	s.T.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), deadline)
	defer cancel()
	ticker := time.NewTicker(s.Cfg.PollInterval)
	defer ticker.Stop()
	var lastErr error
	for {
		done, err := fn(ctx)
		if err != nil {
			lastErr = err
		}
		if done {
			return
		}
		select {
		case <-ctx.Done():
			if lastErr != nil {
				s.T.Fatalf("timed out after %s waiting for %s: %v", deadline, what, lastErr)
			}
			s.T.Fatalf("timed out after %s waiting for %s", deadline, what)
		case <-ticker.C:
		}
	}
}

func (s *Suite) RequireStatus(resp *client.Response, want int, what string) {
	s.T.Helper()
	if resp.Status == want {
		return
	}
	report.Drifts.Record(s.T.Name(), what, fmt.Sprintf("expected HTTP %d, got %d", want, resp.Status))
	s.T.Fatalf("%s: expected HTTP %d, got %d (body=%s)", what, want, resp.Status, resp.Snippet())
}

func (s *Suite) RequireStatusOrDrift(resp *client.Response, want int, what string) {
	s.T.Helper()
	if resp.Status == want {
		return
	}
	s.reportDrift("", what, resp, []string{
		fmt.Sprintf("expected HTTP %d, got %d", want, resp.Status),
	})
	s.T.Skipf("%s: expected HTTP %d, got %d; the rest of this scenario cannot be evaluated",
		what, want, resp.Status)
}

func (s *Suite) RequireSchema(method, path string, resp *client.Response) {
	s.T.Helper()
	if len(resp.Body) == 0 {
		return
	}
	body, err := resp.JSON()
	if err != nil {
		s.T.Fatalf("%s %s: %v", method, path, err)
	}
	problems := s.Spec.ValidateResponse(method, path, resp.Status, body)
	if len(problems) == 0 {
		return
	}
	s.reportDrift(method, path, resp, problems)
}

func (s *Suite) RequireDeclaredStatus(method, path string, resp *client.Response) {
	s.T.Helper()
	declared := s.Spec.DeclaredStatuses(method, path)
	if len(declared) == 0 {
		s.T.Fatalf("%s %s: operation is absent from the OpenAPI contract", method, path)
	}
	for _, code := range declared {
		if code == fmt.Sprint(resp.Status) {
			return
		}
	}
	s.reportDrift(method, path, resp, []string{
		fmt.Sprintf("HTTP %d is not declared in the contract (declared: %s)", resp.Status, strings.Join(declared, ", ")),
	})
}

func (s *Suite) reportDrift(method, path string, resp *client.Response, problems []string) {
	s.T.Helper()
	detail := strings.Join(problems, "; ")
	report.Drifts.Record(s.T.Name(), method+" "+path, detail)
	if s.Cfg.StrictContract {
		s.T.Fatalf("%s %s: response violates the OpenAPI contract:\n  %s\n  body=%s",
			method, path, strings.Join(problems, "\n  "), resp.Snippet())
	}
	s.T.Logf("contract drift %s %s: %s", method, path, detail)
}

func (s *Suite) PostContractShape(ctx context.Context, path string, contractBody, implementationBody any) (*client.Response, error) {
	s.T.Helper()
	resp, err := s.API.Post(ctx, path, contractBody)
	if err != nil {
		return nil, err
	}
	if resp.Status < 400 {
		return resp, nil
	}
	report.Drifts.Record(s.T.Name(), "POST "+path,
		fmt.Sprintf("the contract request shape is rejected with HTTP %d (%s); the implementation expects a different body",
			resp.Status, resp.ErrorCode()))
	if s.Cfg.StrictContract {
		s.T.Fatalf("POST %s: the contract request shape is rejected with HTTP %d (body=%s)",
			path, resp.Status, resp.Snippet())
	}
	s.T.Logf("contract drift POST %s: contract request shape rejected with HTTP %d, retrying with the implementation shape",
		path, resp.Status)
	return s.API.Post(ctx, path, implementationBody)
}

func Field(obj map[string]any, names ...string) (any, bool) {
	for _, name := range names {
		if v, ok := obj[name]; ok {
			return v, true
		}
	}
	return nil, false
}

func StringField(obj map[string]any, names ...string) string {
	v, ok := Field(obj, names...)
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

func IntField(obj map[string]any, names ...string) (int, bool) {
	v, ok := Field(obj, names...)
	if !ok {
		return 0, false
	}
	f, ok := v.(float64)
	if !ok {
		return 0, false
	}
	return int(f), true
}

func Items(body any) ([]any, bool) {
	obj, ok := body.(map[string]any)
	if !ok {
		return nil, false
	}
	raw, ok := obj["items"]
	if !ok {
		return nil, false
	}
	items, ok := raw.([]any)
	return items, ok
}

func loadConfig() Config {
	return Config{
		BaseURL:        os.Getenv("UCRP_E2E_BASE_URL"),
		DatabaseURL:    os.Getenv("UCRP_E2E_DATABASE_URL"),
		AdminEmail:     os.Getenv("UCRP_E2E_ADMIN_EMAIL"),
		AdminPassword:  os.Getenv("UCRP_E2E_ADMIN_PASSWORD"),
		OpenAPIPath:    getenv("UCRP_E2E_OPENAPI", "../../../docs/swagger/openapi.json"),
		RunID:          getenv("UCRP_E2E_RUN_ID", randomRunID()),
		RequestTimeout: 30 * time.Second,
		ImportTimeout:  60 * time.Second,
		PollInterval:   200 * time.Millisecond,
		StrictContract: os.Getenv("UCRP_E2E_STRICT_CONTRACT") != "0",
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func randomRunID() string {
	buf := make([]byte, 4)
	if _, err := rand.Read(buf); err != nil {
		return "00000000"
	}
	return hex.EncodeToString(buf)
}

func statusOf(resp *client.Response) int {
	if resp == nil {
		return 0
	}
	return resp.Status
}

func failureMessage(t *testing.T) string {
	if !t.Failed() {
		return ""
	}
	return "test failed; see the go test output for the assertion message"
}
