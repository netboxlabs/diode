package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/netboxlabs/diode/diode-server/gen/dbstore/postgres"
	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/reconcilerpb"
	"github.com/netboxlabs/diode/diode-server/reconciler/changeset"
)

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

// CreateIngestionLog creates a new ingestion log.
func (r *Repository) CreateIngestionLog(ctx context.Context, ingestionLog *reconcilerpb.IngestionLog, sourceMetadata []byte) (*int32, error) {
	entityJSON, err := protojson.Marshal(ingestionLog.Entity)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal entity: %w", err)
	}
	params := postgres.CreateIngestionLogParams{
		ExternalID:         ingestionLog.Id,
		ObjectType:         pgtype.Text{String: ingestionLog.ObjectType, Valid: true},
		State:              pgtype.Int4{Int32: int32(ingestionLog.State), Valid: true},
		RequestID:          pgtype.Text{String: ingestionLog.RequestId, Valid: true},
		IngestionTs:        pgtype.Int8{Int64: ingestionLog.IngestionTs, Valid: true},
		ProducerAppName:    pgtype.Text{String: ingestionLog.ProducerAppName, Valid: true},
		ProducerAppVersion: pgtype.Text{String: ingestionLog.ProducerAppVersion, Valid: true},
		SdkName:            pgtype.Text{String: ingestionLog.SdkName, Valid: true},
		SdkVersion:         pgtype.Text{String: ingestionLog.SdkVersion, Valid: true},
		Entity:             entityJSON,
		SourceMetadata:     sourceMetadata,
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
func (r *Repository) UpdateIngestionLogStateWithError(ctx context.Context, id int32, state reconcilerpb.State, ingestionError *reconcilerpb.IngestionError) error {
	params := postgres.UpdateIngestionLogStateWithErrorParams{
		ID:    id,
		State: pgtype.Int4{Int32: int32(state), Valid: true},
	}

	if ingestionError != nil {
		ingestionErrJSON, err := json.Marshal(ingestionError)
		if err != nil {
			return fmt.Errorf("failed to marshal error: %w", err)
		}
		params.Error = ingestionErrJSON
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
		var ingestionErr reconcilerpb.IngestionError
		if ingestionLog.Error != nil {
			if err := protojson.Unmarshal(ingestionLog.Error, &ingestionErr); err != nil {
				return nil, fmt.Errorf("failed to unmarshal error: %w", err)
			}
		}

		log := &reconcilerpb.IngestionLog{
			Id:                 ingestionLog.ExternalID,
			ObjectType:         ingestionLog.ObjectType.String,
			State:              reconcilerpb.State(ingestionLog.State.Int32),
			RequestId:          ingestionLog.RequestID.String,
			IngestionTs:        ingestionLog.IngestionTs.Int64,
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
					ChangeID:   dbChange.ExternalID,
					ChangeType: dbChange.ChangeType,
					ObjectType: dbChange.ObjectType,
					Before:     dbChange.Before,
					After:      dbChange.After,
				}

				objID := int(dbChange.ObjectID.Int32)
				if dbChange.ObjectID.Valid {
					change.ObjectID = &objID
				}
				objVersion := int(dbChange.ObjectVersion.Int32)
				if dbChange.ObjectVersion.Valid {
					change.ObjectVersion = &objVersion
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
				ChangeSetID:   row.ChangeSet.ExternalID,
				ChangeSet:     changes,
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
		if err := tx.Rollback(ctx); err != nil {
			panic(fmt.Errorf("failed to rollback transaction: %w", err))
		}
	}

	qtx := r.queries.WithTx(tx)
	params := postgres.CreateChangeSetParams{
		ExternalID:     changeSet.ChangeSetID,
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

	for i, change := range changeSet.ChangeSet {
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

		changeParams := postgres.CreateChangeParams{
			ExternalID:     change.ChangeID,
			ChangeSetID:    cs.ID,
			ChangeType:     change.ChangeType,
			ObjectType:     change.ObjectType,
			Before:         beforeJSON,
			After:          afterJSON,
			SequenceNumber: pgtype.Int4{Int32: int32(i), Valid: true},
		}
		if change.ObjectID != nil {
			changeParams.ObjectID = pgtype.Int4{Int32: int32(*change.ObjectID), Valid: true}
		}
		if change.ObjectVersion != nil {
			changeParams.ObjectVersion = pgtype.Int4{Int32: int32(*change.ObjectVersion), Valid: true}
		}

		if _, err = qtx.CreateChange(ctx, changeParams); err != nil {
			rollback()
			return nil, fmt.Errorf("failed to create change: %w", err)
		}
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
		states := make([]string, 0, len(filter.State))
		for _, state := range filter.State {
			states = append(states, state.String())
		}
		params.State = states
	}
	if len(filter.ObjectType) > 0 {
		params.ObjectType = filter.ObjectType
	}
	if len(filter.BranchId) > 0 {
		params.BranchID = filter.BranchId
	}

	// TODO(mfiedorowicz): filtering by site is not implemented yet

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
	for _, row := range rawDeviations {
		entity := &diodepb.Entity{}
		if err := protojson.Unmarshal(row.Entity, entity); err != nil {
			return nil, fmt.Errorf("failed to unmarshal entity: %w", err)
		}

		// split producer app name by forward slash and get first element if it exists
		source := row.ProducerAppName.String
		if source != "" {
			source = strings.Split(source, "/")[0]
		}

		var branchID *string
		if row.ChangeSet.BranchID.Valid {
			branchID = &row.ChangeSet.BranchID.String
		}

		var deviationErr *reconcilerpb.DeviationError
		if row.Error != nil {
			deviationErr = &reconcilerpb.DeviationError{}
			if err := protojson.Unmarshal(row.Error, deviationErr); err != nil {
				return nil, fmt.Errorf("failed to unmarshal error: %w", err)
			}
		}

		deviation := &reconcilerpb.Deviation{
			Id:             row.ExternalID,
			Name:           row.ChangeSet.DeviationName.String,
			Source:         source,
			State:          reconcilerpb.State(row.State.Int32),
			ObjectType:     row.ObjectType.String,
			BranchId:       branchID,
			IngestedEntity: entity,
			Error:          deviationErr,
		}

		if row.Changes != nil {
			var dbChanges []postgres.Change
			if err := json.Unmarshal(row.Changes, &dbChanges); err != nil {
				return nil, fmt.Errorf("failed to unmarshal changes: %w", err)
			}

			changes := make([]*reconcilerpb.Change, 0, len(dbChanges))
			for _, dbChange := range dbChanges {
				change := &reconcilerpb.Change{
					Id:                 dbChange.ExternalID,
					ChangeType:         dbChange.ChangeType,
					ObjectType:         dbChange.ObjectType,
					ObjectPrimaryValue: dbChange.ObjectPrimaryValue,
				}
				if dbChange.Before != nil {
					change.Before = dbChange.Before.([]byte)
				}
				if dbChange.After != nil {
					change.After = dbChange.After.([]byte)
				}
				changes = append(changes, change)
			}

			deviation.Changes = changes
		}

		deviations = append(deviations, deviation)
	}

	return deviations, nil
}
