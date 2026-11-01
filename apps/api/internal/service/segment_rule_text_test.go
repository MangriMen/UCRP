package service

import (
	"context"
	"testing"
	"time"

	"github.com/MangriMen/UCRP/internal/domain"
)

func TestRuleTextOfUsesTenantNames(t *testing.T) {
	rule := domain.SegmentRule{
		VisitedTenantIDs:  []string{"t-cinema", "t-bean"},
		ExcludedTenantIDs: []string{"t-fit"},
		PeriodDays:        90,
	}
	got := ruleTextOf(rule, map[string]string{
		"t-cinema": "Кинотеатр Синема",
		"t-bean":   "Bean Coffee",
		"t-fit":    "ФитЗона",
	})
	want := "Был в «Кинотеатр Синема», «Bean Coffee», не был в «ФитЗона», за 90 дн."
	if got != want {
		t.Errorf("ruleText = %q, want %q", got, want)
	}
}

func TestRuleTextOfDistinguishesRules(t *testing.T) {
	a := ruleTextOf(domain.SegmentRule{VisitedTenantIDs: []string{"t1"}, ExcludedTenantIDs: []string{"t2"}, PeriodDays: 90}, map[string]string{"t1": "Кино", "t2": "Кофе"})
	b := ruleTextOf(domain.SegmentRule{VisitedTenantIDs: []string{"t3"}, ExcludedTenantIDs: []string{"t4"}, PeriodDays: 60}, map[string]string{"t3": "Спорт", "t4": "Еда"})
	if a == b {
		t.Errorf("different rules produced identical text: %q", a)
	}
}

func TestRuleTextOfGroup(t *testing.T) {
	rule := domain.SegmentRule{
		Operator: "or",
		Conditions: []domain.SegmentRule{
			{VisitedTenantIDs: []string{"t1"}, PeriodDays: 30},
			{ExcludedTenantIDs: []string{"t2"}, PeriodDays: 60},
		},
	}
	got := ruleTextOf(rule, map[string]string{"t1": "Кино", "t2": "Кофе"})
	want := "(Был в «Кино», за 30 дн.) или (не был в «Кофе», за 60 дн.)"
	if got != want {
		t.Errorf("group ruleText = %q, want %q", got, want)
	}
}

func TestRuleTextsBatch(t *testing.T) {
	s := &SegmentService{tenants: &ruleTextTenantStub{names: map[string]string{"t1": "Кино", "t2": "Кофе"}}}
	raws := []string{
		`{"visitedTenantIds":["t1"],"periodDays":90}`,
		``,
		`{"visitedTenantIds":["t2"],"periodDays":30}`,
	}
	got, err := s.ruleTexts(context.Background(), raws)
	if err != nil {
		t.Fatalf("ruleTexts: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("len = %d", len(got))
	}
	if got[0] != "Был в «Кино», за 90 дн." {
		t.Errorf("got[0] = %q", got[0])
	}
	if got[1] != "" {
		t.Errorf("got[1] = %q", got[1])
	}
	if got[2] != "Был в «Кофе», за 30 дн." {
		t.Errorf("got[2] = %q", got[2])
	}
}

type ruleTextTenantStub struct {
	names map[string]string
}

func (s *ruleTextTenantStub) Search(context.Context, domain.TenantFilter, int, int) ([]domain.TenantListEntry, int, error) {
	return nil, 0, nil
}
func (s *ruleTextTenantStub) GetByID(context.Context, string) (*domain.Tenant, error) {
	return nil, domain.ErrNotFound
}
func (s *ruleTextTenantStub) Names(_ context.Context, ids []string) (map[string]string, error) {
	out := map[string]string{}
	for _, id := range ids {
		if name, ok := s.names[id]; ok {
			out[id] = name
		}
	}
	return out, nil
}
func (s *ruleTextTenantStub) Insert(context.Context, *domain.Tenant) error { return nil }
func (s *ruleTextTenantStub) Update(context.Context, *domain.Tenant) error { return nil }
func (s *ruleTextTenantStub) SoftDelete(context.Context, string, time.Time) error {
	return nil
}
