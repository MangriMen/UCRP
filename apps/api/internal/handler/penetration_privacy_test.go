package handler

import (
	"encoding/json"
	"testing"

	"github.com/MangriMen/UCRP/internal/service"
)

func TestPenetrationSuppression(t *testing.T) {
	report := &service.PenetrationResponse{
		Distribution:    service.PenetrationBucket{One: 20, Two: 10, ThreePlus: 10},
		PenetrationRate: 0.5,
		Series:          []service.PenetrationPoint{{Period: "2026-01", PenetrationBucket: service.PenetrationBucket{One: 10, Two: 4, ThreePlus: 6}, PenetrationRate: 0.5}},
	}
	body, err := json.Marshal(penetrationResponse(report))
	if err != nil {
		t.Fatal(err)
	}
	var result struct {
		Distribution struct {
			One        *int `json:"one"`
			Suppressed bool `json:"suppressed"`
		} `json:"distribution"`
		Rate   *float64 `json:"penetrationRate"`
		Series []struct {
			Two        *int `json:"two"`
			Suppressed bool `json:"suppressed"`
		} `json:"series"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatal(err)
	}
	if !result.Distribution.Suppressed || result.Distribution.One != nil || result.Rate != nil || len(result.Series) != 1 || !result.Series[0].Suppressed || result.Series[0].Two != nil {
		t.Fatalf("small period or total visible: %s", body)
	}
	report.Series[0].Two = 5
	resultBody, err := json.Marshal(penetrationResponse(report))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(resultBody, &result); err != nil {
		t.Fatal(err)
	}
	if result.Distribution.Suppressed || result.Distribution.One == nil || result.Rate == nil || result.Series[0].Suppressed || result.Series[0].Two == nil {
		t.Fatalf("eligible distribution hidden: %s", resultBody)
	}
}
