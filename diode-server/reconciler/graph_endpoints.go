package reconciler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/netboxlabs/diode/diode-server/entityhash"
	"github.com/netboxlabs/diode/diode-server/gen/dbstore/postgres"
	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/reconcilerpb"
	"google.golang.org/protobuf/encoding/protojson"
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
