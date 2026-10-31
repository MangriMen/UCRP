package service

import "context"

type TenantVisitAggregate struct {
	VisitsCount *int `json:"visitsCount"`
	Suppressed  bool `json:"suppressed"`
}

func (s *AnalyticsService) TenantVisitAggregate(ctx context.Context, tenantID string, periodDays, threshold int) (*TenantVisitAggregate, error) {
	var visits, guests int
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*), COUNT(DISTINCT e.guest_id)
		FROM event e
		JOIN event_type et ON et.id = e.event_type_id AND et.counts_as_presence
		WHERE e.tenant_id = $1 AND e.guest_id IS NOT NULL
		  AND ($2 = 0 OR e.occurred_at >= now() - $2 * interval '1 day')`, tenantID, periodDays).Scan(&visits, &guests)
	if err != nil {
		return nil, err
	}
	if threshold < 5 {
		threshold = 5
	}
	if guests < threshold {
		return &TenantVisitAggregate{Suppressed: true}, nil
	}
	return &TenantVisitAggregate{VisitsCount: &visits}, nil
}
