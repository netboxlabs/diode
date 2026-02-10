package reconciler

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/reconcilerpb"
	"github.com/netboxlabs/diode/diode-server/graph"
	"github.com/netboxlabs/diode/diode-server/graph/mocks"
)

const CurrentSchemaVersion = graph.CurrentSchemaVersion

func TestCreateEntity_ValidDevice(t *testing.T) {
	ctx := context.Background()
	mockRepo := mocks.NewRepository(t)

	device := &diodepb.Device{
		Name: newPtr("test-device"),
	}
	entity := &diodepb.Entity{
		Entity: &diodepb.Entity_Device{Device: device},
	}
	req := &reconcilerpb.CreateEntityRequest{
		Entity: entity,
	}

	expectedNode := graph.Node{
		ID:                    1,
		ExternalID:            "test-uuid-123",
		NodeType:              "Device",
		Data:                  []byte(`{"name":"test-device"}`),
		MatchingSchemaVersion: CurrentSchemaVersion,
	}

	// No existing node — new entity
	mockRepo.EXPECT().FindNodeByContentHash(ctx, mock.Anything).Return(graph.Node{}, graph.ErrNotFound).Once()

	mockRepo.EXPECT().UpsertNode(ctx, mock.MatchedBy(func(arg graph.UpsertNodeParams) bool {
		return arg.NodeType == "Device" &&
			arg.MatchingSchemaVersion == graph.CurrentSchemaVersion &&
			arg.ExternalID != "" // UUID should be generated
	})).Return(expectedNode, nil).Once()

	resp, err := createEntity(ctx, mockRepo, req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "test-uuid-123", resp.Id)
	assert.Equal(t, "Device", resp.ObjectType)
}

func TestCreateEntity_ValidSite(t *testing.T) {
	ctx := context.Background()
	mockRepo := mocks.NewRepository(t)

	site := &diodepb.Site{
		Name: "test-site",
	}
	entity := &diodepb.Entity{
		Entity: &diodepb.Entity_Site{Site: site},
	}
	req := &reconcilerpb.CreateEntityRequest{
		Entity: entity,
	}

	expectedNode := graph.Node{
		ID:                    2,
		ExternalID:            "site-uuid-456",
		NodeType:              "Site",
		Data:                  []byte(`{"name":"test-site"}`),
		MatchingSchemaVersion: CurrentSchemaVersion,
	}

	mockRepo.EXPECT().FindNodeByContentHash(ctx, mock.Anything).Return(graph.Node{}, graph.ErrNotFound).Once()

	mockRepo.EXPECT().UpsertNode(ctx, mock.MatchedBy(func(arg graph.UpsertNodeParams) bool {
		return arg.NodeType == "Site"
	})).Return(expectedNode, nil).Once()

	resp, err := createEntity(ctx, mockRepo, req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "site-uuid-456", resp.Id)
	assert.Equal(t, "Site", resp.ObjectType)
}

func TestCreateEntity_WithMetadata(t *testing.T) {
	ctx := context.Background()
	mockRepo := mocks.NewRepository(t)

	device := &diodepb.Device{
		Name: newPtr("test-device"),
	}
	entity := &diodepb.Entity{
		Entity: &diodepb.Entity_Device{Device: device},
	}

	metadata, _ := structpb.NewStruct(map[string]interface{}{
		"agent_id":  "agent-123",
		"policy_id": "policy-456",
	})

	req := &reconcilerpb.CreateEntityRequest{
		Entity:   entity,
		Metadata: metadata,
	}

	expectedNode := graph.Node{
		ID:                    1,
		ExternalID:            "test-uuid-123",
		NodeType:              "Device",
		Data:                  []byte(`{"name":"test-device"}`),
		MatchingSchemaVersion: CurrentSchemaVersion,
	}

	mockRepo.EXPECT().FindNodeByContentHash(ctx, mock.Anything).Return(graph.Node{}, graph.ErrNotFound).Once()

	mockRepo.EXPECT().UpsertNode(ctx, mock.MatchedBy(func(arg graph.UpsertNodeParams) bool {
		// Verify metadata is included
		return arg.NodeType == "Device" && len(arg.Metadata) > 2 // Not just "{}"
	})).Return(expectedNode, nil).Once()

	resp, err := createEntity(ctx, mockRepo, req)

	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestCreateEntity_NilRequest(t *testing.T) {
	ctx := context.Background()
	mockRepo := mocks.NewRepository(t)

	// Test with nil entity in request
	req := &reconcilerpb.CreateEntityRequest{
		Entity: nil,
	}

	resp, err := createEntity(ctx, mockRepo, req)

	require.Error(t, err)
	require.Nil(t, resp)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "entity is required")
}

func TestCreateEntity_EmptyEntity(t *testing.T) {
	ctx := context.Background()
	mockRepo := mocks.NewRepository(t)

	// Entity exists but inner entity is nil
	req := &reconcilerpb.CreateEntityRequest{
		Entity: &diodepb.Entity{},
	}

	resp, err := createEntity(ctx, mockRepo, req)

	require.Error(t, err)
	require.Nil(t, resp)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "entity is required")
}

func TestCreateEntity_DatabaseError(t *testing.T) {
	ctx := context.Background()
	mockRepo := mocks.NewRepository(t)

	device := &diodepb.Device{
		Name: newPtr("test-device"),
	}
	entity := &diodepb.Entity{
		Entity: &diodepb.Entity_Device{Device: device},
	}
	req := &reconcilerpb.CreateEntityRequest{
		Entity: entity,
	}

	mockRepo.EXPECT().FindNodeByContentHash(ctx, mock.Anything).Return(graph.Node{}, graph.ErrNotFound).Once()

	dbErr := errors.New("database connection failed")
	mockRepo.EXPECT().UpsertNode(ctx, mock.Anything).Return(graph.Node{}, dbErr).Once()

	resp, err := createEntity(ctx, mockRepo, req)

	require.Error(t, err)
	require.Nil(t, resp)
	assert.Contains(t, err.Error(), "database connection failed")
}

func TestCreateEntity_Idempotency(t *testing.T) {
	ctx := context.Background()

	device := &diodepb.Device{
		Name: newPtr("idempotent-device"),
	}
	entity := &diodepb.Entity{
		Entity: &diodepb.Entity_Device{Device: device},
	}
	req := &reconcilerpb.CreateEntityRequest{
		Entity: entity,
	}

	t.Run("same content hash returns same ID", func(t *testing.T) {
		expectedNode := graph.Node{
			ID:                    1,
			ExternalID:            "idempotent-uuid-123",
			NodeType:              "Device",
			Data:                  []byte(`{"name":"idempotent-device"}`),
			MatchingSchemaVersion: CurrentSchemaVersion,
			ContentHash:           newPtr("hash123"),
		}

		// First call - no existing node, generates new UUID
		mockRepo := mocks.NewRepository(t)
		mockRepo.EXPECT().FindNodeByContentHash(ctx, mock.Anything).Return(graph.Node{}, graph.ErrNotFound).Once()
		mockRepo.EXPECT().UpsertNode(ctx, mock.MatchedBy(func(arg graph.UpsertNodeParams) bool {
			return arg.NodeType == "Device" && arg.ContentHash != nil && *arg.ContentHash != ""
		})).Return(expectedNode, nil).Once()

		resp1, err := createEntity(ctx, mockRepo, req)
		require.NoError(t, err)
		require.NotNil(t, resp1)

		// Second call - existing node found by content hash, reuses ExternalID
		mockRepo2 := mocks.NewRepository(t)
		mockRepo2.EXPECT().FindNodeByContentHash(ctx, mock.Anything).Return(expectedNode, nil).Once()
		mockRepo2.EXPECT().UpsertNode(ctx, mock.MatchedBy(func(arg graph.UpsertNodeParams) bool {
			return arg.NodeType == "Device" &&
				arg.ExternalID == "idempotent-uuid-123" && // reuses existing ID
				arg.ContentHash != nil && *arg.ContentHash != ""
		})).Return(expectedNode, nil).Once()

		resp2, err := createEntity(ctx, mockRepo2, req)
		require.NoError(t, err)
		require.NotNil(t, resp2)

		assert.Equal(t, resp1.Id, resp2.Id, "repeated requests with same content should return same ID")
	})
}

func TestCreateEntity_IdempotencyPreservesExternalID(t *testing.T) {
	// Verifies that when a node already exists (found by content hash),
	// the same ExternalID is reused in the upsert.
	ctx := context.Background()

	device := &diodepb.Device{
		Name: newPtr("idempotent-device"),
	}
	entity := &diodepb.Entity{
		Entity: &diodepb.Entity_Device{Device: device},
	}
	req := &reconcilerpb.CreateEntityRequest{
		Entity: entity,
	}

	existingNode := graph.Node{
		ID:                    1,
		ExternalID:            "existing-uuid-999",
		NodeType:              "Device",
		MatchingSchemaVersion: CurrentSchemaVersion,
	}

	var capturedExternalIDs []string

	mockRepo := mocks.NewRepository(t)
	// First call: not found, generates new UUID
	mockRepo.EXPECT().FindNodeByContentHash(ctx, mock.Anything).Return(graph.Node{}, graph.ErrNotFound).Once()
	// Second call: found, reuses existing ExternalID
	mockRepo.EXPECT().FindNodeByContentHash(ctx, mock.Anything).Return(existingNode, nil).Once()

	mockRepo.EXPECT().UpsertNode(ctx, mock.Anything).Run(func(_ context.Context, arg graph.UpsertNodeParams) {
		capturedExternalIDs = append(capturedExternalIDs, arg.ExternalID)
	}).Return(existingNode, nil).Times(2)

	// First request — new entity
	_, err := createEntity(ctx, mockRepo, req)
	require.NoError(t, err)

	// Second identical request — existing entity found
	_, err = createEntity(ctx, mockRepo, req)
	require.NoError(t, err)

	require.Len(t, capturedExternalIDs, 2)
	assert.NotEqual(t, capturedExternalIDs[0], capturedExternalIDs[1],
		"first call should generate a new UUID")
	assert.Equal(t, "existing-uuid-999", capturedExternalIDs[1],
		"second call should reuse the existing ExternalID")
}

func TestCreateEntity_VariousEntityTypes(t *testing.T) {
	tests := []struct {
		name         string
		entity       *diodepb.Entity
		expectedType string
	}{
		{
			name: "Device",
			entity: &diodepb.Entity{
				Entity: &diodepb.Entity_Device{Device: &diodepb.Device{Name: newPtr("device-1")}},
			},
			expectedType: "Device",
		},
		{
			name: "Site",
			entity: &diodepb.Entity{
				Entity: &diodepb.Entity_Site{Site: &diodepb.Site{Name: "site-1"}},
			},
			expectedType: "Site",
		},
		{
			name: "Interface",
			entity: &diodepb.Entity{
				Entity: &diodepb.Entity_Interface{Interface: &diodepb.Interface{Name: "eth0"}},
			},
			expectedType: "Interface",
		},
		{
			name: "IpAddress",
			entity: &diodepb.Entity{
				Entity: &diodepb.Entity_IpAddress{IpAddress: &diodepb.IPAddress{Address: "192.168.1.1/24"}},
			},
			expectedType: "IpAddress",
		},
		{
			name: "Prefix",
			entity: &diodepb.Entity{
				Entity: &diodepb.Entity_Prefix{Prefix: &diodepb.Prefix{Prefix: "192.168.0.0/16"}},
			},
			expectedType: "Prefix",
		},
		{
			name: "DeviceRole",
			entity: &diodepb.Entity{
				Entity: &diodepb.Entity_DeviceRole{DeviceRole: &diodepb.DeviceRole{Name: "router"}},
			},
			expectedType: "DeviceRole",
		},
		{
			name: "DeviceType",
			entity: &diodepb.Entity{
				Entity: &diodepb.Entity_DeviceType{DeviceType: &diodepb.DeviceType{Model: "Model-X"}},
			},
			expectedType: "DeviceType",
		},
		{
			name: "Manufacturer",
			entity: &diodepb.Entity{
				Entity: &diodepb.Entity_Manufacturer{Manufacturer: &diodepb.Manufacturer{Name: "Cisco"}},
			},
			expectedType: "Manufacturer",
		},
		{
			name: "Platform",
			entity: &diodepb.Entity{
				Entity: &diodepb.Entity_Platform{Platform: &diodepb.Platform{Name: "IOS-XE"}},
			},
			expectedType: "Platform",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockRepo := mocks.NewRepository(t)

			req := &reconcilerpb.CreateEntityRequest{
				Entity: tt.entity,
			}

			expectedNode := graph.Node{
				ID:                    1,
				ExternalID:            "uuid-" + tt.name,
				NodeType:              tt.expectedType,
				MatchingSchemaVersion: CurrentSchemaVersion,
			}

			mockRepo.EXPECT().FindNodeByContentHash(ctx, mock.Anything).Return(graph.Node{}, graph.ErrNotFound).Once()

			mockRepo.EXPECT().UpsertNode(ctx, mock.MatchedBy(func(arg graph.UpsertNodeParams) bool {
				return arg.NodeType == tt.expectedType
			})).Return(expectedNode, nil).Once()

			resp, err := createEntity(ctx, mockRepo, req)

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, tt.expectedType, resp.ObjectType)
		})
	}
}

func TestCreateEntity_ContentHashGeneration(t *testing.T) {
	ctx := context.Background()
	mockRepo := mocks.NewRepository(t)

	device := &diodepb.Device{
		Name: newPtr("hash-test-device"),
	}
	entity := &diodepb.Entity{
		Entity: &diodepb.Entity_Device{Device: device},
	}
	req := &reconcilerpb.CreateEntityRequest{
		Entity: entity,
	}

	var capturedContentHash *string

	expectedNode := graph.Node{
		ID:                    1,
		ExternalID:            "test-uuid",
		NodeType:              "Device",
		MatchingSchemaVersion: CurrentSchemaVersion,
	}

	mockRepo.EXPECT().FindNodeByContentHash(ctx, mock.Anything).Return(graph.Node{}, graph.ErrNotFound).Once()

	mockRepo.EXPECT().UpsertNode(ctx, mock.Anything).Run(func(_ context.Context, arg graph.UpsertNodeParams) {
		capturedContentHash = arg.ContentHash
	}).Return(expectedNode, nil).Once()

	_, err := createEntity(ctx, mockRepo, req)
	require.NoError(t, err)

	// Verify content hash was generated and is valid
	assert.NotEmpty(t, capturedContentHash, "content hash should not be empty")
}

func TestCreateEntity_MetadataDefaultsToEmptyObject(t *testing.T) {
	ctx := context.Background()
	mockRepo := mocks.NewRepository(t)

	device := &diodepb.Device{
		Name: newPtr("test-device"),
	}
	entity := &diodepb.Entity{
		Entity: &diodepb.Entity_Device{Device: device},
	}
	req := &reconcilerpb.CreateEntityRequest{
		Entity:   entity,
		Metadata: nil, // No metadata provided
	}

	var capturedMetadata []byte

	expectedNode := graph.Node{
		ID:                    1,
		ExternalID:            "test-uuid",
		NodeType:              "Device",
		MatchingSchemaVersion: CurrentSchemaVersion,
	}

	mockRepo.EXPECT().FindNodeByContentHash(ctx, mock.Anything).Return(graph.Node{}, graph.ErrNotFound).Once()

	mockRepo.EXPECT().UpsertNode(ctx, mock.Anything).Run(func(_ context.Context, arg graph.UpsertNodeParams) {
		capturedMetadata = arg.Metadata
	}).Return(expectedNode, nil).Once()

	_, err := createEntity(ctx, mockRepo, req)
	require.NoError(t, err)

	// Verify metadata defaults to empty JSON object
	assert.Equal(t, []byte("{}"), capturedMetadata)
}
