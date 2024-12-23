package reconciler

import (
	"context"
	"errors"
	"log/slog"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/reconcilerpb"
	"github.com/netboxlabs/diode/diode-server/netboxdiodeplugin"
	"github.com/netboxlabs/diode/diode-server/reconciler/changeset"
	"github.com/netboxlabs/diode/diode-server/reconciler/differ"
	"github.com/netboxlabs/diode/diode-server/sentry"
)

func generateChangeSet(ctx context.Context, ingestionLogID int32, ingestionLog *reconcilerpb.IngestionLog, branchID string, nbClient netboxdiodeplugin.NetBoxAPI, repository Repository, logger *slog.Logger) (*int32, *changeset.ChangeSet, error) {
	ingestEntity := differ.IngestEntity{
		RequestID: ingestionLog.GetRequestId(),
		DataType:  ingestionLog.GetDataType(),
		Entity:    ingestionLog.GetEntity(),
		State:     int(ingestionLog.GetState()),
	}

	changeSet, err := differ.Diff(ctx, ingestEntity, branchID, nbClient)
	if err != nil {
		tags := map[string]string{
			"request_id": ingestEntity.RequestID,
		}
		contextMap := map[string]any{
			"request_id": ingestEntity.RequestID,
			"data_type":  ingestEntity.DataType,
		}
		sentry.CaptureError(err, tags, "Ingest Entity", contextMap)
		ingestionErr := extractIngestionError(err)

		if err2 := repository.UpdateIngestionLogStateWithError(ctx, ingestionLogID, reconcilerpb.State_FAILED, ingestionErr); err2 != nil {
			err = errors.Join(err, err2)
		}
		return nil, nil, err
	}

	changeSetID, err := repository.CreateChangeSet(ctx, *changeSet, ingestionLogID)
	if err != nil {
		return nil, nil, err
	}

	if len(changeSet.ChangeSet) == 0 {
		if err := repository.UpdateIngestionLogStateWithError(ctx, ingestionLogID, reconcilerpb.State_NO_CHANGES, nil); err != nil {
			logger.Warn("failed to update ingestion log state (ignored)", "ingestionLogID", ingestionLog.GetId(), "error", err)
			// TODO(ltucker): This should be in a transaction.  Can leave an inconsistent state marked on the ingestion log.
			// return nil, err
		}
	}

	logger.Debug("change set generated", "id", changeSetID, "externalID", changeSet.ChangeSetID, "ingestionLogID", ingestionLog.GetId())
	return changeSetID, changeSet, nil
}
