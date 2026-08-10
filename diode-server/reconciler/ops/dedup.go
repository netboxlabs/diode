package ops

import (
	"context"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/reconcilerpb"
)

// DedupRepository is the repository surface available inside a dedup
// transaction (see reconciler.Repository.WithDedupLocks): finding prior
// ingestion logs by entity hash, inserting new ones, and incrementing
// duplicate bookkeeping. All calls run on the same transaction, which holds
// per-entity-hash advisory locks so concurrent same-entity batches serialize
// instead of racing the find-prior -> insert window.
type DedupRepository interface {
	FindPriorIngestionLogByEntityHash(ctx context.Context, entityHash string, currentBranch *string) (*int32, *reconcilerpb.IngestionLog, error)
	FindPriorIngestionLogsByEntityHashes(ctx context.Context, entityHashes []string, currentBranch *string) (map[string]*PriorIngestionLog, error)
	CreateIngestionLog(ctx context.Context, ingestionLog *reconcilerpb.IngestionLog, sourceMetadata []byte, entityHash string) (*int32, error)
	BulkCreateIngestionLogs(ctx context.Context, logs []*reconcilerpb.IngestionLog, sourceMetadata [][]byte, entityHashes []string) (map[string]int32, error)
	// BulkMarkDuplicates increments duplicate bookkeeping by the per-ID amount
	// and requeues drift-eligible rows (APPLIED/FAILED/NO_CHANGES -> QUEUED).
	// Returns requeued flag by ingestion log ID.
	BulkMarkDuplicates(ctx context.Context, increments map[int32]int32) (map[int32]bool, error)
}
