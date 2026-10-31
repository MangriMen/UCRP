package seed

import (
	"context"
	"fmt"

	"github.com/MangriMen/UCRP/internal/recommend"
)

func (s *Seeder) seedRecommendations(ctx context.Context) error {
	recommender := recommend.NewWeightedRecommender(s.db)
	keys := make([]string, 0, s.opts.Guests+len(demoPersonas()))
	for i := 0; i < s.opts.Guests; i++ {
		keys = append(keys, backgroundGuestKey(i))
	}
	for _, p := range demoPersonas() {
		keys = append(keys, p.Key)
	}
	for _, key := range keys {
		guestID := s.guests[key]
		if guestID == "" {
			return fmt.Errorf("recommendation for unknown guest %s", key)
		}
		if _, err := recommender.Refresh(ctx, guestID); err != nil {
			return fmt.Errorf("refresh %s: %w", key, err)
		}
	}
	return s.seedRecommendationFactors(ctx)
}

type factorSpec struct {
	Code   string
	Raw    float64
	Weight float64
}

func (s *Seeder) seedRecommendationFactors(ctx context.Context) error {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id FROM recommendation ORDER BY guest_id, rank LIMIT 30`)
	if err != nil {
		return err
	}
	ids := make([]string, 0, 30)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			_ = rows.Close()
			return err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return err
	}
	_ = rows.Close()

	factors := []factorSpec{
		{"similar_guests", 0.85, 0.40},
		{"category_adjacency", 0.60, 0.25},
		{"recency", 0.70, 0.25},
		{"distance", 0.45, 0.10},
	}
	for _, id := range ids {
		for _, f := range factors {
			if err := s.exec(ctx, `
				INSERT INTO recommendation_factor (recommendation_id, factor_code, raw_value, weight, contribution)
				VALUES ($1,$2,$3,$4,$5)
				ON CONFLICT (recommendation_id, factor_code) DO NOTHING`,
				id, f.Code, f.Raw, f.Weight, f.Raw*f.Weight); err != nil {
				return err
			}
		}
	}
	return nil
}
