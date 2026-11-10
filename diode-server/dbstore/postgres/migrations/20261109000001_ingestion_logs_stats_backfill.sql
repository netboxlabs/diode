-- +goose NO TRANSACTION
-- +goose Up

-- ingestion_logs_stats holds the columns of an ingestion log that change
-- after insert, one narrow row per log. Re-ingesting an unchanged entity
-- updates its log up to three times (duplicate bookkeeping and requeue,
-- claim, final state); each UPDATE writes a new version of the whole row,
-- entity JSON included, plus an entry in every index, because the
-- updated_at trigger makes no update HOT. With the changing columns in
-- their own table, ingestion_logs becomes insert-only and each update
-- rewrites a ~90-byte row instead.
--
-- object_type and ingestion_ts are immutable copies, written once at
-- insert, so the (state, object_type) counters and the (state, ingestion_ts)
-- index stay on one table.
--
-- This migration creates and fills the table without blocking writers:
-- it runs outside a transaction and commits every batch, so a restart
-- resumes where it stopped. The next migration takes the lock, catches up
-- rows written meanwhile and switches over.
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS ingestion_logs_stats
(
    ingestion_log_id    INTEGER PRIMARY KEY REFERENCES ingestion_logs (id) ON DELETE CASCADE,
    state               INTEGER,
    error               JSONB,
    requeued_from_state INTEGER,
    duplicate_count     INTEGER                  DEFAULT 0 NOT NULL,
    last_seen           TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    object_type         VARCHAR(255),
    ingestion_ts        BIGINT
);
-- +goose StatementEnd

-- +goose StatementBegin
DO $$
DECLARE
    batch_size CONSTANT INTEGER := 10000;
    last_id INTEGER;
    max_id INTEGER;
BEGIN
    SELECT coalesce(max(ingestion_log_id), 0) INTO last_id FROM ingestion_logs_stats;
    SELECT coalesce(max(id), 0) INTO max_id FROM ingestion_logs;
    WHILE last_id < max_id LOOP
        INSERT INTO ingestion_logs_stats (ingestion_log_id, state, error, requeued_from_state, duplicate_count,
                                          last_seen, updated_at, object_type, ingestion_ts)
        SELECT id, state, error, requeued_from_state, duplicate_count, last_seen, updated_at, object_type, ingestion_ts
        FROM ingestion_logs
        WHERE id > last_id AND id <= last_id + batch_size
        ON CONFLICT (ingestion_log_id) DO NOTHING;
        last_id := last_id + batch_size;
        COMMIT;
    END LOOP;
END
$$;
-- +goose StatementEnd

-- The indexes that involve state or updated_at, rebuilt on the narrow table
-- after the bulk load and outside the switch-over lock. They mirror the
-- ones the cutover migration drops from ingestion_logs one for one.
CREATE INDEX IF NOT EXISTS idx_ingestion_logs_stats_state_id ON ingestion_logs_stats (state, ingestion_log_id);
CREATE INDEX IF NOT EXISTS idx_ingestion_logs_stats_claim_open ON ingestion_logs_stats (ingestion_log_id) WHERE state = 1;
CREATE INDEX IF NOT EXISTS idx_ingestion_logs_stats_applying ON ingestion_logs_stats (ingestion_log_id) WHERE state = 8;
CREATE INDEX IF NOT EXISTS idx_ingestion_logs_stats_updated_at ON ingestion_logs_stats (updated_at);
CREATE INDEX IF NOT EXISTS idx_ingestion_logs_stats_state_ingestion_ts ON ingestion_logs_stats (state, ingestion_ts);
CREATE INDEX IF NOT EXISTS idx_ingestion_logs_stats_state_updated_at ON ingestion_logs_stats (state, updated_at);
CREATE INDEX IF NOT EXISTS idx_ingestion_logs_stats_updated_at_id_desc ON ingestion_logs_stats (updated_at DESC, ingestion_log_id DESC);

-- +goose Down

DROP TABLE IF EXISTS ingestion_logs_stats;
