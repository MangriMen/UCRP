-- +goose Up

ALTER TABLE export_job DROP CONSTRAINT export_job_format_check;
ALTER TABLE export_job ADD CONSTRAINT export_job_format_check CHECK (format IN ('csv','api','xlsx'));

-- +goose Down

ALTER TABLE export_job DROP CONSTRAINT export_job_format_check;
ALTER TABLE export_job ADD CONSTRAINT export_job_format_check CHECK (format IN ('csv','api'));
