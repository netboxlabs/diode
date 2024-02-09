package diode_test

import (
	"context"
	"os"
	"testing"

	"github.com/netboxlabs/diode/diode-sdk-go/diode/v1"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	cleanUpEnvVars()

	_ = os.Setenv(diode.DiodeAPIKeyEnvVarName, "nothingtoseehere")

	client, err := diode.NewClient(context.Background())
	require.NoError(t, err)
	require.NotNil(t, client)
	require.NoError(t, client.Close())
}

func TestNewClientWithMissingAPIKey(t *testing.T) {
	cleanUpEnvVars()

	client, err := diode.NewClient(context.Background())
	require.Nil(t, client)
	require.EqualError(t, err, "environment variable DIODE_API_KEY not found")
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
			err:        diode.ErrInvalidTimeout,
		},
		{
			desc:       "timeout with non-parseable value",
			timeoutStr: "10a",
			err:        diode.ErrInvalidTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			cleanUpEnvVars()

			_ = os.Setenv(diode.DiodeAPIKeyEnvVarName, "nothingtoseehere")
			_ = os.Setenv(diode.DiodeGRPCTimeoutSecondsEnvVarName, tt.timeoutStr)

			client, err := diode.NewClient(context.Background())
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

func TestNewClientWithInsecureGRPC(t *testing.T) {
	tests := []struct {
		desc  string
		value string
	}{
		{
			desc:  "insecure gRPC enabled",
			value: "true",
		},
		{
			desc:  "insecure gRPC disabled",
			value: "false",
		},
		{
			desc:  "insecure gRPC disabled with invalid value",
			value: "invalidvalue",
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			cleanUpEnvVars()

			_ = os.Setenv(diode.DiodeAPIKeyEnvVarName, "nothingtoseehere")
			_ = os.Setenv(diode.DiodeGRPCInsecureEnvVarName, tt.value)

			client, err := diode.NewClient(context.Background())
			require.NoError(t, err)
			require.NotNil(t, client)
		})
	}
}

func cleanUpEnvVars() {
	_ = os.Unsetenv(diode.DiodeAPIKeyEnvVarName)
	_ = os.Unsetenv(diode.DiodeGRPCHostEnvVarName)
	_ = os.Unsetenv(diode.DiodeGRPCPortEnvVarName)
	_ = os.Unsetenv(diode.DiodeGRPCTimeoutSecondsEnvVarName)
	_ = os.Unsetenv(diode.DiodeGRPCInsecureEnvVarName)
}
