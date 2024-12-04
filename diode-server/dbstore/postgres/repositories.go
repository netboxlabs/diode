package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/netboxlabs/diode/diode-server/gen/dbstore/postgres"
	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/reconcilerpb"
)

// IngestionLogRepository allows interacting with ingestion logs.
type IngestionLogRepository struct {
	pool    *pgxpool.Pool
	queries *postgres.Queries
}

// NewIngestionLogRepository creates a new IngestionLogRepository.
func NewIngestionLogRepository(pool *pgxpool.Pool) *IngestionLogRepository {
	return &IngestionLogRepository{
		pool:    pool,
		queries: postgres.New(pool),
	}
}

// CreateIngestionLog creates a new ingestion log.
func (r *IngestionLogRepository) CreateIngestionLog(ctx context.Context, ingestionLog *reconcilerpb.IngestionLog, sourceMetadata []byte) error {
	entityJSON, err := protojson.Marshal(ingestionLog.Entity)
	if err != nil {
		return fmt.Errorf("failed to marshal entity: %w", err)
	}
	params := postgres.CreateIngestionLogParams{
		IngestionLogKsuid:  ingestionLog.Id,
		DataType:           pgtype.Text{String: ingestionLog.DataType, Valid: true},
		State:              pgtype.Int4{Int32: int32(ingestionLog.State), Valid: true},
		RequestID:          pgtype.Text{String: ingestionLog.RequestId, Valid: true},
		IngestionTs:        pgtype.Int8{Int64: ingestionLog.IngestionTs, Valid: true},
		ProducerAppName:    pgtype.Text{String: ingestionLog.ProducerAppName, Valid: true},
		ProducerAppVersion: pgtype.Text{String: ingestionLog.ProducerAppVersion, Valid: true},
		SdkName:            pgtype.Text{String: ingestionLog.SdkName, Valid: true},
		SdkVersion:         pgtype.Text{String: ingestionLog.SdkVersion, Valid: true},
		Entity:             entityJSON,
		SourceMetadata:     sourceMetadata,
	}

	_, err = r.queries.CreateIngestionLog(ctx, params)
	return err
}

// ChangeSetRepository allows interacting with change sets.
type ChangeSetRepository struct {
	queries *postgres.Queries
}

// NewChangeSetRepository creates a new ChangeSetRepository.
func NewChangeSetRepository(db postgres.DBTX) *ChangeSetRepository {
	return &ChangeSetRepository{
		queries: postgres.New(db),
	}
}

// CreateChangeSet creates a new change set.
func (r *ChangeSetRepository) CreateChangeSet(_ context.Context, _ *reconcilerpb.ChangeSet) error {
	return errors.New("not implemented")
}
