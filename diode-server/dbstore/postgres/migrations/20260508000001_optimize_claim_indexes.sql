-- +goose Up

-- Partial index for ClaimQueuedIngestionLogs: only indexes rows in state=1 (OPEN),
-- which is the transient work-queue state. Rows leave the index when they transition
-- to terminal states, keeping it small under sustained load.
CREATE INDEX IF NOT EXISTS idx_ingestion_logs_claim_open ON ingestion_logs (id) WHERE state = 1;

-- Drop single-column entity_hash index — redundant with the composite
-- (entity_hash, created_at DESC) index which covers all entity_hash lookups.
DROP INDEX IF EXISTS idx_ingestion_logs_entity_hash;

-- Drop request_id index — no query filters on request_id alone.
DROP INDEX IF EXISTS idx_ingestion_logs_request_id;

-- +goose Down

DROP INDEX IF EXISTS idx_ingestion_logs_claim_open;
CREATE INDEX IF NOT EXISTS idx_ingestion_logs_entity_hash ON ingestion_logs USING btree (entity_hash);
CREATE INDEX IF NOT EXISTS idx_ingestion_logs_request_id ON ingestion_logs USING btree (request_id);
