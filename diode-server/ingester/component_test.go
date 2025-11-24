package ingester_test

import (
	"context"
	"log/slog"
	"net"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/kelseyhightower/envconfig"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"github.com/netboxlabs/diode/diode-server/authutil"
	pb "github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
	"github.com/netboxlabs/diode/diode-server/ingester"
	"github.com/netboxlabs/diode/diode-server/reconciler"
	"github.com/netboxlabs/diode/diode-server/reconciler/mocks"
)

func getFreePort() (string, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return strconv.Itoa(0), err
	}

	addr := listener.Addr().(*net.TCPAddr)

	if err = listener.Close(); err != nil {
		return strconv.Itoa(0), err
	}
	return strconv.Itoa(addr.Port), nil
}

func setupEnv(redisAddr string) {
	host, port, _ := net.SplitHostPort(redisAddr)
	grpcPort, _ := getFreePort()
	_ = os.Setenv("GRPC_PORT", grpcPort)
	_ = os.Setenv("REDIS_HOST", host)
	_ = os.Setenv("REDIS_PORT", port)
	_ = os.Setenv("REDIS_USERNAME", "")
	_ = os.Setenv("REDIS_PASSWORD", "")
	_ = os.Setenv("REDIS_DB", "0")
	_ = os.Setenv("REDIS_STREAM_DB", "1")
	_ = os.Setenv("NETBOX_DIODE_PLUGIN_API_BASE_URL", "http://example.com")
	_ = os.Setenv("NETBOX_DIODE_PLUGIN_SKIP_TLS_VERIFY", "true")
	_ = os.Setenv("DIODE_AUTH_TOKEN_URL", "http://example.com")
	_ = os.Setenv("DIODE_TO_NETBOX_CLIENT_ID", "test-client-id")
	_ = os.Setenv("DIODE_TO_NETBOX_CLIENT_SECRET", "test-client-secret")
}

func teardownEnv() {
	_ = os.Unsetenv("GRPC_PORT")
	_ = os.Unsetenv("REDIS_HOST")
	_ = os.Unsetenv("REDIS_PORT")
	_ = os.Unsetenv("REDIS_USERNAME")
	_ = os.Unsetenv("REDIS_PASSWORD")
	_ = os.Unsetenv("REDIS_DB")
	_ = os.Unsetenv("REDIS_STREAM_DB")
	_ = os.Unsetenv("NETBOX_DIODE_PLUGIN_API_BASE_URL")
	_ = os.Unsetenv("NETBOX_DIODE_PLUGIN_SKIP_TLS_VERIFY")
	_ = os.Unsetenv("DIODE_AUTH_TOKEN_URL")
	_ = os.Unsetenv("DIODE_TO_NETBOX_CLIENT_ID")
	_ = os.Unsetenv("DIODE_TO_NETBOX_CLIENT_SECRET")
}

const bufSize = 1024 * 1024

func startReconcilerServer(ctx context.Context, t *testing.T) *reconciler.Server {
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

	errChan := make(chan error, 1)
	go func() {
		errChan <- server.Start(ctx)
	}()

	select {
	case err := <-errChan:
		require.NoError(t, err)
	default:
	}

	return server
}

func startTestComponent(ctx context.Context, t *testing.T, r *miniredis.Miniredis) (*ingester.Component, *grpc.ClientConn) {
	grpcPort, _ := getFreePort()
	_ = os.Setenv("GRPC_PORT", grpcPort)

	listener := bufconn.Listen(bufSize)
	s := grpc.NewServer()

	var cfg ingester.Config
	err := envconfig.Process("", &cfg)
	require.NoError(t, err)

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))

	meter := otel.GetMeterProvider().Meter("test.ingester")

	metrics, err := ingester.NewMetricRecorder(meter, "test")
	require.NoError(t, err)

	redisStreamClient := redis.NewClient(&redis.Options{
		Addr: r.Addr(),
		DB:   1,
	})

	authorizer := authutil.NewContextAuthorizer(logger)
	serverInterceptors := []grpc.UnaryServerInterceptor{
		authutil.NewUnverifiedJWTInterceptor(logger),
		func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
			if err := authorizer.RequireScopes(ctx, []string{authutil.ScopeDiodeIngest}); err != nil {
				return nil, err
			}
			return handler(ctx, req)
		},
	}

	component, err := ingester.New(ctx, logger, cfg, redisStreamClient, metrics, &ingester.DefaultStreamRouter{}, serverInterceptors...)
	require.NoError(t, err)

	pb.RegisterIngesterServiceServer(s, component)
	errChan := make(chan error, 1)
	go func() {
		errChan <- s.Serve(listener)
	}()

	bufDialer := func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}

	conn, err := grpc.NewClient("://bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	select {
	case err := <-errChan:
		require.NoError(t, err)
	default:
	}

	return component, conn
}

func TestNewComponent(t *testing.T) {
	ctx := context.Background()
	r := miniredis.RunT(t)
	defer r.Close()

	setupEnv(r.Addr())
	defer teardownEnv()

	server := startReconcilerServer(ctx, t)

	grpcPort, _ := getFreePort()
	_ = os.Setenv("GRPC_PORT", grpcPort)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))

	var cfg ingester.Config
	err := envconfig.Process("", &cfg)
	require.NoError(t, err)

	meter := otel.GetMeterProvider().Meter("test.ingester")

	metrics, err := ingester.NewMetricRecorder(meter, "test")
	require.NoError(t, err)

	redisStreamClient := redis.NewClient(&redis.Options{
		Addr: r.Addr(),
		DB:   1,
	})
	defer func() {
		_ = redisStreamClient.Close()
	}()
	authorizer := authutil.NewContextAuthorizer(logger)
	serverInterceptors := []grpc.UnaryServerInterceptor{
		authutil.NewUnverifiedJWTInterceptor(logger),
		func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
			if err := authorizer.RequireScopes(ctx, []string{authutil.ScopeDiodeIngest}); err != nil {
				return nil, err
			}
			return handler(ctx, req)
		},
	}

	component, err := ingester.New(ctx, logger, cfg, redisStreamClient, metrics, &ingester.DefaultStreamRouter{}, serverInterceptors...)

	require.NoError(t, err)
	require.NotNil(t, component)

	// Start and stop the component in a separate goroutine
	go func() {
		err = component.Start(ctx)
		require.NoError(t, err)
	}()

	// Wait for the component to start and stop
	time.Sleep(50 * time.Millisecond)

	err = server.Stop()
	require.NoError(t, err)
}

func TestIngest(t *testing.T) {
	tests := []struct {
		name         string
		request      *pb.IngestRequest
		errorMessage string
		hasError     bool
	}{
		{
			name: "missing ID",
			request: &pb.IngestRequest{
				Id:                 "",
				ProducerAppName:    "test-app",
				ProducerAppVersion: "1.0",
				SdkName:            "test-sdk",
				SdkVersion:         "1.0",
				Entities: []*pb.Entity{
					{
						Entity: &pb.Entity_Site{
							Site: &pb.Site{
								Name: "test-site-name",
							},
						},
					},
				},
			},

			errorMessage: "id is empty",
			hasError:     true,
		},
		{
			name: "missing ProducerAppName",
			request: &pb.IngestRequest{
				Id:                 "test-id",
				ProducerAppName:    "",
				ProducerAppVersion: "1.0",
				SdkName:            "test-sdk",
				SdkVersion:         "1.0",
				Entities: []*pb.Entity{
					{
						Entity: &pb.Entity_Site{
							Site: &pb.Site{
								Name: "test-site-name",
							},
						},
					},
				},
			},

			errorMessage: "producer app name is empty",
			hasError:     true,
		},
		{
			name: "missing ProducerAppVersion",
			request: &pb.IngestRequest{
				Id:                 "test-id",
				ProducerAppName:    "test-app",
				ProducerAppVersion: "",
				SdkName:            "test-sdk",
				SdkVersion:         "1.0",
				Entities: []*pb.Entity{
					{
						Entity: &pb.Entity_Site{
							Site: &pb.Site{
								Name: "test-site-name",
							},
						},
					},
				},
			},

			errorMessage: "producer app version is empty",
			hasError:     true,
		},
		{
			name: "missing SdkName",
			request: &pb.IngestRequest{
				Id:                 "test-id",
				ProducerAppName:    "test-app",
				ProducerAppVersion: "1.0",
				SdkName:            "",
				SdkVersion:         "1.0",
				Entities: []*pb.Entity{
					{
						Entity: &pb.Entity_Site{
							Site: &pb.Site{
								Name: "test-site-name",
							},
						},
					},
				},
			},

			errorMessage: "sdk name is empty",
			hasError:     true,
		},
		{
			name: "missing SdkVersion",
			request: &pb.IngestRequest{
				Id:                 "test-id",
				ProducerAppName:    "test-app",
				ProducerAppVersion: "1.0",
				SdkName:            "test-sdk",
				SdkVersion:         "",
				Entities: []*pb.Entity{
					{
						Entity: &pb.Entity_Site{
							Site: &pb.Site{
								Name: "test-site-name",
							},
						},
					},
				},
			},

			errorMessage: "sdk version is empty",
			hasError:     true,
		},
		{
			name: "missing Entities",
			request: &pb.IngestRequest{
				Id:                 "test-id",
				ProducerAppName:    "test-app",
				ProducerAppVersion: "1.0",
				SdkName:            "test-sdk",
				SdkVersion:         "1.0",
				Entities:           []*pb.Entity{},
			},
			errorMessage: "entities is empty",
			hasError:     true,
		},
		{
			name: "valid request",
			request: &pb.IngestRequest{
				Id:                 "test-id",
				ProducerAppName:    "test-app",
				ProducerAppVersion: "1.0",
				SdkName:            "test-sdk",
				SdkVersion:         "1.0",
				Entities: []*pb.Entity{
					{
						Entity: &pb.Entity_Site{
							Site: &pb.Site{
								Name: "test-site-name",
							},
						},
					},
				},
			},
			errorMessage: "",
			hasError:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			r := miniredis.RunT(t)
			defer r.Close()

			setupEnv(r.Addr())
			defer teardownEnv()

			server := startReconcilerServer(ctx, t)
			component, conn := startTestComponent(ctx, t, r)

			client := pb.NewIngesterServiceClient(conn)
			resp, err := client.Ingest(ctx, tt.request)

			if tt.hasError {
				require.Error(t, err)
				require.Nil(t, resp)
				require.Contains(t, err.Error(), tt.errorMessage)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
			}

			err = component.Stop()
			require.NoError(t, err)
			err = conn.Close()
			require.NoError(t, err)
			err = server.Stop()
			require.NoError(t, err)
		})
	}
}
