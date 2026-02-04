package reconciler_test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	"github.com/netboxlabs/diode/diode-server/authutil"
	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/reconcilerpb"
	"github.com/netboxlabs/diode/diode-server/reconciler"
	"github.com/netboxlabs/diode/diode-server/reconciler/mocks"
)

func TestNewServer(t *testing.T) {
	ctx := context.Background()
	s := miniredis.RunT(t)
	defer s.Close()

	setupEnv(s.Addr())
	defer teardownEnv()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))
	mockRepository := mocks.NewRepository(t)
	authorizer := authutil.NewContextAuthorizer(logger)
	serverInterceptors := []grpc.UnaryServerInterceptor{
		authutil.NewUnverifiedJWTInterceptor(logger),
		func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
			if err := authorizer.RequireScopes(ctx, []string{authutil.ScopeDiodeRead}); err != nil {
				return nil, err
			}
			return handler(ctx, req)
		},
	}
	server, err := reconciler.NewServer(ctx, logger, mockRepository, serverInterceptors...)
	require.NoError(t, err)
	require.NotNil(t, server)

	// Start and stop the server in a separate goroutine
	go func() {
		err = server.Start(ctx)
		require.NoError(t, err)
	}()

	// Wait for the server to start and stop
	time.Sleep(50 * time.Millisecond)
}

func TestCreateEntity(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*mocks.Repository)
		request        *reconcilerpb.CreateEntityRequest
		wantID         string
		wantObjectType string
		wantErr        bool
		errCode        string
	}{
		{
			name: "valid IPAddress",
			setupMock: func(m *mocks.Repository) {
				m.EXPECT().CreateEntityInGraph(mock.Anything, mock.Anything).
					Return("ip-address-123", "ipam.ipaddress", nil).
					Once()
			},
			request: &reconcilerpb.CreateEntityRequest{
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_IpAddress{
						IpAddress: &diodepb.IPAddress{
							Address: "192.168.1.1/24",
						},
					},
				},
			},
			wantID:         "ip-address-123",
			wantObjectType: "ipam.ipaddress",
			wantErr:        false,
		},
		{
			name: "valid Prefix",
			setupMock: func(m *mocks.Repository) {
				m.EXPECT().CreateEntityInGraph(mock.Anything, mock.Anything).
					Return("prefix-456", "ipam.prefix", nil).
					Once()
			},
			request: &reconcilerpb.CreateEntityRequest{
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Prefix{
						Prefix: &diodepb.Prefix{
							Prefix: "10.0.0.0/8",
						},
					},
				},
			},
			wantID:         "prefix-456",
			wantObjectType: "ipam.prefix",
			wantErr:        false,
		},
		{
			name: "valid Device",
			setupMock: func(m *mocks.Repository) {
				m.EXPECT().CreateEntityInGraph(mock.Anything, mock.Anything).
					Return("device-789", "dcim.device", nil).
					Once()
			},
			request: &reconcilerpb.CreateEntityRequest{
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Device{
						Device: &diodepb.Device{
							Name: stringPtr("test-device"),
						},
					},
				},
			},
			wantID:         "device-789",
			wantObjectType: "dcim.device",
			wantErr:        false,
		},
		{
			name:      "invalid nil entity",
			setupMock: func(m *mocks.Repository) {},
			request: &reconcilerpb.CreateEntityRequest{
				Entity: nil,
			},
			wantErr: true,
			errCode: "InvalidArgument",
		},
		{
			name: "invalid empty entity",
			setupMock: func(m *mocks.Repository) {
				m.EXPECT().CreateEntityInGraph(mock.Anything, mock.Anything).
					Return("", "", fmt.Errorf("entity has no type specified")).
					Once()
			},
			request: &reconcilerpb.CreateEntityRequest{
				Entity: &diodepb.Entity{},
			},
			wantErr: true,
			errCode: "Internal",
		},
		{
			name: "duplicate entity - idempotent behavior",
			setupMock: func(m *mocks.Repository) {
				m.EXPECT().CreateEntityInGraph(mock.Anything, mock.Anything).
					Return("ip-address-123", "ipam.ipaddress", nil).
					Once()
			},
			request: &reconcilerpb.CreateEntityRequest{
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_IpAddress{
						IpAddress: &diodepb.IPAddress{
							Address: "192.168.1.1/24",
						},
					},
				},
			},
			wantID:         "ip-address-123",
			wantObjectType: "ipam.ipaddress",
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			s := miniredis.RunT(t)
			defer s.Close()

			setupEnv(s.Addr())
			defer teardownEnv()

			logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError, AddSource: false}))
			mockRepository := mocks.NewRepository(t)
			tt.setupMock(mockRepository)

			server, err := reconciler.NewServer(ctx, logger, mockRepository)
			require.NoError(t, err)
			require.NotNil(t, server)

			resp, err := server.CreateEntity(ctx, tt.request)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errCode != "" {
					st, ok := status.FromError(err)
					require.True(t, ok, "error should be a gRPC status")
					require.Contains(t, st.Code().String(), tt.errCode)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Equal(t, tt.wantID, resp.Id)
			require.Equal(t, tt.wantObjectType, resp.ObjectType)
		})
	}
}

func TestListEntities(t *testing.T) {
	tests := []struct {
		name             string
		setupMock        func(*mocks.Repository)
		request          *reconcilerpb.ListEntitiesRequest
		wantEntityCount  int
		wantNextPageToken string
		wantErr          bool
		errCode          string
	}{
		{
			name: "list existing IPAddress entities",
			setupMock: func(m *mocks.Repository) {
				m.EXPECT().ListGraphEntities(mock.Anything, mock.Anything, int32(100), int32(0)).
					Return([]*reconcilerpb.DiodeEntity{
						{
							Id:         "ip-address-1",
							ObjectType: "ipam.ipaddress",
							Entity: &diodepb.Entity{
								Entity: &diodepb.Entity_IpAddress{
									IpAddress: &diodepb.IPAddress{
										Address: "192.168.1.1/24",
									},
								},
							},
						},
						{
							Id:         "ip-address-2",
							ObjectType: "ipam.ipaddress",
							Entity: &diodepb.Entity{
								Entity: &diodepb.Entity_IpAddress{
									IpAddress: &diodepb.IPAddress{
										Address: "192.168.1.2/24",
									},
								},
							},
						},
					}, nil).
					Once()
			},
			request: &reconcilerpb.ListEntitiesRequest{
				ObjectType: []string{"ipam.ipaddress"},
			},
			wantEntityCount:  2,
			wantNextPageToken: "",
			wantErr:          false,
		},
		{
			name: "list existing Prefix entities",
			setupMock: func(m *mocks.Repository) {
				m.EXPECT().ListGraphEntities(mock.Anything, mock.Anything, int32(100), int32(0)).
					Return([]*reconcilerpb.DiodeEntity{
						{
							Id:         "prefix-1",
							ObjectType: "ipam.prefix",
							Entity: &diodepb.Entity{
								Entity: &diodepb.Entity_Prefix{
									Prefix: &diodepb.Prefix{
										Prefix: "10.0.0.0/8",
									},
								},
							},
						},
					}, nil).
					Once()
			},
			request: &reconcilerpb.ListEntitiesRequest{
				ObjectType: []string{"ipam.prefix"},
			},
			wantEntityCount:  1,
			wantNextPageToken: "",
			wantErr:          false,
		},
		{
			name: "list existing Device entities",
			setupMock: func(m *mocks.Repository) {
				m.EXPECT().ListGraphEntities(mock.Anything, mock.Anything, int32(100), int32(0)).
					Return([]*reconcilerpb.DiodeEntity{
						{
							Id:         "device-1",
							ObjectType: "dcim.device",
							Entity: &diodepb.Entity{
								Entity: &diodepb.Entity_Device{
									Device: &diodepb.Device{
										Name: stringPtr("test-device-1"),
									},
								},
							},
						},
						{
							Id:         "device-2",
							ObjectType: "dcim.device",
							Entity: &diodepb.Entity{
								Entity: &diodepb.Entity_Device{
									Device: &diodepb.Device{
										Name: stringPtr("test-device-2"),
									},
								},
							},
						},
					}, nil).
					Once()
			},
			request: &reconcilerpb.ListEntitiesRequest{
				ObjectType: []string{"dcim.device"},
			},
			wantEntityCount:  2,
			wantNextPageToken: "",
			wantErr:          false,
		},
		{
			name: "list non-existing entities - empty result",
			setupMock: func(m *mocks.Repository) {
				m.EXPECT().ListGraphEntities(mock.Anything, mock.Anything, int32(100), int32(0)).
					Return([]*reconcilerpb.DiodeEntity{}, nil).
					Once()
			},
			request: &reconcilerpb.ListEntitiesRequest{
				ObjectType: []string{"nonexistent.type"},
			},
			wantEntityCount:  0,
			wantNextPageToken: "",
			wantErr:          false,
		},
		{
			name: "list all entity types",
			setupMock: func(m *mocks.Repository) {
				m.EXPECT().ListGraphEntities(mock.Anything, mock.Anything, int32(100), int32(0)).
					Return([]*reconcilerpb.DiodeEntity{
						{
							Id:         "ip-address-1",
							ObjectType: "ipam.ipaddress",
							Entity: &diodepb.Entity{
								Entity: &diodepb.Entity_IpAddress{
									IpAddress: &diodepb.IPAddress{
										Address: "192.168.1.1/24",
									},
								},
							},
						},
						{
							Id:         "prefix-1",
							ObjectType: "ipam.prefix",
							Entity: &diodepb.Entity{
								Entity: &diodepb.Entity_Prefix{
									Prefix: &diodepb.Prefix{
										Prefix: "10.0.0.0/8",
									},
								},
							},
						},
						{
							Id:         "device-1",
							ObjectType: "dcim.device",
							Entity: &diodepb.Entity{
								Entity: &diodepb.Entity_Device{
									Device: &diodepb.Device{
										Name: stringPtr("test-device"),
									},
								},
							},
						},
					}, nil).
					Once()
			},
			request:          &reconcilerpb.ListEntitiesRequest{},
			wantEntityCount:  3,
			wantNextPageToken: "",
			wantErr:          false,
		},
		{
			name: "pagination - full page with next token",
			setupMock: func(m *mocks.Repository) {
				pageSize := int32(2)
				entities := make([]*reconcilerpb.DiodeEntity, pageSize)
				for i := range entities {
					entities[i] = &reconcilerpb.DiodeEntity{
						Id:         fmt.Sprintf("device-%d", i+1),
						ObjectType: "dcim.device",
						Entity: &diodepb.Entity{
							Entity: &diodepb.Entity_Device{
								Device: &diodepb.Device{
									Name: stringPtr(fmt.Sprintf("device-%d", i+1)),
								},
							},
						},
					}
				}
				m.EXPECT().ListGraphEntities(mock.Anything, mock.Anything, pageSize, int32(0)).
					Return(entities, nil).
					Once()
			},
			request: &reconcilerpb.ListEntitiesRequest{
				PageSize: int32Ptr(2),
			},
			wantEntityCount:  2,
			wantNextPageToken: "2",
			wantErr:          false,
		},
		{
			name: "pagination - partial page without next token",
			setupMock: func(m *mocks.Repository) {
				m.EXPECT().ListGraphEntities(mock.Anything, mock.Anything, int32(100), int32(0)).
					Return([]*reconcilerpb.DiodeEntity{
						{
							Id:         "device-1",
							ObjectType: "dcim.device",
							Entity: &diodepb.Entity{
								Entity: &diodepb.Entity_Device{
									Device: &diodepb.Device{
										Name: stringPtr("device-1"),
									},
								},
							},
						},
					}, nil).
					Once()
			},
			request: &reconcilerpb.ListEntitiesRequest{
				PageSize: int32Ptr(100),
			},
			wantEntityCount:  1,
			wantNextPageToken: "",
			wantErr:          false,
		},
		{
			name: "invalid page token",
			setupMock: func(m *mocks.Repository) {},
			request: &reconcilerpb.ListEntitiesRequest{
				PageToken: "invalid-token",
			},
			wantErr: true,
			errCode: "InvalidArgument",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			s := miniredis.RunT(t)
			defer s.Close()

			setupEnv(s.Addr())
			defer teardownEnv()

			logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError, AddSource: false}))
			mockRepository := mocks.NewRepository(t)
			tt.setupMock(mockRepository)

			server, err := reconciler.NewServer(ctx, logger, mockRepository)
			require.NoError(t, err)
			require.NotNil(t, server)

			resp, err := server.ListEntities(ctx, tt.request)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errCode != "" {
					st, ok := status.FromError(err)
					require.True(t, ok, "error should be a gRPC status")
					require.Contains(t, st.Code().String(), tt.errCode)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Len(t, resp.Entities, tt.wantEntityCount)
			require.Equal(t, tt.wantNextPageToken, resp.NextPageToken)

			// Verify entity types if filtered
			if len(tt.request.ObjectType) > 0 {
				for _, entity := range resp.Entities {
					require.Contains(t, tt.request.ObjectType, entity.ObjectType)
				}
			}
		})
	}
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}
