package handler

import (
	"testing"

	"github.com/MangriMen/UCRP/internal/service"
)

func TestDashboardMetricSuppression(t *testing.T) {
	metrics := &service.MetricsResponse{MatchRate: 0.8, DisputedMergeRate: 0.2, SegmentToExportMedianSeconds: 20, SegmentToExportP95Seconds: 40}
	result := dashboardMetricsResponse(metrics)
	for _, key := range []string{"matchRate", "disputedMergeRate", "segmentToExportMedianSeconds", "segmentToExportP95Seconds"} {
		if result[key] != nil {
			t.Fatalf("%s must be suppressed: %v", key, result)
		}
	}
	if _, ok := result["campaignConversionRate"]; ok {
		t.Fatal("unverified campaign metric was disclosed")
	}
	metrics.DisputedRateReliable, metrics.ExportLatencyReliable = true, true
	result = dashboardMetricsResponse(metrics)
	if result["disputedMergeRate"] != 0.2 || result["segmentToExportMedianSeconds"] != float64(20) || result["segmentToExportP95Seconds"] != float64(40) || result["matchRate"] != nil {
		t.Fatalf("unexpected metric visibility: %v", result)
	}
}
