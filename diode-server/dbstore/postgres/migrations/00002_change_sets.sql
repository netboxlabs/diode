-- +goose Up

-- Create the change_sets table
CREATE TABLE IF NOT EXISTS change_sets
(
    id               INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    external_id      VARCHAR(255) NOT NULL,
    ingestion_log_id INTEGER      NOT NULL,
    branch_id        VARCHAR(255),
    deviation_name   VARCHAR(255),
    created_at       TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indices
CREATE INDEX IF NOT EXISTS idx_change_sets_external_id ON change_sets (external_id);
CREATE INDEX IF NOT EXISTS idx_change_sets_ingestion_log_id ON change_sets (ingestion_log_id);
CREATE INDEX IF NOT EXISTS idx_change_sets_branch_id ON change_sets (branch_id);

-- Create the changes table
CREATE TABLE IF NOT EXISTS changes
(
    id                   INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    external_id          VARCHAR(255) NOT NULL,
    change_set_id        INTEGER      NOT NULL,
    change_type          VARCHAR(50)  NOT NULL,
    object_type          VARCHAR(255) NOT NULL,
    object_primary_value VARCHAR(255) NOT NULL,
    object_id            INTEGER,
    object_version       INTEGER,
    before               JSONB,
    after                JSONB,
    sequence_number      INTEGER,
    created_at           TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at           TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indices
CREATE INDEX IF NOT EXISTS idx_changes_external_id ON changes (external_id);
CREATE INDEX IF NOT EXISTS idx_changes_change_set_id ON changes (change_set_id);
CREATE INDEX IF NOT EXISTS idx_changes_change_type ON changes (change_type);
CREATE INDEX IF NOT EXISTS idx_changes_object_type ON changes (object_type);

-- Add foreign key constraints
ALTER TABLE change_sets
    ADD CONSTRAINT fk_change_sets_ingestion_logs FOREIGN KEY (ingestion_log_id) REFERENCES ingestion_logs (id) ON DELETE CASCADE;
ALTER TABLE changes
    ADD CONSTRAINT fk_changes_change_sets FOREIGN KEY (change_set_id) REFERENCES change_sets (id) ON DELETE CASCADE;

-- Create a view returning deviations
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

-- +goose Down

-- Drop the v_deviations view
DROP VIEW IF EXISTS v_deviations;

-- Drop the changes table
DROP TABLE changes;

-- Drop the change_sets table
DROP TABLE change_sets;
