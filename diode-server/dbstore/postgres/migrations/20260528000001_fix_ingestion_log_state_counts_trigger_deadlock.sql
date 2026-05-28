-- +goose Up

-- The ingestion_log_state_object_type_counts triggers introduced in
-- 20260525000001 use multi-row `INSERT ... ON CONFLICT DO UPDATE` to
-- maintain the (state, object_type) counter cells. Their SELECT had no
-- ORDER BY, so Postgres acquired row locks on the counter table in the
-- non-deterministic order returned by the hash aggregation. Concurrent
-- statement-level trigger invocations (typically two ClaimQueuedForAutoApply
-- transitions, both 1 -> 8 across overlapping object_types) could then form
-- a lock cycle on the same (state, object_type) cells and deadlock with
-- SQLSTATE 40P01.
--
-- The fix is the documented Postgres mitigation for this anti-pattern:
-- ORDER BY the conflict target so every concurrent invocation acquires
-- counter row locks in the same global order. Output rows are unchanged;
-- only the lock-acquisition order is constrained.
--
-- CREATE OR REPLACE FUNCTION swaps the function body in place; existing
-- triggers pick up the new definition on the next statement, so no
-- ALTER TRIGGER or table change is needed.

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION ingestion_log_state_object_type_counts_on_insert()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO ingestion_log_state_object_type_counts (state, object_type, n)
    SELECT state, COALESCE(object_type, '') AS object_type, COUNT(*)
    FROM new_rows
    WHERE state IS NOT NULL
    GROUP BY state, COALESCE(object_type, '')
    ORDER BY state, object_type
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
    ORDER BY state, object_type
    ON CONFLICT (state, object_type)
    DO UPDATE SET n = ingestion_log_state_object_type_counts.n + EXCLUDED.n;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose Down

-- Restore the original (pre-fix) function bodies from 20260525000001.
-- The Down path reintroduces the deadlock risk; it exists only to make
-- the migration reversible during local testing.

-- +goose StatementBegin
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
