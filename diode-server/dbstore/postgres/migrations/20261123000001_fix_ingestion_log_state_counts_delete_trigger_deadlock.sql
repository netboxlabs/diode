-- +goose Up

-- The ingestion_log_state_object_type_counts delete trigger from
-- 20260525000001 decrements the counter cells of the deleted rows with an
-- UPDATE ... FROM, which locks the cells in whatever order the join yields.
-- 20260528000001 made the insert and update triggers lock their cells in
-- (state, object_type) order. A DELETE whose rows span several cells could
-- still take them out of that order and deadlock (SQLSTATE 40P01) with a
-- concurrent state transition.
--
-- The delete trigger now locks its cells in (state, object_type) order
-- before it updates them. It keeps the UPDATE instead of the INSERT ... ON
-- CONFLICT form of the other two triggers, so that a cell missing through
-- counter drift stays missing instead of getting a negative count.

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION ingestion_log_state_object_type_counts_on_delete()
RETURNS TRIGGER AS $$
BEGIN
    PERFORM 1
    FROM ingestion_log_state_object_type_counts c
    WHERE (c.state, c.object_type) IN (
        SELECT state, COALESCE(object_type, '')
        FROM old_rows
        WHERE state IS NOT NULL
    )
    ORDER BY c.state, c.object_type
    FOR UPDATE OF c;

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

-- +goose Down

-- Restore the unordered body from 20260525000001.

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
