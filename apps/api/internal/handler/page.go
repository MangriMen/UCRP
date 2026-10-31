package handler

import (
	"net/http"
	"strconv"

	"github.com/MangriMen/UCRP/internal/platform/httpx"
	"github.com/MangriMen/UCRP/internal/service"
)

func pageParams(r *http.Request) (int, int) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit, _ = strconv.Atoi(r.URL.Query().Get("page_size"))
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}
	if page, _ := strconv.Atoi(r.URL.Query().Get("page")); page > 1 && offset == 0 {
		if limit <= 0 {
			limit = 20
		}
		offset = (page - 1) * limit
	}
	return limit, offset
}

func pageMeta(limit, offset, total int) service.PageMeta {
	if limit <= 0 {
		limit = 20
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}
	totalPages := (total + limit - 1) / limit
	if totalPages < 1 {
		totalPages = 1
	}
	return service.PageMeta{
		Page:       offset/limit + 1,
		PageSize:   limit,
		Total:      total,
		TotalPages: totalPages,
	}
}

func writePage(w http.ResponseWriter, items any, limit, offset, total int) {
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"items": items,
		"meta":  pageMeta(limit, offset, total),
	})
}
