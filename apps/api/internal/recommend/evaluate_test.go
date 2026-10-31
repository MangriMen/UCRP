package recommend

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/MangriMen/UCRP/internal/domain"
	"github.com/MangriMen/UCRP/internal/platform/kev"
	"github.com/MangriMen/UCRP/internal/repository"
)

type stubEventRepo struct {
	entries []domain.GuestActionLogEntry
}

func (s *stubEventRepo) ListByGuest(context.Context, string, int, int) ([]domain.Event, error) {
	return nil, nil
}
func (s *stubEventRepo) ListUnmatchedByBatch(context.Context, string, int) ([]domain.Event, error) {
	return nil, nil
}
func (s *stubEventRepo) ActionLog(context.Context, string, int) ([]domain.GuestActionLogEntry, error) {
	return s.entries, nil
}
func (s *stubEventRepo) Insert(context.Context, *domain.Event) error { return nil }
func (s *stubEventRepo) InsertIgnoreDuplicate(context.Context, *domain.Event) (bool, error) {
	return false, nil
}
func (s *stubEventRepo) AttachGuest(context.Context, string, string) error { return nil }
func (s *stubEventRepo) MoveToGuest(context.Context, string, string) error { return nil }
func (s *stubEventRepo) RestoreOrigin(context.Context, string) error       { return nil }

type stubRecoRepo struct {
	recos []domain.Recommendation
}

func (s *stubRecoRepo) ListByGuest(context.Context, string) ([]domain.Recommendation, error) {
	return s.recos, nil
}

type stubEvalRepo struct {
	got *domain.RecommendationEval
}

func (s *stubEvalRepo) Insert(_ context.Context, ev *domain.RecommendationEval) error {
	s.got = ev
	return nil
}

type stubTenants struct {
	list []repository.TenantCandidate
}

func (s *stubTenants) ActiveTenants(context.Context) ([]repository.TenantCandidate, error) {
	return s.list, nil
}

type stubRanker struct {
	ranking     *kev.Ranking
	gotState    string
	gotCriteria map[string]string
	gotInstr    string
	calls       int
}

func (s *stubRanker) Rank(_ context.Context, state string, criteria map[string]string, instructions string) (*kev.Ranking, error) {
	s.gotState = state
	s.gotCriteria = criteria
	s.gotInstr = instructions
	s.calls++
	return s.ranking, nil
}

func TestEvaluateStoresWorstToBestRanking(t *testing.T) {
	computedAt := time.Date(2026, 2, 1, 10, 0, 0, 0, time.UTC)
	amount := int64(1500)
	evalRepo := &stubEvalRepo{}
	ranker := &stubRanker{ranking: &kev.Ranking{
		Choice:        "t-b",
		Confidence:    nil,
		Probabilities: map[string]float64{"t-a": 0.7, "t-b": 0.1, "t-c": 0.5},
	}}
	e := NewEvaluator(
		&stubEventRepo{entries: []domain.GuestActionLogEntry{{
			OccurredAt:  time.Date(2026, 1, 20, 12, 30, 0, 0, time.UTC),
			TenantName:  "Bean Coffee",
			EventType:   "purchase",
			AmountMinor: &amount,
			Currency:    "RUB",
		}}},
		&stubRecoRepo{recos: []domain.Recommendation{
			{ID: "r1", GuestID: "g-1", TenantID: "t-a", Rank: 1, Score: 0.4, Explanation: "similar guests", ComputedAt: computedAt},
			{ID: "r2", GuestID: "g-1", TenantID: "t-b", Rank: 2, Score: 0.5, Explanation: "adjacent categories", ComputedAt: computedAt},
			{ID: "r3", GuestID: "g-1", TenantID: "t-c", Rank: 3, Score: 0.3, Explanation: "freshness", ComputedAt: computedAt},
		}},
		evalRepo,
		&stubTenants{list: []repository.TenantCandidate{
			{ID: "t-a", Name: "Cinema"},
			{ID: "t-b", Name: "FitZone"},
			{ID: "t-c", Name: "Bookstore"},
		}},
		ranker,
	)

	if err := e.Evaluate(context.Background(), "g-1"); err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if ranker.calls != 1 {
		t.Fatalf("ranker calls = %d", ranker.calls)
	}
	if !strings.Contains(ranker.gotState, "Bean Coffee") || !strings.Contains(ranker.gotState, "2026-01-20T12:30:00Z") {
		t.Errorf("state misses action log:\n%s", ranker.gotState)
	}
	if !strings.Contains(ranker.gotState, "Cinema") || !strings.Contains(ranker.gotState, "similar guests") {
		t.Errorf("state misses recommendations:\n%s", ranker.gotState)
	}
	if len(ranker.gotCriteria) != 3 {
		t.Errorf("criteria = %v", ranker.gotCriteria)
	}
	if !strings.Contains(ranker.gotInstr, "worst to best") {
		t.Errorf("instructions = %q", ranker.gotInstr)
	}

	got := evalRepo.got
	if got == nil {
		t.Fatal("eval not stored")
	}
	if got.GuestID != "g-1" || got.ID == "" {
		t.Errorf("eval ids: %+v", got)
	}
	wantRanking := []string{"t-b", "t-c", "t-a"}
	if len(got.Ranking) != len(wantRanking) {
		t.Fatalf("ranking = %v", got.Ranking)
	}
	for i, id := range wantRanking {
		if got.Ranking[i] != id {
			t.Errorf("ranking[%d] = %s, want %s", i, got.Ranking[i], id)
		}
	}
	if !got.RecommendationComputedAt.Equal(computedAt) {
		t.Errorf("RecommendationComputedAt = %s", got.RecommendationComputedAt)
	}
	if got.EvaluatedAt.IsZero() {
		t.Error("EvaluatedAt is zero")
	}
	if len(got.Probabilities) != 3 {
		t.Errorf("probabilities = %v", got.Probabilities)
	}
}

func TestEvaluateSkipsGuestWithoutRecommendations(t *testing.T) {
	evalRepo := &stubEvalRepo{}
	ranker := &stubRanker{ranking: &kev.Ranking{Choice: "t-a", Probabilities: map[string]float64{"t-a": 1}}}
	e := NewEvaluator(&stubEventRepo{}, &stubRecoRepo{}, evalRepo, &stubTenants{}, ranker)

	if err := e.Evaluate(context.Background(), "g-empty"); err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if ranker.calls != 0 {
		t.Errorf("ranker calls = %d", ranker.calls)
	}
	if evalRepo.got != nil {
		t.Error("eval stored for empty recommendations")
	}
}

func TestBuildEvalStateEmptyLog(t *testing.T) {
	state := buildEvalState(nil, nil, nil)
	if !strings.Contains(state, "(empty)") {
		t.Errorf("state = %q", state)
	}
}

func TestOrderRanking(t *testing.T) {
	criteria := map[string]string{"t-a": "", "t-b": "", "t-c": "", "t-d": ""}
	probs := map[string]float64{"t-a": 0.4, "t-b": 0.4, "t-c": 0.9}
	got := orderRanking(probs, criteria)
	want := []string{"t-d", "t-a", "t-b", "t-c"}
	if len(got) != len(want) {
		t.Fatalf("ranking = %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("ranking = %v, want %v", got, want)
		}
	}
}
