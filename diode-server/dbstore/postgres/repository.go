package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/netboxlabs/diode/diode-server/entityhash"
	"github.com/netboxlabs/diode/diode-server/gen/dbstore/postgres"
	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/reconcilerpb"
	"github.com/netboxlabs/diode/diode-server/reconciler"
	"github.com/netboxlabs/diode/diode-server/reconciler/changeset"
)

// GraphBuilder interface for graph operations
type GraphBuilder interface {
	ExtractGraph(ctx context.Context, entity *diodepb.Entity) error
	GetCompleteNodeData(ctx context.Context, nodeType, externalID string) (*reconciler.CompleteNodeData, error)
}

// Repository is an interface for interacting with ingestion logs and change sets.
type Repository struct {
	pool         *pgxpool.Pool
	queries      *postgres.Queries
	graphBuilder GraphBuilder
}

// NewRepository creates a new Repository.
func NewRepository(pool *pgxpool.Pool, graphBuilder GraphBuilder) *Repository {
	return &Repository{
		pool:         pool,
		queries:      postgres.New(pool),
		graphBuilder: graphBuilder,
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
		if err := tx.Rollback(ctx); err != nil {
			panic(fmt.Errorf("failed to rollback transaction: %w", err))
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

		changeParams := postgres.CreateChangeParams{
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
			changeParams.ObjectID = pgtype.Int4{Int32: int32(*change.ObjectID), Valid: true}
		}
		if change.ObjectVersion != nil {
			changeParams.ObjectVersion = pgtype.Int4{Int32: int32(*change.ObjectVersion), Valid: true}
		}
		if change.RefID != nil {
			changeParams.RefID = pgtype.Text{String: *change.RefID, Valid: true}
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

// CreateEntityInGraph creates an entity in the graph database using GraphBuilder
func (r *Repository) CreateEntityInGraph(ctx context.Context, entity *diodepb.Entity) (externalID string, nodeType string, err error) {
	if r.graphBuilder == nil {
		return "", "", fmt.Errorf("graph builder not configured")
	}

	// Extract node type from entity before processing
	nodeType = getEntityTypeName(entity)
	if nodeType == "" {
		return "", "", fmt.Errorf("failed to get entity type name")
	}

	// Extract external ID (entity hash) from entity
	fingerprinter := entityhash.NewEntityFingerprinter()
	externalID, err = fingerprinter.GenerateEntityHash(entity)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate entity hash: %w", err)
	}

	// Use GraphBuilder to extract and persist the entity graph
	if err := r.graphBuilder.ExtractGraph(ctx, entity); err != nil {
		return "", "", fmt.Errorf("failed to extract entity graph: %w", err)
	}

	return externalID, nodeType, nil
}

// getEntityTypeName extracts the entity type name from an entity using proto names
// This is similar to the private function in graph_builder.go
func getEntityTypeName(entity *diodepb.Entity) string {
	if entity == nil || entity.GetEntity() == nil {
		return ""
	}

	// Simple approach using reflection to get the type name
	entityWrapper := entity.GetEntity()
	entityType := reflect.TypeOf(entityWrapper)
	if entityType == nil {
		return ""
	}

	// Extract the type name from the wrapper (e.g., "Entity_Device" -> "Device")
	typeName := entityType.Elem().Name()
	if name, found := strings.CutPrefix(typeName, "Entity_"); found {
		return name
	}

	return typeName
}

// ListGraphEntities lists entities from the graph database with filtering
func (r *Repository) ListGraphEntities(ctx context.Context, filter *reconcilerpb.ListEntitiesRequest, limit int32, offset int32) ([]*reconcilerpb.DiodeEntity, error) {
	if r.graphBuilder == nil {
		return nil, fmt.Errorf("graph builder not configured")
	}

	// Build query parameters based on filters
	var nodes []postgres.GraphNode
	var err error

	if len(filter.ObjectType) > 0 {
		// Filter by object types - for now, query the first type
		// TODO: Support multiple object types in a single query
		nodes, err = r.queries.GetGraphNodesByType(ctx, postgres.GetGraphNodesByTypeParams{
			NodeType: filter.ObjectType[0],
			Limit:    limit,
			Offset:   offset,
		})
	} else {
		// No filter - get all nodes with pagination
		// Use a high limit and offset to paginate through all types
		// TODO: Add a proper "GetAllGraphNodes" query
		return nil, fmt.Errorf("listing all entities without object_type filter is not yet implemented")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query graph nodes: %w", err)
	}

	// Convert graph nodes to DiodeEntity responses
	entities := make([]*reconcilerpb.DiodeEntity, 0, len(nodes))
	for _, node := range nodes {
		// Get complete node data with snapshot
		completeData, err := r.graphBuilder.GetCompleteNodeData(ctx, node.NodeType, node.ExternalID)
		if err != nil {
			// Log warning but continue processing other nodes
			continue
		}

		// Parse the complete entity data from snapshot
		entity := &diodepb.Entity{}
		if err := protojson.Unmarshal(completeData.CompleteData, entity); err != nil {
			// Skip nodes with invalid data
			continue
		}

		// Convert timestamps
		createdAt := node.CreatedAt.Time.UnixNano()
		updatedAt := node.UpdatedAt.Time.UnixNano()
		lastSeenTs := int64(0)
		if node.LastSeenTs.Valid {
			lastSeenTs = node.LastSeenTs.Time.UnixNano()
		}

		diodeEntity := &reconcilerpb.DiodeEntity{
			Id:             node.ExternalID, // Using external_id as the UUID
			ObjectType:     node.NodeType,
			Entity:         entity,
			CreatedAt:      createdAt,
			UpdatedAt:      updatedAt,
			LastSeenTs:     lastSeenTs,
			DuplicateCount: node.DuplicateCount,
			// Note: source_metadata, ingestion_ts, and source_ts are not stored in graph_nodes
			// These would need to be added to the schema if required
		}

		entities = append(entities, diodeEntity)
	}

	return entities, nil
}
