package diode_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/netboxlabs/diode/diode-sdk-go/diode/v1"
)

func TestNewClient(t *testing.T) {
	cleanUpEnvVars()

	_ = os.Setenv(diode.DiodeAPIKeyEnvVarName, "nothingtoseehere")

	client, err := diode.NewClient()
	require.NoError(t, err)
	require.NotNil(t, client)
	require.NoError(t, client.Close())
}

func TestNewClientWithMissingAPIKey(t *testing.T) {
	cleanUpEnvVars()

	client, err := diode.NewClient()
	require.Nil(t, client)
	require.EqualError(t, err, "environment variable DIODE_API_KEY not found")
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

			client, err := diode.NewClient()
			require.NoError(t, err)
			require.NotNil(t, client)
		})
	}
}

func cleanUpEnvVars() {
	_ = os.Unsetenv(diode.DiodeAPIKeyEnvVarName)
	_ = os.Unsetenv(diode.DiodeGRPCHostEnvVarName)
	_ = os.Unsetenv(diode.DiodeGRPCPortEnvVarName)
	_ = os.Unsetenv(diode.DiodeGRPCInsecureEnvVarName)
}
