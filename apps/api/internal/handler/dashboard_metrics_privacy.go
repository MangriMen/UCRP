package handler

import "github.com/MangriMen/UCRP/internal/service"

func dashboardMetricsResponse(metrics *service.MetricsResponse) map[string]any {
	result := map[string]any{
		"from": metrics.From, "to": metrics.To,
		"matchRate": nil, "disputedMergeRate": nil,
		"segmentToExportMedianSeconds": nil, "segmentToExportP95Seconds": nil,
		"suppressed": true,
	}
	if metrics.DisputedRateReliable {
		result["disputedMergeRate"] = metrics.DisputedMergeRate
	}
	if metrics.ExportLatencyReliable {
		result["segmentToExportMedianSeconds"] = metrics.SegmentToExportMedianSeconds
		result["segmentToExportP95Seconds"] = metrics.SegmentToExportP95Seconds
	}
	return result
}
