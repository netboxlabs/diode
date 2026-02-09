package reconciler

import (
	"context"
	"encoding/json"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/netboxlabs/diode/diode-server/entityhash"
	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/reconcilerpb"
	"github.com/netboxlabs/diode/diode-server/graph"
)

func createEntity(ctx context.Context, gr graph.Repository, req *reconcilerpb.CreateEntityRequest) (*reconcilerpb.CreateEntityResponse, error) {
	entity := req.GetEntity()
	if entity == nil || entity.GetEntity() == nil {
		return nil, status.Error(codes.InvalidArgument, "entity is required")
	}

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

	args := graph.UpsertNodeParams{
		Data:        entityData,
		Metadata:    metadata,
		ContentHash: newPtr(contentHash),
	}

	node, err := gr.UpsertNode(ctx, args)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create entity: %v", err)
	}

	// Build response from upsert result
	return &reconcilerpb.CreateEntityResponse{
		Id:         node.ExternalID,
		ObjectType: node.NodeType,
	}, nil
}

// can be replaced by `new()` once we move to go1.26+
func newPtr[T any](v T) *T {
	return &v
}
