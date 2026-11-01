package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPageParams(t *testing.T) {
	cases := []struct {
		name       string
		query      string
		wantLimit  int
		wantOffset int
	}{
		{"defaults", "", 50, 0},
		{"limit", "limit=10", 10, 0},
		{"page_size", "page_size=25", 25, 0},
		{"limit wins over page_size", "limit=10&page_size=25", 10, 0},
		{"zero falls back to default", "limit=0", 50, 0},
		{"negative falls back to default", "page_size=-5", 50, 0},
		{"clamped to max", "limit=1000", 200, 0},
		{"page to offset", "page=3&page_size=20", 20, 40},
		{"page uses default size", "page=2", 50, 50},
		{"page uses clamped size", "page=2&limit=500", 200, 200},
		{"explicit offset wins over page", "page=3&limit=20&offset=5", 20, 5},
		{"negative offset", "offset=-3", 50, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			limit, offset := pageParams(httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/x?"+tc.query, nil))
			if limit != tc.wantLimit || offset != tc.wantOffset {
				t.Fatalf("got limit=%d offset=%d, want limit=%d offset=%d", limit, offset, tc.wantLimit, tc.wantOffset)
			}
		})
	}
}

func TestPageMetaMatchesParams(t *testing.T) {
	meta := pageMeta(50, 100, 120)
	if meta.Page != 3 || meta.PageSize != 50 || meta.TotalPages != 3 {
		t.Fatalf("unexpected meta %+v", meta)
	}
	empty := pageMeta(0, 0, 0)
	if empty.PageSize != defaultPageSize || empty.TotalPages != 1 {
		t.Fatalf("unexpected empty meta %+v", empty)
	}
}
