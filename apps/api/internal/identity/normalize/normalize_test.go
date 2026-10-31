package normalize

import "testing"

func TestPhone(t *testing.T) {
	cases := map[string]string{
		"+7 (900) 123-45-67": "79001234567",
		"89001234567":        "79001234567",
		"79001234567":        "79001234567",
		"9001234567":         "79001234567",
		"":                   "",
		"abc":                "",
	}
	for in, want := range cases {
		if got := Phone(in); got != want {
			t.Errorf("Phone(%q)=%q want %q", in, got, want)
		}
	}
}

func TestEmail(t *testing.T) {
	if got := Email("  Anna@Example.COM "); got != "anna@example.com" {
		t.Fatalf("Email=%q", got)
	}
	if !ValidEmail("a@b.ru") {
		t.Fatal("valid email rejected")
	}
	if ValidEmail("bad-email") {
		t.Fatal("invalid email accepted")
	}
}

func TestCard(t *testing.T) {
	if got := Card(" 4111 1111 1111 1111 "); got != "4111111111111111" {
		t.Fatalf("Card=%q", got)
	}
}
