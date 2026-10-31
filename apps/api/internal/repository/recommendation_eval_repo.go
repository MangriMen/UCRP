package repository

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/MangriMen/UCRP/internal/domain"
)

type RecommendationEvalSQL struct {
	db *sql.DB
}

func NewRecommendationEvalSQL(db *sql.DB) *RecommendationEvalSQL {
	return &RecommendationEvalSQL{db: db}
}

func (r *RecommendationEvalSQL) Insert(ctx context.Context, ev *domain.RecommendationEval) error {
	probabilities, err := json.Marshal(ev.Probabilities)
	if err != nil {
		return err
	}
	ranking, err := json.Marshal(ev.Ranking)
	if err != nil {
		return err
	}
	var confidence any
	if ev.Confidence != nil {
		confidence = *ev.Confidence
	}
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO recommendation_eval (id, guest_id, probabilities, ranking, confidence, recommendation_computed_at, evaluated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		ev.ID, ev.GuestID, string(probabilities), string(ranking), confidence, ev.RecommendationComputedAt, ev.EvaluatedAt,
	)
	return err
}
