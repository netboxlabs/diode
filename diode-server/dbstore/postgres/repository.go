package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/netboxlabs/diode/diode-server/gen/dbstore/postgres"
	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/reconcilerpb"
	"github.com/netboxlabs/diode/diode-server/reconciler/changeset"
	"github.com/netboxlabs/diode/diode-server/reconciler/ops"
)

// rollbackTimeout bounds how long we wait for a rollback during shutdown.
// The parent context may already be canceled, so rollback uses a detached
// context with this timeout to ensure cleanup completes.
const rollbackTimeout = 5 * time.Second

// Repository is an interface for interacting with ingestion logs and change sets.
type Repository struct {
	pool    *pgxpool.Pool
	queries *postgres.Queries
}

// NewRepository creates a new Repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool:    pool,
		queries: postgres.New(pool),
	}
}

// CreateIngestionLog creates a new ingestion log with entity hash and deduplication fields.
func (r *Repository) CreateIngestionLog(ctx context.Context, ingestionLog *reconcilerpb.IngestionLog, sourceMetadata []byte, entityHash string) (*int32, error) {
	marshaler := protojson.MarshalOptions{
		UseProtoNames: true,
	}
	entityJSON, err := marshaler.Marshal(ingestionLog.Entity)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal entity: %w", err)
	}

	params := postgres.CreateIngestionLogParams{
		ExternalID:         ingestionLog.Id,
		ObjectType:         pgtype.Text{String: ingestionLog.ObjectType, Valid: true},
		State:              pgtype.Int4{Int32: int32(ingestionLog.State), Valid: true},
		RequestID:          pgtype.Text{String: ingestionLog.RequestId, Valid: true},
		IngestionTs:        pgtype.Int8{Int64: ingestionLog.IngestionTs, Valid: true},
		SourceTs:           pgtype.Int8{Int64: ingestionLog.SourceTs, Valid: true},
		ProducerAppName:    pgtype.Text{String: ingestionLog.ProducerAppName, Valid: true},
		ProducerAppVersion: pgtype.Text{String: ingestionLog.ProducerAppVersion, Valid: true},
		SdkName:            pgtype.Text{String: ingestionLog.SdkName, Valid: true},
		SdkVersion:         pgtype.Text{String: ingestionLog.SdkVersion, Valid: true},
		Entity:             entityJSON,
		SourceMetadata:     sourceMetadata,
		EntityHash:         pgtype.Text{String: entityHash, Valid: true},
	}

	createdIngestionLog, err := r.queries.CreateIngestionLog(ctx, params)
	if err != nil {
		return nil, err
	}
	return &createdIngestionLog.ID, nil
}

// RetrieveIngestionLogByExternalID looks up an ingestion log using its external identifier (uuid)
func (r *Repository) RetrieveIngestionLogByExternalID(ctx context.Context, uuid string) (*int32, *reconcilerpb.IngestionLog, error) {
	ingestionLog, err := r.queries.RetrieveIngestionLogByExternalID(ctx, uuid)
	if err != nil {
		return nil, nil, err
	}
	log, err := ingestionLog.ToProto()
	if err != nil {
		return nil, nil, err
	}

	return &ingestionLog.ID, log, nil
}

// UpdateIngestionLogStateWithError updates an ingestion log with a new state and error.
func (r *Repository) UpdateIngestionLogStateWithError(ctx context.Context, id int32, state reconcilerpb.State, err error) error {
	params := postgres.UpdateIngestionLogStateWithErrorParams{
		ID:    id,
		State: pgtype.Int4{Int32: int32(state), Valid: true},
	}

	if err != nil {
		errJSON, err := json.Marshal(err)
		if err != nil {
			return fmt.Errorf("failed to marshal error: %w", err)
		}
		params.Error = errJSON
	}
	return r.queries.UpdateIngestionLogStateWithError(ctx, params)
}

// CountIngestionLogsPerState counts ingestion logs per state.
func (r *Repository) CountIngestionLogsPerState(ctx context.Context) (map[reconcilerpb.State]int32, error) {
	counts, err := r.queries.CountIngestionLogsPerState(ctx)
	if err != nil {
		return nil, err
	}

	stateCounts := make(map[reconcilerpb.State]int32)
	for _, stateCount := range counts {
		stateCounts[reconcilerpb.State(stateCount.State.Int32)] = int32(stateCount.Count)
	}
	return stateCounts, nil
}

// RetrieveIngestionLogs retrieves ingestion logs.
func (r *Repository) RetrieveIngestionLogs(ctx context.Context, filter *reconcilerpb.RetrieveIngestionLogsRequest, limit int32, offset int32) ([]*reconcilerpb.IngestionLog, error) {
	params := postgres.RetrieveIngestionLogsWithChangeSetsParams{
		Limit:  limit,
		Offset: offset,
	}
	if filter.State != nil {
		params.State = pgtype.Int4{Int32: int32(*filter.State), Valid: true}
	}

	// backwards compatibility (dataType -> objectType)
	if filter.DataType != "" {
		params.ObjectType = pgtype.Text{String: filter.DataType, Valid: true}
	}
	if filter.ObjectType != "" {
		params.ObjectType = pgtype.Text{String: filter.ObjectType, Valid: true}
	}
	if filter.IngestionTsStart > 0 {
		params.IngestionTsStart = pgtype.Int8{Int64: filter.IngestionTsStart, Valid: true}
	}
	if filter.IngestionTsEnd > 0 {
		params.IngestionTsEnd = pgtype.Int8{Int64: filter.IngestionTsEnd, Valid: true}
	}

	rawIngestionLogs, err := r.queries.RetrieveIngestionLogsWithChangeSets(ctx, params)
	if err != nil {
		return nil, err
	}

	ingestionLogs := make([]*reconcilerpb.IngestionLog, 0, len(rawIngestionLogs))
	for _, row := range rawIngestionLogs {
		ingestionLog := row
		entity := &diodepb.Entity{}
		if err := protojson.Unmarshal(ingestionLog.Entity, entity); err != nil {
			return nil, fmt.Errorf("failed to unmarshal entity: %w", err)
		}
		var ingestionErr reconcilerpb.DeviationError
		if ingestionLog.Error != nil {
			if err := postgres.LoadDeviationError(ingestionLog.Error, &ingestionErr); err != nil {
				return nil, fmt.Errorf("failed to unmarshal error: %w", err)
			}
		}

		log := &reconcilerpb.IngestionLog{
			Id:                 ingestionLog.ExternalID,
			ObjectType:         ingestionLog.ObjectType.String,
			State:              reconcilerpb.State(ingestionLog.State.Int32),
			RequestId:          ingestionLog.RequestID.String,
			IngestionTs:        ingestionLog.IngestionTs.Int64,
			SourceTs:           ingestionLog.SourceTs.Int64,
			ProducerAppName:    ingestionLog.ProducerAppName.String,
			ProducerAppVersion: ingestionLog.ProducerAppVersion.String,
			SdkName:            ingestionLog.SdkName.String,
			SdkVersion:         ingestionLog.SdkVersion.String,
			Entity:             entity,
			Error:              &ingestionErr,
		}

		if row.Changes != nil {
			var dbChanges []postgres.Change
			if err := json.Unmarshal(row.Changes, &dbChanges); err != nil {
				return nil, fmt.Errorf("failed to unmarshal changes: %w", err)
			}

			changes := make([]changeset.Change, 0, len(dbChanges))
			for _, dbChange := range dbChanges {
				change := changeset.Change{
					ID:         dbChange.ExternalID,
					ChangeType: dbChange.ChangeType,
					ObjectType: dbChange.ObjectType,
					Before:     dbChange.Before,
					After:      dbChange.After,
					NewRefs:    dbChange.NewRefs,
				}

				objID := int(dbChange.ObjectID.Int32)
				if dbChange.ObjectID.Valid {
					change.ObjectID = &objID
				}
				objVersion := int(dbChange.ObjectVersion.Int32)
				if dbChange.ObjectVersion.Valid {
					change.ObjectVersion = &objVersion
				}
				if dbChange.RefID.Valid {
					change.RefID = &dbChange.RefID.String
				}

				changes = append(changes, change)
			}

			var branchID *string
			if row.ChangeSet.BranchID.Valid {
				branchID = &row.ChangeSet.BranchID.String
			}

			var deviationName *string
			if row.ChangeSet.DeviationName.Valid {
				deviationName = &row.ChangeSet.DeviationName.String
			}

			changeSet := &changeset.ChangeSet{
				ID:            row.ChangeSet.ExternalID,
				Changes:       changes,
				BranchID:      branchID,
				DeviationName: deviationName,
			}

			var compressedChangeSet []byte
			if len(changes) > 0 {
				b, err := changeset.CompressChangeSet(changeSet)
				if err != nil {
					return nil, fmt.Errorf("failed to compress change set: %w", err)
				}
				compressedChangeSet = b
			}

			log.ChangeSet = &reconcilerpb.ChangeSet{
				Id:            row.ChangeSet.ExternalID,
				Data:          compressedChangeSet,
				BranchId:      branchID,
				DeviationName: deviationName,
			}
		}

		ingestionLogs = append(ingestionLogs, log)
	}

	return ingestionLogs, nil
}

// CreateChangeSet creates a new change set.
func (r *Repository) CreateChangeSet(ctx context.Context, changeSet changeset.ChangeSet, ingestionLogID int32) (*int32, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}

	rollback := func() {
		// Detached context so the rollback can complete even when the
		// caller's context is already canceled (pod shutdown, parent
		// timeout). pgx5 otherwise fails to DEALLOCATE prepared statements
		// and returns an error here. Panicking on that would crash the
		// reconciler over what is at worst a logged warning - the
		// transaction is cleaned up when the connection returns to the
		// pool either way.
		rollbackCtx, cancel := context.WithTimeout(context.Background(), rollbackTimeout)
		defer cancel()
		if err := tx.Rollback(rollbackCtx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			slog.WarnContext(ctx, "failed to rollback transaction", "error", err)
		}
	}

	qtx := r.queries.WithTx(tx)
	params := postgres.CreateChangeSetParams{
		ExternalID:     changeSet.ID,
		IngestionLogID: ingestionLogID,
	}
	if changeSet.BranchID != nil {
		params.BranchID = pgtype.Text{String: *changeSet.BranchID, Valid: true}
	}
	if changeSet.DeviationName != nil {
		params.DeviationName = pgtype.Text{String: *changeSet.DeviationName, Valid: true}
	}

	cs, err := qtx.CreateChangeSet(ctx, params)
	if err != nil {
		rollback()
		return nil, fmt.Errorf("failed to create change set: %w", err)
	}

	bulkParams := make([]postgres.BulkCreateChangesParams, 0, len(changeSet.Changes))
	for i, change := range changeSet.Changes {
		beforeJSON, err := json.Marshal(change.Before)
		if err != nil {
			rollback()
			return nil, fmt.Errorf("failed to marshal before state: %w", err)
		}

		afterJSON, err := json.Marshal(change.After)
		if err != nil {
			rollback()
			return nil, fmt.Errorf("failed to marshal after state: %w", err)
		}

		p := postgres.BulkCreateChangesParams{
			ExternalID:         change.ID,
			ChangeSetID:        cs.ID,
			ChangeType:         change.ChangeType,
			ObjectType:         change.ObjectType,
			ObjectPrimaryValue: change.ObjectPrimaryValue,
			Before:             beforeJSON,
			After:              afterJSON,
			NewRefs:            change.NewRefs,
			SequenceNumber:     pgtype.Int4{Int32: int32(i), Valid: true},
		}
		if change.ObjectID != nil {
			p.ObjectID = pgtype.Int4{Int32: int32(*change.ObjectID), Valid: true}
		}
		if change.ObjectVersion != nil {
			p.ObjectVersion = pgtype.Int4{Int32: int32(*change.ObjectVersion), Valid: true}
		}
		if change.RefID != nil {
			p.RefID = pgtype.Text{String: *change.RefID, Valid: true}
		}
		bulkParams = append(bulkParams, p)
	}

	if _, err = qtx.BulkCreateChanges(ctx, bulkParams); err != nil {
		rollback()
		return nil, fmt.Errorf("failed to bulk create changes: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		rollback()
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return &cs.ID, nil
}

// RetrieveDeviations retrieves deviations.
func (r *Repository) RetrieveDeviations(ctx context.Context, filter *reconcilerpb.RetrieveDeviationsRequest, limit int32, offset int32) ([]*reconcilerpb.Deviation, error) {
	params := postgres.RetrieveDeviationsParams{
		Limit:  limit,
		Offset: offset,
	}

	if len(filter.State) > 0 {
		states := make([]int32, 0, len(filter.State))
		for _, state := range filter.State {
			states = append(states, int32(state))
		}
		params.State = states
	}
	if len(filter.ObjectType) > 0 {
		params.ObjectType = filter.ObjectType
	}
	if len(filter.BranchId) > 0 {
		params.BranchID = filter.BranchId
	}
	if filter.IngestionTsStart > 0 {
		params.IngestionTsStart = pgtype.Int8{Int64: filter.IngestionTsStart, Valid: true}
	}
	if filter.IngestionTsEnd > 0 {
		params.IngestionTsEnd = pgtype.Int8{Int64: filter.IngestionTsEnd, Valid: true}
	}

	rawDeviations, err := r.queries.RetrieveDeviations(ctx, params)
	if err != nil {
		return nil, err
	}

	deviations := make([]*reconcilerpb.Deviation, 0, len(rawDeviations))
	for _, rawDeviation := range rawDeviations {
		deviationPb, err := deviationToProto(rawDeviation)
		if err != nil {
			return nil, fmt.Errorf("failed to convert deviation to proto: %w", err)
		}

		deviations = append(deviations, deviationPb)
	}

	return deviations, nil
}

// RetrieveDeviationByID retrieves a deviation by its external identifier.
func (r *Repository) RetrieveDeviationByID(ctx context.Context, externalID string) (*reconcilerpb.Deviation, error) {
	rawDeviation, err := r.queries.RetrieveDeviationByID(ctx, externalID)
	if err != nil {
		return nil, err
	}

	deviationPb, err := deviationToProto(rawDeviation)
	if err != nil {
		return nil, fmt.Errorf("failed to convert deviation to proto: %w", err)
	}

	return deviationPb, nil
}

func deviationToProto(dbDeviation postgres.VDeviation) (*reconcilerpb.Deviation, error) {
	entity := &diodepb.Entity{}
	if err := protojson.Unmarshal(dbDeviation.Entity, entity); err != nil {
		return nil, fmt.Errorf("failed to unmarshal entity: %w", err)
	}

	source := dbDeviation.ProducerAppName.String

	var deviationName string
	var branchID *string
	if dbDeviation.ChangeSet != nil {
		deviationName = dbDeviation.ChangeSet.DeviationName.String
		if dbDeviation.ChangeSet.BranchID.Valid {
			branchID = &dbDeviation.ChangeSet.BranchID.String
		}
	}

	var deviationErr *reconcilerpb.DeviationError
	if dbDeviation.Error != nil {
		if err := postgres.LoadDeviationError(dbDeviation.Error, deviationErr); err != nil {
			return nil, fmt.Errorf("failed to load deviation error: %w", err)
		}
	}

	deviation := &reconcilerpb.Deviation{
		Id:             dbDeviation.ExternalID,
		Name:           deviationName,
		Source:         source,
		State:          reconcilerpb.State(dbDeviation.State.Int32),
		ObjectType:     dbDeviation.ObjectType.String,
		BranchId:       branchID,
		IngestedEntity: entity,
		Error:          deviationErr,
		IngestionTs:    dbDeviation.IngestionTs.Int64,
		LastUpdateTs:   dbDeviation.UpdatedAt.Time.UnixNano(),
		SourceTs:       dbDeviation.SourceTs.Int64,
	}

	if dbDeviation.Changes != nil {
		var dbChanges []postgres.Change
		if err := json.Unmarshal(dbDeviation.Changes, &dbChanges); err != nil {
			return nil, fmt.Errorf("failed to unmarshal changes: %w", err)
		}

		changes := make([]*reconcilerpb.Change, 0, len(dbChanges))
		for _, dbChange := range dbChanges {
			change := &reconcilerpb.Change{
				Id:                 dbChange.ExternalID,
				ChangeType:         dbChange.ChangeType,
				ObjectType:         dbChange.ObjectType,
				ObjectPrimaryValue: dbChange.ObjectPrimaryValue,
				Before:             dbChange.Before,
				After:              dbChange.After,
			}
			changes = append(changes, change)
		}

		deviation.Changes = changes
	}

	return deviation, nil
}

// FindPriorIngestionLogByEntityHash finds a prior deviation with the same entity hash, considering branch context.
func (r *Repository) FindPriorIngestionLogByEntityHash(ctx context.Context, entityHash string, currentBranch *string) (*int32, *reconcilerpb.IngestionLog, error) {
	params := postgres.FindPriorIngestionLogByEntityHashParams{
		EntityHash: pgtype.Text{String: entityHash, Valid: true},
	}
	if currentBranch != nil {
		params.BranchID = pgtype.Text{String: *currentBranch, Valid: true}
	}

	dbLog, err := r.queries.FindPriorIngestionLogByEntityHash(ctx, params)
	if err != nil {
		return nil, nil, err
	}

	log, err := dbLog.ToProto()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to convert to proto: %w", err)
	}

	return &dbLog.ID, log, nil
}

// IncrementDuplicateCount increments the duplicate count for an ingestion log
func (r *Repository) IncrementDuplicateCount(ctx context.Context, id int32) error {
	return r.queries.IncrementDuplicateCount(ctx, id)
}

// TruncateChangeSets truncates change sets for an ingestion log to the given limit (keeps latest n)
func (r *Repository) TruncateChangeSets(ctx context.Context, ingestionLogID int32, limit int32) error {
	return r.queries.TruncateChangeSets(ctx, postgres.TruncateChangeSetsParams{
		IngestionLogID: ingestionLogID,
		Limit:          limit,
	})
}

// FindPriorIngestionLogsByEntityHashes finds prior ingestion logs matching the given entity hashes, scoped by branch.
func (r *Repository) FindPriorIngestionLogsByEntityHashes(ctx context.Context, entityHashes []string, currentBranch *string) (map[string]*ops.PriorIngestionLog, error) {
	params := postgres.FindPriorIngestionLogsByEntityHashesParams{
		EntityHashes: entityHashes,
	}
	if currentBranch != nil {
		params.BranchID = pgtype.Text{String: *currentBranch, Valid: true}
	}

	dbLogs, err := r.queries.FindPriorIngestionLogsByEntityHashes(ctx, params)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*ops.PriorIngestionLog, len(dbLogs))
	for _, dbLog := range dbLogs {
		log, err := dbLog.ToProto()
		if err != nil {
			return nil, fmt.Errorf("failed to convert to proto: %w", err)
		}
		result[dbLog.EntityHash.String] = &ops.PriorIngestionLog{
			ID:           dbLog.ID,
			IngestionLog: log,
		}
	}
	return result, nil
}

// BulkCreateIngestionLogs bulk inserts ingestion logs using the COPY protocol.
// It pre-allocates sequence IDs and returns a map of external_id → id.
func (r *Repository) BulkCreateIngestionLogs(ctx context.Context, logs []*reconcilerpb.IngestionLog, sourceMetadata [][]byte, entityHashes []string) (map[string]int32, error) {
	ids, err := r.allocateIngestionLogIDs(ctx, len(logs))
	if err != nil {
		return nil, fmt.Errorf("failed to allocate IDs: %w", err)
	}

	marshaler := protojson.MarshalOptions{
		UseProtoNames: true,
	}

	params := make([]postgres.BulkCreateIngestionLogsParams, 0, len(logs))
	idMap := make(map[string]int32, len(logs))
	for i, log := range logs {
		entityJSON, err := marshaler.Marshal(log.Entity)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal entity at index %d: %w", i, err)
		}

		var sm []byte
		if i < len(sourceMetadata) {
			sm = sourceMetadata[i]
		}

		params = append(params, postgres.BulkCreateIngestionLogsParams{
			ID:                 ids[i],
			ExternalID:         log.Id,
			ObjectType:         pgtype.Text{String: log.ObjectType, Valid: true},
			State:              pgtype.Int4{Int32: int32(log.State), Valid: true},
			RequestID:          pgtype.Text{String: log.RequestId, Valid: true},
			IngestionTs:        pgtype.Int8{Int64: log.IngestionTs, Valid: true},
			SourceTs:           pgtype.Int8{Int64: log.SourceTs, Valid: true},
			ProducerAppName:    pgtype.Text{String: log.ProducerAppName, Valid: true},
			ProducerAppVersion: pgtype.Text{String: log.ProducerAppVersion, Valid: true},
			SdkName:            pgtype.Text{String: log.SdkName, Valid: true},
			SdkVersion:         pgtype.Text{String: log.SdkVersion, Valid: true},
			Entity:             entityJSON,
			SourceMetadata:     sm,
			EntityHash:         pgtype.Text{String: entityHashes[i], Valid: true},
		})
		idMap[log.Id] = ids[i]
	}

	if _, err := r.queries.BulkCreateIngestionLogs(ctx, params); err != nil {
		return nil, err
	}
	return idMap, nil
}

func (r *Repository) allocateIngestionLogIDs(ctx context.Context, n int) ([]int32, error) {
	rows, err := r.pool.Query(ctx, "SELECT nextval('ingestion_logs_id_seq')::int4 FROM generate_series(1, $1)", n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make([]int32, 0, n)
	for rows.Next() {
		var id int32
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// BulkIncrementDuplicateCounts increments the duplicate count for multiple ingestion logs.
func (r *Repository) BulkIncrementDuplicateCounts(ctx context.Context, ids []int32) error {
	return r.queries.BulkIncrementDuplicateCounts(ctx, ids)
}

// BulkPersistChangeSets persists multiple changesets and their changes in a
// single transaction, then bulk-updates ingestion log states.
func (r *Repository) BulkPersistChangeSets(ctx context.Context, items []ops.BulkPersistItem, maxChangeSetsPerLog int32) ([]ops.BulkPersistResult, error) {
	results := make([]ops.BulkPersistResult, len(items))

	var withChanges []int
	for i, item := range items {
		results[i].IngestionLogID = item.IngestionLogID
		if len(item.ChangeSet.Changes) > 0 {
			withChanges = append(withChanges, i)
		}
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	qtx := r.queries.WithTx(tx)

	if len(withChanges) > 0 {
		csIDs, err := r.allocateChangeSetIDs(ctx, len(withChanges))
		if err != nil {
			return nil, fmt.Errorf("failed to allocate change set IDs: %w", err)
		}

		csParams := make([]postgres.BulkCreateChangeSetsParams, 0, len(withChanges))
		var allChangeParams []postgres.BulkCreateChangesParams

		for j, idx := range withChanges {
			item := items[idx]
			csID := csIDs[j]
			results[idx].ChangeSetID = &csID

			p := postgres.BulkCreateChangeSetsParams{
				ID:             csID,
				ExternalID:     item.ChangeSet.ID,
				IngestionLogID: item.IngestionLogID,
			}
			if item.ChangeSet.BranchID != nil {
				p.BranchID = pgtype.Text{String: *item.ChangeSet.BranchID, Valid: true}
			}
			if item.ChangeSet.DeviationName != nil {
				p.DeviationName = pgtype.Text{String: *item.ChangeSet.DeviationName, Valid: true}
			}
			csParams = append(csParams, p)

			for seq, change := range item.ChangeSet.Changes {
				beforeJSON, err := json.Marshal(change.Before)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal before state: %w", err)
				}
				afterJSON, err := json.Marshal(change.After)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal after state: %w", err)
				}

				cp := postgres.BulkCreateChangesParams{
					ExternalID:         change.ID,
					ChangeSetID:        csID,
					ChangeType:         change.ChangeType,
					ObjectType:         change.ObjectType,
					ObjectPrimaryValue: change.ObjectPrimaryValue,
					Before:             beforeJSON,
					After:              afterJSON,
					NewRefs:            change.NewRefs,
					SequenceNumber:     pgtype.Int4{Int32: int32(seq), Valid: true},
				}
				if change.ObjectID != nil {
					cp.ObjectID = pgtype.Int4{Int32: int32(*change.ObjectID), Valid: true}
				}
				if change.ObjectVersion != nil {
					cp.ObjectVersion = pgtype.Int4{Int32: int32(*change.ObjectVersion), Valid: true}
				}
				if change.RefID != nil {
					cp.RefID = pgtype.Text{String: *change.RefID, Valid: true}
				}
				allChangeParams = append(allChangeParams, cp)
			}
		}

		if _, err := qtx.BulkCreateChangeSets(ctx, csParams); err != nil {
			return nil, fmt.Errorf("failed to bulk create change sets: %w", err)
		}
		if _, err := qtx.BulkCreateChanges(ctx, allChangeParams); err != nil {
			return nil, fmt.Errorf("failed to bulk create changes: %w", err)
		}
	}

	stateIDs := make([]int32, len(items))
	states := make([]int32, len(items))
	for i, item := range items {
		stateIDs[i] = item.IngestionLogID
		states[i] = int32(item.NewState)
	}
	if err := qtx.BulkUpdateIngestionLogStates(ctx, postgres.BulkUpdateIngestionLogStatesParams{
		Ids:    stateIDs,
		States: states,
	}); err != nil {
		return nil, fmt.Errorf("failed to bulk update ingestion log states: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	if maxChangeSetsPerLog > 0 && len(withChanges) > 0 {
		ingLogIDs := make([]int32, len(withChanges))
		for j, idx := range withChanges {
			ingLogIDs[j] = items[idx].IngestionLogID
		}
		if err := r.queries.BulkTruncateChangeSets(ctx, postgres.BulkTruncateChangeSetsParams{
			IngestionLogIds: ingLogIDs,
			KeepCount:       maxChangeSetsPerLog,
		}); err != nil {
			return nil, fmt.Errorf("failed to bulk truncate change sets: %w", err)
		}
	}

	return results, nil
}

func (r *Repository) allocateChangeSetIDs(ctx context.Context, n int) ([]int32, error) {
	rows, err := r.pool.Query(ctx, "SELECT nextval('change_sets_id_seq')::int4 FROM generate_series(1, $1)", n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make([]int32, 0, n)
	for rows.Next() {
		var id int32
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// ClaimQueuedIngestionLogs returns a batch of ingestion logs in QUEUED state for processing.
func (r *Repository) ClaimQueuedIngestionLogs(ctx context.Context, batchSize int32) ([]ops.QueuedIngestionLog, error) {
	dbLogs, err := r.queries.ClaimQueuedIngestionLogs(ctx, batchSize)
	if err != nil {
		return nil, err
	}

	result := make([]ops.QueuedIngestionLog, 0, len(dbLogs))
	for _, dbLog := range dbLogs {
		log, err := dbLog.ToProto()
		if err != nil {
			return nil, fmt.Errorf("failed to convert to proto: %w", err)
		}
		result = append(result, ops.QueuedIngestionLog{
			ID:           dbLog.ID,
			IngestionLog: log,
		})
	}
	return result, nil
}

// ClaimQueuedForAutoApply claims a batch of QUEUED ingestion logs for the
// AutoApplyProcessor (combined plan + apply). Each claimed row transitions to
// APPLYING for the duration of the NetBox round-trip; stuck rows are returned
// to QUEUED on startup via ResetApplyingIngestionLogs.
func (r *Repository) ClaimQueuedForAutoApply(ctx context.Context, batchSize int32) ([]ops.QueuedIngestionLog, error) {
	dbLogs, err := r.queries.ClaimQueuedForAutoApply(ctx, batchSize)
	if err != nil {
		return nil, fmt.Errorf("failed to claim queued ingestion logs for auto-apply: %w", err)
	}

	result := make([]ops.QueuedIngestionLog, 0, len(dbLogs))
	for _, dbLog := range dbLogs {
		log, err := dbLog.ToProto()
		if err != nil {
			return nil, fmt.Errorf("failed to convert ingestion log %d to proto: %w", dbLog.ID, err)
		}
		result = append(result, ops.QueuedIngestionLog{
			ID:           dbLog.ID,
			IngestionLog: log,
		})
	}
	return result, nil
}

// ResetApplyingIngestionLogs resets any ingestion logs stuck in APPLYING state back to OPEN.
func (r *Repository) ResetApplyingIngestionLogs(ctx context.Context) error {
	return r.queries.ResetApplyingIngestionLogs(ctx)
}
