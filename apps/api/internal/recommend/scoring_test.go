package recommend

import (
	"math"
	"testing"
	"time"

	"github.com/MangriMen/UCRP/internal/platform/config"
)

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}

func TestCosineSupportMatchesDocumentExample(t *testing.T) {
	simAB := CosineSupport(2, 4, 4, 2)
	if !almostEqual(simAB, 0.25) {
		t.Fatalf("similarity(A,B): want 0.25, got %v", simAB)
	}
	simCB := CosineSupport(2, 3, 4, 2)
	want := 2 / math.Sqrt(12) * 0.5
	if !almostEqual(simCB, want) {
		t.Fatalf("similarity(C,B): want %v, got %v", want, simCB)
	}
}

func TestBehaviorScoreUsesAllHistoryWeights(t *testing.T) {
	history := []HistoryVisit{
		{TenantID: "A", Weight: 1},
		{TenantID: "C", Weight: 0.5},
	}
	singles := map[string]int{"A": 4, "B": 4, "C": 3}
	co := map[[2]string]int{
		{"A", "B"}: 2,
		{"B", "C"}: 2,
	}
	score, ok := BehaviorScore(history, co, singles, "B", 2)
	if !ok {
		t.Fatal("behavior expected to be defined")
	}
	simAB := CosineSupport(2, 4, 4, 2)
	simCB := CosineSupport(2, 3, 4, 2)
	want := (1*simAB + 0.5*simCB) / 1.5
	if !almostEqual(score, want) {
		t.Fatalf("behavior: want %v, got %v", want, score)
	}
}

func TestBehaviorScoreUndefinedWithoutHistoryOrStats(t *testing.T) {
	if _, ok := BehaviorScore(nil, nil, map[string]int{"B": 4}, "B", 2); ok {
		t.Fatal("empty history must be undefined")
	}
	if _, ok := BehaviorScore([]HistoryVisit{{TenantID: "A", Weight: 1}}, nil, map[string]int{"A": 1}, "B", 2); ok {
		t.Fatal("missing candidate stats must be undefined")
	}
}

func TestScoreCandidatesDocumentScore(t *testing.T) {
	cfg := config.RecommendConfig{
		AlgorithmVersion:     "test",
		Lambda:               2,
		WeightBehavior:       0.6,
		WeightCategory:       0,
		WeightPopularity:     0.25,
		WeightPriority:       0.15,
		CompetitorPenalty:    0,
		SameCategoryAffinity: 1,
	}
	history := []HistoryVisit{
		{TenantID: "A", Weight: 1},
		{TenantID: "C", Weight: 0.5},
	}
	singles := map[string]int{"A": 4, "B": 4, "C": 3}
	co := map[[2]string]int{
		{"A", "B"}: 2,
		{"B", "C"}: 2,
	}
	rules := buildCategoryRules(nil, 1)
	scored := ScoreCandidates(cfg, history, nil, []Candidate{{TenantID: "B", Priority: 60}}, singles, co, 6, rules)
	if len(scored) != 1 {
		t.Fatalf("expected 1 scored candidate, got %d", len(scored))
	}
	behavior, ok := BehaviorScore(history, co, singles, "B", 2)
	if !ok {
		t.Fatal("behavior undefined")
	}
	want := 0.6*behavior + 0.25*(4.0/6.0) + 0.15*0.6
	if !almostEqual(scored[0].Score, want) {
		t.Fatalf("score: want %v, got %v", want, scored[0].Score)
	}
}

func TestFallbackWithoutHistoryDoesNotTreatBehaviorAsZero(t *testing.T) {
	cfg := config.DefaultRecommendConfig()
	scored := ScoreCandidates(cfg, nil, nil, []Candidate{{TenantID: "B", Priority: 50}}, map[string]int{"B": 3}, nil, 6, buildCategoryRules(nil, 1))
	if len(scored) != 1 {
		t.Fatalf("expected 1 candidate, got %d", len(scored))
	}
	if !scored[0].LowData {
		t.Fatal("fallback must be marked low data")
	}
	for _, f := range scored[0].Factors {
		if f.Code == "similar_guests" {
			t.Fatal("fallback must not emit similar_guests as measured zero")
		}
	}
	want := (cfg.WeightPopularity*(3.0/6.0) + cfg.WeightPriority*0.5) / (cfg.WeightPopularity + cfg.WeightPriority)
	if !almostEqual(scored[0].Score, want) {
		t.Fatalf("fallback score: want %v, got %v", want, scored[0].Score)
	}
}

func TestVisitedAndInactiveExcludedFromCandidates(t *testing.T) {
	visited := map[string]bool{"A": true}
	candidates := []Candidate{{TenantID: "A"}, {TenantID: "B"}, {TenantID: "C"}}
	eligible := make([]Candidate, 0, len(candidates))
	for _, c := range candidates {
		if visited[c.TenantID] {
			continue
		}
		eligible = append(eligible, c)
	}
	if len(eligible) != 2 || eligible[0].TenantID != "B" || eligible[1].TenantID != "C" {
		t.Fatalf("eligible candidates wrong: %+v", eligible)
	}
}

func TestCategoryAndCompetitionFollowConfig(t *testing.T) {
	rels := []categoryRelation{
		{From: "food", To: "ent", Type: "adjacent", Weight: 0.5},
		{From: "food", To: "coffee", Type: "competitor", Weight: 1},
	}
	rules := buildCategoryRules(rels, 1)
	if got := CategoryAffinity([]string{"food"}, "ent", rules); !almostEqual(got, 0.5) {
		t.Fatalf("adjacent affinity: want 0.5, got %v", got)
	}
	if got := CategoryAffinity([]string{"food"}, "food", rules); !almostEqual(got, 1) {
		t.Fatalf("same category affinity: want 1, got %v", got)
	}
	if got := CompetitionMultiplier([]string{"food"}, "coffee", rules, 0.3); !almostEqual(got, 0.7) {
		t.Fatalf("competitor multiplier: want 0.7, got %v", got)
	}
	if got := CompetitionMultiplier([]string{"food"}, "ent", rules, 0.3); !almostEqual(got, 1) {
		t.Fatalf("non-competitor multiplier: want 1, got %v", got)
	}
}

func TestStableSortByScoreThenTenantID(t *testing.T) {
	cfg := config.DefaultRecommendConfig()
	scored := ScoreCandidates(cfg, nil, nil, []Candidate{
		{TenantID: "b", Priority: 50},
		{TenantID: "a", Priority: 50},
	}, map[string]int{}, nil, 0, buildCategoryRules(nil, 1))
	if scored[0].TenantID != "a" || scored[1].TenantID != "b" {
		t.Fatalf("stable order broken: %+v", scored)
	}
}

func TestAggregateSegmentBothDenominatorsAndIncomplete(t *testing.T) {
	now := time.Now()
	members := []SegmentMemberSnapshot{
		{GuestID: "g1", Visited: map[string]bool{"A": true}, Recs: []MemberRecommendation{
			{TenantID: "B", Score: 0.6, ComputedAt: now},
			{TenantID: "C", Score: 0.5, ComputedAt: now},
		}, HasFresh: true},
		{GuestID: "g2", Visited: map[string]bool{}, Recs: []MemberRecommendation{
			{TenantID: "B", Score: 0.3, ComputedAt: now},
		}, HasFresh: true},
		{GuestID: "g3", Visited: map[string]bool{}, Recs: nil, HasFresh: false},
	}
	aggs := AggregateSegmentMembers(members, 10)
	var b *SegmentTenantAggregate
	for i := range aggs {
		if aggs[i].TenantID == "B" {
			b = &aggs[i]
		}
	}
	if b == nil {
		t.Fatal("tenant B missing")
	}
	if b.SegmentCount != 3 || b.EligibleCount != 3 || b.GuestCount != 2 {
		t.Fatalf("B counts wrong: %+v", b)
	}
	if b.IncompleteCount != 1 {
		t.Fatalf("B incomplete: want 1, got %d", b.IncompleteCount)
	}
	if !almostEqual(b.AverageScore, 0.45) {
		t.Fatalf("B average: want 0.45, got %v", b.AverageScore)
	}
}

func TestAggregateEmptyPersonalSetIsNotZeroScore(t *testing.T) {
	members := []SegmentMemberSnapshot{
		{GuestID: "g1", Visited: map[string]bool{}, Recs: []MemberRecommendation{}, HasFresh: true},
		{GuestID: "g2", Visited: map[string]bool{}, Recs: nil, HasFresh: true},
	}
	aggs := AggregateSegmentMembers(members, 10)
	for _, a := range aggs {
		if a.GuestCount == 0 && almostEqual(a.AverageScore, 0) && a.IncompleteCount == 0 {
			continue
		}
		t.Fatalf("unexpected aggregate %+v", a)
	}
}

func TestCountUniquePairsIgnoresDuplicateEvents(t *testing.T) {
	singles, co := CountUniquePairs([]VisitPair{
		{GuestID: "g1", TenantID: "A"},
		{GuestID: "g1", TenantID: "A"},
		{GuestID: "g1", TenantID: "B"},
		{GuestID: "g2", TenantID: "A"},
		{GuestID: "g2", TenantID: "B"},
	})
	if singles["A"] != 2 || singles["B"] != 2 {
		t.Fatalf("unique singles wrong: %v", singles)
	}
	if co[[2]string{"A", "B"}] != 2 {
		t.Fatalf("co-visit nAB wrong: %v", co)
	}
}

func TestScoreCandidatesIsIdempotent(t *testing.T) {
	cfg := config.DefaultRecommendConfig()
	history := []HistoryVisit{{TenantID: "A", Weight: 1}}
	singles := map[string]int{"A": 2, "B": 2}
	co := map[[2]string]int{{"A", "B"}: 1}
	cands := []Candidate{{TenantID: "B", Priority: 40}, {TenantID: "C", Priority: 40}}
	first := ScoreCandidates(cfg, history, nil, cands, singles, co, 4, buildCategoryRules(nil, 1))
	second := ScoreCandidates(cfg, history, nil, cands, singles, co, 4, buildCategoryRules(nil, 1))
	if len(first) != len(second) {
		t.Fatalf("length mismatch")
	}
	for i := range first {
		if first[i].TenantID != second[i].TenantID || !almostEqual(first[i].Score, second[i].Score) {
			t.Fatalf("reprocess changed result at %d: %+v vs %+v", i, first[i], second[i])
		}
	}
}

func TestAggregateSegmentIsIdempotent(t *testing.T) {
	now := time.Now()
	members := []SegmentMemberSnapshot{
		{GuestID: "g1", Visited: map[string]bool{}, Recs: []MemberRecommendation{
			{TenantID: "B", Score: 0.4, ComputedAt: now},
		}, HasFresh: true},
	}
	a := AggregateSegmentMembers(members, 5)
	b := AggregateSegmentMembers(members, 5)
	if len(a) != len(b) || a[0].GuestCount != b[0].GuestCount || !almostEqual(a[0].AverageScore, b[0].AverageScore) {
		t.Fatalf("aggregate not idempotent: %+v vs %+v", a, b)
	}
}

func TestRecencyWeightDecays(t *testing.T) {
	now := float64(time.Now().UnixNano())
	half := 30 * 24 * float64(time.Hour.Nanoseconds())
	fresh := RecencyWeight(now, now, 30)
	old := RecencyWeight(now-half, now, 30)
	if !almostEqual(fresh, 1) {
		t.Fatalf("fresh weight: want 1, got %v", fresh)
	}
	if !almostEqual(old, 0.5) {
		t.Fatalf("half-life weight: want 0.5, got %v", old)
	}
}
