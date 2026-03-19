package ops

import "github.com/netboxlabs/diode/diode-server/gen/diode/v1/reconcilerpb"

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
