package reconciler

import (
	"context"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/reconcilerpb"
	"github.com/netboxlabs/diode/diode-server/reconciler/changeset"
)

// Repository is an interface for interacting with ingestion logs and change sets.
type Repository interface {
	CreateIngestionLog(ctx context.Context, ingestionLog *reconcilerpb.IngestionLog, sourceMetadata []byte, entityHash string) (*int32, error)
	UpdateIngestionLogStateWithError(ctx context.Context, id int32, state reconcilerpb.State, err error) error
	RetrieveIngestionLogByExternalID(ctx context.Context, uuid string) (*int32, *reconcilerpb.IngestionLog, error)
	RetrieveIngestionLogs(ctx context.Context, filter *reconcilerpb.RetrieveIngestionLogsRequest, limit int32, offset int32) ([]*reconcilerpb.IngestionLog, error)
	CountIngestionLogsPerState(ctx context.Context) (map[reconcilerpb.State]int32, error)
	CreateChangeSet(ctx context.Context, changeSet changeset.ChangeSet, ingestionLogID int32) (*int32, error)
	RetrieveDeviations(ctx context.Context, filter *reconcilerpb.RetrieveDeviationsRequest, limit int32, offset int32) ([]*reconcilerpb.Deviation, error)
	RetrieveDeviationByID(ctx context.Context, externalID string) (*reconcilerpb.Deviation, error)
	ListResultsByJob(ctx context.Context, filter *reconcilerpb.ListResultsByJobRequest, limit int32, offset int32) ([]*reconcilerpb.Deviation, error)

	FindPriorIngestionLogByEntityHash(ctx context.Context, entityHash string, currentBranch *string) (*int32, *reconcilerpb.IngestionLog, error)
	IncrementDuplicateCount(ctx context.Context, id int32) error
	TruncateChangeSets(ctx context.Context, ingestionLogID int32, limit int32) error
}
