package repository

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/MangriMen/UCRP/internal/domain"
)

type ExportSQL struct {
	db *sql.DB
}

func NewExportSQL(db *sql.DB) *ExportSQL {
	return &ExportSQL{db: db}
}

const exportSelect = `
	SELECT e.id, e.segment_id, e.segment_version_id, COALESCE(e.campaign_id, ''), e.channel, e.format, e.control_percent, e.control_seed,
		e.idempotency_key, e.status, e.total_in_segment, e.no_contact_count, e.blocked_by_limit_count, e.control_count, e.exported_count,
		e.created_by, e.created_at, e.finished_at, COALESCE(s.name, ''), COALESCE(v.version_no, 0)
	FROM export_job e
	LEFT JOIN segment s ON s.id = e.segment_id
	LEFT JOIN segment_version v ON v.id = e.segment_version_id`

func (r *ExportSQL) FindByKey(ctx context.Context, userID, key string) (*domain.ExportJob, error) {
	return r.one(ctx, exportSelect+` WHERE e.created_by = $1 AND e.idempotency_key = $2`, userID, key)
}

func (r *ExportSQL) GetByID(ctx context.Context, id string) (*domain.ExportJob, error) {
	return r.one(ctx, exportSelect+` WHERE e.id = $1`, id)
}

func (r *ExportSQL) ListMembers(ctx context.Context, exportID string) ([]domain.ExportMember, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT m.export_id, m.guest_id, m.identifier_id, m.is_control, COALESCE(i.value_normalized, '')
		FROM export_member m
		LEFT JOIN guest_identifier i ON i.id = m.identifier_id
		WHERE m.export_id = $1
		ORDER BY m.is_control, m.guest_id`, exportID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]domain.ExportMember, 0)
	for rows.Next() {
		var m domain.ExportMember
		if err := rows.Scan(&m.ExportID, &m.GuestID, &m.IdentifierID, &m.IsControl, &m.Contact); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *ExportSQL) ListMemberPage(ctx context.Context, exportID string, isControl *bool, limit, offset int) ([]domain.ExportMember, int, error) {
	limit, offset = clampPage(limit, offset)
	where := ` WHERE m.export_id = $1`
	args := []any{exportID}
	if isControl != nil {
		args = append(args, *isControl)
		where += ` AND m.is_control = $2`
	}
	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM export_member m`+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, `
		SELECT m.export_id, m.guest_id, m.identifier_id, m.is_control, COALESCE(i.value_normalized, ''), COALESCE(i.id_type, '')
		FROM export_member m
		LEFT JOIN guest_identifier i ON i.id = m.identifier_id`+where+`
		ORDER BY m.is_control, m.guest_id
		LIMIT $`+strconv.Itoa(len(args)-1)+` OFFSET $`+strconv.Itoa(len(args)), args...)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]domain.ExportMember, 0)
	for rows.Next() {
		var m domain.ExportMember
		if err := rows.Scan(&m.ExportID, &m.GuestID, &m.IdentifierID, &m.IsControl, &m.Contact, &m.ContactType); err != nil {
			return nil, 0, err
		}
		out = append(out, m)
	}
	return out, total, rows.Err()
}

func (r *ExportSQL) Insert(ctx context.Context, job *domain.ExportJob, members []domain.ExportMember) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	_, err = tx.ExecContext(ctx, `
		INSERT INTO export_job (
			id, segment_id, segment_version_id, campaign_id, channel, format, control_percent, control_seed,
			idempotency_key, status, total_in_segment, no_contact_count, blocked_by_limit_count, control_count,
			exported_count, created_by, created_at, finished_at
		) VALUES ($1,$2,$3,NULLIF($4,''),$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)`,
		job.ID, job.SegmentID, job.SegmentVersionID, job.CampaignID, string(job.Channel), string(job.Format),
		job.ControlPercent, job.ControlSeed, job.IdempotencyKey, string(job.Status),
		job.TotalInSegment, job.NoContactCount, job.BlockedByLimitCount, job.ControlCount, job.ExportedCount,
		job.CreatedBy, job.CreatedAt, job.FinishedAt,
	)
	if err != nil {
		return err
	}
	for _, m := range members {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO export_member (export_id, guest_id, identifier_id, is_control)
			VALUES ($1,$2,$3,$4)`, m.ExportID, m.GuestID, m.IdentifierID, m.IsControl); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *ExportSQL) List(ctx context.Context, filter domain.ExportFilter, limit, offset int) ([]domain.ExportJob, int, error) {
	limit, offset = clampPage(limit, offset)
	var conditions []string
	var args []any
	for _, condition := range []struct {
		column string
		value  string
	}{
		{"e.segment_id", filter.SegmentID},
		{"e.campaign_id", filter.CampaignID},
		{"e.channel", filter.Channel},
		{"e.status", filter.Status},
	} {
		if condition.value == "" {
			continue
		}
		args = append(args, condition.value)
		conditions = append(conditions, condition.column+" = $"+strconv.Itoa(len(args)))
	}
	where := ""
	if len(conditions) > 0 {
		where = " WHERE " + strings.Join(conditions, " AND ")
	}
	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM export_job e`+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, exportSelect+where+
		` ORDER BY e.created_at DESC, e.id LIMIT $`+strconv.Itoa(len(args)-1)+` OFFSET $`+strconv.Itoa(len(args)), args...)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]domain.ExportJob, 0)
	for rows.Next() {
		j, err := scanExport(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *j)
	}
	return out, total, rows.Err()
}

func clampPage(limit, offset int) (int, int) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	return limit, max(offset, 0)
}

func (r *ExportSQL) one(ctx context.Context, query string, args ...any) (*domain.ExportJob, error) {
	j, err := scanExport(r.db.QueryRowContext(ctx, query, args...))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return j, err
}

func scanExport(row interface{ Scan(dest ...any) error }) (*domain.ExportJob, error) {
	var (
		j       domain.ExportJob
		channel string
		format  string
		status  string
		fin     sql.NullTime
	)
	if err := row.Scan(
		&j.ID, &j.SegmentID, &j.SegmentVersionID, &j.CampaignID, &channel, &format, &j.ControlPercent, &j.ControlSeed,
		&j.IdempotencyKey, &status, &j.TotalInSegment, &j.NoContactCount, &j.BlockedByLimitCount, &j.ControlCount, &j.ExportedCount,
		&j.CreatedBy, &j.CreatedAt, &fin, &j.SegmentName, &j.SegmentVersionNo,
	); err != nil {
		return nil, err
	}
	j.Channel = domain.ExportChannel(channel)
	j.Format = domain.ExportFormat(format)
	j.Status = domain.ExportStatus(status)
	if fin.Valid {
		t := fin.Time
		j.FinishedAt = &t
	}
	return &j, nil
}

type ContactLimitSQL struct {
	db *sql.DB
}

func NewContactLimitSQL(db *sql.DB) *ContactLimitSQL {
	return &ContactLimitSQL{db: db}
}

func (r *ContactLimitSQL) List(ctx context.Context, limit, offset int) ([]domain.ContactLimit, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, COALESCE(channel, ''), max_contacts, period_days, is_active, COALESCE(created_by, ''), created_at, updated_at
		FROM contact_limit ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]domain.ContactLimit, 0)
	for rows.Next() {
		var c domain.ContactLimit
		if err := rows.Scan(&c.ID, &c.Channel, &c.MaxContacts, &c.PeriodDays, &c.IsActive, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *ContactLimitSQL) ListActive(ctx context.Context) ([]domain.ContactLimit, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, COALESCE(channel, ''), max_contacts, period_days, is_active, COALESCE(created_by, ''), created_at, updated_at
		FROM contact_limit WHERE is_active`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]domain.ContactLimit, 0)
	for rows.Next() {
		var c domain.ContactLimit
		if err := rows.Scan(&c.ID, &c.Channel, &c.MaxContacts, &c.PeriodDays, &c.IsActive, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *ContactLimitSQL) GetByID(ctx context.Context, id string) (*domain.ContactLimit, error) {
	var c domain.ContactLimit
	err := r.db.QueryRowContext(ctx, `
		SELECT id, COALESCE(channel, ''), max_contacts, period_days, is_active, COALESCE(created_by, ''), created_at, updated_at
		FROM contact_limit WHERE id = $1`, id).
		Scan(&c.ID, &c.Channel, &c.MaxContacts, &c.PeriodDays, &c.IsActive, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &c, err
}

func (r *ContactLimitSQL) Insert(ctx context.Context, c *domain.ContactLimit) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO contact_limit (id, channel, max_contacts, period_days, is_active, created_by, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		c.ID, nullIfEmpty(c.Channel), c.MaxContacts, c.PeriodDays, c.IsActive, nullIfEmpty(c.CreatedBy), c.CreatedAt, c.UpdatedAt,
	)
	return err
}

func (r *ContactLimitSQL) Update(ctx context.Context, c *domain.ContactLimit) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE contact_limit SET channel = $2, max_contacts = $3, period_days = $4, is_active = $5, updated_at = $6
		WHERE id = $1`,
		c.ID, nullIfEmpty(c.Channel), c.MaxContacts, c.PeriodDays, c.IsActive, c.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *ContactLimitSQL) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM contact_limit WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

type ContactLogSQL struct {
	db *sql.DB
}

func NewContactLogSQL(db *sql.DB) *ContactLogSQL {
	return &ContactLogSQL{db: db}
}

func (r *ContactLogSQL) CountSince(ctx context.Context, guestIDs []string, since time.Time) (map[string]int, error) {
	return r.count(ctx, guestIDs, since, "", false)
}

func (r *ContactLogSQL) CountSinceChannel(ctx context.Context, guestIDs []string, since time.Time, channel string) (map[string]int, error) {
	return r.count(ctx, guestIDs, since, channel, channel != "")
}

func (r *ContactLogSQL) count(ctx context.Context, guestIDs []string, since time.Time, channel string, byChannel bool) (map[string]int, error) {
	out := map[string]int{}
	if len(guestIDs) == 0 {
		return out, nil
	}
	ph := make([]string, len(guestIDs))
	args := make([]any, 0, len(guestIDs)+3)
	for i, id := range guestIDs {
		ph[i] = "$" + strconv.Itoa(i+1)
		args = append(args, id)
	}
	sinceIdx := len(guestIDs) + 1
	args = append(args, since)
	query := `
		SELECT cl.guest_id, COUNT(*)
		FROM contact_log cl
		JOIN export_job j ON j.id = cl.export_id
		WHERE cl.guest_id IN (` + strings.Join(ph, ",") + `)
		  AND cl.contacted_at >= $` + strconv.Itoa(sinceIdx)
	if byChannel {
		args = append(args, channel)
		query += ` AND j.channel = $` + strconv.Itoa(sinceIdx+1)
	}
	query += ` GROUP BY cl.guest_id`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var (
			id string
			n  int
		)
		if err := rows.Scan(&id, &n); err != nil {
			return nil, err
		}
		out[id] = n
	}
	return out, rows.Err()
}

func (r *ContactLogSQL) InsertBatch(ctx context.Context, logs []domain.ContactLog) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	for _, l := range logs {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO contact_log (id, guest_id, export_id, contacted_at)
			VALUES ($1,$2,$3,$4)`, l.ID, l.GuestID, l.ExportID, l.ContactedAt); err != nil {
			return err
		}
	}
	return tx.Commit()
}
