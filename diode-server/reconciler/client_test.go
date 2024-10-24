package reconciler_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	pb "github.com/netboxlabs/diode/diode-server/gen/diode/v1/reconcilerpb"
	"github.com/netboxlabs/diode/diode-server/reconciler"
)

func TestNewClient(t *testing.T) {
	cleanUpEnvVars()

	client, err := reconciler.NewClient()
	require.NoError(t, err)
	require.NotNil(t, client)

	ctx := context.Background()
	req := &pb.RetrieveIngestionDataSourcesRequest{}

	resp, err := client.RetrieveIngestionDataSources(ctx, req)
	require.Error(t, err)
	require.Nil(t, resp)

	require.NoError(t, client.Close())
}

func cleanUpEnvVars() {
	_ = os.Unsetenv(reconciler.GRPCHostEnvVarName)
	_ = os.Unsetenv(reconciler.GRPCPortEnvVarName)
}
