package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/MangriMen/UCRP/internal/domain"
)

type SegmentSQL struct {
	db *sql.DB
}

func NewSegmentSQL(db *sql.DB) *SegmentSQL {
	return &SegmentSQL{db: db}
}

func (r *SegmentSQL) EnsureTemplates(ctx context.Context, clock domain.Clock, idGen domain.IDGen) error {
	templates := []struct {
		code, name, desc string
		was              int
	}{
		{"sleep_neighbor", "Спящий сосед", "Был у одного арендатора, не был у соседей", 1},
		{"single_tenant", "Один магазин", "Посещал только один арендатор", 1},
		{"churn_risk", "Риск ухода", "Давно не посещал выбранных арендаторов", 1},
		{"newcomer", "Новичок", "Мало истории посещений", 1},
		{"high_value", "Высокая ценность", "Высокая частота посещений", 1},
	}
	now := clock.Now()
	for _, t := range templates {
		var exists int
		err := r.db.QueryRowContext(ctx, `SELECT 1 FROM segment WHERE template_code = $1`, t.code).Scan(&exists)
		if err == nil {
			continue
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		segID := idGen.NewID(ctx)
		rule := `{"was":[],"was_not":[],"period_days":30}`
		tx, err := r.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO segment (id, name, description, template_code, is_template, current_version_no, status, created_by, created_at, updated_at)
			VALUES ($1,$2,$3,$4,TRUE,1,'active',NULL,$5,$5)`,
			segID, t.name, t.desc, t.code, now); err != nil {
			_ = tx.Rollback()
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO segment_version (id, segment_id, version_no, rule, status, member_count, computed_at, created_by, created_at)
			VALUES ($1,$2,1,$3,'ready',0,$4,NULL,$4)`,
			idGen.NewID(ctx), segID, rule, now); err != nil {
			_ = tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}

func (r *SegmentSQL) Insert(ctx context.Context, s *domain.Segment, v *domain.SegmentVersion) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	_, err = tx.ExecContext(ctx, `
		INSERT INTO segment (id, name, description, template_code, is_template, current_version_no, status, created_by, created_at, updated_at)
		VALUES ($1,$2,$3,NULLIF($4,''),$5,$6,$7,$8,$9,$10)`,
		s.ID, s.Name, nullIfEmpty(s.Description), s.TemplateCode, s.IsTemplate, s.CurrentVersionNo, string(s.Status), nullIfEmpty(s.CreatedBy), s.CreatedAt, s.UpdatedAt,
	)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO segment_version (id, segment_id, version_no, rule, status, member_count, computed_at, created_by, created_at)
		VALUES ($1,$2,$3,$4,$5,NULL,NULL,$6,$7)`,
		v.ID, v.SegmentID, v.VersionNo, v.Rule, string(v.Status), nullIfEmpty(v.CreatedBy), v.CreatedAt,
	)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *SegmentSQL) GetByID(ctx context.Context, id string) (*domain.Segment, *domain.SegmentVersion, error) {
	var (
		s     domain.Segment
		v     domain.SegmentVersion
		desc  sql.NullString
		tmpl  sql.NullString
		sstat string
		vstat string
	)
	err := r.db.QueryRowContext(ctx, `
		SELECT s.id, s.name, COALESCE(s.description, ''), COALESCE(s.template_code, ''), s.is_template, s.current_version_no, s.status, COALESCE(s.created_by, ''), s.created_at, s.updated_at,
			v.id, v.version_no, v.rule, v.status, v.member_count, v.computed_at, COALESCE(v.created_by, ''), v.created_at
		FROM segment s
		JOIN segment_version v ON v.segment_id = s.id AND v.version_no = s.current_version_no
		WHERE s.id = $1`, id).Scan(
		&s.ID, &s.Name, &desc, &tmpl, &s.IsTemplate, &s.CurrentVersionNo, &sstat, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt,
		&v.ID, &v.VersionNo, &v.Rule, &vstat, &v.MemberCount, &v.ComputedAt, &v.CreatedBy, &v.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, nil, err
	}
	s.Status = domain.SegmentStatus(sstat)
	v.Status = domain.SegmentVersionStatus(vstat)
	v.SegmentID = s.ID
	return &s, &v, nil
}

func (r *SegmentSQL) List(ctx context.Context, limit, offset int) ([]domain.Segment, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, COALESCE(description, ''), COALESCE(template_code, ''), is_template, current_version_no, status, COALESCE(created_by, ''), created_at, updated_at
		FROM segment
		WHERE status <> 'archived'
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]domain.Segment, 0)
	for rows.Next() {
		var (
			s    domain.Segment
			desc string
			tmpl string
			stat string
		)
		if err := rows.Scan(&s.ID, &s.Name, &desc, &tmpl, &s.IsTemplate, &s.CurrentVersionNo, &stat, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		s.Description = desc
		s.TemplateCode = tmpl
		s.Status = domain.SegmentStatus(stat)
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *SegmentSQL) ReplaceMembers(ctx context.Context, versionID string, guestIDs []string, computedAt time.Time) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `DELETE FROM segment_member WHERE segment_version_id = $1`, versionID); err != nil {
		return err
	}
	for _, gid := range guestIDs {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO segment_member (segment_version_id, guest_id, added_at)
			VALUES ($1, $2, $3)`, versionID, gid, computedAt); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE segment_version
		SET status = 'ready', member_count = $2, computed_at = $3
		WHERE id = $1`, versionID, len(guestIDs), computedAt); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *SegmentSQL) ListTemplates(ctx context.Context) ([]domain.Segment, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, COALESCE(description, ''), COALESCE(template_code, ''), is_template, current_version_no, status, COALESCE(created_by, ''), created_at, updated_at
		FROM segment WHERE is_template AND status <> 'archived'
		ORDER BY template_code`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]domain.Segment, 0)
	for rows.Next() {
		var (
			s    domain.Segment
			desc string
			tmpl string
			stat string
		)
		if err := rows.Scan(&s.ID, &s.Name, &desc, &tmpl, &s.IsTemplate, &s.CurrentVersionNo, &stat, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		s.Description = desc
		s.TemplateCode = tmpl
		s.Status = domain.SegmentStatus(stat)
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *SegmentSQL) ListVersions(ctx context.Context, segmentID string) ([]domain.SegmentVersion, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, segment_id, version_no, rule, status, member_count, computed_at, COALESCE(created_by, ''), created_at
		FROM segment_version WHERE segment_id = $1 ORDER BY version_no DESC`, segmentID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]domain.SegmentVersion, 0)
	for rows.Next() {
		var (
			v     domain.SegmentVersion
			stat  string
			count sql.NullInt64
		)
		if err := rows.Scan(&v.ID, &v.SegmentID, &v.VersionNo, &v.Rule, &stat, &count, &v.ComputedAt, &v.CreatedBy, &v.CreatedAt); err != nil {
			return nil, err
		}
		v.Status = domain.SegmentVersionStatus(stat)
		if count.Valid {
			n := int(count.Int64)
			v.MemberCount = &n
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (r *SegmentSQL) GetVersion(ctx context.Context, segmentID string, versionNo int) (*domain.Segment, *domain.SegmentVersion, error) {
	var (
		s     domain.Segment
		v     domain.SegmentVersion
		desc  sql.NullString
		tmpl  sql.NullString
		sstat string
		vstat string
	)
	err := r.db.QueryRowContext(ctx, `
		SELECT s.id, s.name, COALESCE(s.description, ''), COALESCE(s.template_code, ''), s.is_template, s.current_version_no, s.status, COALESCE(s.created_by, ''), s.created_at, s.updated_at,
			v.id, v.version_no, v.rule, v.status, v.member_count, v.computed_at, COALESCE(v.created_by, ''), v.created_at
		FROM segment s
		JOIN segment_version v ON v.segment_id = s.id AND v.version_no = $2
		WHERE s.id = $1`, segmentID, versionNo).Scan(
		&s.ID, &s.Name, &desc, &tmpl, &s.IsTemplate, &s.CurrentVersionNo, &sstat, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt,
		&v.ID, &v.VersionNo, &v.Rule, &vstat, &v.MemberCount, &v.ComputedAt, &v.CreatedBy, &v.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, nil, err
	}
	s.Status = domain.SegmentStatus(sstat)
	v.Status = domain.SegmentVersionStatus(vstat)
	v.SegmentID = s.ID
	return &s, &v, nil
}

func (r *SegmentSQL) InsertVersion(ctx context.Context, v *domain.SegmentVersion) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO segment_version (id, segment_id, version_no, rule, status, member_count, computed_at, created_by, created_at)
		VALUES ($1,$2,$3,$4,$5,NULL,NULL,$6,$7)`,
		v.ID, v.SegmentID, v.VersionNo, v.Rule, string(v.Status), nullIfEmpty(v.CreatedBy), v.CreatedAt,
	); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE segment SET current_version_no = $2, updated_at = $3 WHERE id = $1`,
		v.SegmentID, v.VersionNo, v.CreatedAt,
	); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *SegmentSQL) Archive(ctx context.Context, id string, now time.Time) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE segment SET status = 'archived', updated_at = $2 WHERE id = $1 AND status <> 'archived'`, id, now)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *SegmentSQL) MarkComputing(ctx context.Context, versionID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE segment_version SET status = 'computing' WHERE id = $1`, versionID)
	return err
}

func (r *SegmentSQL) ListMembers(ctx context.Context, versionID string, limit, offset int) ([]string, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT guest_id FROM segment_member WHERE segment_version_id = $1
		ORDER BY guest_id LIMIT $2 OFFSET $3`, versionID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

func (r *SegmentSQL) ListMemberContacts(ctx context.Context, versionID string) ([]domain.GuestContact, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT m.guest_id, i.id, i.value_normalized, i.id_type
		FROM segment_member m
		JOIN guest_identifier i ON i.guest_id = m.guest_id
		WHERE m.segment_version_id = $1
		ORDER BY m.guest_id`, versionID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]domain.GuestContact, 0)
	for rows.Next() {
		var c domain.GuestContact
		if err := rows.Scan(&c.GuestID, &c.IdentifierID, &c.Contact, &c.IDType); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
