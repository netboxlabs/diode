package reconciler_test

import (
	"context"
	"log/slog"
	"net"
	"os"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/netboxlabs/diode/diode-server/reconciler"
	pb "github.com/netboxlabs/diode/diode-server/reconciler/v1/reconcilerpb"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func startTestServer(ctx context.Context, t *testing.T, redisAddr string) (*reconciler.Server, *grpc.ClientConn) {
	setupEnv(redisAddr)
	defer teardownEnv()

	listener := bufconn.Listen(bufSize)
	s := grpc.NewServer()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))
	server, err := reconciler.NewServer(ctx, logger)
	require.NoError(t, err)

	pb.RegisterReconcilerServiceServer(s, server)
	errChan := make(chan error, 1)
	go func() {
		errChan <- s.Serve(listener)
	}()

	bufDialer := func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}

	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	select {
	case err := <-errChan:
		require.NoError(t, err)
	default:
	}

	return server, conn
}

func TestNewServer(t *testing.T) {
	ctx := context.Background()
	s := miniredis.RunT(t)
	defer s.Close()

	setupEnv(s.Addr())
	defer teardownEnv()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))
	server, err := reconciler.NewServer(ctx, logger)
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

func TestRetrieveIngestionDataSources(t *testing.T) {
	tests := []struct {
		name         string
		requestName  string
		sdkVersion   string
		sdkName      string
		errorMessage string
		hasError     bool
	}{
		{
			name:         "missing SDK name",
			requestName:  "",
			sdkVersion:   "1.0",
			sdkName:      "",
			errorMessage: "sdk name is empty",
			hasError:     true,
		},
		{
			name:         "missing SDK version",
			requestName:  "",
			sdkVersion:   "",
			sdkName:      "test-sdk",
			errorMessage: "sdk version is empty",
			hasError:     true,
		},
		{
			name:         "invalid data source",
			requestName:  "INVALID",
			sdkVersion:   "1.0",
			sdkName:      "test-sdk",
			errorMessage: "",
			hasError:     true,
		},
		{
			name:         "valid request",
			requestName:  "",
			sdkVersion:   "1.0",
			sdkName:      "test-sdk",
			errorMessage: "",
			hasError:     false,
		},
		{
			name:         "valid request with filter",
			requestName:  "DIODE",
			sdkVersion:   "1.0",
			sdkName:      "test-sdk",
			errorMessage: "",
			hasError:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			s := miniredis.RunT(t)
			defer s.Close()

			server, conn := startTestServer(ctx, t, s.Addr())

			req := &pb.RetrieveIngestionDataSourcesRequest{
				Name:       tt.requestName,
				SdkVersion: tt.sdkVersion,
				SdkName:    tt.sdkName,
			}
			resp, err := server.RetrieveIngestionDataSources(ctx, req)
			if tt.hasError {
				require.Error(t, err)
				require.Nil(t, resp)
				require.Contains(t, err.Error(), tt.errorMessage)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
			}

			err = server.Stop()
			require.NoError(t, err)
			err = conn.Close()
			require.NoError(t, err)
		})
	}
}
