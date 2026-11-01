package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

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

func (r *RecommendationEvalSQL) LatestForGuest(ctx context.Context, guestID string) (*domain.RecommendationEval, error) {
	var (
		ev         domain.RecommendationEval
		probRaw    string
		rankRaw    string
		confidence sql.NullFloat64
	)
	err := r.db.QueryRowContext(ctx, `
		SELECT id, guest_id, probabilities, ranking, confidence, recommendation_computed_at, evaluated_at
		FROM recommendation_eval
		WHERE guest_id = $1
		ORDER BY evaluated_at DESC
		LIMIT 1`, guestID).Scan(
		&ev.ID, &ev.GuestID, &probRaw, &rankRaw, &confidence, &ev.RecommendationComputedAt, &ev.EvaluatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(probRaw), &ev.Probabilities); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(rankRaw), &ev.Ranking); err != nil {
		return nil, err
	}
	if confidence.Valid {
		v := confidence.Float64
		ev.Confidence = &v
	}
	return &ev, nil
}
