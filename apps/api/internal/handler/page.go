package handler

import (
	"net/http"
	"strconv"

	"github.com/MangriMen/UCRP/internal/platform/httpx"
	"github.com/MangriMen/UCRP/internal/service"
)

const (
	defaultPageSize = 50
	maxPageSize     = 200
)

func pageParams(r *http.Request) (int, int) {
	query := r.URL.Query()
	limit, _ := strconv.Atoi(query.Get("limit"))
	if limit <= 0 {
		limit, _ = strconv.Atoi(query.Get("page_size"))
	}
	if limit <= 0 {
		limit = defaultPageSize
	}
	limit = min(limit, maxPageSize)
	offset, _ := strconv.Atoi(query.Get("offset"))
	offset = max(offset, 0)
	if page, _ := strconv.Atoi(query.Get("page")); page > 1 && offset == 0 {
		offset = (page - 1) * limit
	}
	return limit, offset
}

func pageMeta(limit, offset, total int) service.PageMeta {
	if limit <= 0 {
		limit = defaultPageSize
	}
	offset = max(offset, 0)
	totalPages := max((total+limit-1)/limit, 1)
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
