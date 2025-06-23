-- +goose Up

ALTER TABLE ingestion_logs
    ADD COLUMN entity_hash VARCHAR(64),
    ADD COLUMN duplicate_of_id INTEGER REFERENCES ingestion_logs(id),
    ADD COLUMN last_seen TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    ADD COLUMN duplicate_count INTEGER DEFAULT 0 NOT NULL;

CREATE INDEX idx_ingestion_logs_entity_hash_primary ON ingestion_logs(entity_hash) WHERE duplicate_of_id IS NULL;
CREATE INDEX idx_ingestion_logs_duplicate_of_id ON ingestion_logs(duplicate_of_id);

-- Must be recreated after adding duplicate_of_id column (not stored as query in pg)
-- It also must not alter the order of the columns in the view in order to be
-- performed as a CREATE OR REPLACE (vs a DROP and re-CREATE)
CREATE OR REPLACE VIEW v_deviations AS
SELECT DISTINCT ON (ingestion_logs.id)
    ingestion_logs.id,
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
    FILTER ( WHERE changes.id IS NOT NULL ) AS changes,
    ingestion_logs.duplicate_of_id,
    ingestion_logs.last_seen,
    ingestion_logs.duplicate_count
FROM ingestion_logs
         LEFT JOIN change_sets on ingestion_logs.id = change_sets.ingestion_log_id
         LEFT JOIN changes on change_sets.id = changes.change_set_id
GROUP BY ingestion_logs.id, change_sets.id
ORDER BY ingestion_logs.id DESC, change_sets.id DESC;

-- Create trigger function to maintain last_seen and duplicate_count
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_ingestion_logs_duplicate_stats()
RETURNS TRIGGER AS $$
DECLARE
    primary_id INTEGER;
    latest_created_at TIMESTAMP WITH TIME ZONE;
    dup_count INTEGER;
BEGIN
    -- Skip if this is a recursive trigger call (avoid infinite recursion)
    -- We detect this by checking if we're only updating last_seen/duplicate_count
    IF TG_OP = 'UPDATE' AND 
       OLD.duplicate_of_id IS NOT DISTINCT FROM NEW.duplicate_of_id AND
       OLD.created_at IS NOT DISTINCT FROM NEW.created_at AND
       OLD.entity_hash IS NOT DISTINCT FROM NEW.entity_hash THEN
        RETURN NEW;
    END IF;
    
    -- Determine the primary record ID
    IF TG_OP = 'INSERT' THEN
        primary_id := COALESCE(NEW.duplicate_of_id, NEW.id);
    ELSIF TG_OP = 'UPDATE' THEN
        primary_id := COALESCE(NEW.duplicate_of_id, NEW.id);
    END IF;
    
    -- Calculate latest created_at and duplicate count for the primary record
    SELECT 
        MAX(created_at),
        COUNT(*) - 1  -- Subtract 1 to exclude the primary record itself
    INTO latest_created_at, dup_count
    FROM ingestion_logs 
    WHERE id = primary_id OR duplicate_of_id = primary_id;
    
    -- Update only the primary record with duplicate stats
    -- This will trigger the UPDATE trigger again, but the recursion check above will prevent infinite loops
    UPDATE ingestion_logs 
    SET 
        last_seen = latest_created_at,
        duplicate_count = dup_count
    WHERE id = primary_id;
    
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- Create triggers
CREATE TRIGGER trg_ingestion_logs_duplicate_stats_insert
    AFTER INSERT ON ingestion_logs
    FOR EACH ROW
    EXECUTE FUNCTION update_ingestion_logs_duplicate_stats();

CREATE TRIGGER trg_ingestion_logs_duplicate_stats_update
    AFTER UPDATE ON ingestion_logs
    FOR EACH ROW
    EXECUTE FUNCTION update_ingestion_logs_duplicate_stats();

-- Initialize existing data: set last_seen to created_at for all records
UPDATE ingestion_logs SET last_seen = created_at WHERE last_seen IS NULL;


-- +goose Down

DROP VIEW IF EXISTS v_deviations;

DROP TRIGGER IF EXISTS trg_ingestion_logs_duplicate_stats_update ON ingestion_logs;
DROP TRIGGER IF EXISTS trg_ingestion_logs_duplicate_stats_insert ON ingestion_logs;
DROP FUNCTION IF EXISTS update_ingestion_logs_duplicate_stats();

DROP INDEX IF EXISTS idx_ingestion_logs_entity_hash_primary;
DROP INDEX IF EXISTS idx_ingestion_logs_duplicate_of_id;

ALTER TABLE ingestion_logs
    DROP COLUMN IF EXISTS entity_hash,
    DROP COLUMN IF EXISTS duplicate_of_id,
    DROP COLUMN IF EXISTS last_seen,
    DROP COLUMN IF EXISTS duplicate_count;

-- Rereate a view returning deviations without the duplicate_of_id column
CREATE VIEW v_deviations AS
SELECT DISTINCT ON (ingestion_logs.id) ingestion_logs.*,
                                       row_to_json(change_sets.*)              AS change_set,
                                       JSON_AGG(changes.* ORDER BY changes.sequence_number ASC)
                                       FILTER ( WHERE changes.id IS NOT NULL ) AS changes
FROM ingestion_logs
         LEFT JOIN change_sets on ingestion_logs.id = change_sets.ingestion_log_id
         LEFT JOIN changes on change_sets.id = changes.change_set_id
GROUP BY ingestion_logs.id, change_sets.id
ORDER BY ingestion_logs.id DESC, change_sets.id DESC;
