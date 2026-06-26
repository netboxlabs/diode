-- +goose Up

-- Track graph-upsert progress per ingestion log. Graph upsert runs on a
-- separate processor (GraphUpsertProcessor) that is independent of the
-- ingestion state machine; using its own columns keeps the two concerns
-- decoupled and lets graph upsert run on rows in any state.
--
-- graph_upserted_at        — set when graph upsert completes successfully
-- graph_upsert_attempts    — number of attempts so far (terminal at the
--                            processor-configured max)
-- graph_upsert_claimed_at  — soft lease for the worker that currently owns
--                            the row; cleared on startup so a crashed worker
--                            does not leak its claim
ALTER TABLE ingestion_logs
    ADD COLUMN graph_upserted_at       TIMESTAMP WITH TIME ZONE,
    ADD COLUMN graph_upsert_attempts   INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN graph_upsert_claimed_at TIMESTAMP WITH TIME ZONE;

-- Partial index for the claim query: only un-upserted, un-claimed rows are
-- ever scanned. Terminal-failed rows fall out of the index once attempts
-- exceeds the processor's max-attempts setting; the WHERE clause in the
-- claim query handles that filter.
CREATE INDEX IF NOT EXISTS idx_ingestion_logs_graph_upsert_pending
    ON ingestion_logs (id)
    WHERE graph_upserted_at IS NULL AND graph_upsert_claimed_at IS NULL;

-- +goose Down

DROP INDEX IF EXISTS idx_ingestion_logs_graph_upsert_pending;

ALTER TABLE ingestion_logs
    DROP COLUMN IF EXISTS graph_upsert_claimed_at,
    DROP COLUMN IF EXISTS graph_upsert_attempts,
    DROP COLUMN IF EXISTS graph_upserted_at;
