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
	Requeued     bool   // true if a duplicate's prior ingestion log was requeued to re-check NetBox state drift
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
	// RequeuedFromState is the terminal state the log was in when a duplicate
	// observation requeued it for re-plan (STATE_UNSPECIFIED when the log is a
	// first-time plan). A log requeued from APPLIED spawns a new deviation on
	// drift and is restored to APPLIED, preserving its history.
	RequeuedFromState reconcilerpb.State
}

// DriftDeviationItem describes a new deviation to create for an entity whose
// NetBox state drifted since its prior ingestion log was applied.
type DriftDeviationItem struct {
	PriorIngestionLogID int32
	NewExternalID       string
	NewState            reconcilerpb.State
	ChangeSet           changeset.ChangeSet
}

// DriftDeviationResult holds the outcome of creating one drift deviation.
type DriftDeviationResult struct {
	PriorIngestionLogID int32
	NewIngestionLogID   int32
	ChangeSetID         *int32
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

// BulkPlanApplyResult holds the outcome of one entity in a /bulk-plan-apply call.
// ChangeSet is populated when the plan phase produced one (regardless of apply outcome).
// PlanErr and ApplyErr are split so callers can attribute the failure phase; both nil
// means the entity was planned and applied successfully.
type BulkPlanApplyResult struct {
	IngestionLogID int32
	ChangeSetID    *int32
	ChangeSet      *changeset.ChangeSet
	PlanErr        error
	ApplyErr       error
}
