package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/MangriMen/UCRP/internal/domain"
)

type MallSQL struct {
	db *sql.DB
}

func NewMallSQL(db *sql.DB) *MallSQL {
	return &MallSQL{db: db}
}

const mallSelect = `
	SELECT m.id, m.code, m.name, m.is_active, m.floor_min, m.floor_max, m.has_zero_floor, m.created_at
	FROM mall m`

func (r *MallSQL) List(ctx context.Context) ([]domain.Mall, error) {
	rows, err := r.db.QueryContext(ctx, mallSelect+` ORDER BY m.code`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]domain.Mall, 0)
	for rows.Next() {
		m, err := scanMall(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *m)
	}
	return out, rows.Err()
}

func (r *MallSQL) GetByID(ctx context.Context, id string) (*domain.Mall, error) {
	m, err := scanMall(r.db.QueryRowContext(ctx, mallSelect+` WHERE m.id = $1`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return m, err
}

func (r *MallSQL) Insert(ctx context.Context, m *domain.Mall) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO mall (id, code, name, is_active, floor_min, floor_max, has_zero_floor, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		m.ID, m.Code, m.Name, m.IsActive, m.FloorMin, m.FloorMax, m.HasZero, m.CreatedAt,
	)
	if isDuplicate(err) {
		return domain.ErrConflict
	}
	return err
}

func (r *MallSQL) Update(ctx context.Context, m *domain.Mall) error {
	var outside int
	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM tenant
		WHERE mall_id = $1 AND deleted_at IS NULL AND floor IS NOT NULL
				AND (floor < $2 OR floor > $3 OR (floor = 0 AND NOT $4))`,
		m.ID, m.FloorMin, m.FloorMax, m.HasZero).Scan(&outside); err != nil {
		return err
	}
	if outside > 0 {
		return domain.ErrConflict
	}
	res, err := r.db.ExecContext(ctx, `
		UPDATE mall SET name = $2, is_active = $3, floor_min = $4, floor_max = $5, has_zero_floor = $6 WHERE id = $1`,
		m.ID, m.Name, m.IsActive, m.FloorMin, m.FloorMax, m.HasZero)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *MallSQL) Delete(ctx context.Context, id string) error {
	var tenants int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM tenant WHERE mall_id = $1`, id).Scan(&tenants); err != nil {
		return err
	}
	if tenants > 0 {
		return domain.ErrConflict
	}
	res, err := r.db.ExecContext(ctx, `DELETE FROM mall WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func scanMall(s scanner) (*domain.Mall, error) {
	var m domain.Mall
	if err := s.Scan(&m.ID, &m.Code, &m.Name, &m.IsActive, &m.FloorMin, &m.FloorMax, &m.HasZero, &m.CreatedAt); err != nil {
		return nil, err
	}
	return &m, nil
}
