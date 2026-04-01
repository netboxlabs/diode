package graph

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
	"github.com/netboxlabs/diode/diode-server/matching"
)

// Exported wrappers for unexported Service methods used by external tests.

func (s *Service) TestFindNodeByTypeAndID(ctx context.Context, nodeType, externalID string) (*Node, error) {
	return s.findNodeByTypeAndID(ctx, nodeType, externalID)
}

func (s *Service) TestExtractFieldByPath(data map[string]any, fieldPath string) any {
	return s.extractFieldByPath(data, fieldPath)
}

func (s *Service) TestSetFieldByPath(data map[string]any, fieldPath string, value any) {
	s.setFieldByPath(data, fieldPath, value)
}

func (s *Service) TestFindMatchByMetadata(ctx context.Context, entity *diodepb.Entity, nodeType string) *matching.MatchResult {
	return s.findMatchByMetadata(ctx, entity, nodeType)
}

func (s *Service) TestFindNodeBySourceMatchKey(ctx context.Context, nodeType, key string, value any) *matching.MatchResult {
	return s.findNodeBySourceMatchKey(ctx, nodeType, key, value)
}

func (s *Service) TestExtractMetadata(entity *diodepb.Entity) (json.RawMessage, error) {
	return s.extractMetadata(entity)
}

// TestExtractMetadataWithRequest sets request-level metadata the same way UpsertEntity does, then runs extractMetadata.
func (s *Service) TestExtractMetadataWithRequest(entity *diodepb.Entity, requestMetadata map[string]any) (json.RawMessage, error) {
	s.requestMetadata = requestMetadata
	return s.extractMetadata(entity)
}

func (s *Service) TestFindNodeByExternalID(ctx context.Context, nodeType, externalID string) *matching.MatchResult {
	return s.findNodeByExternalID(ctx, nodeType, externalID)
}

func (s *Service) TestEnsureDiodeID(metadata json.RawMessage, externalID string) json.RawMessage {
	return s.ensureDiodeID(metadata, externalID)
}

func (s *Service) TestFindNodeByContentHash(ctx context.Context, nodeType, contentHash string) *matching.MatchResult {
	return s.findNodeByContentHash(ctx, nodeType, contentHash)
}

// Exported wrappers for standalone functions.

func ExportCanBeNil(v reflect.Value) bool {
	return canBeNil(v)
}

func ExportGetEntityTypeName(entity *diodepb.Entity) string {
	return getEntityTypeName(entity)
}

// Exported wrappers for propagate functions.

func ExportUpdateRefsInData(data map[string]any, updates []ParsedUpdateForTest) {
	internal := make([]parsedUpdate, len(updates))
	for i, u := range updates {
		internal[i] = parsedUpdate{
			key:       u.Key,
			nodeType:  u.NodeType,
			entityMap: u.EntityMap,
		}
	}
	updateRefsInData(data, internal)
}

func ExportCheckForUpdatedRefs(data any, excludeNodeKey string, updates []ParsedUpdateForTest) bool {
	internal := make([]parsedUpdate, len(updates))
	for i, u := range updates {
		internal[i] = parsedUpdate{
			key:       u.Key,
			nodeType:  u.NodeType,
			entityMap: u.EntityMap,
		}
	}
	return checkForUpdatedRefs(data, excludeNodeKey, internal)
}

// ParsedUpdateForTest is an exported version of parsedUpdate for tests.
type ParsedUpdateForTest struct {
	Key       string
	NodeType  string
	EntityMap map[string]any
}

// Field accessors for Service.

func (s *Service) TestSnapshotRetentionDays() int            { return s.snapshotRetentionDays }
func (s *Service) TestEntityMatcher() matching.EntityMatcher { return s.entityMatcher }
func (s *Service) TestNodeCache() map[string]*Node           { return s.nodeCache }
func (s *Service) TestUpdatedNodes() map[string]*Node        { return s.updatedNodes }
func (s *Service) TestSeenInThisRequest() map[string]bool    { return s.seenInThisRequest }
