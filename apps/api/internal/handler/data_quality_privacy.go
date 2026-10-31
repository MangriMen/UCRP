package handler

import "github.com/MangriMen/UCRP/internal/service"

type dataQualityItemView struct {
	DataSourceID  string   `json:"dataSourceId"`
	Name          string   `json:"name"`
	LastImportAt  string   `json:"lastImportAt,omitempty"`
	RejectedRate  *float64 `json:"rejectedRate"`
	PendingMerges *int     `json:"pendingMerges"`
	Warning       *bool    `json:"warning"`
	Suppressed    bool     `json:"suppressed"`
}

func dataQualityResponse(items []service.DataQualityItem) []dataQualityItemView {
	result := make([]dataQualityItemView, 0, len(items))
	for _, item := range items {
		view := dataQualityItemView{
			DataSourceID: item.DataSourceID,
			Name:         item.Name,
			LastImportAt: item.LastImportAt,
			Suppressed:   true,
		}
		if item.PendingMerges >= 5 {
			n := item.PendingMerges
			view.PendingMerges = &n
		}
		result = append(result, view)
	}
	return result
}
