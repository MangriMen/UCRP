package recommend

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/MangriMen/UCRP/internal/domain"
	"github.com/MangriMen/UCRP/internal/platform/idgen"
	"github.com/MangriMen/UCRP/internal/platform/kev"
	"github.com/MangriMen/UCRP/internal/repository"
)

const (
	evalActionLogLimit = 50
	evalInstructions   = "Given the visitor's action log, rank these top candidate places from worst to best for this visitor."
)

type Ranker interface {
	Rank(ctx context.Context, state string, criteria map[string]string, instructions string) (*kev.Ranking, error)
}

type ActiveTenantsSource interface {
	ActiveTenants(ctx context.Context) ([]repository.TenantCandidate, error)
}

type Evaluator struct {
	events   domain.EventRepo
	recos    domain.RecommendationRepo
	evals    domain.RecommendationEvalRepo
	presence ActiveTenantsSource
	ranker   Ranker
	clock    domain.Clock
	idGen    domain.IDGen
}

func NewEvaluator(
	events domain.EventRepo,
	recos domain.RecommendationRepo,
	evals domain.RecommendationEvalRepo,
	presence ActiveTenantsSource,
	ranker Ranker,
) *Evaluator {
	return &Evaluator{
		events:   events,
		recos:    recos,
		evals:    evals,
		presence: presence,
		ranker:   ranker,
		clock:    domain.SystemClock{},
		idGen:    idgen.UUID{},
	}
}

func (e *Evaluator) Evaluate(ctx context.Context, guestID string) error {
	recos, err := e.recos.ListByGuest(ctx, guestID)
	if err != nil {
		return err
	}
	if len(recos) == 0 {
		return nil
	}
	actionLog, err := e.events.ActionLog(ctx, guestID, evalActionLogLimit)
	if err != nil {
		return err
	}
	names, err := e.tenantNames(ctx, recos)
	if err != nil {
		return err
	}

	state := buildEvalState(actionLog, recos, names)
	criteria := buildEvalCriteria(recos)
	ranking, err := e.ranker.Rank(ctx, state, criteria, evalInstructions)
	if err != nil {
		return err
	}

	ev := &domain.RecommendationEval{
		ID:                       e.idGen.NewID(ctx),
		GuestID:                  guestID,
		Probabilities:            ranking.Probabilities,
		Ranking:                  orderRanking(ranking.Probabilities, criteria),
		Confidence:               ranking.Confidence,
		RecommendationComputedAt: recos[0].ComputedAt,
		EvaluatedAt:              e.clock.Now(),
	}
	return e.evals.Insert(ctx, ev)
}

func (e *Evaluator) tenantNames(ctx context.Context, recos []domain.Recommendation) (map[string]string, error) {
	wanted := make(map[string]bool, len(recos))
	for _, rec := range recos {
		wanted[rec.TenantID] = true
	}
	candidates, err := e.presence.ActiveTenants(ctx)
	if err != nil {
		return nil, err
	}
	names := make(map[string]string, len(wanted))
	for _, c := range candidates {
		if wanted[c.ID] {
			names[c.ID] = c.Name
		}
	}
	return names, nil
}

func buildEvalState(actionLog []domain.GuestActionLogEntry, recos []domain.Recommendation, names map[string]string) string {
	var b strings.Builder
	b.WriteString("Visitor action log (most recent first):\n")
	if len(actionLog) == 0 {
		b.WriteString("(empty)\n")
	}
	for _, entry := range actionLog {
		line := fmt.Sprintf("- %s | %s | %s", entry.OccurredAt.Format(time.RFC3339), entry.TenantName, entry.EventType)
		if entry.AmountMinor != nil {
			line += fmt.Sprintf(" | %d %s", *entry.AmountMinor, entry.Currency)
		}
		b.WriteString(line + "\n")
	}
	b.WriteString("Top-3 recommendations (rank | score | reason):\n")
	for _, rec := range recos {
		fmt.Fprintf(&b, "- %d | %s | %.4f | %s\n", rec.Rank, tenantLabel(rec.TenantID, names), rec.Score, rec.Explanation)
	}
	return b.String()
}

func buildEvalCriteria(recos []domain.Recommendation) map[string]string {
	criteria := make(map[string]string, len(recos))
	for _, rec := range recos {
		criteria[rec.TenantID] = fmt.Sprintf("rank %d, score %.4f, reason: %s", rec.Rank, rec.Score, rec.Explanation)
	}
	return criteria
}

func orderRanking(probabilities map[string]float64, criteria map[string]string) []string {
	ids := make([]string, 0, len(criteria))
	for id := range criteria {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		pi, pj := probabilities[ids[i]], probabilities[ids[j]]
		if pi == pj {
			return ids[i] < ids[j]
		}
		return pi < pj
	})
	return ids
}

func tenantLabel(tenantID string, names map[string]string) string {
	if name, ok := names[tenantID]; ok && name != "" {
		return name
	}
	return tenantID
}
