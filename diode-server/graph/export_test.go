package graph

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
	"github.com/netboxlabs/diode/diode-server/matching"
)

// Exported wrappers for unexported Builder methods used by external tests.

func (gb *Builder) TestFindNodeByTypeAndID(ctx context.Context, nodeType, externalID string) (*Node, error) {
	return gb.findNodeByTypeAndID(ctx, nodeType, externalID)
}

func (gb *Builder) TestExtractFieldByPath(data map[string]any, fieldPath string) any {
	return gb.extractFieldByPath(data, fieldPath)
}

func (gb *Builder) TestSetFieldByPath(data map[string]any, fieldPath string, value any) {
	gb.setFieldByPath(data, fieldPath, value)
}

func (gb *Builder) TestNeedsSchemaUpdate(node *Node) bool {
	return gb.needsSchemaUpdate(node)
}

func (gb *Builder) TestFindMatchByMetadata(ctx context.Context, entity *diodepb.Entity, nodeType string) *matching.MatchResult {
	return gb.findMatchByMetadata(ctx, entity, nodeType)
}

func (gb *Builder) TestFindNodeBySourceMatchKey(ctx context.Context, nodeType, key string, value any) *matching.MatchResult {
	return gb.findNodeBySourceMatchKey(ctx, nodeType, key, value)
}

func (gb *Builder) TestExtractMetadata(entity *diodepb.Entity) (json.RawMessage, error) {
	return gb.extractMetadata(entity)
}

func (gb *Builder) TestFindNodeByExternalID(ctx context.Context, nodeType, externalID string) *matching.MatchResult {
	return gb.findNodeByExternalID(ctx, nodeType, externalID)
}

func (gb *Builder) TestEnsureDiodeID(metadata json.RawMessage, externalID string) json.RawMessage {
	return gb.ensureDiodeID(metadata, externalID)
}

func (gb *Builder) TestFindNodeByContentHash(ctx context.Context, nodeType, contentHash string) *matching.MatchResult {
	return gb.findNodeByContentHash(ctx, nodeType, contentHash)
}

// Exported wrappers for standalone functions.

func ExportCanBeNil(v reflect.Value) bool {
	return canBeNil(v)
}

func ExportGetEntityTypeName(entity *diodepb.Entity) string {
	return getEntityTypeName(entity)
}

// Field accessors for Builder.

func (gb *Builder) TestSnapshotRetention() int                { return gb.snapshotRetention }
func (gb *Builder) TestEntityMatcher() matching.EntityMatcher { return gb.entityMatcher }
func (gb *Builder) TestNodeCache() map[string]*Node           { return gb.nodeCache }
func (gb *Builder) TestUpdatedNodes() map[string]*Node        { return gb.updatedNodes }
func (gb *Builder) TestSeenInThisRequest() map[string]bool    { return gb.seenInThisRequest }
