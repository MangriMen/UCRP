package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/MangriMen/UCRP/internal/domain"
	"github.com/google/uuid"
)

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
	a, c := b.GuestAID, b.GuestBID
	if a > c {
		a, c = c, a
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO unmerge_block (id, guest_a_id, guest_b_id, reason, candidate_id, created_by, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT DO NOTHING`,
		b.ID, a, c, b.Reason, nullIfEmpty(b.CandidateID), nullIfEmpty(b.CreatedBy), b.CreatedAt,
	)
	return err
}

func (r *MergeSQL) CreatePending(ctx context.Context, guestA, guestB, matchType, reasonCode, reasonText string, score int, now time.Time) (*domain.MergeCandidate, error) {
	a, b := guestA, guestB
	if a > b {
		a, b = b, a
	}
	var existing string
	err := r.db.QueryRowContext(ctx, `
		SELECT id FROM merge_candidate WHERE guest_a_id = $1 AND guest_b_id = $2`, a, b).Scan(&existing)
	if err == nil {
		return r.candidate(ctx, existing)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	id := uuid.NewString()
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO merge_candidate (
			id, guest_a_id, guest_b_id, match_type, reason_code, reason_text, score, resolution, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,'pending',$8)`,
		id, a, b, matchType, reasonCode, reasonText, score, now,
	)
	if err != nil {
		return nil, err
	}
	return r.candidate(ctx, id)
}

func (r *MergeSQL) AutoMerge(ctx context.Context, guestA, guestB, matchType, reasonCode, reasonText string, score int, now time.Time) (*domain.MergeCandidate, error) {
	a, b := guestA, guestB
	if a > b {
		a, b = b, a
	}
	c, err := r.CreatePending(ctx, a, b, matchType, reasonCode, reasonText, score, now)
	if err != nil {
		return nil, err
	}
	if c.Resolution != domain.MergePending {
		return c, nil
	}
	err = r.tx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `
			UPDATE guest SET status = 'merged', merged_into_id = $2, updated_at = $3 WHERE id = $1 AND status = 'active'`,
			b, a, now); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE event SET origin_guest_id = guest_id WHERE guest_id = $1 AND origin_guest_id IS NULL`,
			b); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE event SET guest_id = $2 WHERE guest_id = $1`,
			b, a); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			DELETE FROM guest_identifier
			WHERE guest_id = $2 AND origin_guest_id = $2
			  AND EXISTS (
				SELECT 1 FROM guest_identifier w
				WHERE w.guest_id = $1 AND w.id_type = guest_identifier.id_type
				  AND w.value_normalized = guest_identifier.value_normalized
			  )`,
			a, b); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE guest_identifier SET guest_id = $2
			WHERE guest_id = $1 AND origin_guest_id = $1`,
			b, a); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE merge_candidate
			SET resolution = 'auto_merged', winner_guest_id = $2, merged_guest_id = $3, decided_at = $4
			WHERE id = $1 AND resolution = 'pending'`,
			c.ID, a, b, now); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx, `
			INSERT INTO audit_log (id, actor_user_id, actor_role, action, entity_type, entity_id, details, request_id, created_at)
			VALUES (gen_random_uuid()::text, NULL, 'system', 'merge.auto', 'merge_candidate', $1, '{}'::text, NULL, $2)`,
			c.ID, now)
		return err
	})
	if err != nil {
		return nil, err
	}
	return r.candidate(ctx, c.ID)
}
