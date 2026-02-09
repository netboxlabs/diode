package reconciler

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/netboxlabs/diode/diode-server/gen/dbstore/postgres"
	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/reconcilerpb"
	"github.com/netboxlabs/diode/diode-server/reconciler/mocks"
)

func TestCreateEntity_ValidDevice(t *testing.T) {
	ctx := context.Background()
	mockRepo := mocks.NewGraphRepository(t)

	device := &diodepb.Device{
		Name: ptrString("test-device"),
	}
	entity := &diodepb.Entity{
		Entity: &diodepb.Entity_Device{Device: device},
	}
	req := &reconcilerpb.CreateEntityRequest{
		Entity: entity,
	}

	expectedNode := postgres.GraphNode{
		ID:                    1,
		ExternalID:            "test-uuid-123",
		NodeType:              "Device",
		Data:                  []byte(`{"name":"test-device"}`),
		MatchingSchemaVersion: CurrentSchemaVersion,
	}

	mockRepo.EXPECT().UpsertGraphNode(ctx, mock.MatchedBy(func(arg postgres.UpsertGraphNodeParams) bool {
		return arg.NodeType == "Device" &&
			arg.MatchingSchemaVersion == CurrentSchemaVersion &&
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
	mockRepo := mocks.NewGraphRepository(t)

	site := &diodepb.Site{
		Name: "test-site",
	}
	entity := &diodepb.Entity{
		Entity: &diodepb.Entity_Site{Site: site},
	}
	req := &reconcilerpb.CreateEntityRequest{
		Entity: entity,
	}

	expectedNode := postgres.GraphNode{
		ID:                    2,
		ExternalID:            "site-uuid-456",
		NodeType:              "Site",
		Data:                  []byte(`{"name":"test-site"}`),
		MatchingSchemaVersion: CurrentSchemaVersion,
	}

	mockRepo.EXPECT().UpsertGraphNode(ctx, mock.MatchedBy(func(arg postgres.UpsertGraphNodeParams) bool {
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
	mockRepo := mocks.NewGraphRepository(t)

	device := &diodepb.Device{
		Name: ptrString("test-device"),
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

	expectedNode := postgres.GraphNode{
		ID:                    1,
		ExternalID:            "test-uuid-123",
		NodeType:              "Device",
		Data:                  []byte(`{"name":"test-device"}`),
		MatchingSchemaVersion: CurrentSchemaVersion,
	}

	mockRepo.EXPECT().UpsertGraphNode(ctx, mock.MatchedBy(func(arg postgres.UpsertGraphNodeParams) bool {
		// Verify metadata is included
		return arg.NodeType == "Device" && len(arg.Metadata) > 2 // Not just "{}"
	})).Return(expectedNode, nil).Once()

	resp, err := createEntity(ctx, mockRepo, req)

	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestCreateEntity_NilRequest(t *testing.T) {
	ctx := context.Background()
	mockRepo := mocks.NewGraphRepository(t)

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
	mockRepo := mocks.NewGraphRepository(t)

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
	mockRepo := mocks.NewGraphRepository(t)

	device := &diodepb.Device{
		Name: ptrString("test-device"),
	}
	entity := &diodepb.Entity{
		Entity: &diodepb.Entity_Device{Device: device},
	}
	req := &reconcilerpb.CreateEntityRequest{
		Entity: entity,
	}

	dbErr := errors.New("database connection failed")
	mockRepo.EXPECT().UpsertGraphNode(ctx, mock.Anything).Return(postgres.GraphNode{}, dbErr).Once()

	resp, err := createEntity(ctx, mockRepo, req)

	require.Error(t, err)
	require.Nil(t, resp)
	assert.Contains(t, err.Error(), "database connection failed")
}

func TestCreateEntity_Idempotency(t *testing.T) {
	ctx := context.Background()

	device := &diodepb.Device{
		Name: ptrString("idempotent-device"),
	}
	entity := &diodepb.Entity{
		Entity: &diodepb.Entity_Device{Device: device},
	}
	req := &reconcilerpb.CreateEntityRequest{
		Entity: entity,
	}

	t.Run("same content hash returns same ID", func(t *testing.T) {
		mockRepo := mocks.NewGraphRepository(t)

		// First call - new entity
		expectedNode := postgres.GraphNode{
			ID:                    1,
			ExternalID:            "idempotent-uuid-123",
			NodeType:              "Device",
			Data:                  []byte(`{"name":"idempotent-device"}`),
			MatchingSchemaVersion: CurrentSchemaVersion,
			ContentHash:           pgtype.Text{String: "hash123", Valid: true},
		}

		mockRepo.EXPECT().UpsertGraphNode(ctx, mock.MatchedBy(func(arg postgres.UpsertGraphNodeParams) bool {
			return arg.NodeType == "Device" && arg.ContentHash.Valid
		})).Return(expectedNode, nil).Once()

		resp1, err := createEntity(ctx, mockRepo, req)
		require.NoError(t, err)
		require.NotNil(t, resp1)

		// Second call - should match on content hash via upsert
		mockRepo2 := mocks.NewGraphRepository(t)
		mockRepo2.EXPECT().UpsertGraphNode(ctx, mock.MatchedBy(func(arg postgres.UpsertGraphNodeParams) bool {
			return arg.NodeType == "Device" && arg.ContentHash.Valid
		})).Return(expectedNode, nil).Once()

		resp2, err := createEntity(ctx, mockRepo2, req)
		require.NoError(t, err)
		require.NotNil(t, resp2)

		// Both responses should reference the same entity ID
		// NOTE: Currently this will fail because a new UUID is generated each time
		// This test documents the expected idempotent behavior
		assert.Equal(t, resp1.Id, resp2.Id, "repeated requests with same content should return same ID")
	})
}

func TestCreateEntity_IdempotencyBroken(t *testing.T) {
	// This test demonstrates the current broken idempotency behavior
	// where each call generates a new external ID regardless of content hash
	ctx := context.Background()

	device := &diodepb.Device{
		Name: ptrString("idempotent-device"),
	}
	entity := &diodepb.Entity{
		Entity: &diodepb.Entity_Device{Device: device},
	}
	req := &reconcilerpb.CreateEntityRequest{
		Entity: entity,
	}

	var capturedExternalIDs []string

	mockRepo := mocks.NewGraphRepository(t)
	mockRepo.EXPECT().UpsertGraphNode(ctx, mock.Anything).Run(func(_ context.Context, arg postgres.UpsertGraphNodeParams) {
		capturedExternalIDs = append(capturedExternalIDs, arg.ExternalID)
	}).Return(postgres.GraphNode{
		ID:                    1,
		ExternalID:            "will-be-overwritten",
		NodeType:              "Device",
		MatchingSchemaVersion: CurrentSchemaVersion,
	}, nil).Times(2)

	// First request
	_, err := createEntity(ctx, mockRepo, req)
	require.NoError(t, err)

	// Second identical request
	_, err = createEntity(ctx, mockRepo, req)
	require.NoError(t, err)

	// Verify two different UUIDs were generated (demonstrating broken idempotency)
	require.Len(t, capturedExternalIDs, 2)
	assert.NotEqual(t, capturedExternalIDs[0], capturedExternalIDs[1],
		"BUG: each createEntity call generates a new UUID instead of being idempotent")
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
				Entity: &diodepb.Entity_Device{Device: &diodepb.Device{Name: ptrString("device-1")}},
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
			mockRepo := mocks.NewGraphRepository(t)

			req := &reconcilerpb.CreateEntityRequest{
				Entity: tt.entity,
			}

			expectedNode := postgres.GraphNode{
				ID:                    1,
				ExternalID:            "uuid-" + tt.name,
				NodeType:              tt.expectedType,
				MatchingSchemaVersion: CurrentSchemaVersion,
			}

			mockRepo.EXPECT().UpsertGraphNode(ctx, mock.MatchedBy(func(arg postgres.UpsertGraphNodeParams) bool {
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
	mockRepo := mocks.NewGraphRepository(t)

	device := &diodepb.Device{
		Name: ptrString("hash-test-device"),
	}
	entity := &diodepb.Entity{
		Entity: &diodepb.Entity_Device{Device: device},
	}
	req := &reconcilerpb.CreateEntityRequest{
		Entity: entity,
	}

	var capturedContentHash pgtype.Text

	expectedNode := postgres.GraphNode{
		ID:                    1,
		ExternalID:            "test-uuid",
		NodeType:              "Device",
		MatchingSchemaVersion: CurrentSchemaVersion,
	}

	mockRepo.EXPECT().UpsertGraphNode(ctx, mock.Anything).Run(func(_ context.Context, arg postgres.UpsertGraphNodeParams) {
		capturedContentHash = arg.ContentHash
	}).Return(expectedNode, nil).Once()

	_, err := createEntity(ctx, mockRepo, req)
	require.NoError(t, err)

	// Verify content hash was generated and is valid
	assert.True(t, capturedContentHash.Valid, "content hash should be generated")
	assert.NotEmpty(t, capturedContentHash.String, "content hash should not be empty")
}

func TestCreateEntity_MetadataDefaultsToEmptyObject(t *testing.T) {
	ctx := context.Background()
	mockRepo := mocks.NewGraphRepository(t)

	device := &diodepb.Device{
		Name: ptrString("test-device"),
	}
	entity := &diodepb.Entity{
		Entity: &diodepb.Entity_Device{Device: device},
	}
	req := &reconcilerpb.CreateEntityRequest{
		Entity:   entity,
		Metadata: nil, // No metadata provided
	}

	var capturedMetadata []byte

	expectedNode := postgres.GraphNode{
		ID:                    1,
		ExternalID:            "test-uuid",
		NodeType:              "Device",
		MatchingSchemaVersion: CurrentSchemaVersion,
	}

	mockRepo.EXPECT().UpsertGraphNode(ctx, mock.Anything).Run(func(_ context.Context, arg postgres.UpsertGraphNodeParams) {
		capturedMetadata = arg.Metadata
	}).Return(expectedNode, nil).Once()

	_, err := createEntity(ctx, mockRepo, req)
	require.NoError(t, err)

	// Verify metadata defaults to empty JSON object
	assert.Equal(t, []byte("{}"), capturedMetadata)
}
