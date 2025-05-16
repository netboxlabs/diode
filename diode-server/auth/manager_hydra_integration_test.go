//go:build integration_test

package auth_test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/netboxlabs/diode/diode-server/auth"
)

type logDumpConsumer struct {
	enabled bool
}

func (c *logDumpConsumer) Accept(log testcontainers.Log) {
	if !c.enabled {
		return
	}
	fmt.Printf("HYDRA LOG: %s\n", string(log.Content))
}

func TestHydraClientManager(t *testing.T) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "oryd/hydra:v2.3.0",
		ExposedPorts: []string{"4445/tcp"},
		WaitingFor:   wait.ForLog("Setting up http server on :4445"),
		Cmd:          []string{"serve", "all", "--dev"},
		Env: map[string]string{
			"DSN":                                      "memory",
			"STRATEGIES_ACCESS_TOKEN":                  "jwt",
			"STRATEGIES_REFRESH_TOKEN":                 "jwt",
			"STRATEGIES_JWT_SCOPE_CLAIM":               "both",
			"TTL_ACCESS_TOKEN":                         "1h",
			"OIDC_SUBJECT_IDENTIFIERS_SUPPORTED_TYPES": "public",
			"URLS_SELF_ISSUER":                         "http://127.0.0.1:4444",
			"SECRETS_SYSTEM":                           "some_very_secret_bytes",
		},
		LogConsumerCfg: &testcontainers.LogConsumerConfig{
			Opts:      []testcontainers.LogProductionOption{testcontainers.WithLogProductionTimeout(10 * time.Second)},
			Consumers: []testcontainers.LogConsumer{&logDumpConsumer{}},
		},
	}
	hydraC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	defer testcontainers.TerminateContainer(hydraC)
	require.NoError(t, err)

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	var manager auth.ClientManager
	authEndpoint, err := hydraC.Endpoint(ctx, "")
	require.NoError(t, err)
	authEndpoint = fmt.Sprintf("http://%s", authEndpoint)
	manager = auth.NewHydraClientManager(authEndpoint, logger)

	// no client initially
	currentClients, err := manager.RetrieveClients(ctx, auth.RetrieveClientsRequest{})
	require.NoError(t, err)
	require.Equal(t, 0, len(currentClients.Clients))

	// create a client
	createdClient, err := manager.CreateClient(ctx, auth.ClientInfo{
		ClientID:     "diode-test-client-1",
		ClientSecret: "secret-material",
		Scope:        "test:diode:1 test:diode:2",
	})
	require.NoError(t, err)
	require.NotNil(t, createdClient)
	require.Equal(t, "diode-test-client-1", createdClient.ClientID)
	require.Equal(t, "secret-material", createdClient.ClientSecret)
	require.Equal(t, "test:diode:1 test:diode:2", createdClient.Scope)

	// fetch the client by id
	client, err := manager.RetrieveClientByID(ctx, "diode-test-client-1")
	require.NoError(t, err)
	require.Equal(t, "diode-test-client-1", client.ClientID)
	require.Equal(t, "", client.ClientSecret)
	require.Equal(t, "test:diode:1 test:diode:2", client.Scope)

	// client should be in the list of clients
	currentClients, err = manager.RetrieveClients(ctx, auth.RetrieveClientsRequest{})
	require.NoError(t, err)
	require.Equal(t, 1, len(currentClients.Clients))
	require.Equal(t, "diode-test-client-1", currentClients.Clients[0].ClientID)
	require.Equal(t, "", currentClients.Clients[0].ClientSecret)
	require.Equal(t, "test:diode:1 test:diode:2", currentClients.Clients[0].Scope)

	// delete the client
	err = manager.DeleteClientByID(ctx, "diode-test-client-1")
	require.NoError(t, err)

	// should not be able to fetch the client by id
	_, err = manager.RetrieveClientByID(ctx, "diode-test-client-1")
	require.Error(t, err)
	require.Contains(t, err.Error(), "client diode-test-client-1 not found")

	// client should not be in the list of clients
	currentClients, err = manager.RetrieveClients(ctx, auth.RetrieveClientsRequest{})
	require.NoError(t, err)
	require.Equal(t, 0, len(currentClients.Clients))
}
