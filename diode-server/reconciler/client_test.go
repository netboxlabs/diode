package reconciler_test

import (
	"context"
	"github.com/netboxlabs/diode/diode-server/reconciler"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	cleanUpEnvVars()

	client, err := reconciler.NewClient(context.Background())
	require.NoError(t, err)
	require.NotNil(t, client)
	require.NoError(t, client.Close())
}

func TestNewClientWithTimeout(t *testing.T) {
	tests := []struct {
		desc       string
		timeoutStr string
		err        error
	}{
		{
			desc:       "timeout with valid value",
			timeoutStr: "10",
			err:        nil,
		},
		{
			desc:       "timeout with negative value",
			timeoutStr: "-1",
			err:        reconciler.ErrInvalidTimeout,
		},
		{
			desc:       "timeout with non-parseable value",
			timeoutStr: "10a",
			err:        reconciler.ErrInvalidTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			cleanUpEnvVars()

			_ = os.Setenv(reconciler.GRPCTimeoutSecondsEnvVarName, tt.timeoutStr)

			client, err := reconciler.NewClient(context.Background())
			if tt.err == nil {
				require.NoError(t, err)
				require.NotNil(t, client)
				require.NoError(t, client.Close())
			} else {
				require.Nil(t, client)
				require.EqualError(t, err, tt.err.Error())
			}
		})
	}
}

func cleanUpEnvVars() {
	_ = os.Unsetenv(reconciler.GRPCHostEnvVarName)
	_ = os.Unsetenv(reconciler.GRPCPortEnvVarName)
	_ = os.Unsetenv(reconciler.GRPCTimeoutSecondsEnvVarName)
}
