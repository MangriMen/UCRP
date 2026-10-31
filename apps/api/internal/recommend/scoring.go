package recommend

import (
	"math"
	"sort"
	"time"

	"github.com/MangriMen/UCRP/internal/platform/config"
)

type HistoryVisit struct {
	TenantID    string
	Weight      float64
	LastVisitAt float64
}

type Candidate struct {
	TenantID   string
	Name       string
	CategoryID string
	Priority   int
	Floor      *int
}

type FactorValue struct {
	Code     string
	RawValue float64
	Weight   float64
}

type ScoredCandidate struct {
	TenantID    string
	Score       float64
	Factors     []FactorValue
	Explanation string
	LowData     bool
	HasBehavior bool
}

type VisitPair struct {
	GuestID  string
	TenantID string
}

func CountUniquePairs(pairs []VisitPair) (map[string]int, map[[2]string]int) {
	unique := map[[2]string]bool{}
	for _, p := range pairs {
		if p.GuestID == "" || p.TenantID == "" {
			continue
		}
		unique[[2]string{p.GuestID, p.TenantID}] = true
	}
	singles := map[string]int{}
	for key := range unique {
		singles[key[1]]++
	}
	co := map[[2]string]int{}
	byGuest := map[string][]string{}
	for key := range unique {
		byGuest[key[0]] = append(byGuest[key[0]], key[1])
	}
	for _, tenants := range byGuest {
		sort.Strings(tenants)
		for i := 0; i < len(tenants); i++ {
			for j := i + 1; j < len(tenants); j++ {
				co[[2]string{tenants[i], tenants[j]}]++
			}
		}
	}
	return singles, co
}

func CosineSupport(nAB, nA, nB int, lambda float64) float64 {
	if nA <= 0 || nB <= 0 || nAB <= 0 {
		return 0
	}
	cosine := float64(nAB) / math.Sqrt(float64(nA)*float64(nB))
	smooth := float64(nAB) / (float64(nAB) + lambda)
	return cosine * smooth
}

func RecencyWeight(lastVisit float64, now float64, halfLifeDays float64) float64 {
	if halfLifeDays <= 0 {
		return 1
	}
	days := (now - lastVisit) / (24 * float64(time.Hour))
	if days < 0 {
		days = 0
	}
	return math.Exp(-math.Ln2 * days / halfLifeDays)
}

func BehaviorScore(history []HistoryVisit, coVisits map[[2]string]int, singles map[string]int, candidateID string, lambda float64) (float64, bool) {
	var num, den float64
	hasStat := false
	for _, h := range history {
		nAB := coVisitCount(coVisits, h.TenantID, candidateID)
		nA := singles[h.TenantID]
		nB := singles[candidateID]
		if nA > 0 && nB > 0 {
			hasStat = true
		}
		sim := CosineSupport(nAB, nA, nB, lambda)
		num += h.Weight * sim
		den += h.Weight
	}
	if den <= 0 || !hasStat {
		return 0, false
	}
	return num / den, true
}

func coVisitCount(co map[[2]string]int, a, b string) int {
	if a == b {
		return 0
	}
	if a < b {
		if n, ok := co[[2]string{a, b}]; ok {
			return n
		}
		return 0
	}
	if n, ok := co[[2]string{b, a}]; ok {
		return n
	}
	return 0
}

type categoryRules struct {
	sameAffinity   float64
	adjacent       map[[2]string]float64
	competitorFrom map[string]map[string]bool
}

func buildCategoryRules(rels []categoryRelation, sameAffinity float64) categoryRules {
	out := categoryRules{
		sameAffinity:   sameAffinity,
		adjacent:       map[[2]string]float64{},
		competitorFrom: map[string]map[string]bool{},
	}
	for _, rel := range rels {
		key := [2]string{rel.From, rel.To}
		rev := [2]string{rel.To, rel.From}
		switch rel.Type {
		case "adjacent":
			w := rel.Weight
			if out.adjacent[key] < w {
				out.adjacent[key] = w
			}
			if out.adjacent[rev] < w {
				out.adjacent[rev] = w
			}
		case "competitor":
			if out.competitorFrom[rel.From] == nil {
				out.competitorFrom[rel.From] = map[string]bool{}
			}
			if out.competitorFrom[rel.To] == nil {
				out.competitorFrom[rel.To] = map[string]bool{}
			}
			out.competitorFrom[rel.From][rel.To] = true
			out.competitorFrom[rel.To][rel.From] = true
		}
	}
	return out
}

type categoryRelation struct {
	From   string
	To     string
	Type   string
	Weight float64
}

func CategoryAffinity(historyCategories []string, candidateCategory string, rules categoryRules) float64 {
	if candidateCategory == "" {
		return 0
	}
	best := 0.0
	for _, hc := range historyCategories {
		if hc == candidateCategory {
			if rules.sameAffinity > best {
				best = rules.sameAffinity
			}
			continue
		}
		if w, ok := rules.adjacent[[2]string{hc, candidateCategory}]; ok && w > best {
			best = w
		}
	}
	return best
}

func CompetitionMultiplier(historyCategories []string, candidateCategory string, rules categoryRules, penalty float64) float64 {
	if candidateCategory == "" || penalty <= 0 {
		return 1
	}
	for _, hc := range historyCategories {
		if rules.competitorFrom[hc][candidateCategory] {
			return 1 - penalty
		}
	}
	return 1
}

func PopularityFraction(nB, totalGuests int) float64 {
	if totalGuests <= 0 {
		return 0
	}
	return float64(nB) / float64(totalGuests)
}

func PriorityNormalized(priority int) float64 {
	if priority <= 0 {
		return 0
	}
	if priority > 100 {
		return 1
	}
	return float64(priority) / 100
}

func ScoreCandidates(cfg config.RecommendConfig, history []HistoryVisit, historyCategories []string, candidates []Candidate, singles map[string]int, coVisits map[[2]string]int, totalGuests int, rules categoryRules) []ScoredCandidate {
	out := make([]ScoredCandidate, 0, len(candidates))
	for _, cand := range candidates {
		var factors []FactorValue
		var base float64
		lowData := len(history) == 0
		hasBehavior := false

		if len(history) > 0 {
			if behavior, ok := BehaviorScore(history, coVisits, singles, cand.TenantID, cfg.Lambda); ok {
				hasBehavior = true
				base += cfg.WeightBehavior * behavior
				factors = append(factors, FactorValue{Code: "similar_guests", RawValue: clamp01(behavior), Weight: cfg.WeightBehavior})
			} else {
				lowData = true
			}
			cat := CategoryAffinity(historyCategories, cand.CategoryID, rules)
			base += cfg.WeightCategory * cat
			factors = append(factors, FactorValue{Code: "category_adjacency", RawValue: clamp01(cat), Weight: cfg.WeightCategory})
		}

		pop := PopularityFraction(singles[cand.TenantID], totalGuests)
		pri := PriorityNormalized(cand.Priority)

		if lowData {
			wSum := cfg.WeightPopularity + cfg.WeightPriority
			if wSum <= 0 {
				wSum = 1
			}
			base = (cfg.WeightPopularity*pop + cfg.WeightPriority*pri) / wSum
			factors = []FactorValue{
				{Code: "low_data", RawValue: 1, Weight: 0},
				{Code: "popularity", RawValue: clamp01(pop), Weight: cfg.WeightPopularity / wSum},
				{Code: "tenant_priority", RawValue: pri, Weight: cfg.WeightPriority / wSum},
			}
		} else {
			base += cfg.WeightPopularity * pop
			base += cfg.WeightPriority * pri
			factors = append(factors,
				FactorValue{Code: "popularity", RawValue: clamp01(pop), Weight: cfg.WeightPopularity},
				FactorValue{Code: "tenant_priority", RawValue: pri, Weight: cfg.WeightPriority},
			)
		}

		mult := CompetitionMultiplier(historyCategories, cand.CategoryID, rules, cfg.CompetitorPenalty)
		score := base * mult
		if mult < 1 {
			factors = append(factors, FactorValue{Code: "competitor_penalty", RawValue: clamp01(mult), Weight: 1})
		}

		out = append(out, ScoredCandidate{
			TenantID:    cand.TenantID,
			Score:       score,
			Factors:     factors,
			Explanation: explain(factors, lowData, hasBehavior, cand),
			LowData:     lowData,
			HasBehavior: hasBehavior,
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Score != out[j].Score {
			return out[i].Score > out[j].Score
		}
		return out[i].TenantID < out[j].TenantID
	})
	return out
}

func explain(factors []FactorValue, lowData, hasBehavior bool, cand Candidate) string {
	if lowData {
		return "мало данных: популярность и приоритет арендатора"
	}
	parts := make([]string, 0, 4)
	for _, f := range factors {
		switch f.Code {
		case "similar_guests":
			if f.RawValue > 0 {
				parts = append(parts, "другие гости также посещали "+cand.Name)
			}
		case "category_adjacency":
			if f.RawValue > 0 {
				parts = append(parts, "категории смежные")
			}
		case "competitor_penalty":
			parts = append(parts, "понижен из-за прямой конкуренции")
		}
	}
	if !hasBehavior {
		parts = append(parts, "недостаточно совместных посещений")
	}
	if len(parts) == 0 {
		return "приоритет и популярность арендатора"
	}
	return joinParts(parts)
}

func joinParts(parts []string) string {
	out := parts[0]
	for _, p := range parts[1:] {
		out += "; " + p
	}
	return out
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
