package segments

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/MangriMen/UCRP/internal/domain"
)

const (
	purchaseEventCode = "purchase"
	operatorAnd       = "and"
	operatorOr        = "or"
	maxRuleDepth      = 4
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
	if err := validateRule(rule, 0, "rule"); err != nil {
		return nil, err
	}
	return e.evaluateNode(ctx, rule)
}

func (e *SqlEvaluator) evaluateNode(ctx context.Context, rule domain.SegmentRule) ([]string, error) {
	if !rule.IsGroup() {
		return e.evaluateLeaf(ctx, rule)
	}
	var result []string
	for i, condition := range rule.Conditions {
		guestIDs, err := e.evaluateNode(ctx, condition)
		if err != nil {
			return nil, err
		}
		if i == 0 {
			result = guestIDs
			continue
		}
		result = combineGuestIDs(rule.Operator, result, guestIDs)
		if rule.Operator == operatorAnd && len(result) == 0 {
			break
		}
	}
	sort.Strings(result)
	return result, nil
}

func combineGuestIDs(operator string, left, right []string) []string {
	inRight := make(map[string]struct{}, len(right))
	for _, id := range right {
		inRight[id] = struct{}{}
	}
	if operator == operatorOr {
		inLeft := make(map[string]struct{}, len(left))
		out := append([]string(nil), left...)
		for _, id := range left {
			inLeft[id] = struct{}{}
		}
		for _, id := range right {
			if _, seen := inLeft[id]; !seen {
				out = append(out, id)
			}
		}
		return out
	}
	out := make([]string, 0)
	for _, id := range left {
		if _, ok := inRight[id]; ok {
			out = append(out, id)
		}
	}
	return out
}

func (e *SqlEvaluator) evaluateLeaf(ctx context.Context, rule domain.SegmentRule) ([]string, error) {
	if rule.PeriodDays <= 0 {
		rule.PeriodDays = 90
	}
	since := time.Now().UTC().AddDate(0, 0, -rule.PeriodDays)

	query, args := buildQuery(rule, since)
	rows, err := e.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("evaluate segment: %w", err)
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

type queryArgs struct {
	values []any
}

func (a *queryArgs) add(value any) string {
	a.values = append(a.values, value)
	return "$" + strconv.Itoa(len(a.values))
}

func (a *queryArgs) addAll(values []string) string {
	placeholders := make([]string, len(values))
	for i, value := range values {
		placeholders[i] = a.add(value)
	}
	return strings.Join(placeholders, ",")
}

func buildQuery(rule domain.SegmentRule, since time.Time) (string, []any) {
	args := &queryArgs{}
	sinceArg := args.add(since)

	var base string
	var conditions []string
	switch {
	case len(rule.VisitedTenantIDs) > 0:
		base = `SELECT DISTINCT e.guest_id ` + presenceAtTenants(args, rule.VisitedTenantIDs, sinceArg) + ` AND e.guest_id IS NOT NULL`
		if len(rule.VisitedCategoryIDs) > 0 {
			conditions = append(conditions, `EXISTS (SELECT 1 `+presenceInCategories(args, rule.VisitedCategoryIDs, sinceArg)+` AND e.guest_id = base.guest_id)`)
		}
	case len(rule.VisitedCategoryIDs) > 0:
		base = `SELECT DISTINCT e.guest_id ` + presenceInCategories(args, rule.VisitedCategoryIDs, sinceArg) + ` AND e.guest_id IS NOT NULL`
	default:
		base = `SELECT DISTINCT ep.guest_id ` + purchasesInZones(args, rule.PurchasedZones, sinceArg) + ` AND ep.guest_id IS NOT NULL`
	}
	if len(rule.PurchasedZones) > 0 && (len(rule.VisitedTenantIDs) > 0 || len(rule.VisitedCategoryIDs) > 0) {
		conditions = append(conditions, `EXISTS (SELECT 1 `+purchasesInZones(args, rule.PurchasedZones, sinceArg)+` AND ep.guest_id = base.guest_id)`)
	}
	if len(rule.ExcludedCategoryIDs) > 0 {
		conditions = append(conditions, `NOT EXISTS (SELECT 1 `+presenceInCategories(args, rule.ExcludedCategoryIDs, sinceArg)+` AND e.guest_id = base.guest_id)`)
	}
	if rule.MinVisits != nil {
		conditions = append(conditions, `(SELECT COUNT(*) FROM event ev
				JOIN event_type evt ON evt.id = ev.event_type_id AND evt.counts_as_presence
				WHERE ev.guest_id = base.guest_id AND ev.occurred_at >= `+sinceArg+`) >= `+args.add(*rule.MinVisits))
	}
	if rule.MinAmountMinor != nil {
		conditions = append(conditions, `(SELECT COALESCE(SUM(ev.amount_minor), 0) FROM event ev
				JOIN event_type evt ON evt.id = ev.event_type_id AND evt.code = '`+purchaseEventCode+`'
				WHERE ev.guest_id = base.guest_id AND ev.occurred_at >= `+sinceArg+`) >= `+args.add(*rule.MinAmountMinor))
	}

	if len(rule.ExcludedTenantIDs) > 0 {
		conditions = append(conditions, `NOT EXISTS (
				SELECT 1 FROM event e2
				JOIN event_type et2 ON et2.id = e2.event_type_id AND et2.counts_as_presence
				WHERE e2.guest_id = base.guest_id
					AND e2.tenant_id IN (`+args.addAll(rule.ExcludedTenantIDs)+`)
					AND e2.occurred_at >= `+sinceArg+`
			)`)
	}
	if rule.RequireContact != "" {
		conditions = append(conditions, `EXISTS (
				SELECT 1 FROM guest_identifier i
				WHERE i.guest_id = base.guest_id AND i.id_type = `+args.add(rule.RequireContact)+`
			)`)
	}

	query := `SELECT guest_id FROM (` + base + `) base`
	if len(conditions) > 0 {
		query += ` WHERE ` + strings.Join(conditions, ` AND `)
	}
	return query + ` ORDER BY guest_id`, args.values
}

func presenceAtTenants(args *queryArgs, tenantIDs []string, sinceArg string) string {
	return `FROM event e
				JOIN event_type et ON et.id = e.event_type_id AND et.counts_as_presence
				WHERE e.tenant_id IN (` + args.addAll(tenantIDs) + `)
					AND e.occurred_at >= ` + sinceArg
}

func presenceInCategories(args *queryArgs, categoryIDs []string, sinceArg string) string {
	return `FROM event e
				JOIN event_type et ON et.id = e.event_type_id AND et.counts_as_presence
				JOIN tenant tc ON tc.id = e.tenant_id
				WHERE tc.category_id IN (` + args.addAll(categoryIDs) + `)
					AND e.occurred_at >= ` + sinceArg
}

func purchasesInZones(args *queryArgs, zones []domain.ZoneRef, sinceArg string) string {
	pairs := make([]string, len(zones))
	for i, zone := range zones {
		pairs[i] = "(tz.mall_id = " + args.add(zone.MallID) + " AND tz.floor = " + args.add(zone.Floor) + ")"
	}
	return `FROM event ep
				JOIN event_type ept ON ept.id = ep.event_type_id AND ept.code = '` + purchaseEventCode + `'
				JOIN tenant tz ON tz.id = ep.tenant_id
				JOIN mall mz ON mz.id = tz.mall_id AND mz.is_active
				WHERE ep.occurred_at >= ` + sinceArg + `
					AND (` + strings.Join(pairs, " OR ") + `)`
}

func validateRule(rule domain.SegmentRule, depth int, path string) error {
	if !rule.IsGroup() {
		return validateLeaf(rule, path)
	}
	if depth >= maxRuleDepth || len(rule.Conditions) == 0 {
		return domain.ErrValidation
	}
	if rule.Operator != operatorAnd && rule.Operator != operatorOr {
		return domain.ErrValidation
	}
	for i, condition := range rule.Conditions {
		if err := validateRule(condition, depth+1, fmt.Sprintf("%s.conditions[%d]", path, i)); err != nil {
			return err
		}
	}
	return nil
}

func validateLeaf(rule domain.SegmentRule, path string) error {
	if len(rule.VisitedTenantIDs) == 0 && len(rule.VisitedCategoryIDs) == 0 && len(rule.PurchasedZones) == 0 {
		return &domain.FieldError{Field: path + ".visitedTenantIds", Reason: "Выберите арендатора, категорию или этаж покупки"}
	}
	for _, excluded := range rule.ExcludedCategoryIDs {
		if slices.Contains(rule.VisitedCategoryIDs, excluded) {
			return &domain.FieldError{Field: path + ".excludedCategoryIds", Reason: "Категория указана и среди посещённых, и среди исключённых"}
		}
	}
	if rule.MinVisits != nil && *rule.MinVisits < 1 {
		return &domain.FieldError{Field: path + ".minVisits", Reason: "Не меньше 1"}
	}
	if rule.MinAmountMinor != nil && *rule.MinAmountMinor < 0 {
		return &domain.FieldError{Field: path + ".minAmountMinor", Reason: "Не меньше 0"}
	}
	for _, excluded := range rule.ExcludedTenantIDs {
		if slices.Contains(rule.VisitedTenantIDs, excluded) {
			return &domain.FieldError{Field: path + ".excludedTenantIds", Reason: "Арендатор указан и среди посещённых, и среди исключённых"}
		}
	}
	if !validZones(rule.PurchasedZones) {
		return domain.ErrValidation
	}
	switch rule.RequireContact {
	case "":
	case "phone", "email", "card":
	default:
		return domain.ErrValidation
	}
	return nil
}

func validZones(zones []domain.ZoneRef) bool {
	seen := make(map[domain.ZoneRef]struct{}, len(zones))
	for _, zone := range zones {
		if zone.MallID == "" {
			return false
		}
		if _, duplicate := seen[zone]; duplicate {
			return false
		}
		seen[zone] = struct{}{}
	}
	return true
}
