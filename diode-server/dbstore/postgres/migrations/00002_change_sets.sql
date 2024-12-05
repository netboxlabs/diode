-- +goose Up

-- Create the change_sets table
CREATE TABLE IF NOT EXISTS change_sets
(
    id               INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    change_set_ksuid VARCHAR(27) NOT NULL,
    ingestion_log_id INTEGER     NOT NULL,
    branch_name      VARCHAR(255),
    created_at       TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indices
CREATE INDEX IF NOT EXISTS idx_change_sets_change_set_ksuid ON change_sets (change_set_ksuid);

-- Create the changes table
CREATE TABLE IF NOT EXISTS changes
(
    id              INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    change_ksuid    VARCHAR(27)  NOT NULL,
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
CREATE INDEX IF NOT EXISTS idx_changes_change_ksuid ON changes (change_ksuid);
CREATE INDEX IF NOT EXISTS idx_changes_change_set_id ON changes (change_set_id);
CREATE INDEX IF NOT EXISTS idx_changes_change_type ON changes (change_type);
CREATE INDEX IF NOT EXISTS idx_changes_object_type ON changes (object_type);

-- Add foreign key constraints
ALTER TABLE change_sets
    ADD CONSTRAINT fk_change_sets_ingestion_logs FOREIGN KEY (ingestion_log_id) REFERENCES ingestion_logs (id);
ALTER TABLE changes
    ADD CONSTRAINT fk_changes_change_sets FOREIGN KEY (change_set_id) REFERENCES change_sets (id);

-- +goose Down

-- Drop the changes table
DROP TABLE changes;

-- Drop the change_sets table
DROP TABLE change_sets;
