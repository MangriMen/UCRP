package seed

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

type batchSpec struct {
	Key        string
	Source     string
	Mode       string
	FileName   string
	Status     string
	Total      int
	Accepted   int
	Rejected   int
	Duplicate  int
	Error      string
	DaysAgo    int
	WindowFrom int
	WindowTo   int
}

type rejectedSpec struct {
	Batch     string
	RowNumber int
	Code      string
	Message   string
	RawLine   string
}

func demoBatches() []batchSpec {
	return []batchSpec{
		{"bean-0901", "pos-bean", "file", "pos_bean_2026-09-01.csv", "completed", 20, 20, 0, 0, "", 60, 61, 60},
		{"bean-0915", "pos-bean", "file", "pos_bean_2026-09-15.csv", "completed", 18, 18, 0, 0, "", 46, 47, 46},
		{"fit-0901", "crm-fit", "file", "crm_fit_2026-09-01.csv", "completed", 15, 15, 0, 0, "", 60, 61, 60},
		{"cinema-1001", "pos-cinema", "file", "pos_cinema_2026-10-01.csv", "completed", 25, 25, 0, 0, "", 30, 31, 30},
		{"cinema-1015", "pos-cinema", "file", "pos_cinema_2026-10-15.csv", "completed", 22, 21, 0, 1, "", 16, 17, 16},
		{"trend-1005", "pos-trend", "file", "pos_trend_2026-10-05.csv", "completed", 12, 12, 0, 0, "", 26, 27, 26},
		{"food-0920", "csv-food", "file", "csv_food_bad_rows.csv", "completed_with_errors", 10, 6, 4, 0, "", 41, 42, 41},
		{"cinema-1020", "pos-cinema", "file", "pos_cinema_2026-10-20.csv", "completed_with_errors", 14, 12, 2, 0, "", 11, 12, 11},
		{"wifi-1010", "wifi-mall", "file", "mall_wifi_2026-10.csv", "completed_with_errors", 30, 27, 3, 0, "", 21, 22, 21},
		{"kids-pull-1025", "api-kids", "pull", "", "failed", 0, 0, 0, 0, "Таймаут ответа API детского центра", 6, 7, 6},
		{"fit-pull-1028", "crm-fit", "pull", "", "queued", 0, 0, 0, 0, "", 3, 4, 3},
		{"bean-1030", "pos-bean", "file", "pos_bean_2026-10-30.csv", "running", 0, 0, 0, 0, "", 1, 2, 1},
	}
}

func demoRejectedRows() []rejectedSpec {
	return []rejectedSpec{
		{"food-0920", 2, "no_external_id", "Не указан внешний идентификатор", ",2026-09-20T12:10:00Z,purchase,45000,79161234567,,"},
		{"food-0920", 5, "bad_date", "Не распознана дата события", "CHK-9005,20.09.2026 12:40,purchase,32000,79161234568,,"},
		{"food-0920", 7, "bad_tenant", "Неизвестный арендатор: food-unknown", "CHK-9007,2026-09-20T13:05:00Z,purchase,28000,79161234569,,food-unknown"},
		{"food-0920", 9, "bad_amount", "Не распознана сумма", "CHK-9009,2026-09-20T13:45:00Z,purchase,двадцать,79161234570,,"},
		{"cinema-1020", 4, "bad_event_type", "Неизвестный тип события: refund", "CHK-9104,2026-10-20T18:20:00Z,refund,60000,79161234571,,"},
		{"cinema-1020", 11, "bad_email", "Не распознан email", "CHK-9111,2026-10-20T20:05:00Z,booking,90000,,broken@@mail,"},
		{"wifi-1010", 3, "no_external_id", "Не указан внешний идентификатор", ",2026-10-10T10:15:00Z,visit,,79161234572,,"},
		{"wifi-1010", 8, "bad_date", "Не распознана дата события", "WIFI-1008,10/10/2026 11:20,visit,,79161234573,,"},
		{"wifi-1010", 17, "bad_tenant", "Неизвестный арендатор: zone-404", "WIFI-1017,2026-10-10T14:35:00Z,visit,,79161234574,,zone-404"},
		{"bean-1030", 6, "no_external_id", "Не указан внешний идентификатор", ",2026-10-30T09:40:00Z,purchase,21000,79161234575,,"},
	}
}

func (s *Seeder) seedBatches(ctx context.Context) error {
	for _, b := range demoBatches() {
		id := demoID("import_batch", b.Key)
		s.batches[b.Key] = id
		createdAt := s.now.AddDate(0, 0, -b.DaysAgo)
		var startedAt, finishedAt any
		if b.Status != "queued" {
			startedAt = createdAt.Add(30 * time.Second)
		}
		if b.Status == "completed" || b.Status == "completed_with_errors" || b.Status == "failed" {
			finishedAt = createdAt.Add(2 * time.Minute)
		}
		var fileName, fileHash, fileSize any
		if b.Mode == "file" {
			fileName = b.FileName
			sum := sha256.Sum256([]byte("ucrp-demo/" + b.Key))
			fileHash = hex.EncodeToString(sum[:])
			fileSize = int64(b.Total * 96)
		}
		var windowFrom, windowTo any
		if b.Mode == "pull" {
			windowFrom = s.now.AddDate(0, 0, -b.WindowFrom)
			windowTo = s.now.AddDate(0, 0, -b.WindowTo)
		}
		var errMsg any
		if b.Error != "" {
			errMsg = b.Error
		}
		if err := s.exec(ctx, `
			INSERT INTO import_batch (id, data_source_id, mode, file_name, file_hash, file_size_bytes, window_from, window_to,
				status, total_rows, accepted_rows, rejected_rows, duplicate_rows, error_message, created_by, created_at, started_at, finished_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)
			ON CONFLICT (id) DO NOTHING`,
			id, s.sources[b.Source], b.Mode, fileName, fileHash, fileSize, windowFrom, windowTo,
			b.Status, b.Total, b.Accepted, b.Rejected, b.Duplicate, errMsg, s.users["analyst"],
			createdAt, startedAt, finishedAt); err != nil {
			return err
		}
	}
	for _, r := range demoRejectedRows() {
		if err := s.exec(ctx, `
			INSERT INTO rejected_row (id, import_batch_id, row_number, error_code, error_message, raw_line, created_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7)
			ON CONFLICT (id) DO NOTHING`,
			demoID("rejected_row", fmt.Sprintf("%s/%d", r.Batch, r.RowNumber)), s.batches[r.Batch], r.RowNumber,
			r.Code, r.Message, r.RawLine, s.now.AddDate(0, 0, -20)); err != nil {
			return err
		}
	}
	return nil
}
