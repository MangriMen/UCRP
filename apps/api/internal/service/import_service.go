package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/MangriMen/UCRP/internal/domain"
	"github.com/MangriMen/UCRP/internal/ingestion"
	"github.com/MangriMen/UCRP/internal/platform/idgen"
)

const (
	importMaxRows      = 100000
	importMaxFileBytes = 50 << 20
	jobImportProcess   = "import.process"
	jobConnectorPull   = "connector.pull"
)

type ImportError struct {
	Status  int
	Code    string
	Message string
}

func (e *ImportError) Error() string { return e.Code }

type ImportService struct {
	batches  domain.ImportBatchRepo
	rejected domain.RejectedRowRepo
	sources  domain.DataSourceRepo
	storage  domain.BlobStore
	pipeline *ingestion.Pipeline
	clock    domain.Clock
	idGen    domain.IDGen
}

func NewImportService(
	batches domain.ImportBatchRepo,
	rejected domain.RejectedRowRepo,
	sources domain.DataSourceRepo,
	events domain.EventRepo,
	catalog domain.ImportCatalog,
	storage domain.BlobStore,
) *ImportService {
	clock := domain.SystemClock{}
	idGen := idgen.UUID{}
	return &ImportService{
		batches:  batches,
		rejected: rejected,
		sources:  sources,
		storage:  storage,
		pipeline: ingestion.NewPipeline(batches, rejected, events, sources, catalog, clock, idGen),
		clock:    clock,
		idGen:    idGen,
	}
}

type ImportDTO struct {
	ID           string     `json:"id"`
	DataSourceID string     `json:"data_source_id"`
	Mode         string     `json:"mode"`
	Status       string     `json:"status"`
	FileName     string     `json:"file_name,omitempty"`
	Total        int        `json:"total"`
	Accepted     int        `json:"accepted"`
	Rejected     int        `json:"rejected"`
	Duplicate    int        `json:"duplicate"`
	ErrorMessage string     `json:"error_message,omitempty"`
	CreatedBy    string     `json:"created_by,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	FinishedAt   *time.Time `json:"finished_at,omitempty"`
}

type CreateImportResult struct {
	BatchID    string `json:"batch_id"`
	Status     string `json:"status"`
	Duplicate  bool   `json:"duplicate"`
	HTTPStatus int    `json:"-"`
}

type RejectedRowDTO struct {
	RowNumber    int    `json:"row_number"`
	ErrorCode    string `json:"error_code"`
	ErrorMessage string `json:"error_message"`
	RawLine      string `json:"raw_line"`
}

type CreateImportInput struct {
	DataSourceID string
	FileName     string
	File         io.Reader
	ActorUserID  string
}

func (s *ImportService) Create(ctx context.Context, in CreateImportInput) (*CreateImportResult, error) {
	if strings.TrimSpace(in.DataSourceID) == "" || in.File == nil {
		return nil, domain.ErrValidation
	}
	ds, err := s.sources.GetByID(ctx, in.DataSourceID)
	if err != nil {
		return nil, err
	}
	if !ds.IsActive {
		return nil, &ImportError{http.StatusUnprocessableEntity, "bad_format", "Источник деактивирован"}
	}
	if !ds.SupportsFile {
		return nil, &ImportError{http.StatusUnprocessableEntity, "bad_format", "Источник не поддерживает файлы"}
	}

	data, err := io.ReadAll(io.LimitReader(in.File, importMaxFileBytes+1))
	if err != nil {
		return nil, &ImportError{http.StatusUnprocessableEntity, "bad_format", "Не удалось прочитать файл"}
	}
	if int64(len(data)) > importMaxFileBytes {
		return nil, &ImportError{http.StatusRequestEntityTooLarge, "file_too_large", "Файл превышает допустимый размер"}
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, &ImportError{http.StatusUnprocessableEntity, "bad_format", "Пустой файл"}
	}
	mapping, err := ingestion.ParseFieldMapping(ds.FieldMapping)
	if err != nil {
		return nil, &ImportError{http.StatusUnprocessableEntity, "bad_format", "Некорректный маппинг полей источника"}
	}
	if err := ingestion.ValidateHeader(bytes.NewReader(data), mapping); err != nil {
		return nil, importReadError(err)
	}

	sum := sha256.Sum256(data)
	hash := hex.EncodeToString(sum[:])

	if existing, err := s.batches.FindByHash(ctx, in.DataSourceID, hash); err == nil {
		return &CreateImportResult{
			BatchID:    existing.ID,
			Status:     string(existing.Status),
			Duplicate:  true,
			HTTPStatus: http.StatusOK,
		}, nil
	} else if !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}

	batchID := s.idGen.NewID(ctx)
	storageKey := "imports/" + batchID + ".csv"
	if err := s.storage.Save(ctx, storageKey, bytes.NewReader(data)); err != nil {
		return nil, err
	}

	batch := &domain.ImportBatch{
		ID:            batchID,
		DataSourceID:  in.DataSourceID,
		Mode:          domain.ImportModeFile,
		FileName:      sanitizeFileName(in.FileName),
		FileHash:      hash,
		FileSizeBytes: int64(len(data)),
		Status:        domain.ImportQueued,
		CreatedBy:     in.ActorUserID,
		CreatedAt:     s.clock.Now(),
		StorageKey:    storageKey,
	}
	if err := s.batches.CreateQueued(ctx, batch, jobImportProcess, "import:"+batchID); err != nil {
		return nil, err
	}
	return &CreateImportResult{
		BatchID:    batchID,
		Status:     string(domain.ImportQueued),
		Duplicate:  false,
		HTTPStatus: http.StatusAccepted,
	}, nil
}

func (s *ImportService) Sync(ctx context.Context, dataSourceID, actorUserID string) (*CreateImportResult, error) {
	ds, err := s.sources.GetByID(ctx, dataSourceID)
	if err != nil {
		return nil, err
	}
	if !ds.SupportsPull {
		return nil, domain.ErrValidation
	}
	inProgress, err := s.batches.HasActivePull(ctx, dataSourceID)
	if err != nil {
		return nil, err
	}
	if inProgress {
		return nil, &ImportError{http.StatusConflict, "sync_in_progress", "Синхронизация уже идёт"}
	}
	now := s.clock.Now()
	from := now.Add(-24 * time.Hour)
	batch := &domain.ImportBatch{
		ID:           s.idGen.NewID(ctx),
		DataSourceID: dataSourceID,
		Mode:         domain.ImportModePull,
		Status:       domain.ImportQueued,
		CreatedBy:    actorUserID,
		CreatedAt:    now,
		WindowFrom:   &from,
		WindowTo:     &now,
	}
	dedup := "pull:" + dataSourceID + ":" + from.UTC().Format(time.RFC3339)
	if err := s.batches.CreateQueued(ctx, batch, jobConnectorPull, dedup); err != nil {
		return nil, err
	}
	return &CreateImportResult{
		BatchID:    batch.ID,
		Status:     string(domain.ImportQueued),
		HTTPStatus: http.StatusAccepted,
	}, nil
}

func (s *ImportService) List(ctx context.Context, dataSourceID, status string, limit, offset int) ([]ImportDTO, error) {
	items, err := s.batches.List(ctx, dataSourceID, status, limit, offset)
	if err != nil {
		return nil, err
	}
	out := make([]ImportDTO, 0, len(items))
	for _, b := range items {
		out = append(out, toImportDTO(b))
	}
	return out, nil
}

func (s *ImportService) Get(ctx context.Context, id string) (*ImportDTO, error) {
	b, err := s.batches.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	d := toImportDTO(*b)
	return &d, nil
}

func (s *ImportService) RejectedRows(ctx context.Context, batchID string, limit, offset int) ([]RejectedRowDTO, error) {
	if _, err := s.batches.GetByID(ctx, batchID); err != nil {
		return nil, err
	}
	rows, err := s.rejected.ListByBatch(ctx, batchID, limit, offset)
	if err != nil {
		return nil, err
	}
	out := make([]RejectedRowDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, RejectedRowDTO{
			RowNumber:    r.RowNumber,
			ErrorCode:    r.ErrorCode,
			ErrorMessage: r.ErrorMessage,
			RawLine:      r.RawLine,
		})
	}
	return out, nil
}

func (s *ImportService) ProcessBatch(ctx context.Context, batchID string) error {
	batch, err := s.batches.GetByID(ctx, batchID)
	if err != nil {
		return err
	}
	switch batch.Status {
	case domain.ImportCompleted, domain.ImportCompletedWithErrors, domain.ImportFailed:
		return nil
	}
	if batch.Mode == domain.ImportModePull {
		return s.processPull(ctx, batch)
	}
	return s.processFile(ctx, batch)
}

func (s *ImportService) processFile(ctx context.Context, batch *domain.ImportBatch) error {
	ds, err := s.sources.GetByID(ctx, batch.DataSourceID)
	if err != nil {
		return s.pipeline.Fail(ctx, batch.ID, "Источник не найден")
	}
	mapping, err := ingestion.ParseFieldMapping(ds.FieldMapping)
	if err != nil {
		return s.pipeline.Fail(ctx, batch.ID, "Некорректный маппинг полей источника")
	}
	rc, err := s.storage.Open(ctx, batch.StorageKey)
	if err != nil {
		return s.pipeline.Fail(ctx, batch.ID, "Файл не найден в хранилище")
	}
	defer rc.Close()
	rows, err := readCSV(rc, mapping)
	if err != nil {
		return s.pipeline.Fail(ctx, batch.ID, csvError(err).Message)
	}
	_, err = s.pipeline.ProcessRows(ctx, batch, rows, mapping)
	return err
}

func (s *ImportService) processPull(ctx context.Context, batch *domain.ImportBatch) error {
	ds, err := s.sources.GetByID(ctx, batch.DataSourceID)
	if err != nil {
		return s.pipeline.Fail(ctx, batch.ID, "Источник не найден")
	}
	var cfg struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal([]byte(ds.Config), &cfg); err != nil || cfg.URL == "" {
		return s.pipeline.Fail(ctx, batch.ID, "Для pull-источника не задан config.url")
	}
	mapping, err := ingestion.ParseFieldMapping(ds.FieldMapping)
	if err != nil {
		return s.pipeline.Fail(ctx, batch.ID, "Некорректный маппинг полей источника")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.URL, nil)
	if err != nil {
		return s.pipeline.Fail(ctx, batch.ID, "Некорректный URL источника")
	}
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return s.pipeline.Fail(ctx, batch.ID, "Не удалось получить данные источника")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return s.pipeline.Fail(ctx, batch.ID, fmt.Sprintf("Источник ответил %d", resp.StatusCode))
	}
	rows, err := readCSV(resp.Body, mapping)
	if err != nil {
		return s.pipeline.Fail(ctx, batch.ID, csvError(err).Message)
	}
	_, err = s.pipeline.ProcessRows(ctx, batch, rows, mapping)
	return err
}

func readCSV(r io.Reader, mapping ingestion.FieldMapping) ([]ingestion.Row, error) {
	rows, err := ingestion.NewCSVConnector(importMaxRows).Read(r, mapping)
	if err != nil {
		return nil, csvError(err)
	}
	return rows, nil
}

func toImportDTO(b domain.ImportBatch) ImportDTO {
	return ImportDTO{
		ID:           b.ID,
		DataSourceID: b.DataSourceID,
		Mode:         string(b.Mode),
		Status:       string(b.Status),
		FileName:     b.FileName,
		Total:        b.TotalRows,
		Accepted:     b.AcceptedRows,
		Rejected:     b.RejectedRows,
		Duplicate:    b.DuplicateRows,
		ErrorMessage: b.ErrorMessage,
		CreatedBy:    b.CreatedBy,
		CreatedAt:    b.CreatedAt,
		StartedAt:    b.StartedAt,
		FinishedAt:   b.FinishedAt,
	}
}

func sanitizeFileName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "upload.csv"
	}
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, "/", "_")
	if len(name) > 255 {
		name = name[:255]
	}
	return name
}

func importReadError(err error) error {
	ie := csvError(err)
	switch {
	case errors.Is(err, ingestion.ErrCSVFileTooLarge):
		return &ImportError{http.StatusRequestEntityTooLarge, ie.Code, ie.Message}
	default:
		return &ImportError{http.StatusUnprocessableEntity, ie.Code, ie.Message}
	}
}

func csvError(err error) *ImportError {
	switch {
	case errors.Is(err, ingestion.ErrCSVFileTooLarge):
		return &ImportError{http.StatusRequestEntityTooLarge, "file_too_large", "Больше 100 тыс. строк"}
	case errors.Is(err, ingestion.ErrCSVMissingColumns):
		return &ImportError{http.StatusUnprocessableEntity, "missing_columns", "Нет обязательных колонок"}
	default:
		return &ImportError{http.StatusUnprocessableEntity, "bad_format", "Неверный формат файла"}
	}
}
