package recommend

import (
	"sort"
	"time"
)

type SegmentMemberSnapshot struct {
	GuestID  string
	Visited  map[string]bool
	Recs     []MemberRecommendation
	HasFresh bool
}

type MemberRecommendation struct {
	TenantID   string
	Score      float64
	ComputedAt time.Time
}

type SegmentTenantAggregate struct {
	TenantID        string
	TenantName      string
	CategoryID      string
	GuestCount      int
	EligibleCount   int
	SegmentCount    int
	AverageScore    float64
	IncompleteCount int
}

func AggregateSegmentMembers(members []SegmentMemberSnapshot, limit int) []SegmentTenantAggregate {
	if limit <= 0 {
		limit = 10
	}
	tenantIDs := map[string]bool{}
	for _, m := range members {
		for tid := range m.Visited {
			tenantIDs[tid] = true
		}
		for _, rec := range m.Recs {
			tenantIDs[rec.TenantID] = true
		}
	}
	ids := make([]string, 0, len(tenantIDs))
	for tid := range tenantIDs {
		ids = append(ids, tid)
	}
	sort.Strings(ids)

	byTenant := map[string]*SegmentTenantAggregate{}
	for _, tid := range ids {
		byTenant[tid] = &SegmentTenantAggregate{
			TenantID:     tid,
			SegmentCount: len(members),
		}
	}

	for _, m := range members {
		for _, tid := range ids {
			a := byTenant[tid]
			visited := m.Visited[tid]
			if !visited {
				a.EligibleCount++
				if !m.HasFresh {
					a.IncompleteCount++
				}
			}
			for _, rec := range m.Recs {
				if rec.TenantID != tid || visited {
					continue
				}
				a.GuestCount++
				a.AverageScore += rec.Score
			}
		}
	}

	out := make([]SegmentTenantAggregate, 0, len(byTenant))
	for _, a := range byTenant {
		if a.GuestCount == 0 && a.IncompleteCount == 0 {
			continue
		}
		if a.GuestCount > 0 {
			a.AverageScore /= float64(a.GuestCount)
		}
		out = append(out, *a)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].GuestCount != out[j].GuestCount {
			return out[i].GuestCount > out[j].GuestCount
		}
		if out[i].AverageScore != out[j].AverageScore {
			return out[i].AverageScore > out[j].AverageScore
		}
		return out[i].TenantID < out[j].TenantID
	})
	if len(out) > limit {
		out = out[:limit]
	}
	return out
}
