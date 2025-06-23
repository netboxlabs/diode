package ops

import "github.com/netboxlabs/diode/diode-server/gen/diode/v1/reconcilerpb"

// IngestionLogRef is a reference to an ingestion log including its database id
type IngestionLogRef struct {
	ID           int32
	IngestionLog *reconcilerpb.IngestionLog
}

// CreateIngestionLogResult represents the result of creating an ingestion log.
type CreateIngestionLogResult struct {
	Created     IngestionLogRef
	DuplicateOf *IngestionLogRef
}
