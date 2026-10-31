package recommend

import (
	"context"
	"database/sql"
	"fmt"
	"sort"

	"github.com/MangriMen/UCRP/internal/domain"
	"github.com/MangriMen/UCRP/internal/platform/idgen"
)

const algorithmVersion = "v1"

type WeightedRecommender struct {
	db    *sql.DB
	idGen domain.IDGen
	clock domain.Clock
}

func NewWeightedRecommender(db *sql.DB) *WeightedRecommender {
	return &WeightedRecommender{db: db, idGen: idgen.UUID{}, clock: domain.SystemClock{}}
}

type candidate struct {
	id       string
	name     string
	priority int
}

func (r *WeightedRecommender) Refresh(ctx context.Context, guestID string) ([]domain.Recommendation, error) {
	visited, err := r.visitedTenants(ctx, guestID)
	if err != nil {
		return nil, err
	}
	tenants, err := r.activeTenants(ctx)
	if err != nil {
		return nil, err
	}
	similar, err := r.similarGuestTenants(ctx, guestID)
	if err != nil {
		return nil, err
	}

	now := r.clock.Now()
	out := make([]domain.Recommendation, 0, 3)
	for _, t := range tenants {
		if visited[t.id] {
			continue
		}
		score := 0.2 * float64(t.priority)
		why := fmt.Sprintf("приоритет арендатора %d/100", t.priority)
		if similar[t.id] > 0 {
			score += 40
			why = "похожие гости ходят в " + t.name + "; " + why
		}
		out = append(out, domain.Recommendation{
			GuestID:          guestID,
			TenantID:         t.id,
			Score:            score,
			Explanation:      why,
			AlgorithmVersion: algorithmVersion,
			ComputedAt:       now,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Score != out[j].Score {
			return out[i].Score > out[j].Score
		}
		return out[i].TenantID < out[j].TenantID
	})
	if len(out) > 3 {
		out = out[:3]
	}
	for i := range out {
		out[i].ID = r.idGen.NewID(ctx)
		out[i].Rank = i + 1
	}
	return out, r.replace(ctx, guestID, out)
}

func (r *WeightedRecommender) replace(ctx context.Context, guestID string, items []domain.Recommendation) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `DELETE FROM recommendation WHERE guest_id = $1`, guestID); err != nil {
		return err
	}
	for _, it := range items {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO recommendation (id, guest_id, tenant_id, rank, score, explanation, algorithm_version, computed_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
			it.ID, it.GuestID, it.TenantID, it.Rank, it.Score, it.Explanation, it.AlgorithmVersion, it.ComputedAt,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *WeightedRecommender) visitedTenants(ctx context.Context, guestID string) (map[string]bool, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT DISTINCT e.tenant_id
		FROM event e
		JOIN event_type et ON et.id = e.event_type_id
		WHERE e.guest_id = $1 AND et.counts_as_presence`, guestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out[id] = true
	}
	return out, rows.Err()
}

func (r *WeightedRecommender) activeTenants(ctx context.Context) ([]candidate, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, priority FROM tenant WHERE is_active ORDER BY priority DESC, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]candidate, 0)
	for rows.Next() {
		var t candidate
		if err := rows.Scan(&t.id, &t.name, &t.priority); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *WeightedRecommender) similarGuestTenants(ctx context.Context, guestID string) (map[string]int, error) {
	rows, err := r.db.QueryContext(ctx, `
		WITH visited AS (
			SELECT DISTINCT e.tenant_id
			FROM event e
			JOIN event_type et ON et.id = e.event_type_id
			WHERE e.guest_id = $1 AND et.counts_as_presence
		),
		peers AS (
			SELECT DISTINCT e2.guest_id
			FROM event e2
			JOIN visited v ON v.tenant_id = e2.tenant_id
			WHERE e2.guest_id IS NOT NULL AND e2.guest_id <> $1
		)
		SELECT e.tenant_id, COUNT(*) AS c
		FROM event e
		JOIN peers p ON p.guest_id = e.guest_id
		JOIN event_type et ON et.id = e.event_type_id AND et.counts_as_presence
		WHERE e.tenant_id NOT IN (SELECT tenant_id FROM visited)
		GROUP BY e.tenant_id
		ORDER BY c DESC
		LIMIT 10`, guestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]int{}
	for rows.Next() {
		var (
			id string
			c  int
		)
		if err := rows.Scan(&id, &c); err != nil {
			return nil, err
		}
		out[id] = c
	}
	return out, rows.Err()
}
