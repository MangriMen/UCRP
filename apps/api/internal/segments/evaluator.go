package segments

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/MangriMen/UCRP/internal/domain"
)

type SqlEvaluator struct {
	db *sql.DB
}

func NewSqlEvaluator(db *sql.DB) *SqlEvaluator {
	return &SqlEvaluator{db: db}
}

func (e *SqlEvaluator) Preview(ctx context.Context, rule domain.SegmentRule) (int, []string, error) {
	guestIDs, err := e.Evaluate(ctx, rule)
	if err != nil {
		return 0, nil, err
	}
	examples := guestIDs
	if len(examples) > 20 {
		examples = examples[:20]
	}
	return len(guestIDs), examples, nil
}

func (e *SqlEvaluator) Evaluate(ctx context.Context, rule domain.SegmentRule) ([]string, error) {
	if err := validateRule(rule); err != nil {
		return nil, err
	}
	if rule.PeriodDays <= 0 {
		rule.PeriodDays = 90
	}
	since := time.Now().UTC().AddDate(0, 0, -rule.PeriodDays)

	args := make([]any, 0, len(rule.Was)+len(rule.WasNot)+4)
	for _, id := range rule.Was {
		args = append(args, id)
	}
	args = append(args, since)

	query := `
		SELECT guest_id FROM (
			SELECT e.guest_id, COUNT(DISTINCT e.tenant_id) AS hit
			FROM event e
			JOIN event_type et ON et.id = e.event_type_id AND et.counts_as_presence
			WHERE e.guest_id IS NOT NULL
				AND e.tenant_id IN (` + inPlaceholders(1, len(rule.Was)) + `)
				AND e.occurred_at >= $` + strconv.Itoa(len(rule.Was)+1) + `
			GROUP BY e.guest_id
			HAVING COUNT(DISTINCT e.tenant_id) = $` + strconv.Itoa(len(rule.Was)+2)
	args = append(args, len(rule.Was))

	if len(rule.WasNot) > 0 {
		notStart := len(args) + 1
		for _, id := range rule.WasNot {
			args = append(args, id)
		}
		args = append(args, since)
		query += `
		) base
		WHERE NOT EXISTS (
			SELECT 1 FROM event e2
			JOIN event_type et2 ON et2.id = e2.event_type_id AND et2.counts_as_presence
			WHERE e2.guest_id = base.guest_id
				AND e2.tenant_id IN (` + inPlaceholders(notStart, len(rule.WasNot)) + `)
				AND e2.occurred_at >= $` + strconv.Itoa(notStart+len(rule.WasNot)) + `
		)`
	} else {
		query += `
		) base`
	}

	if rule.RequireContact != "" {
		args = append(args, rule.RequireContact)
		query += `
		AND EXISTS (
			SELECT 1 FROM guest_identifier i
			WHERE i.guest_id = base.guest_id AND i.id_type = $` + strconv.Itoa(len(args)) + `
		)`
	}

	query += ` ORDER BY guest_id`

	rows, err := e.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("evaluate segment: %w", err)
	}
	defer rows.Close()
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

func validateRule(rule domain.SegmentRule) error {
	if len(rule.Was) == 0 {
		return domain.ErrValidation
	}
	for _, t := range rule.WasNot {
		for _, w := range rule.Was {
			if t == w {
				return domain.ErrValidation
			}
		}
	}
	switch rule.RequireContact {
	case "":
	case "phone", "email", "card":
	default:
		return domain.ErrValidation
	}
	return nil
}

func inPlaceholders(start, n int) string {
	parts := make([]string, n)
	for i := range parts {
		parts[i] = "$" + strconv.Itoa(start+i)
	}
	return strings.Join(parts, ",")
}
