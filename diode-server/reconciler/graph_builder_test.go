package reconciler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"reflect"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/netboxlabs/diode/diode-server/gen/dbstore/postgres"
	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
	"github.com/netboxlabs/diode/diode-server/reconciler/mocks"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestNewGraphBuilder(t *testing.T) {
	repo := new(mocks.GraphRepository)
	logger := newTestLogger()

	gb := NewGraphBuilder(repo, logger)

	assert.NotNil(t, gb)
	assert.NotNil(t, gb.nodeCache)
	assert.NotNil(t, gb.updatedNodes)
	assert.NotNil(t, gb.seenInThisRequest)
	assert.Equal(t, DefaultSnapshotRetention, gb.snapshotRetention)
	assert.Nil(t, gb.entityMatcher)
}

func TestNewGraphBuilderWithMatcher(t *testing.T) {
	repo := new(mocks.GraphRepository)
	logger := newTestLogger()

	// Test with nil matcher
	gb := NewGraphBuilderWithMatcher(repo, logger, nil)

	assert.NotNil(t, gb)
	assert.Nil(t, gb.entityMatcher)
}

func TestSetSnapshotRetention(t *testing.T) {
	repo := new(mocks.GraphRepository)
	logger := newTestLogger()
	gb := NewGraphBuilder(repo, logger)

	// Test setting valid retention
	gb.SetSnapshotRetention(10)
	assert.Equal(t, 10, gb.snapshotRetention)

	// Test setting zero (should not change)
	gb.SetSnapshotRetention(0)
	assert.Equal(t, 10, gb.snapshotRetention)

	// Test setting negative (should not change)
	gb.SetSnapshotRetention(-5)
	assert.Equal(t, 10, gb.snapshotRetention)
}

func TestExtractGraph_NilEntity(t *testing.T) {
	repo := new(mocks.GraphRepository)
	logger := newTestLogger()
	gb := NewGraphBuilder(repo, logger)

	err := gb.ExtractGraph(context.Background(), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "entity or entity content is nil")
}

func TestExtractGraph_EmptyEntity(t *testing.T) {
	repo := new(mocks.GraphRepository)
	logger := newTestLogger()
	gb := NewGraphBuilder(repo, logger)

	entity := &diodepb.Entity{}
	err := gb.ExtractGraph(context.Background(), entity)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "entity or entity content is nil")
}

func TestExtractGraph_SimpleDevice(t *testing.T) {
	repo := new(mocks.GraphRepository)
	logger := newTestLogger()
	gb := NewGraphBuilder(repo, logger)

	ctx := context.Background()

	// Create a simple device entity
	device := &diodepb.Device{
		Name: ptrString("test-device"),
	}
	entity := &diodepb.Entity{
		Entity: &diodepb.Entity_Device{Device: device},
	}

	// Setup mock expectations - content hash lookup returns no match (triggers new node creation)
	repo.EXPECT().FindNodeByContentHash(ctx, mock.AnythingOfType("postgres.FindNodeByContentHashParams")).
		Return(postgres.GraphNode{}, pgx.ErrNoRows).Once()

	repo.EXPECT().UpsertGraphNode(ctx, mock.AnythingOfType("postgres.UpsertGraphNodeParams")).
		Return(postgres.GraphNode{
			ID:                    1,
			ExternalID:            "test-hash",
			NodeType:              "Device",
			Data:                  json.RawMessage(`{}`),
			DuplicateCount:        1,
			MatchingSchemaVersion: 1,
		}, nil).Once()

	repo.EXPECT().InsertSnapshot(ctx, mock.AnythingOfType("postgres.InsertSnapshotParams")).
		Return(postgres.GraphNodeSnapshot{}, nil).Once()

	repo.EXPECT().CleanupOldSnapshots(ctx, mock.AnythingOfType("postgres.CleanupOldSnapshotsParams")).
		Return(nil).Once()

	err := gb.ExtractGraph(ctx, entity)
	assert.NoError(t, err)

	repo.AssertExpectations(t)
}

func TestExtractGraph_DeviceWithSite(t *testing.T) {
	repo := new(mocks.GraphRepository)
	logger := newTestLogger()
	gb := NewGraphBuilder(repo, logger)

	ctx := context.Background()

	// Create a device with a site reference
	site := &diodepb.Site{
		Name: "test-site",
	}
	device := &diodepb.Device{
		Name: ptrString("test-device"),
		Site: site,
	}
	entity := &diodepb.Entity{
		Entity: &diodepb.Entity_Device{Device: device},
	}

	// Setup mock expectations - content hash lookup returns no match for both Device and Site
	repo.EXPECT().FindNodeByContentHash(ctx, mock.MatchedBy(func(params postgres.FindNodeByContentHashParams) bool {
		return params.NodeType == "Device"
	})).Return(postgres.GraphNode{}, pgx.ErrNoRows).Once()

	repo.EXPECT().FindNodeByContentHash(ctx, mock.MatchedBy(func(params postgres.FindNodeByContentHashParams) bool {
		return params.NodeType == "Site"
	})).Return(postgres.GraphNode{}, pgx.ErrNoRows).Once()

	// Setup mock expectations for device node
	repo.EXPECT().UpsertGraphNode(ctx, mock.MatchedBy(func(params postgres.UpsertGraphNodeParams) bool {
		return params.NodeType == "Device"
	})).Return(postgres.GraphNode{
		ID:                    1,
		ExternalID:            "device-hash",
		NodeType:              "Device",
		Data:                  json.RawMessage(`{}`),
		DuplicateCount:        1,
		MatchingSchemaVersion: 1,
	}, nil).Once()

	repo.EXPECT().InsertSnapshot(ctx, mock.MatchedBy(func(params postgres.InsertSnapshotParams) bool {
		return params.NodeID == 1
	})).Return(postgres.GraphNodeSnapshot{}, nil).Once()

	repo.EXPECT().CleanupOldSnapshots(ctx, mock.MatchedBy(func(params postgres.CleanupOldSnapshotsParams) bool {
		return params.NodeID == 1
	})).Return(nil).Once()

	// Setup mock expectations for site node
	repo.EXPECT().UpsertGraphNode(ctx, mock.MatchedBy(func(params postgres.UpsertGraphNodeParams) bool {
		return params.NodeType == "Site"
	})).Return(postgres.GraphNode{
		ID:                    2,
		ExternalID:            "site-hash",
		NodeType:              "Site",
		Data:                  json.RawMessage(`{}`),
		DuplicateCount:        1,
		MatchingSchemaVersion: 1,
	}, nil).Once()

	repo.EXPECT().InsertSnapshot(ctx, mock.MatchedBy(func(params postgres.InsertSnapshotParams) bool {
		return params.NodeID == 2
	})).Return(postgres.GraphNodeSnapshot{}, nil).Once()

	repo.EXPECT().CleanupOldSnapshots(ctx, mock.MatchedBy(func(params postgres.CleanupOldSnapshotsParams) bool {
		return params.NodeID == 2
	})).Return(nil).Once()

	// Setup mock expectations for edges (bidirectional)
	repo.EXPECT().UpsertGraphEdge(ctx, mock.MatchedBy(func(params postgres.UpsertGraphEdgeParams) bool {
		return params.EdgeType == "BELONGS_TO_SITE"
	})).Return(nil).Once()

	repo.EXPECT().UpsertGraphEdge(ctx, mock.MatchedBy(func(params postgres.UpsertGraphEdgeParams) bool {
		return params.EdgeType == "HAS_DEVICE"
	})).Return(nil).Once()

	err := gb.ExtractGraph(ctx, entity)
	assert.NoError(t, err)

	repo.AssertExpectations(t)
}

func TestFindNodeByTypeAndID_Found(t *testing.T) {
	repo := new(mocks.GraphRepository)
	logger := newTestLogger()
	gb := NewGraphBuilder(repo, logger)

	ctx := context.Background()

	repo.EXPECT().FindGraphNode(ctx, postgres.FindGraphNodeParams{
		NodeType:   "Device",
		ExternalID: "test-id",
	}).Return(postgres.GraphNode{
		ID:             1,
		ExternalID:     "test-id",
		NodeType:       "Device",
		Data:           json.RawMessage(`{}`),
		DuplicateCount: 1,
	}, nil).Once()

	node, err := gb.findNodeByTypeAndID(ctx, "Device", "test-id")
	require.NoError(t, err)
	assert.NotNil(t, node)
	assert.Equal(t, int64(1), node.ID)
	assert.Equal(t, "test-id", node.ExternalID)
	assert.Equal(t, "Device", node.NodeType)

	repo.AssertExpectations(t)
}

func TestFindNodeByTypeAndID_NotFound(t *testing.T) {
	repo := new(mocks.GraphRepository)
	logger := newTestLogger()
	gb := NewGraphBuilder(repo, logger)

	ctx := context.Background()

	repo.EXPECT().FindGraphNode(ctx, postgres.FindGraphNodeParams{
		NodeType:   "Device",
		ExternalID: "nonexistent",
	}).Return(postgres.GraphNode{}, pgx.ErrNoRows).Once()

	node, err := gb.findNodeByTypeAndID(ctx, "Device", "nonexistent")
	require.NoError(t, err)
	assert.Nil(t, node)

	repo.AssertExpectations(t)
}

func TestCanBeNil(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected bool
	}{
		{"pointer", new(string), true},
		{"nil pointer", (*string)(nil), true},
		{"slice", []string{}, true},
		{"map", map[string]string{}, true},
		{"interface", any(nil), true},
		{"string", "test", false},
		{"int", 42, false},
		{"struct", struct{}{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := reflect.ValueOf(tt.value)
			// For nil interface, create an interface-typed reflect.Value
			if tt.name == "interface" {
				var i any
				v = reflect.ValueOf(&i).Elem()
			}
			result := canBeNil(v)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetEntityTypeName(t *testing.T) {
	tests := []struct {
		name     string
		entity   *diodepb.Entity
		expected string
	}{
		{
			name:     "nil entity",
			entity:   nil,
			expected: "",
		},
		{
			name:     "empty entity",
			entity:   &diodepb.Entity{},
			expected: "",
		},
		{
			name: "device entity",
			entity: &diodepb.Entity{
				Entity: &diodepb.Entity_Device{Device: &diodepb.Device{}},
			},
			expected: "Device",
		},
		{
			name: "site entity",
			entity: &diodepb.Entity{
				Entity: &diodepb.Entity_Site{Site: &diodepb.Site{}},
			},
			expected: "Site",
		},
		{
			name: "interface entity",
			entity: &diodepb.Entity{
				Entity: &diodepb.Entity_Interface{Interface: &diodepb.Interface{}},
			},
			expected: "Interface",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getEntityTypeName(tt.entity)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractFieldByPath(t *testing.T) {
	repo := new(mocks.GraphRepository)
	logger := newTestLogger()
	gb := NewGraphBuilder(repo, logger)

	data := map[string]any{
		"name": "test",
		"device": map[string]any{
			"name": "test-device",
			"site": map[string]any{
				"name": "test-site",
			},
		},
	}

	tests := []struct {
		name     string
		path     string
		expected any
	}{
		{"simple field", "name", "test"},
		{"nested field", "device.name", "test-device"},
		{"deeply nested", "device.site.name", "test-site"},
		{"nonexistent field", "nonexistent", nil},
		{"nonexistent nested", "device.nonexistent", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gb.extractFieldByPath(data, tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSetFieldByPath(t *testing.T) {
	repo := new(mocks.GraphRepository)
	logger := newTestLogger()
	gb := NewGraphBuilder(repo, logger)

	tests := []struct {
		name     string
		path     string
		value    any
		expected map[string]any
	}{
		{
			name:     "simple field",
			path:     "name",
			value:    "test",
			expected: map[string]any{"name": "test"},
		},
		{
			name:  "nested field",
			path:  "device.name",
			value: "test-device",
			expected: map[string]any{
				"device": map[string]any{
					"name": "test-device",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make(map[string]any)
			gb.setFieldByPath(data, tt.path, tt.value)
			assert.Equal(t, tt.expected, data)
		})
	}
}

func TestGetCompleteNodeData(t *testing.T) {
	repo := new(mocks.GraphRepository)
	logger := newTestLogger()
	gb := NewGraphBuilder(repo, logger)

	ctx := context.Background()

	repo.EXPECT().GetNodeWithLatestSnapshot(ctx, postgres.GetNodeWithLatestSnapshotParams{
		NodeType:   "Device",
		ExternalID: "test-id",
	}).Return(postgres.GetNodeWithLatestSnapshotRow{
		ID:                    1,
		ExternalID:            "test-id",
		NodeType:              "Device",
		MatchingData:          json.RawMessage(`{"name":"test"}`),
		SnapshotData:          json.RawMessage(`{"name":"test","serial":"123"}`),
		DuplicateCount:        1,
		MatchingSchemaVersion: 1,
		SequenceNumber:        1,
		SnapshotCreatedAt:     pgtype.Timestamptz{Valid: true},
	}, nil).Once()

	result, err := gb.GetCompleteNodeData(ctx, "Device", "test-id")
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(1), result.ID)
	assert.Equal(t, "test-id", result.ExternalID)
	assert.Equal(t, json.RawMessage(`{"name":"test","serial":"123"}`), result.CompleteData)

	repo.AssertExpectations(t)
}

func TestGetNodeSnapshots(t *testing.T) {
	repo := new(mocks.GraphRepository)
	logger := newTestLogger()
	gb := NewGraphBuilder(repo, logger)

	ctx := context.Background()

	expectedSnapshots := []postgres.GraphNodeSnapshot{
		{ID: 1, NodeID: 1, SequenceNumber: 1, SnapshotData: json.RawMessage(`{}`)},
		{ID: 2, NodeID: 1, SequenceNumber: 2, SnapshotData: json.RawMessage(`{}`)},
	}

	repo.EXPECT().GetSnapshotsByNode(ctx, postgres.GetSnapshotsByNodeParams{
		NodeID: 1,
		Limit:  10,
		Offset: 0,
	}).Return(expectedSnapshots, nil).Once()

	snapshots, err := gb.GetNodeSnapshots(ctx, 1, 10, 0)
	require.NoError(t, err)
	assert.Len(t, snapshots, 2)

	repo.AssertExpectations(t)
}

func TestNeedsSchemaUpdate(t *testing.T) {
	repo := new(mocks.GraphRepository)
	logger := newTestLogger()
	gb := NewGraphBuilder(repo, logger)

	tests := []struct {
		name     string
		version  int32
		expected bool
	}{
		{"older version", 0, true},
		{"current version", CurrentSchemaVersion, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &postgres.GraphNode{MatchingSchemaVersion: tt.version}
			result := gb.needsSchemaUpdate(node)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper function to create string pointer
func ptrString(s string) *string {
	return &s
}

func TestFindMatchByMetadata_DiodeIDMatch(t *testing.T) {
	repo := new(mocks.GraphRepository)
	logger := newTestLogger()
	gb := NewGraphBuilder(repo, logger)

	ctx := context.Background()

	// Create metadata with source_match.diode_id
	metadata, err := structpb.NewStruct(map[string]any{
		"source_match": map[string]any{
			"diode_id": "device-123",
		},
	})
	require.NoError(t, err)

	device := &diodepb.Device{
		Name:     ptrString("test-device"),
		Metadata: metadata,
	}
	entity := &diodepb.Entity{
		Entity: &diodepb.Entity_Device{Device: device},
	}

	// Mock FindGraphNode for direct diode_id lookup (diode_id == externalID)
	repo.EXPECT().FindGraphNode(ctx, postgres.FindGraphNodeParams{
		NodeType:   "Device",
		ExternalID: "device-123", // diode_id is used as externalID
	}).Return(postgres.GraphNode{
		ID:             1,
		ExternalID:     "device-123",
		NodeType:       "Device",
		Data:           json.RawMessage(`{"device":{"name":"test-device"}}`),
		DuplicateCount: 1,
	}, nil).Once()

	result := gb.findMatchByMetadata(ctx, entity, "Device")

	require.NotNil(t, result)
	assert.Equal(t, int64(1), *result.NodeID)
	assert.Equal(t, "device-123", *result.ExternalID)
	assert.Equal(t, float64(1.0), float64(result.Confidence))
	assert.Contains(t, result.MatchingFields, "source_match.diode_id")

	repo.AssertExpectations(t)
}

func TestFindMatchByMetadata_OtherSourceMatchKey(t *testing.T) {
	repo := new(mocks.GraphRepository)
	logger := newTestLogger()
	gb := NewGraphBuilder(repo, logger)

	ctx := context.Background()

	// Create metadata with source_match.netbox_id (not diode_id)
	metadata, err := structpb.NewStruct(map[string]any{
		"source_match": map[string]any{
			"netbox_id": "456",
		},
	})
	require.NoError(t, err)

	device := &diodepb.Device{
		Name:     ptrString("test-device"),
		Metadata: metadata,
	}
	entity := &diodepb.Entity{
		Entity: &diodepb.Entity_Device{Device: device},
	}

	// Mock FindNodeByMetadata to return a match for netbox_id
	repo.EXPECT().FindNodeByMetadata(ctx, mock.MatchedBy(func(params postgres.FindNodeByMetadataParams) bool {
		var filter map[string]any
		if err := json.Unmarshal(params.MetadataFilter, &filter); err != nil {
			return false
		}
		sourceMatch, ok := filter["source_match"].(map[string]any)
		return params.NodeType == "Device" && ok && sourceMatch["netbox_id"] == "456"
	})).Return(postgres.GraphNode{
		ID:             2,
		ExternalID:     "existing-device-hash-2",
		NodeType:       "Device",
		Data:           json.RawMessage(`{"device":{"name":"test-device"}}`),
		DuplicateCount: 1,
	}, nil).Once()

	result := gb.findMatchByMetadata(ctx, entity, "Device")

	require.NotNil(t, result)
	assert.Equal(t, int64(2), *result.NodeID)
	assert.Equal(t, "existing-device-hash-2", *result.ExternalID)
	assert.Equal(t, float64(1.0), float64(result.Confidence))
	assert.Contains(t, result.MatchingFields, "source_match.netbox_id")

	repo.AssertExpectations(t)
}

func TestFindMatchByMetadata_DiodeIDPriorityOverOther(t *testing.T) {
	repo := new(mocks.GraphRepository)
	logger := newTestLogger()
	gb := NewGraphBuilder(repo, logger)

	ctx := context.Background()

	// Create metadata with both diode_id and netbox_id
	metadata, err := structpb.NewStruct(map[string]any{
		"source_match": map[string]any{
			"diode_id":  "device-123",
			"netbox_id": "456",
		},
	})
	require.NoError(t, err)

	device := &diodepb.Device{
		Name:     ptrString("test-device"),
		Metadata: metadata,
	}
	entity := &diodepb.Entity{
		Entity: &diodepb.Entity_Device{Device: device},
	}

	// Mock: diode_id lookup should be called first and return a match
	// (netbox_id should NOT be called because diode_id takes priority)
	repo.EXPECT().FindGraphNode(ctx, postgres.FindGraphNodeParams{
		NodeType:   "Device",
		ExternalID: "device-123", // diode_id is used as externalID
	}).Return(postgres.GraphNode{
		ID:             1,
		ExternalID:     "device-123",
		NodeType:       "Device",
		Data:           json.RawMessage(`{}`),
		DuplicateCount: 1,
	}, nil).Once()

	result := gb.findMatchByMetadata(ctx, entity, "Device")

	require.NotNil(t, result)
	assert.Equal(t, int64(1), *result.NodeID)
	assert.Contains(t, result.MatchingFields, "source_match.diode_id")

	repo.AssertExpectations(t)
}

func TestFindMatchByMetadata_NoSourceMatchKey(t *testing.T) {
	repo := new(mocks.GraphRepository)
	logger := newTestLogger()
	gb := NewGraphBuilder(repo, logger)

	ctx := context.Background()

	// Create metadata WITHOUT source_match key
	metadata, err := structpb.NewStruct(map[string]any{
		"some_other_key": "value",
	})
	require.NoError(t, err)

	device := &diodepb.Device{
		Name:     ptrString("test-device"),
		Metadata: metadata,
	}
	entity := &diodepb.Entity{
		Entity: &diodepb.Entity_Device{Device: device},
	}

	// FindNodeByMetadata should NOT be called because there's no source_match
	result := gb.findMatchByMetadata(ctx, entity, "Device")

	assert.Nil(t, result)
	repo.AssertExpectations(t)
}

func TestFindMatchByMetadata_EmptySourceMatch(t *testing.T) {
	repo := new(mocks.GraphRepository)
	logger := newTestLogger()
	gb := NewGraphBuilder(repo, logger)

	ctx := context.Background()

	// Create metadata with empty source_match
	metadata, err := structpb.NewStruct(map[string]any{
		"source_match": map[string]any{},
	})
	require.NoError(t, err)

	device := &diodepb.Device{
		Name:     ptrString("test-device"),
		Metadata: metadata,
	}
	entity := &diodepb.Entity{
		Entity: &diodepb.Entity_Device{Device: device},
	}

	// FindNodeByMetadata should NOT be called because source_match is empty
	result := gb.findMatchByMetadata(ctx, entity, "Device")

	assert.Nil(t, result)
	repo.AssertExpectations(t)
}

func TestFindMatchByMetadata_NilEntity(t *testing.T) {
	repo := new(mocks.GraphRepository)
	logger := newTestLogger()
	gb := NewGraphBuilder(repo, logger)

	ctx := context.Background()

	result := gb.findMatchByMetadata(ctx, nil, "Device")

	assert.Nil(t, result)
	repo.AssertExpectations(t)
}

func TestFindMatchByMetadata_NoMetadata(t *testing.T) {
	repo := new(mocks.GraphRepository)
	logger := newTestLogger()
	gb := NewGraphBuilder(repo, logger)

	ctx := context.Background()

	// Create device without metadata
	device := &diodepb.Device{
		Name: ptrString("test-device"),
	}
	entity := &diodepb.Entity{
		Entity: &diodepb.Entity_Device{Device: device},
	}

	result := gb.findMatchByMetadata(ctx, entity, "Device")

	assert.Nil(t, result)
	repo.AssertExpectations(t)
}

func TestFindMatchByMetadata_NoMatch(t *testing.T) {
	repo := new(mocks.GraphRepository)
	logger := newTestLogger()
	gb := NewGraphBuilder(repo, logger)

	ctx := context.Background()

	// Create metadata with source_match that won't find a match
	metadata, err := structpb.NewStruct(map[string]any{
		"source_match": map[string]any{
			"vcenter_id": "vm-999",
		},
	})
	require.NoError(t, err)

	device := &diodepb.Device{
		Name:     ptrString("test-device"),
		Metadata: metadata,
	}
	entity := &diodepb.Entity{
		Entity: &diodepb.Entity_Device{Device: device},
	}

	// Mock FindNodeByMetadata to return no match
	repo.EXPECT().FindNodeByMetadata(ctx, mock.AnythingOfType("postgres.FindNodeByMetadataParams")).
		Return(postgres.GraphNode{}, pgx.ErrNoRows).Once()

	result := gb.findMatchByMetadata(ctx, entity, "Device")

	assert.Nil(t, result)
	repo.AssertExpectations(t)
}

func TestFindNodeBySourceMatchKey(t *testing.T) {
	repo := new(mocks.GraphRepository)
	logger := newTestLogger()
	gb := NewGraphBuilder(repo, logger)

	ctx := context.Background()

	// Test successful lookup
	repo.EXPECT().FindNodeByMetadata(ctx, mock.MatchedBy(func(params postgres.FindNodeByMetadataParams) bool {
		var filter map[string]any
		if err := json.Unmarshal(params.MetadataFilter, &filter); err != nil {
			return false
		}
		sourceMatch, ok := filter["source_match"].(map[string]any)
		return params.NodeType == "Site" && ok && sourceMatch["site_id"] == "site-abc"
	})).Return(postgres.GraphNode{
		ID:             5,
		ExternalID:     "site-hash",
		NodeType:       "Site",
		Data:           json.RawMessage(`{"site":{"name":"test-site"}}`),
		DuplicateCount: 1,
	}, nil).Once()

	result := gb.findNodeBySourceMatchKey(ctx, "Site", "site_id", "site-abc")

	require.NotNil(t, result)
	assert.Equal(t, int64(5), *result.NodeID)
	assert.Equal(t, "site-hash", *result.ExternalID)
	assert.Equal(t, float64(1.0), float64(result.Confidence))
	assert.Contains(t, result.MatchReason, "Metadata match: source_match.site_id=site-abc")

	repo.AssertExpectations(t)
}

func TestFindNodeBySourceMatchKey_NotFound(t *testing.T) {
	repo := new(mocks.GraphRepository)
	logger := newTestLogger()
	gb := NewGraphBuilder(repo, logger)

	ctx := context.Background()

	repo.EXPECT().FindNodeByMetadata(ctx, mock.AnythingOfType("postgres.FindNodeByMetadataParams")).
		Return(postgres.GraphNode{}, pgx.ErrNoRows).Once()

	result := gb.findNodeBySourceMatchKey(ctx, "Site", "site_id", "nonexistent")

	assert.Nil(t, result)
	repo.AssertExpectations(t)
}

func TestFindNodeBySourceMatchKey_DatabaseError(t *testing.T) {
	repo := new(mocks.GraphRepository)
	logger := newTestLogger()
	gb := NewGraphBuilder(repo, logger)

	ctx := context.Background()

	// Simulate a database connection error (not ErrNoRows)
	dbErr := errors.New("database connection failed")
	repo.EXPECT().FindNodeByMetadata(ctx, mock.AnythingOfType("postgres.FindNodeByMetadataParams")).
		Return(postgres.GraphNode{}, dbErr).Once()

	result := gb.findNodeBySourceMatchKey(ctx, "Site", "site_id", "test-value")

	assert.Nil(t, result)
	repo.AssertExpectations(t)
}

func TestExtractMetadata(t *testing.T) {
	repo := new(mocks.GraphRepository)
	logger := newTestLogger()
	gb := NewGraphBuilder(repo, logger)

	tests := []struct {
		name     string
		entity   *diodepb.Entity
		expected string
	}{
		{
			name:     "nil entity",
			entity:   nil,
			expected: "{}",
		},
		{
			name:     "entity without metadata",
			entity:   &diodepb.Entity{Entity: &diodepb.Entity_Device{Device: &diodepb.Device{Name: ptrString("test")}}},
			expected: "{}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := gb.extractMetadata(tt.entity)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, string(result))
		})
	}
}

func TestExtractMetadata_WithMetadata(t *testing.T) {
	repo := new(mocks.GraphRepository)
	logger := newTestLogger()
	gb := NewGraphBuilder(repo, logger)

	metadata, err := structpb.NewStruct(map[string]any{
		"key": "value",
	})
	require.NoError(t, err)

	device := &diodepb.Device{
		Name:     ptrString("test-device"),
		Metadata: metadata,
	}
	entity := &diodepb.Entity{
		Entity: &diodepb.Entity_Device{Device: device},
	}

	result, err := gb.extractMetadata(entity)
	require.NoError(t, err)

	var parsed map[string]any
	err = json.Unmarshal(result, &parsed)
	require.NoError(t, err)
	assert.Equal(t, "value", parsed["key"])
}

func TestFindNodeByExternalID(t *testing.T) {
	repo := new(mocks.GraphRepository)
	logger := newTestLogger()
	gb := NewGraphBuilder(repo, logger)

	ctx := context.Background()

	repo.EXPECT().FindGraphNode(ctx, postgres.FindGraphNodeParams{
		NodeType:   "Device",
		ExternalID: "device-abc-123",
	}).Return(postgres.GraphNode{
		ID:             10,
		ExternalID:     "device-abc-123",
		NodeType:       "Device",
		Data:           json.RawMessage(`{"device":{"name":"test"}}`),
		DuplicateCount: 1,
	}, nil).Once()

	result := gb.findNodeByExternalID(ctx, "Device", "device-abc-123")

	require.NotNil(t, result)
	assert.Equal(t, int64(10), *result.NodeID)
	assert.Equal(t, "device-abc-123", *result.ExternalID)
	assert.Equal(t, float64(1.0), float64(result.Confidence))
	assert.Contains(t, result.MatchReason, "Direct lookup")

	repo.AssertExpectations(t)
}

func TestFindNodeByExternalID_NotFound(t *testing.T) {
	repo := new(mocks.GraphRepository)
	logger := newTestLogger()
	gb := NewGraphBuilder(repo, logger)

	ctx := context.Background()

	repo.EXPECT().FindGraphNode(ctx, mock.AnythingOfType("postgres.FindGraphNodeParams")).
		Return(postgres.GraphNode{}, pgx.ErrNoRows).Once()

	result := gb.findNodeByExternalID(ctx, "Device", "nonexistent")

	assert.Nil(t, result)
	repo.AssertExpectations(t)
}

func TestEnsureDiodeID_AddsWhenMissing(t *testing.T) {
	repo := new(mocks.GraphRepository)
	logger := newTestLogger()
	gb := NewGraphBuilder(repo, logger)

	// Empty metadata
	metadata := json.RawMessage(`{}`)
	result := gb.ensureDiodeID(metadata, "ext-123")

	var parsed map[string]any
	err := json.Unmarshal(result, &parsed)
	require.NoError(t, err)

	sourceMatch, ok := parsed["source_match"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "ext-123", sourceMatch["diode_id"])
}

func TestEnsureDiodeID_OverwritesExisting(t *testing.T) {
	repo := new(mocks.GraphRepository)
	logger := newTestLogger()
	gb := NewGraphBuilder(repo, logger)

	// Metadata with existing diode_id (client-provided)
	metadata := json.RawMessage(`{"source_match":{"diode_id":"client-provided-id","other":"value"}}`)
	result := gb.ensureDiodeID(metadata, "system-uuid-123")

	var parsed map[string]any
	err := json.Unmarshal(result, &parsed)
	require.NoError(t, err)

	sourceMatch, ok := parsed["source_match"].(map[string]any)
	require.True(t, ok)
	// Should overwrite client-provided diode_id with system UUID
	assert.Equal(t, "system-uuid-123", sourceMatch["diode_id"])
	// Other metadata preserved
	assert.Equal(t, "value", sourceMatch["other"])
}

func TestEnsureDiodeID_AddsToExistingSourceMatch(t *testing.T) {
	repo := new(mocks.GraphRepository)
	logger := newTestLogger()
	gb := NewGraphBuilder(repo, logger)

	// Metadata with source_match but no diode_id
	metadata := json.RawMessage(`{"source_match":{"netbox_id":"456"}}`)
	result := gb.ensureDiodeID(metadata, "ext-789")

	var parsed map[string]any
	err := json.Unmarshal(result, &parsed)
	require.NoError(t, err)

	sourceMatch, ok := parsed["source_match"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "ext-789", sourceMatch["diode_id"])
	assert.Equal(t, "456", sourceMatch["netbox_id"])
}

func TestFindNodeByContentHash(t *testing.T) {
	repo := new(mocks.GraphRepository)
	logger := newTestLogger()
	gb := NewGraphBuilder(repo, logger)

	ctx := context.Background()

	// Mock returns existing node
	repo.EXPECT().FindNodeByContentHash(ctx, postgres.FindNodeByContentHashParams{
		NodeType:    "Device",
		ContentHash: pgtype.Text{String: "abc123hash", Valid: true},
	}).Return(postgres.GraphNode{
		ID:                    42,
		ExternalID:            "existing-uuid",
		NodeType:              "Device",
		Data:                  json.RawMessage(`{"name":"test-device"}`),
		DuplicateCount:        5,
		MatchingSchemaVersion: 1,
	}, nil).Once()

	result := gb.findNodeByContentHash(ctx, "Device", "abc123hash")

	require.NotNil(t, result)
	assert.Equal(t, int64(42), *result.NodeID)
	assert.Equal(t, "existing-uuid", *result.ExternalID)
	assert.Equal(t, float64(0.9), float64(result.Confidence))
	assert.Contains(t, result.MatchingFields, "content_hash")
	assert.Contains(t, result.MatchReason, "Content hash match")

	repo.AssertExpectations(t)
}

func TestFindNodeByContentHash_NotFound(t *testing.T) {
	repo := new(mocks.GraphRepository)
	logger := newTestLogger()
	gb := NewGraphBuilder(repo, logger)

	ctx := context.Background()

	// Mock returns no rows error
	repo.EXPECT().FindNodeByContentHash(ctx, postgres.FindNodeByContentHashParams{
		NodeType:    "Device",
		ContentHash: pgtype.Text{String: "unknown-hash", Valid: true},
	}).Return(postgres.GraphNode{}, pgx.ErrNoRows).Once()

	result := gb.findNodeByContentHash(ctx, "Device", "unknown-hash")

	assert.Nil(t, result)

	repo.AssertExpectations(t)
}

func TestFindNodeByContentHash_DatabaseError(t *testing.T) {
	repo := new(mocks.GraphRepository)
	logger := newTestLogger()
	gb := NewGraphBuilder(repo, logger)

	ctx := context.Background()

	// Mock returns database error
	repo.EXPECT().FindNodeByContentHash(ctx, postgres.FindNodeByContentHashParams{
		NodeType:    "Device",
		ContentHash: pgtype.Text{String: "some-hash", Valid: true},
	}).Return(postgres.GraphNode{}, errors.New("database connection failed")).Once()

	result := gb.findNodeByContentHash(ctx, "Device", "some-hash")

	assert.Nil(t, result)

	repo.AssertExpectations(t)
}
