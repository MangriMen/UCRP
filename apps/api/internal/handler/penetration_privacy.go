package handler

import "github.com/MangriMen/UCRP/internal/service"

type penetrationBucketView struct {
	One        *int `json:"one"`
	Two        *int `json:"two"`
	ThreePlus  *int `json:"threePlus"`
	Suppressed bool `json:"suppressed"`
}

type penetrationPointView struct {
	penetrationBucketView
	Period          string   `json:"period"`
	PenetrationRate *float64 `json:"penetrationRate"`
}

func visiblePenetration(bucket service.PenetrationBucket, rate float64, hide bool) (penetrationBucketView, *float64) {
	if hide || bucket.One < 5 || bucket.Two < 5 || bucket.ThreePlus < 5 {
		return penetrationBucketView{Suppressed: true}, nil
	}
	return penetrationBucketView{One: &bucket.One, Two: &bucket.Two, ThreePlus: &bucket.ThreePlus}, &rate
}

func penetrationResponse(report *service.PenetrationResponse) map[string]any {
	points := make([]penetrationPointView, 0, len(report.Series))
	hideTotal := false
	for _, point := range report.Series {
		bucket, rate := visiblePenetration(point.PenetrationBucket, point.PenetrationRate, false)
		hideTotal = hideTotal || bucket.Suppressed
		points = append(points, penetrationPointView{penetrationBucketView: bucket, Period: point.Period, PenetrationRate: rate})
	}
	distribution, rate := visiblePenetration(report.Distribution, report.PenetrationRate, hideTotal)
	return map[string]any{"from": report.From, "to": report.To, "granularity": report.Granularity,
		"distribution": distribution, "penetrationRate": rate, "series": points, "suppressed": distribution.Suppressed}
}
