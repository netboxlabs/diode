package reconciler

import (
	"context"
	"errors"
	"log/slog"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/reconcilerpb"
	"github.com/netboxlabs/diode/diode-server/netboxdiodeplugin"
	"github.com/netboxlabs/diode/diode-server/reconciler/applier"
	"github.com/netboxlabs/diode/diode-server/reconciler/changeset"
	"github.com/netboxlabs/diode/diode-server/reconciler/differ"
	"github.com/netboxlabs/diode/diode-server/sentry"
)

// GenerateChangeSet creates a change set based on current NetBox state with optional branch
func GenerateChangeSet(ctx context.Context, ingestionLogID int32, ingestionLog *reconcilerpb.IngestionLog, branchID string, nbClient netboxdiodeplugin.NetBoxAPI, repository Repository, logger *slog.Logger) (*int32, *changeset.ChangeSet, error) {
	ingestEntity := differ.IngestEntity{
		RequestID:  ingestionLog.GetRequestId(),
		ObjectType: ingestionLog.GetObjectType(),
		Entity:     ingestionLog.GetEntity(),
		State:      int(ingestionLog.GetState()),
	}

	changeSet, err := differ.Diff(ctx, ingestEntity, branchID, nbClient)
	if err != nil {
		tags := map[string]string{
			"request_id": ingestEntity.RequestID,
		}
		contextMap := map[string]any{
			"request_id":  ingestEntity.RequestID,
			"object_type": ingestEntity.ObjectType,
		}
		sentry.CaptureError(err, tags, "Ingest Entity", contextMap)
		ingestionErr := extractIngestionError(err)

		ingestionLog.State = reconcilerpb.State_FAILED
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
		ingestionLog.State = reconcilerpb.State_NO_CHANGES
		if err := repository.UpdateIngestionLogStateWithError(ctx, ingestionLogID, reconcilerpb.State_NO_CHANGES, nil); err != nil {
			logger.Warn("failed to update ingestion log state (error ignored)", "ingestionLogID", ingestionLogID, "error", err)
			// TODO(ltucker): This should be in a transaction.  Can leave an inconsistent state marked on the ingestion log.
			// return nil, err
		}
	}

	logger.Debug("change set generated", "id", changeSetID, "externalID", changeSet.ChangeSetID, "ingestionLogID", ingestionLogID)
	return changeSetID, changeSet, nil
}

// ApplyChangeSet applies change set to NetBox and updates related states
func ApplyChangeSet(ctx context.Context, ingestionLogID int32, ingestionLog *reconcilerpb.IngestionLog, changeSetID int32, changeSet *changeset.ChangeSet, nbClient netboxdiodeplugin.NetBoxAPI, repository Repository, logger *slog.Logger) error {
	if err := applier.ApplyChangeSet(ctx, logger, *changeSet, nbClient); err != nil {
		logger.Debug("failed to apply change set", "id", changeSetID, "externalID", changeSet.ChangeSetID, "ingestionLogID", ingestionLogID, "error", err)
		ingestionErr := extractIngestionError(err)

		if err2 := repository.UpdateIngestionLogStateWithError(ctx, ingestionLogID, reconcilerpb.State_FAILED, ingestionErr); err2 != nil {
			err = errors.Join(err, err2)
		}
		return err
	}

	ingestionLog.State = reconcilerpb.State_APPLIED
	if err := repository.UpdateIngestionLogStateWithError(ctx, ingestionLogID, reconcilerpb.State_APPLIED, nil); err != nil {
		logger.Warn("failed to update ingestion log state (error ignored)", "ingestionLogID", ingestionLogID, "error", err)
		// TODO(ltucker): This should be in a transaction.  Can leave an inconsistent state marked on the ingestion log.
		// return nil, err
	}

	logger.Debug("change set applied", "id", changeSetID, "externalID", changeSet.ChangeSetID, "ingestionLogID", ingestionLogID)
	return nil
}
