package handler

import (
	"encoding/json"
	"testing"

	"github.com/MangriMen/UCRP/internal/service"
)

func TestCrossTenantSuppression(t *testing.T) {
	report := &service.CrossTenantResponse{
		Distribution:    service.CrossTenantDistribution{One: 20, Two: 10, ThreePlus: 10},
		CrossTenantRate: 0.5,
		Series:          []service.CrossTenantPoint{{Period: "2026-01", CrossTenantDistribution: service.CrossTenantDistribution{One: 10, Two: 4, ThreePlus: 6}, CrossTenantRate: 0.5}},
	}
	body, err := json.Marshal(crossTenantResponse(report))
	if err != nil {
		t.Fatal(err)
	}
	var result struct {
		Distribution struct {
			One        *int `json:"one"`
			Suppressed bool `json:"suppressed"`
		} `json:"distribution"`
		Rate   *float64 `json:"crossTenantRate"`
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
	resultBody, err := json.Marshal(crossTenantResponse(report))
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
