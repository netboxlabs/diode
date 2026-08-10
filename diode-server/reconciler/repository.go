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
	TruncateChangeSets(ctx context.Context, ingestionLogID int32, limit int32) error

	// Bulk operations
	FindPriorIngestionLogsByEntityHashes(ctx context.Context, entityHashes []string, currentBranch *string) (map[string]*ops.PriorIngestionLog, error)
	BulkCreateIngestionLogs(ctx context.Context, logs []*reconcilerpb.IngestionLog, sourceMetadata [][]byte, entityHashes []string) (map[string]int32, error)
	// BulkMarkDuplicates increments duplicate bookkeeping for the given prior
	// ingestion logs by the per-ID amount and requeues drift-eligible ones
	// (APPLIED/FAILED/NO_CHANGES -> QUEUED) so they get re-planned. Returns
	// requeued flag by ingestion log ID.
	BulkMarkDuplicates(ctx context.Context, increments map[int32]int32) (map[int32]bool, error)
	// WithDedupLocks runs fn within one transaction holding per-entity-hash
	// advisory locks, serializing concurrent dedup of the same entity so the
	// find-prior -> insert window cannot race across batches or replicas. An
	// error from fn rolls the transaction back.
	WithDedupLocks(ctx context.Context, entityHashes []string, fn func(ops.DedupRepository) error) error

	// Bulk changeset persistence
	BulkPersistChangeSets(ctx context.Context, items []ops.BulkPersistItem, maxChangeSetsPerLog int32) ([]ops.BulkPersistResult, error)
	// BulkCreateDriftDeviations creates, in one transaction, a new ingestion
	// log per item (clone of the prior with a fresh external ID) carrying the
	// drift change set, and restores each prior log to APPLIED so its history
	// stays intact.
	BulkCreateDriftDeviations(ctx context.Context, items []ops.DriftDeviationItem) ([]ops.DriftDeviationResult, error)

	// Inbox processing
	ClaimQueuedIngestionLogs(ctx context.Context, batchSize int32) ([]ops.QueuedIngestionLog, error)
	ClaimQueuedForAutoApply(ctx context.Context, batchSize int32) ([]ops.QueuedIngestionLog, error)
	ResetApplyingIngestionLogs(ctx context.Context) error
}
