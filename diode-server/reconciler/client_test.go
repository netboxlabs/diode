package reconciler_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/netboxlabs/diode/diode-server/reconciler"
)

func TestNewClient(t *testing.T) {
	cleanUpEnvVars()

	client, err := reconciler.NewClient()
	require.NoError(t, err)
	require.NotNil(t, client)
	require.NoError(t, client.Close())
}

func cleanUpEnvVars() {
	_ = os.Unsetenv(reconciler.GRPCHostEnvVarName)
	_ = os.Unsetenv(reconciler.GRPCPortEnvVarName)
}
