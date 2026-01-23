package reconciler

import (
	"context"

	"github.com/netboxlabs/diode/diode-server/gen/dbstore/postgres"
)

// GraphNode is an alias to the SQLC-generated GraphNode type
type GraphNode = postgres.GraphNode

// GraphRepository defines the interface for graph database operations.
// This interface wraps the SQLC-generated Queries to allow for mocking in tests.
// Used by EntityMatcher and GraphBuilder.
type GraphRepository interface {
	// FindNodesByFieldMatch finds nodes by matching JSON fields
	FindNodesByFieldMatch(ctx context.Context, arg postgres.FindNodesByFieldMatchParams) ([]postgres.GraphNode, error)

	// GetGraphNodesByType retrieves nodes of a specific type with pagination
	GetGraphNodesByType(ctx context.Context, arg postgres.GetGraphNodesByTypeParams) ([]postgres.GraphNode, error)
}
