package auth_test

import (
	"context"
	"log/slog"
	"net"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/netboxlabs/diode/diode-server/auth"
)

func TestNewServer(t *testing.T) {
	ctx := context.Background()

	setupEnv()
	defer teardownEnv()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))
	server, err := auth.NewServer(ctx, logger)
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

func setupEnv() {
	httpPort, _ := getFreePort()
	_ = os.Setenv("HTTP_PORT", httpPort)
}

func teardownEnv() {
	_ = os.Unsetenv("HTTP_PORT")
}
