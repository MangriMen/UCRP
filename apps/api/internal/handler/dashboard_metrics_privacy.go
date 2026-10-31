package handler

import "github.com/MangriMen/UCRP/internal/service"

func dashboardMetricsResponse(metrics *service.MetricsResponse) map[string]any {
	result := map[string]any{
		"from": metrics.From, "to": metrics.To,
		"matchRate": nil, "disputedMergeRate": nil,
		"segmentToExportMedianSeconds": nil, "segmentToExportP95Seconds": nil,
		"suppressed": true,
	}
	if metrics.DisputedSafe {
		result["disputedMergeRate"] = metrics.ContestedMergeRate
	}
	if metrics.ExportLatencySafe {
		result["segmentToExportMedianSeconds"] = metrics.AvgSegmentToExport
		result["segmentToExportP95Seconds"] = metrics.SegmentToExportP95
	}
	return result
}
