package reconciler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/netboxlabs/diode/diode-server/entityhash"
	diodeErrors "github.com/netboxlabs/diode/diode-server/errors"
	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/reconcilerpb"
	"github.com/netboxlabs/diode/diode-server/netboxdiodeplugin"
	"github.com/netboxlabs/diode/diode-server/reconciler/applier"
	"github.com/netboxlabs/diode/diode-server/reconciler/changeset"
	"github.com/netboxlabs/diode/diode-server/reconciler/differ"
	"github.com/netboxlabs/diode/diode-server/reconciler/ops"
	"github.com/netboxlabs/diode/diode-server/sentry"
)

// Limits is an interface that provides limits for the reconciler operations to enforce
type Limits interface {
	MaxChangeSetsPerIngestionLog() int32
}

// DefaultLimits is the default implementation of the Limits interface
type DefaultLimits struct{}

// MaxChangeSetsPerIngestionLog returns retention limit for change sets per ingestion log
func (l *DefaultLimits) MaxChangeSetsPerIngestionLog() int32 {
	return 5
}

// Ops high level operations performed during ingestion processing
type Ops struct {
	repository Repository
	nbClient   netboxdiodeplugin.NetBoxAPI
	logger     *slog.Logger
	limits     Limits
}

// NewOps creates a new Ops
func NewOps(repository Repository, nbClient netboxdiodeplugin.NetBoxAPI, logger *slog.Logger, limits Limits) *Ops {
	if limits == nil {
		limits = &DefaultLimits{}
	}

	return &Ops{
		repository: repository,
		nbClient:   nbClient,
		logger:     logger,
		limits:     limits,
	}
}

// CreateIngestionLog creates a record for a newly received ingestion log
func (o *Ops) CreateIngestionLog(ctx context.Context, ingestionLog *reconcilerpb.IngestionLog, sourceMetadata []byte) (*ops.CreateIngestionLogResult, error) {
	// TODO: this should be in a transaction.

	fingerprinter := entityhash.NewEntityFingerprinter()
	entityHash, err := fingerprinter.GenerateEntityHash(ingestionLog.Entity)
	if err != nil {
		return nil, fmt.Errorf("failed to generate entity hash: %w", err)
	}

	existingID, existingLog, err := o.repository.FindPriorIngestionLogByEntityHash(ctx, entityHash, nil)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed to search for prior deviation: %w", err)
	}

	if existingID == nil {
		id, err := o.repository.CreateIngestionLog(ctx, ingestionLog, sourceMetadata, entityHash)
		if err != nil {
			return nil, err
		}

		result := &ops.CreateIngestionLogResult{
			ID:           *id,
			IngestionLog: ingestionLog,
		}
		return result, nil
	}

	// It was a duplicate, increment the duplicate count and return the prior ingestion log
	if err := o.repository.IncrementDuplicateCount(ctx, *existingID); err != nil {
		return nil, fmt.Errorf("failed to mark record as duplicate: %w", err)
	}

	result := &ops.CreateIngestionLogResult{
		ID:           *existingID,
		IngestionLog: existingLog,
		WasDuplicate: true,
	}

	return result, nil
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

		ingestionLog.State = reconcilerpb.State_FAILED

		changeSetErr := handleChangeSetError(err)
		if err2 := o.repository.UpdateIngestionLogStateWithError(ctx, ingestionLogID, reconcilerpb.State_FAILED, changeSetErr); err2 != nil {
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

	// if the change set has no changes and the ingestion log is already in the no changes or applied state,
	// we don't record another changeset in the database, we just bump the updated at time.
	if len(changeSet.Changes) == 0 && (ingestionLog.State == reconcilerpb.State_NO_CHANGES || ingestionLog.State == reconcilerpb.State_APPLIED) {
		if err := o.repository.UpdateIngestionLogStateWithError(ctx, ingestionLogID, ingestionLog.State, nil); err != nil {
			o.logger.Error("failed to update ingestion log state (error ignored)", "ingestionLogID", ingestionLogID, "error", err)
		}
		return nil, changeSet, nil
	}

	// TODO: At this point if the prior ingestion log is in the applied state, and we have
	// a new set of changes, we could "clone" the ingestion log to open a new "deviation"
	// and leave the prior one as applied. This would be more historically accurate /
	// less surprising.  For now we just re-open the previously applied change set.
	//
	// If we did create a new one, we would need to communicate that back to the rest
	// of the pipeline and also this operation's name would be a bit of misnomer.
	// Possibly some refactoring/renaming of the operations (which are meant to
	// keep rpc and pipeline behavior in sync) would be warranted.

	changeSetID, err := o.repository.CreateChangeSet(ctx, *changeSet, ingestionLogID)
	if err != nil {
		return nil, nil, err
	}

	maxChangeSets := o.limits.MaxChangeSetsPerIngestionLog()
	if maxChangeSets > 0 {
		if err := o.repository.TruncateChangeSets(ctx, ingestionLogID, maxChangeSets); err != nil {
			o.logger.Error("failed to truncate change sets (error ignored)", "ingestionLogID", ingestionLogID, "error", err)
		}
	}

	if len(changeSet.Changes) > 0 {
		ingestionLog.State = reconcilerpb.State_OPEN
	} else {
		ingestionLog.State = reconcilerpb.State_NO_CHANGES
	}
	if err := o.repository.UpdateIngestionLogStateWithError(ctx, ingestionLogID, ingestionLog.State, nil); err != nil {
		o.logger.Error("failed to update ingestion log state (error ignored)", "ingestionLogID", ingestionLogID, "error", err)
	}

	o.logger.Debug("change set generated", "id", changeSetID, "externalID", changeSet.ID, "ingestionLogID", ingestionLogID)
	return changeSetID, changeSet, nil
}

// ApplyChangeSet applies change set to NetBox and updates related states
func (o *Ops) ApplyChangeSet(ctx context.Context, ingestionLogID int32, ingestionLog *reconcilerpb.IngestionLog, changeSetID int32, changeSet *changeset.ChangeSet) error {
	if err := applier.ApplyChangeSet(ctx, o.logger, *changeSet, o.nbClient); err != nil {
		o.logger.Debug("failed to apply change set", "id", changeSetID, "externalID", changeSet.ID, "ingestionLogID", ingestionLogID, "error", err)

		changeSetErr := handleChangeSetError(err)

		if err2 := o.repository.UpdateIngestionLogStateWithError(ctx, ingestionLogID, reconcilerpb.State_FAILED, changeSetErr); err2 != nil {
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

func handleChangeSetError(err error) error {
	var changeSetErr *changeset.Error
	if errors.As(err, &changeSetErr) {
		return err
	}

	return &changeset.Error{
		Message: err.Error(),
		Code:    diodeErrors.ErrCodeInternal,
		Details: []byte{},
	}
}
