-- +goose Up

-- Create the ingestion_logs table
CREATE TABLE IF NOT EXISTS ingestion_logs
(
    id                   INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    external_id          VARCHAR(255) NOT NULL,
    object_type          VARCHAR(255),
    state                INTEGER,
    request_id           VARCHAR(255),
    ingestion_ts         BIGINT,
    producer_app_name    VARCHAR(255),
    producer_app_version VARCHAR(255),
    sdk_name             VARCHAR(255),
    sdk_version          VARCHAR(255),
    entity               JSONB,
    error                JSONB,
    source_metadata      JSONB,
    created_at           TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at           TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indices
CREATE INDEX IF NOT EXISTS idx_ingestion_logs_external_id ON ingestion_logs (external_id);
CREATE INDEX IF NOT EXISTS idx_ingestion_logs_object_type ON ingestion_logs (object_type);
CREATE INDEX IF NOT EXISTS idx_ingestion_logs_state ON ingestion_logs (state);
CREATE INDEX IF NOT EXISTS idx_ingestion_logs_request_id ON ingestion_logs (request_id);

-- +goose Down

-- Drop the ingestion_logs table
DROP TABLE ingestion_logs;
