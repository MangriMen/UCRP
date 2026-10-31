package service

import "testing"

func TestNormalizeQuery(t *testing.T) {
	if got := normalizeQuery("+7 (900) 123-45-67"); got != "79001234567" {
		t.Fatalf("phone query=%q", got)
	}
	if got := normalizeQuery("89001234567"); got != "79001234567" {
		t.Fatalf("8-prefix query=%q", got)
	}
	if got := normalizeQuery("Anna@Example.COM"); got != "anna@example.com" {
		t.Fatalf("email query=%q", got)
	}
	if got := normalizeQuery("  Анна  "); got != "Анна" {
		t.Fatalf("name query=%q", got)
	}
	if got := normalizeQuery(""); got != "" {
		t.Fatalf("empty=%q", got)
	}
}
