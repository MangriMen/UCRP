package segments

import (
	"context"
	"database/sql"
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

	var query string
	args := make([]any, 0, len(rule.Was)+len(rule.WasNot)+2)
	for _, id := range rule.Was {
		args = append(args, id)
	}
	if len(rule.WasNot) == 0 {
		args = append(args, since)
		query = `
			SELECT DISTINCT e.guest_id
			FROM event e
			JOIN event_type et ON et.id = e.event_type_id AND et.counts_as_presence
			WHERE e.guest_id IS NOT NULL
				AND e.tenant_id IN (` + inPlaceholders(1, len(rule.Was)) + `)
				AND e.occurred_at >= $` + strconv.Itoa(len(rule.Was)+1) + `
			ORDER BY e.guest_id`
	} else {
		sinceIdx := 1 + len(rule.Was)
		notStart := sinceIdx + 1
		notSinceIdx := notStart + len(rule.WasNot)
		args = append(args, since)
		for _, id := range rule.WasNot {
			args = append(args, id)
		}
		args = append(args, since)
		query = `
			WITH was AS (
				SELECT DISTINCT e.guest_id
				FROM event e
				JOIN event_type et ON et.id = e.event_type_id AND et.counts_as_presence
				WHERE e.guest_id IS NOT NULL
					AND e.tenant_id IN (` + inPlaceholders(1, len(rule.Was)) + `)
					AND e.occurred_at >= $` + strconv.Itoa(sinceIdx) + `
			),
			was_not AS (
				SELECT DISTINCT e.guest_id
				FROM event e
				JOIN event_type et ON et.id = e.event_type_id AND et.counts_as_presence
				WHERE e.guest_id IS NOT NULL
					AND e.tenant_id IN (` + inPlaceholders(notStart, len(rule.WasNot)) + `)
					AND e.occurred_at >= $` + strconv.Itoa(notSinceIdx) + `
			)
			SELECT w.guest_id
			FROM was w
			LEFT JOIN was_not n ON n.guest_id = w.guest_id
			WHERE n.guest_id IS NULL
			ORDER BY w.guest_id`
	}

	rows, err := e.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
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
	return nil
}

func inPlaceholders(start, n int) string {
	parts := make([]string, n)
	for i := range parts {
		parts[i] = "$" + strconv.Itoa(start+i)
	}
	return strings.Join(parts, ",")
}
