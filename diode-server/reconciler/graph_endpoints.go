package reconciler

import (
	"context"
	"fmt"

	"github.com/netboxlabs/diode/diode-server/gen/dbstore/postgres"
	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/reconcilerpb"
)

func createEntity(ctx context.Context, graphdb GraphRepository, req *reconcilerpb.CreateEntityRequest) (*reconcilerpb.CreateEntityResponse, error) {
	entity := req.GetEntity()
	if entity == nil || entity.GetEntity() == nil {
		return nil, fmt.Errorf("entity is required")
	}

	args := postgres.UpsertGraphNodeParams{}
	node, err := graphdb.UpsertGraphNode(ctx, args)
	if err != nil {
		return nil, err
	}

	// // Build response from upsert result
	return &reconcilerpb.CreateEntityResponse{
		Id:         node.ExternalID,
		ObjectType: node.NodeType,
	}, nil
}
