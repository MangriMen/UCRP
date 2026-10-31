package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/MangriMen/UCRP/internal/domain"
)

type MergeCandidateSQL struct {
	db *sql.DB
}

func NewMergeCandidateSQL(db *sql.DB) *MergeCandidateSQL {
	return &MergeCandidateSQL{db: db}
}

func (r *MergeCandidateSQL) GetByID(ctx context.Context, id string) (*domain.MergeCandidate, error) {
	var (
		c         domain.MergeCandidate
		res       string
		winner    sql.NullString
		merged    sql.NullString
		decidedBy sql.NullString
		decidedAt sql.NullTime
	)
	err := r.db.QueryRowContext(ctx, `
		SELECT id, guest_a_id, guest_b_id, match_type, reason_code, reason_text, score, resolution,
			winner_guest_id, merged_guest_id, decided_by, decided_at, created_at
		FROM merge_candidate WHERE id = $1`, id).Scan(
		&c.ID, &c.GuestAID, &c.GuestBID, &c.MatchType, &c.ReasonCode, &c.ReasonText, &c.Score, &res,
		&winner, &merged, &decidedBy, &decidedAt, &c.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	c.Resolution = domain.MergeResolution(res)
	c.WinnerGuestID = winner.String
	c.MergedGuestID = merged.String
	c.DecidedBy = decidedBy.String
	if decidedAt.Valid {
		t := decidedAt.Time
		c.DecidedAt = &t
	}
	return &c, nil
}

func (r *MergeCandidateSQL) List(ctx context.Context, resolution domain.MergeResolution, limit, offset int) ([]domain.MergeCandidate, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, guest_a_id, guest_b_id, match_type, reason_code, reason_text, score, resolution,
			COALESCE(winner_guest_id, ''), COALESCE(merged_guest_id, ''), COALESCE(decided_by, ''), decided_at, created_at
		FROM merge_candidate
		WHERE ($1 = '' OR resolution = $1)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`, string(resolution), limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.MergeCandidate, 0)
	for rows.Next() {
		var (
			c   domain.MergeCandidate
			res string
			at  sql.NullTime
		)
		if err := rows.Scan(
			&c.ID, &c.GuestAID, &c.GuestBID, &c.MatchType, &c.ReasonCode, &c.ReasonText, &c.Score, &res,
			&c.WinnerGuestID, &c.MergedGuestID, &c.DecidedBy, &at, &c.CreatedAt,
		); err != nil {
			return nil, err
		}
		c.Resolution = domain.MergeResolution(res)
		if at.Valid {
			t := at.Time
			c.DecidedAt = &t
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *MergeCandidateSQL) Insert(ctx context.Context, c *domain.MergeCandidate) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO merge_candidate (
			id, guest_a_id, guest_b_id, match_type, reason_code, reason_text, score, resolution,
			winner_guest_id, merged_guest_id, decided_by, decided_at, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		c.ID, c.GuestAID, c.GuestBID, c.MatchType, c.ReasonCode, c.ReasonText, c.Score, string(c.Resolution),
		nullIfEmpty(c.WinnerGuestID), nullIfEmpty(c.MergedGuestID), nullIfEmpty(c.DecidedBy), c.DecidedAt, c.CreatedAt,
	)
	return err
}

func (r *MergeCandidateSQL) UpdateDecision(ctx context.Context, c *domain.MergeCandidate) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE merge_candidate
		SET resolution = $2, winner_guest_id = $3, merged_guest_id = $4, decided_by = $5, decided_at = $6
		WHERE id = $1`,
		c.ID, string(c.Resolution), nullIfEmpty(c.WinnerGuestID), nullIfEmpty(c.MergedGuestID), nullIfEmpty(c.DecidedBy), c.DecidedAt,
	)
	return err
}

type UnmergeBlockSQL struct {
	db *sql.DB
}

func NewUnmergeBlockSQL(db *sql.DB) *UnmergeBlockSQL {
	return &UnmergeBlockSQL{db: db}
}

func (r *UnmergeBlockSQL) HasBlock(ctx context.Context, guestA, guestB string) (bool, error) {
	var n int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM unmerge_block
		WHERE (guest_a_id = $1 AND guest_b_id = $2) OR (guest_a_id = $2 AND guest_b_id = $1)`,
		guestA, guestB).Scan(&n)
	return n > 0, err
}

func (r *UnmergeBlockSQL) Insert(ctx context.Context, b *domain.UnmergeBlock) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO unmerge_block (id, guest_a_id, guest_b_id, reason, candidate_id, created_by, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		b.ID, b.GuestAID, b.GuestBID, b.Reason, nullIfEmpty(b.CandidateID), b.CreatedBy, b.CreatedAt,
	)
	return err
}
