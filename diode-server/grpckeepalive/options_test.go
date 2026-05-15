package grpckeepalive

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func TestServerOptions_ServeBufConn(t *testing.T) {
	t.Parallel()

	const bufSize = 1024 * 1024
	lis := bufconn.Listen(bufSize)
	srv := grpc.NewServer(ServerOptions()...)

	errCh := make(chan error, 1)
	go func() { errCh <- srv.Serve(lis) }()
	t.Cleanup(func() {
		srv.Stop()
		if err := <-errCh; err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			t.Errorf("Serve returned unexpected error: %v", err)
		}
	})

	dial := func(context.Context, string) (net.Conn, error) { return lis.Dial() }
	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(dial),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	conn.Connect()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for {
		st := conn.GetState()
		if st == connectivity.Ready {
			break
		}
		if !conn.WaitForStateChange(ctx, st) {
			require.Failf(t, "connection did not become ready", "state=%s", st.String())
		}
	}
}
