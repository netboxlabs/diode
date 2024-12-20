package postgres

import (
	"context"
	"encoding/json"
	"fmt"

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
		IngestionLogUuid:   ingestionLog.Id,
		DataType:           pgtype.Text{String: ingestionLog.DataType, Valid: true},
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
	if filter.DataType != "" {
		params.DataType = pgtype.Text{String: filter.DataType, Valid: true}
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
			Id:                 ingestionLog.IngestionLogUuid,
			DataType:           ingestionLog.DataType.String,
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
					ChangeID:   dbChange.ChangeUuid,
					ChangeType: dbChange.ChangeType,
					ObjectType: dbChange.ObjectType,
					Data:       dbChange.Data,
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

			changeSet := &changeset.ChangeSet{
				ChangeSetID: row.ChangeSet.ChangeSetUuid,
				ChangeSet:   changes,
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
				Id:   row.ChangeSet.ChangeSetUuid,
				Data: compressedChangeSet,
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
		ChangeSetUuid:  changeSet.ChangeSetID,
		IngestionLogID: ingestionLogID,
	}
	if changeSet.BranchID != nil {
		params.BranchID = pgtype.Text{String: *changeSet.BranchID, Valid: true}
	}
	cs, err := qtx.CreateChangeSet(ctx, params)
	if err != nil {
		rollback()
		return nil, fmt.Errorf("failed to create change set: %w", err)
	}

	for i, change := range changeSet.ChangeSet {
		dataJSON, err := json.Marshal(change.Data)
		if err != nil {
			rollback()
			return nil, fmt.Errorf("failed to marshal entity: %w", err)
		}

		changeParams := postgres.CreateChangeParams{
			ChangeUuid:     change.ChangeID,
			ChangeSetID:    cs.ID,
			ChangeType:     change.ChangeType,
			ObjectType:     change.ObjectType,
			Data:           dataJSON,
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
