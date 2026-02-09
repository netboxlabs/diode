package graph

import (
	"context"
	"log/slog"
)

// ListNodesParams contains parameters for listing nodes with pagination and filtering.
type ListNodesParams struct {
	ObjectTypes []string
	PageSize    int32
	PageToken   string
}

// ListNodesResult contains the result of listing nodes.
type ListNodesResult struct {
	Nodes         []NodeWithLatestSnapshot
	NextPageToken string
}

// Reader provides read operations for the entity graph.
type Reader struct {
	repo   Repository
	logger *slog.Logger
}

// NewReader creates a new Reader.
func NewReader(repo Repository, logger *slog.Logger) *Reader {
	return &Reader{
		repo:   repo,
		logger: logger,
	}
}

// ListEntities returns a paginated list of entities with optional filtering.
func (r *Reader) ListEntities(_ context.Context, _ ListNodesParams) (*ListNodesResult, error) {
	// TODO: implement once ListNodes query is added to Repository
	return &ListNodesResult{}, nil
}
