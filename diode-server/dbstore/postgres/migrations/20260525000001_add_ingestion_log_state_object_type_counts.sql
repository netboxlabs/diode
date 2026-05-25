-- +goose Up
-- +goose StatementBegin

-- Counter table for fast SUM-based aggregation over (state, object_type).
-- Read path: list queries that filter by any subset of {state, object_type}
-- can answer their total via a single sub-millisecond index lookup instead
-- of a parallel sequential scan over the whole ingestion_logs table.
-- For tenants with several million rows that drops the count cost from
-- ~600ms per call to <1ms.
CREATE TABLE IF NOT EXISTS ingestion_log_state_object_type_counts (
    state       INT4   NOT NULL,
    object_type TEXT   NOT NULL,
    n           BIGINT NOT NULL,
    PRIMARY KEY (state, object_type)
);

-- Initial backfill from the current data. Runs once at migration time; after
-- this point the counter is maintained transactionally by the triggers below.
INSERT INTO ingestion_log_state_object_type_counts (state, object_type, n)
SELECT state, COALESCE(object_type, ''), COUNT(*)
FROM ingestion_logs
WHERE state IS NOT NULL
GROUP BY state, COALESCE(object_type, '')
ON CONFLICT (state, object_type) DO UPDATE SET n = EXCLUDED.n;

-- +goose StatementEnd

-- +goose StatementBegin
-- Statement-level triggers with transition tables. Each fires once per
-- INSERT / UPDATE / DELETE statement against ingestion_logs regardless of
-- how many rows it touches, so a 100-row bulk transition costs one trigger
-- invocation, not a hundred. The trigger body collapses the affected rows
-- into a small aggregated UPSERT.
--
-- Rows with NULL state are skipped: state is only NULL transiently during
-- creation; the existing list queries filter `state = ANY(...)` which also
-- excludes nulls, so counter semantics match.

CREATE OR REPLACE FUNCTION ingestion_log_state_object_type_counts_on_insert()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO ingestion_log_state_object_type_counts (state, object_type, n)
    SELECT state, COALESCE(object_type, ''), COUNT(*)
    FROM new_rows
    WHERE state IS NOT NULL
    GROUP BY state, COALESCE(object_type, '')
    ON CONFLICT (state, object_type)
    DO UPDATE SET n = ingestion_log_state_object_type_counts.n + EXCLUDED.n;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION ingestion_log_state_object_type_counts_on_update()
RETURNS TRIGGER AS $$
BEGIN
    -- Aggregate the net delta per (state, object_type) cell. A row that
    -- changes state from A to B contributes -1 to (A, *) and +1 to (B, *);
    -- a row whose (state, object_type) is unchanged nets to zero and is
    -- filtered out by the HAVING clause.
    WITH deltas AS (
        SELECT state, COALESCE(object_type, '') AS object_type, -COUNT(*) AS delta
        FROM old_rows
        WHERE state IS NOT NULL
        GROUP BY state, COALESCE(object_type, '')
        UNION ALL
        SELECT state, COALESCE(object_type, ''), COUNT(*)
        FROM new_rows
        WHERE state IS NOT NULL
        GROUP BY state, COALESCE(object_type, '')
    ),
    net AS (
        SELECT state, object_type, SUM(delta) AS delta
        FROM deltas
        GROUP BY state, object_type
        HAVING SUM(delta) <> 0
    )
    INSERT INTO ingestion_log_state_object_type_counts (state, object_type, n)
    SELECT state, object_type, delta FROM net
    ON CONFLICT (state, object_type)
    DO UPDATE SET n = ingestion_log_state_object_type_counts.n + EXCLUDED.n;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION ingestion_log_state_object_type_counts_on_delete()
RETURNS TRIGGER AS $$
BEGIN
    WITH deltas AS (
        SELECT state, COALESCE(object_type, '') AS object_type, COUNT(*) AS n
        FROM old_rows
        WHERE state IS NOT NULL
        GROUP BY state, COALESCE(object_type, '')
    )
    UPDATE ingestion_log_state_object_type_counts c
    SET n = c.n - d.n
    FROM deltas d
    WHERE c.state = d.state AND c.object_type = d.object_type;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
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
-- Reconciliation function: replace the counter contents with the truth as
-- of NOW(). Called at reconciler startup as a defense-in-depth check against
-- drift (admin SQL that bypassed the trigger, replication anomalies, etc.).
-- Bounded to a single transaction so concurrent readers either see the old
-- value or the new value, not partial state.
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

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS ingestion_log_state_object_type_counts_after_delete ON ingestion_logs;
DROP TRIGGER IF EXISTS ingestion_log_state_object_type_counts_after_update ON ingestion_logs;
DROP TRIGGER IF EXISTS ingestion_log_state_object_type_counts_after_insert ON ingestion_logs;
DROP FUNCTION IF EXISTS rebuild_ingestion_log_state_object_type_counts();
DROP FUNCTION IF EXISTS ingestion_log_state_object_type_counts_on_delete();
DROP FUNCTION IF EXISTS ingestion_log_state_object_type_counts_on_update();
DROP FUNCTION IF EXISTS ingestion_log_state_object_type_counts_on_insert();
DROP TABLE IF EXISTS ingestion_log_state_object_type_counts;
-- +goose StatementEnd
