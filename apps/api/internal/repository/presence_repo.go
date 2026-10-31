package repository

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"time"
)

type PresenceSQL struct {
	db *sql.DB
}

func NewPresenceSQL(db *sql.DB) *PresenceSQL {
	return &PresenceSQL{db: db}
}

type TenantPresenceStat struct {
	TenantID    string
	GuestCount  int
	LastVisitAt time.Time
}

func (r *PresenceSQL) GuestHistory(ctx context.Context, guestID string, since time.Time) ([]TenantPresenceStat, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT e.tenant_id, COUNT(DISTINCT e.guest_id), MAX(e.occurred_at)
		FROM event e
		JOIN event_type et ON et.id = e.event_type_id AND et.counts_as_presence
		WHERE e.guest_id = $1 AND e.occurred_at >= $2
		GROUP BY e.tenant_id`, guestID, since)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]TenantPresenceStat, 0)
	for rows.Next() {
		var s TenantPresenceStat
		if err := rows.Scan(&s.TenantID, &s.GuestCount, &s.LastVisitAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *PresenceSQL) VisitedTenants(ctx context.Context, guestID string) (map[string]bool, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT DISTINCT e.tenant_id
		FROM event e
		JOIN event_type et ON et.id = e.event_type_id AND et.counts_as_presence
		WHERE e.guest_id = $1`, guestID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := map[string]bool{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out[id] = true
	}
	return out, rows.Err()
}

type TenantCandidate struct {
	ID         string
	Name       string
	CategoryID string
	Priority   int
	Floor      *int
}

func (r *PresenceSQL) ActiveTenants(ctx context.Context) ([]TenantCandidate, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT t.id, t.name, t.category_id, t.priority, t.floor
		FROM tenant t
		JOIN mall m ON m.id = t.mall_id AND m.is_active
		WHERE t.is_active ORDER BY t.id`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]TenantCandidate, 0)
	for rows.Next() {
		var t TenantCandidate
		if err := rows.Scan(&t.ID, &t.Name, &t.CategoryID, &t.Priority, &t.Floor); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *PresenceSQL) PairGuestCounts(ctx context.Context, since time.Time, tenantIDs []string) (map[string]int, map[[2]string]int, error) {
	if len(tenantIDs) == 0 {
		return map[string]int{}, map[[2]string]int{}, nil
	}
	placeholders := make([]string, len(tenantIDs))
	args := make([]any, 0, len(tenantIDs)+1)
	args = append(args, since)
	for i, id := range tenantIDs {
		placeholders[i] = "$" + strconv.Itoa(i+2)
		args = append(args, id)
	}
	inList := strings.Join(placeholders, ",")
	pairsSQL := `
		SELECT tenant_id, guest_id FROM (
			SELECT DISTINCT e.tenant_id, e.guest_id
			FROM event e
			JOIN event_type et ON et.id = e.event_type_id AND et.counts_as_presence
			WHERE e.guest_id IS NOT NULL AND e.occurred_at >= $1
			  AND e.tenant_id IN (` + inList + `)
		) p`
	rows, err := r.db.QueryContext(ctx, pairsSQL, args...)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = rows.Close() }()
	single := map[string]int{}
	for rows.Next() {
		var tenantID, guestID string
		if err := rows.Scan(&tenantID, &guestID); err != nil {
			return nil, nil, err
		}
		single[tenantID]++
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	coSQL := `
		SELECT a.tenant_id, b.tenant_id, COUNT(DISTINCT a.guest_id)
		FROM (
			SELECT DISTINCT e.tenant_id, e.guest_id
			FROM event e
			JOIN event_type et ON et.id = e.event_type_id AND et.counts_as_presence
			WHERE e.guest_id IS NOT NULL AND e.occurred_at >= $1
			  AND e.tenant_id IN (` + inList + `)
		) a
		JOIN (
			SELECT DISTINCT e.tenant_id, e.guest_id
			FROM event e
			JOIN event_type et ON et.id = e.event_type_id AND et.counts_as_presence
			WHERE e.guest_id IS NOT NULL AND e.occurred_at >= $1
			  AND e.tenant_id IN (` + inList + `)
		) b ON a.guest_id = b.guest_id AND a.tenant_id < b.tenant_id
		GROUP BY a.tenant_id, b.tenant_id`
	co := map[[2]string]int{}
	pairRows, err := r.db.QueryContext(ctx, coSQL, args...)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = pairRows.Close() }()
	for pairRows.Next() {
		var a, b string
		var n int
		if err := pairRows.Scan(&a, &b, &n); err != nil {
			return nil, nil, err
		}
		co[[2]string{a, b}] = n
	}
	return single, co, pairRows.Err()
}

func (r *PresenceSQL) UniquePresenceGuests(ctx context.Context, since time.Time) (int, error) {
	var n int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM (
			SELECT DISTINCT e.guest_id
			FROM event e
			JOIN event_type et ON et.id = e.event_type_id AND et.counts_as_presence
			WHERE e.guest_id IS NOT NULL AND e.occurred_at >= $1
		) g`, since).Scan(&n)
	return n, err
}

type CategoryRelationRow struct {
	FromCategoryID string
	ToCategoryID   string
	RelationType   string
	WeightPct      int
}

func (r *PresenceSQL) CategoryRelations(ctx context.Context, categoryIDs []string) ([]CategoryRelationRow, error) {
	if len(categoryIDs) == 0 {
		return nil, nil
	}
	placeholders := make([]string, len(categoryIDs))
	args := make([]any, len(categoryIDs))
	for i, id := range categoryIDs {
		placeholders[i] = "$" + strconv.Itoa(i+1)
		args[i] = id
	}
	joined := strings.Join(placeholders, ",")
	rows, err := r.db.QueryContext(ctx, `
		SELECT category_id, related_category_id, relation_type, weight_pct
		FROM tenant_category_relation
		WHERE category_id IN (`+joined+`)
		   OR related_category_id IN (`+joined+`)`, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]CategoryRelationRow, 0)
	for rows.Next() {
		var rel CategoryRelationRow
		if err := rows.Scan(&rel.FromCategoryID, &rel.ToCategoryID, &rel.RelationType, &rel.WeightPct); err != nil {
			return nil, err
		}
		out = append(out, rel)
	}
	return out, rows.Err()
}

type SegmentGuestPresence struct {
	GuestID   string
	TenantIDs map[string]bool
}

func (r *PresenceSQL) SegmentMembers(ctx context.Context, segmentID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT m.guest_id
		FROM segment_member m
		JOIN segment_version v ON v.id = m.segment_version_id
		JOIN segment s ON s.id = v.segment_id AND s.current_version_no = v.version_no
		WHERE s.id = $1
		ORDER BY m.guest_id`, segmentID)
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

func (r *PresenceSQL) SegmentMembersWithPresence(ctx context.Context, segmentID string) ([]SegmentGuestPresence, error) {
	members, err := r.SegmentMembers(ctx, segmentID)
	if err != nil {
		return nil, err
	}
	out := make([]SegmentGuestPresence, 0, len(members))
	for _, guestID := range members {
		visited, err := r.VisitedTenants(ctx, guestID)
		if err != nil {
			return nil, err
		}
		out = append(out, SegmentGuestPresence{GuestID: guestID, TenantIDs: visited})
	}
	return out, nil
}

type StoredRecommendation struct {
	TenantID   string
	Score      float64
	ComputedAt time.Time
}

func (r *PresenceSQL) RecommendationsByGuests(ctx context.Context, guestIDs []string) (map[string][]StoredRecommendation, error) {
	out := map[string][]StoredRecommendation{}
	if len(guestIDs) == 0 {
		return out, nil
	}
	placeholders := make([]string, len(guestIDs))
	args := make([]any, len(guestIDs))
	for i, id := range guestIDs {
		placeholders[i] = "$" + strconv.Itoa(i+1)
		args[i] = id
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT guest_id, tenant_id, score, computed_at
		FROM recommendation
		WHERE guest_id IN (`+strings.Join(placeholders, ",")+`)
		ORDER BY guest_id, rank`, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var guestID, tenantID string
		var score float64
		var computedAt time.Time
		if err := rows.Scan(&guestID, &tenantID, &score, &computedAt); err != nil {
			return nil, err
		}
		out[guestID] = append(out[guestID], StoredRecommendation{
			TenantID:   tenantID,
			Score:      score,
			ComputedAt: computedAt,
		})
	}
	return out, rows.Err()
}

func (r *PresenceSQL) GuestIDsByBatch(ctx context.Context, batchID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT DISTINCT guest_id
		FROM event
		WHERE import_batch_id = $1 AND guest_id IS NOT NULL
		ORDER BY guest_id`, batchID)
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
