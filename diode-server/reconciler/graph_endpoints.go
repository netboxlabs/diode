package reconciler

import (
	"context"
	"fmt"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/reconcilerpb"
)

func createEntity(ctx context.Context, graphdb *GraphBuilder, req *reconcilerpb.CreateEntityRequest) (*reconcilerpb.CreateEntityResponse, error) {
	entity := req.GetEntity()
	if entity == nil || entity.GetEntity() == nil {
		return nil, fmt.Errorf("entity is required")
	}

	// Upsert the node
	ack, err := graphdb.processEntityRecursively(ctx, entity) //.UpsertGraphNode(ctx, params)
	if err != nil {
		return nil, err
	}

	// Build response from upsert result
	return &reconcilerpb.CreateEntityResponse{
		Id:         ack.ExternalID,
		ObjectType: ack.NodeType,
	}, nil
}
