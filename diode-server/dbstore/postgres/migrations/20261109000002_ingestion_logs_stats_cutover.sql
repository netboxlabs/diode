-- +goose Up

-- Switch the changing columns of ingestion_logs over to ingestion_logs_stats
-- (created and filled by the previous migration). One transaction under an
-- exclusive lock: catch up rows written since the backfill, move the
-- triggers, drop the moved columns and their indexes, and add the
-- v_ingestion_logs view that readers use for the old row shape.
--
-- DROP COLUMN only marks the columns dropped; existing rows keep their bytes
-- until they are rewritten (VACUUM FULL or pg_repack). New rows are narrow.
LOCK TABLE ingestion_logs IN ACCESS EXCLUSIVE MODE;

-- Rows inserted, or committed with a lower id, after the backfill passed them.
INSERT INTO ingestion_logs_stats (ingestion_log_id, state, error, requeued_from_state, duplicate_count,
                                  last_seen, updated_at, object_type, ingestion_ts)
SELECT il.id, il.state, il.error, il.requeued_from_state, il.duplicate_count, il.last_seen, il.updated_at,
       il.object_type, il.ingestion_ts
FROM ingestion_logs il
WHERE NOT EXISTS (SELECT 1 FROM ingestion_logs_stats s WHERE s.ingestion_log_id = il.id);

-- Rows updated after the backfill copied them: every UPDATE moved updated_at.
UPDATE ingestion_logs_stats s
SET state               = il.state,
    error               = il.error,
    requeued_from_state = il.requeued_from_state,
    duplicate_count     = il.duplicate_count,
    last_seen           = il.last_seen,
    updated_at          = il.updated_at
FROM ingestion_logs il
WHERE il.id = s.ingestion_log_id
  AND il.updated_at IS DISTINCT FROM s.updated_at;

-- The updated_at and counter triggers move to the stats table. The counter
-- functions read only state and object_type from their transition tables,
-- which both tables carry, so they are reused as they are.
DROP TRIGGER IF EXISTS update_ingestion_logs_updated_at ON ingestion_logs;
DROP TRIGGER IF EXISTS ingestion_log_state_object_type_counts_after_insert ON ingestion_logs;
DROP TRIGGER IF EXISTS ingestion_log_state_object_type_counts_after_update ON ingestion_logs;
DROP TRIGGER IF EXISTS ingestion_log_state_object_type_counts_after_delete ON ingestion_logs;

-- +goose StatementBegin
CREATE TRIGGER update_ingestion_logs_stats_updated_at
    BEFORE UPDATE ON ingestion_logs_stats
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER ingestion_log_state_object_type_counts_after_insert
AFTER INSERT ON ingestion_logs_stats
REFERENCING NEW TABLE AS new_rows
FOR EACH STATEMENT
EXECUTE FUNCTION ingestion_log_state_object_type_counts_on_insert();
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER ingestion_log_state_object_type_counts_after_update
AFTER UPDATE ON ingestion_logs_stats
REFERENCING OLD TABLE AS old_rows NEW TABLE AS new_rows
FOR EACH STATEMENT
EXECUTE FUNCTION ingestion_log_state_object_type_counts_on_update();
-- +goose StatementEnd

-- Deleting an ingestion log cascades to its stats row, which fires this.
-- +goose StatementBegin
CREATE TRIGGER ingestion_log_state_object_type_counts_after_delete
AFTER DELETE ON ingestion_logs_stats
REFERENCING OLD TABLE AS old_rows
FOR EACH STATEMENT
EXECUTE FUNCTION ingestion_log_state_object_type_counts_on_delete();
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION rebuild_ingestion_log_state_object_type_counts()
RETURNS VOID AS $$
BEGIN
    LOCK TABLE ingestion_log_state_object_type_counts IN EXCLUSIVE MODE;
    DELETE FROM ingestion_log_state_object_type_counts;
    INSERT INTO ingestion_log_state_object_type_counts (state, object_type, n)
    SELECT state, COALESCE(object_type, ''), COUNT(*)
    FROM ingestion_logs_stats
    WHERE state IS NOT NULL
    GROUP BY state, COALESCE(object_type, '');
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

DROP VIEW IF EXISTS v_deviations;

DROP INDEX IF EXISTS idx_ingestion_logs_state_id;
DROP INDEX IF EXISTS idx_ingestion_logs_claim_open;
DROP INDEX IF EXISTS idx_ingestion_logs_applying;
DROP INDEX IF EXISTS idx_ingestion_logs_updated_at;
DROP INDEX IF EXISTS idx_ingestion_logs_state_ingestion_ts;
DROP INDEX IF EXISTS idx_ingestion_logs_state_updated_at;
DROP INDEX IF EXISTS idx_ingestion_logs_updated_at_id_desc;

-- CASCADE: views of other components that select these columns (diode-pro's
-- v_deviations_with_deviation_type) are dropped here and rebuilt by their
-- own migrations on top of the stats table.
ALTER TABLE ingestion_logs
    DROP COLUMN state CASCADE,
    DROP COLUMN error CASCADE,
    DROP COLUMN requeued_from_state CASCADE,
    DROP COLUMN duplicate_count CASCADE,
    DROP COLUMN last_seen CASCADE,
    DROP COLUMN updated_at CASCADE;

-- The pre-split row shape, column for column, for readers.
CREATE VIEW v_ingestion_logs AS
SELECT il.id,
       il.external_id,
       il.object_type,
       s.state,
       il.request_id,
       il.ingestion_ts,
       il.source_ts,
       il.producer_app_name,
       il.producer_app_version,
       il.sdk_name,
       il.sdk_version,
       il.entity,
       s.error,
       il.source_metadata,
       il.created_at,
       s.updated_at,
       il.entity_hash,
       s.last_seen,
       s.duplicate_count,
       s.requeued_from_state
FROM ingestion_logs il
JOIN ingestion_logs_stats s ON s.ingestion_log_id = il.id;

-- Same columns as before the split (the ingestion_logs columns of 2025).
-- Built on the base tables so GROUP BY can rely on both primary keys.
CREATE VIEW v_deviations AS
SELECT DISTINCT ON (il.id) il.id,
                           il.external_id,
                           il.object_type,
                           s.state,
                           il.request_id,
                           il.ingestion_ts,
                           il.source_ts,
                           il.producer_app_name,
                           il.producer_app_version,
                           il.sdk_name,
                           il.sdk_version,
                           il.entity,
                           s.error,
                           il.source_metadata,
                           il.created_at,
                           s.updated_at,
                           row_to_json(change_sets.*)              AS change_set,
                           JSON_AGG(changes.* ORDER BY changes.sequence_number ASC)
                           FILTER ( WHERE changes.id IS NOT NULL ) AS changes
FROM ingestion_logs il
         JOIN ingestion_logs_stats s ON s.ingestion_log_id = il.id
         LEFT JOIN change_sets ON il.id = change_sets.ingestion_log_id
         LEFT JOIN changes ON change_sets.id = changes.change_set_id
GROUP BY il.id, s.ingestion_log_id, change_sets.id
ORDER BY il.id DESC, change_sets.id DESC;

-- +goose Down

-- Restores the columns with their values. They come back at the end of the
-- column list, so the physical column order differs from the pre-split one.
LOCK TABLE ingestion_logs IN ACCESS EXCLUSIVE MODE;

DROP VIEW IF EXISTS v_deviations;
DROP VIEW IF EXISTS v_ingestion_logs;

ALTER TABLE ingestion_logs
    ADD COLUMN state INTEGER,
    ADD COLUMN error JSONB,
    ADD COLUMN updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    ADD COLUMN last_seen TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    ADD COLUMN duplicate_count INTEGER DEFAULT 0 NOT NULL,
    ADD COLUMN requeued_from_state INTEGER;

UPDATE ingestion_logs il
SET state               = s.state,
    error               = s.error,
    updated_at          = s.updated_at,
    last_seen           = s.last_seen,
    duplicate_count     = s.duplicate_count,
    requeued_from_state = s.requeued_from_state
FROM ingestion_logs_stats s
WHERE s.ingestion_log_id = il.id;

DROP TRIGGER IF EXISTS update_ingestion_logs_stats_updated_at ON ingestion_logs_stats;
DROP TRIGGER IF EXISTS ingestion_log_state_object_type_counts_after_insert ON ingestion_logs_stats;
DROP TRIGGER IF EXISTS ingestion_log_state_object_type_counts_after_update ON ingestion_logs_stats;
DROP TRIGGER IF EXISTS ingestion_log_state_object_type_counts_after_delete ON ingestion_logs_stats;

-- +goose StatementBegin
CREATE TRIGGER update_ingestion_logs_updated_at
    BEFORE UPDATE ON ingestion_logs
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER ingestion_log_state_object_type_counts_after_insert
AFTER INSERT ON ingestion_logs
REFERENCING NEW TABLE AS new_rows
FOR EACH STATEMENT
EXECUTE FUNCTION ingestion_log_state_object_type_counts_on_insert();
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER ingestion_log_state_object_type_counts_after_update
AFTER UPDATE ON ingestion_logs
REFERENCING OLD TABLE AS old_rows NEW TABLE AS new_rows
FOR EACH STATEMENT
EXECUTE FUNCTION ingestion_log_state_object_type_counts_on_update();
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER ingestion_log_state_object_type_counts_after_delete
AFTER DELETE ON ingestion_logs
REFERENCING OLD TABLE AS old_rows
FOR EACH STATEMENT
EXECUTE FUNCTION ingestion_log_state_object_type_counts_on_delete();
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION rebuild_ingestion_log_state_object_type_counts()
RETURNS VOID AS $$
BEGIN
    LOCK TABLE ingestion_log_state_object_type_counts IN EXCLUSIVE MODE;
    DELETE FROM ingestion_log_state_object_type_counts;
    INSERT INTO ingestion_log_state_object_type_counts (state, object_type, n)
    SELECT state, COALESCE(object_type, ''), COUNT(*)
    FROM ingestion_logs
    WHERE state IS NOT NULL
    GROUP BY state, COALESCE(object_type, '');
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE INDEX IF NOT EXISTS idx_ingestion_logs_state_id ON ingestion_logs (state, id);
CREATE INDEX IF NOT EXISTS idx_ingestion_logs_claim_open ON ingestion_logs (id) WHERE state = 1;
CREATE INDEX IF NOT EXISTS idx_ingestion_logs_applying ON ingestion_logs (id) WHERE state = 8;
CREATE INDEX IF NOT EXISTS idx_ingestion_logs_updated_at ON ingestion_logs (updated_at);
CREATE INDEX IF NOT EXISTS idx_ingestion_logs_state_ingestion_ts ON ingestion_logs (state, ingestion_ts);
CREATE INDEX IF NOT EXISTS idx_ingestion_logs_state_updated_at ON ingestion_logs (state, updated_at);
CREATE INDEX IF NOT EXISTS idx_ingestion_logs_updated_at_id_desc ON ingestion_logs (updated_at DESC, id DESC);

CREATE VIEW v_deviations AS
SELECT DISTINCT ON (ingestion_logs.id) ingestion_logs.id,
                                       ingestion_logs.external_id,
                                       ingestion_logs.object_type,
                                       ingestion_logs.state,
                                       ingestion_logs.request_id,
                                       ingestion_logs.ingestion_ts,
                                       ingestion_logs.source_ts,
                                       ingestion_logs.producer_app_name,
                                       ingestion_logs.producer_app_version,
                                       ingestion_logs.sdk_name,
                                       ingestion_logs.sdk_version,
                                       ingestion_logs.entity,
                                       ingestion_logs.error,
                                       ingestion_logs.source_metadata,
                                       ingestion_logs.created_at,
                                       ingestion_logs.updated_at,
                                       row_to_json(change_sets.*)              AS change_set,
                                       JSON_AGG(changes.* ORDER BY changes.sequence_number ASC)
                                       FILTER ( WHERE changes.id IS NOT NULL ) AS changes
FROM ingestion_logs
         LEFT JOIN change_sets on ingestion_logs.id = change_sets.ingestion_log_id
         LEFT JOIN changes on change_sets.id = changes.change_set_id
GROUP BY ingestion_logs.id, change_sets.id
ORDER BY ingestion_logs.id DESC, change_sets.id DESC;
