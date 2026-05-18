package reconciler

import (
	"context"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/reconcilerpb"
	"github.com/netboxlabs/diode/diode-server/reconciler/changeset"
	"github.com/netboxlabs/diode/diode-server/reconciler/ops"
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

	FindPriorIngestionLogByEntityHash(ctx context.Context, entityHash string, currentBranch *string) (*int32, *reconcilerpb.IngestionLog, error)
	IncrementDuplicateCount(ctx context.Context, id int32) error
	TruncateChangeSets(ctx context.Context, ingestionLogID int32, limit int32) error

	// Bulk operations
	FindPriorIngestionLogsByEntityHashes(ctx context.Context, entityHashes []string, currentBranch *string) (map[string]*ops.PriorIngestionLog, error)
	BulkCreateIngestionLogs(ctx context.Context, logs []*reconcilerpb.IngestionLog, sourceMetadata [][]byte, entityHashes []string) (map[string]int32, error)
	BulkIncrementDuplicateCounts(ctx context.Context, ids []int32) error

	// Bulk changeset persistence
	BulkPersistChangeSets(ctx context.Context, items []ops.BulkPersistItem, maxChangeSetsPerLog int32) ([]ops.BulkPersistResult, error)

	// Inbox processing
	ClaimQueuedIngestionLogs(ctx context.Context, batchSize int32) ([]ops.QueuedIngestionLog, error)
	ClaimQueuedForAutoApply(ctx context.Context, batchSize int32) ([]ops.QueuedIngestionLog, error)
	ResetApplyingIngestionLogs(ctx context.Context) error

	// Graph-upsert processing (independent of the ingestion state machine)
	ClaimGraphUpsertCandidates(ctx context.Context, batchSize, maxAttempts int32) ([]ops.QueuedIngestionLog, error)
	MarkGraphUpserted(ctx context.Context, ids []int32) error
	ReleaseGraphUpsertClaims(ctx context.Context, ids []int32) error
	ResetClaimedGraphUpserts(ctx context.Context) error
}
