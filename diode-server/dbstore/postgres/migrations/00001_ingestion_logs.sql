-- +goose Up

-- Create the ingestion_logs table
CREATE TABLE IF NOT EXISTS ingestion_logs
(
    id                   SERIAL PRIMARY KEY,
    ingestion_log_ksuid  CHAR(27) NOT NULL,
    data_type            VARCHAR(255),
    state                INTEGER,
    request_id           VARCHAR(255),
    ingestion_ts         BIGINT,
    producer_app_name    VARCHAR(255),
    producer_app_version VARCHAR(255),
    sdk_name             VARCHAR(255),
    sdk_version          VARCHAR(255),
    entity               JSONB,              -- protojson output
    error                JSONB,
    source_metadata      JSONB,              -- kv for source metadata that came from agent, i.e. source, hostname, ip
    created_at           TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at           TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indices
CREATE INDEX IF NOT EXISTS idx_ingestion_logs_ingestion_log_ksuid ON ingestion_logs(ingestion_log_ksuid);
CREATE INDEX IF NOT EXISTS idx_ingestion_logs_data_type ON ingestion_logs(data_type);
CREATE INDEX IF NOT EXISTS idx_ingestion_logs_state ON ingestion_logs(state);
CREATE INDEX IF NOT EXISTS idx_ingestion_logs_request_id ON ingestion_logs(request_id);

-- +goose Down

-- Drop the ingestion_logs table
DROP TABLE ingestion_logs;
