package service

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/MangriMen/UCRP/internal/domain"
)

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

type searchGuestRepo struct {
	domain.GuestRepo
	entries []domain.GuestListEntry
}

func (r searchGuestRepo) Search(context.Context, domain.GuestFilter, int, int) ([]domain.GuestListEntry, int, error) {
	return r.entries, len(r.entries), nil
}

type pageIdentifierRepo struct {
	domain.GuestIdentifierRepo
	rows  []domain.GuestIdentifier
	calls [][]string
}

func (r *pageIdentifierRepo) ListByGuests(_ context.Context, guestIDs []string) ([]domain.GuestIdentifier, error) {
	r.calls = append(r.calls, guestIDs)
	return r.rows, nil
}

func TestProfileSearch_MasksIdentifiersOfThePageInOneQuery_WhenAllowed(t *testing.T) {
	guests := searchGuestRepo{entries: []domain.GuestListEntry{{Guest: domain.Guest{ID: "g1"}}, {Guest: domain.Guest{ID: "g2"}}}}
	identifiers := &pageIdentifierRepo{rows: []domain.GuestIdentifier{
		{GuestID: "g1", IDType: "phone", ValueNormalized: "79001234567"},
		{GuestID: "g1", IDType: "email", ValueNormalized: "ivan@example.com"},
	}}
	svc := NewProfileService(guests, identifiers, nil, nil)

	page, err := svc.Search(t.Context(), domain.GuestFilter{}, 20, 0, GuestListAccess{Identifiers: true})
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(identifiers.calls, [][]string{{"g1", "g2"}}) {
		t.Fatalf("expected one identifier query for the whole page, got %v", identifiers.calls)
	}
	want := []MaskedGuestIdentifier{{IDType: "phone", ValueMasked: "***4567"}, {IDType: "email", ValueMasked: "i***@example.com"}}
	if !reflect.DeepEqual(page.Items[0].MaskedIdentifiers, want) {
		t.Fatalf("g1 identifiers = %v, want %v", page.Items[0].MaskedIdentifiers, want)
	}
	body, err := json.Marshal(page.Items[1])
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), `"maskedIdentifiers":[]`) {
		t.Fatalf("a guest without identifiers must carry an empty list: %s", body)
	}
}

func TestProfileSearch_OmitsIdentifiers_WhenNotAllowed(t *testing.T) {
	guests := searchGuestRepo{entries: []domain.GuestListEntry{{Guest: domain.Guest{ID: "g1"}}}}
	identifiers := &pageIdentifierRepo{rows: []domain.GuestIdentifier{{GuestID: "g1", IDType: "phone", ValueNormalized: "79001234567"}}}
	svc := NewProfileService(guests, identifiers, nil, nil)

	page, err := svc.Search(t.Context(), domain.GuestFilter{}, 20, 0, GuestListAccess{})
	if err != nil {
		t.Fatal(err)
	}

	if len(identifiers.calls) != 0 {
		t.Fatalf("identifiers must not be read without permission, got %v", identifiers.calls)
	}
	body, err := json.Marshal(page.Items[0])
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(body), "maskedIdentifiers") || strings.Contains(string(body), "4567") {
		t.Fatalf("identifiers leaked without permission: %s", body)
	}
}
