package reconciler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

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

const (
	// DefaultBranchRefreshInterval is how often the background refresher
	// re-fetches the default branch from the NetBox plugin. Picked short
	// enough to pick up a UI-side default-branch change within ~a minute,
	// long enough to be invisible to NetBox load.
	DefaultBranchRefreshInterval = 60 * time.Second

	// DefaultBranchFetchTimeout bounds each individual refresh attempt.
	// Kept short so a stuck NetBox/Hydra doesn't extend cycle time.
	DefaultBranchFetchTimeout = 5 * time.Second
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

// Ops high level operations performed during ingestion processing.
//
// DefaultBranch lookups are served exclusively from an in-memory cache that
// a background goroutine (Start) keeps current. Hot-path callers (consume
// loop, AutoApply) never do synchronous NetBox HTTP via DefaultBranch — so
// a NetBox/Hydra outage cannot block Redis→inbox draining.
type Ops struct {
	repository            Repository
	nbClient              netboxdiodeplugin.NetBoxAPI
	logger                *slog.Logger
	limits                Limits
	bulkOperationsEnabled bool

	// Default-branch state. Updated only by the background refresher.
	branchMu     sync.RWMutex
	branch       *netboxdiodeplugin.Branch
	branchLoaded bool          // true once we've recorded ANY result (success, nil, or known-absent)
	refreshSig   chan struct{} // signal channel for on-demand refresh; buffered=1
}

// NewOps creates a new Ops. The background DefaultBranch refresher is NOT
// started until Start(ctx) is called; until then DefaultBranch() returns
// (nil, nil) and callers degrade to "no branch context".
func NewOps(repository Repository, nbClient netboxdiodeplugin.NetBoxAPI, logger *slog.Logger, limits Limits, bulkOperationsEnabled bool) *Ops {
	if limits == nil {
		limits = &DefaultLimits{}
	}

	return &Ops{
		repository:            repository,
		nbClient:              nbClient,
		logger:                logger,
		limits:                limits,
		bulkOperationsEnabled: bulkOperationsEnabled,
		refreshSig:            make(chan struct{}, 1),
	}
}

// Start launches the background default-branch refresher. It fetches once
// immediately (fire-and-forget) and then on a fixed interval, stopping when
// ctx is cancelled. Safe to call once per Ops; subsequent calls are no-ops.
func (o *Ops) Start(ctx context.Context) {
	go o.runBranchRefresher(ctx)
}

func (o *Ops) runBranchRefresher(ctx context.Context) {
	o.fetchAndStoreBranch(ctx)

	ticker := time.NewTicker(DefaultBranchRefreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			o.fetchAndStoreBranch(ctx)
		case <-o.refreshSig:
			o.fetchAndStoreBranch(ctx)
		}
	}
}

func (o *Ops) fetchAndStoreBranch(ctx context.Context) {
	fetchCtx, cancel := context.WithTimeout(ctx, DefaultBranchFetchTimeout)
	defer cancel()

	branch, err := o.nbClient.GetDefaultBranch(fetchCtx)
	if err != nil {
		if errors.Is(err, netboxdiodeplugin.ErrDefaultBranchNotFound) {
			// Older plugin version — there is no default-branch endpoint.
			// Remember nil so we don't keep retrying frivolously.
			o.setBranch(nil)
			return
		}
		// Transient failure (auth, network, NetBox down). Leave the
		// last-known value in place; consume loop and AutoApply keep
		// reading the previous successful result.
		o.logger.Warn("default branch refresh failed; serving last-known value",
			"error", err,
			"have_value", o.HasBranchLoaded(),
		)
		return
	}
	o.setBranch(branch)
}

func (o *Ops) setBranch(b *netboxdiodeplugin.Branch) {
	o.branchMu.Lock()
	defer o.branchMu.Unlock()
	o.branch = b
	o.branchLoaded = true
}

// HasBranchLoaded reports whether the refresher has successfully recorded
// any result (a real branch, nil-known-absent, or a 404-on-older-plugin).
// False means the cache is cold and DefaultBranch will return (nil, nil).
func (o *Ops) HasBranchLoaded() bool {
	o.branchMu.RLock()
	defer o.branchMu.RUnlock()
	return o.branchLoaded
}

// DefaultBranch returns the cached default branch. It never makes a network
// call — the background refresher (started via Start) owns NetBox HTTP for
// this value. Returns (nil, nil) if the cache is still cold; callers must
// tolerate that and degrade gracefully (e.g., no branch filter on lookups).
func (o *Ops) DefaultBranch(_ context.Context) (*netboxdiodeplugin.Branch, error) {
	o.branchMu.RLock()
	defer o.branchMu.RUnlock()
	return o.branch, nil
}

// RefreshDefaultBranch signals the background refresher to refetch ASAP.
// Non-blocking — returns the current cached value immediately. The actual
// refresh happens asynchronously; callers that need the freshly-fetched
// value should poll DefaultBranch after a short delay.
func (o *Ops) RefreshDefaultBranch(ctx context.Context) (*netboxdiodeplugin.Branch, error) {
	select {
	case o.refreshSig <- struct{}{}:
	default:
		// Refresh already pending; nothing to do.
	}
	return o.DefaultBranch(ctx)
}

// CreateIngestionLog creates a record for a newly received ingestion log
func (o *Ops) CreateIngestionLog(ctx context.Context, ingestionLog *reconcilerpb.IngestionLog, sourceMetadata []byte) (*ops.CreateIngestionLogResult, error) {
	// TODO: this should be in a transaction.

	fingerprinter := entityhash.NewEntityFingerprinter()
	entityHash, err := fingerprinter.GenerateEntityHash(ingestionLog.Entity)
	if err != nil {
		return nil, fmt.Errorf("failed to generate entity hash: %w", err)
	}

	// Fetch default branch from NetBox plugin (cached for 5 minutes) to ensure we search for prior ingestion logs in the correct branch context
	var defaultBranchID *string
	var branchIDForResult string
	if branch, err := o.DefaultBranch(ctx); err != nil {
		o.logger.Warn("failed to fetch default branch from NetBox plugin", "error", err)
		// Continue with nil branch (main branch) if we can't fetch default branch
	} else if branch != nil {
		branchID := fmt.Sprintf("%s (%s)", branch.Name, branch.ID)
		defaultBranchID = &branchID
		branchIDForResult = branch.ID // Store the schema_id for GenerateChangeSet
		o.logger.Debug("using default branch for ingestion log deduplication", "branch", branch.Name, "branchID", branch.ID)
	}

	existingID, existingLog, err := o.repository.FindPriorIngestionLogByEntityHash(ctx, entityHash, defaultBranchID)
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
			BranchID:     branchIDForResult,
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
		BranchID:     branchIDForResult,
	}

	return result, nil
}

// BulkCreateIngestionLogs creates multiple ingestion logs in bulk, deduplicating against existing records.
func (o *Ops) BulkCreateIngestionLogs(ctx context.Context, ingestionLogs []*reconcilerpb.IngestionLog, sourceMetadata [][]byte, entityHashes []string) ([]*ops.CreateIngestionLogResult, error) {
	// Fetch default branch (cached)
	var defaultBranchID *string
	var branchIDForResult string
	if branch, err := o.DefaultBranch(ctx); err != nil {
		o.logger.Warn("failed to fetch default branch from NetBox plugin", "error", err)
	} else if branch != nil {
		branchID := fmt.Sprintf("%s (%s)", branch.Name, branch.ID)
		defaultBranchID = &branchID
		branchIDForResult = branch.ID
		o.logger.Debug("using default branch for bulk ingestion log deduplication", "branch", branch.Name, "branchID", branch.ID)
	}

	// Collect unique hashes for the batch find
	uniqueHashes := make(map[string]struct{}, len(entityHashes))
	for _, h := range entityHashes {
		uniqueHashes[h] = struct{}{}
	}
	hashSlice := make([]string, 0, len(uniqueHashes))
	for h := range uniqueHashes {
		hashSlice = append(hashSlice, h)
	}

	// Batch find priors
	priors, err := o.repository.FindPriorIngestionLogsByEntityHashes(ctx, hashSlice, defaultBranchID)
	if err != nil {
		return nil, fmt.Errorf("failed to batch find prior ingestion logs: %w", err)
	}

	// Partition into new and duplicate
	var newLogs []*reconcilerpb.IngestionLog
	var newSourceMetadata [][]byte
	var newEntityHashes []string
	var newIndices []int // original indices for correlation

	var duplicateIDs []int32
	results := make([]*ops.CreateIngestionLogResult, len(ingestionLogs))

	for i, log := range ingestionLogs {
		hash := entityHashes[i]
		prior, found := priors[hash]
		if !found {
			newLogs = append(newLogs, log)
			var sm []byte
			if i < len(sourceMetadata) {
				sm = sourceMetadata[i]
			}
			newSourceMetadata = append(newSourceMetadata, sm)
			newEntityHashes = append(newEntityHashes, hash)
			newIndices = append(newIndices, i)
		} else {
			duplicateIDs = append(duplicateIDs, prior.ID)
			results[i] = &ops.CreateIngestionLogResult{
				ID:           prior.ID,
				IngestionLog: prior.IngestionLog,
				WasDuplicate: true,
				BranchID:     branchIDForResult,
			}
		}
	}

	// Bulk increment duplicate counts
	if len(duplicateIDs) > 0 {
		if err := o.repository.BulkIncrementDuplicateCounts(ctx, duplicateIDs); err != nil {
			return nil, fmt.Errorf("failed to bulk increment duplicate counts: %w", err)
		}
	}

	// Bulk insert new logs
	if len(newLogs) > 0 {
		idMap, err := o.repository.BulkCreateIngestionLogs(ctx, newLogs, newSourceMetadata, newEntityHashes)
		if err != nil {
			return nil, fmt.Errorf("failed to bulk create ingestion logs: %w", err)
		}

		for j, log := range newLogs {
			origIdx := newIndices[j]
			id, ok := idMap[log.Id]
			if !ok {
				return nil, fmt.Errorf("inserted ingestion log ID not found for external_id %s", log.Id)
			}
			results[origIdx] = &ops.CreateIngestionLogResult{
				ID:           id,
				IngestionLog: log,
				WasDuplicate: false,
				BranchID:     branchIDForResult,
			}
		}
	}

	return results, nil
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

// BulkGenerateChangeSets generates change sets for a batch of ingestion logs.
// When bulk operations are enabled, it uses a single BulkPlan HTTP call.
// Otherwise, it falls back to per-item differ.Diff calls.
func (o *Ops) BulkGenerateChangeSets(ctx context.Context, items []ops.QueuedIngestionLog, branchID string) []ops.BulkGenerateChangeSetResult {
	results := make([]ops.BulkGenerateChangeSetResult, len(items))

	var persistItems []ops.BulkPersistItem
	var persistIndex []int

	if o.bulkOperationsEnabled {
		persistItems, persistIndex = o.bulkPlanDiff(ctx, items, branchID, results)
	} else {
		persistItems, persistIndex = o.perItemDiff(ctx, items, branchID, results)
	}

	if len(persistItems) > 0 {
		persistResults, err := o.repository.BulkPersistChangeSets(ctx, persistItems, o.limits.MaxChangeSetsPerIngestionLog())
		if err != nil {
			o.logger.Error("bulk persist failed, falling back to per-item persist", "error", err)
			for j, idx := range persistIndex {
				item := items[idx]
				results[idx] = o.persistChangeSet(ctx, item.ID, item.IngestionLog, &persistItems[j].ChangeSet)
			}
		} else {
			for j, pr := range persistResults {
				idx := persistIndex[j]
				results[idx].ChangeSetID = pr.ChangeSetID
			}
		}
	}

	return results
}

func (o *Ops) bulkPlanDiff(ctx context.Context, items []ops.QueuedIngestionLog, branchID string, results []ops.BulkGenerateChangeSetResult) ([]ops.BulkPersistItem, []int) {
	entities := make([]netboxdiodeplugin.BulkPlanEntity, len(items))
	for i, item := range items {
		entities[i] = netboxdiodeplugin.BulkPlanEntity{
			ID:         fmt.Sprintf("%d", item.ID),
			ObjectType: item.IngestionLog.GetObjectType(),
			Entity:     item.IngestionLog.GetEntity(),
		}
	}

	bulkResp, bulkErr := o.nbClient.BulkPlan(ctx, netboxdiodeplugin.BulkPlanRequest{
		Entities: entities,
		BranchID: branchID,
	})

	if bulkErr != nil {
		for i, item := range items {
			results[i] = o.handleGenerateChangeSetFailure(ctx, item, branchID, bulkErr)
		}
		return nil, nil
	}

	resultByID := make(map[string]netboxdiodeplugin.BulkPlanResult, len(bulkResp.Results))
	for _, r := range bulkResp.Results {
		resultByID[r.ID] = r
	}

	var persistItems []ops.BulkPersistItem
	var persistIndex []int

	for i, item := range items {
		entityID := fmt.Sprintf("%d", item.ID)
		planResult, found := resultByID[entityID]
		if !found {
			results[i] = o.handleGenerateChangeSetFailure(ctx, item, branchID, fmt.Errorf("no result returned for ingestion log %d", item.ID))
			continue
		}

		cs, err := differ.ConvertBulkPlanResult(planResult, item.IngestionLog.GetObjectType())
		if err != nil {
			results[i] = o.handleGenerateChangeSetFailure(ctx, item, branchID, err)
			continue
		}

		persistItems, persistIndex = collectPersistItem(persistItems, persistIndex, results, i, item, cs)
	}

	return persistItems, persistIndex
}

func (o *Ops) perItemDiff(ctx context.Context, items []ops.QueuedIngestionLog, branchID string, results []ops.BulkGenerateChangeSetResult) ([]ops.BulkPersistItem, []int) {
	var persistItems []ops.BulkPersistItem
	var persistIndex []int

	for i, item := range items {
		ingestEntity := differ.IngestEntity{
			RequestID:  item.IngestionLog.GetRequestId(),
			ObjectType: item.IngestionLog.GetObjectType(),
			Entity:     item.IngestionLog.GetEntity(),
			State:      int(item.IngestionLog.GetState()),
		}

		cs, err := differ.Diff(ctx, ingestEntity, branchID, o.nbClient)
		if err != nil {
			results[i] = o.handleGenerateChangeSetFailure(ctx, item, branchID, err)
			continue
		}

		persistItems, persistIndex = collectPersistItem(persistItems, persistIndex, results, i, item, cs)
	}

	return persistItems, persistIndex
}

func collectPersistItem(persistItems []ops.BulkPersistItem, persistIndex []int, results []ops.BulkGenerateChangeSetResult, i int, item ops.QueuedIngestionLog, cs *changeset.ChangeSet) ([]ops.BulkPersistItem, []int) {
	stripNoopOnlyChanges(cs)

	newState := reconcilerpb.State_OPEN
	if len(cs.Changes) == 0 {
		newState = reconcilerpb.State_NO_CHANGES
	}

	persistItems = append(persistItems, ops.BulkPersistItem{
		IngestionLogID: item.ID,
		ChangeSet:      *cs,
		NewState:       newState,
	})
	persistIndex = append(persistIndex, i)
	results[i] = ops.BulkGenerateChangeSetResult{
		IngestionLogID: item.ID,
		ChangeSet:      cs,
	}

	return persistItems, persistIndex
}

func stripNoopOnlyChanges(cs *changeset.ChangeSet) {
	for _, c := range cs.Changes {
		if c.ChangeType != changeset.ChangeTypeNoop {
			return
		}
	}
	cs.Changes = nil
}

func (o *Ops) handleGenerateChangeSetFailure(ctx context.Context, item ops.QueuedIngestionLog, branchID string, err error) ops.BulkGenerateChangeSetResult {
	ingestEntity := differ.IngestEntity{
		RequestID:  item.IngestionLog.GetRequestId(),
		ObjectType: item.IngestionLog.GetObjectType(),
		Entity:     item.IngestionLog.GetEntity(),
		State:      int(item.IngestionLog.GetState()),
	}

	tags := map[string]string{"request_id": ingestEntity.RequestID}
	contextMap := map[string]any{"request_id": ingestEntity.RequestID, "object_type": ingestEntity.ObjectType}
	sentry.CaptureError(err, tags, "Ingest Entity", contextMap)

	item.IngestionLog.State = reconcilerpb.State_FAILED
	changeSetErr := handleChangeSetError(err)
	if err2 := o.repository.UpdateIngestionLogStateWithError(ctx, item.ID, reconcilerpb.State_FAILED, changeSetErr); err2 != nil {
		err = errors.Join(err, err2)
	}

	cs := differ.FailedDiffChangeSet(ingestEntity, branchID)
	id, err1 := o.repository.CreateChangeSet(ctx, *cs, item.ID)
	if err1 != nil {
		o.logger.Error("error creating failure placeholder change set", "ingestionLogID", item.ID)
		return ops.BulkGenerateChangeSetResult{IngestionLogID: item.ID, Err: errors.Join(err, err1)}
	}

	return ops.BulkGenerateChangeSetResult{IngestionLogID: item.ID, ChangeSetID: id, ChangeSet: cs, Err: err}
}

func (o *Ops) persistChangeSet(ctx context.Context, ingestionLogID int32, ingestionLog *reconcilerpb.IngestionLog, cs *changeset.ChangeSet) ops.BulkGenerateChangeSetResult {
	if len(cs.Changes) == 0 && (ingestionLog.State == reconcilerpb.State_NO_CHANGES || ingestionLog.State == reconcilerpb.State_APPLIED) {
		if err := o.repository.UpdateIngestionLogStateWithError(ctx, ingestionLogID, ingestionLog.State, nil); err != nil {
			o.logger.Error("failed to update ingestion log state (error ignored)", "ingestionLogID", ingestionLogID, "error", err)
		}
		return ops.BulkGenerateChangeSetResult{IngestionLogID: ingestionLogID, ChangeSet: cs}
	}

	changeSetID, err := o.repository.CreateChangeSet(ctx, *cs, ingestionLogID)
	if err != nil {
		return ops.BulkGenerateChangeSetResult{IngestionLogID: ingestionLogID, Err: err}
	}

	maxChangeSets := o.limits.MaxChangeSetsPerIngestionLog()
	if maxChangeSets > 0 {
		if err := o.repository.TruncateChangeSets(ctx, ingestionLogID, maxChangeSets); err != nil {
			o.logger.Error("failed to truncate change sets (error ignored)", "ingestionLogID", ingestionLogID, "error", err)
		}
	}

	if len(cs.Changes) > 0 {
		ingestionLog.State = reconcilerpb.State_OPEN
	} else {
		ingestionLog.State = reconcilerpb.State_NO_CHANGES
	}
	if err := o.repository.UpdateIngestionLogStateWithError(ctx, ingestionLogID, ingestionLog.State, nil); err != nil {
		o.logger.Error("failed to update ingestion log state (error ignored)", "ingestionLogID", ingestionLogID, "error", err)
	}

	o.logger.Debug("change set generated", "id", changeSetID, "externalID", cs.ID, "ingestionLogID", ingestionLogID)
	return ops.BulkGenerateChangeSetResult{IngestionLogID: ingestionLogID, ChangeSetID: changeSetID, ChangeSet: cs}
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

// BulkPlanApply runs combined plan + apply for a batch of QUEUED ingestion logs
// via a single /bulk-plan-apply HTTP call. Per-entity outcomes:
//   - plan failed                 -> state FAILED, no change_set persisted
//   - plan ok, no changes         -> state NO_CHANGES, no change_set persisted
//   - plan ok, apply ok           -> state APPLIED, change_set persisted
//   - plan ok, apply failed       -> state FAILED, change_set persisted (audit/retry)
//
// On a whole-batch transport error, every entity is marked FAILED with that error.
// Returns one BulkPlanApplyResult per input item, in order.
func (o *Ops) BulkPlanApply(ctx context.Context, items []ops.QueuedIngestionLog, branchID string) []ops.BulkPlanApplyResult {
	results := make([]ops.BulkPlanApplyResult, len(items))

	entities := make([]netboxdiodeplugin.BulkPlanEntity, len(items))
	for i, item := range items {
		entities[i] = netboxdiodeplugin.BulkPlanEntity{
			ID:         fmt.Sprintf("%d", item.ID),
			ObjectType: item.IngestionLog.GetObjectType(),
			Entity:     item.IngestionLog.GetEntity(),
		}
	}

	resp, err := o.nbClient.BulkPlanApply(ctx, netboxdiodeplugin.BulkPlanApplyRequest{
		Entities: entities,
		BranchID: branchID,
	})
	if err != nil {
		for i, item := range items {
			results[i] = o.persistPlanApplyFailurePlaceholder(ctx, item, branchID, err)
		}
		return results
	}

	resultByID := make(map[string]netboxdiodeplugin.BulkPlanApplyResult, len(resp.Results))
	for _, r := range resp.Results {
		resultByID[r.ID] = r
	}

	var persistItems []ops.BulkPersistItem
	var persistIndex []int

	for i, item := range items {
		entityID := fmt.Sprintf("%d", item.ID)
		planApplyResult, found := resultByID[entityID]
		if !found {
			results[i] = o.persistPlanApplyFailurePlaceholder(ctx, item, branchID, fmt.Errorf("no result returned for ingestion log %d", item.ID))
			continue
		}

		cs, planErr, applyErr := differ.ConvertBulkPlanApplyResult(planApplyResult, item.IngestionLog.GetObjectType())

		results[i].IngestionLogID = item.ID
		results[i].ChangeSet = cs
		results[i].PlanErr = planErr
		results[i].ApplyErr = applyErr

		if planErr != nil {
			// Plan failed — record state FAILED with error, and persist a
			// failure-placeholder change_set so audit/deviation-type tooling
			// has something to associate against (matches BulkGenerateChangeSets).
			results[i] = o.persistPlanApplyFailurePlaceholder(ctx, item, branchID, planErr)
			continue
		}

		// Plan succeeded — collect for bulk persist with the right terminal state.
		stripNoopOnlyChanges(cs)
		newState := reconcilerpb.State_APPLIED
		switch {
		case len(cs.Changes) == 0:
			newState = reconcilerpb.State_NO_CHANGES
		case applyErr != nil:
			newState = reconcilerpb.State_FAILED
		}

		persistItems = append(persistItems, ops.BulkPersistItem{
			IngestionLogID: item.ID,
			ChangeSet:      *cs,
			NewState:       newState,
		})
		persistIndex = append(persistIndex, i)
	}

	if len(persistItems) > 0 {
		persistResults, err := o.repository.BulkPersistChangeSets(ctx, persistItems, o.limits.MaxChangeSetsPerIngestionLog())
		if err != nil {
			o.logger.Error("bulk persist failed during bulk-plan-apply", "error", err)
			for _, idx := range persistIndex {
				if results[idx].PlanErr == nil && results[idx].ApplyErr == nil {
					results[idx].ApplyErr = err
				}
			}
		} else {
			for j, pr := range persistResults {
				idx := persistIndex[j]
				results[idx].ChangeSetID = pr.ChangeSetID
			}
		}
	}

	// For entities whose apply phase failed, attach the apply error message to the
	// ingestion log row. BulkPersistChangeSets clears the error column when it sets
	// the state, so this second pass annotates the FAILED rows with their reason.
	for _, idx := range persistIndex {
		if results[idx].ApplyErr != nil {
			changeSetErr := handleChangeSetError(results[idx].ApplyErr)
			if err := o.repository.UpdateIngestionLogStateWithError(ctx, results[idx].IngestionLogID, reconcilerpb.State_FAILED, changeSetErr); err != nil {
				o.logger.Warn("failed to annotate apply failure", "ingestionLogID", results[idx].IngestionLogID, "error", err)
			}
		}
	}

	return results
}

// persistPlanApplyFailurePlaceholder records a plan-phase failure as a
// failure-placeholder change_set + ingestion log state=FAILED with the error
// detail. Mirrors handleGenerateChangeSetFailure in the plan-only flow so
// downstream consumers (deviation type association, audit) have a change_set
// ID to attach to. Returns a fully populated BulkPlanApplyResult.
func (o *Ops) persistPlanApplyFailurePlaceholder(ctx context.Context, item ops.QueuedIngestionLog, branchID string, planErr error) ops.BulkPlanApplyResult {
	ingestEntity := differ.IngestEntity{
		RequestID:  item.IngestionLog.GetRequestId(),
		ObjectType: item.IngestionLog.GetObjectType(),
		Entity:     item.IngestionLog.GetEntity(),
		State:      int(item.IngestionLog.GetState()),
	}

	tags := map[string]string{"request_id": ingestEntity.RequestID}
	contextMap := map[string]any{"request_id": ingestEntity.RequestID, "object_type": ingestEntity.ObjectType}
	sentry.CaptureError(planErr, tags, "BulkPlanApply", contextMap)

	changeSetErr := handleChangeSetError(planErr)
	if err2 := o.repository.UpdateIngestionLogStateWithError(ctx, item.ID, reconcilerpb.State_FAILED, changeSetErr); err2 != nil {
		planErr = errors.Join(planErr, err2)
	}

	cs := differ.FailedDiffChangeSet(ingestEntity, branchID)
	id, err1 := o.repository.CreateChangeSet(ctx, *cs, item.ID)
	if err1 != nil {
		o.logger.Error("error creating failure placeholder change set", "ingestionLogID", item.ID, "error", err1)
		return ops.BulkPlanApplyResult{IngestionLogID: item.ID, PlanErr: errors.Join(planErr, err1)}
	}

	return ops.BulkPlanApplyResult{IngestionLogID: item.ID, ChangeSetID: id, ChangeSet: cs, PlanErr: planErr}
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
