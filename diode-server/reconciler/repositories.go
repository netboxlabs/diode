package reconciler

import (
	"context"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/reconcilerpb"
)

// IngestionLogRepository is an interface for interacting with ingestion logs.
type IngestionLogRepository interface {
	CreateIngestionLog(ctx context.Context, ingestionLog *reconcilerpb.IngestionLog, sourceMetadata []byte) error
}

// ChangeSetRepository is an interface for interacting with change sets.
type ChangeSetRepository interface {
	CreateChangeSet(ctx context.Context, changeSet *reconcilerpb.ChangeSet) error
}
