package handler

import "github.com/MangriMen/UCRP/internal/service"

type crossTenantDistributionView struct {
	One        *int `json:"one"`
	Two        *int `json:"two"`
	ThreePlus  *int `json:"threePlus"`
	Suppressed bool `json:"suppressed"`
}

type crossTenantPointView struct {
	crossTenantDistributionView
	Period          string   `json:"period"`
	CrossTenantRate *float64 `json:"crossTenantRate"`
}

func visibleCrossTenant(bucket service.CrossTenantDistribution, rate float64, hide bool) (crossTenantDistributionView, *float64) {
	if hide || bucket.One < 5 || bucket.Two < 5 || bucket.ThreePlus < 5 {
		return crossTenantDistributionView{Suppressed: true}, nil
	}
	return crossTenantDistributionView{One: &bucket.One, Two: &bucket.Two, ThreePlus: &bucket.ThreePlus}, &rate
}

func crossTenantResponse(report *service.CrossTenantResponse) map[string]any {
	points := make([]crossTenantPointView, 0, len(report.Series))
	hideTotal := false
	for _, point := range report.Series {
		bucket, rate := visibleCrossTenant(point.CrossTenantDistribution, point.CrossTenantRate, false)
		hideTotal = hideTotal || bucket.Suppressed
		points = append(points, crossTenantPointView{crossTenantDistributionView: bucket, Period: point.Period, CrossTenantRate: rate})
	}
	distribution, rate := visibleCrossTenant(report.Distribution, report.CrossTenantRate, hideTotal)
	return map[string]any{"from": report.From, "to": report.To, "granularity": report.Granularity,
		"distribution": distribution, "crossTenantRate": rate, "series": points, "suppressed": distribution.Suppressed}
}
