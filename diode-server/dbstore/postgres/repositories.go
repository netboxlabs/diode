package postgres

import (
	"context"
	"errors"

	"github.com/netboxlabs/diode/diode-server/gen/dbstore/postgres"
	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/reconcilerpb"
)

// IngestionLogRepository allows interacting with ingestion logs.
type IngestionLogRepository struct {
	queries *postgres.Queries
}

// NewIngestionLogRepository creates a new IngestionLogRepository.
func NewIngestionLogRepository(db postgres.DBTX) *IngestionLogRepository {
	return &IngestionLogRepository{
		queries: postgres.New(db),
	}
}

// CreateIngestionLog creates a new ingestion log.
func (r *IngestionLogRepository) CreateIngestionLog(_ context.Context, _ *reconcilerpb.IngestionLog, _ []byte) error {
	return errors.New("not implemented")
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
