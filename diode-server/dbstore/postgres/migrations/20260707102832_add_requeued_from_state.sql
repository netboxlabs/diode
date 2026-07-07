-- +goose Up
-- +goose StatementBegin
-- requeued_from_state records the terminal state an ingestion log was in when
-- a duplicate observation requeued it for re-plan (see BulkMarkDuplicates).
-- Consumed at persist time: a log requeued from APPLIED spawns a new
-- deviation on drift and is restored to APPLIED, preserving its history.
-- Cleared on every final state write.
ALTER TABLE ingestion_logs ADD COLUMN requeued_from_state INTEGER;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE ingestion_logs DROP COLUMN requeued_from_state;
-- +goose StatementEnd
