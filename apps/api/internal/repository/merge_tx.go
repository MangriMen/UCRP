package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/MangriMen/UCRP/internal/domain"
)

type MergeSQL struct {
	db *sql.DB
}

func NewMergeSQL(db *sql.DB) *MergeSQL {
	return &MergeSQL{db: db}
}

func (r *MergeSQL) Confirm(ctx context.Context, candidateID, actorUserID string, now time.Time) (*domain.MergeCandidate, error) {
	c, err := r.candidate(ctx, candidateID)
	if err != nil {
		return nil, err
	}
	if c.Resolution != domain.MergePending {
		return nil, domain.ErrConflict
	}
	blocked, err := r.hasBlock(ctx, c.GuestAID, c.GuestBID)
	if err != nil {
		return nil, err
	}
	if blocked {
		return nil, domain.ErrConflict
	}
	err = r.tx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `
			UPDATE guest SET status = 'merged', merged_into_id = $2, updated_at = $3 WHERE id = $1`,
			c.GuestBID, c.GuestAID, now); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE event SET origin_guest_id = guest_id WHERE guest_id = $1 AND origin_guest_id IS NULL`,
			c.GuestBID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE event SET guest_id = $2 WHERE guest_id = $1`,
			c.GuestBID, c.GuestAID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			DELETE FROM guest_identifier
			WHERE guest_id = $2
			  AND origin_guest_id = $2
			  AND EXISTS (
				SELECT 1 FROM guest_identifier w
				WHERE w.guest_id = $1 AND w.id_type = guest_identifier.id_type
				  AND w.value_normalized = guest_identifier.value_normalized
			  )`,
			c.GuestAID, c.GuestBID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE guest_identifier SET guest_id = $2
			WHERE guest_id = $1 AND origin_guest_id = $1`,
			c.GuestBID, c.GuestAID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE merge_candidate
			SET resolution = 'confirmed', winner_guest_id = $2, merged_guest_id = $3, decided_by = $4, decided_at = $5
			WHERE id = $1 AND resolution = 'pending'`,
			c.ID, c.GuestAID, c.GuestBID, actorUserID, now); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO audit_log (id, actor_user_id, actor_role, action, entity_type, entity_id, details, request_id, created_at)
			VALUES (gen_random_uuid()::text, $1, 'analyst', 'merge.confirm', 'merge_candidate', $2, '{}'::text, NULL, $3)`,
			actorUserID, c.ID, now); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return r.candidate(ctx, candidateID)
}

func (r *MergeSQL) Reject(ctx context.Context, candidateID, actorUserID string, now time.Time) (*domain.MergeCandidate, error) {
	c, err := r.candidate(ctx, candidateID)
	if err != nil {
		return nil, err
	}
	if c.Resolution != domain.MergePending {
		return nil, domain.ErrConflict
	}
	err = r.tx(ctx, func(tx *sql.Tx) error {
		if err := insertBlockTx(ctx, tx, c, "rejected", actorUserID, now, r.idGen()); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE merge_candidate
			SET resolution = 'rejected', decided_by = $2, decided_at = $3
			WHERE id = $1 AND resolution = 'pending'`,
			c.ID, actorUserID, now); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx, `
			INSERT INTO audit_log (id, actor_user_id, actor_role, action, entity_type, entity_id, details, request_id, created_at)
			VALUES (gen_random_uuid()::text, $1, 'analyst', 'merge.reject', 'merge_candidate', $2, '{}'::text, NULL, $3)`,
			actorUserID, c.ID, now)
		return err
	})
	if err != nil {
		return nil, err
	}
	return r.candidate(ctx, candidateID)
}

func (r *MergeSQL) Revert(ctx context.Context, candidateID, actorUserID string, now time.Time) (*domain.MergeCandidate, error) {
	c, err := r.candidate(ctx, candidateID)
	if err != nil {
		return nil, err
	}
	if c.Resolution != domain.MergeConfirmed && c.Resolution != domain.MergeAutoMerged {
		return nil, domain.ErrConflict
	}
	if c.MergedGuestID == "" {
		return nil, domain.ErrConflict
	}
	err = r.tx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `
			UPDATE event SET guest_id = origin_guest_id
			WHERE origin_guest_id = $1 AND guest_id <> origin_guest_id`,
			c.MergedGuestID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE guest_identifier SET guest_id = origin_guest_id
			WHERE origin_guest_id = $1 AND guest_id <> origin_guest_id`,
			c.MergedGuestID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE guest SET status = 'active', merged_into_id = NULL, updated_at = $2 WHERE id = $1`,
			c.MergedGuestID, now); err != nil {
			return err
		}
		if err := insertBlockTx(ctx, tx, c, "manual_unmerge", actorUserID, now, r.idGen()); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE merge_candidate
			SET resolution = 'reverted', decided_by = $2, decided_at = $3
			WHERE id = $1`,
			c.ID, actorUserID, now); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx, `
			INSERT INTO audit_log (id, actor_user_id, actor_role, action, entity_type, entity_id, details, request_id, created_at)
			VALUES (gen_random_uuid()::text, $1, 'analyst', 'merge.revert', 'merge_candidate', $2, '{}'::text, NULL, $3)`,
			actorUserID, c.ID, now)
		return err
	})
	if err != nil {
		return nil, err
	}
	return r.candidate(ctx, candidateID)
}

func (r *MergeSQL) GetByID(ctx context.Context, id string) (*domain.MergeCandidate, error) {
	return r.candidate(ctx, id)
}

func (r *MergeSQL) List(ctx context.Context, resolution domain.MergeResolution, limit, offset int) ([]domain.MergeCandidate, error) {
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
	defer func() { _ = rows.Close() }()
	out := make([]domain.MergeCandidate, 0)
	for rows.Next() {
		c, err := scanCandidate(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *c)
	}
	return out, rows.Err()
}

func (r *MergeSQL) candidate(ctx context.Context, id string) (*domain.MergeCandidate, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, guest_a_id, guest_b_id, match_type, reason_code, reason_text, score, resolution,
			COALESCE(winner_guest_id, ''), COALESCE(merged_guest_id, ''), COALESCE(decided_by, ''), decided_at, created_at
		FROM merge_candidate WHERE id = $1`, id)
	c, err := scanCandidateRow(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return c, err
}

func (r *MergeSQL) hasBlock(ctx context.Context, guestA, guestB string) (bool, error) {
	var n int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM merge_block
		WHERE (guest_a_id = $1 AND guest_b_id = $2) OR (guest_a_id = $2 AND guest_b_id = $1)`,
		guestA, guestB).Scan(&n)
	return n > 0, err
}

func (r *MergeSQL) tx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *MergeSQL) idGen() string {
	var s string
	err := r.db.QueryRowContext(context.Background(), `SELECT gen_random_uuid()::text`).Scan(&s)
	if err != nil {
		return "unknown"
	}
	return s
}

func insertBlockTx(ctx context.Context, tx *sql.Tx, c *domain.MergeCandidate, reason, actor string, now time.Time, id string) error {
	a, b := c.GuestAID, c.GuestBID
	if a > b {
		a, b = b, a
	}
	var n int
	if err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM merge_block
		WHERE (guest_a_id = $1 AND guest_b_id = $2) OR (guest_a_id = $2 AND guest_b_id = $1)`, a, b).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO merge_block (id, guest_a_id, guest_b_id, reason, candidate_id, created_by, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`, id, a, b, reason, c.ID, actor, now)
	return err
}

type scanner interface {
	Scan(dest ...any) error
}

func scanCandidateRow(row scanner) (*domain.MergeCandidate, error) {
	var (
		c   domain.MergeCandidate
		res string
		at  sql.NullTime
	)
	if err := row.Scan(
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
	return &c, nil
}

func scanCandidate(rows *sql.Rows) (*domain.MergeCandidate, error) {
	return scanCandidateRow(rows)
}
