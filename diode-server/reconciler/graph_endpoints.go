package reconciler

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/netboxlabs/diode/diode-server/entityhash"
	"github.com/netboxlabs/diode/diode-server/gen/dbstore/postgres"
	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/reconcilerpb"
)

func createEntity(ctx context.Context, graphdb GraphRepository, req *reconcilerpb.CreateEntityRequest) (*reconcilerpb.CreateEntityResponse, error) {
	entity := req.GetEntity()
	if entity == nil || entity.GetEntity() == nil {
		return nil, fmt.Errorf("entity is required")
	}

	// Extract node type from entity
	nodeType := getEntityTypeName(entity)
	if nodeType == "" {
		return nil, fmt.Errorf("failed to determine entity type")
	}

	// Generate new UUID for external ID
	externalID := uuid.New().String()

	// Marshal entity data
	entityData, err := protojson.Marshal(entity)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal entity: %w", err)
	}

	// Convert request metadata to JSON
	var metadata json.RawMessage
	if req.GetMetadata() != nil {
		metadataBytes, err := req.GetMetadata().MarshalJSON()
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		metadata = metadataBytes
	} else {
		metadata = json.RawMessage("{}")
	}

	// Generate content hash for deduplication
	fingerprinter := entityhash.NewEntityFingerprinter()
	contentHash, err := fingerprinter.GenerateEntityHash(entity)
	if err != nil {
		return nil, fmt.Errorf("failed to generate entity hash: %w", err)
	}

	args := postgres.UpsertGraphNodeParams{
		ExternalID:            externalID,
		NodeType:              nodeType,
		Data:                  entityData,
		MatchingSchemaVersion: CurrentSchemaVersion,
		Metadata:              metadata,
		ContentHash:           pgtype.Text{String: contentHash, Valid: true},
	}

	node, err := graphdb.UpsertGraphNode(ctx, args)
	if err != nil {
		return nil, err
	}

	// Build response from upsert result
	return &reconcilerpb.CreateEntityResponse{
		Id:         node.ExternalID,
		ObjectType: node.NodeType,
	}, nil
}

const (
	defaultPageSize = 100
	maxPageSize     = 1000
)

func listEntities(ctx context.Context, graphdb GraphRepository, req *reconcilerpb.ListEntitiesRequest) (*reconcilerpb.ListEntitiesResponse, error) {
	// Parse pagination parameters
	pageSize := defaultPageSize
	if req.GetPageSize() > 0 {
		pageSize = int(req.GetPageSize())
		if pageSize > maxPageSize {
			pageSize = maxPageSize
		}
	}

	// Parse page token (offset encoded as base64)
	offset := int32(0)
	if req.GetPageToken() != "" {
		offsetBytes, err := base64.StdEncoding.DecodeString(req.GetPageToken())
		if err != nil {
			return nil, fmt.Errorf("invalid page token: %w", err)
		}
		offsetVal, err := strconv.ParseInt(string(offsetBytes), 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid page token: %w", err)
		}
		if offsetVal < 0 {
			return nil, fmt.Errorf("invalid page token: negative offset")
		}
		offset = int32(offsetVal)
	}

	// Build query parameters
	params := postgres.ListGraphNodesParams{
		Offset: offset,
		Limit:  int32(pageSize + 1), // Fetch one extra to determine if there's a next page
	}

	// Set object type filter
	if len(req.GetObjectType()) > 0 {
		params.NodeTypes = req.GetObjectType()
	}

	// Set metadata filter
	if req.GetMetadataFilters() != nil {
		metadataJSON, err := req.GetMetadataFilters().MarshalJSON()
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata filters: %w", err)
		}
		params.MetadataFilter = metadataJSON
	}

	// Set timestamp range filters
	if req.IngestionTsStart != nil {
		params.TsStart = pgtype.Timestamptz{
			Time:  time.Unix(0, req.GetIngestionTsStart()*int64(time.Millisecond)),
			Valid: true,
		}
	}
	if req.IngestionTsEnd != nil {
		params.TsEnd = pgtype.Timestamptz{
			Time:  time.Unix(0, req.GetIngestionTsEnd()*int64(time.Millisecond)),
			Valid: true,
		}
	}

	// Execute query
	nodes, err := graphdb.ListGraphNodes(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list entities: %w", err)
	}

	// Determine if there's a next page
	hasNextPage := len(nodes) > pageSize
	if hasNextPage {
		nodes = nodes[:pageSize] // Trim to requested page size
	}

	// Convert nodes to DiodeEntity
	entities := make([]*reconcilerpb.DiodeEntity, 0, len(nodes))
	for _, node := range nodes {
		entity, err := graphNodeToDiodeEntity(node)
		if err != nil {
			return nil, fmt.Errorf("failed to convert node to entity: %w", err)
		}
		entities = append(entities, entity)
	}

	// Build response
	resp := &reconcilerpb.ListEntitiesResponse{
		Entities: entities,
	}

	// Set next page token if there are more results
	if hasNextPage {
		nextOffset := offset + int32(pageSize)
		resp.NextPageToken = base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(int(nextOffset))))
	}

	return resp, nil
}

// graphNodeToDiodeEntity converts a GraphNode to a DiodeEntity
func graphNodeToDiodeEntity(node postgres.GraphNode) (*reconcilerpb.DiodeEntity, error) {
	// Unmarshal entity data
	var entity diodepb.Entity
	if err := protojson.Unmarshal(node.Data, &entity); err != nil {
		return nil, fmt.Errorf("failed to unmarshal entity data: %w", err)
	}

	// Convert metadata to structpb.Struct
	var sourceMetadata *structpb.Struct
	if len(node.Metadata) > 0 {
		var metadataMap map[string]interface{}
		if err := json.Unmarshal(node.Metadata, &metadataMap); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
		var err error
		sourceMetadata, err = structpb.NewStruct(metadataMap)
		if err != nil {
			return nil, fmt.Errorf("failed to create struct from metadata: %w", err)
		}
	}

	// Convert timestamps to milliseconds
	var ingestionTs, createdAt, updatedAt, lastSeenTs int64
	if node.LastSeenTs.Valid {
		ingestionTs = node.LastSeenTs.Time.UnixMilli()
		lastSeenTs = node.LastSeenTs.Time.UnixMilli()
	}
	if node.CreatedAt.Valid {
		createdAt = node.CreatedAt.Time.UnixMilli()
	}
	if node.UpdatedAt.Valid {
		updatedAt = node.UpdatedAt.Time.UnixMilli()
	}

	return &reconcilerpb.DiodeEntity{
		Id:             node.ExternalID,
		ObjectType:     node.NodeType,
		IngestionTs:    ingestionTs,
		Entity:         &entity,
		SourceMetadata: sourceMetadata,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
		LastSeenTs:     lastSeenTs,
		DuplicateCount: node.DuplicateCount,
	}, nil
}
