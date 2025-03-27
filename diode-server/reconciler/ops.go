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

// Ops high level operations performed during ingestion processing
type Ops struct {
	repository Repository
	nbClient   netboxdiodeplugin.NetBoxAPI
	logger     *slog.Logger
}

// NewOps creates a new Ops
func NewOps(repository Repository, nbClient netboxdiodeplugin.NetBoxAPI, logger *slog.Logger) *Ops {
	return &Ops{
		repository: repository,
		nbClient:   nbClient,
		logger:     logger,
	}
}

// CreateIngestionLog creates a record for a newly received ingestion log
func (o *Ops) CreateIngestionLog(ctx context.Context, ingestionLog *reconcilerpb.IngestionLog, sourceMetadata []byte) (*int32, error) {
	return o.repository.CreateIngestionLog(ctx, ingestionLog, sourceMetadata)
}

// GenerateChangeSet creates a change set based on current NetBox state with optional branch
func (o *Ops) GenerateChangeSet(ctx context.Context, ingestionLogID int32, ingestionLog *reconcilerpb.IngestionLog, branchID string) (*int32, *changeset.ChangeSet, error) {
	ingestEntity := differ.IngestEntity{
		RequestID:  ingestionLog.GetRequestId(),
		ObjectType: ingestionLog.GetObjectType(),
		Entity:     ingestionLog.GetEntity(),
		State:      int(ingestionLog.GetState()),
	}

	changeSet, err := differ.Diff(ctx, ingestEntity, branchID, o.nbClient)
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
		if err2 := o.repository.UpdateIngestionLogStateWithError(ctx, ingestionLogID, reconcilerpb.State_FAILED, ingestionErr); err2 != nil {
			err = errors.Join(err, err2)
		}

		cs := differ.FailedDiffChangeSet(ingestEntity, branchID)
		id, err1 := o.repository.CreateChangeSet(ctx, *cs, ingestionLogID)
		if err1 != nil {
			o.logger.Error("error generating diff failure placeholder change set")
			return nil, nil, errors.Join(err, err1)
		}

		return id, cs, err
	}

	changeSetID, err := o.repository.CreateChangeSet(ctx, *changeSet, ingestionLogID)
	if err != nil {
		return nil, nil, err
	}

	state := reconcilerpb.State_OPEN
	if len(changeSet.Changes) == 0 {
		state = reconcilerpb.State_NO_CHANGES
	}

	ingestionLog.State = state

	if err := o.repository.UpdateIngestionLogStateWithError(ctx, ingestionLogID, state, nil); err != nil {
		o.logger.Warn("failed to update ingestion log state (error ignored)", "ingestionLogID", ingestionLogID, "error", err)
		// TODO(ltucker): This should be in a transaction.  Can leave an inconsistent state marked on the ingestion log.
		// return nil, err
	}

	o.logger.Debug("change set generated", "id", changeSetID, "externalID", changeSet.ID, "ingestionLogID", ingestionLogID)
	return changeSetID, changeSet, nil
}

// ApplyChangeSet applies change set to NetBox and updates related states
func (o *Ops) ApplyChangeSet(ctx context.Context, ingestionLogID int32, ingestionLog *reconcilerpb.IngestionLog, changeSetID int32, changeSet *changeset.ChangeSet) error {
	if err := applier.ApplyChangeSet(ctx, o.logger, *changeSet, o.nbClient); err != nil {
		o.logger.Debug("failed to apply change set", "id", changeSetID, "externalID", changeSet.ID, "ingestionLogID", ingestionLogID, "error", err)
		ingestionErr := extractIngestionError(err)

		if err2 := o.repository.UpdateIngestionLogStateWithError(ctx, ingestionLogID, reconcilerpb.State_FAILED, ingestionErr); err2 != nil {
			err = errors.Join(err, err2)
		}
		return err
	}

	ingestionLog.State = reconcilerpb.State_APPLIED
	if err := o.repository.UpdateIngestionLogStateWithError(ctx, ingestionLogID, reconcilerpb.State_APPLIED, nil); err != nil {
		o.logger.Warn("failed to update ingestion log state (error ignored)", "ingestionLogID", ingestionLogID, "error", err)
		// TODO(ltucker): This should be in a transaction.  Can leave an inconsistent state marked on the ingestion log.
		// return nil, err
	}

	o.logger.Debug("change set applied", "id", changeSetID, "externalID", changeSet.ID, "ingestionLogID", ingestionLogID)
	return nil
}
