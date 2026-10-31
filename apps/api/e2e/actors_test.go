package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/MangriMen/UCRP/e2e/client"
	"github.com/MangriMen/UCRP/e2e/harness"
)

func actor(t *testing.T, s *harness.Suite, role string) *client.Client {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	suffix := role + "-" + s.Reg.NewID()[:8]
	user, err := s.Reg.User(ctx, suffix, role, "")
	if err != nil {
		t.Fatalf("fixture %s: %v", role, err)
	}
	return s.As(user.Email, user.Password)
}

func jsonObject(t *testing.T, resp *client.Response, what string) map[string]any {
	t.Helper()
	body := mustJSON(t, resp, what)
	obj, ok := body.(map[string]any)
	if !ok {
		t.Fatalf("%s: expected a JSON object, got %T (%s)", what, body, resp.Snippet())
	}
	return obj
}

func mustJSON(t *testing.T, resp *client.Response, what string) any {
	t.Helper()
	body, err := resp.JSON()
	if err != nil {
		t.Fatalf("%s: decode body: %v", what, err)
	}
	return body
}
