-- +goose Up

-- Create the change_sets table
CREATE TABLE IF NOT EXISTS change_sets
(
    id               INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    change_set_uuid  VARCHAR(255) NOT NULL,
    ingestion_log_id INTEGER      NOT NULL,
    branch_id        VARCHAR(255),
    created_at       TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indices
CREATE INDEX IF NOT EXISTS idx_change_sets_change_set_uuid ON change_sets (change_set_uuid);

-- Create the changes table
CREATE TABLE IF NOT EXISTS changes
(
    id              INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    change_uuid     VARCHAR(255) NOT NULL,
    change_set_id   INTEGER      NOT NULL,
    change_type     VARCHAR(50)  NOT NULL,
    object_type     VARCHAR(100) NOT NULL,
    object_id       INTEGER,
    object_version  INTEGER,
    data            JSONB,
    sequence_number INTEGER,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indices
CREATE INDEX IF NOT EXISTS idx_changes_change_uuid ON changes (change_uuid);
CREATE INDEX IF NOT EXISTS idx_changes_change_set_id ON changes (change_set_id);
CREATE INDEX IF NOT EXISTS idx_changes_change_type ON changes (change_type);
CREATE INDEX IF NOT EXISTS idx_changes_object_type ON changes (object_type);

-- Add foreign key constraints
ALTER TABLE change_sets
    ADD CONSTRAINT fk_change_sets_ingestion_logs FOREIGN KEY (ingestion_log_id) REFERENCES ingestion_logs (id);
ALTER TABLE changes
    ADD CONSTRAINT fk_changes_change_sets FOREIGN KEY (change_set_id) REFERENCES change_sets (id);

-- Create a view to join ingestion_logs with change_sets
CREATE VIEW v_ingestion_logs_change_sets AS
(
SELECT change_sets.*
FROM ingestion_logs
         LEFT JOIN change_sets on ingestion_logs.id = change_sets.ingestion_log_id
    );

-- Create a view to join change_sets with changes
CREATE VIEW v_change_sets_changes AS
(
SELECT changes.*
FROM change_sets
         LEFT JOIN changes on change_sets.id = changes.change_set_id
    );

-- +goose Down

-- Drop the v_ingestion_logs_change_sets view
DROP VIEW IF EXISTS v_ingestion_logs_change_sets;

-- Drop the v_change_sets_with_changes view
DROP VIEW IF EXISTS v_change_sets_with_changes;

-- Drop the changes table
DROP TABLE changes;

-- Drop the change_sets table
DROP TABLE change_sets;
