package recommend

import (
	"context"
	"database/sql"

	"github.com/MangriMen/UCRP/internal/domain"
	"github.com/MangriMen/UCRP/internal/platform/config"
	"github.com/MangriMen/UCRP/internal/platform/idgen"
	"github.com/MangriMen/UCRP/internal/repository"
)

type WeightedRecommender struct {
	db       *sql.DB
	presence *repository.PresenceSQL
	cfg      config.RecommendConfig
	idGen    domain.IDGen
	clock    domain.Clock
}

func NewWeightedRecommender(db *sql.DB, cfg config.RecommendConfig) *WeightedRecommender {
	return &WeightedRecommender{
		db:       db,
		presence: repository.NewPresenceSQL(db),
		cfg:      cfg,
		idGen:    idgen.UUID{},
		clock:    domain.SystemClock{},
	}
}

func (r *WeightedRecommender) Refresh(ctx context.Context, guestID string) ([]domain.Recommendation, error) {
	now := r.clock.Now().UTC()
	since := now.AddDate(0, 0, -r.cfg.WindowDays)

	visited, err := r.presence.VisitedTenants(ctx, guestID)
	if err != nil {
		return nil, err
	}
	historyRows, err := r.presence.GuestHistory(ctx, guestID, since)
	if err != nil {
		return nil, err
	}
	candidates, err := r.presence.ActiveTenants(ctx)
	if err != nil {
		return nil, err
	}

	tenantCat := map[string]string{}
	for _, c := range candidates {
		tenantCat[c.ID] = c.CategoryID
	}
	history := make([]HistoryVisit, 0, len(historyRows))
	historyCategories := make([]string, 0, len(historyRows))
	seenHistCat := map[string]bool{}
	for _, h := range historyRows {
		history = append(history, HistoryVisit{
			TenantID:    h.TenantID,
			Weight:      RecencyWeight(float64(h.LastVisitAt.UnixNano()), float64(now.UnixNano()), r.cfg.RecencyHalfLifeDays),
			LastVisitAt: float64(h.LastVisitAt.UnixNano()),
		})
		if cid, ok := tenantCat[h.TenantID]; ok && !seenHistCat[cid] {
			seenHistCat[cid] = true
			historyCategories = append(historyCategories, cid)
		}
	}

	eligible := make([]Candidate, 0, len(candidates))
	for _, c := range candidates {
		if visited[c.ID] {
			continue
		}
		eligible = append(eligible, Candidate{
			TenantID:   c.ID,
			Name:       c.Name,
			CategoryID: c.CategoryID,
			Priority:   c.Priority,
			Floor:      c.Floor,
		})
	}

	tenantIDs := make([]string, 0, len(eligible)+len(history))
	seen := map[string]bool{}
	for _, h := range history {
		if !seen[h.TenantID] {
			seen[h.TenantID] = true
			tenantIDs = append(tenantIDs, h.TenantID)
		}
	}
	for _, c := range eligible {
		if !seen[c.TenantID] {
			seen[c.TenantID] = true
			tenantIDs = append(tenantIDs, c.TenantID)
		}
	}

	singles, coVisits, err := r.presence.PairGuestCounts(ctx, since, tenantIDs)
	if err != nil {
		return nil, err
	}
	totalGuests, err := r.presence.UniquePresenceGuests(ctx, since)
	if err != nil {
		return nil, err
	}

	catIDs := map[string]bool{}
	for _, c := range candidates {
		if c.CategoryID != "" {
			catIDs[c.CategoryID] = true
		}
	}
	for _, cid := range historyCategories {
		catIDs[cid] = true
	}
	catList := make([]string, 0, len(catIDs))
	for id := range catIDs {
		catList = append(catList, id)
	}
	rels, err := r.presence.CategoryRelations(ctx, catList)
	if err != nil {
		return nil, err
	}
	relRows := make([]categoryRelation, 0, len(rels))
	for _, rel := range rels {
		relRows = append(relRows, categoryRelation{
			From:   rel.FromCategoryID,
			To:     rel.ToCategoryID,
			Type:   rel.RelationType,
			Weight: float64(rel.WeightPct) / 100,
		})
	}
	rules := buildCategoryRules(relRows, r.cfg.SameCategoryAffinity)

	scored := ScoreCandidates(r.cfg, history, historyCategories, eligible, singles, coVisits, totalGuests, rules)
	if len(scored) > 3 {
		scored = scored[:3]
	}

	out := make([]domain.Recommendation, 0, len(scored))
	for i, sc := range scored {
		rec := domain.Recommendation{
			ID:               r.idGen.NewID(ctx),
			GuestID:          guestID,
			TenantID:         sc.TenantID,
			Rank:             i + 1,
			Score:            sc.Score,
			Explanation:      sc.Explanation,
			AlgorithmVersion: r.cfg.AlgorithmVersion,
			ComputedAt:       now,
		}
		for _, f := range sc.Factors {
			rec.Factors = append(rec.Factors, domain.RecommendationFactor{
				RecommendationID: rec.ID,
				FactorCode:       f.Code,
				RawValue:         f.RawValue,
				Weight:           f.Weight,
				Contribution:     f.Weight * f.RawValue,
			})
		}
		out = append(out, rec)
	}
	if err := r.replace(ctx, guestID, out); err != nil {
		return nil, err
	}
	return out, nil
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
		for _, f := range it.Factors {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO recommendation_factor (recommendation_id, factor_code, raw_value, weight, contribution)
				VALUES ($1,$2,$3,$4,$5)`,
				it.ID, f.FactorCode, f.RawValue, f.Weight, f.Contribution,
			); err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}

func (r *WeightedRecommender) AggregateSegment(ctx context.Context, segmentID string, limit int) ([]domain.SegmentTenantScore, error) {
	members, err := r.presence.SegmentMembersWithPresence(ctx, segmentID)
	if err != nil {
		return nil, err
	}
	guestIDs := make([]string, 0, len(members))
	for _, m := range members {
		guestIDs = append(guestIDs, m.GuestID)
	}
	recs, err := r.presence.RecommendationsByGuests(ctx, guestIDs)
	if err != nil {
		return nil, err
	}

	cands, err := r.presence.ActiveTenants(ctx)
	if err != nil {
		return nil, err
	}
	meta := map[string]struct {
		Name     string
		Category string
	}{}
	for _, c := range cands {
		meta[c.ID] = struct {
			Name     string
			Category string
		}{c.Name, c.CategoryID}
	}

	now := r.clock.Now().UTC()
	staleBefore := now.Add(-r.cfg.StaleAfter)
	snaps := make([]SegmentMemberSnapshot, 0, len(members))
	for _, m := range members {
		memberRecs := recs[m.GuestID]
		converted := make([]MemberRecommendation, 0, len(memberRecs))
		fresh := false
		for _, rec := range memberRecs {
			converted = append(converted, MemberRecommendation{
				TenantID:   rec.TenantID,
				Score:      rec.Score,
				ComputedAt: rec.ComputedAt,
			})
			if !rec.ComputedAt.Before(staleBefore) {
				fresh = true
			}
		}
		snaps = append(snaps, SegmentMemberSnapshot{
			GuestID:  m.GuestID,
			Visited:  m.TenantIDs,
			Recs:     converted,
			HasFresh: fresh,
		})
	}

	aggs := AggregateSegmentMembers(snaps, limit)
	out := make([]domain.SegmentTenantScore, 0, len(aggs))
	for _, a := range aggs {
		name := a.TenantID
		cat := ""
		if m, ok := meta[a.TenantID]; ok {
			name = m.Name
			cat = m.Category
		}
		out = append(out, domain.SegmentTenantScore{
			TenantID:        a.TenantID,
			TenantName:      name,
			CategoryName:    cat,
			GuestCount:      a.GuestCount,
			EligibleCount:   a.EligibleCount,
			SegmentCount:    a.SegmentCount,
			AverageScore:    a.AverageScore,
			IncompleteCount: a.IncompleteCount,
		})
	}
	return out, nil
}

func (r *WeightedRecommender) AlgorithmVersion() string {
	return r.cfg.AlgorithmVersion
}

var _ domain.Recommender = (*WeightedRecommender)(nil)
