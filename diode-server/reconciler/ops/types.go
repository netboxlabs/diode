package ops

import (
	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/reconcilerpb"
	"github.com/netboxlabs/diode/diode-server/reconciler/changeset"
)

// CreateIngestionLogResult represents the result of creating an ingestion log.
type CreateIngestionLogResult struct {
	ID           int32
	IngestionLog *reconcilerpb.IngestionLog
	WasDuplicate bool   // true if the ingestion log was a duplicate, in this case the prior ingestion log is returned
	BranchID     string // the branch ID used for this ingestion log (empty string means main branch)
}

// PriorIngestionLog represents a prior ingestion log found by entity hash
type PriorIngestionLog struct {
	ID           int32
	IngestionLog *reconcilerpb.IngestionLog
}

// QueuedIngestionLog represents an ingestion log in QUEUED state ready for processing
type QueuedIngestionLog struct {
	ID           int32
	IngestionLog *reconcilerpb.IngestionLog
}

// BulkGenerateChangeSetResult holds the result of generating a change set for a single item in a bulk operation.
type BulkGenerateChangeSetResult struct {
	IngestionLogID int32
	ChangeSetID    *int32
	ChangeSet      *changeset.ChangeSet
	Err            error
}

// BulkPersistItem holds data for one changeset to be bulk-persisted.
type BulkPersistItem struct {
	IngestionLogID int32
	ChangeSet      changeset.ChangeSet
	NewState       reconcilerpb.State
}

// BulkPersistResult holds the outcome of persisting one changeset.
type BulkPersistResult struct {
	IngestionLogID int32
	ChangeSetID    *int32
}

