package reconciler

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/netboxlabs/diode/diode-server/entityhash"
	"github.com/netboxlabs/diode/diode-server/gen/dbstore/postgres"
	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/reconcilerpb"
)

func createEntity(ctx context.Context, graphdb GraphRepository, req *reconcilerpb.CreateEntityRequest) (*reconcilerpb.CreateEntityResponse, error) {
	entity := req.GetEntity()
	if entity == nil || entity.GetEntity() == nil {
		return nil, status.Error(codes.InvalidArgument, "entity is required")
	}

	// Extract node type from entity
	nodeType := getEntityTypeName(entity)
	if nodeType == "" {
		return nil, status.Error(codes.InvalidArgument, "failed to determine entity type")
	}

	// Generate new UUID for external ID
	externalID := uuid.New().String()

	// Marshal entity data
	entityData, err := protojson.Marshal(entity)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to marshal entity: %v", err)
	}

	// Convert request metadata to JSON
	var metadata json.RawMessage
	if req.GetMetadata() != nil {
		metadataBytes, err := req.GetMetadata().MarshalJSON()
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to marshal metadata: %v", err)
		}
		metadata = metadataBytes
	} else {
		metadata = json.RawMessage("{}")
	}

	// Generate content hash for deduplication
	fingerprinter := entityhash.NewEntityFingerprinter()
	contentHash, err := fingerprinter.GenerateEntityHash(entity)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate entity hash: %v", err)
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
